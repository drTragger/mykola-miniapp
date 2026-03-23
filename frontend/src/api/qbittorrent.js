import { apiGet, apiPost } from './client'

export async function fetchTorrents() {
  const data = await apiGet('/api/qbittorrent/torrents')

  if (!data.ok) {
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
  const data = await apiGet(`/api/qbittorrent/torrents/${encodeURIComponent(hash)}/peers`)
  return data.peers || []
}

async function postAction(url, payload) {
  const data = await apiPost(url, payload)

  if (!data.ok) {
    throw new Error(data.error || 'Помилка запиту')
  }

  return data
}