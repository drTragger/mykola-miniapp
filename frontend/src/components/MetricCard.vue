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
  <div class="card">
    <div class="card-label">{{ label }}</div>
    <div class="card-value">{{ value }}</div>
    <div class="card-subvalue">{{ subvalue }}</div>
  </div>
</template>