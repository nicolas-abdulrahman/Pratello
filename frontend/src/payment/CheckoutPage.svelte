<script lang="ts">
  import { untrack } from 'svelte';
  import Checkout from './checkout/Checkout.svelte';
  import WhatsAppSent, { type Order as WhatsAppOrder } from './checkout/WhatsAppSent.svelte';
  import { cart, cart_total, clear_cart, calc_line_price, type CartItem } from '../lib/stores/Cart';
  import { account } from '../lib/stores/account';
  import {
    createOrder,
    preparePayment,
    submitOrder,
    settlePayment,
    refreshOrder,
    isPaid,
    isTestEnv,
    visibleOrders,
    type SavedOrder
  } from '../lib/stores/Transaction';
  import { get } from 'svelte/store';
  import { ArrowLeft } from 'lucide-svelte';

  let {
    onCancel = () => {},
    slug = '',
    total = 0,
    onOrderConfirmed = undefined as ((restaurantId: string | number, orderId: string | number) => void) | undefined
  }: {
    onCancel?: () => void;
    slug?: string;
    total?: number;
    onOrderConfirmed?: (restaurantId: string | number, orderId: string | number) => void;
  } = $props();

  // Exactly what the API returns. null until loaded.
  let restaurantInfo = $state<Record<string, any> | null>(null);

  let whatsappOrder = $state<WhatsAppOrder | null>(null);
  let confirmedOrder = $state<{
    orderId: string;
    orderNumber: number;
    restaurantId: string;
    restaurantName: string;
    displayCode: string;
    paid: boolean;
  } | null>(null);

  // Production PIX: waiting for the bank to settle.
  let pixPending = $state<SavedOrder | null>(null);
  let pixCopied = $state(false);
  let pixContext: { items: CartItem[]; payload: any } | null = null;

  // Load restaurant
  $effect(() => {
    const currentSlug = slug;
    if (!currentSlug) return;
    const controller = new AbortController();
    fetch(`/api/restaurants/${encodeURIComponent(currentSlug)}`, { signal: controller.signal })
      .then((res) => (res.ok ? res.json() : null))
      .then((data) => {
        if (data && !controller.signal.aborted) restaurantInfo = data;
      })
      .catch(() => {});
    return () => controller.abort();
  });

  // Restore WhatsApp screen on /pedido-whatsapp/:id
  $effect(() => {
    const waMatch = window.location.pathname.match(/^\/pedido-whatsapp\/([^/]+)\/?$/);
    if (!waMatch) return;
    untrack(() => {
      if (whatsappOrder) return;
      const found = visibleOrders().find((o) => String(o.order.id) === waMatch[1]);
      if (!found) return;
      whatsappOrder = {
        id: String(found.order.id),
        number: found.order.id,
        clientId: found.userId ? String(found.userId) : 'guest',
        customerName: $account.name || 'Cliente',
        phone: restaurantInfo?.phone ?? '',
        location: { mode: 'mesa', table: '01' },
        items: found.order.items.map((i) => ({
          productId: String(i.item_id),
          name: i.item_name,
          price: i.unit_price_cents / 100,
          quantity: i.quantity
        })),
        notes: '',
        subtotal: found.order.total_cents / 100,
        deliveryFee: 0,
        total: found.order.total_cents / 100,
        payment: { method: 'whatsapp', status: 'pago' },
        status: 'preparo',
        createdAt: Date.now()
      };
    });
  });

  // Production PIX: poll until the transaction is paid.
  $effect(() => {
    const current = pixPending;
    if (!current) return;
    let stopped = false;
    const tick = async () => {
      try {
        const fresh = await refreshOrder(current);
        if (!stopped && isPaid(fresh.order) && pixContext) {
          finalize(fresh, pixContext.items, pixContext.payload, true);
        }
      } catch {
        /* keep polling */
      }
    };
    const timer = window.setInterval(tick, 4000);
    return () => {
      stopped = true;
      window.clearInterval(timer);
    };
  });

  function getLinePrice(line: any): number {
    if (typeof line?.item?.basePrice === 'number' && !isNaN(line.item.basePrice) && line.item.basePrice > 0) {
      return line.item.basePrice;
    }
    if (typeof line?.item?.price === 'number' && !isNaN(line.item.price) && line.item.price > 0) {
      return line.item.price;
    }
    const calc = calc_line_price(line);
    if (typeof calc === 'number' && !isNaN(calc) && (line?.quantity || 1) > 0) {
      return calc / (line?.quantity || 1);
    }
    return 0;
  }

  let cartLines = $derived(
    $cart.map((line) => {
      const price = getLinePrice(line);
      return {
        product: {
          id: String(line.item.id || ''),
          name: line.item.name || 'Item',
          price: isNaN(price) ? 0 : price,
          image:
            line.item.imageUrl ||
            (line.item as any).image ||
            (line.item as any).image_url ||
            (line.item as any).icon ||
            ''
        },
        quantity: line.quantity || 1
      };
    })
  );

  let clientSession = $derived({
    clientId: $account.id ? String($account.id) : undefined,
    name: $account.name || undefined,
    phone: $account.phone || undefined
  });

  /** Maps the checkout payment draft to a method key understood by preparePayment. */
  function resolveMethod(payload: any): string {
    if (payload.channel === 'whatsapp') return 'whatsapp';
    const m = payload.payment?.method;
    if (m === 'cartao' || m === 'card') return payload.payment?.cardType === 'debito' ? 'debito' : 'credito';
    return m || 'pix';
  }

  function finalize(prepared: SavedOrder, items: CartItem[], payload: any, paid: boolean) {
    const isWhatsapp = payload.channel === 'whatsapp';
    const orderId = String(prepared.order.id);
    const orderNumber = prepared.order.id;
    const restId = String(prepared.restaurantId || prepared.slug || slug || '');
    const displayCode = prepared.order.display_code || `#${orderId}`;
    const discount = (payload.summary?.discount ?? 0) + (payload.summary?.shippingDiscount ?? 0);

    pixPending = null;
    pixContext = null;

    if (isWhatsapp) {
      whatsappOrder = {
        id: orderId,
        number: orderNumber,
        clientId: $account.id ? String($account.id) : 'guest',
        customerName: $account.name || 'Cliente',
        phone: payload.phone || restaurantInfo?.phone || '',
        location: payload.location,
        items: items.map((l) => {
          const price = getLinePrice(l);
          return {
            productId: String(l.item.id),
            name: l.item.name || 'Item',
            price: isNaN(price) ? 0 : price,
            quantity: l.quantity || 1
          };
        }),
        notes: payload.notes || '',
        subtotal: payload.summary?.subtotal ?? 0,
        deliveryFee: payload.summary?.shippingFee ?? 0,
        discount,
        coupons: payload.couponCodes,
        channel: 'whatsapp',
        shipping: payload.shipping,
        total: payload.summary?.total ?? 0,
        payment: {
          method: payload.payment?.method || 'whatsapp',
          cardType: payload.payment?.cardType,
          installments: payload.payment?.installments,
          needsMachine: payload.payment?.needsMachine,
          changeFor: payload.payment?.changeFor ? Number(payload.payment.changeFor) : undefined,
          status: 'pago'
        },
        status: 'preparo',
        createdAt: Date.now()
      };
    } else {
      confirmedOrder = {
        orderId,
        orderNumber,
        restaurantId: restId,
        restaurantName: prepared.restaurantName || restaurantInfo?.name || '',
        displayCode,
        paid
      };
    }

    clear_cart();
  }

  async function handlePlaceOrder(payload: any): Promise<{ id: string; number: number } | null> {
    try {
      const items = get(cart);
      const isWhatsapp = payload.channel === 'whatsapp';
      const method = resolveMethod(payload);
      const mode = payload.location?.mode;

      const saved = await createOrder(slug, items, {
        couponCodes: payload.couponCodes,
        tableNumber: mode === 'mesa' ? payload.location.table : undefined,
        deliveryAddress: mode === 'delivery' || mode === 'whatsapp' ? payload.location?.address : undefined,
        phone: payload.phone,
        notes: payload.notes,
        channel: payload.channel,
        mode,
        shippingFee: payload.summary?.shippingFee ?? 0,
        discount: (payload.summary?.discount ?? 0) + (payload.summary?.shippingDiscount ?? 0)
      });

      // WhatsApp: just submit the order, no transaction.
      if (isWhatsapp) {
        let submitted = saved;
        try {
          submitted = await submitOrder(saved);
        } catch (e) {
          if (!isTestEnv) throw e;
        }
        finalize(submitted, items, payload, false);
        return { id: String(submitted.order.id), number: submitted.order.id };
      }

      let prepared = await preparePayment(saved, method);

      // TEST_ENV: every method (pix, card, cash) always passes.
      if (isTestEnv) {
        prepared = await settlePayment(prepared);
        finalize(prepared, items, payload, true);
        return { id: String(prepared.order.id), number: prepared.order.id };
      }

      // Production PIX: show QR code and wait for payment.
      if (prepared.paymentMethod === 'pix') {
        pixContext = { items, payload };
        pixPending = prepared;
        return { id: String(prepared.order.id), number: prepared.order.id };
      }

      // Production card / cash: order is sent, payment happens at the restaurant.
      finalize(prepared, items, payload, false);
      return { id: String(prepared.order.id), number: prepared.order.id };
    } catch (err) {
      console.error('Failed to place order:', err);
      throw err; // Checkout shows err.message
    }
  }

  function handleConfirmPopupOk() {
    if (!confirmedOrder) return;
    const { restaurantId: restId, orderId } = confirmedOrder;
    confirmedOrder = null;
    if (onOrderConfirmed) {
      onOrderConfirmed(restId, orderId);
    } else {
      window.history.pushState({}, '', `/me/pedidos/${restId}/${orderId}`);
      window.dispatchEvent(new PopStateEvent('popstate'));
    }
  }

  function handleNavigate(path: string) {
    if (confirmedOrder) return;
    if (path === '/cardapio' || path === '/order' || path === '/') {
      onCancel();
    } else {
      window.history.pushState({}, '', path);
      window.dispatchEvent(new PopStateEvent('popstate'));
    }
  }

  async function copyPix() {
    try {
      await navigator.clipboard.writeText(pixPending!.pix!.copy_paste_code);
      pixCopied = true;
      setTimeout(() => (pixCopied = false), 2000);
    } catch {
      /* clipboard unavailable */
    }
  }

  function leavePix() {
    // The order stays saved; the cart is kept so the customer can retry.
    pixPending = null;
    pixContext = null;
    onCancel();
  }
