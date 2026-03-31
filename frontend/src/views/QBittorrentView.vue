<script setup>
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Tag from 'primevue/tag'
import Dialog from 'primevue/dialog'
import Checkbox from 'primevue/checkbox'
import ProgressBar from 'primevue/progressbar'
import Menu from 'primevue/menu'
import {
  fetchTorrents,
  fetchTorrentPeers,
  pauseTorrents,
  resumeTorrents,
  deleteTorrents
} from '../api/qbittorrent'

const props = defineProps({
  active: {
    type: Boolean,
    default: false
  }
})

const torrents = ref([])
const loading = ref(false)
const refreshing = ref(false)
const error = ref('')
const selectedTorrents = ref([])
const search = ref('')

const stateFilter = ref('all')
const sortBy = ref('activity')
const sortDirection = ref('desc')

const deleteDialogVisible = ref(false)
const deleteFiles = ref(false)
const deleteTargetHashes = ref([])

const refreshIntervalId = ref(null)
const peersRefreshIntervalId = ref(null)
const actionLoading = ref(false)

const rowActionMenu = ref()
const rowActionTorrent = ref(null)
const searchInputRef = ref(null)

const expandedTorrentHashes = ref([])
const peersByHash = ref({})
const peersLoadingByHash = ref({})
const peersErrorByHash = ref({})

const filterOptions = [
  { value: 'all', label: 'Усі' },
  { value: 'downloading', label: 'Качаються' },
  { value: 'uploading', label: 'Роздаються' },
  { value: 'paused', label: 'Пауза' },
  { value: 'queued', label: 'У черзі' },
  { value: 'checking', label: 'Перевірка' },
  { value: 'error', label: 'Проблемні' },
  { value: 'completed', label: 'Готові' }
]

const sortOptions = [
  { value: 'activity', label: 'Активність' },
  { value: 'name', label: 'Назва' },
  { value: 'progress', label: 'Прогрес' },
  { value: 'addedOn', label: 'Дата додавання' },
  { value: 'size', label: 'Розмір' },
  { value: 'downloadSpeed', label: 'Швидкість ↓' },
  { value: 'uploadSpeed', label: 'Швидкість ↑' },
  { value: 'eta', label: 'ETA' },
  { value: 'ratio', label: 'Ratio' },
  { value: 'seeds', label: 'Seeds' }
]

const selectedCount = computed(() => selectedTorrents.value.length)

const stats = computed(() => {
  const list = torrents.value

  return {
    total: list.length,
    downloading: list.filter((torrent) => isDownloadingState(torrent.state)).length,
    uploading: list.filter((torrent) => isUploadingState(torrent.state)).length,
    paused: list.filter((torrent) => isPausedState(torrent.state)).length,
    queued: list.filter((torrent) => isQueuedState(torrent.state)).length,
    error: list.filter((torrent) => isErrorState(torrent.state)).length
  }
})

const rowActionItems = computed(() => {
  if (!rowActionTorrent.value) {
    return []
  }

  const torrent = rowActionTorrent.value
  const items = []

  if (isPausedState(torrent.state)) {
    items.push({
      label: 'Продовжити',
      icon: 'pi pi-play !text-emerald-500',
      command: async () => {
        await resumeOne(torrent)
      }
    })
  } else {
    items.push({
      label: 'Пауза',
      icon: 'pi pi-pause !text-amber-500',
      command: async () => {
        await pauseOne(torrent)
      }
    })
  }

  items.push({
    label: 'Видалити',
    icon: 'pi pi-trash !text-red-500',
    command: () => {
      askDeleteOne(torrent)
    }
  })

  return items
})

const filteredTorrents = computed(() => {
  const term = search.value.trim().toLowerCase()

  return torrents.value.filter((torrent) => {
    const matchesSearch = !term || [
      torrent.name,
      torrent.state,
      torrent.category,
      torrent.savePath
    ].some((value) => String(value || '').toLowerCase().includes(term))

    if (!matchesSearch) {
      return false
    }

    switch (stateFilter.value) {
      case 'downloading':
        return isDownloadingState(torrent.state)
      case 'uploading':
        return isUploadingState(torrent.state)
      case 'paused':
        return isPausedState(torrent.state)
      case 'queued':
        return isQueuedState(torrent.state)
      case 'checking':
        return isCheckingState(torrent.state)
      case 'error':
        return isErrorState(torrent.state)
      case 'completed':
        return isCompletedState(torrent)
      default:
        return true
    }
  })
})

const sortedTorrents = computed(() => {
  const list = [...filteredTorrents.value]

  list.sort((a, b) => {
    const direction = sortDirection.value === 'asc' ? 1 : -1

    const aValue = getSortValue(a, sortBy.value)
    const bValue = getSortValue(b, sortBy.value)

    if (typeof aValue === 'string' || typeof bValue === 'string') {
      return String(aValue).localeCompare(String(bValue), 'uk', { sensitivity: 'base' }) * direction
    }

    if (aValue === bValue) return 0
    return (aValue > bValue ? 1 : -1) * direction
  })

  return list
})

