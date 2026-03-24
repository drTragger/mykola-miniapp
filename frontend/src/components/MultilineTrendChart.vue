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
  labels: {
    type: Array,
    default: () => []
  },
  datasets: {
    type: Array,
    default: () => []
  },
  min: {
    type: Number,
    default: null
  },
  max: {
    type: Number,
    default: null
  },
  stepSize: {
    type: Number,
    default: null
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

const chartData = computed(() => ({
  labels: props.labels,
  datasets: props.datasets.map((dataset) => ({
    label: dataset.label,
    data: dataset.data,
    fill: false,
    borderColor: dataset.color,
    backgroundColor: hexToRgba(dataset.color, 0.12),
    tension: 0.35,
    borderWidth: 2,
    pointRadius: 0,
    pointHoverRadius: 3,
    pointHitRadius: 10,
    clip: 0
  }))
}))

const chartOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  animation: false,
  layout: {
    padding: {
      top: 4,
      right: 4,
      bottom: 0,
      left: 4
    }
  },
  plugins: {
    legend: {
      display: true,
      position: 'top',
      labels: {
        color: 'rgba(255,255,255,0.70)',
        boxWidth: 10,
        boxHeight: 10,
        usePointStyle: true,
        pointStyle: 'circle',
        font: {
          size: 11
        }
      }
    },
    tooltip: {
      displayColors: true,
      callbacks: {
        title: (items) => items?.[0]?.label || '',
        label: (context) => `${context.dataset.label}: ${props.formatter(context.parsed.y)}`
      }
    }
  },
  scales: {
    x: {
      display: true,
      ticks: {
        color: 'rgba(255,255,255,0.38)',
        font: {
          size: 10
        },
        maxRotation: 0,
        autoSkip: true,
        maxTicksLimit: 6
      },
      grid: {
        display: false
      },
      border: {
        display: false
      }
    },
    y: {
      display: true,
      min: props.min ?? undefined,
      max: props.max ?? undefined,
      ticks: {
        color: 'rgba(255,255,255,0.38)',
        font: {
          size: 10
        },
        stepSize: props.stepSize ?? undefined,
        callback: (value) => props.formatter(Number(value))
      },
      grid: {
        color: 'rgba(255,255,255,0.06)',
        drawBorder: false
      },
      border: {
        display: false
      }
    }
  }
}))
</script>

<template>
  <div class="bg-panel rounded-2xl p-4 shadow-custom border border-white/10 flex flex-col overflow-hidden min-h-[290px]">
    <div class="text-[10px] sm:text-xs uppercase tracking-wide text-white/70 mb-1">
      {{ title }}
    </div>

    <div class="text-[10px] sm:text-xs text-white/50 mb-3">
      {{ subtitle }}
    </div>

    <div class="h-[220px] sm:h-[260px] relative overflow-hidden">
      <Chart
        type="line"
        :data="chartData"
        :options="chartOptions"
        class="w-full h-full"
      />
    </div>
  </div>
</template>