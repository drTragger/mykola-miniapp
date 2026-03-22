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
  <section class="grid-single">
    <InfoListCard label="Мережа" :items="items" />

    <div class="footer-note">
      Далі сюди можна додати ping, RX/TX, зовнішній IP, статус Jellyfin, qBittorrent, Sonarr і решту сервісів.
    </div>
  </section>
</template>