function blurSearch() {
  const input =
    searchInputRef.value?.$el?.querySelector('input') ||
    searchInputRef.value?.input ||
    searchInputRef.value

  input?.blur?.()
}

function toggleRowActionMenu(event, torrent) {
  blurSearch()
  rowActionTorrent.value = torrent
  rowActionMenu.value?.toggle(event)
}

function activityScore(torrent) {
  let score = Number(torrent.dlSpeed || 0) + Number(torrent.upSpeed || 0)

  if (isDownloadingState(torrent.state)) score += 10_000_000_000
  if (['uploading', 'forcedUP'].includes(torrent.state)) score += 8_000_000_000
  if (torrent.state === 'stalledUP') score += 2_000_000_000
  if (isPausedState(torrent.state)) score -= 1_000_000_000
  if (isErrorState(torrent.state)) score += 5_000_000_000

  return score
}

function getSortValue(torrent, key) {
  switch (key) {
    case 'name':
      return torrent.name || ''
    case 'progress':
      return Number(torrent.progress || 0)
    case 'addedOn':
      return Number(torrent.addedOn || 0)
    case 'size':
      return Number(torrent.totalSize || torrent.size || 0)
    case 'downloadSpeed':
      return Number(torrent.dlSpeed || 0)
    case 'uploadSpeed':
      return Number(torrent.upSpeed || 0)
    case 'eta': {
      const eta = Number(torrent.eta || 0)
      return eta <= 0 || eta === 8640000 ? Number.MAX_SAFE_INTEGER : eta
    }
    case 'ratio':
      return Number(torrent.ratio || 0)
    case 'seeds':
      return Number(torrent.numSeeds || 0)
    case 'activity':
    default:
      return activityScore(torrent)
  }
}

function isDownloadingState(state) {
  return ['downloading', 'metaDL', 'forcedDL'].includes(state)
}

function isUploadingState(state) {
  return ['uploading', 'forcedUP', 'stalledUP'].includes(state)
}

function isPausedState(state) {
  return ['pausedDL', 'pausedUP'].includes(state)
}

function isQueuedState(state) {
  return ['queuedDL', 'queuedUP', 'stalledDL'].includes(state)
}

function isCheckingState(state) {
  return ['checkingUP', 'checkingDL', 'checkingResumeData'].includes(state)
}

function isErrorState(state) {
  return ['error', 'missingFiles'].includes(state)
}

function isCompletedState(torrent) {
  return Number(torrent.progress || 0) >= 1
}

function isExpanded(hash) {
  return expandedTorrentHashes.value.includes(hash)
}

async function toggleExpanded(torrent) {
  const hash = torrent.hash

  if (isExpanded(hash)) {
    expandedTorrentHashes.value = expandedTorrentHashes.value.filter((item) => item !== hash)
    return
  }

  expandedTorrentHashes.value = [...expandedTorrentHashes.value, hash]

  if (!peersByHash.value[hash]) {
    await loadTorrentPeers(hash)
  }
}

async function loadTorrentPeers(hash, { silent = false } = {}) {
  if (!silent) {
    peersLoadingByHash.value = {
      ...peersLoadingByHash.value,
      [hash]: true
    }
  }

  peersErrorByHash.value = {
    ...peersErrorByHash.value,
    [hash]: ''
  }

  try {
    const peers = await fetchTorrentPeers(hash)
    peersByHash.value = {
      ...peersByHash.value,
      [hash]: peers
    }
  } catch (e) {
    peersErrorByHash.value = {
      ...peersErrorByHash.value,
      [hash]: e.message || 'Не вдалося завантажити пірів'
    }
  } finally {
    peersLoadingByHash.value = {
      ...peersLoadingByHash.value,
      [hash]: false
    }
  }
}

async function refreshExpandedPeers() {
  const hashes = [...expandedTorrentHashes.value]
  if (!hashes.length) return

  await Promise.allSettled(
    hashes.map((hash) => loadTorrentPeers(hash, { silent: true }))
  )
}

async function loadTorrents({ silent = false } = {}) {
  if (silent) {
    refreshing.value = true
  } else {
    loading.value = true
  }

  error.value = ''

  try {
    const data = await fetchTorrents()
    torrents.value = data
    syncSelectedTorrents()
  } catch (e) {
    error.value = e.message || 'Не вдалося завантажити торенти'
  } finally {
    if (silent) {
      refreshing.value = false
    } else {
      loading.value = false
    }
  }
}

function syncSelectedTorrents() {
  const selectedHashes = new Set(selectedTorrents.value.map((torrent) => torrent.hash))
  selectedTorrents.value = torrents.value.filter((torrent) => selectedHashes.has(torrent.hash))
}

function startAutoRefresh() {
  stopAutoRefresh()

  refreshIntervalId.value = setInterval(() => {
    loadTorrents({ silent: true })
  }, 10000)

  peersRefreshIntervalId.value = setInterval(() => {
    refreshExpandedPeers()
  }, 12000)
}

