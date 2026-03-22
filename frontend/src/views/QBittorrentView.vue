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
const error = ref('')
const selectedTorrents = ref([])
const search = ref('')

const deleteDialogVisible = ref(false)
const deleteFiles = ref(false)
const deleteTargetHashes = ref([])

const refreshIntervalId = ref(null)
const actionLoading = ref(false)

const rowActionMenu = ref()
const rowActionTorrent = ref(null)

const filteredTorrents = computed(() => {
  const term = search.value.trim().toLowerCase()
  if (!term) return torrents.value

  return torrents.value.filter((torrent) => {
    return [
      torrent.name,
      torrent.state,
      torrent.category,
      torrent.savePath
    ].some((value) => String(value || '').toLowerCase().includes(term))
  })
})

const selectedCount = computed(() => selectedTorrents.value.length)

const rowActionItems = computed(() => {
  if (!rowActionTorrent.value) {
    return []
  }

  const items = []

  if (isPausedState(rowActionTorrent.value.state)) {
    items.push({
      label: 'Продовжити',
      icon: 'pi pi-play !text-emerald-500',
      command: async () => {
        if (rowActionTorrent.value) {
          await resumeOne(rowActionTorrent.value)
        }
      }
    })
  } else {
    items.push({
      label: 'Пауза',
      icon: 'pi pi-pause !text-amber-500',
      command: async () => {
        if (rowActionTorrent.value) {
          await pauseOne(rowActionTorrent.value)
        }
      }
    })
  }

  items.push({
    label: 'Видалити',
    icon: 'pi pi-trash !text-red-500',
    command: () => {
      if (rowActionTorrent.value) {
        askDeleteOne(rowActionTorrent.value)
      }
    }
  })

  return [
    {
        label: 'Дії',
        items: items
    }
  ]
})

function toggleRowActionMenu(event, torrent) {
  rowActionTorrent.value = torrent
  rowActionMenu.value?.toggle(event)
}

function isPausedState(state) {
  return ['pausedDL', 'pausedUP'].includes(state)
}

async function loadTorrents({ silent = false } = {}) {
  if (!silent) {
    loading.value = true
  }

  error.value = ''

  try {
    torrents.value = await fetchTorrents()
  } catch (e) {
    error.value = e.message || 'Не вдалося завантажити торенти'
  } finally {
    if (!silent) {
      loading.value = false
    }
  }
}

function startAutoRefresh() {
  stopAutoRefresh()

  refreshIntervalId.value = setInterval(() => {
    loadTorrents({ silent: true })
  }, 10000)
}

function stopAutoRefresh() {
  if (refreshIntervalId.value) {
    clearInterval(refreshIntervalId.value)
    refreshIntervalId.value = null
  }
}

function getSelectedHashes() {
  return selectedTorrents.value.map((torrent) => torrent.hash)
}

async function handlePauseSelected() {
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
  deleteTargetHashes.value = hashes
  deleteFiles.value = false
  deleteDialogVisible.value = true
}

function askDeleteSelected() {
  const hashes = getSelectedHashes()
  if (!hashes.length) return
  openDeleteDialog(hashes)
}

function askDeleteOne(torrent) {
  openDeleteDialog([torrent.hash])
}

