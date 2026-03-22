<script setup>
defineProps({
  label: {
    type: String,
    required: true
  },
  items: {
    type: Array,
    required: true
  }
})
</script>

<template>
  <div class="card">
    <div class="card-label">{{ label }}</div>
    <div class="info-list">
      <div v-for="item in items" :key="item.key" class="info-row">
        <div class="info-key">{{ item.key }}</div>
        <div class="info-value">{{ item.value }}</div>
      </div>
    </div>
  </div>
</template>