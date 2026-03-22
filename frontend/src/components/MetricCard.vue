<script setup>
defineProps({
  label: {
    type: String,
    required: true
  },
  value: {
    type: String,
    required: true
  },
  subvalue: {
    type: String,
    default: ''
  }
})
</script>

<template>
  <div class="metric-card">
    <div class="metric-card-label">{{ label }}</div>
    <div class="metric-card-value">{{ value }}</div>
    <div class="metric-card-subvalue">{{ subvalue }}</div>
  </div>
</template>