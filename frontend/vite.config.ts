import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [vue(), tailwindcss()],
  build: {
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (id.includes('node_modules/primevue') || id.includes('node_modules/@primevue')) {
            return 'primevue'
          }
          if (id.includes('node_modules/vue')) {
            return 'vue'
          }
        }
      }
    }
  },
  server: {
    host: '127.0.0.1'
  }
})
