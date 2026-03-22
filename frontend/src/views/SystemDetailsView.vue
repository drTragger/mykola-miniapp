<script setup>
import { computed } from 'vue'

const props = defineProps({
  systemData: {
    type: Object,
    required: true
  }
})

const system = computed(() => props.systemData?.system || {})
const network = computed(() => props.systemData?.network || {})
const vpn = computed(() => props.systemData?.vpn || {})
const users = computed(() => props.systemData?.users || [])

const bootTime = computed(() => {
  return system.value?.bootTimeUnix
    ? new Date(system.value.bootTimeUnix * 1000).toLocaleString()
    : '—'
})

const pingText = computed(() => {
  return typeof network.value?.pingMs === 'number'
    ? `${network.value.pingMs.toFixed(1)} ms`
    : '—'
})

const rxSpeedText = computed(() => network.value?.rxSpeedHuman || '—')
const txSpeedText = computed(() => network.value?.txSpeedHuman || '—')

const vpnStatusText = computed(() => {
  return vpn.value?.ok ? 'Активний' : 'Неактивний'
})

const vpnStatusClass = computed(() => {
  return vpn.value?.ok
    ? 'bg-green-500/10 text-green-300 border-green-500/20'
    : 'bg-red-500/10 text-red-300 border-red-500/20'
})

const qbitBindingOk = computed(() => vpn.value?.qbit?.binding === 'wg0')
const qbitBindingText = computed(() => {
  if (!vpn.value?.qbit?.binding) return '—'
  return qbitBindingOk.value
    ? `${vpn.value.qbit.binding} (OK)`
    : `${vpn.value.qbit.binding} (не wg0)`
})

const detailRowsLeft = computed(() => [
  { label: 'Хост', value: system.value?.hostname || '—' },
  { label: 'Платформа', value: system.value?.platform || '—' },
  { label: 'Версія ОС', value: system.value?.platformVersion || '—' },
  { label: 'Ядро', value: system.value?.kernelVersion || '—' },
  { label: 'Архітектура', value: system.value?.architecture || '—' },
  { label: 'Процеси', value: String(system.value?.processes ?? '—') },
  { label: 'Час запуску', value: bootTime.value }
])

const detailRowsRight = computed(() => [
  { label: 'Модель CPU', value: system.value?.cpuModel || '—' },
  {
    label: 'Частота',
    value: system.value?.cpuFrequencyMHz
      ? `${system.value.cpuFrequencyMHz.toFixed(0)} MHz`
      : '—'
  },
  { label: 'Ядер (логічних)', value: String(system.value?.logicalCpuCount ?? '—') },
  {
    label: 'Load 1 хв',
    value: typeof system.value?.load1 === 'number' ? system.value.load1.toFixed(2) : '—'
  },
  {
    label: 'Load 5 хв',
    value: typeof system.value?.load5 === 'number' ? system.value.load5.toFixed(2) : '—'
  },
  {
    label: 'Load 15 хв',
    value: typeof system.value?.load15 === 'number' ? system.value.load15.toFixed(2) : '—'
  }
])

const networkRows = computed(() => [
  { label: 'Локальний IP', value: network.value?.localIpv4 || '—' },
  { label: 'Публічний IP', value: network.value?.publicIp || '—' },
  { label: 'Пінг', value: pingText.value },
  { label: 'RX (всього)', value: network.value?.rxTotalHuman || '—' },
  { label: 'TX (всього)', value: network.value?.txTotalHuman || '—' },
  { label: 'RX швидкість', value: rxSpeedText.value },
  { label: 'TX швидкість', value: txSpeedText.value }
])

function normalizeConnectionType(type) {
  return String(type || '').trim().toLowerCase()
}

function connectionLabel(type) {
  const normalizedType = normalizeConnectionType(type)

  if (normalizedType === 'ssh') return 'SSH'
  if (normalizedType === 'local') return 'Локально'
  if (normalizedType === 'terminal') return 'Термінал'

  return type || 'Невідомо'
}

function connectionBadgeClass(type) {
  const normalizedType = normalizeConnectionType(type)

  if (normalizedType === 'ssh') {
    return 'bg-cyan-500/10 text-cyan-300 border-cyan-500/20'
  }

  if (normalizedType === 'local') {
    return 'bg-emerald-500/10 text-emerald-300 border-emerald-500/20'
  }

  if (normalizedType === 'terminal') {
    return 'bg-violet-500/10 text-violet-300 border-violet-500/20'
  }

  return 'bg-white/5 text-white/70 border-white/10'
}

function connectionIcon(type) {
  const normalizedType = normalizeConnectionType(type)

  if (normalizedType === 'ssh') return '🖥️'
  if (normalizedType === 'local') return '💻'
  if (normalizedType === 'terminal') return '⌨️'

  return '👤'
}
</script>

