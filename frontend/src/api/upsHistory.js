import { apiGet } from './client'

export async function fetchUpsHistory(limit = 288) {
  const data = await apiGet(`/api/ups/history?limit=${limit}`)

  if (!data.ok) {
    throw new Error(data.error || 'UPS history request failed')
  }

  return data.points || []
}