function stopAutoRefresh() {
  if (refreshIntervalId.value) {
    clearInterval(refreshIntervalId.value)
    refreshIntervalId.value = null
  }

  if (peersRefreshIntervalId.value) {
    clearInterval(peersRefreshIntervalId.value)
    peersRefreshIntervalId.value = null
  }
}

function getSelectedHashes() {
  return selectedTorrents.value.map((torrent) => torrent.hash)
}

async function handlePauseSelected() {
  blurSearch()
  const hashes = getSelectedHashes()
  if (!hashes.length) return

  actionLoading.value = true
  try {
    await pauseTorrents(hashes)
    selectedTorrents.value = []
    await loadTorrents()
  } catch (e) {
    error.value = e.message || 'Не вдалося поставити торенти на паузу'
  } finally {
    actionLoading.value = false
  }
}

async function handleResumeSelected() {
  blurSearch()
  const hashes = getSelectedHashes()
  if (!hashes.length) return

  actionLoading.value = true
  try {
    await resumeTorrents(hashes)
    selectedTorrents.value = []
    await loadTorrents()
  } catch (e) {
    error.value = e.message || 'Не вдалося продовжити торенти'
  } finally {
    actionLoading.value = false
  }
}

function openDeleteDialog(hashes) {
  blurSearch()
  deleteTargetHashes.value = hashes
  deleteFiles.value = false
  deleteDialogVisible.value = true
}

function askDeleteSelected() {
  blurSearch()
  const hashes = getSelectedHashes()
  if (!hashes.length) return
  openDeleteDialog(hashes)
}

function askDeleteOne(torrent) {
  blurSearch()
  openDeleteDialog([torrent.hash])
}

async function confirmDelete() {
  blurSearch()
  if (!deleteTargetHashes.value.length) return

  actionLoading.value = true
  try {
    await deleteTorrents(deleteTargetHashes.value, deleteFiles.value)
    deleteDialogVisible.value = false
    deleteTargetHashes.value = []
    deleteFiles.value = false
    selectedTorrents.value = []
    await loadTorrents()
  } catch (e) {
    error.value = e.message || 'Не вдалося видалити торенти'
  } finally {
    actionLoading.value = false
  }
}

async function pauseOne(torrent) {
  blurSearch()
  actionLoading.value = true
  try {
    await pauseTorrents([torrent.hash])
    await loadTorrents()
  } catch (e) {
    error.value = e.message || 'Не вдалося поставити торент на паузу'
  } finally {
    actionLoading.value = false
  }
}

async function resumeOne(torrent) {
  blurSearch()
  actionLoading.value = true
  try {
    await resumeTorrents([torrent.hash])
    await loadTorrents()
  } catch (e) {
    error.value = e.message || 'Не вдалося продовжити торент'
  } finally {
    actionLoading.value = false
  }
}

function toggleTorrentSelection(torrent) {
  const exists = selectedTorrents.value.some((item) => item.hash === torrent.hash)

  if (exists) {
    selectedTorrents.value = selectedTorrents.value.filter((item) => item.hash !== torrent.hash)
    return
  }

  selectedTorrents.value = [...selectedTorrents.value, torrent]
}

function isSelected(torrent) {
  return selectedTorrents.value.some((item) => item.hash === torrent.hash)
}

function formatBytes(bytes) {
  const value = Number(bytes || 0)
  if (value < 1024) return `${value} B`

  const units = ['KiB', 'MiB', 'GiB', 'TiB']
  let size = value
  let unitIndex = -1

  do {
    size /= 1024
    unitIndex++
  } while (size >= 1024 && unitIndex < units.length - 1)

  return `${size.toFixed(2)} ${units[unitIndex]}`
}

function formatSpeed(bytes) {
  const numeric = Number(bytes || 0)
  if (numeric <= 0) return '0 B/s'
  return `${formatBytes(numeric)}/s`
}

function formatPercent(progress) {
  return `${((Number(progress || 0)) * 100).toFixed(1)}%`
}

function formatEta(seconds) {
  const value = Number(seconds || 0)
  if (value <= 0 || value === 8640000) return '—'

  const h = Math.floor(value / 3600)
  const m = Math.floor((value % 3600) / 60)

  if (h > 0) return `${h}г ${m}хв`
  return `${m}хв`
}

function formatRatio(value) {
  return Number(value || 0).toFixed(2)
}

function formatPeerPercent(progress) {
  return `${((Number(progress || 0)) * 100).toFixed(0)}%`
}

function peerConnectionLabel(value) {
  const map = {
    'µTP': 'uTP',
    BT: 'TCP'
  }

  return map[value] || value || '—'
}

