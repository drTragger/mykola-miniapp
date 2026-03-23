import { apiGet } from './client'

export async function fetchUpsBattery() {
  const data = await apiGet('/api/ups/battery')

  if (!data.ok) {
    throw new Error(data.error || 'UPS battery request failed')
  }

  return data
}