import { mount } from 'svelte';
import './styles/app.css';
import App from './App.svelte';

const BACKEND_URL = (import.meta.env.VITE_API_URL || '').replace(/\/$/, '');
if (BACKEND_URL) {
  const originalFetch = window.fetch;
  window.fetch = (input, init) => {
    if (typeof input === 'string' && (input.startsWith('/api') || input.startsWith('/media'))) {
      input = `${BACKEND_URL}${input}`;
    }
    return originalFetch(input, init);
  };
}

const app = mount(App, {
  target: document.getElementById('app'),
});

export default app;
