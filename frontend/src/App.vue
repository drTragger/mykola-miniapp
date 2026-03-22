<script setup>
import { computed, onMounted, onBeforeUnmount, ref, watch } from 'vue'
import AppHeader from './components/AppHeader.vue'
import DeviceHero from './components/DeviceHero.vue'
import BottomNav from './components/BottomNav.vue'
import OverviewView from './views/OverviewView.vue'
import SystemView from './views/SystemView.vue'
import NetworkView from './views/NetworkView.vue'
import UpsView from './views/UpsView.vue'
import { fetchMetrics } from './api/metrics'
import { fetchUps } from './api/ups'
import { fetchUpsBattery } from './api/upsBattery'
import { formatCollectedAt, formatUptime } from './utils/formatters'
import { useMetricsHistory } from './composables/useMetricsHistory'
import mykolaImage from './assets/mykola-1.png'

const tg = window.Telegram?.WebApp
tg?.ready()
tg?.expand()

const activeTab = ref('overview')
const status = ref('Завантаження...')
const lastUpdated = ref('—')

const metricsIntervalId = ref(null)
const heroBatteryIntervalId = ref(null)

const metrics = ref({
  overview: {},
  system: {},
  network: {},
  services: {},
  vpn: {}
})

const ups = ref(null)
const upsLoading = ref(false)
const upsError = ref('')

const heroBattery = ref(null)
const heroBatteryLoading = ref(false)

const {
  cpuUsageHistory,
  cpuTempHistory,
  ramUsageHistory,
  appendMetrics
} = useMetricsHistory()

const user = tg?.initDataUnsafe?.user

const subtitle = computed(() => {
  if (!user) {
    return 'Панель моніторингу Raspberry Pi'
  }

  const fullName = `${user.first_name ?? ''} ${user.last_name ?? ''}`.trim()
  return `Панель моніторингу Raspberry Pi • ${fullName || user.username || 'Telegram user'}`
})

const heroTitle = computed(() => metrics.value.system?.hostname || 'mykola-1')
const heroUptime = computed(() => formatUptime(metrics.value.overview?.uptimeSeconds || 0))

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

watch(activeTab, async (tab) => {
  if (tab === 'ups' && !ups.value && !upsLoading.value) {
    await loadUps()
  }
})

onMounted(() => {
  loadMetrics()
  loadHeroBattery()

  metricsIntervalId.value = setInterval(loadMetrics, 5000)
  heroBatteryIntervalId.value = setInterval(loadHeroBattery, 30000)
})

onBeforeUnmount(() => {
  if (metricsIntervalId.value) {
    clearInterval(metricsIntervalId.value)
  }

  if (heroBatteryIntervalId.value) {
    clearInterval(heroBatteryIntervalId.value)
  }
})
</script>

<template>
  <div class="max-w-[920px] mx-auto px-4 pb-32 pt-4 space-y-4">
    <AppHeader :subtitle="subtitle" :status="status" />

    <DeviceHero
      :title="heroTitle"
      :subtitle="lastUpdated"
      :status="status"
      :uptime="heroUptime"
      :hero-image="mykolaImage"
      :battery-percent="heroBatteryPercent"
      :refreshing="metricsRefreshing"
      @refresh="loadMetrics"
    />

    <OverviewView
      v-if="activeTab === 'overview'"
      :metrics="metrics"
      :cpu-usage-history="cpuUsageHistory"
      :cpu-temp-history="cpuTempHistory"
      :ram-usage-history="ramUsageHistory"
    />

    <SystemView v-else-if="activeTab === 'system'" :metrics="metrics" />
    <NetworkView v-else-if="activeTab === 'network'" :metrics="metrics" />
    <UpsView
      v-else-if="activeTab === 'ups'"
      :ups="ups"
      :loading="upsLoading"
      :error="upsError"
    />

    <BottomNav v-model="activeTab" />
  </div>
</template>