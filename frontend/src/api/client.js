const tg = window.Telegram?.WebApp
const isDev = import.meta.env.DEV

function getHeaders(extra = {}) {
  const headers = { ...extra }

  if (tg?.initData) {
    headers['X-Telegram-Init-Data'] = tg.initData
    return headers
  }

  if (isDev) {
    headers['X-Debug-Dev-Access'] = '1'
    return headers
  }

  throw new Error('Mini App доступний лише в Telegram')
}

export async function apiGet(url) {
  const response = await fetch(url, {
    method: 'GET',
    cache: 'no-store',
    headers: getHeaders()
  })

  if (!response.ok) {
    throw new Error(`HTTP ${response.status}`)
  }

  return response.json()
}

export async function apiPost(url, body) {
  const response = await fetch(url, {
    method: 'POST',
    cache: 'no-store',
    headers: getHeaders({
      'Content-Type': 'application/json'
    }),
    body: JSON.stringify(body)
  })

  if (!response.ok) {
    throw new Error(`HTTP ${response.status}`)
  }

  return response.json()
}