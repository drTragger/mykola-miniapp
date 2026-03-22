<script setup>
import { computed } from 'vue'
import Knob from 'primevue/knob'

const props = defineProps({
  label: {
    type: String,
    required: true
  },
  value: {
    type: Number,
    required: true
  },
  suffix: {
    type: String,
    default: '%'
  },
  color: {
    type: String,
    default: '#7c83ff'
  },
  compact: {
    type: Boolean,
    default: true
  }
})

const clampedValue = computed(() => {
  if (typeof props.value !== 'number' || Number.isNaN(props.value)) return 0
  return Math.max(0, Math.min(100, props.value))
})

const knobSize = computed(() => (props.compact ? 84 : 120))
const valueFontSize = computed(() => (props.compact ? '0.95rem' : '1.15rem'))
</script>

<template>
  <div class="flex flex-col items-center justify-start rounded-xl px-1 py-1">
    <div class="text-[10px] sm:text-xs uppercase tracking-wide text-white/70 text-center mb-1">
      {{ label }}
    </div>

    <Knob
      :model-value="clampedValue"
      :min="0"
      :max="100"
      :size="knobSize"
      :stroke-width="10"
      :value-template="`${clampedValue.toFixed(1)}${suffix}`"
      :value-color="color"
      range-color="rgba(255,255,255,0.08)"
      text-color="#ffffff"
      readonly
      class="donut-knob"
      :style="{ '--knob-value-font-size': valueFontSize }"
    />
  </div>
</template>

<style scoped>
:deep(.p-knob-value) {
  stroke-linecap: round;
}

:deep(.p-knob-text) {
  font-weight: 700;
  font-size: var(--knob-value-font-size, 1rem);
}
</style>