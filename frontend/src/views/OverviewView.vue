<script setup>
import { computed } from 'vue'
import DonutMetricCard from '../components/DonutMetricCard.vue'
import MiniTrendChart from '../components/MiniTrendChart.vue'
import MetricCard from '../components/MetricCard.vue'
import StatusPill from '../components/StatusPill.vue'
import {
  formatBytes,
  formatTemperature
} from '../utils/formatters'

const props = defineProps({
  metrics: {
    type: Object,
    required: true
  },
  cpuUsageHistory: {
    type: Array,
    default: () => []
  },
  cpuTempHistory: {
    type: Array,
    default: () => []
  },
  ramUsageHistory: {
    type: Array,
    default: () => []
  }
})

const ramValue = computed(() => {
  return `${formatBytes(props.metrics.overview?.ramUsedBytes)} / ${formatBytes(props.metrics.overview?.ramTotalBytes)}`
})

const pingValue = computed(() => {
  const ping = props.metrics.network?.pingMs
  return typeof ping === 'number' ? `${ping.toFixed(1)} ms` : '—'
})

const publicIpValue = computed(() => {
  return props.metrics.network?.publicIp || '—'
})

const rxTxValue = computed(() => {
  const rx = props.metrics.network?.rxSpeedHuman || '—'
  const tx = props.metrics.network?.txSpeedHuman || '—'
  return `↓ ${rx} · ↑ ${tx}`
})

const rxTxSubvalue = computed(() => {
  const rxTotal = props.metrics.network?.rxTotalHuman || ''
  const txTotal = props.metrics.network?.txTotalHuman || ''
  if (!rxTotal && !txTotal) return ''
  return `RX ${rxTotal} · TX ${txTotal}`
})

const services = computed(() => [
  { label: 'Jellyfin', ok: !!props.metrics.services?.jellyfin },
  { label: 'qBittorrent', ok: !!props.metrics.services?.qBittorrent },
  { label: 'Sonarr', ok: !!props.metrics.services?.sonarr },
  { label: 'Radarr', ok: !!props.metrics.services?.radarr },
  { label: 'Prowlarr', ok: !!props.metrics.services?.prowlarr },
  { label: 'Fail2Ban', ok: !!props.metrics.services?.fail2ban },
  { label: 'VPN', ok: !!props.metrics.vpn?.ok }
])
</script>

<template>
  <section class="space-y-3">
    <div class="bg-panel rounded-2xl px-2 py-3 shadow-custom border border-white/10">
      <div class="grid grid-cols-3 gap-1 sm:gap-2 lg:gap-3">
        <DonutMetricCard
          compact
          label="CPU"
          :value="metrics.overview.cpuUsagePercent || 0"
          color="#7C83FF"
        />

        <DonutMetricCard
          compact
          label="RAM"
          :value="metrics.overview.ramUsagePercent || 0"
          color="#31D0AA"
        />

        <DonutMetricCard
          compact
          label="Disk"
          :value="metrics.overview.diskUsagePercent || 0"
          color="#FF8A65"
        />
      </div>
    </div>

    <div class="grid grid-cols-2 xl:grid-cols-4 gap-3">
      <MetricCard
        label="Температура"
        :value="formatTemperature(metrics.overview.cpuTemperatureCelsius)"
      />

      <MetricCard
        label="RAM"
        :value="ramValue"
      />

      <MetricCard
        label="Ping / IP"
        :value="pingValue"
        :subvalue="publicIpValue"
      />

      <MetricCard
        label="RX / TX"
        :value="rxTxValue"
        :subvalue="rxTxSubvalue"
      />
    </div>

    <div class="bg-panel rounded-2xl p-3 shadow-custom border border-white/10">
      <div class="text-[10px] sm:text-xs uppercase tracking-wide text-white/60 mb-3">
        Сервіси
      </div>

      <div class="flex flex-wrap gap-2">
        <StatusPill
          v-for="service in services"
          :key="service.label"
          :label="service.label"
          :ok="service.ok"
        />
      </div>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-3">
      <MiniTrendChart
        title="CPU"
        subtitle="Останні виміри"
        :points="cpuUsageHistory"
        color="#7C83FF"
        :formatter="(value) => `${value.toFixed(1)}%`"
      />

      <MiniTrendChart
        title="Температура CPU"
        subtitle="Останні виміри"
        :points="cpuTempHistory"
        color="#FF8A65"
        :formatter="(value) => `${value.toFixed(1)}°C`"
      />
    </div>
  </section>
</template>