// ─────────────────────────────────────────────────────────────
// Types
// ─────────────────────────────────────────────────────────────

export type Review = {
  rating: number;
  comment?: string;
  author?: string;
  date?: string;
};

export type Ingredient = {
  id: number;
  name: string;
  selected: boolean;
  price?: number;
};

export type Extra = {
  id: number;
  name: string;
  price: number;
  selected: boolean;
};

export type ComboComponent = {
  name: string;
  quantity: number;
  imageUrl?: string;
};

export type ItemType = 'simple' | 'combo';

export async function get_description(
  slug: string,
  item: { id: number; is_combo: boolean; description?: string | null },
  signal?: AbortSignal,
): Promise<string | undefined> {
  if (!item.is_combo) return item.description ?? undefined;

  const response = await fetch(`/api/restaurants/${encodeURIComponent(slug)}/items/${item.id}`, { signal });
  if (!response.ok) return item.description ?? undefined;
  const { combo_items }: { combo_items?: { sub_item_name: string; quantity: number }[] } = await response.json();
  return combo_items?.map(({ sub_item_name, quantity }) => `${quantity} ${sub_item_name}`).join(', ')
    || item.description || undefined;
}

export type Item = {
  id: number;
  name: string;
  description?: string;
  imageUrl?: string;
  basePrice: number;
  type: ItemType;
  components?: ComboComponent[];
  ingredients: Ingredient[];
  extras: Extra[];
  review?: Review;
};

// ─────────────────────────────────────────────────────────────
// Config — flip to false once your backend is ready
// ─────────────────────────────────────────────────────────────

const API = '/api';
const USE_MOCK_FALLBACK = true;

// ─────────────────────────────────────────────────────────────
// Pricing
// ─────────────────────────────────────────────────────────────

export function calc_price(item: Item): number {
  const ingredientsTotal = item.ingredients
    .filter((i) => i.selected && typeof i.price === 'number')
    .reduce((sum, i) => sum + (i.price as number), 0);

  const extrasTotal = item.extras
    .filter((e) => e.selected)
    .reduce((sum, e) => sum + e.price, 0);

  return item.basePrice + ingredientsTotal + extrasTotal;
}

// ─────────────────────────────────────────────────────────────
// Safe fetch helper
// ─────────────────────────────────────────────────────────────

async function safe_fetch<T>(
  url: string,
  fallback: () => T,
  what: string,
): Promise<T> {
  try {
    const res = await fetch(url);
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    return (await res.json()) as T;
  } catch (err) {
    if (!USE_MOCK_FALLBACK) throw err;
    console.warn(`[Item] ${what} failed — using placeholder.`, err);
    return fallback();
  }
}

// ─────────────────────────────────────────────────────────────
// Placeholder data (deterministic, id-based)
// ─────────────────────────────────────────────────────────────

function placeholder_item(id: number): Item {
  return {
    id,
    name: `Item #${id}`,
    description:
      'Este item ainda não foi carregado do servidor. Os dados exibidos são um placeholder.',
    imageUrl:
      'data:image/svg+xml;utf8,' +
      encodeURIComponent(
        `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 300 300">
          <rect width="300" height="300" fill="#f5f5f4"/>
          <g fill="none" stroke="#a8a29e" stroke-width="8" stroke-linecap="round" stroke-linejoin="round">
            <rect x="70" y="70" width="160" height="160" rx="20"/>
            <circle cx="120" cy="120" r="18"/>
            <path d="M90 210l50-50 35 35 25-25 40 40"/>
          </g>
        </svg>`,
      ),
    basePrice: 0,
    type: 'simple',
    ingredients: [],
    extras: [],
  };
}

function placeholder_ingredients(_itemId: number): Ingredient[] {
  return [];
}

function placeholder_extras(_itemId: number): Extra[] {
  return [];
}

function placeholder_review(_itemId: number): Review | undefined {
  return undefined;
}

// ─────────────────────────────────────────────────────────────
// Fetchers
// ─────────────────────────────────────────────────────────────

export async function fetch_item(id: number): Promise<Item> {
  return safe_fetch(
    `${API}/items/${id}`,
    () => placeholder_item(id),
    `fetch_item(${id})`,
  );
}

export async function fetch_ingredients(itemId: number): Promise<Ingredient[]> {
  return safe_fetch(
    `${API}/items/${itemId}/ingredients`,
    () => placeholder_ingredients(itemId),
    `fetch_ingredients(${itemId})`,
  );
}

export async function fetch_extras(itemId: number): Promise<Extra[]> {
  return safe_fetch(
    `${API}/items/${itemId}/extras`,
    () => placeholder_extras(itemId),
    `fetch_extras(${itemId})`,
  );
}

export async function fetch_review(itemId: number): Promise<Review | undefined> {
  return safe_fetch(
    `${API}/items/${itemId}/review`,
    () => placeholder_review(itemId),
    `fetch_review(${itemId})`,
  );
}

export async function fetch_extras_and_ingredients(
  itemId: number,
): Promise<{ ingredients: Ingredient[]; extras: Extra[] }> {
  const [ingredients, extras] = await Promise.all([
    fetch_ingredients(itemId),
    fetch_extras(itemId),
  ]);
  return { ingredients, extras };
}
