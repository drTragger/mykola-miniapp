<script setup>
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import * as echarts from 'echarts'

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

const chartEl = ref(null)
let chartInstance = null

function renderChart() {
  if (!chartEl.value) return

  if (!chartInstance) {
    chartInstance = echarts.init(chartEl.value)
  }

  chartInstance.setOption({
    backgroundColor: 'transparent',
    grid: {
      top: 18,
      left: 8,
      right: 8,
      bottom: 8,
      containLabel: false
    },
    tooltip: {
      trigger: 'axis',
      backgroundColor: '#111827',
      borderColor: 'rgba(255,255,255,0.08)',
      textStyle: {
        color: '#fff'
      },
      formatter: (params) => {
        const point = params?.[0]
        if (!point) return ''
        return `${point.axisValue}<br/>${props.formatter(point.data)}`
      }
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: props.points.map((item) => item.time),
      axisLine: { show: false },
      axisTick: { show: false },
      axisLabel: { show: false }
    },
    yAxis: {
      type: 'value',
      axisLine: { show: false },
      axisTick: { show: false },
      splitLine: {
        lineStyle: {
          color: 'rgba(255,255,255,0.05)'
        }
      },
      axisLabel: { show: false }
    },
    series: [
      {
        type: 'line',
        smooth: true,
        symbol: 'none',
        data: props.points.map((item) => item.value),
        lineStyle: {
          width: 3,
          color: props.color
        },
        areaStyle: {
          opacity: 0.16,
          color: props.color
        }
      }
    ]
  })
}

function handleResize() {
  chartInstance?.resize()
}

watch(() => props.points, renderChart, { deep: true })

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
  <div class="trend-card">
    <div class="trend-card-header">
      <div class="trend-card-title">{{ title }}</div>
      <div class="trend-card-subtitle">{{ subtitle }}</div>
    </div>

    <div ref="chartEl" class="trend-chart"></div>
  </div>
</template>