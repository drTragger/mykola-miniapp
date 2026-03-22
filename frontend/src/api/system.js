export async function fetchSystemDetails() {
  const response = await fetch('/api/system', {
    method: 'GET',
    cache: 'no-store'
  })

  if (!response.ok) {
    throw new Error(`HTTP ${response.status}`)
  }

  const data = await response.json()

  if (!data.ok) {
    throw new Error(data.error || 'System request failed')
  }

  return data
}