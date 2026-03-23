import { apiGet } from './client'

export async function fetchUps() {
  const data = await apiGet('/api/ups')

  if (!data.ok) {
    throw new Error(data.error || 'UPS request failed')
  }

  return data
}