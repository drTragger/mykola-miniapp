<script setup>
import { computed } from 'vue'

const props = defineProps({
  title: String,
  subtitle: String,
  status: String,
  uptime: String,
  heroImage: String
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
</script>

<template>
  <section
    class="relative overflow-hidden rounded-3xl p-5 border border-white/10 bg-white/[0.02] backdrop-blur"
  >
    <div class="flex flex-col gap-4 sm:flex-row sm:justify-between sm:items-center">
      
      <!-- LEFT -->
      <div class="flex items-center gap-4">
        <div
          class="w-16 h-16 rounded-xl overflow-hidden bg-white/10 border border-white/10"
        >
          <img :src="heroImage" class="w-full h-full object-cover" />
        </div>

        <div>
          <div class="text-[10px] uppercase tracking-wider text-white/60 mb-1">
            MYKOLA HUB
          </div>

          <h1 class="text-2xl font-bold text-white">
            {{ title }}
          </h1>

          <p class="text-xs text-white/50 mt-1">
            {{ subtitle }}
          </p>
        </div>
      </div>

      <!-- RIGHT -->
      <div class="flex items-center gap-2 flex-wrap sm:flex-nowrap">
        
        <!-- UPTIME -->
        <div class="px-3 py-1.5 rounded-full bg-white/5 border border-white/10 text-xs text-white/70">
          ⏱ {{ uptime }}
        </div>

        <!-- REFRESH -->
        <Button
          icon="pi pi-refresh"
          size="small"
          text
          rounded
          @click="emit('refresh')"
        />

        <!-- STATUS -->
        <div
          class="flex items-center gap-2 rounded-full px-3 py-1.5"
          :style="statusWrapperStyle"
        >
          <span
            class="w-2 h-2 rounded-full"
            :style="statusDotStyle"
          ></span>

          <span class="text-xs font-semibold">
            {{ status }}
          </span>
        </div>

      </div>
    </div>
  </section>
</template>