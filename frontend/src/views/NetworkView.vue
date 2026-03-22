<script setup>
import { computed } from 'vue'
import InfoListCard from '../components/InfoListCard.vue'

const props = defineProps({
  metrics: {
    type: Object,
    required: true
  }
})

const items = computed(() => [
  { key: 'Local IPv4', value: props.metrics.network?.localIpv4 || '—' }
])
</script>

<template>
  <section class="space-y-4">
    <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
      <InfoListCard label="Мережа" :items="items" />
    </div>
    <p class="text-sm text-white/60">
      Далі сюди можна додати ping, RX/TX, зовнішній IP, статус Jellyfin, qBittorrent, Sonarr і решту сервісів.
    </p>
  </section>
</template>