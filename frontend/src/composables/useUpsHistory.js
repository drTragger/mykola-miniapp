import { ref } from 'vue'

const MAX_POINTS = 24

export function useUpsHistory() {
  const batteryPercentHistory = ref([])
  const cellDeltaHistory = ref([])

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

  function appendUps(data) {
    if (!data) return

    pushPoint(batteryPercentHistory, data.batteryPercent)
    pushPoint(cellDeltaHistory, data.cellDeltaMv)
  }

  return {
    batteryPercentHistory,
    cellDeltaHistory,
    appendUps
  }
}