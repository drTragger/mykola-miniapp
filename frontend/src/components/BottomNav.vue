<script setup>
// Receive the currently active tab from v-model
defineProps({
  modelValue: {
    type: String,
    required: true
  }
})

const emit = defineEmits(['update:modelValue'])

// Define navigation items with PrimeIcons
const items = [
  { key: 'overview', label: 'Огляд', icon: 'pi-chart-line' },
  { key: 'system', label: 'Система', icon: 'pi-desktop' },
  { key: 'network', label: 'Мережа', icon: 'pi-sitemap' }
]
</script>

<template>
  <nav
    class="fixed bottom-0 left-0 right-0 bg-panel/80 backdrop-blur-md border-t border-white/10 z-50"
  >
    <ul class="flex justify-between px-4 py-2">
      <li v-for="item in items" :key="item.key" class="flex-1 px-1">
        <button
          class="flex flex-col items-center justify-center gap-1 w-full py-2 rounded-xl transition-colors"
          :class="
            modelValue === item.key
              ? 'bg-primary/20 text-primary'
              : 'text-white/60 hover:bg-white/10'
          "
          @click="emit('update:modelValue', item.key)"
        >
          <i :class="['pi', item.icon, 'text-xl']"></i>
          <span class="text-xs">{{ item.label }}</span>
        </button>
      </li>
    </ul>
  </nav>
</template>