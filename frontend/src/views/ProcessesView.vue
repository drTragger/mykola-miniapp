<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { fetchProcesses, sendProcessAction } from '../api/processes'
import { formatBytes, formatUptime } from '../utils/formatters'

const props = defineProps({
  active: {
    type: Boolean,
    default: false
  }
})

const loading = ref(false)
const error = ref('')
const items = ref([])
const total = ref(0)

const search = ref('')
const debouncedSearch = ref('')

const sort = ref('cpu')
const dir = ref('desc')

const page = ref(1)
const limit = ref(25)

const actionLoadingPid = ref(null)
const autoRefreshId = ref(null)
const searchDebounceId = ref(null)

const sortOptions = [
  { label: 'CPU', value: 'cpu' },
  { label: 'Памʼять', value: 'memory' },
  { label: 'Імʼя', value: 'name' },
  { label: 'PID', value: 'pid' },
  { label: 'Аптайм', value: 'uptime' }
]

const dirOptions = [
  { label: 'Спадання', value: 'desc' },
  { label: 'Зростання', value: 'asc' }
]

const totalPages = computed(() => {
  if (!total.value || !limit.value) return 1
  return Math.max(1, Math.ceil(total.value / limit.value))
})

const offset = computed(() => {
  return (page.value - 1) * limit.value
})

const summaryText = computed(() => {
  if (loading.value) return 'Завантаження...'
  return `${total.value} процесів`
})

function statusLabel(status) {
  const map = {
    R: 'Працює',
    S: 'Сон',
    D: 'Очікування',
    T: 'Зупинений',
    Z: 'Zombie',
    I: 'Idle'
  }

  return map[status] || status || '—'
}

function statusClass(status) {
  switch (status) {
    case 'R':
      return 'bg-green-500/10 text-green-300 border-green-500/20'
    case 'S':
    case 'I':
      return 'bg-white/5 text-white/70 border-white/10'
    case 'T':
      return 'bg-yellow-500/10 text-yellow-300 border-yellow-500/20'
    case 'D':
      return 'bg-orange-500/10 text-orange-300 border-orange-500/20'
    case 'Z':
      return 'bg-red-500/10 text-red-300 border-red-500/20'
    default:
      return 'bg-white/5 text-white/70 border-white/10'
  }
}

function shortCmdline(item) {
  if (!item?.cmdline) return '—'
  if (item.cmdline.length <= 110) return item.cmdline
  return item.cmdline.slice(0, 110) + '...'
}

function formatCpu(value) {
  return `${Number(value || 0).toFixed(1)}%`
}

function formatMemPercent(value) {
  return `${Number(value || 0).toFixed(1)}%`
}

function formatThreads(value) {
  return typeof value === 'number' ? String(value) : '—'
}

function formatProcessUptime(value) {
  if (!value || value <= 0) return '—'
  return formatUptime(value)
}

async function loadProcesses() {
  loading.value = true
  error.value = ''

  try {
    const data = await fetchProcesses({
      search: debouncedSearch.value,
      sort: sort.value,
      dir: dir.value,
      limit: limit.value,
      offset: offset.value
    })

    items.value = Array.isArray(data.items) ? data.items : []
    total.value = Number(data.total || 0)
  } catch (err) {
    console.error(err)
    error.value = err.message || 'Не вдалося завантажити процеси'
  } finally {
    loading.value = false
  }
}

async function runAction(pid, action) {
  const confirmed = window.confirm(`Виконати дію "${action}" для PID ${pid}?`)
  if (!confirmed) return

  actionLoadingPid.value = pid

  try {
    await sendProcessAction(pid, action)
    await loadProcesses()
  } catch (err) {
    console.error(err)
    window.alert(err.message || 'Не вдалося виконати дію')
  } finally {
    actionLoadingPid.value = null
  }
}

function setSort(value) {
  if (sort.value === value) {
    dir.value = dir.value === 'desc' ? 'asc' : 'desc'
    return
  }

  sort.value = value
  dir.value = 'desc'
}

function prevPage() {
  if (page.value > 1) {
    page.value -= 1
  }
}

function nextPage() {
  if (page.value < totalPages.value) {
    page.value += 1
  }
}

function startAutoRefresh() {
  stopAutoRefresh()

  autoRefreshId.value = setInterval(() => {
    loadProcesses()
  }, 10000)
}

function stopAutoRefresh() {
  if (autoRefreshId.value) {
    clearInterval(autoRefreshId.value)
    autoRefreshId.value = null
  }
}

watch(search, (value) => {
  if (searchDebounceId.value) {
    clearTimeout(searchDebounceId.value)
  }

  searchDebounceId.value = setTimeout(() => {
    debouncedSearch.value = value
    page.value = 1
  }, 400)
})

watch([debouncedSearch, sort, dir, limit], () => {
  page.value = 1
})

watch([debouncedSearch, sort, dir, limit, page], () => {
  if (props.active) {
    loadProcesses()
  }
})

