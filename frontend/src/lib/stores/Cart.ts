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
  if (!line || !line.item) return 0;
  const unit = calc_item_unit_price(line.item);
  const qty = typeof line.quantity === 'number' && !isNaN(line.quantity) && line.quantity > 0 ? line.quantity : 1;
  const total = (isNaN(unit) ? 0 : unit) * qty;
  return isNaN(total) ? 0 : total;
}

export function calc_price(items: CartItem[]): number {
  if (!Array.isArray(items)) return 0;
  return items.reduce((sum, line) => {
    const val = calc_line_price(line);
    return sum + (isNaN(val) ? 0 : val);
  }, 0);
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
function signature(item: any): string {
  const ingr = (item.ingredients || [])
    .filter((i: any) => i.selected)
    .map((i: any) => i.id)
    .sort()
    .join(',');
  const extr = (item.extras || [])
    .filter((e: any) => e.selected)
    .map((e: any) => e.id)
    .sort()
    .join(',');
  return `${item.id || item.name}|${item.type || 'simple'}|${ingr}|${extr}`;
}

function clone_item(item: any): Item {
  return {
    ...item,
    id: item.id || Date.now(),
    name: item.name || item.title || 'Item',
    description: item.description || item.desc,
    imageUrl: item.imageUrl || item.image || item.icon,
    basePrice: typeof item.basePrice === 'number' ? item.basePrice : (typeof item.price === 'number' ? item.price : 0),
    type: item.type || (item.combo || item.isCombo ? 'combo' : 'simple'),
    components: item.components ? item.components.map((c: any) => ({ ...c })) : undefined,
    ingredients: Array.isArray(item.ingredients) ? item.ingredients.map((i: any) => ({ ...i })) : [],
    extras: Array.isArray(item.extras) ? item.extras.map((e: any) => ({ ...e })) : [],
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
