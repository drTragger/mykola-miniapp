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
  <div class="app-topbar">
    <div>
      <div class="app-topbar-title">Mykola Hub</div>
      <div class="app-topbar-subtitle">{{ subtitle }}</div>
    </div>
  </div>
</template>