<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import * as echarts from 'echarts'

const props = defineProps({
  label: {
    type: String,
    required: true
  },
  value: {
    type: Number,
    required: true
  },
  suffix: {
    type: String,
    default: '%'
  },
  subtitle: {
    type: String,
    default: ''
  },
  color: {
    type: String,
    default: '#7c83ff'
  }
})

const chartEl = ref(null)
let chartInstance = null

const clampedValue = computed(() => {
  if (typeof props.value !== 'number' || Number.isNaN(props.value)) return 0
  return Math.max(0, Math.min(100, props.value))
})

function renderChart() {
  if (!chartEl.value) return

  if (!chartInstance) {
    chartInstance = echarts.init(chartEl.value)
  }

  chartInstance.setOption({
    animation: true,
    backgroundColor: 'transparent',
    series: [
      {
        type: 'pie',
        radius: ['72%', '90%'],
        center: ['50%', '50%'],
        silent: true,
        label: { show: false },
        data: [
          {
            value: clampedValue.value,
            itemStyle: {
              color: props.color,
              shadowBlur: 18,
              shadowColor: props.color
            }
          },
          {
            value: 100 - clampedValue.value,
            itemStyle: {
              color: 'rgba(255,255,255,0.08)'
            }
          }
        ]
      }
    ]
  })
}

function handleResize() {
  chartInstance?.resize()
}

watch(() => props.value, renderChart)

onMounted(() => {
  renderChart()
  window.addEventListener('resize', handleResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  chartInstance?.dispose()
  chartInstance = null
})
</script>

<template>
  <div class="metric-donut-card">
    <div class="metric-donut-header">
      <div class="metric-donut-label">{{ label }}</div>
      <div class="metric-donut-subtitle">{{ subtitle }}</div>
    </div>

    <div class="metric-donut-body">
      <div ref="chartEl" class="metric-donut-chart"></div>

      <div class="metric-donut-center">
        <div class="metric-donut-value">{{ clampedValue.toFixed(1) }}{{ suffix }}</div>
      </div>
    </div>
  </div>
</template>