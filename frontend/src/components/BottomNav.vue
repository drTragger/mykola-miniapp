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
  { key: 'system', label: 'Система', icon: 'pi-desktop' },
]
</script>

<template>
  <nav
    class="fixed bottom-0 left-0 right-0 bg-panel/80 backdrop-blur-md border-t border-white/10 z-50"
  >
    <ul class="grid grid-cols-3 gap-1 px-3 py-2 max-w-[920px] mx-auto">
      <li v-for="item in items" :key="item.key">
        <button
          type="button"
          class="flex flex-col items-center justify-center gap-1 w-full py-2 rounded-xl transition-colors"
          :class="
            modelValue === item.key
              ? 'bg-primary/20 text-primary'
              : 'text-white/60 hover:bg-white/10'
          "
          @click="emit('update:modelValue', item.key)"
        >
          <i :class="['pi', item.icon, 'text-lg']"></i>
          <span class="text-[11px] sm:text-xs leading-none">{{ item.label }}</span>
        </button>
      </li>
    </ul>
  </nav>
</template>