<script setup>
defineProps({
  subtitleLines: {
    type: Array,
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

      <div class="text-sm leading-snug mt-1">
        <div class="text-white/60">
          {{ subtitleLines[0] }}
        </div>

        <div
          v-if="subtitleLines[1]"
          class="text-xs text-white/40 mt-0.5"
        >
          {{ subtitleLines[1] }}
        </div>
      </div>
    </div>
  </div>
</template>