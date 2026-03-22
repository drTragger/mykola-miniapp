export async function fetchUpsBattery() {
  const response = await fetch('/api/ups/battery', {
    method: 'GET',
    cache: 'no-store'
  })

  if (!response.ok) {
    throw new Error(`HTTP ${response.status}`)
  }

  const data = await response.json()

  if (!data.ok) {
    throw new Error(data.error || 'UPS battery request failed')
  }

  return data
}