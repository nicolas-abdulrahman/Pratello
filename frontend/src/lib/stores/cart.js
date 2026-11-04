import { writable } from 'svelte/store';

// Initialize with an empty cart
export const cart = writable([]);

// In cart.js
export function addToCart(newItem) {
  cart.update((items) => {
    const existingItemIndex = items.findIndex((i) => i.name === newItem.name);
    
    if (existingItemIndex !== -1) {
      const updatedItems = [...items];
      updatedItems[existingItemIndex].quantity += 1;
      return updatedItems;
    } else {
      // Spread newItem (...newItem) to ensure the 'combo' property is kept!
      return [...items, { ...newItem, id: Date.now(), quantity: 1 }];
    }
  });
}

export function removeFromCart(id) {
  cart.update((items) => items.filter((item) => item.id !== id));
}
