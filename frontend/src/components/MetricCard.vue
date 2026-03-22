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
  },
  compact: {
    type: Boolean,
    default: true
  }
})
</script>

<template>
  <div
    class="bg-panel rounded-2xl border border-white/10 shadow-custom flex flex-col"
    :class="compact ? 'p-3 min-h-[92px]' : 'p-4 min-h-[120px]'"
  >
    <div class="text-[10px] sm:text-xs uppercase tracking-wide text-white/60">
      {{ label }}
    </div>

    <div
      class="mt-1 text-white font-semibold leading-tight"
      :class="compact ? 'text-lg sm:text-xl' : 'text-xl sm:text-2xl'"
    >
      {{ value }}
    </div>

    <div v-if="subvalue" class="mt-1 text-xs sm:text-sm text-white/50 leading-snug">
      {{ subvalue }}
    </div>
  </div>
</template>