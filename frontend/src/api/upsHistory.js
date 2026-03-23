export async function fetchUpsHistory(limit = 288) {
  const response = await fetch(`/api/ups/history?limit=${limit}`, {
    method: 'GET',
    cache: 'no-store'
  })

  if (!response.ok) {
    throw new Error(`HTTP ${response.status}`)
  }

  const data = await response.json()

  if (!data.ok) {
    throw new Error(data.error || 'UPS history request failed')
  }

  return data.points || []
}