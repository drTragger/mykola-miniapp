<script setup>
defineProps({
  modelValue: {
    type: String,
    required: true
  }
})

const emit = defineEmits(['update:modelValue'])

const tabs = [
  { key: 'overview', label: 'Огляд' },
  { key: 'system', label: 'Система' },
  { key: 'network', label: 'Мережа' }
]
</script>

<template>
  <div class="toolbar">
    <button
      v-for="tab in tabs"
      :key="tab.key"
      class="tab"
      :class="{ active: modelValue === tab.key }"
      @click="emit('update:modelValue', tab.key)"
    >
      {{ tab.label }}
    </button>
  </div>
</template>