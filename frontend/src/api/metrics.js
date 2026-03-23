import { apiGet } from './client'

export async function fetchMetrics() {
  const data = await apiGet('/api/metrics')

  if (!data.ok) {
    throw new Error(data.error || 'Unknown error')
  }

  return data
}