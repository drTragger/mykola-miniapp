import { apiGet } from './client'

export async function fetchUpsHistory(limit = 288) {
  const response = await apiGet(`/api/ups/history?limit=${limit}`)
  return response.points || []
}