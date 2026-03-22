<script setup>
import { computed } from 'vue'
import MetricCard from '../components/MetricCard.vue'

const props = defineProps({
  ups: {
    type: Object,
    default: null
  },
  loading: {
    type: Boolean,
    default: false
  },
  error: {
    type: String,
    default: ''
  }
})

const data = computed(() => props.ups?.data ?? null)

const batteryValue = computed(() => {
  if (!data.value) return '—'
  return `${data.value.batteryPercent}%`
})

const batterySubvalue = computed(() => {
  if (!data.value) return ''
  return `${data.value.remainingMAh} mAh`
})

const vbusValue = computed(() => {
  if (!data.value) return '—'
  return `${data.value.vbusVoltageV.toFixed(3)} V`
})

const vbusSubvalue = computed(() => {
  if (!data.value) return ''
  return `${data.value.vbusCurrentA.toFixed(3)} A · ${data.value.vbusPowerW.toFixed(3)} W`
})

const batteryElectricalValue = computed(() => {
  if (!data.value) return '—'
  return `${data.value.batteryVoltageV.toFixed(3)} V`
})

const batteryElectricalSubvalue = computed(() => {
  if (!data.value) return ''
  return `${data.value.batteryCurrentA.toFixed(3)} A`
})

const cellsValue = computed(() => {
  if (!data.value) return '—'
  return `${data.value.cell1Mv} | ${data.value.cell2Mv} | ${data.value.cell3Mv} | ${data.value.cell4Mv} mV`
})
</script>

<template>
  <section class="space-y-3">
    <div
      v-if="loading"
      class="bg-panel rounded-2xl p-4 border border-white/10 text-white/60 text-sm"
    >
      Завантаження UPS...
    </div>

    <div
      v-else-if="error"
      class="bg-panel rounded-2xl p-4 border border-red-500/20 text-red-300 text-sm"
    >
      {{ error }}
    </div>

    <template v-else-if="data">
      <div class="grid grid-cols-2 xl:grid-cols-4 gap-3">
        <MetricCard
          label="Режим"
          :value="data.modeText"
          :subvalue="data.powerSourceText"
        />

        <MetricCard
          label="Заряд"
          :value="batteryValue"
          :subvalue="batterySubvalue"
        />

        <MetricCard
          label="VBUS"
          :value="vbusValue"
          :subvalue="vbusSubvalue"
        />

        <MetricCard
          label="Батарея"
          :value="batteryElectricalValue"
          :subvalue="batteryElectricalSubvalue"
        />
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-2 gap-3">
        <MetricCard
          label="Стан зарядки"
          :value="data.chargeText"
        />

        <MetricCard
          label="Дельта банок"
          :value="`${data.cellDeltaMv} mV`"
          :subvalue="data.cellDeltaText"
        />
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-2 gap-3">
        <MetricCard
          label="Банки"
          :value="cellsValue"
        />

        <MetricCard
          label="Система"
          :value="data.commText"
          :subvalue="`Прошивка: ${data.firmwareText}`"
        />
      </div>
    </template>
  </section>
</template>