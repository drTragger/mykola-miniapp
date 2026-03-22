<script setup>
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
  }
})
</script>

<template>
  <section class="section active">
    <div class="grid">
      <MetricCard
        label="Температура CPU"
        :value="formatTemperature(metrics.overview.cpuTemperatureCelsius)"
        subvalue="Поточна температура процесора"
      />

      <MetricCard
        label="CPU Usage"
        :value="formatPercent(metrics.overview.cpuUsagePercent)"
        subvalue="Завантаження процесора"
      />

      <MetricCard
        label="RAM"
        :value="formatPercent(metrics.overview.ramUsagePercent)"
        :subvalue="`${formatBytes(metrics.overview.ramUsedBytes)} / ${formatBytes(metrics.overview.ramTotalBytes)}`"
      />

      <MetricCard
        label="Disk"
        :value="formatPercent(metrics.overview.diskUsagePercent)"
        :subvalue="`${formatBytes(metrics.overview.diskUsedBytes)} / ${formatBytes(metrics.overview.diskTotalBytes)}`"
      />

      <MetricCard
        label="Uptime"
        :value="formatUptime(metrics.overview.uptimeSeconds)"
        subvalue="Час безперервної роботи"
      />

      <MetricCard
        label="Hostname"
        :value="metrics.system.hostname || '—'"
        subvalue="Поточний вузол"
      />
    </div>
  </section>
</template>