<script setup>
import { computed } from 'vue'
import Chart from 'primevue/chart'

const props = defineProps({
  title: {
    type: String,
    required: true
  },
  subtitle: {
    type: String,
    default: ''
  },
  points: {
    type: Array,
    default: () => []
  },
  color: {
    type: String,
    default: '#7c83ff'
  },
  formatter: {
    type: Function,
    default: (value) => String(value)
  }
})

function hexToRgba(hex, alpha) {
  let h = hex.replace('#', '')

  if (h.length === 3) {
    h = h
      .split('')
      .map((ch) => ch + ch)
      .join('')
  }

  const bigint = parseInt(h, 16)
  const r = (bigint >> 16) & 255
  const g = (bigint >> 8) & 255
  const b = bigint & 255

  return `rgba(${r}, ${g}, ${b}, ${alpha})`
}

const labels = computed(() => props.points.map((p) => p.time))
const values = computed(() => props.points.map((p) => p.value))

const chartData = computed(() => ({
  labels: labels.value,
  datasets: [
    {
      data: values.value,
      fill: true,
      borderColor: props.color,
      backgroundColor: hexToRgba(props.color, 0.12),
      tension: 0.35,
      borderWidth: 3,
      pointRadius: 0,
      pointHoverRadius: 3,
      clip: 0
    }
  ]
}))

const chartOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  animation: false,
  layout: {
    padding: {
      top: 6,
      right: 6,
      bottom: 0,
      left: 6
    }
  },
  plugins: {
    legend: { display: false },
    tooltip: {
      displayColors: false,
      callbacks: {
        label: (context) => props.formatter(context.parsed.y)
      }
    }
  },
  scales: {
    x: {
      display: false,
      grid: { display: false },
      border: { display: false }
    },
    y: {
      display: false,
      grid: { display: false },
      border: { display: false }
    }
  }
}))
</script>

<template>
  <div
    class="bg-panel rounded-2xl p-4 shadow-custom border border-white/10 flex flex-col overflow-hidden min-h-[180px]"
  >
    <div class="text-[10px] sm:text-xs uppercase tracking-wide text-white/70 mb-1">
      {{ title }}
    </div>

    <div class="text-[10px] sm:text-xs text-white/50 mb-3">
      {{ subtitle }}
    </div>

    <div class="h-[110px] sm:h-[130px] relative overflow-hidden">
      <Chart
        type="line"
        :data="chartData"
        :options="chartOptions"
        class="w-full h-full"
      />
    </div>
  </div>
</template>