import { writable, derived, get } from 'svelte/store';
import type { Item } from './Item';
import { calc_price as calc_item_unit_price } from './Item';

// ─────────────────────────────────────────────────────────────
// Types
// ─────────────────────────────────────────────────────────────

export type CartItem = {
  /** cart-line id (client-generated; backend replaces this later) */
  id: number;
  item: Item;
  quantity: number;
  notes?: string;
};

// ─────────────────────────────────────────────────────────────
// Store
// ─────────────────────────────────────────────────────────────

export const cart = writable<CartItem[]>([]);

export const cart_count = derived(cart, ($c) =>
  $c.reduce((sum, line) => sum + line.quantity, 0),
);

// ─────────────────────────────────────────────────────────────
// Pricing
// ─────────────────────────────────────────────────────────────

export function calc_line_price(line: CartItem): number {
  return calc_item_unit_price(line.item) * line.quantity;
}

export function calc_price(items: CartItem[]): number {
  return items.reduce((sum, line) => sum + calc_line_price(line), 0);
}

export const cart_total = derived(cart, ($c) => calc_price($c));

// ─────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────

// Monotonic counter, seeded with Date.now() so cart-line ids can never
// collide with backend ids (which are small integers).
let next_line_id = Date.now();

function generate_line_id(): number {
  return next_line_id++;
}

/** Lines merge when menu item + selected ingredients + extras match. */
function signature(item: Item): string {
  const ingr = item.ingredients
    .filter((i) => i.selected)
    .map((i) => i.id)
    .sort()
    .join(',');
  const extr = item.extras
    .filter((e) => e.selected)
    .map((e) => e.id)
    .sort()
    .join(',');
  return `${item.id}|${item.type}|${ingr}|${extr}`;
}

function clone_item(item: Item): Item {
  return {
    ...item,
    components: item.components ? item.components.map((c) => ({ ...c })) : undefined,
    ingredients: item.ingredients.map((i) => ({ ...i })),
    extras: item.extras.map((e) => ({ ...e })),
    review: item.review ? { ...item.review } : undefined,
  };
}

// ─────────────────────────────────────────────────────────────
// Mutations
// ─────────────────────────────────────────────────────────────

export function add_item(item: Item, quantity: number = 1): void {
  const cloned = clone_item(item);
  const sig = signature(cloned);

  cart.update((items) => {
    const existing = items.find((line) => signature(line.item) === sig);
    if (existing) {
      return items.map((line) =>
        line.id === existing.id
          ? { ...line, quantity: line.quantity + quantity }
          : line,
      );
    }
    return [
      ...items,
      { id: generate_line_id(), item: cloned, quantity },
    ];
  });
}

export function remove_item(id: number): void {
  cart.update((items) => items.filter((line) => line.id !== id));
}

export function update_quantity(id: number, quantity: number): void {
  if (quantity <= 0) {
    remove_item(id);
    return;
  }
  cart.update((items) =>
    items.map((line) => (line.id === id ? { ...line, quantity } : line)),
  );
}

export function toggle_ingredient(lineId: number, ingredientId: number): void {
  cart.update((items) =>
    items.map((line) =>
      line.id === lineId
        ? {
            ...line,
            item: {
              ...line.item,
              ingredients: line.item.ingredients.map((ing) =>
                ing.id === ingredientId ? { ...ing, selected: !ing.selected } : ing,
              ),
            },
          }
        : line,
    ),
  );
}

export function toggle_extra(lineId: number, extraId: number): void {
  cart.update((items) =>
    items.map((line) =>
      line.id === lineId
        ? {
            ...line,
            item: {
              ...line.item,
              extras: line.item.extras.map((ex) =>
                ex.id === extraId ? { ...ex, selected: !ex.selected } : ex,
              ),
            },
          }
        : line,
    ),
  );
}

export function set_notes(lineId: number, notes: string): void {
  cart.update((items) =>
    items.map((line) => (line.id === lineId ? { ...line, notes } : line)),
  );
}

export function clear_cart(): void {
  cart.set([]);
}

// ─────────────────────────────────────────────────────────────
// Read helpers
// ─────────────────────────────────────────────────────────────

export function get_line(id: number): CartItem | undefined {
  return get(cart).find((line) => line.id === id);
}
