export async function fetchMetrics() {
  const response = await fetch('/api/metrics', {
    method: 'GET',
    cache: 'no-store'
  })

  if (!response.ok) {
    throw new Error(`HTTP ${response.status}`)
  }

  const data = await response.json()

  if (!data.ok) {
    throw new Error(data.error || 'Unknown error')
  }

  return data
}