// lib/stores/account.js
import { writable, derived } from 'svelte/store';

const STORAGE_KEY = 'pratello:account';

const emptyState = {
  isLoggedIn: false,
  name: '',
  email: '',
  role: '',       // 'cliente' | 'funcionario'
  initials: ''
};

// Hydrate from localStorage on module load (client only)
function loadFromStorage() {
  if (typeof localStorage === 'undefined') return emptyState;
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return emptyState;
    const parsed = JSON.parse(raw);
    // Basic sanity check so a corrupted blob doesn't poison the app
    if (typeof parsed?.isLoggedIn === 'boolean') return { ...emptyState, ...parsed };
    return emptyState;
  } catch {
    return emptyState;
  }
}

function makeInitials(name = '') {
  return name
    .split(' ')
    .filter(Boolean)
    .map((p) => p[0])
    .slice(0, 2)
    .join('')
    .toUpperCase();
}

function createAccountStore() {
  const { subscribe, set } = writable(loadFromStorage());

  return {
    subscribe,
    login({ name, email, role }) {
      const next = {
        isLoggedIn: true,
        name,
        email,
        role,
        initials: makeInitials(name)
      };
      if (typeof localStorage !== 'undefined') {
        localStorage.setItem(STORAGE_KEY, JSON.stringify(next));
      }
      set(next);
    },
    logout() {
      if (typeof localStorage !== 'undefined') {
        localStorage.removeItem(STORAGE_KEY);
      }
      set(emptyState);
    }
  };
}

export const account = createAccountStore();
export const isLoggedIn = derived(account, ($a) => $a.isLoggedIn);
