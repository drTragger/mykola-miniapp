export function formatPercent(value) {
  if (typeof value !== 'number') return '—'
  return `${value.toFixed(1)}%`
}

export function formatTemperature(value) {
  if (typeof value !== 'number' || value <= 0) return '—'
  return `${value.toFixed(1)}°C`
}

export function formatFrequency(value) {
  if (typeof value !== 'number' || value <= 0) return '—'
  return `${value.toFixed(0)} MHz`
}

export function formatFixed(value) {
  if (typeof value !== 'number') return '—'
  return value.toFixed(2)
}

export function formatBytes(bytes) {
  if (typeof bytes !== 'number' || bytes < 0) return '—'

  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let value = bytes
  let unitIndex = 0

  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024
    unitIndex++
  }

  return `${value.toFixed(value >= 10 ? 1 : 2)} ${units[unitIndex]}`
}

export function formatUptime(seconds) {
  if (typeof seconds !== 'number' || seconds < 0) return '—'

  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)

  const parts = []
  if (days > 0) parts.push(`${days}д`)
  if (hours > 0) parts.push(`${hours}г`)
  parts.push(`${minutes}х`)

  return parts.join(' ')
}

export function formatDateTime(unixSeconds) {
  if (typeof unixSeconds !== 'number' || unixSeconds <= 0) return '—'
  return new Date(unixSeconds * 1000).toLocaleString('uk-UA')
}

export function formatCollectedAt(value) {
  if (!value) return '—'
  return new Date(value).toLocaleTimeString('uk-UA', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}