import { ref } from 'vue'

const MAX_POINTS = 24

export function useMetricsHistory() {
  const cpuUsageHistory = ref([])
  const cpuTempHistory = ref([])
  const ramUsageHistory = ref([])
  const rxSpeedHistory = ref([])
  const txSpeedHistory = ref([])

  function pushPoint(collectionRef, value) {
    const now = new Date()
    const time = now.toLocaleTimeString('uk-UA', {
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit'
    })

    collectionRef.value.push({
      time,
      value: typeof value === 'number' ? value : 0
    })

    if (collectionRef.value.length > MAX_POINTS) {
      collectionRef.value.shift()
    }
  }

  function appendMetrics(metrics) {
    if (!metrics) return

    pushPoint(cpuUsageHistory, metrics.overview?.cpuUsagePercent)
    pushPoint(cpuTempHistory, metrics.overview?.cpuTemperatureCelsius)
    pushPoint(ramUsageHistory, metrics.overview?.ramUsagePercent)
    pushPoint(rxSpeedHistory, metrics.network?.rxSpeedBps)
    pushPoint(txSpeedHistory, metrics.network?.txSpeedBps)
  }

  return {
    cpuUsageHistory,
    cpuTempHistory,
    ramUsageHistory,
    rxSpeedHistory,
    txSpeedHistory,
    appendMetrics
  }
}