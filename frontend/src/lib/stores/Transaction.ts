import { get, writable } from 'svelte/store';
import { account } from './account';
import type { CartItem } from './Cart';

export interface OrderLine {
  id: number;
  item_id: number;
  item_name: string;
  quantity: number;
  unit_price_cents: number;
  voided_at?: string | null;
}

export interface OrderDetail {
  id: number;
  status: string;
  display_code?: string | null;
  total_cents: number;
  submitted_at?: string | null;
  created_at: string;
  items: OrderLine[];
  transactions?: { id: number; status: string; amount_cents: number }[];
}

export interface SavedOrder {
  order: OrderDetail;
  slug: string;
  restaurantName: string;
  branchId: number;
  userId: number | null;
  guestToken: string | null;
  paymentKey: string;
  pixKey: string;
  transactionId?: number;
  pix?: { copy_paste_code: string; qr_code_base64: string };
  reviewedItems: number[];
  serviceReviewed: boolean;
  popupDismissed: boolean;
  pickupAcknowledgedAt?: number;
}

const storageKey = 'pratello:transactions';
function load(): SavedOrder[] {
  try {
    const value = JSON.parse(localStorage.getItem(storageKey) || '[]');
    return Array.isArray(value) ? value : [];
  } catch { return []; }
}

export const transactions = writable<SavedOrder[]>(load());
transactions.subscribe((orders) => {
  try { localStorage.setItem(storageKey, JSON.stringify(orders)); } catch { /* storage unavailable */ }
});

export function visibleOrders(): SavedOrder[] {
  const user = get(account);
  return get(transactions).filter((o) => user.isLoggedIn ? o.userId === user.id : o.userId === null);
}

export function isPaid(order: OrderDetail): boolean {
  return (order.transactions ?? []).filter((t) => t.status === 'paid')
    .reduce((total, t) => total + t.amount_cents, 0) >= order.total_cents && order.total_cents > 0;
}

export function updateOrder(id: number, patch: Partial<SavedOrder>) {
  transactions.update((orders) => orders.map((o) => o.order.id === id ? { ...o, ...patch } : o));
}

async function request(path: string, init: RequestInit = {}, saved?: SavedOrder): Promise<Response> {
  if (saved && saved.userId !== null && (!get(account).isLoggedIn || get(account).id !== saved.userId)) {
    throw new Error('Entre na conta que criou este pedido.');
  }
  const headers = new Headers(init.headers);
  if (saved?.guestToken) headers.set('X-Order-Token', saved.guestToken);
  const response = saved?.userId === null
    ? await fetch(`/api${path}`, { ...init, headers })
    : await account.authenticatedFetch(path, { ...init, headers });
  if (!response.ok) {
    let detail = '';
    try { const body = await response.json(); detail = typeof body.detail === 'string' ? body.detail : ''; } catch { /* no JSON */ }
    throw new Error(detail || `Erro da API (${response.status}).`);
  }
  return response;
}

function json(body: object): RequestInit {
  return { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) };
}

