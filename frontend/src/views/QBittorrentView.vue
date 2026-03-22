<script setup>
import { computed, onMounted, ref } from 'vue'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Tag from 'primevue/tag'
import Dialog from 'primevue/dialog'
import Checkbox from 'primevue/checkbox'
import ProgressBar from 'primevue/progressbar'
import {
  fetchTorrents,
  pauseTorrents,
  resumeTorrents,
  deleteTorrents
} from '../api/qbittorrent'

const torrents = ref([])
const loading = ref(false)
const error = ref('')
const selectedTorrents = ref([])
const search = ref('')

const deleteDialogVisible = ref(false)
const deleteFiles = ref(false)
const deleteTargetHashes = ref([])

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

async function loadTorrents() {
  loading.value = true
  error.value = ''

  try {
    torrents.value = await fetchTorrents()
  } catch (e) {
    error.value = e.message || 'Не вдалося завантажити торенти'
  } finally {
    loading.value = false
  }
}

function getSelectedHashes() {
  return selectedTorrents.value.map((torrent) => torrent.hash)
}

async function handlePauseSelected() {
  const hashes = getSelectedHashes()
  if (!hashes.length) return

  await pauseTorrents(hashes)
  await loadTorrents()
  selectedTorrents.value = []
}

async function handleResumeSelected() {
  const hashes = getSelectedHashes()
  if (!hashes.length) return

  await resumeTorrents(hashes)
  await loadTorrents()
  selectedTorrents.value = []
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

  await deleteTorrents(deleteTargetHashes.value, deleteFiles.value)
  deleteDialogVisible.value = false
  deleteTargetHashes.value = []
  deleteFiles.value = false
  selectedTorrents.value = []
  await loadTorrents()
}

async function pauseOne(torrent) {
  await pauseTorrents([torrent.hash])
  await loadTorrents()
}

async function resumeOne(torrent) {
  await resumeTorrents([torrent.hash])
  await loadTorrents()
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
  return `${formatBytes(bytes)}/s`
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
    metaDL: 'Отримання метаданих',
    error: 'Помилка',
    missingFiles: 'Немає файлів'
  }

  return map[state] || state || '—'
}

function stateSeverity(state) {
  if (['downloading', 'forcedDL', 'metaDL'].includes(state)) return 'info'
  if (['uploading', 'stalledUP', 'forcedUP'].includes(state)) return 'success'
  if (['pausedDL', 'pausedUP'].includes(state)) return 'warning'
  if (['error', 'missingFiles'].includes(state)) return 'danger'
  return 'secondary'
}

onMounted(loadTorrents)
</script>

<template>
  <section class="space-y-4">
    <div class="bg-panel rounded-2xl border border-white/10 p-4 shadow-custom">
      <div class="flex flex-col lg:flex-row lg:items-center lg:justify-between gap-3">
        <div>
          <div class="text-[10px] uppercase tracking-wide text-white/60 mb-1">
            qBittorrent
          </div>
          <div class="text-sm text-white/50">
            Керування торентами, пауза, продовження, видалення
          </div>
        </div>

        <div class="flex items-center gap-2">
          <InputText
            v-model="search"
            placeholder="Пошук торентів"
            class="w-full lg:w-72"
          />

          <Button
            icon="pi pi-refresh"
            text
            rounded
            @click="loadTorrents"
          />
        </div>
      </div>
    </div>

    <div class="bg-panel rounded-2xl border border-white/10 p-4 shadow-custom">
      <div class="flex flex-wrap gap-2">
        <Button
          label="Пауза вибраних"
          icon="pi pi-pause"
          size="small"
          :disabled="!selectedTorrents.length"
          @click="handlePauseSelected"
        />

        <Button
          label="Продовжити вибрані"
          icon="pi pi-play"
          size="small"
          severity="success"
          :disabled="!selectedTorrents.length"
          @click="handleResumeSelected"
        />

        <Button
          label="Видалити вибрані"
          icon="pi pi-trash"
          size="small"
          severity="danger"
          :disabled="!selectedTorrents.length"
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

    <div class="bg-panel rounded-2xl border border-white/10 p-2 shadow-custom overflow-hidden">
      <DataTable
        v-model:selection="selectedTorrents"
        :value="filteredTorrents"
        dataKey="hash"
        scrollable
        scrollHeight="70vh"
        stripedRows
        :loading="loading"
        class="rounded-xl overflow-hidden"
      >
        <Column selectionMode="multiple" headerStyle="width: 3rem" />

        <Column field="name" header="Торент" style="min-width: 22rem">
          <template #body="{ data }">
            <div class="space-y-2">
              <div class="font-semibold text-white">
                {{ data.name }}
              </div>

              <div class="flex flex-wrap gap-2">
                <Tag :value="stateLabel(data.state)" :severity="stateSeverity(data.state)" />
                <Tag v-if="data.category" :value="data.category" severity="secondary" />
              </div>

              <ProgressBar
                :value="Number((data.progress || 0) * 100)"
                :showValue="false"
                style="height: 8px"
              />

              <div class="text-xs text-white/50">
                {{ formatPercent(data.progress) }}
              </div>
            </div>
          </template>
        </Column>

        <Column header="Швидкість" style="min-width: 12rem">
          <template #body="{ data }">
            <div class="space-y-1 text-sm">
              <div>↓ {{ formatSpeed(data.dlSpeed) }}</div>
              <div>↑ {{ formatSpeed(data.upSpeed) }}</div>
            </div>
          </template>
        </Column>

        <Column header="Розмір" style="min-width: 11rem">
          <template #body="{ data }">
            <div class="space-y-1 text-sm">
              <div>{{ formatBytes(data.downloaded) }} / {{ formatBytes(data.totalSize || data.size) }}</div>
              <div class="text-white/50">ETA: {{ formatEta(data.eta) }}</div>
            </div>
          </template>
        </Column>

        <Column header="Сіди / Лічи" style="min-width: 10rem">
          <template #body="{ data }">
            <div class="space-y-1 text-sm">
              <div>S: {{ data.numSeeds ?? 0 }}</div>
              <div>L: {{ data.numLeechs ?? 0 }}</div>
            </div>
          </template>
        </Column>

        <Column header="Ratio" style="min-width: 8rem">
          <template #body="{ data }">
            {{ Number(data.ratio || 0).toFixed(2) }}
          </template>
        </Column>

        <Column header="Дії" style="min-width: 14rem">
          <template #body="{ data }">
            <div class="flex flex-wrap gap-2">
              <Button
                icon="pi pi-pause"
                size="small"
                text
                rounded
                @click="pauseOne(data)"
              />

              <Button
                icon="pi pi-play"
                size="small"
                text
                rounded
                severity="success"
                @click="resumeOne(data)"
              />

              <Button
                icon="pi pi-trash"
                size="small"
                text
                rounded
                severity="danger"
                @click="askDeleteOne(data)"
              />
            </div>
          </template>
        </Column>
      </DataTable>
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
            text
            @click="deleteDialogVisible = false"
          />
          <Button
            label="Видалити"
            severity="danger"
            @click="confirmDelete"
          />
        </div>
      </div>
    </Dialog>
  </section>
</template>