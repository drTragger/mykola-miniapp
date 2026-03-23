<script setup>
import { computed, onMounted, onBeforeUnmount, ref, watch } from 'vue'
import AppHeader from './components/AppHeader.vue'
import DeviceHero from './components/DeviceHero.vue'
import BottomNav from './components/BottomNav.vue'
import OverviewView from './views/OverviewView.vue'
import UpsView from './views/UpsView.vue'
import QBittorrentView from './views/QBittorrentView.vue'
import SystemDetailsView from './views/SystemDetailsView.vue'
import { fetchMetrics } from './api/metrics'
import { fetchUps } from './api/ups'
import { fetchUpsBattery } from './api/upsBattery'
import { fetchUpsHistory } from './api/upsHistory'
import { fetchSystemDetails } from './api/system'
import { fetchVpnSummary } from './api/vpnSummary'
import { formatCollectedAt, formatUptime } from './utils/formatters'
import { useMetricsHistory } from './composables/useMetricsHistory'
import mykolaImage from './assets/mykola-1.png'

const tg = window.Telegram?.WebApp

const isDev = import.meta.env.DEV

const isTelegramApp = computed(() => {
  return Boolean(tg && tg.initData && tg.initData.length > 0)
})

const allowStandaloneDebug = computed(() => {
  return isDev
})

const canRenderApp = computed(() => {
  return isTelegramApp.value || allowStandaloneDebug.value
})

if (isTelegramApp.value) {
  tg.ready()
  tg.expand()
}

const activeTab = ref('overview')
const status = ref('Завантаження...')
const lastUpdated = ref('—')

const metricsIntervalId = ref(null)
const heroBatteryIntervalId = ref(null)

const metrics = ref({
  overview: {},
  network: {},
  services: {}
})

const systemData = ref({
  system: {},
  network: {},
  vpn: {}
})
const systemLoading = ref(false)
const systemError = ref('')

const ups = ref(null)
const upsLoading = ref(false)
const upsError = ref('')

const heroBattery = ref(null)
const heroBatteryLoading = ref(false)

const vpnSummary = ref(null)
const vpnSummaryIntervalId = ref(null)

const batteryPercentHistory = ref([])
const cellDeltaHistory = ref([])
const upsHistoryIntervalId = ref(null)

const {
  cpuUsageHistory,
  cpuTempHistory,
  ramUsageHistory,
  rxSpeedHistory,
  txSpeedHistory,
  appendMetrics
} = useMetricsHistory()

const user = tg?.initDataUnsafe?.user

const subtitleLines = computed(() => {
  if (!user) {
    return ['Панель моніторингу Raspberry Pi']
  }

  const fullName = `${user.first_name ?? ''} ${user.last_name ?? ''}`.trim()

  return [
    'Панель моніторингу Raspberry Pi',
    fullName || user.username || 'Telegram user'
  ]
})

const heroTitle = computed(() => {
  return systemData.value?.system?.hostname || 'mykola-1'
})

const heroUptime = computed(() => {
  return formatUptime(metrics.value.overview?.uptimeSeconds || 0)
})

const heroBatteryPercent = computed(() => {
  return heroBattery.value?.batteryPercent ?? null
})

const metricsRefreshing = ref(false)

async function loadMetrics() {
  metricsRefreshing.value = true

  try {
    const data = await fetchMetrics()
    metrics.value = data
    appendMetrics(data)
    status.value = 'Онлайн'
    lastUpdated.value = `Оновлено: ${formatCollectedAt(data.collectedAt)}`
  } catch (error) {
    console.error(error)
    status.value = 'Помилка'
  } finally {
    metricsRefreshing.value = false
  }
}

async function loadUps() {
  upsLoading.value = true
  upsError.value = ''

  try {
    ups.value = await fetchUps()
  } catch (error) {
    console.error(error)
    upsError.value = error.message || 'Не вдалося завантажити UPS'
  } finally {
    upsLoading.value = false
  }
}

async function loadHeroBattery() {
  heroBatteryLoading.value = true

  try {
    heroBattery.value = await fetchUpsBattery()
  } catch (error) {
    console.error(error)
  } finally {
    heroBatteryLoading.value = false
  }
}

async function loadUpsHistory() {
  try {
    const points = await fetchUpsHistory(288)

    batteryPercentHistory.value = points.map((point) => ({
      time: point.time,
      value: point.batteryPercent
    }))

    cellDeltaHistory.value = points.map((point) => ({
      time: point.time,
      value: point.cellDeltaMv
    }))
  } catch (error) {
    console.error(error)
  }
}

function startUpsHistoryRefresh() {
  stopUpsHistoryRefresh()

  upsHistoryIntervalId.value = setInterval(() => {
    loadUpsHistory()
  }, 60000)
}

function stopUpsHistoryRefresh() {
  if (upsHistoryIntervalId.value) {
    clearInterval(upsHistoryIntervalId.value)
    upsHistoryIntervalId.value = null
  }
}

