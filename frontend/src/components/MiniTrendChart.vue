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
  showTimeAxis: {
    type: Boolean,
    default: false
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
      backgroundColor: hexToRgba(props.color, 0.10),
      tension: 0.35,
      borderWidth: 2,
      pointRadius: 0,
      pointHoverRadius: 3,
      pointHitRadius: 10,
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
      top: 2,
      right: 2,
      bottom: 0,
      left: 2
    }
  },
  plugins: {
    legend: { display: false },
    tooltip: {
      displayColors: false,
      callbacks: {
        title: (items) => items?.[0]?.label || '',
        label: (context) => props.formatter(context.parsed.y)
      }
    }
  },
  scales: {
    x: {
      display: props.showTimeAxis,
      ticks: {
        color: 'rgba(255,255,255,0.38)',
        font: {
          size: 10
        },
        maxRotation: 0,
        autoSkip: true,
        maxTicksLimit: 5
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
  <div
    class="bg-panel rounded-2xl p-3 shadow-custom border border-white/10 flex flex-col overflow-hidden min-h-[210px]"
  >
    <div class="text-[10px] sm:text-xs uppercase tracking-wide text-white/70 mb-1">
      {{ title }}
    </div>

    <div class="text-[10px] sm:text-xs text-white/50 mb-2">
      {{ subtitle }}
    </div>

    <div class="h-[145px] sm:h-[165px] relative overflow-hidden">
      <Chart
        type="line"
        :data="chartData"
        :options="chartOptions"
        class="w-full h-full"
      />
    </div>
  </div>
</template>