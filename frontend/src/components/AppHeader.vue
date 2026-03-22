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
  <div class="flex items-center justify-between mb-4">
    <div>
      <div class="text-sm font-bold uppercase tracking-wider text-white/60">
        Mykola Hub
      </div>
      <div class="mt-1 text-sm text-white/50">
        {{ subtitle }}
      </div>
    </div>
  </div>
</template>