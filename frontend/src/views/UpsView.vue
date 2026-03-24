<script setup>
import { computed } from 'vue'
import ProgressBar from 'primevue/progressbar'
import MiniTrendChart from '../components/MiniTrendChart.vue'
import MultiLineTrendChart from '../components/MultiLineTrendChart.vue'
import batteryIcon from '../assets/battery.png'

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
  },
  batteryPercentHistory: {
    type: Array,
    default: () => []
  },
  cellDeltaHistory: {
    type: Array,
    default: () => []
  },
    cell1History: {
    type: Array,
    default: () => []
  },
  cell2History: {
    type: Array,
    default: () => []
  },
  cell3History: {
    type: Array,
    default: () => []
  },
  cell4History: {
    type: Array,
    default: () => []
  }
})

const data = computed(() => props.ups?.data ?? null)

const batteryPercent = computed(() => data.value?.batteryPercent ?? 0)

const batteryValue = computed(() => {
  if (!data.value) return '—'
  return `${data.value.batteryPercent}%`
})

const batteryCapacityText = computed(() => {
  if (!data.value) return ''
  if (!data.value.fullCapacityMAh) return `${data.value.remainingMAh} mAh`
  return `${data.value.remainingMAh} / ${data.value.fullCapacityMAh} mAh`
})

const vbusValue = computed(() => {
  if (!data.value) return '—'
  return `${data.value.vbusVoltageV.toFixed(3)} V`
})

const vbusCurrentText = computed(() => {
  if (!data.value) return '—'
  return `${data.value.vbusCurrentA.toFixed(3)} A`
})

const vbusPowerText = computed(() => {
  if (!data.value) return '—'
  return `${data.value.vbusPowerW.toFixed(3)} W`
})

const batteryElectricalValue = computed(() => {
  if (!data.value) return '—'
  return `${data.value.batteryVoltageV.toFixed(3)} V`
})

const batteryElectricalCurrent = computed(() => {
  if (!data.value) return '—'
  return `${data.value.batteryCurrentA.toFixed(3)} A`
})

const cells = computed(() => {
  if (!data.value) return []

  const raw = [
    { label: 'Банка 1', value: data.value.cell1Mv },
    { label: 'Банка 2', value: data.value.cell2Mv },
    { label: 'Банка 3', value: data.value.cell3Mv },
    { label: 'Банка 4', value: data.value.cell4Mv }
  ]

  return raw.map((cell) => ({
    ...cell,
    percent: normalizeCellMv(cell.value)
  }))
})

const deltaSeverityClass = computed(() => {
  const delta = data.value?.cellDeltaMv ?? 0

  if (delta < 50) return 'bg-green-500/10 text-green-300 border-green-500/20'
  if (delta < 100) return 'bg-emerald-500/10 text-emerald-300 border-emerald-500/20'
  if (delta < 200) return 'bg-yellow-500/10 text-yellow-300 border-yellow-500/20'
  return 'bg-red-500/10 text-red-300 border-red-500/20'
})

const etaValue = computed(() => {
  if (!data.value) return '—'

  if (data.value.timeToChargeText && data.value.timeToChargeText !== '—') {
    return data.value.timeToChargeText
  }

  if (data.value.timeToDischargeText && data.value.timeToDischargeText !== '—') {
    return data.value.timeToDischargeText
  }

  return '—'
})

const etaLabel = computed(() => {
  if (!data.value) return 'ETA'

  if (data.value.timeToChargeText && data.value.timeToChargeText !== '—') {
    return 'До повного заряду'
  }

  if (data.value.timeToDischargeText && data.value.timeToDischargeText !== '—') {
    return 'Час роботи'
  }

  return 'ETA'
})

function normalizeCellMv(mv) {
  const min = 3000
  const max = 4200
  const value = ((mv - min) / (max - min)) * 100
  return Math.max(0, Math.min(100, Math.round(value)))
}

const cellVoltageLabels = computed(() => {
  return props.cell1History.map((point) => point.time)
})

