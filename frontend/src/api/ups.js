export async function fetchUps() {
  const response = await fetch('/api/ups', {
    method: 'GET',
    cache: 'no-store'
  })

  if (!response.ok) {
    throw new Error(`HTTP ${response.status}`)
  }

  const data = await response.json()

  if (!data.ok) {
    throw new Error(data.error || 'UPS request failed')
  }

  return data
}