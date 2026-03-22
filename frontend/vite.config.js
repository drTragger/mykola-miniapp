import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'node:path'

export default defineConfig({
  plugins: [vue()],
  base: '/',
  build: {
    outDir: resolve(__dirname, '../internal/web/dist'),
    emptyOutDir: true
  },
  server: {
    host: '0.0.0.0',
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://192.168.110.221:8090',
        changeOrigin: true
      }
    }
  }
})