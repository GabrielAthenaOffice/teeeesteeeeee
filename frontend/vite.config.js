import react from '@vitejs/plugin-react'
import { defineConfig, loadEnv } from 'vite'

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')

  // Porta real do backend-golang (backend-golang/.env => PORT).
  const API_TARGET = env.VITE_API_PROXY_TARGET || 'http://localhost:8082'

  return {
    plugins: [react()],
    server: {
      port: 5173,
      // Evita CORS: /api e /health são encaminhados para a API em dev.
      proxy: {
        '/api': { target: API_TARGET, changeOrigin: true },
        '/health': { target: API_TARGET, changeOrigin: true },
      },
    },
  }
})
