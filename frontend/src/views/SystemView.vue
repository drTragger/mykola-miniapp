<script setup>
import { computed } from 'vue'
import InfoListCard from '../components/InfoListCard.vue'
import {
  formatDateTime,
  formatFixed,
  formatFrequency
} from '../utils/formatters'

const props = defineProps({
  metrics: {
    type: Object,
    required: true
  }
})

const systemItems = computed(() => [
  { key: 'Hostname', value: props.metrics.system?.hostname || '—' },
  { key: 'Platform', value: props.metrics.system?.platform || '—' },
  { key: 'Platform Version', value: props.metrics.system?.platformVersion || '—' },
  { key: 'Kernel', value: props.metrics.system?.kernelVersion || '—' },
  { key: 'Architecture', value: props.metrics.system?.architecture || '—' },
  { key: 'Processes', value: String(props.metrics.system?.processes ?? '—') },
  { key: 'Boot Time', value: formatDateTime(props.metrics.system?.bootTimeUnix) }
])

const cpuItems = computed(() => [
  { key: 'Model', value: props.metrics.system?.cpuModel || '—' },
  { key: 'Frequency', value: formatFrequency(props.metrics.system?.cpuFrequencyMHz) },
  { key: 'Logical Cores', value: String(props.metrics.system?.logicalCpuCount ?? '—') }
])

const loadItems = computed(() => [
  { key: '1 min', value: formatFixed(props.metrics.system?.load1) },
  { key: '5 min', value: formatFixed(props.metrics.system?.load5) },
  { key: '15 min', value: formatFixed(props.metrics.system?.load15) }
])
</script>

<template>
  <section class="grid-single">
    <InfoListCard label="Система" :items="systemItems" />
    <InfoListCard label="CPU" :items="cpuItems" />
    <InfoListCard label="Load Average" :items="loadItems" />
  </section>
</template>