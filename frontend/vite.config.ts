import { defineConfig } from 'vite';
import tailwindcss from '@tailwindcss/vite';

export default defineConfig({
  plugins: [
    tailwindcss(),
  ],
  server: {
    proxy: {
      '/convert': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/calculate': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/extract': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/auth': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/repair': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/nifty50': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/market': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/compress-image': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
});