export async function createOrder(slug: string, lines: CartItem[]): Promise<SavedOrder> {
  if (!slug) throw new Error('Escolha um restaurante antes de finalizar o pedido.');
  if (get(account).isLoggedIn && !get(account).id) throw new Error('Não foi possível identificar sua conta. Entre novamente.');
  const root = `/restaurants/${encodeURIComponent(slug)}`;
  const [restaurantResponse, branchResponse] = await Promise.all([
    request(root),
    request(`${root}/branches?page=1&page_size=100`),
  ]);
  const restaurant = await restaurantResponse.json();
  const branches = await branchResponse.json();
  const branch = branches.data?.find((b: { is_active: boolean; cnpj_verified: boolean }) => b.is_active && b.cnpj_verified);
  if (!branch) throw new Error('Nenhuma filial ativa e verificada está disponível.');
  const items = lines.map((line) => ({
    item_id: line.item.id,
    quantity: line.quantity,
    option_ids: line.item.extras.filter((option) => option.selected).map((option) => option.id),
    remove_ingredient_ids: line.item.ingredients.filter((ingredient) => !ingredient.selected).map((ingredient) => ingredient.id),
    ...(line.notes ? { notes: line.notes } : {}),
  }));
  const response = await request(`${root}/branches/${branch.id}/orders/local`, json({ items, coupon_codes: [], coins: 0 }));
  const order: OrderDetail = await response.json();
  const user = get(account);
  const saved: SavedOrder = {
    order, slug, restaurantName: restaurant.name, branchId: branch.id,
    userId: user.isLoggedIn ? user.id ?? null : null,
    guestToken: response.headers.get('X-Order-Token'),
    paymentKey: crypto.randomUUID(), pixKey: crypto.randomUUID(),
    reviewedItems: [], serviceReviewed: false, popupDismissed: false,
  };
  transactions.update((orders) => [saved, ...orders]);
  return saved;
}

export async function preparePayment(saved: SavedOrder): Promise<SavedOrder> {
  if (saved.userId === null && !saved.guestToken) throw new Error('O servidor não retornou o token do pedido de visitante. Não crie outro pedido; entre em contato com o restaurante.');
  saved = await refreshOrder(saved);
  if (!saved.order.submitted_at) {
    const response = await request(`/orders/${saved.order.id}/submit`, json({ coupon_codes: [], saved_coupon_ids: [], coins: 0 }), saved);
    saved = { ...saved, order: await response.json() };
    updateOrder(saved.order.id, { order: saved.order });
  }
  if (!saved.transactionId) {
    const init = json({ payment_method: 'pix', amount_cents: saved.order.total_cents });
    (init.headers as Record<string, string>)['Idempotency-Key'] = saved.paymentKey;
    const response = await request(`/orders/${saved.order.id}/transactions`, init, saved);
    const transaction = await response.json();
    saved = { ...saved, transactionId: transaction.id };
    updateOrder(saved.order.id, { transactionId: transaction.id });
  }
  if (!saved.pix) {
    const response = await request(`/transactions/${saved.transactionId}/pix`, { method: 'POST', headers: { 'Idempotency-Key': saved.pixKey } }, saved);
    const pix = await response.json();
    saved = { ...saved, pix };
    updateOrder(saved.order.id, { pix });
  }
  return saved;
}

export async function refreshOrder(saved: SavedOrder): Promise<SavedOrder> {
  const response = await request(`/orders/${saved.order.id}`, {}, saved);
  const order: OrderDetail = await response.json();
  updateOrder(saved.order.id, { order });
  return { ...saved, order };
}

export async function confirmPayment(saved: SavedOrder): Promise<SavedOrder> {
  const response = await request(`/transactions/${saved.transactionId}/pix/confirm`, { method: 'POST' }, saved);
  const transaction = await response.json();
  if (transaction.status !== 'paid') throw new Error('O pagamento ainda não foi confirmado.');
  const order = {
    ...saved.order,
    transactions: [...(saved.order.transactions ?? []).filter((entry) => entry.id !== transaction.id), transaction],
  };
  updateOrder(saved.order.id, { order });
  const paid = { ...saved, order };
  try {
    return await refreshOrder(paid);
  } catch {
    // Confirmation succeeded; keep it visible even if refreshing the order fails.
    return paid;
  }
}

export async function submitReview(saved: SavedOrder, itemId: number | null, rating: number, comment: string) {
  const path = itemId === null ? `/orders/${saved.order.id}/review` : `/orders/${saved.order.id}/items/${itemId}/review`;
  await request(path, json({ rating, ...(saved.userId !== null && comment.trim() ? { comment: comment.trim() } : {}) }), saved);
  if (itemId === null) updateOrder(saved.order.id, { serviceReviewed: true });
  else updateOrder(saved.order.id, { reviewedItems: [...saved.reviewedItems, itemId] });
}
