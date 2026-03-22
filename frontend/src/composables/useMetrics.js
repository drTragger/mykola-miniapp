import { onBeforeUnmount, onMounted, ref } from 'vue'
import { fetchMetrics } from '@/api/metrics'

export function useMetrics() {
  const metrics = ref(null)
  const loading = ref(true)
  const error = ref('')
  const online = ref(false)

  let intervalId = null

  async function loadMetrics() {
    try {
      if (!metrics.value) {
        loading.value = true
      }

      error.value = ''
      const data = await fetchMetrics()
      metrics.value = data
      online.value = true
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Unknown error'
      online.value = false
    } finally {
      loading.value = false
    }
  }

  function startPolling() {
    stopPolling()
    intervalId = setInterval(loadMetrics, 5000)
  }

  function stopPolling() {
    if (intervalId) {
      clearInterval(intervalId)
      intervalId = null
    }
  }

  onMounted(async () => {
    await loadMetrics()
    startPolling()
  })

  onBeforeUnmount(() => {
    stopPolling()
  })

  return {
    metrics,
    loading,
    error,
    online,
    reload: loadMetrics,
  }
}