<template>
  <section class="space-y-4">
    <!-- SUMMARY -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-3">
      <div class="bg-panel rounded-2xl border border-white/10 p-3 shadow-custom">
        <div class="text-[10px] uppercase tracking-wide text-white/50 mb-1">Система</div>
        <div class="text-sm font-semibold text-white truncate">
          {{ system.hostname || '—' }}
        </div>
        <div class="text-xs text-white/50 mt-1 truncate">
          {{ system.platform || '—' }} {{ system.platformVersion || '' }}
        </div>
      </div>

      <div class="bg-panel rounded-2xl border border-white/10 p-3 shadow-custom">
        <div class="text-[10px] uppercase tracking-wide text-white/50 mb-1">CPU</div>
        <div class="text-sm font-semibold text-white truncate">
          {{ system.cpuModel || '—' }}
        </div>
        <div class="text-xs text-white/50 mt-1">
          {{ system.logicalCpuCount ?? '—' }} ядер •
          {{ system.cpuFrequencyMHz ? `${system.cpuFrequencyMHz.toFixed(0)} MHz` : '—' }}
        </div>
      </div>

      <div class="bg-panel rounded-2xl border border-white/10 p-3 shadow-custom">
        <div class="text-[10px] uppercase tracking-wide text-white/50 mb-1">Мережа</div>
        <div class="text-sm font-semibold text-white">
          {{ pingText }}
        </div>
        <div class="text-xs text-white/50 mt-1 truncate">
          {{ network.publicIp || '—' }}
        </div>
      </div>

      <div class="bg-panel rounded-2xl border border-white/10 p-3 shadow-custom">
        <div class="text-[10px] uppercase tracking-wide text-white/50 mb-1">VPN</div>
        <div
          class="inline-flex items-center gap-2 rounded-full border px-2.5 py-1 text-xs"
          :class="vpnStatusClass"
        >
          <span class="font-medium">{{ vpnStatusText }}</span>
        </div>
        <div class="text-xs text-white/50 mt-2">
          Handshake: {{ vpn.lastHandshakeAgo || '—' }}
        </div>
      </div>
    </div>

    <!-- USERS -->
    <div
      v-if="users.length"
      class="bg-panel rounded-2xl border border-white/10 p-4 shadow-custom"
    >
      <div class="flex items-center justify-between gap-3 mb-3">
        <div class="text-[10px] uppercase tracking-wide text-white/60">
          Активні користувачі
        </div>

        <div class="text-xs text-white/45">
          {{ users.length }} сес.
        </div>
      </div>

      <div class="grid grid-cols-1 xl:grid-cols-2 gap-3">
        <div
          v-for="user in users"
          :key="`${user.username}-${user.tty}-${user.loginAt}`"
          class="rounded-2xl border border-white/10 bg-white/[0.03] p-3"
        >
          <div class="flex items-start justify-between gap-3">
            <div class="flex items-center gap-3 min-w-0">
              <div
                class="w-10 h-10 rounded-xl bg-white/5 border border-white/10 flex items-center justify-center text-base shrink-0"
              >
                {{ connectionIcon(user.connectionType) }}
              </div>

              <div class="min-w-0">
                <div class="text-sm font-semibold text-white truncate">
                  {{ user.username || '—' }}
                </div>

                <div class="text-xs text-white/45 mt-1 truncate">
                  {{ user.tty || '—' }}
                </div>
              </div>
            </div>

            <div
              class="inline-flex items-center gap-1.5 rounded-full border px-2.5 py-1 text-[10px] shrink-0"
              :class="connectionBadgeClass(user.connectionType)"
            >
              <span>{{ connectionIcon(user.connectionType) }}</span>
              <span>{{ connectionLabel(user.connectionType) }}</span>
            </div>
          </div>

          <div class="mt-3 grid grid-cols-2 gap-x-4 gap-y-2 text-sm">
            <div class="text-white/45">IP / Host</div>
            <div class="text-white text-right break-all">
              {{ user.remoteIp || user.from || '—' }}
            </div>

            <div class="text-white/45">Локація</div>
            <div class="text-white text-right break-all">
              {{ user.location || '—' }}
            </div>

            <div class="text-white/45">Вхід</div>
            <div class="text-white text-right break-all">
              {{ user.loginAt || '—' }}
            </div>

            <div class="text-white/45">Idle</div>
            <div class="text-white text-right">
              {{ user.idle || '—' }}
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- DETAILS -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
      <div class="bg-panel rounded-2xl border border-white/10 p-4 shadow-custom">
        <div class="text-[10px] uppercase tracking-wide text-white/60 mb-3">
          Система
        </div>

        <div class="space-y-2">
          <div
            v-for="row in detailRowsLeft"
            :key="row.label"
            class="flex items-start justify-between gap-3 border-b border-white/5 pb-2 last:border-b-0 last:pb-0"
          >
            <span class="text-sm text-white/45">{{ row.label }}</span>
            <span class="text-sm text-white text-right break-all">{{ row.value }}</span>
          </div>
        </div>
      </div>

      <div class="bg-panel rounded-2xl border border-white/10 p-4 shadow-custom">
        <div class="text-[10px] uppercase tracking-wide text-white/60 mb-3">
          CPU та навантаження
        </div>

        <div class="space-y-2">
          <div
            v-for="row in detailRowsRight"
            :key="row.label"
            class="flex items-start justify-between gap-3 border-b border-white/5 pb-2 last:border-b-0 last:pb-0"
          >
            <span class="text-sm text-white/45">{{ row.label }}</span>
            <span class="text-sm text-white text-right break-all">{{ row.value }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- NETWORK -->
    <div class="bg-panel rounded-2xl border border-white/10 p-4 shadow-custom">
      <div class="text-[10px] uppercase tracking-wide text-white/60 mb-3">
        Мережа
      </div>

      <div class="grid grid-cols-1 sm:grid-cols-2 gap-x-6 gap-y-2">
        <div
          v-for="row in networkRows"
          :key="row.label"
          class="flex items-start justify-between gap-3 border-b border-white/5 pb-2 last:border-b-0 sm:last:border-b"
        >
          <span class="text-sm text-white/45">{{ row.label }}</span>
          <span class="text-sm text-white text-right break-all">{{ row.value }}</span>
        </div>
      </div>
    </div>

    <!-- VPN -->
    <div class="bg-panel rounded-2xl p-4 shadow-custom border border-white/10 space-y-4">
      <div class="flex items-center justify-between gap-3 flex-wrap">
        <div class="text-[10px] uppercase tracking-wide text-white/60">
          VPN / WireGuard
        </div>

        <div
          class="inline-flex items-center gap-2 rounded-full border px-3 py-1.5 text-xs"
          :class="vpnStatusClass"
        >
          <span class="font-medium">{{ vpnStatusText }}</span>
        </div>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-3 gap-3">
        <!-- WireGuard -->
        <div class="rounded-2xl border border-white/10 bg-white/[0.03] p-3">
          <div class="text-xs text-white/40 mb-3">WireGuard</div>

          <div class="space-y-2 text-sm">
            <div class="flex justify-between gap-3">
              <span class="text-white/45">IP</span>
              <span class="text-white text-right break-all">{{ vpn.wgIp || '—' }}</span>
            </div>
            <div class="flex justify-between gap-3">
              <span class="text-white/45">Endpoint</span>
              <span class="text-white text-right break-all">{{ vpn.endpoint || '—' }}</span>
            </div>
            <div class="flex justify-between gap-3">
              <span class="text-white/45">Handshake</span>
              <span class="text-white text-right">{{ vpn.lastHandshakeAgo || '—' }}</span>
            </div>
            <div class="flex justify-between gap-3">
              <span class="text-white/45">Отримано</span>
              <span class="text-white text-right">{{ vpn.rx || '—' }}</span>
            </div>
            <div class="flex justify-between gap-3">
              <span class="text-white/45">Відправлено</span>
              <span class="text-white text-right">{{ vpn.tx || '—' }}</span>
            </div>
          </div>
        </div>

        <!-- qBittorrent -->
        <div class="rounded-2xl border border-white/10 bg-white/[0.03] p-3">
          <div class="text-xs text-white/40 mb-3">qBittorrent</div>

          <div class="space-y-2 text-sm">
            <div class="flex justify-between gap-3">
              <span class="text-white/45">Сервіс</span>
              <span class="text-white">{{ vpn.qbit?.serviceOk ? '✅' : '❌' }}</span>
            </div>
            <div class="flex justify-between gap-3">
              <span class="text-white/45">User</span>
              <span class="text-white text-right break-all">{{ vpn.qbit?.user || '—' }}</span>
            </div>
            <div class="flex justify-between gap-3">
              <span class="text-white/45">Інтерфейс</span>
              <span
                class="text-right"
                :class="qbitBindingOk ? 'text-green-300' : 'text-red-300'"
              >
                {{ qbitBindingText }}
              </span>
            </div>
            <div class="flex justify-between gap-3">
              <span class="text-white/45">Web UI</span>
              <span class="text-white text-right break-all">{{ vpn.qbit?.webui || '—' }}</span>
            </div>
          </div>
        </div>

        <!-- Routing -->
        <div class="rounded-2xl border border-white/10 bg-white/[0.03] p-3">
          <div class="text-xs text-white/40 mb-3">Routing</div>

          <div class="space-y-2 text-sm">
            <div class="flex justify-between gap-3">
              <span class="text-white/45">ip rule</span>
              <span class="text-white">{{ vpn.ruleOk ? '✅' : '❌' }}</span>
            </div>
            <div class="flex justify-between gap-3">
              <span class="text-white/45">Route через wg0</span>
              <span class="text-white">{{ vpn.routeOk ? '✅' : '❌' }}</span>
            </div>
            <div class="pt-1">
              <div class="text-white/45 mb-1">Таблиця</div>
              <div class="text-white break-all text-sm">
                {{ vpn.routeTable || '—' }}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>