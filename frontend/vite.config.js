import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    port: 5173,
    proxy: {
      // 开发时将 API 代理到 Go 后端
      '/api': 'http://localhost:8080',
    },
  },
})
