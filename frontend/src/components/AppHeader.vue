<script setup>
defineProps({
  subtitle: {
    type: String,
    required: true
  },
  status: {
    type: String,
    required: true
  }
})
</script>

<template>
  <div class="header">
    <div class="title-row">
      <div class="title">Mykola Hub</div>
      <div class="status-chip">{{ status }}</div>
    </div>
    <div class="subtitle">{{ subtitle }}</div>
  </div>
</template>