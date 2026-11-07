import path from 'node:path'
import { fileURLToPath } from 'node:url'
import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

const projectDir = path.dirname(fileURLToPath(import.meta.url))

export default defineConfig({
  base: '/arsy/',
  plugins: [react(), tailwindcss()],
  server: {
    proxy: {
      '/api': process.env.ARSY_API_TARGET || 'http://127.0.0.1:8090',
    },
  },
  build: {
    outDir: path.resolve(projectDir, '../arsy/dashboard/dist'),
    emptyOutDir: true,
  },
  resolve: {
    alias: {
      '@': path.resolve(projectDir, 'src'),
    },
  },
})
