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
  <div
    class="bg-panel rounded-2xl p-4 shadow-custom border border-white/10 flex flex-col"
  >
    <div class="text-xs uppercase tracking-wide text-white/70 mb-2">
      {{ label }}
    </div>
    <ul class="space-y-1">
      <li
        v-for="item in items"
        :key="item.key"
        class="flex justify-between text-sm"
      >
        <span class="text-white/60">{{ item.key }}</span>
        <span class="font-medium text-white">{{ item.value }}</span>
      </li>
    </ul>
  </div>
</template>