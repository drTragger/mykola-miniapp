<script setup>
import { computed, onMounted, onBeforeUnmount, ref } from 'vue'
import AppHeader from './components/AppHeader.vue'
import SectionTabs from './components/SectionTabs.vue'
import OverviewView from './views/OverviewView.vue'
import SystemView from './views/SystemView.vue'
import NetworkView from './views/NetworkView.vue'
import { fetchMetrics } from './api/metrics'
import { formatCollectedAt } from './utils/formatters'

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

const user = tg?.initDataUnsafe?.user

const subtitle = computed(() => {
  if (!user) {
    return 'Панель моніторингу Raspberry Pi'
  }

  const fullName = `${user.first_name ?? ''} ${user.last_name ?? ''}`.trim()
  return `Панель моніторингу Raspberry Pi • ${fullName || user.username || 'Telegram user'}`
})

async function loadMetrics() {
  status.value = 'Оновлення...'

  try {
    const data = await fetchMetrics()
    metrics.value = data
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
  <div class="app">
    <AppHeader :subtitle="subtitle" :status="status" />

    <SectionTabs v-model="activeTab" />

    <div class="actions">
      <button class="action-btn" @click="loadMetrics">Оновити</button>
      <button class="action-btn">{{ lastUpdated }}</button>
    </div>

    <OverviewView v-if="activeTab === 'overview'" :metrics="metrics" />
    <SystemView v-else-if="activeTab === 'system'" :metrics="metrics" />
    <NetworkView v-else :metrics="metrics" />
  </div>
</template>