function formatDate(ts) {
  const value = Number(ts || 0)
  if (!value) return '—'
  return new Date(value * 1000).toLocaleString('uk-UA', {
    day: '2-digit',
    month: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

function compactPath(path) {
  if (!path) return '—'
  return path.length > 56 ? `…${path.slice(-56)}` : path
}

function peersCount(hash) {
  return peersByHash.value[hash]?.length || 0
}

function stateLabel(state) {
  const map = {
    downloading: 'Завантаження',
    pausedDL: 'Пауза',
    pausedUP: 'Пауза',
    uploading: 'Роздача',
    stalledUP: 'Очікує пірів',
    stalledDL: 'Очікує',
    queuedDL: 'У черзі',
    queuedUP: 'У черзі',
    checkingUP: 'Перевірка',
    checkingDL: 'Перевірка',
    checkingResumeData: 'Перевірка',
    forcedDL: 'Примусово',
    forcedUP: 'Примусово',
    metaDL: 'Метадані',
    error: 'Помилка',
    missingFiles: 'Немає файлів'
  }

  return map[state] || state || '—'
}

function stateIcon(state) {
  if (['downloading', 'forcedDL', 'metaDL'].includes(state)) return 'pi pi-download'
  if (['uploading', 'forcedUP'].includes(state)) return 'pi pi-upload'
  if (['stalledUP'].includes(state)) return 'pi pi-share-alt'
  if (['pausedDL', 'pausedUP'].includes(state)) return 'pi pi-pause-circle'
  if (['queuedDL', 'queuedUP', 'stalledDL'].includes(state)) return 'pi pi-clock'
  if (['checkingUP', 'checkingDL', 'checkingResumeData'].includes(state)) return 'pi pi-spin pi-spinner'
  if (['error', 'missingFiles'].includes(state)) return 'pi pi-exclamation-triangle'
  return 'pi pi-circle'
}

function statePillClass(state) {
  if (['downloading', 'forcedDL', 'metaDL'].includes(state)) return 'torrent-state-info'
  if (['uploading', 'forcedUP'].includes(state)) return 'torrent-state-success'
  if (['stalledUP'].includes(state)) return 'torrent-state-seeding'
  if (['pausedDL', 'pausedUP'].includes(state)) return 'torrent-state-warn'
  if (['queuedDL', 'queuedUP', 'stalledDL'].includes(state)) return 'torrent-state-muted'
  if (['checkingUP', 'checkingDL', 'checkingResumeData'].includes(state)) return 'torrent-state-checking'
  if (['error', 'missingFiles'].includes(state)) return 'torrent-state-danger'
  return 'torrent-state-muted'
}

function torrentHealthLabel(torrent) {
  if (isErrorState(torrent.state)) return 'Проблема'
  if (isDownloadingState(torrent.state)) return 'Активне завантаження'
  if (['uploading', 'forcedUP'].includes(torrent.state)) return 'Активна роздача'
  if (torrent.state === 'stalledUP') return 'Готовий до роздачі'
  if (isPausedState(torrent.state)) return 'На паузі'
  if (isQueuedState(torrent.state)) return 'У черзі'
  if (isCheckingState(torrent.state)) return 'Перевіряється'
  if (isCompletedState(torrent)) return 'Готовий'
  return 'Невідомо'
}

function torrentHealthClass(torrent) {
  if (isErrorState(torrent.state)) return 'border-red-500/20 bg-red-500/10 text-red-300'
  if (isDownloadingState(torrent.state)) return 'border-sky-500/20 bg-sky-500/10 text-sky-300'
  if (['uploading', 'forcedUP'].includes(torrent.state)) return 'border-emerald-500/20 bg-emerald-500/10 text-emerald-300'
  if (torrent.state === 'stalledUP') return 'border-indigo-500/20 bg-indigo-500/10 text-indigo-300'
  if (isPausedState(torrent.state)) return 'border-amber-500/20 bg-amber-500/10 text-amber-300'
  if (isCompletedState(torrent)) return 'border-violet-500/20 bg-violet-500/10 text-violet-300'
  return 'border-white/10 bg-white/5 text-white/70'
}

watch(
  () => props.active,
  async (isActive) => {
    if (isActive) {
      await loadTorrents()
      startAutoRefresh()
      return
    }

    stopAutoRefresh()
  },
  { immediate: true }
)

onBeforeUnmount(() => {
  stopAutoRefresh()
})
</script>

<template>
  <section class="space-y-3 relative">
    <div
      v-if="refreshing"
      class="pointer-events-none absolute right-0 top-0 z-20"
    >
      <div class="refresh-badge">
        <span class="refresh-dot" />
        Оновлення
      </div>
    </div>

    <div class="bg-panel rounded-2xl border border-white/10 p-4 shadow-custom">
      <div class="flex flex-col gap-3">
        <div class="flex items-start justify-between gap-3">
          <div>
            <div class="text-[10px] uppercase tracking-wide text-white/60 mb-1">
              qBittorrent
            </div>
            <div class="text-sm text-white/50">
              Керуйте торентами з телефону
            </div>
          </div>

          <Button
            icon="pi pi-refresh"
            text
            rounded
            :loading="loading || refreshing"
            @click="loadTorrents"
          />
        </div>

        <div class="grid grid-cols-3 gap-2">
          <div class="rounded-xl border border-white/10 bg-white/5 px-3 py-2">
            <div class="text-[10px] uppercase tracking-wide text-white/40">Усі</div>
            <div class="mt-1 text-lg font-semibold text-white">{{ stats.total }}</div>
          </div>

          <div class="rounded-xl border border-emerald-500/15 bg-emerald-500/10 px-3 py-2">
            <div class="text-[10px] uppercase tracking-wide text-emerald-200/70">Активні</div>
            <div class="mt-1 text-lg font-semibold text-emerald-200">{{ stats.downloading + stats.uploading }}</div>
          </div>

          <div class="rounded-xl border border-amber-500/15 bg-amber-500/10 px-3 py-2">
            <div class="text-[10px] uppercase tracking-wide text-amber-200/70">Пауза</div>
            <div class="mt-1 text-lg font-semibold text-amber-200">{{ stats.paused }}</div>
          </div>
        </div>

        <div class="flex items-center gap-2">
          <InputText
            ref="searchInputRef"
            v-model="search"
            placeholder="Пошук торентів"
            size="small"
            fluid
            @keydown.enter="blurSearch"
          />

          <Button
            icon="pi pi-check"
            text
            rounded
            aria-label="Готово"
            @click="blurSearch"
          />
        </div>

        <div class="flex gap-2 overflow-x-auto no-scrollbar pb-1">
          <button
            v-for="option in filterOptions"
            :key="option.value"
            type="button"
            class="shrink-0 rounded-full border px-3 py-1.5 text-xs transition"
            :class="stateFilter === option.value
              ? 'border-sky-400/30 bg-sky-400/15 text-sky-200'
              : 'border-white/10 bg-white/5 text-white/65'"
            @click="stateFilter = option.value"
          >
            {{ option.label }}
          </button>
        </div>

        <div class="grid grid-cols-2 gap-2">
          <div class="rounded-xl border border-white/10 bg-white/5 px-3 py-2">
            <label class="block text-[10px] uppercase tracking-wide text-white/40 mb-1">
              Сортування
            </label>
            <select
              v-model="sortBy"
              class="w-full bg-transparent text-sm text-white outline-none"
            >
              <option
                v-for="option in sortOptions"
                :key="option.value"
                :value="option.value"
                class="bg-slate-900 text-white"
              >
                {{ option.label }}
              </option>
            </select>
          </div>

          <div class="rounded-xl border border-white/10 bg-white/5 px-3 py-2">
            <label class="block text-[10px] uppercase tracking-wide text-white/40 mb-1">
              Порядок
            </label>
            <select
              v-model="sortDirection"
              class="w-full bg-transparent text-sm text-white outline-none"
            >
              <option value="desc" class="bg-slate-900 text-white">Спадання</option>
              <option value="asc" class="bg-slate-900 text-white">Зростання</option>
            </select>
          </div>
        </div>
      </div>
    </div>

    <div
      v-if="selectedTorrents.length"
      class="bg-panel rounded-2xl border border-white/10 p-3 shadow-custom"
    >
      <div class="flex flex-wrap items-center gap-2">
        <div class="text-xs text-white/50 mr-1">
          Вибрано: {{ selectedCount }}
        </div>

        <Button
          icon="pi pi-pause"
          size="small"
          severity="warn"
          :disabled="actionLoading"
          @click="handlePauseSelected"
        />

        <Button
          icon="pi pi-play"
          size="small"
          severity="success"
          :disabled="actionLoading"
          @click="handleResumeSelected"
        />

        <Button
          icon="pi pi-trash"
          size="small"
          severity="danger"
          :disabled="actionLoading"
          @click="askDeleteSelected"
        />
      </div>
    </div>

    <div
      v-if="error"
      class="bg-panel rounded-2xl border border-red-500/20 p-4 text-red-300 shadow-custom"
    >
      {{ error }}
    </div>

    <Menu
      ref="rowActionMenu"
      :model="rowActionItems"
      popup
    />

    <div
      v-if="loading && !torrents.length"
      class="space-y-3"
    >
      <div
        v-for="n in 3"
        :key="n"
        class="bg-panel rounded-2xl border border-white/10 p-4 shadow-custom animate-pulse"
      >
        <div class="h-4 w-2/3 rounded bg-white/10" />
        <div class="mt-3 h-2 rounded bg-white/10" />
        <div class="mt-3 grid grid-cols-2 gap-2">
          <div class="h-10 rounded-xl bg-white/5" />
          <div class="h-10 rounded-xl bg-white/5" />
        </div>
      </div>
    </div>

    <div v-else class="space-y-3">
      <div
        v-for="torrent in sortedTorrents"
        :key="torrent.hash"
        class="bg-panel rounded-2xl border border-white/10 shadow-custom overflow-hidden transition"
        :class="isExpanded(torrent.hash) ? 'border-sky-400/20' : ''"
      >
        <div class="p-3 sm:p-4">
          <div class="flex items-start gap-3">
            <div @click.stop>
              <Checkbox
                :modelValue="isSelected(torrent)"
                binary
                @update:modelValue="toggleTorrentSelection(torrent)"
              />
            </div>

            <div class="min-w-0 flex-1">
              <div class="flex items-start justify-between gap-3">
                <div class="min-w-0">
                  <div class="font-semibold text-white leading-snug break-words">
                    {{ torrent.name }}
                  </div>

                  <div class="mt-2 flex flex-wrap gap-2">
                    <div
                      class="torrent-state-pill"
                      :class="statePillClass(torrent.state)"
                    >
                      <i :class="stateIcon(torrent.state)" />
                      <span>{{ stateLabel(torrent.state) }}</span>
                    </div>

                    <Tag
                      v-if="torrent.category"
                      :value="torrent.category"
                      severity="secondary"
                    />

                    <Tag
                      v-if="torrent.numSeeds > 0 && (['uploading', 'forcedUP'].includes(torrent.state) || torrent.state === 'stalledUP')"
                      :value="`Seeds ${torrent.numSeeds}`"
                      severity="contrast"
                    />
                  </div>
                </div>

                <div class="flex items-center gap-1 shrink-0">
                  <Button
                    icon="pi pi-ellipsis-v"
                    size="small"
                    text
                    rounded
                    aria-label="Дії"
                    :disabled="actionLoading"
                    @click="toggleRowActionMenu($event, torrent)"
                  />

                  <Button
                    :icon="isExpanded(torrent.hash) ? 'pi pi-chevron-up' : 'pi pi-chevron-down'"
                    size="small"
                    text
                    rounded
                    aria-label="Розгорнути"
                    @click="toggleExpanded(torrent)"
                  />
                </div>
              </div>

              <div class="mt-3">
                <div class="flex items-center justify-between gap-3 text-xs text-white/50 mb-1">
                  <span>{{ formatPercent(torrent.progress) }}</span>
                  <span>{{ formatBytes(torrent.downloaded) }} / {{ formatBytes(torrent.totalSize || torrent.size) }}</span>
                </div>

                <ProgressBar
                  :value="Number((torrent.progress || 0) * 100)"
                  :showValue="false"
                  style="height: 8px"
                />
              </div>

              <div class="mt-2.5 grid grid-cols-2 gap-2">
                <div class="rounded-xl border border-white/10 bg-white/5 px-3 py-1.5">
                  <div class="text-[10px] uppercase tracking-wide text-white/40">Швидкість</div>
                  <div class="mt-1 text-[12px] sm:text-sm text-white whitespace-nowrap">
                    ↓ {{ formatSpeed(torrent.dlSpeed) }} · ↑ {{ formatSpeed(torrent.upSpeed) }}
                  </div>
                </div>

                <div class="rounded-xl border border-white/10 bg-white/5 px-3 py-1.5">
                  <div class="text-[10px] uppercase tracking-wide text-white/40">ETA</div>
                  <div class="mt-1 text-[12px] sm:text-sm text-white whitespace-nowrap">
                    {{ formatEta(torrent.eta) }}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <Transition name="torrent-expand">
          <div
            v-if="isExpanded(torrent.hash)"
            class="border-t border-white/10 bg-black/10 px-3 pb-3 pt-3 sm:px-4 sm:pb-4"
          >
            <div class="rounded-2xl border border-white/10 bg-white/[0.03] p-3 sm:p-4 space-y-4">
              <div class="flex items-center justify-between gap-3">
                <div>
                  <div class="text-[10px] uppercase tracking-wide text-white/45">
                    Деталі торенту
                  </div>
                  <div class="text-sm text-white/60 mt-1">
                    Додаткова інформація та піри
                  </div>
                </div>

                <div
                  class="shrink-0 rounded-full border px-3 py-1.5 text-xs font-medium"
                  :class="torrentHealthClass(torrent)"
                >
                  {{ torrentHealthLabel(torrent) }}
                </div>
              </div>

              <div class="grid grid-cols-2 gap-2">
                <div class="rounded-xl border border-white/10 bg-white/5 px-3 py-2">
                  <div class="text-[10px] uppercase tracking-wide text-white/40">
                    Seeds / Leechs
                  </div>
                  <div class="mt-1 text-sm font-medium text-white">
                    {{ torrent.numSeeds ?? 0 }} / {{ torrent.numLeechs ?? 0 }}
                  </div>
                </div>

                <div class="rounded-xl border border-white/10 bg-white/5 px-3 py-2">
                  <div class="text-[10px] uppercase tracking-wide text-white/40">
                    Ratio
                  </div>
                  <div class="mt-1 text-sm font-medium text-white">
                    {{ formatRatio(torrent.ratio) }}
                  </div>
                </div>

                <div class="rounded-xl border border-white/10 bg-white/5 px-3 py-2">
                  <div class="text-[10px] uppercase tracking-wide text-white/40">
                    Розмір
                  </div>
                  <div class="mt-1 text-sm font-medium text-white">
                    {{ formatBytes(torrent.totalSize || torrent.size) }}
                  </div>
                </div>

                <div class="rounded-xl border border-white/10 bg-white/5 px-3 py-2">
                  <div class="text-[10px] uppercase tracking-wide text-white/40">
                    Пірів зараз
                  </div>
                  <div class="mt-1 text-sm font-medium text-white">
                    {{ peersCount(torrent.hash) }}
                  </div>
                </div>

                <div class="rounded-xl border border-white/10 bg-white/5 px-3 py-2">
                  <div class="text-[10px] uppercase tracking-wide text-white/40">
                    Додано
                  </div>
                  <div class="mt-1 text-sm font-medium text-white">
                    {{ formatDate(torrent.addedOn) }}
                  </div>
                </div>

                <div class="rounded-xl border border-white/10 bg-white/5 px-3 py-2">
                  <div class="text-[10px] uppercase tracking-wide text-white/40">
                    Завершено
                  </div>
                  <div class="mt-1 text-sm font-medium text-white">
                    {{ formatDate(torrent.completedOn) }}
                  </div>
                </div>
              </div>

              <div class="space-y-2">
                <div class="text-[10px] uppercase tracking-wide text-white/45">
                  Шляхи
                </div>

                <div class="rounded-xl border border-white/10 bg-white/5 px-3 py-3">
                  <div class="text-[10px] uppercase tracking-wide text-white/35">
                    Save path
                  </div>
                  <div class="mt-1 text-sm text-white break-all">
                    {{ compactPath(torrent.savePath) }}
                  </div>
                </div>

                <div
                  v-if="torrent.contentPath"
                  class="rounded-xl border border-white/10 bg-white/5 px-3 py-3"
                >
                  <div class="text-[10px] uppercase tracking-wide text-white/35">
                    Content path
                  </div>
                  <div class="mt-1 text-sm text-white break-all">
                    {{ compactPath(torrent.contentPath) }}
                  </div>
                </div>
              </div>

              <div class="rounded-2xl border border-white/10 bg-black/10 p-3">
                <div class="flex items-center justify-between gap-3 mb-3">
                  <div>
                    <div class="text-[10px] uppercase tracking-wide text-white/45">
                      Піри
                    </div>
                    <div class="text-xs text-white/50 mt-1">
                      Підключені клієнти для цього торенту
                    </div>
                  </div>

                  <Button
                    icon="pi pi-refresh"
                    size="small"
                    text
                    rounded
                    :loading="!!peersLoadingByHash[torrent.hash]"
                    @click="loadTorrentPeers(torrent.hash)"
                  />
                </div>

                <div
                  v-if="peersErrorByHash[torrent.hash]"
                  class="rounded-xl border border-red-500/20 bg-red-500/10 px-3 py-3 text-sm text-red-300"
                >
                  {{ peersErrorByHash[torrent.hash] }}
                </div>

                <div
                  v-else-if="peersLoadingByHash[torrent.hash] && !peersByHash[torrent.hash]?.length"
                  class="rounded-xl border border-white/10 bg-white/5 px-3 py-3 text-sm text-white/50"
                >
                  Завантаження пірів...
                </div>

                <div
                  v-else-if="!peersByHash[torrent.hash]?.length"
                  class="rounded-xl border border-white/10 bg-white/5 px-3 py-3 text-sm text-white/50"
                >
                  Немає активних пірів
                </div>

                <div
                  v-else
                  class="space-y-2 max-h-[420px] overflow-y-auto pr-1"
                >
                  <div
                    v-for="peer in peersByHash[torrent.hash]"
                    :key="`${torrent.hash}-${peer.ip}-${peer.port}`"
                    class="rounded-xl border border-white/10 bg-white/[0.04] p-3"
                  >
                    <div class="flex items-start justify-between gap-3">
                      <div class="min-w-0">
                        <div class="text-sm font-medium text-white break-all">
                          {{ peer.ip }}:{{ peer.port }}
                        </div>

                        <div class="mt-2 flex flex-wrap gap-2">
                          <Tag
                            v-if="peer.country"
                            :value="peer.country"
                            severity="secondary"
                          />
                          <Tag
                            :value="peerConnectionLabel(peer.connection)"
                            severity="info"
                          />
                          <Tag
                            v-if="peer.client"
                            :value="peer.client"
                            severity="contrast"
                          />
                        </div>
                      </div>

                      <div class="shrink-0 text-xs text-white/45">
                        {{ formatPeerPercent(peer.progress) }}
                      </div>
                    </div>

                    <div class="mt-3 grid grid-cols-2 gap-2">
                      <div class="rounded-lg bg-black/10 px-2.5 py-2">
                        <div class="text-[10px] uppercase tracking-wide text-white/35">
                          ↓ Download
                        </div>
                        <div class="mt-1 text-[13px] text-white">
                          {{ formatSpeed(peer.dlRate) }}
                        </div>
                      </div>

                      <div class="rounded-lg bg-black/10 px-2.5 py-2">
                        <div class="text-[10px] uppercase tracking-wide text-white/35">
                          ↑ Upload
                        </div>
                        <div class="mt-1 text-[13px] text-white">
                          {{ formatSpeed(peer.ulRate) }}
                        </div>
                      </div>

                      <div class="rounded-lg bg-black/10 px-2.5 py-2">
                        <div class="text-[10px] uppercase tracking-wide text-white/35">
                          Отримано
                        </div>
                        <div class="mt-1 text-[13px] text-white">
                          {{ formatBytes(peer.downloaded) }}
                        </div>
                      </div>

                      <div class="rounded-lg bg-black/10 px-2.5 py-2">
                        <div class="text-[10px] uppercase tracking-wide text-white/35">
                          Віддано
                        </div>
                        <div class="mt-1 text-[13px] text-white">
                          {{ formatBytes(peer.uploaded) }}
                        </div>
                      </div>
                    </div>

                    <div class="mt-2 flex flex-wrap gap-x-3 gap-y-1 text-[11px] text-white/45">
                      <span>Relevance: {{ formatRatio(peer.relevance) }}</span>
                      <span>Flags: {{ peer.flags || '—' }}</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </Transition>
      </div>

      <div
        v-if="!loading && !sortedTorrents.length"
        class="bg-panel rounded-2xl border border-white/10 p-4 shadow-custom text-sm text-white/50"
      >
        Торентів не знайдено
      </div>
    </div>

    <Dialog
      v-model:visible="deleteDialogVisible"
      modal
      header="Видалення торентів"
      :style="{ width: '28rem' }"
    >
      <div class="space-y-4">
        <p class="text-sm text-white/70">
          Точно видалити вибрані торенти?
        </p>

        <div class="flex items-center gap-2">
          <Checkbox v-model="deleteFiles" binary inputId="deleteFiles" />
          <label for="deleteFiles" class="text-sm text-white/80">
            Також видалити завантажені файли
          </label>
        </div>

        <div class="flex justify-end gap-2">
          <Button
            label="Скасувати"
            icon="pi pi-times"
            text
            size="small"
            @click="deleteDialogVisible = false"
          />
          <Button
            label="Видалити"
            icon="pi pi-trash"
            severity="danger"
            size="small"
            :loading="actionLoading"
            @click="confirmDelete"
          />
        </div>
      </div>
    </Dialog>
  </section>
</template>

<style scoped>
.no-scrollbar::-webkit-scrollbar {
  display: none;
}

.no-scrollbar {
  -ms-overflow-style: none;
  scrollbar-width: none;
}

.refresh-badge {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: 9999px;
  border: 1px solid rgba(34, 211, 238, 0.18);
  background: rgba(8, 47, 73, 0.72);
  backdrop-filter: blur(10px);
  color: rgba(186, 230, 253, 0.95);
  font-size: 12px;
  line-height: 1;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.18);
}

.refresh-dot {
  width: 8px;
  height: 8px;
  border-radius: 9999px;
  background: #22d3ee;
  box-shadow: 0 0 0 0 rgba(34, 211, 238, 0.7);
  animation: qb-pulse 1.4s infinite;
}

.torrent-state-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  border-radius: 9999px;
  border-width: 1px;
  padding: 6px 10px;
  font-size: 12px;
  line-height: 1;
  font-weight: 500;
}

.torrent-state-info {
  border-color: rgba(56, 189, 248, 0.25);
  background: rgba(14, 165, 233, 0.12);
  color: rgb(186 230 253);
}

.torrent-state-success {
  border-color: rgba(16, 185, 129, 0.22);
  background: rgba(16, 185, 129, 0.12);
  color: rgb(167 243 208);
}

.torrent-state-seeding {
  border-color: rgba(129, 140, 248, 0.24);
  background: rgba(99, 102, 241, 0.12);
  color: rgb(199 210 254);
}

.torrent-state-checking {
  border-color: rgba(56, 189, 248, 0.18);
  background: rgba(148, 163, 184, 0.12);
  color: rgb(226 232 240);
}

.torrent-state-warn {
  border-color: rgba(245, 158, 11, 0.22);
  background: rgba(245, 158, 11, 0.12);
  color: rgb(253 230 138);
}

.torrent-state-muted {
  border-color: rgba(255, 255, 255, 0.1);
  background: rgba(255, 255, 255, 0.05);
  color: rgba(255, 255, 255, 0.72);
}

.torrent-state-danger {
  border-color: rgba(239, 68, 68, 0.22);
  background: rgba(239, 68, 68, 0.12);
  color: rgb(252 165 165);
}

.torrent-expand-enter-active,
.torrent-expand-leave-active {
  transition: all 0.2s ease;
}

.torrent-expand-enter-from,
.torrent-expand-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}

@keyframes qb-pulse {
  0% {
    box-shadow: 0 0 0 0 rgba(34, 211, 238, 0.7);
  }
  70% {
    box-shadow: 0 0 0 8px rgba(34, 211, 238, 0);
  }
  100% {
    box-shadow: 0 0 0 0 rgba(34, 211, 238, 0);
  }
}
</style>