</script>

{#if whatsappOrder}
  <WhatsAppSent
    order={whatsappOrder}
    restaurantName={restaurantInfo?.name}
    onNavigate={handleNavigate}
  />
{:else if pixPending?.pix}
  <div class="min-h-screen bg-zinc-100">
    <header class="sticky top-0 z-30 border-b border-zinc-200 bg-white/95 backdrop-blur">
      <div class="mx-auto flex max-w-6xl items-center gap-3 px-4 py-3 sm:px-6">
        <button onclick={leavePix} class="rounded-lg p-1.5 text-zinc-600 hover:bg-zinc-100" aria-label="Voltar">
          <ArrowLeft size={20} />
        </button>
        <h1 class="text-lg font-bold text-zinc-900">Pagamento PIX</h1>
      </div>
    </header>

    <main class="mx-auto max-w-xl px-4 py-6 sm:px-6">
      <div class="space-y-5 rounded-2xl bg-white p-6 text-center shadow-sm ring-1 ring-zinc-200 sm:p-8">
        <div class="space-y-1">
          <h2 class="text-2xl font-black text-zinc-900">Pague com PIX</h2>
          <p class="text-sm text-zinc-600">
            Total <b>{(pixPending.order.total_cents / 100).toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })}</b>
          </p>
        </div>

        {#if pixPending.pix.qr_code_base64}
          <img
            class="mx-auto h-48 w-48"
            src={`data:image/png;base64,${pixPending.pix.qr_code_base64}`}
            alt="Código QR PIX"
          />
        {/if}

        <div class="space-y-2 text-left">
          <label for="pix-code" class="block font-bold text-zinc-900">PIX copia e cola</label>
          <input
            id="pix-code"
            class="w-full rounded-xl border border-zinc-200 bg-white p-3 text-sm text-zinc-900"
            readonly
            value={pixPending.pix.copy_paste_code}
          />
          <button type="button" class="font-bold text-yellow-600 hover:text-yellow-700" onclick={copyPix}>
            {pixCopied ? 'Copiado!' : 'Copiar código'}
          </button>
        </div>

        <p class="text-sm text-zinc-500" role="status">Aguardando confirmação do pagamento…</p>
      </div>
    </main>
  </div>
{:else}
  <Checkout
    {cartLines}
    subtotal={isNaN($cart_total) ? 0 : $cart_total}
    {clientSession}
    {restaurantInfo}
    onPlaceOrder={handlePlaceOrder}
    onNavigate={handleNavigate}
    {onCancel}
  />
{/if}

{#if confirmedOrder}
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4 backdrop-blur-xs">
    <div role="dialog" aria-modal="true" aria-labelledby="order-confirmed-title" class="w-full max-w-md rounded-3xl bg-surface p-8 shadow-2xl space-y-5 border border-border">
      <div class="text-center space-y-2">
        <div class="mx-auto w-16 h-16 rounded-full bg-emerald-500/10 text-emerald-600 flex items-center justify-center mb-3">
          <svg class="w-8 h-8" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
          </svg>
        </div>
        <h2 id="order-confirmed-title" class="text-2xl font-bold text-text">Pedido confirmado!</h2>
        <p class="text-sm text-text-muted">
          {#if confirmedOrder.paid}
            Seu pagamento foi confirmado com sucesso em {confirmedOrder.restaurantName}.
          {:else}
            Seu pedido foi enviado para {confirmedOrder.restaurantName}. O pagamento é feito diretamente com o restaurante.
          {/if}
        </p>
      </div>

      <div class="bg-surface-hover border border-border rounded-2xl py-4 px-4 text-center space-y-1">
        <span class="text-xs uppercase tracking-wider text-text-muted font-bold block">Código de retirada / identificação</span>
        <span class="text-3xl font-black text-primary tracking-widest font-mono">{confirmedOrder.displayCode}</span>
        <p class="text-[11px] text-text-muted mt-1">Guarde este código para acompanhar e retirar seu pedido</p>
      </div>

      <button
        type="button"
        onclick={handleConfirmPopupOk}
        class="w-full rounded-xl bg-primary hover:bg-primary-hover px-5 py-3.5 font-bold text-white transition cursor-pointer shadow-lg shadow-primary/20 text-base"
      >
        OK
      </button>
    </div>
  </div>
{/if}
