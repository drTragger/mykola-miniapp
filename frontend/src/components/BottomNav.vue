<script setup>
defineProps({
  modelValue: {
    type: String,
    required: true
  }
})

const emit = defineEmits(['update:modelValue'])

const items = [
  { key: 'overview', label: 'Огляд', icon: 'pi-chart-line' },
  { key: 'ups', label: 'UPS', icon: 'pi-bolt' },
  { key: 'qbittorrent', label: 'Торенти', icon: 'pi pi-download' },
  { key: 'system', label: 'Система', icon: 'pi-desktop' },
]
</script>

<template>
  <nav
    class="fixed bottom-0 left-0 right-0 bg-panel border-t border-white/10 z-50"
  >
    <ul class="grid grid-cols-4 gap-2 px-3 py-3 max-w-[920px] mx-auto">
      <li v-for="item in items" :key="item.key">
        <button
          type="button"
          class="flex flex-col items-center justify-center gap-0 w-full py-2.5 rounded-xl transition-colors"
          :class="
            modelValue === item.key
              ? 'bg-primary/20 text-primary shadow-[0_0_16px_rgba(124,131,255,0.18)]'
              : 'text-white/60 hover:bg-white/10'
          "
          @click="emit('update:modelValue', item.key)"
        >
          <i :class="['pi', item.icon, 'text-xl']"></i>
          <span class="text-[11px] sm:text-xs leading-none">{{ item.label }}</span>
        </button>
      </li>
    </ul>
  </nav>
</template>