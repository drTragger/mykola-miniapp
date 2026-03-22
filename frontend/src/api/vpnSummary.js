export async function fetchVpnSummary() {
  const response = await fetch('/api/vpn/summary', {
    method: 'GET',
    cache: 'no-store'
  })

  if (!response.ok) {
    throw new Error(`HTTP ${response.status}`)
  }

  const data = await response.json()

  if (!data.ok) {
    throw new Error(data.error || 'VPN summary request failed')
  }

  return data
}