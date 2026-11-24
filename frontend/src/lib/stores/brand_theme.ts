import { writable } from 'svelte/store';

const key = 'pratello-brand-dark';

function initialDark(): boolean {
  try {
    return typeof localStorage !== 'undefined' && localStorage.getItem(key) === 'true';
  } catch {
    return false;
  }
}

export const brandDark = writable(initialDark());

export function toggleBrandDark() {
  brandDark.update((dark) => {
    const next = !dark;
    try {
      localStorage.setItem(key, String(next));
    } catch {
      // Keep the toggle usable when browser storage is unavailable.
    }
    return next;
  });
}
