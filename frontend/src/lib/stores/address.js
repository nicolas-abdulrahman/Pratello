import { writable, derived } from 'svelte/store';

const STORAGE_KEY = 'pratello:address';

const empty = { state: '', stateName: '', city: '' };

function load() {
  if (typeof localStorage === 'undefined') return empty;
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    return raw ? { ...empty, ...JSON.parse(raw) } : empty;
  } catch {
    return empty;
  }
}

function createAddressStore() {
  const { subscribe, set } = writable(load());

  return {
    subscribe,
    setLocation({ state, stateName, city }) {
      const next = { state, stateName, city };
      if (typeof localStorage !== 'undefined') {
        localStorage.setItem(STORAGE_KEY, JSON.stringify(next));
      }
      set(next);
    }
  };
}

export const address = createAddressStore();

export const addressLabel = derived(address, ($a) =>
  $a.city && $a.state ? `${$a.city}, ${$a.state}` : 'Escolha um endereço'
);
