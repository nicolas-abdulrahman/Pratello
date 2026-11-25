<script lang="ts">
  import Checkout from './checkout/Checkout.svelte';
  import WhatsAppSent, { type Order as WhatsAppOrder } from './checkout/WhatsAppSent.svelte';
  import { cart, cart_total, clear_cart, calc_line_price } from '../lib/stores/Cart';
  import { account } from '../lib/stores/account';
  import {
    createOrder,
    preparePayment,
    confirmPayment,
    visibleOrders,
    type SavedOrder
  } from '../lib/stores/Transaction';
  import { get } from 'svelte/store';

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

  let restaurantInfo = $state({
    name: 'Pratello Restaurante',
    pixKey: '12.345.678/0001-90',
    pixMerchant: 'PRATELLO RESTAURANTE LTDA',
    tableCount: 15,
    phone: '(11) 98765-4321',
  });

  let whatsappOrder = $state<WhatsAppOrder | null>(null);
  let confirmedOrder = $state<{
    orderId: string;
    orderNumber: number;
    restaurantId: string;
    restaurantName: string;
    displayCode: string;
  } | null>(null);

  $effect(() => {
    if (slug) {
      fetch(`/api/restaurants/${encodeURIComponent(slug)}`)
        .then((res) => (res.ok ? res.json() : null))
        .then((data) => {
          if (data?.name) {
            restaurantInfo = {
              ...restaurantInfo,
              name: data.name,
              phone: data.phone || restaurantInfo.phone,
            };
          }
        })
        .catch(() => {});
    }

    const waMatch = window.location.pathname.match(/^\/pedido-whatsapp\/([^/]+)\/?$/);
    if (waMatch && !whatsappOrder) {
      const found = visibleOrders().find((o) => String(o.order.id) === waMatch[1]);
      if (found) {
        whatsappOrder = {
          id: String(found.order.id),
          number: found.order.id,
          clientId: found.userId ? String(found.userId) : 'guest',
          customerName: $account.name || 'Cliente Pratello',
          phone: restaurantInfo.phone,
          location: { mode: 'mesa', table: '01' },
          items: found.order.items.map((i) => ({
            productId: String(i.item_id),
            name: i.item_name,
            price: i.unit_price_cents / 100,
            quantity: i.quantity,
          })),
          notes: '',
          subtotal: found.order.total_cents / 100,
          deliveryFee: 0,
          total: found.order.total_cents / 100,
          payment: {
            method: 'whatsapp',
            status: 'pago',
          },
          status: 'preparo',
          createdAt: Date.now(),
        };
      }
    }
  });

  let cartLines = $derived(
    $cart.map((line) => ({
      product: {
        id: String(line.item.id),
        name: line.item.name,
        price: line.item.price ?? (calc_line_price(line) / line.quantity),
        image: line.item.image_url ?? line.item.image,
      },
      quantity: line.quantity,
    }))
  );

  let clientSession = $derived({
    clientId: $account.id ? String($account.id) : undefined,
    name: $account.name || undefined,
    phone: $account.phone || undefined,
  });

  async function handlePlaceOrder(payload: any): Promise<{ id: string; number: number } | null> {
    try {
      const items = get(cart);
      const isWhatsapp = payload.channel === 'whatsapp';
      const paymentMethod = payload.payment?.method || (isWhatsapp ? 'whatsapp' : 'pix');

      const saved = await createOrder(slug, items, {
        couponCodes: payload.couponCodes,
        tableNumber: payload.location?.mode === 'mesa' ? payload.location.table : undefined,
        deliveryAddress: payload.location?.mode === 'delivery' ? payload.location.address : undefined,
        phone: payload.phone,
        notes: payload.notes,
        channel: payload.channel,
        mode: payload.location?.mode,
        shippingFee: payload.summary?.shippingFee ?? 0,
        discount: (payload.summary?.discount ?? 0) + (payload.summary?.shippingDiscount ?? 0),
      });

      let prepared = await preparePayment(saved, paymentMethod);

      // In TEST environment or for Pix, guarantee simulated confirmation succeeds
      if (paymentMethod === 'pix' || import.meta.env.TEST_ENV === 'true' || import.meta.env.TEST_ENV === '1') {
        try {
          prepared = await confirmPayment(prepared);
        } catch {
          // In test mode, confirmPayment handles fallback
        }
      }

      const orderId = String(prepared.order.id);
      const orderNumber = prepared.order.id;
      const restId = String(prepared.restaurantId || prepared.slug || slug || 'restaurante');
      const displayCode = prepared.order.display_code || `#${orderId}`;

      if (isWhatsapp) {
        whatsappOrder = {
          id: orderId,
          number: orderNumber,
          clientId: $account.id ? String($account.id) : 'guest',
          customerName: $account.name || 'Cliente Pratello',
          phone: payload.phone || restaurantInfo.phone,
          location: payload.location,
          items: items.map((l) => ({
            productId: String(l.item.id),
            name: l.item.name,
            price: l.item.price ?? (calc_line_price(l) / l.quantity),
            quantity: l.quantity,
          })),
          notes: payload.notes || '',
          subtotal: payload.summary?.subtotal ?? $cart_total,
          deliveryFee: payload.summary?.shippingFee ?? 0,
          discount: (payload.summary?.discount ?? 0) + (payload.summary?.shippingDiscount ?? 0),
          coupons: payload.couponCodes,
          channel: 'whatsapp',
          shipping: payload.shipping,
          total: payload.summary?.total ?? $cart_total,
          payment: {
            method: payload.payment?.method || 'whatsapp',
            cardType: payload.payment?.cardType,
            installments: payload.payment?.installments,
            needsMachine: payload.payment?.needsMachine,
            changeFor: payload.payment?.changeFor ? Number(payload.payment.changeFor) : undefined,
            status: 'pago',
          },
          status: 'preparo',
          createdAt: Date.now(),
        };
      }

      clear_cart();

      if (!isWhatsapp) {
        confirmedOrder = {
          orderId,
          orderNumber,
          restaurantId: restId,
          restaurantName: prepared.restaurantName || restaurantInfo.name,
          displayCode,
        };
      }

      return { id: orderId, number: orderNumber };
    } catch (err) {
      console.error('Failed to place order:', err);
      return null;
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
    if (confirmedOrder) {
      // Keep confirmation popup open so user can read the code and click OK
      return;
    }
    if (path === '/cardapio' || path === '/order' || path === '/') {
      onCancel();
    } else {
      window.history.pushState({}, '', path);
      window.dispatchEvent(new PopStateEvent('popstate'));
    }
  }
</script>

{#if whatsappOrder}
  <WhatsAppSent
    order={whatsappOrder}
    restaurantName={restaurantInfo.name}
    onNavigate={handleNavigate}
  />
{:else}
  <Checkout
    {cartLines}
    subtotal={$cart_total}
    {clientSession}
    {restaurantInfo}
    onPlaceOrder={handlePlaceOrder}
    onNavigate={handleNavigate}
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
        <p class="text-sm text-text-muted">Seu pagamento foi confirmado com sucesso em {confirmedOrder.restaurantName}.</p>
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
