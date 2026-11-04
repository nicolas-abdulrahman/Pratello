import { writable } from 'svelte/store';

export const checkoutData = writable({
  total: 0,
  items: [],
  restaurant: null
});
