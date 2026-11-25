import { writable } from 'svelte/store';

export type DineSession = {
  /** true when the customer arrived through /r/{slug}/dine-in */
  active: boolean;
  /** restaurant the dine-in session belongs to */
  slug: string;
};

const KEY = 'pratello:dine-in';
const empty: DineSession = { active: false, slug: '' };

function load(): DineSession {
  try {
    const value = JSON.parse(sessionStorage.getItem(KEY) || 'null');
    return value && typeof value.active === 'boolean' ? { ...empty, ...value } : empty;
  } catch {
    return empty;
  }
}

export const dineSession = writable<DineSession>(load());

dineSession.subscribe((value) => {
  try { sessionStorage.setItem(KEY, JSON.stringify(value)); } catch { /* storage unavailable */ }
});

export function enterDineIn(slug: string) {
  dineSession.set({ active: true, slug });
}

export function leaveDineIn() {
  dineSession.set(empty);
}
