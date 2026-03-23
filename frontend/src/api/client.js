const tg = window.Telegram?.WebApp

export async function apiGet(url) {
  const response = await fetch(url, {
    method: 'GET',
    cache: 'no-store',
    headers: {
      'X-Telegram-Init-Data': tg?.initData || ''
    }
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
    headers: {
      'Content-Type': 'application/json',
      'X-Telegram-Init-Data': tg?.initData || ''
    },
    body: JSON.stringify(body)
  })

  if (!response.ok) {
    throw new Error(`HTTP ${response.status}`)
  }

  return response.json()
}