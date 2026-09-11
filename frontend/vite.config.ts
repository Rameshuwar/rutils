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
    },
  },
});
