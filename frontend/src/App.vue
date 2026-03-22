<script setup>
import { computed, onMounted, onBeforeUnmount, ref } from 'vue'
import AppHeader from './components/AppHeader.vue'
import DeviceHero from './components/DeviceHero.vue'
import BottomNav from './components/BottomNav.vue'
import OverviewView from './views/OverviewView.vue'
import SystemView from './views/SystemView.vue'
import NetworkView from './views/NetworkView.vue'
import { fetchMetrics } from './api/metrics'
import { formatCollectedAt, formatUptime } from './utils/formatters'
import { useMetricsHistory } from './composables/useMetricsHistory'

const tg = window.Telegram?.WebApp
tg?.ready()
tg?.expand()

const activeTab = ref('overview')
const status = ref('Завантаження...')
const lastUpdated = ref('—')
const intervalId = ref(null)

const metrics = ref({
  overview: {},
  system: {},
  network: {}
})

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
const heroIp = computed(() => metrics.value.network?.localIpv4 || '—')

async function loadMetrics() {
  status.value = 'Оновлення...'

  try {
    const data = await fetchMetrics()
    metrics.value = data
    appendMetrics(data)
    status.value = 'Онлайн'
    lastUpdated.value = `Оновлено: ${formatCollectedAt(data.collectedAt)}`
  } catch (error) {
    console.error(error)
    status.value = 'Помилка'
  }
}

onMounted(() => {
  loadMetrics()
  intervalId.value = setInterval(loadMetrics, 5000)
})

onBeforeUnmount(() => {
  if (intervalId.value) {
    clearInterval(intervalId.value)
  }
})
</script>

<template>
  <div class="app shell">
    <AppHeader :subtitle="subtitle" :status="status" />

    <DeviceHero
      :title="heroTitle"
      :subtitle="lastUpdated"
      :status="status"
      :hostname="metrics.system?.hostname || '—'"
      :uptime="heroUptime"
      :local-ip="heroIp"
      hero-image="/hero.png"
    />

    <div class="top-actions">
      <button class="primary-action" @click="loadMetrics">Оновити зараз</button>
    </div>

    <OverviewView
      v-if="activeTab === 'overview'"
      :metrics="metrics"
      :cpu-usage-history="cpuUsageHistory"
      :cpu-temp-history="cpuTempHistory"
      :ram-usage-history="ramUsageHistory"
    />

    <SystemView v-else-if="activeTab === 'system'" :metrics="metrics" />
    <NetworkView v-else :metrics="metrics" />

    <BottomNav v-model="activeTab" />
  </div>
</template>