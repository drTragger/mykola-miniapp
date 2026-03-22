<script setup>
defineProps({
  modelValue: {
    type: String,
    required: true
  }
})

const emit = defineEmits(['update:modelValue'])

const items = [
  { key: 'overview', label: 'Overview', icon: '◉' },
  { key: 'system', label: 'System', icon: '▣' },
  { key: 'network', label: 'Network', icon: '◎' },
  { key: 'ups', label: 'UPS', icon: '◌' }
]
</script>

<template>
  <nav class="bottom-nav">
    <button
      v-for="item in items"
      :key="item.key"
      class="bottom-nav-item"
      :class="{ active: modelValue === item.key }"
      @click="emit('update:modelValue', item.key)"
    >
      <span class="bottom-nav-icon">{{ item.icon }}</span>
      <span class="bottom-nav-label">{{ item.label }}</span>
    </button>
  </nav>
</template>