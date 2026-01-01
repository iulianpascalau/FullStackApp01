import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    allowedHosts: ['app.jls-software.net'],
    proxy: {
      '^/(login|register|change-password|counter|version)': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  }
})
