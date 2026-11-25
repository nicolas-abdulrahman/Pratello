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
        if (onOrderConfirmed) {
          onOrderConfirmed(restId, orderId);
        } else {
          window.history.pushState({}, '', `/me/pedidos/${restId}/${orderId}`);
          window.dispatchEvent(new PopStateEvent('popstate'));
        }
      }

      return { id: orderId, number: orderNumber };
    } catch (err) {
      console.error('Failed to place order:', err);
      return null;
    }
  }

  function handleNavigate(path: string) {
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
