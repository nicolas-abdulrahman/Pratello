import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import tailwindcss from '@tailwindcss/vite';

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    tailwindcss(),
    svelte()
  ],
  server: {
    port: 5173,
    proxy: {
      // Proxy API requests to Go backend during local frontend dev
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/media': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
});
