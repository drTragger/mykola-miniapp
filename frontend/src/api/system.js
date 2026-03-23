import { apiGet } from './client'

export async function fetchSystemDetails() {
  const data = await apiGet('/api/system')

  if (!data.ok) {
    throw new Error(data.error || 'System request failed')
  }

  return data
}