const cellVoltageDatasets = computed(() => [
  {
    label: 'Банка 1',
    color: '#60A5FA',
    data: props.cell1History.map((point) => point.value)
  },
  {
    label: 'Банка 2',
    color: '#34D399',
    data: props.cell2History.map((point) => point.value)
  },
  {
    label: 'Банка 3',
    color: '#F59E0B',
    data: props.cell3History.map((point) => point.value)
  },
  {
    label: 'Банка 4',
    color: '#F472B6',
    data: props.cell4History.map((point) => point.value)
  }
])

const allCellVoltageValues = computed(() => {
  return [
    ...props.cell1History.map((point) => point.value),
    ...props.cell2History.map((point) => point.value),
    ...props.cell3History.map((point) => point.value),
    ...props.cell4History.map((point) => point.value)
  ].filter((value) => typeof value === 'number' && value > 0)
})

const cellVoltageMin = computed(() => {
  if (!allCellVoltageValues.value.length) {
    return 3000
  }

  const min = Math.min(...allCellVoltageValues.value)
  return Math.floor((min - 20) / 10) * 10
})

const cellVoltageMax = computed(() => {
  if (!allCellVoltageValues.value.length) {
    return 4300
  }

  const max = Math.max(...allCellVoltageValues.value)
  return Math.ceil((max + 20) / 10) * 10
})

const cellVoltageStep = computed(() => {
  const range = cellVoltageMax.value - cellVoltageMin.value

  if (range <= 80) return 10
  if (range <= 160) return 20
  return 50
})

const batteryHistoryValues = computed(() => {
  return props.batteryPercentHistory
    .map((point) => point.value)
    .filter((value) => typeof value === 'number')
})

const batteryChartMin = computed(() => {
  if (!batteryHistoryValues.value.length) return 0

  const min = Math.min(...batteryHistoryValues.value)
  const max = Math.max(...batteryHistoryValues.value)
  const range = max - min

  const padding = Math.max(3, Math.ceil(range * 0.25))
  return Math.max(0, min - padding)
})

const batteryChartMax = computed(() => {
  if (!batteryHistoryValues.value.length) return 100

  const min = Math.min(...batteryHistoryValues.value)
  const max = Math.max(...batteryHistoryValues.value)
  const range = max - min

  const padding = Math.max(3, Math.ceil(range * 0.25))
  return Math.min(100, max + padding)
})

const batteryChartStep = computed(() => {
  const range = batteryChartMax.value - batteryChartMin.value

  if (range <= 10) return 2
  if (range <= 20) return 5
  if (range <= 40) return 10
  return 20
})

const cellDeltaValues = computed(() => {
  return props.cellDeltaHistory
    .map((point) => point.value)
    .filter((value) => typeof value === 'number')
})

const cellDeltaChartMin = computed(() => {
  if (!cellDeltaValues.value.length) return 0

  const min = Math.min(...cellDeltaValues.value)
  const max = Math.max(...cellDeltaValues.value)
  const range = max - min

  const padding = Math.max(10, Math.ceil(range * 0.25))
  return Math.max(0, min - padding)
})

const cellDeltaChartMax = computed(() => {
  if (!cellDeltaValues.value.length) return 300

  const min = Math.min(...cellDeltaValues.value)
  const max = Math.max(...cellDeltaValues.value)
  const range = max - min

  const padding = Math.max(10, Math.ceil(range * 0.25))
  return max + padding
})

