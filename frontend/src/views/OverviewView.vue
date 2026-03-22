<script setup>
import DonutMetricCard from '../components/DonutMetricCard.vue'
import MiniTrendChart from '../components/MiniTrendChart.vue'
import MetricCard from '../components/MetricCard.vue'
import {
  formatBytes,
  formatPercent,
  formatTemperature,
  formatUptime
} from '../utils/formatters'

defineProps({
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
</script>

<template>
  <section class="overview-layout">
    <div class="overview-donuts">
      <DonutMetricCard
        label="CPU"
        :value="metrics.overview.cpuUsagePercent || 0"
        subtitle="Поточне завантаження"
        color="#7C83FF"
      />

      <DonutMetricCard
        label="RAM"
        :value="metrics.overview.ramUsagePercent || 0"
        subtitle="Використання пам'яті"
        color="#31D0AA"
      />

      <DonutMetricCard
        label="Disk"
        :value="metrics.overview.diskUsagePercent || 0"
        subtitle="Зайняте місце"
        color="#FF8A65"
      />
    </div>

    <div class="overview-main-grid">
      <MetricCard
        label="Температура CPU"
        :value="formatTemperature(metrics.overview.cpuTemperatureCelsius)"
        subvalue="Поточна температура процесора"
      />

      <MetricCard
        label="Uptime"
        :value="formatUptime(metrics.overview.uptimeSeconds)"
        subvalue="Час безперервної роботи"
      />

      <MetricCard
        label="RAM"
        :value="formatPercent(metrics.overview.ramUsagePercent)"
        :subvalue="`${formatBytes(metrics.overview.ramUsedBytes)} / ${formatBytes(metrics.overview.ramTotalBytes)}`"
      />

      <MetricCard
        label="Hostname"
        :value="metrics.system.hostname || '—'"
        subvalue="Поточний вузол"
      />
    </div>

    <div class="overview-trends">
      <MiniTrendChart
        title="CPU Usage Trend"
        subtitle="Останні виміри"
        :points="cpuUsageHistory"
        color="#7C83FF"
        :formatter="(value) => `${value.toFixed(1)}%`"
      />

      <MiniTrendChart
        title="CPU Temperature Trend"
        subtitle="Останні виміри"
        :points="cpuTempHistory"
        color="#FF8A65"
        :formatter="(value) => `${value.toFixed(1)}°C`"
      />

      <MiniTrendChart
        title="RAM Usage Trend"
        subtitle="Останні виміри"
        :points="ramUsageHistory"
        color="#31D0AA"
        :formatter="(value) => `${value.toFixed(1)}%`"
      />
    </div>
  </section>
</template>