import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import fs from 'node:fs'

// emptyOutDir 会清空 embed 目录，构建后补回 .gitkeep 占位，
// 保证全新克隆（未构建前端）也能通过 go:embed 编译
function keepEmbedDir() {
  return {
    name: 'keep-embed-dir',
    closeBundle() {
      fs.writeFileSync(
        new URL('../backend/internal/web/dist/.gitkeep', import.meta.url),
        '占位文件：本目录由 frontend 的 pnpm build 生成。\n',
      )
    },
  }
}

export default defineConfig({
  plugins: [vue(), keepEmbedDir()],
  build: {
    // 直接输出到 Go 的 embed 目录，由 go:embed 打进二进制
    outDir: '../backend/internal/web/dist',
    emptyOutDir: true,
  },
  server: {
    port: 5173,
    proxy: {
      // 开发时将 API 代理到 Go 后端
      '/api': 'http://localhost:8080',
    },
  },
})
