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

export async function fetchTorrentPeers(hash) {
  const response = await fetch(`/api/qbittorrent/torrents/${encodeURIComponent(hash)}/peers`)

  if (!response.ok) {
    throw new Error('Не вдалося завантажити піри')
  }

  const data = await response.json()
  return data.peers || []
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