async function loadSystemDetails() {
  systemLoading.value = true
  systemError.value = ''

  try {
    systemData.value = await fetchSystemDetails()
  } catch (error) {
    console.error(error)
    systemError.value = error.message || 'Не вдалося завантажити системні дані'
  } finally {
    systemLoading.value = false
  }
}

async function loadVpnSummary() {
  try {
    vpnSummary.value = await fetchVpnSummary()
  } catch (error) {
    console.error(error)
  }
}

async function refreshAllData() {
  const tasks = [
    loadMetrics(),
    loadHeroBattery(),
    loadVpnSummary()
  ]

  if (activeTab.value === 'ups') {
    tasks.push(loadUps())
  }

  if (activeTab.value === 'system') {
    tasks.push(loadSystemDetails())
  }

  await Promise.allSettled(tasks)
}

watch(activeTab, async (tab) => {
  if (!canRenderApp.value) {
    return
  }

  if (tab === 'ups') {
    await Promise.all([
      loadUps(),
      loadUpsHistory()
    ])
    startUpsHistoryRefresh()
  } else {
    stopUpsHistoryRefresh()
  }

  if (tab === 'system' && !systemLoading.value && !systemData.value?.collectedAt) {
    await loadSystemDetails()
  }
})

onMounted(() => {
  if (!canRenderApp.value) {
    status.value = 'Недоступно поза Telegram'
    return
  }

  const topInset = tg?.contentSafeAreaInset?.top
    ?? tg?.safeAreaInset?.top
    ?? 0

  document.documentElement.style.setProperty('--tg-safe-top', `${topInset}px`)
  document.documentElement.style.setProperty(
    '--tg-extra-top',
    topInset > 0 ? '16px' : '0px'
  )

  loadMetrics()
  loadHeroBattery()
  loadSystemDetails()
  loadVpnSummary()
  loadUpsHistory()

  metricsIntervalId.value = setInterval(loadMetrics, 5000)
  heroBatteryIntervalId.value = setInterval(loadHeroBattery, 30000)
  vpnSummaryIntervalId.value = setInterval(loadVpnSummary, 30000)
})

onBeforeUnmount(() => {
  if (metricsIntervalId.value) {
    clearInterval(metricsIntervalId.value)
  }

  if (heroBatteryIntervalId.value) {
    clearInterval(heroBatteryIntervalId.value)
  }

  if (vpnSummaryIntervalId.value) {
    clearInterval(vpnSummaryIntervalId.value)
  }

  stopUpsHistoryRefresh()
})
</script>

<template>
  <div
    v-if="!canRenderApp"
    class="min-h-screen flex items-center justify-center px-4"
  >
    <div class="max-w-md w-full bg-panel rounded-3xl border border-white/10 shadow-custom p-6 text-center">
      <div
        class="w-16 h-16 mx-auto rounded-2xl bg-primary/15 border border-primary/20 flex items-center justify-center text-2xl mb-4"
      >
        📱
      </div>

      <div class="text-xl font-semibold text-white mb-2">
        Доступ лише через Telegram
      </div>

      <div class="text-sm text-white/60 leading-relaxed">
        Цей застосунок працює тільки всередині Telegram Mini App.
        Відкрий його через свого бота або через головний застосунок у Telegram.
      </div>
    </div>
  </div>

  <div
    v-else
    class="app-shell max-w-[920px] mx-auto px-4 pb-32 space-y-4"
  >
    <AppHeader :subtitle-lines="subtitleLines" :status="status" />

    <DeviceHero
      :title="heroTitle"
      :subtitle="lastUpdated"
      :status="status"
      :uptime="heroUptime"
      :hero-image="mykolaImage"
      :battery-percent="heroBatteryPercent"
      :refreshing="metricsRefreshing"
      @refresh="refreshAllData"
    />

    <OverviewView
      v-if="activeTab === 'overview'"
      :metrics="metrics"
      :vpn-summary="vpnSummary"
      :cpu-usage-history="cpuUsageHistory"
      :cpu-temp-history="cpuTempHistory"
      :ram-usage-history="ramUsageHistory"
      :rx-speed-history="rxSpeedHistory"
      :tx-speed-history="txSpeedHistory"
    />

    <QBittorrentView
      v-else-if="activeTab === 'qbittorrent'"
      :active="activeTab === 'qbittorrent'"
    />

    <UpsView
      v-else-if="activeTab === 'ups'"
      :ups="ups"
      :loading="upsLoading"
      :battery-percent-history="batteryPercentHistory"
      :cell-delta-history="cellDeltaHistory"
      :error="upsError"
    />

    <SystemDetailsView
      v-else-if="activeTab === 'system'"
      :system-data="systemData"
    />

    <BottomNav v-model="activeTab" />
  </div>
</template>