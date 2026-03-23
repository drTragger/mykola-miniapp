import { apiGet } from './client'

export async function fetchVpnSummary() {
  const data = await apiGet('/api/vpn/summary')

  if (!data.ok) {
    throw new Error(data.error || 'VPN summary request failed')
  }

  return data
}