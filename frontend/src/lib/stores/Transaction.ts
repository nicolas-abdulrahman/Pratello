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
itemImages?: Record<number, string>;
  slug: string;
  restaurantId?: number | string;
  restaurantName: string;
  branchId: number;
  userId: number | null;
  guestToken: string | null;
  paymentKey: string;
  pixKey: string;
  transactionId?: number;
  paymentMethod?: string;
  /** TEST_ENV only: order exists only locally. Never call the API for it. */
  mock?: boolean;
  /** TEST_ENV only: payment was forced to "paid" locally. */
  testPaid?: boolean;
  pix?: { copy_paste_code: string; qr_code_base64: string };
  reviewedItems: number[];
  serviceReviewed: boolean;
  isReviewed?: boolean;
  popupDismissed: boolean;
  pickupAcknowledgedAt?: number;
}

export const isTestEnv = import.meta.env.TEST_ENV === 'true' || import.meta.env.TEST_ENV === '1';

/**
 * Set to true once you confirm the branch has the table the customer picked
 * (curl: POST .../orders/local with {"table_number":1} must return 201).
 */
const SEND_TABLE_NUMBER = false;

// ─────────────────────────────────────────────────────────────
// Persistence
// ─────────────────────────────────────────────────────────────

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

// /checkout has no slug in the URL, so remember the last restaurant the user visited.
const slugKey = 'pratello:last-slug';
export function rememberSlug(slug: string) {
  if (!slug) return;
  try { sessionStorage.setItem(slugKey, slug); } catch { /* ignore */ }
}
function recallSlug(): string {
  try { return sessionStorage.getItem(slugKey) || ''; } catch { return ''; }
}

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

// ─────────────────────────────────────────────────────────────
// HTTP
// ─────────────────────────────────────────────────────────────

export class ApiError extends Error {
  status: number;
  constructor(message: string, status: number) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
  }
}

const json = (body: unknown): RequestInit & { headers: Record<string, string> } => ({
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify(body),
});

async function request(path: string, init: RequestInit = {}, saved?: Pick<SavedOrder, 'userId' | 'guestToken'>): Promise<Response> {
  if (saved && saved.userId !== null && (!get(account).isLoggedIn || get(account).id !== saved.userId)) {
    throw new Error('Entre na conta que criou este pedido.');
  }
  const headers = new Headers(init.headers);
  if (saved?.guestToken) headers.set('X-Order-Token', saved.guestToken);
  const response = saved?.userId === null
    ? await fetch(`/api${path}`, { ...init, headers })
    : await account.authenticatedFetch(path, { ...init, headers });
  if (!response.ok) {
    const text = await response.text().catch(() => '');
    let detail = text;
    try {
      const body = JSON.parse(text);
      detail = typeof body.detail === 'string' ? body.detail : JSON.stringify(body.detail ?? body);
    } catch { /* not JSON */ }
    const method = init.method ?? 'GET';
    console.error(`[API ${response.status}] ${method} ${path}`, text, init.body);
    const clean = detail && detail !== '""' ? detail : 'Erro da API';
    throw new ApiError(`${clean} (${response.status} em ${method} ${path})`, response.status);
  }
  return response;
}

// ─────────────────────────────────────────────────────────────
// Create order
// ─────────────────────────────────────────────────────────────

export interface CreateOrderOptions {
  couponCodes?: string[];
  tableNumber?: string | number;
  deliveryAddress?: any;
  phone?: string;
  notes?: string;
  channel?: 'app' | 'whatsapp';
  mode?: 'mesa' | 'delivery' | 'whatsapp';
  shippingFee?: number;
  discount?: number;
}

function unitPrice(l: CartItem): number {
  const item: any = l.item;
  const unit = typeof item?.basePrice === 'number' && !isNaN(item.basePrice)
    ? item.basePrice
    : (typeof item?.price === 'number' && !isNaN(item.price) ? item.price : 0);
  return isNaN(unit) ? 0 : unit;
}

