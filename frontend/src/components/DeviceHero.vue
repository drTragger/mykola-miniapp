<script setup>
defineProps({
  title: {
    type: String,
    required: true
  },
  subtitle: {
    type: String,
    required: true
  },
  status: {
    type: String,
    required: true
  },
  hostname: {
    type: String,
    default: '—'
  },
  uptime: {
    type: String,
    default: '—'
  },
  localIp: {
    type: String,
    default: '—'
  },
  heroImage: {
    type: String,
    default: '/hero.png'
  }
})
</script>

<template>
  <section class="hero-card">
    <div class="hero-content">
      <div class="hero-topline">
        <div class="hero-device">
          <div class="hero-avatar">
            <img :src="heroImage" alt="Mykola device" />
          </div>

          <div class="hero-text">
            <div class="hero-kicker">MYKOLA HUB</div>
            <h1 class="hero-title">{{ title }}</h1>
            <p class="hero-subtitle">{{ subtitle }}</p>
          </div>
        </div>

        <div class="hero-status" :class="status === 'Онлайн' ? 'is-online' : ''">
          <span class="hero-status-dot"></span>
          {{ status }}
        </div>
      </div>

      <div class="hero-quick-stats">
        <div class="hero-pill">
          <span class="hero-pill-label">Hostname</span>
          <span class="hero-pill-value">{{ hostname }}</span>
        </div>

        <div class="hero-pill">
          <span class="hero-pill-label">Uptime</span>
          <span class="hero-pill-value">{{ uptime }}</span>
        </div>

        <div class="hero-pill">
          <span class="hero-pill-label">IP</span>
          <span class="hero-pill-value">{{ localIp }}</span>
        </div>
      </div>
    </div>
  </section>
</template>