const cellDeltaChartStep = computed(() => {
  const range = cellDeltaChartMax.value - cellDeltaChartMin.value

  if (range <= 40) return 5
  if (range <= 80) return 10
  if (range <= 160) return 20
  return 50
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
      <div class="bg-panel rounded-3xl border border-white/10 shadow-custom p-4">
        <div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
          <div class="flex items-center gap-4 min-w-0">
            <div
              class="w-16 h-16 rounded-2xl bg-gradient-to-br from-emerald-500/20 to-cyan-500/10 border border-white/10 flex items-center justify-center shrink-0"
            >
              <img
                :src="batteryIcon"
                alt="battery"
                class="w-20 h-20 object-cover opacity-90"
              />
            </div>

            <div class="min-w-0">
              <div class="text-[10px] uppercase tracking-[0.2em] text-white/50 mb-1">
                UPS HAT (E)
              </div>
              <div class="text-2xl sm:text-3xl font-bold text-white">
                {{ batteryValue }}
              </div>
              <div class="text-sm text-white/50 mt-1 break-words">
                {{ batteryCapacityText }}
              </div>
            </div>
          </div>

          <div class="flex flex-wrap gap-2">
            <div class="rounded-full border border-white/10 bg-white/5 px-3 py-1.5 text-xs text-white/70">
              {{ data.modeText }}
            </div>

            <div class="rounded-full border border-white/10 bg-white/5 px-3 py-1.5 text-xs text-white/70">
              {{ data.powerSourceText }}
            </div>

            <div class="rounded-full border border-white/10 bg-white/5 px-3 py-1.5 text-xs text-white/70">
              {{ data.chargeText }}
            </div>
          </div>
        </div>

        <div class="mt-4">
          <ProgressBar
            :value="batteryPercent"
            :showValue="false"
            class="ups-battery-bar"
          />

          <div class="mt-2 flex items-start justify-between gap-3">
            <div class="text-xs text-white/40">0%</div>

            <div
              class="max-w-[65%] rounded-full border border-violet-400/15 bg-violet-400/10 px-3 py-1 text-[11px] sm:text-xs text-violet-200 text-right leading-tight"
            >
              <span class="text-white/55 mr-1">{{ etaLabel }}:</span>
              <span class="font-medium">{{ etaValue }}</span>
            </div>

            <div class="text-xs text-white/40">100%</div>
          </div>
        </div>
      </div>

      <div class="grid grid-cols-2 gap-3">
        <div class="bg-panel rounded-2xl border border-white/10 shadow-custom p-3">
          <div class="flex items-start justify-between gap-3">
            <div class="min-w-0">
              <div class="text-[10px] uppercase tracking-wide text-white/50 mb-1">
                VBUS
              </div>
              <div class="text-xl sm:text-2xl font-semibold text-white leading-none">
                {{ vbusValue }}
              </div>
            </div>

            <div
              class="w-9 h-9 rounded-xl bg-cyan-400/10 border border-cyan-300/10 flex items-center justify-center text-cyan-300 shrink-0"
            >
              ⚡
            </div>
          </div>

          <div class="mt-3 grid grid-cols-1 sm:grid-cols-2 gap-2">
            <div class="rounded-xl bg-white/5 border border-white/10 px-2 py-1.5">
              <div class="text-[10px] uppercase tracking-wide text-white/40">
                Струм
              </div>
              <div class="text-sm font-medium text-white mt-1 leading-none">
                {{ vbusCurrentText }}
              </div>
            </div>

            <div class="rounded-xl bg-white/5 border border-white/10 px-2 py-1.5">
              <div class="text-[10px] uppercase tracking-wide text-white/40">
                Потужність
              </div>
              <div class="text-sm font-medium text-white mt-1 leading-none">
                {{ vbusPowerText }}
              </div>
            </div>
          </div>
        </div>

        <div class="bg-panel rounded-2xl border border-white/10 shadow-custom p-3">
          <div class="flex items-start justify-between gap-3">
            <div class="min-w-0">
              <div class="text-[10px] uppercase tracking-wide text-white/50 mb-1">
                Батарея
              </div>
              <div class="text-xl sm:text-2xl font-semibold text-white leading-none">
                {{ batteryElectricalValue }}
              </div>
            </div>

            <div
              class="w-9 h-9 rounded-xl bg-emerald-400/10 border border-emerald-300/10 flex items-center justify-center text-emerald-300 shrink-0"
            >
              🔋
            </div>
          </div>

          <div class="mt-3 grid grid-cols-1 sm:grid-cols-2 gap-2">
            <div class="rounded-xl bg-white/5 border border-white/10 px-2 py-1.5">
              <div class="text-[10px] uppercase tracking-wide text-white/40">
                Струм
              </div>
              <div class="text-sm font-medium text-white mt-1 leading-none">
                {{ batteryElectricalCurrent }}
              </div>
            </div>

            <div class="rounded-xl bg-white/5 border border-white/10 px-2 py-1.5">
              <div class="text-[10px] uppercase tracking-wide text-white/40">
                Заряд
              </div>
              <div class="text-sm font-medium text-white mt-1 leading-none">
                {{ batteryValue }}
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="bg-panel rounded-2xl border border-white/10 shadow-custom p-4">
        <div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between mb-4">
          <div>
            <div class="text-[10px] uppercase tracking-wide text-white/50">
              Банки
            </div>
            <div class="text-sm text-white/60 mt-1">
              Напруга по кожній банці
            </div>
          </div>

          <div
            class="inline-flex items-center gap-2 rounded-full border px-3 py-1.5 text-xs"
            :class="deltaSeverityClass"
          >
            <span class="font-medium">Δ {{ data.cellDeltaText }}</span>
          </div>
        </div>

        <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
          <div
            v-for="cell in cells"
            :key="cell.label"
            class="rounded-xl border border-white/10 bg-white/5 p-3"
          >
            <div class="flex items-center justify-between mb-2">
              <span class="text-sm text-white/70">{{ cell.label }}</span>
              <span class="text-sm font-semibold text-white">{{ cell.value }} mV</span>
            </div>

            <ProgressBar
              :value="cell.percent"
              :showValue="false"
              class="ups-cell-bar"
            />
          </div>
        </div>
      </div>

      <div class="space-y-2">
        <div class="px-1 text-[10px] sm:text-xs uppercase tracking-wide text-white/60">
          Історія напруги банок
        </div>

        <MultiLineTrendChart
          title="Напруга по банках"
          subtitle="Усі 4 банки на одному графіку"
          :labels="cellVoltageLabels"
          :datasets="cellVoltageDatasets"
          :min="cellVoltageMin"
          :max="cellVoltageMax"
          :step-size="cellVoltageStep"
          :formatter="(value) => `${value.toFixed(0)} mV`"
        />
      </div>

      <div class="space-y-2">
        <div class="px-1 text-[10px] sm:text-xs uppercase tracking-wide text-white/60">
          Історія UPS
        </div>

        <div class="-mx-4 px-4 overflow-x-auto no-scrollbar">
          <div class="flex gap-4 min-w-max pr-4">
            <div class="w-[320px] sm:w-[420px] lg:w-[480px] shrink-0">
              <MiniTrendChart
                title="Заряд акумулятора"
                subtitle="Останні виміри"
                :points="batteryPercentHistory"
                color="#34D399"
                :min="batteryChartMin"
                :max="batteryChartMax"
                :step-size="batteryChartStep"
                :show-time-axis="true"
                :formatter="(value) => `${value.toFixed(0)}%`"
              />
            </div>

            <div class="w-[320px] sm:w-[420px] lg:w-[480px] shrink-0">
              <MiniTrendChart
                title="Дельта банок"
                subtitle="Різниця між банками"
                :points="cellDeltaHistory"
                color="#F59E0B"
                :min="cellDeltaChartMin"
                :max="cellDeltaChartMax"
                :step-size="cellDeltaChartStep"
                :show-time-axis="true"
                :formatter="(value) => `${value.toFixed(0)} mV`"
              />
            </div>
          </div>
        </div>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-2 gap-3">
        <div class="bg-panel rounded-2xl border border-white/10 shadow-custom p-4">
          <div class="text-[10px] uppercase tracking-wide text-white/50 mb-2">
            Комунікація
          </div>
          <div class="text-base font-medium text-white leading-snug">
            {{ data.commText }}
          </div>
        </div>

        <div class="bg-panel rounded-2xl border border-white/10 shadow-custom p-4">
          <div class="text-[10px] uppercase tracking-wide text-white/50 mb-2">
            Прошивка
          </div>
          <div class="text-2xl font-semibold text-white">
            {{ data.firmwareText }}
          </div>
        </div>
      </div>
    </template>
  </section>
</template>

<style scoped>
:deep(.ups-battery-bar) {
  height: 10px;
  background: rgba(255, 255, 255, 0.06);
  border-radius: 9999px;
  overflow: hidden;
}

:deep(.ups-battery-bar .p-progressbar-value) {
  background: linear-gradient(90deg, #34d399 0%, #22c55e 100%);
  border-radius: 9999px;
}

:deep(.ups-cell-bar) {
  height: 8px;
  background: rgba(255, 255, 255, 0.06);
  border-radius: 9999px;
  overflow: hidden;
}

:deep(.ups-cell-bar .p-progressbar-value) {
  background: linear-gradient(90deg, #60a5fa 0%, #22d3ee 100%);
  border-radius: 9999px;
}
</style>