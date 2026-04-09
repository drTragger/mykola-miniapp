import { apiGet, apiPost } from './client'

export async function fetchProcesses(params = {}) {
  const query = new URLSearchParams()

  if (params.search) query.set('search', params.search)
  if (params.sort) query.set('sort', params.sort)
  if (params.dir) query.set('dir', params.dir)
  if (typeof params.limit === 'number') query.set('limit', String(params.limit))
  if (typeof params.offset === 'number') query.set('offset', String(params.offset))

  const suffix = query.toString() ? `?${query.toString()}` : ''
  const data = await apiGet(`/api/processes${suffix}`)

  if (!data.ok) {
    throw new Error(data.error || 'Processes request failed')
  }

  return data
}

export async function sendProcessAction(pid, action) {
  const data = await apiPost(`/api/processes/${pid}/action`, {
    action
  })

  if (!data.ok) {
    throw new Error(data.error || 'Process action failed')
  }

  return data
}