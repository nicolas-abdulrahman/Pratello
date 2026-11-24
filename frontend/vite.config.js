import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import tailwindcss from '@tailwindcss/vite';
import { fileURLToPath } from 'node:url';

// https://vitejs.dev/config/
export default defineConfig({
  envDir: fileURLToPath(new URL('..', import.meta.url)),
  envPrefix: ['VITE_', 'TEST_ENV'],
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
