import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  build: {
    outDir: 'dist',
    sourcemap: true,
    rollupOptions: {
      output: {
        manualChunks: (id) => {
          if (id.includes('chart.js') || id.includes('react-chartjs-2')) {
            return 'chartjs';
          }
          if (id.includes('cytoscape') || id.includes('react-cytoscapejs')) {
            return 'cytoscape';
          }
          return undefined;
        },
      },
    },
  },
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./src/test/setup.ts'],
    include: ['src/**/*.{test,spec}.{js,ts,tsx}'],
    exclude: ['node_modules', 'dist'],
  },
  server: {
    port: 7575,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/health': 'http://localhost:8080',
      '/configuration': 'http://localhost:8080',
      '/projects': 'http://localhost:8080',
      '/search': 'http://localhost:8080',
      '/relationships': 'http://localhost:8080',
      '/statistics': 'http://localhost:8080',
    },
  },
})
