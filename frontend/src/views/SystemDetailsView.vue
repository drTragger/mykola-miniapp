<script setup>
import { computed } from 'vue'
import InfoListCard from '../components/InfoListCard.vue'

const props = defineProps({
  systemData: {
    type: Object,
    required: true
  }
})

const systemItems = computed(() => [
  { key: 'Хост', value: props.systemData.system?.hostname || '—' },
  { key: 'Платформа', value: props.systemData.system?.platform || '—' },
  { key: 'Версія ОС', value: props.systemData.system?.platformVersion || '—' },
  { key: 'Ядро', value: props.systemData.system?.kernelVersion || '—' },
  { key: 'Архітектура', value: props.systemData.system?.architecture || '—' },
  { key: 'Процеси', value: String(props.systemData.system?.processes ?? '—') },
  { key: 'Час запуску', value: props.systemData.system?.bootTimeUnix ? new Date(props.systemData.system.bootTimeUnix * 1000).toLocaleString() : '—' }
])

const cpuItems = computed(() => [
  { key: 'Модель CPU', value: props.systemData.system?.cpuModel || '—' },
  { key: 'Частота', value: props.systemData.system?.cpuFrequencyMHz ? `${props.systemData.system.cpuFrequencyMHz.toFixed(0)} MHz` : '—' },
  { key: 'Ядер (логічних)', value: String(props.systemData.system?.logicalCpuCount ?? '—') }
])

const loadItems = computed(() => [
  { key: '1 хв', value: props.systemData.system?.load1?.toFixed?.(2) ?? '—' },
  { key: '5 хв', value: props.systemData.system?.load5?.toFixed?.(2) ?? '—' },
  { key: '15 хв', value: props.systemData.system?.load15?.toFixed?.(2) ?? '—' }
])

const networkItems = computed(() => [
  { key: 'Локальний IP', value: props.systemData.network?.localIpv4 || '—' },
  { key: 'Публічний IP', value: props.systemData.network?.publicIp || '—' },
  { key: 'Пінг', value: typeof props.systemData.network?.pingMs === 'number' ? `${props.systemData.network.pingMs.toFixed(1)} ms` : '—' },
  { key: 'RX (всього)', value: props.systemData.network?.rxTotalHuman || '—' },
  { key: 'TX (всього)', value: props.systemData.network?.txTotalHuman || '—' },
  { key: 'RX швидкість', value: props.systemData.network?.rxSpeedHuman || '—' },
  { key: 'TX швидкість', value: props.systemData.network?.txSpeedHuman || '—' }
])

const vpn = computed(() => props.systemData.vpn || {})
</script>

<template>
  <section class="space-y-4">
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
      <InfoListCard label="Система" :items="systemItems" />
      <InfoListCard label="CPU" :items="cpuItems" />
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
      <InfoListCard label="Навантаження (Load Average)" :items="loadItems" />
      <InfoListCard label="Мережа" :items="networkItems" />
    </div>

    <div class="bg-panel rounded-2xl p-4 shadow-custom border border-white/10 space-y-4">
      <div class="text-[10px] uppercase tracking-wide text-white/60">
        VPN / WireGuard
      </div>

      <div class="space-y-2">
        <div class="text-xs text-white/40">WireGuard</div>

        <div class="grid grid-cols-1 sm:grid-cols-2 gap-2 text-sm">
          <div>Статус: <span class="text-white">{{ vpn.ok ? '✅ Активний' : '❌ Неактивний' }}</span></div>
          <div>IP: <span class="text-white">{{ vpn.wgIp || '—' }}</span></div>
          <div>Endpoint: <span class="text-white">{{ vpn.endpoint || '—' }}</span></div>
          <div>Handshake: <span class="text-white">{{ vpn.lastHandshakeAgo || '—' }}</span></div>
          <div>Отримано: <span class="text-white">{{ vpn.rx || '—' }}</span></div>
          <div>Відправлено: <span class="text-white">{{ vpn.tx || '—' }}</span></div>
        </div>
      </div>

      <div class="space-y-2">
        <div class="text-xs text-white/40">qBittorrent</div>

        <div class="grid grid-cols-1 sm:grid-cols-2 gap-2 text-sm">
          <div>Сервіс: <span class="text-white">{{ vpn.qbit?.serviceOk ? '✅' : '❌' }}</span></div>
          <div>User: <span class="text-white">{{ vpn.qbit?.user || '—' }}</span></div>
          <div>Інтерфейс: <span class="text-white">{{ vpn.qbit?.binding || '—' }}</span></div>
          <div>Web UI: <span class="text-white">{{ vpn.qbit?.webui || '—' }}</span></div>
        </div>
      </div>

      <div class="space-y-2">
        <div class="text-xs text-white/40">Routing</div>

        <div class="grid grid-cols-1 sm:grid-cols-2 gap-2 text-sm">
          <div>ip rule: <span class="text-white">{{ vpn.ruleOk ? '✅' : '❌' }}</span></div>
          <div>Route через wg0: <span class="text-white">{{ vpn.routeOk ? '✅' : '❌' }}</span></div>
          <div class="col-span-2">
            Таблиця:
            <span class="text-white block break-all mt-1">
              {{ vpn.routeTable || '—' }}
            </span>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>