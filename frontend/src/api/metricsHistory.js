import { apiGet } from './client'

export async function fetchMetricsHistory(limit = 288) {
  const data = await apiGet(`/api/metrics/history?limit=${limit}`)

  if (!data.ok) {
    throw new Error(data.error || 'Unknown error')
  }

  return data.points
}