async function confirmDelete() {
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

function stateLabel(state) {
  const map = {
    downloading: 'Завантаження',
    pausedDL: 'Пауза',
    pausedUP: 'Пауза',
    uploading: 'Роздача',
    stalledUP: 'Сідування',
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

function stateSeverity(state) {
  if (['downloading', 'forcedDL', 'metaDL'].includes(state)) return 'info'
  if (['uploading', 'forcedUP'].includes(state)) return 'success'
  if (['stalledUP'].includes(state)) return 'contrast'
  if (['pausedDL', 'pausedUP'].includes(state)) return 'warn'
  if (['queuedDL', 'queuedUP', 'stalledDL'].includes(state)) return 'secondary'
  if (['checkingUP', 'checkingDL', 'checkingResumeData'].includes(state)) return 'info'
  if (['error', 'missingFiles'].includes(state)) return 'danger'
  return 'secondary'
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
  <section class="space-y-3">
    <div class="bg-panel rounded-2xl border border-white/10 p-4 shadow-custom">
      <div class="flex flex-col gap-3">
        <div class="flex items-start justify-between gap-3">
          <div>
            <div class="text-[10px] uppercase tracking-wide text-white/60 mb-1">
              qBittorrent
            </div>
            <div class="text-sm text-white/50">
              Торенти, пауза, продовження, видалення
            </div>
          </div>

          <Button
            icon="pi pi-refresh"
            text
            rounded
            :loading="loading"
            @click="loadTorrents"
          />
        </div>

        <InputText
          v-model="search"
          placeholder="Пошук торентів"
          size="small"
          fluid
        />
      </div>
    </div>

    <div class="bg-panel rounded-2xl border border-white/10 p-3 shadow-custom">
      <div class="flex flex-wrap items-center gap-2">
        <div class="text-xs text-white/50 mr-1">
          Вибрано: {{ selectedCount }}
        </div>

        <Button
          icon="pi pi-pause"
          size="small"
          severity="warn"
          :disabled="!selectedTorrents.length || actionLoading"
          @click="handlePauseSelected"
        />

        <Button
          icon="pi pi-play"
          size="small"
          severity="success"
          :disabled="!selectedTorrents.length || actionLoading"
          @click="handleResumeSelected"
        />

        <Button
          icon="pi pi-trash"
          size="small"
          severity="danger"
          :disabled="!selectedTorrents.length || actionLoading"
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

    <div class="space-y-3">
      <div
        v-for="torrent in filteredTorrents"
        :key="torrent.hash"
        class="bg-panel rounded-2xl border border-white/10 p-3 sm:p-4 shadow-custom"
      >
        <div class="flex items-start gap-3">
          <Checkbox
            :modelValue="isSelected(torrent)"
            binary
            @update:modelValue="toggleTorrentSelection(torrent)"
          />

          <div class="min-w-0 flex-1">
            <div class="flex items-start justify-between gap-3">
              <div class="min-w-0">
                <div class="font-semibold text-white leading-snug break-words">
                  {{ torrent.name }}
                </div>

                <div class="mt-2 flex flex-wrap gap-2">
                  <Tag :value="stateLabel(torrent.state)" :severity="stateSeverity(torrent.state)" />
                  <Tag v-if="torrent.category" :value="torrent.category" severity="secondary" />
                </div>
              </div>

              <div class="shrink-0">
                <Button
                  icon="pi pi-ellipsis-v"
                  size="small"
                  text
                  rounded
                  aria-label="Дії"
                  :disabled="actionLoading"
                  @click="toggleRowActionMenu($event, torrent)"
                />
              </div>
            </div>

            <div class="mt-3">
              <ProgressBar
                :value="Number((torrent.progress || 0) * 100)"
                :showValue="false"
                style="height: 8px"
              />
              <div class="mt-1 text-xs text-white/50">
                {{ formatPercent(torrent.progress) }}
              </div>
            </div>

            <div class="mt-3 grid grid-cols-2 xl:grid-cols-3 gap-x-4 gap-y-2 text-sm">
                <div class="text-white/45">Швидкість</div>
                <div class="text-white text-right xl:text-left xl:col-span-2 whitespace-nowrap text-[13px] sm:text-sm">
                    ↓ {{ formatSpeed(torrent.dlSpeed) }} · ↑ {{ formatSpeed(torrent.upSpeed) }}
                </div>

                <div class="text-white/45">Розмір</div>
                <div class="text-white text-right xl:text-left xl:col-span-2 break-words text-[13px] sm:text-sm">
                    {{ formatBytes(torrent.downloaded) }} / {{ formatBytes(torrent.totalSize || torrent.size) }}
                </div>

                <div class="text-white/45">ETA</div>
                <div class="text-white text-right xl:text-left xl:col-span-2 text-[13px] sm:text-sm">
                    {{ formatEta(torrent.eta) }}
                </div>

                <div class="text-white/45">Seeds / Leechs</div>
                <div class="text-white text-right xl:text-left xl:col-span-2 whitespace-nowrap text-[13px] sm:text-sm">
                    {{ torrent.numSeeds ?? 0 }} / {{ torrent.numLeechs ?? 0 }}
                </div>

                <div class="text-white/45">Ratio</div>
                <div class="text-white text-right xl:text-left xl:col-span-2 whitespace-nowrap text-[13px] sm:text-sm">
                    {{ formatRatio(torrent.ratio) }}
                </div>
            </div>
          </div>
        </div>
      </div>

      <div
        v-if="!loading && !filteredTorrents.length"
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