export async function createOrder(
  slug: string,
  lines: CartItem[],
  options: CreateOrderOptions = {}
): Promise<SavedOrder> {
  let effectiveSlug = slug;
  if (!effectiveSlug) {
    const match = window.location.pathname.match(/^\/(?:r|restaurant)\/([^/]+)/);
    if (match) effectiveSlug = match[1];
  }
  if (!effectiveSlug) effectiveSlug = recallSlug();
  if (!effectiveSlug) throw new Error('Restaurante não identificado. Volte ao cardápio do restaurante e tente de novo.');
  rememberSlug(effectiveSlug);
    const itemImages: Record<number, string> = {};

for (const line of lines) {
  if (line.item.imageUrl) {
    itemImages[Number(line.item.id)] = line.item.imageUrl;
  }
}

  if (get(account).isLoggedIn && !get(account).id) throw new Error('Não foi possível identificar sua conta. Entre novamente.');

  const couponCodes = options.couponCodes ?? [];
  const isRemote = !!options.deliveryAddress;
  if (isRemote && !options.phone) throw new Error('Informe um telefone com DDD para contato.');

  // Validation errors must NOT fall back to a mock order.
  const items = lines.map((line) => {
    const itemId = Number(line.item.id);
    if (!Number.isInteger(itemId) || itemId <= 0) {
      throw new Error(`Item inválido no carrinho: ${line.item.name}`);
    }
    return {
      item_id: itemId,
      quantity: line.quantity || 1,
      option_ids: line.item.extras?.filter((option) => option.selected).map((option) => option.id) ?? [],
      remove_ingredient_ids: line.item.ingredients?.filter((ingredient) => !ingredient.selected).map((ingredient) => ingredient.id) ?? [],
      ...(line.notes || options.notes ? { notes: line.notes || options.notes } : {}),
    };
  });

  try {
    const root = `/restaurants/${encodeURIComponent(effectiveSlug)}`;
    const [restaurantResponse, branchResponse] = await Promise.all([
      request(root),
      request(`${root}/branches?page=1&page_size=100`),
    ]);
    const restaurant = await restaurantResponse.json();
    const branches = await branchResponse.json();
    const list: any[] = branches.data ?? [];

    // Never guess a branch id: use a real one from this restaurant.
    const branch =
      list.find((b) => b.is_active && b.cnpj_verified) ??
      (isTestEnv ? (list.find((b) => b.is_active) ?? list[0]) : undefined);
    if (!branch) throw new Error('Nenhuma filial ativa e verificada está disponível.');
    const branchId: number = branch.id;
    console.info(`[createOrder] slug=${effectiveSlug} branch=${branchId}`);

    // 1) Create the order as an empty draft.
    const orderPayload: any = { coupon_codes: couponCodes, coins: 0 };
    let endpoint = `${root}/branches/${branchId}/orders/local`;
    if (isRemote) {
      orderPayload.delivery_address = options.deliveryAddress;
      orderPayload.phone = options.phone;
      endpoint = `${root}/branches/${branchId}/orders/remote`;
    } else if (SEND_TABLE_NUMBER && options.tableNumber && Number(options.tableNumber) > 0) {
      orderPayload.table_number = Number(options.tableNumber);
    }

    const response = await request(endpoint, json(orderPayload));
    const draft: OrderDetail = await response.json();
    const user = get(account);
    const auth = {
      userId: user.isLoggedIn ? user.id ?? null : null,
      guestToken: response.headers.get('X-Order-Token'),
    };

    // 2) Add items one by one so a bad item gives a clear error.
    for (const item of items) {
      await request(`/orders/${draft.id}/items`, json(item), auth);
    }

    // 3) Reload the order with items and totals.
    const order: OrderDetail = items.length
      ? await (await request(`/orders/${draft.id}`, {}, auth)).json()
      : draft;

    const saved: SavedOrder = {
      order,
              itemImages,
      slug: effectiveSlug,
      restaurantName: restaurant?.name ?? effectiveSlug,
      branchId,
      restaurantId: restaurant?.id ?? branch?.restaurant_id ?? effectiveSlug,
      userId: auth.userId,
      guestToken: auth.guestToken,
      paymentKey: crypto.randomUUID(),
      pixKey: crypto.randomUUID(),
      reviewedItems: [],
      serviceReviewed: false,
      popupDismissed: false,
    };
    transactions.update((orders) => [saved, ...orders]);
    return saved;
  } catch (err) {
    // Real HTTP errors (404, 400, ...) must be shown, never hidden behind a mock.
    if (!isTestEnv || err instanceof ApiError) throw err;
    console.warn('[TEST_ENV] createOrder failed (network), using local MOCK order (ids will look huge).', err);

    const mockId = Date.now();
    const subtotalCents = Math.round(lines.reduce((sum, l) => sum + unitPrice(l) * (l.quantity || 1), 0) * 100);
    const discountCents = Math.round((options.discount ?? 0) * 100);
    const shippingCents = Math.round((options.shippingFee ?? 0) * 100);
    const totalCents = Math.max(0, subtotalCents - discountCents + shippingCents);

    const mockOrder: OrderDetail = {
      id: mockId,
      status: 'pending',
      display_code: `#${mockId.toString().slice(-4)}`,
      total_cents: totalCents,
      submitted_at: new Date().toISOString(),
      created_at: new Date().toISOString(),
      items: lines.map((l, idx) => ({
        id: idx + 1,
        item_id: Number(l.item.id) || idx + 1,
        item_name: l.item.name || 'Item',
        quantity: l.quantity || 1,
        unit_price_cents: Math.round(unitPrice(l) * 100),
      })),
      transactions: [],
    };

    const user = get(account);
    const saved: SavedOrder = {
      order: mockOrder,
              itemImages,
      mock: true,
      slug: effectiveSlug,
      restaurantName: effectiveSlug,
      branchId: 0,
      restaurantId: effectiveSlug,
      userId: user.isLoggedIn ? user.id ?? null : null,
      guestToken: 'test-guest-token',
      paymentKey: crypto.randomUUID(),
      pixKey: crypto.randomUUID(),
      reviewedItems: [],
      serviceReviewed: false,
      popupDismissed: false,
    };
    transactions.update((orders) => [saved, ...orders]);
    return saved;
  }
}

