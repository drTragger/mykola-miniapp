<script setup>
import InfoListCard from '../components/InfoListCard.vue'

const props = defineProps({
  metrics: {
    type: Object,
    required: true
  }
})

const items = [
  { key: 'Local IPv4', value: props.metrics.network.localIpv4 || '—' }
]
</script>

<template>
  <section class="section active">
    <div class="grid-single">
      <InfoListCard label="Мережа" :items="items" />
    </div>

    <div class="footer-note">
      Далі сюди можна додати ping, RX/TX, зовнішній IP, статус Jellyfin, qBittorrent, Sonarr і решту сервісів.
    </div>
  </section>
</template>