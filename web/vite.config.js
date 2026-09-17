import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      // forward the websocket to the Go server on :9000
      '/ws': {
        target: 'ws://localhost:9000',
        ws: true,
      },
    },
  },
})