// ─────────────────────────────────────────────────────────────
// Order lifecycle
// ─────────────────────────────────────────────────────────────

export async function refreshOrder(saved: SavedOrder): Promise<SavedOrder> {
  if (saved.mock) return saved; // never call the API for local mock orders
  try {
    const response = await request(`/orders/${saved.order.id}`, {}, saved);
    let order: OrderDetail = await response.json();
    // TEST_ENV: keep the forced "paid" state even if the server still says pending.
    if (isTestEnv && saved.testPaid && !isPaid(order)) {
      order = { ...order, transactions: saved.order.transactions };
    }
    updateOrder(saved.order.id, { order });
    return { ...saved, order };
  } catch (err) {
    if (isTestEnv) return saved;
    throw err;
  }
}

/** POST /orders/{id}/submit (skipped when already submitted). Also generates display_code. */
export async function submitOrder(saved: SavedOrder): Promise<SavedOrder> {
  if (saved.mock || saved.order.submitted_at) return saved;
  const response = await request(
    `/orders/${saved.order.id}/submit`,
    json({ coupon_codes: [], saved_coupon_ids: [], coins: 0 }),
    saved
  );
  const order: OrderDetail = await response.json();
  updateOrder(saved.order.id, { order });
  return { ...saved, order };
}

const methodMap: Record<string, string> = {
  pix: 'pix',
  cartao: 'credit_card',
  card: 'credit_card',
  credito: 'credit_card',
  debito: 'debit_card',
  dinheiro: 'cash',
  cash: 'cash',
};

