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
  vpnSummary: {
    type: Object,
    default: null
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
  },
  rxSpeedHistory: {
    type: Array,
    default: () => []
  },
  txSpeedHistory: {
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
  return `↓ ${rx}\n↑ ${tx}`
})

const services = computed(() => [
  { label: 'Jellyfin', ok: !!props.metrics.services?.jellyfin },
  { label: 'qBittorrent', ok: !!props.metrics.services?.qBittorrent },
  { label: 'Sonarr', ok: !!props.metrics.services?.sonarr },
  { label: 'Radarr', ok: !!props.metrics.services?.radarr },
  { label: 'Prowlarr', ok: !!props.metrics.services?.prowlarr },
  { label: 'VPN', ok: !!props.vpnSummary?.vpnOk }
])

function formatSpeedChartValue(value) {
  return formatBytes(Number(value || 0)) + '/s'
}
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

    <div class="space-y-2">
      <div class="px-1 text-[10px] sm:text-xs uppercase tracking-wide text-white/60">
        Історія метрик
      </div>

      <div class="-mx-4 px-4 overflow-x-auto no-scrollbar">
        <div class="flex gap-4 min-w-max pr-4">
          <div class="w-[300px] sm:w-[360px] shrink-0">
            <MiniTrendChart
              title="CPU"
              subtitle="Останні виміри"
              :points="cpuUsageHistory"
              color="#7C83FF"
              :min="0"
              :max="100"
              :step-size="25"
              :formatter="(value) => `${value.toFixed(0)}%`"
            />
          </div>

          <div class="w-[300px] sm:w-[360px] shrink-0">
            <MiniTrendChart
              title="Температура CPU"
              subtitle="Останні виміри"
              :points="cpuTempHistory"
              color="#FF8A65"
              :min="30"
              :max="90"
              :step-size="15"
              :formatter="(value) => `${value.toFixed(0)}°C`"
            />
          </div>

          <div class="w-[300px] sm:w-[360px] shrink-0">
            <MiniTrendChart
              title="RAM"
              subtitle="Останні виміри"
              :points="ramUsageHistory"
              color="#31D0AA"
              :min="0"
              :max="100"
              :step-size="25"
              :formatter="(value) => `${value.toFixed(0)}%`"
            />
          </div>

          <div class="w-[300px] sm:w-[360px] shrink-0">
            <MiniTrendChart
              title="RX"
              subtitle="Швидкість отримання"
              :points="rxSpeedHistory"
              color="#60A5FA"
              :formatter="formatSpeedChartValue"
            />
          </div>

          <div class="w-[300px] sm:w-[360px] shrink-0">
            <MiniTrendChart
              title="TX"
              subtitle="Швидкість відправлення"
              :points="txSpeedHistory"
              color="#A78BFA"
              :formatter="formatSpeedChartValue"
            />
          </div>
        </div>
      </div>
    </div>
  </section>
</template>