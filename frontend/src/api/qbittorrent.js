export async function fetchTorrents() {
  const response = await fetch('/api/qbittorrent/torrents')
  const data = await response.json()

  if (!response.ok || !data.ok) {
    throw new Error(data.error || 'Не вдалося завантажити торенти')
  }

  return data.torrents || []
}

export async function pauseTorrents(hashes) {
  return postAction('/api/qbittorrent/torrents/pause', { hashes })
}

export async function resumeTorrents(hashes) {
  return postAction('/api/qbittorrent/torrents/resume', { hashes })
}

export async function deleteTorrents(hashes, deleteFiles = false) {
  return postAction('/api/qbittorrent/torrents/delete', { hashes, deleteFiles })
}

async function postAction(url, payload) {
  const response = await fetch(url, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(payload)
  })

  const data = await response.json()

  if (!response.ok || !data.ok) {
    throw new Error(data.error || 'Помилка запиту')
  }

  return data
}