export async function preparePayment(saved: SavedOrder, paymentMethod: string = 'pix'): Promise<SavedOrder> {
  const method = methodMap[paymentMethod] ?? 'pix';

  // Switching method on retry needs a fresh transaction + idempotency keys.
  if (saved.transactionId && saved.paymentMethod && saved.paymentMethod !== method) {
    saved = {
      ...saved,
      transactionId: undefined,
      pix: undefined,
      paymentKey: crypto.randomUUID(),
      pixKey: crypto.randomUUID(),
    };
  }

  try {
    if (saved.mock) throw new Error('mock order'); // jumps to the TEST_ENV mock below, no network
    if (saved.userId === null && !saved.guestToken) {
      throw new Error('O servidor não retornou o token do pedido de visitante. Não crie outro pedido; entre em contato com o restaurante.');
    }
    try { saved = await refreshOrder(saved); } catch { /* ignore, submit will surface real errors */ }
    saved = await submitOrder(saved);

    if (!saved.transactionId) {
      const init = json({ payment_method: method, amount_cents: saved.order.total_cents });
      init.headers['Idempotency-Key'] = saved.paymentKey;
      const response = await request(`/orders/${saved.order.id}/transactions`, init, saved);
      const transaction = await response.json();
      saved = { ...saved, transactionId: transaction.id, paymentMethod: method };
      updateOrder(saved.order.id, { transactionId: transaction.id, paymentMethod: method });
    }

    if (method === 'pix' && !saved.pix) {
      const response = await request(
        `/transactions/${saved.transactionId}/pix`,
        { method: 'POST', headers: { 'Idempotency-Key': saved.pixKey } },
        saved
      );
      const pix = await response.json();
      saved = { ...saved, pix };
      updateOrder(saved.order.id, { pix });
    }
    return saved;
  } catch (err) {
    if (!isTestEnv) throw err;
    // Real HTTP errors are shown; only mock orders / network failures use the local mock.
    if (!saved.mock && err instanceof ApiError) throw err;
    if (!saved.mock) console.warn('[TEST_ENV] preparePayment failed (network), using local mock transaction.', err);

    const transactionId = saved.transactionId ?? Date.now();
    const mockPix = method === 'pix' ? {
      copy_paste_code: `00020126360014BR.GOV.BCB.PIX0114+5511987654321520400005303986540${(saved.order.total_cents / 100).toFixed(2)}5802BR5920RESTAURANTE6009SAO PAULO62070503***6304`,
      qr_code_base64: 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==',
    } : undefined;

    const order: OrderDetail = {
      ...saved.order,
      submitted_at: saved.order.submitted_at ?? new Date().toISOString(),
      transactions: [{ id: transactionId, status: 'pending', amount_cents: saved.order.total_cents }],
    };
    const mocked: SavedOrder = {
      ...saved,
      transactionId,
      paymentMethod: method,
      order,
      ...(mockPix ? { pix: mockPix } : {}),
    };
    updateOrder(saved.order.id, mocked);
    return mocked;
  }
}

/** Marks the order as paid locally. TEST_ENV only. */
function forcePaid(saved: SavedOrder): SavedOrder {
  const txId = saved.transactionId ?? Date.now();
  const paidTx = { id: txId, status: 'paid', amount_cents: saved.order.total_cents };
  const order: OrderDetail = {
    ...saved.order,
    transactions: [...(saved.order.transactions ?? []).filter((t) => t.id !== txId), paidTx],
  };
  const paid: SavedOrder = { ...saved, transactionId: txId, order, testPaid: true };
  updateOrder(saved.order.id, { order, transactionId: txId, testPaid: true });
  return paid;
}

/** PIX sandbox confirm (POST /transactions/{id}/pix/confirm). */
export async function confirmPayment(saved: SavedOrder): Promise<SavedOrder> {
  try {
    if (saved.mock) return forcePaid(saved);
    const response = await request(`/transactions/${saved.transactionId}/pix/confirm`, { method: 'POST' }, saved);
    const transaction = await response.json();
    if (transaction.status !== 'paid') throw new Error('O pagamento ainda não foi confirmado.');
    const order: OrderDetail = {
      ...saved.order,
      transactions: [...(saved.order.transactions ?? []).filter((t) => t.id !== transaction.id), transaction],
    };
    updateOrder(saved.order.id, { order });
    const paid = { ...saved, order };
    try { return await refreshOrder(paid); } catch { return paid; }
  } catch (err) {
    if (!isTestEnv) throw err;
    return forcePaid(saved);
  }
}

/**
 * TEST_ENV: settles ANY payment method (pix, card, cash) so the flow always passes.
 * PIX tries the real sandbox confirm first; card/cash have no endpoint, so they are forced to paid.
 */
export async function settlePayment(saved: SavedOrder): Promise<SavedOrder> {
  if (!isTestEnv) throw new Error('A confirmação simulada só está disponível no ambiente de teste.');
  if (saved.paymentMethod === 'pix') return confirmPayment(saved);
  return forcePaid(saved);
}

export async function submitReview(saved: SavedOrder, itemId: number | null, rating: number, comment: string) {
  const path = itemId === null ? `/orders/${saved.order.id}/review` : `/orders/${saved.order.id}/items/${itemId}/review`;
  await request(path, json({ rating, ...(saved.userId !== null && comment.trim() ? { comment: comment.trim() } : {}) }), saved);
  if (itemId === null) updateOrder(saved.order.id, { serviceReviewed: true });
  else updateOrder(saved.order.id, { reviewedItems: [...saved.reviewedItems, itemId] });
}
