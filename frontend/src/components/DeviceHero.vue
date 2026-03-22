<script setup>
import { computed } from 'vue'

const props = defineProps({
  title: String,
  subtitle: String,
  status: String,
  uptime: String,
  heroImage: String,
  batteryPercent: {
    type: Number,
    default: null
  },
  refreshing: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['refresh'])

const isOnline = computed(() => props.status === 'Онлайн')

const statusWrapperStyle = computed(() =>
  isOnline.value
    ? {
        backgroundColor: 'rgba(34, 197, 94, 0.18)',
        color: '#86efac',
        border: '1px solid rgba(34, 197, 94, 0.22)'
      }
    : {
        backgroundColor: 'rgba(239, 68, 68, 0.18)',
        color: '#fca5a5',
        border: '1px solid rgba(239, 68, 68, 0.22)'
      }
)

const statusDotStyle = computed(() =>
  isOnline.value
    ? {
        backgroundColor: '#4ade80',
        boxShadow: '0 0 10px rgba(74, 222, 128, 0.8)'
      }
    : {
        backgroundColor: '#f87171',
        boxShadow: '0 0 10px rgba(248, 113, 113, 0.8)'
      }
)

const batteryStyle = computed(() => {
  const p = props.batteryPercent

  if (p === null) return null

  if (p > 60) {
    return {
      backgroundColor: 'rgba(34, 197, 94, 0.15)',
      color: '#86efac',
      border: '1px solid rgba(34, 197, 94, 0.25)'
    }
  }

  if (p > 30) {
    return {
      backgroundColor: 'rgba(234, 179, 8, 0.15)',
      color: '#fde68a',
      border: '1px solid rgba(234, 179, 8, 0.25)'
    }
  }

  return {
    backgroundColor: 'rgba(239, 68, 68, 0.15)',
      color: '#fca5a5',
      border: '1px solid rgba(239, 68, 68, 0.25)'
    }
})
</script>

<template>
  <section
    class="relative overflow-hidden rounded-3xl p-4 sm:p-5 border border-white/10 bg-white/[0.02] backdrop-blur"
  >
    <Button
      :icon="refreshing ? 'pi pi-spin pi-spinner' : 'pi pi-refresh'"
      size="small"
      text
      rounded
      aria-label="Оновити"
      class="!absolute top-3 right-3 z-10"
      @click="emit('refresh')"
    />

    <div class="flex items-center justify-between gap-4">
      <div class="flex items-center gap-3 sm:gap-4 min-w-0">
        <div
          class="w-14 h-14 sm:w-16 sm:h-16 rounded-xl overflow-hidden bg-white/10 border border-white/10 shrink-0"
        >
          <img :src="heroImage" class="w-full h-full object-cover" />
        </div>

        <div class="min-w-0">
          <div class="text-[10px] uppercase tracking-wider text-white/60 mb-1">
            MYKOLA HUB
          </div>

          <h1 class="text-xl sm:text-2xl font-bold text-white truncate">
            {{ title }}
          </h1>

          <p class="text-xs text-white/50 mt-1 truncate">
            {{ subtitle }}
          </p>
        </div>
      </div>

      <div class="hidden sm:flex items-center gap-2 flex-wrap justify-end pr-10">
        <div class="px-3 py-1.5 rounded-full bg-white/5 border border-white/10 text-xs text-white/70">
          ⏱ {{ uptime }}
        </div>

        <div
          v-if="batteryPercent !== null"
          class="flex items-center gap-2 rounded-full px-3 py-1.5 text-xs"
          :style="batteryStyle"
        >
          <span>🔋</span>
          <span class="font-semibold">{{ batteryPercent }}%</span>
        </div>

        <div
          class="flex items-center justify-center gap-2 rounded-full px-3 py-1.5"
          :style="statusWrapperStyle"
        >
          <span class="w-2 h-2 rounded-full" :style="statusDotStyle"></span>
         <span class="text-xs font-semibold text-center whitespace-nowrap">{{ status }}</span>
        </div>
      </div>
    </div>

    <div class="mt-4 flex sm:hidden flex-wrap gap-2 pr-10">
      <div class="px-3 py-1.5 rounded-full bg-white/5 border border-white/10 text-xs text-white/70">
        ⏱ {{ uptime }}
      </div>

      <div
        v-if="batteryPercent !== null"
        class="flex items-center gap-2 rounded-full px-3 py-1.5 text-xs"
        :style="batteryStyle"
      >
        <span>🔋</span>
        <span class="font-semibold">{{ batteryPercent }}%</span>
      </div>

      <div
        class="flex items-center justify-center gap-2 rounded-full px-3 py-1.5"
        :style="statusWrapperStyle"
      >
        <span class="w-2 h-2 rounded-full shrink-0" :style="statusDotStyle"></span>
        <span class="text-xs font-semibold text-center whitespace-nowrap">{{ status }}</span>
      </div>
    </div>
  </section>
</template>