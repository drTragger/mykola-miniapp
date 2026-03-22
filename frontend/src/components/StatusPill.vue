<script setup>
const props = defineProps({
  label: {
    type: String,
    required: true
  },
  ok: {
    type: Boolean,
    default: false
  }
})
</script>

<template>
  <div
    class="inline-flex items-center gap-2 rounded-full px-3 py-1.5 text-xs border"
    :class="
      ok
        ? 'bg-green-500/10 text-green-300 border-green-500/20'
        : 'bg-red-500/10 text-red-300 border-red-500/20'
    "
  >
    <span
      class="w-2 h-2 rounded-full"
      :class="ok ? 'bg-green-400 shadow-[0_0_10px_rgba(74,222,128,0.8)]' : 'bg-red-400 shadow-[0_0_10px_rgba(248,113,113,0.8)]'"
    ></span>
    <span class="font-medium">{{ label }}</span>
  </div>
</template>