watch(
  () => props.active,
  async (active) => {
    if (active) {
      await loadProcesses()
      startAutoRefresh()
    } else {
      stopAutoRefresh()
    }
  },
  { immediate: true }
)

onMounted(() => {
  if (props.active) {
    startAutoRefresh()
  }
})

onBeforeUnmount(() => {
  stopAutoRefresh()

  if (searchDebounceId.value) {
    clearTimeout(searchDebounceId.value)
  }
})
</script>

<template>
  <section class="space-y-4">
    <div class="bg-panel rounded-2xl border border-white/10 p-4 shadow-custom">
      <div class="flex items-start justify-between gap-3 flex-wrap">
        <div>
          <div class="text-[10px] uppercase tracking-wide text-white/60 mb-1">
            Процеси
          </div>

          <div class="text-lg font-semibold text-white">
            Активні процеси
          </div>

          <div class="text-sm text-white/45 mt-1">
            {{ summaryText }}
          </div>
        </div>

        <button
          class="inline-flex items-center gap-2 rounded-full border border-white/10 bg-white/5 px-3 py-2 text-sm text-white hover:bg-white/10 transition"
          @click="loadProcesses"
        >
          Оновити
        </button>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-4 gap-3 mt-4">
        <div class="lg:col-span-2">
          <label class="text-xs text-white/45 mb-1 block">Пошук</label>
          <input
            v-model="search"
            type="text"
            placeholder="Імʼя, команда, користувач..."
            class="w-full rounded-2xl border border-white/10 bg-white/5 px-4 py-3 text-sm text-white outline-none placeholder:text-white/30"
          >
        </div>

        <div>
          <label class="text-xs text-white/45 mb-1 block">Сортування</label>
          <select
            v-model="sort"
            class="w-full rounded-2xl border border-white/10 bg-white/5 px-4 py-3 text-sm text-white outline-none"
          >
            <option
              v-for="option in sortOptions"
              :key="option.value"
              :value="option.value"
              class="bg-slate-900"
            >
              {{ option.label }}
            </option>
          </select>
        </div>

        <div>
          <label class="text-xs text-white/45 mb-1 block">Напрям</label>
          <select
            v-model="dir"
            class="w-full rounded-2xl border border-white/10 bg-white/5 px-4 py-3 text-sm text-white outline-none"
          >
            <option
              v-for="option in dirOptions"
              :key="option.value"
              :value="option.value"
              class="bg-slate-900"
            >
              {{ option.label }}
            </option>
          </select>
        </div>
      </div>

      <div class="flex flex-wrap gap-2 mt-3">
        <button
          class="rounded-full border px-3 py-1.5 text-xs transition"
          :class="sort === 'cpu' ? 'border-primary/30 bg-primary/15 text-primary' : 'border-white/10 bg-white/5 text-white/70'"
          @click="setSort('cpu')"
        >
          CPU
        </button>

        <button
          class="rounded-full border px-3 py-1.5 text-xs transition"
          :class="sort === 'memory' ? 'border-primary/30 bg-primary/15 text-primary' : 'border-white/10 bg-white/5 text-white/70'"
          @click="setSort('memory')"
        >
          Памʼять
        </button>

        <button
          class="rounded-full border px-3 py-1.5 text-xs transition"
          :class="sort === 'uptime' ? 'border-primary/30 bg-primary/15 text-primary' : 'border-white/10 bg-white/5 text-white/70'"
          @click="setSort('uptime')"
        >
          Аптайм
        </button>

        <button
          class="rounded-full border px-3 py-1.5 text-xs transition"
          :class="sort === 'pid' ? 'border-primary/30 bg-primary/15 text-primary' : 'border-white/10 bg-white/5 text-white/70'"
          @click="setSort('pid')"
        >
          PID
        </button>
      </div>
    </div>

    <div
      v-if="error"
      class="bg-red-500/10 border border-red-500/20 rounded-2xl p-4 text-sm text-red-200"
    >
      {{ error }}
    </div>

    <div
      v-if="loading"
      class="bg-panel rounded-2xl border border-white/10 p-4 shadow-custom"
    >
      <div class="text-sm text-white/60">
        Завантаження процесів...
      </div>
    </div>

    <div
      v-else-if="!items.length"
      class="bg-panel rounded-2xl border border-white/10 p-4 shadow-custom"
    >
      <div class="text-sm text-white/60">
        Нічого не знайдено.
      </div>
    </div>

    <div
      v-else
      class="space-y-3"
    >
      <div
        v-for="item in items"
        :key="item.pid"
        class="bg-panel rounded-2xl border border-white/10 p-4 shadow-custom"
      >
        <div class="flex items-start justify-between gap-3 flex-wrap">
          <div class="min-w-0">
            <div class="flex items-center gap-2 flex-wrap">
              <div class="text-base font-semibold text-white break-all">
                {{ item.name || 'unknown' }}
              </div>

              <div class="text-xs rounded-full border border-white/10 bg-white/5 px-2 py-1 text-white/70">
                PID {{ item.pid }}
              </div>

              <div
                class="text-xs rounded-full border px-2 py-1"
                :class="statusClass(item.status)"
              >
                {{ statusLabel(item.status) }}
              </div>

              <div
                v-if="item.isCritical"
                class="text-xs rounded-full border border-yellow-500/20 bg-yellow-500/10 px-2 py-1 text-yellow-300"
              >
                critical
              </div>

              <div
                v-if="item.serviceName"
                class="text-xs rounded-full border border-cyan-500/20 bg-cyan-500/10 px-2 py-1 text-cyan-300"
              >
                {{ item.serviceName }}
              </div>
            </div>

            <div class="text-sm text-white/45 mt-2 break-all">
              {{ shortCmdline(item) }}
            </div>
          </div>

          <div class="flex gap-2 flex-wrap">
            <button
              class="rounded-xl border border-yellow-500/20 bg-yellow-500/10 px-3 py-2 text-xs text-yellow-200 disabled:opacity-50"
              :disabled="item.isCritical || actionLoadingPid === item.pid"
              @click="runAction(item.pid, 'stop')"
            >
              Stop
            </button>

            <button
              class="rounded-xl border border-green-500/20 bg-green-500/10 px-3 py-2 text-xs text-green-200 disabled:opacity-50"
              :disabled="item.isCritical || actionLoadingPid === item.pid"
              @click="runAction(item.pid, 'cont')"
            >
              Continue
            </button>

            <button
              class="rounded-xl border border-orange-500/20 bg-orange-500/10 px-3 py-2 text-xs text-orange-200 disabled:opacity-50"
              :disabled="item.isCritical || actionLoadingPid === item.pid"
              @click="runAction(item.pid, 'term')"
            >
              Term
            </button>

            <button
              class="rounded-xl border border-red-500/20 bg-red-500/10 px-3 py-2 text-xs text-red-200 disabled:opacity-50"
              :disabled="item.isCritical || actionLoadingPid === item.pid"
              @click="runAction(item.pid, 'kill')"
            >
              Kill
            </button>
          </div>
        </div>

        <div class="grid grid-cols-2 lg:grid-cols-4 gap-3 mt-4">
          <div class="rounded-2xl border border-white/10 bg-white/[0.03] p-3">
            <div class="text-[10px] uppercase tracking-wide text-white/45 mb-1">CPU</div>
            <div class="text-lg font-semibold text-white">{{ formatCpu(item.cpuPercent) }}</div>
          </div>

          <div class="rounded-2xl border border-white/10 bg-white/[0.03] p-3">
            <div class="text-[10px] uppercase tracking-wide text-white/45 mb-1">RAM</div>
            <div class="text-lg font-semibold text-white">{{ formatMemPercent(item.memoryPercent) }}</div>
            <div class="text-xs text-white/45 mt-1">{{ formatBytes(item.rssBytes) }}</div>
          </div>

          <div class="rounded-2xl border border-white/10 bg-white/[0.03] p-3">
            <div class="text-[10px] uppercase tracking-wide text-white/45 mb-1">Threads</div>
            <div class="text-lg font-semibold text-white">{{ formatThreads(item.threads) }}</div>
          </div>

          <div class="rounded-2xl border border-white/10 bg-white/[0.03] p-3">
            <div class="text-[10px] uppercase tracking-wide text-white/45 mb-1">Аптайм</div>
            <div class="text-lg font-semibold text-white">{{ formatProcessUptime(item.uptimeSec) }}</div>
          </div>
        </div>

        <div class="grid grid-cols-1 lg:grid-cols-2 gap-3 mt-3">
          <div class="rounded-2xl border border-white/10 bg-white/[0.03] p-3">
            <div class="text-xs text-white/45 mb-1">User</div>
            <div class="text-sm text-white break-all">{{ item.user || '—' }}</div>
          </div>

          <div class="rounded-2xl border border-white/10 bg-white/[0.03] p-3">
            <div class="text-xs text-white/45 mb-1">Executable</div>
            <div class="text-sm text-white break-all">{{ item.executable || '—' }}</div>
          </div>
        </div>
      </div>

      <div class="bg-panel rounded-2xl border border-white/10 p-4 shadow-custom">
        <div class="flex items-center justify-between gap-3 flex-wrap">
          <div class="text-sm text-white/60">
            Сторінка {{ page }} / {{ totalPages }}
          </div>

          <div class="flex items-center gap-2">
            <button
              class="rounded-xl border border-white/10 bg-white/5 px-3 py-2 text-sm text-white disabled:opacity-40"
              :disabled="page <= 1"
              @click="prevPage"
            >
              Назад
            </button>

            <button
              class="rounded-xl border border-white/10 bg-white/5 px-3 py-2 text-sm text-white disabled:opacity-40"
              :disabled="page >= totalPages"
              @click="nextPage"
            >
              Далі
            </button>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>