/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{svelte,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        background: 'var(--color-background)',
        text: 'var(--color-text)',
        primary: {
          DEFAULT: 'var(--color-primary)',
          hover: 'var(--color-primary-hover)',
          soft: 'var(--color-primary-soft)',
        },
        surface: {
          DEFAULT: 'var(--color-surface)',
          hover: 'var(--color-surface-hover)',
        },
        muted: 'var(--color-text-muted)',
        border: 'var(--color-border)',
        brand: {
          DEFAULT: 'var(--color-brand-primary)',
          hover: 'var(--color-brand-primary-hover)',
          soft: 'var(--color-brand-primary-soft)',
          bg: 'var(--color-brand-bg)',
          surface: 'var(--color-brand-surface)',
          'surface-hover': 'var(--color-brand-surface-hover)',
          text: 'var(--color-brand-text)',
          muted: 'var(--color-brand-text-muted)',
          border: 'var(--color-brand-border)',
        },
        restaurant: {
          DEFAULT: 'var(--color-restaurant-primary)',
          hover: 'var(--color-restaurant-primary-hover)',
          soft: 'var(--color-restaurant-primary-soft)',
          bg: 'var(--color-restaurant-bg)',
          surface: 'var(--color-restaurant-surface)',
          'surface-hover': 'var(--color-restaurant-surface-hover)',
          text: 'var(--color-restaurant-text)',
          muted: 'var(--color-restaurant-text-muted)',
          border: 'var(--color-restaurant-border)',
        },
        navbar: {
          bg: 'var(--color-navbar-bg)',
          border: 'var(--color-navbar-border)',
        },
      },
      spacing: {
        'page-x': 'var(--spacing-page-x)',
        'page-y': 'var(--spacing-page-y)',
        'section-x': 'var(--spacing-section-x)',
        'section-y': 'var(--spacing-section-y)',
        card: 'var(--spacing-card)',
        'card-sm': 'var(--spacing-card-sm)',
        'card-lg': 'var(--spacing-card-lg)',
        'btn-x': 'var(--spacing-btn-x)',
        'btn-y': 'var(--spacing-btn-y)',
        'input-x': 'var(--spacing-input-x)',
        'input-y': 'var(--spacing-input-y)',
      },
      borderRadius: {
        card: 'var(--radius-2xl)',
        btn: 'var(--radius-xl)',
        input: 'var(--radius-xl)',
        badge: 'var(--radius-full)',
      },
      fontFamily: {
        sans: ['var(--font-sans)'],
        mono: ['var(--font-mono)'],
      },
    },
  },
  plugins: [],
};
