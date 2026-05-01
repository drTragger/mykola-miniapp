import { ref } from 'vue'
import { fetchMetricsHistory } from '../api/metricsHistory'

const MAX_POINTS = 288

export function useMetricsHistory() {
  const cpuUsageHistory = ref([])
  const cpuTempHistory = ref([])
  const ramUsageHistory = ref([])
  const rxSpeedHistory = ref([])
  const txSpeedHistory = ref([])

  function pushPoint(collectionRef, value, time) {
    const t = time ?? new Date().toLocaleTimeString('uk-UA', {
      hour: '2-digit',
      minute: '2-digit'
    })

    collectionRef.value.push({
      time: t,
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

  async function loadHistory() {
    try {
      const points = await fetchMetricsHistory(288)

      cpuUsageHistory.value = points.map(p => ({ time: p.time, value: p.cpuUsagePercent }))
      cpuTempHistory.value  = points.map(p => ({ time: p.time, value: p.cpuTempCelsius }))
      ramUsageHistory.value = points.map(p => ({ time: p.time, value: p.ramUsagePercent }))
      rxSpeedHistory.value  = points.map(p => ({ time: p.time, value: p.rxSpeedBps }))
      txSpeedHistory.value  = points.map(p => ({ time: p.time, value: p.txSpeedBps }))
    } catch (error) {
      console.error('loadHistory error:', error)
    }
  }

  return {
    cpuUsageHistory,
    cpuTempHistory,
    ramUsageHistory,
    rxSpeedHistory,
    txSpeedHistory,
    appendMetrics,
    loadHistory
  }
}