<script lang="ts">
  import { onMount } from 'svelte';
  import { account } from '../../lib/stores/account';
  import { transactions, refreshOrder, isPaid, type SavedOrder } from '../../lib/stores/Transaction';
  export interface OrderItem {
    id: string;
    name: string;
    description?: string;
    quantity: number;
    price: number;
    imageUrl: string;
  }

  export interface HistoricalOrder {
    id: string;
    orderNumber: string;
    date: string;
    restaurantName: string;
    status: 'Entregue' | 'Cancelado' | 'Em andamento' | '';
    items: OrderItem[];
    total: number;
    isReviewed?: boolean;
  }

  let {
    onSelectOrder = (order: HistoricalOrder) => {}
  }: {
    onSelectOrder?: (order: HistoricalOrder) => void;
  } = $props();

  // Mock data of past orders
  const unusedMockOrders: HistoricalOrder[] = [
    {
      id: 'ord-1042',
      orderNumber: '#1042',
      date: 'Hoje, 12:45',
      restaurantName: 'Pratello Gourmet Burger',
      status: 'Entregue',
      total: 54.9,
      items: [
        {
          id: 'item-combo-1',
          name: 'Combo Smash Bacon Duplo',
          description: 'Pão brioche, 2x smash 90g, cheddar cremoso, bacon crocante + Batata frita M + Refrigerante',
          quantity: 1,
          price: 39.9,
          imageUrl: 'https://images.unsplash.com/photo-1568901346375-23c9450c58cd?w=400&auto=format&fit=crop&q=80'
        },
        {
          id: 'item-icecream-1',
          name: 'Sorvete Artesanal de Baunilha com Calda',
          description: 'Taça de sorvete premium com calda quente de chocolate e crocante de castanhas',
          quantity: 1,
          price: 15.0,
          imageUrl: 'https://images.unsplash.com/photo-1570197788417-0e82375c9371?w=400&auto=format&fit=crop&q=80'
        }
      ]
    },
    {
      id: 'ord-1028',
      orderNumber: '#1028',
      date: '22 de Nov, 19:30',
      restaurantName: 'Pratello Smash & Fries',
      status: 'Entregue',
      total: 48.0,
      items: [
        {
          id: 'item-combo-2',
          name: 'Combo Chicken Supreme Crocante',
          description: 'Sobrecoxa empanada ultra crocante, salada coleslaw, maionese verde + Batata rústica',
          quantity: 1,
          price: 36.0,
          imageUrl: 'https://images.unsplash.com/photo-1625813506062-0aeb1d7a094b?w=400&auto=format&fit=crop&q=80'
        },
        {
          id: 'item-icecream-2',
          name: 'Sundae Especial de Morango',
          description: 'Sorvete cremoso com pedaços de morango fresco e chantilly',
          quantity: 1,
          price: 12.0,
          imageUrl: 'https://images.unsplash.com/photo-1563805042-7684c019e1cb?w=400&auto=format&fit=crop&q=80'
        }
      ]
    },
    {
      id: 'ord-0985',
      orderNumber: '#0985',
      date: '17 de Nov, 20:15',
      restaurantName: 'Pratello Burger',
      status: 'Entregue',
      total: 62.5,
      items: [
        {
          id: 'item-combo-3',
          name: 'Combo Cheese Picanha Blend',
          description: 'Blend de picanha 180g, queijo prato maçaricado, cebola caramelizada + Batata crinkle',
          quantity: 1,
          price: 44.5,
          imageUrl: 'https://images.unsplash.com/photo-1586190848861-99aa4a171e90?w=400&auto=format&fit=crop&q=80'
        },
        {
          id: 'item-icecream-3',
          name: 'Sorvete de Pistache Italiano',
          description: 'Gelato de pistache puro com praliné crocante',
          quantity: 1,
          price: 18.0,
          imageUrl: 'https://images.unsplash.com/photo-1501443762994-82bd5dace89a?w=400&auto=format&fit=crop&q=80'
        }
      ]
    }
  ];

  let visible = $derived($transactions.filter((o) => $account.isLoggedIn ? o.userId === $account.id : o.userId === null));
  let mockOrders = $derived(visible.map(toHistoricalOrder));

  export function toHistoricalOrder(saved: SavedOrder): HistoricalOrder {
    const isDone = saved.order.status === 'delivered' || saved.order.status === 'ready' || !!saved.pickupAcknowledgedAt || isPaid(saved.order);
    const isReviewed = !!saved.isReviewed || !!saved.serviceReviewed;
    return {
      id: String(saved.order.id),
      orderNumber: `#${saved.order.id}`,
      date: new Date(saved.order.created_at).toLocaleString('pt-BR'),
      restaurantName: saved.restaurantName,
      status: saved.order.status === 'cancelled' ? 'Cancelado' : isDone ? 'Entregue' : '',
      total: saved.order.total_cents / 100,
      isReviewed,
      items: (saved.order.items || []).filter((item) => !item.voided_at).map((item) => ({
        id: String(item.id), name: item.item_name, quantity: item.quantity,
        price: item.unit_price_cents / 100, imageUrl: '',
      })),
    };
  }

  onMount(() => {
    for (const saved of visible) void refreshOrder(saved).catch(() => {});
  });

  function formatPrice(val: number) {
    return val.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' });
  }
</script>

<section class="space-y-6">
  <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-border pb-4">
    <div>
      <h1 class="text-2xl sm:text-3xl font-extrabold text-text tracking-tight">Meus Pedidos</h1>
      <p class="text-sm text-text-muted mt-1">Acompanhe seus pedidos anteriores e deixe sua avaliação</p>
    </div>
    <div class="inline-flex items-center gap-2 bg-surface-hover text-primary text-xs font-semibold px-3 py-1.5 rounded-full self-start sm:self-auto border border-border">
      <svg class="w-4 h-4 text-primary" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
        <path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
      {mockOrders.length} pedidos realizados
    </div>
  </div>

  <div class="grid gap-5">
    {#each mockOrders as order (order.id)}
      <div
        class="bg-surface rounded-2xl border border-border shadow-xs hover:shadow-md hover:border-primary transition-all duration-200 overflow-hidden group"
      >
        <!-- Header do Pedido -->
        <div class="p-4 sm:p-5 bg-surface-hover border-b border-border flex flex-wrap items-center justify-between gap-3">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-xl bg-surface-hover flex items-center justify-center text-primary font-bold text-sm">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                <path d="M16 11V7a4 4 0 00-8 0v4M5 9h14l1 12H4L5 9z" />
              </svg>
            </div>
            <div>
              <div class="flex items-center gap-2">
                <span class="font-bold text-text">{order.restaurantName}</span>
                <span class="text-xs text-text-muted font-mono">{order.orderNumber}</span>
              </div>
              <p class="text-xs text-text-muted">{order.date}</p>
            </div>
          </div>

          <div class="flex items-center gap-3">
            {#if order.status === 'Cancelado'}
              <span class="inline-flex items-center gap-1.5 text-xs font-medium px-2.5 py-1 rounded-full bg-rose-500/10 text-rose-600 border border-rose-500/20">
                Cancelado
              </span>
            {:else if order.isReviewed}
              <span class="inline-flex items-center gap-1.5 text-xs font-medium px-2.5 py-1 rounded-full bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20">
                <svg class="w-3.5 h-3.5 text-emerald-500" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
                </svg>
                Avaliado
              </span>
            {/if}
            <span class="font-bold text-text text-sm">{formatPrice(order.total)}</span>
          </div>
        </div>

        <!-- Itens do Pedido -->
        <div class="p-4 sm:p-5 divide-y divide-border">
          {#each order.items as item (item.id)}
            <div class="py-3 first:pt-0 last:pb-0 flex items-center gap-4">
              {#if item.imageUrl}<img
                src={item.imageUrl}
                alt={item.name}
                class="w-14 h-14 rounded-xl object-cover border border-border flex-shrink-0 bg-surface-hover"
                loading="lazy"
              />{/if}
              <div class="flex-1 min-w-0">
                <h4 class="text-sm font-semibold text-text truncate">{item.name}</h4>
                {#if item.description}
                  <p class="text-xs text-text-muted line-clamp-1 mt-0.5">{item.description}</p>
                {/if}
                <div class="flex items-center gap-3 mt-1 text-xs text-text-muted">
                  <span>Qtd: {item.quantity}</span>
                  <span>•</span>
                  <span class="font-medium text-text">{formatPrice(item.price)}</span>
                </div>
              </div>
            </div>
          {/each}
        </div>

        <!-- Ação: Avaliar Pedido -->
        <div class="px-4 sm:px-5 py-3.5 bg-surface-hover border-t border-border flex items-center justify-between">
          <span class="text-xs text-text-muted">
            {#if order.status === 'Cancelado'}
              Pedido cancelado
            {:else if order.isReviewed}
              Pedido avaliado. Obrigado pelo seu feedback!
            {:else}
              Avalie sua experiência e ajude a melhorar nosso serviço
            {/if}
          </span>
          {#if visible.find((entry) => String(entry.order.id) === order.id && !isPaid(entry.order) && entry.order.status !== 'cancelled')}
            <a href={`/checkout?resume=${order.id}`} class="text-primary font-bold text-xs">Continuar pagamento</a>
          {/if}
          <button
            type="button"
            disabled={order.status === 'Cancelado' || !!order.isReviewed}
            onclick={() => onSelectOrder(order)}
            class="inline-flex items-center gap-2 px-4 py-2 {order.isReviewed ? 'bg-surface border border-border text-text-muted cursor-not-allowed opacity-80' : 'bg-primary hover:bg-primary-hover active:bg-primary-hover text-white shadow-xs cursor-pointer group-hover:scale-102'} font-semibold text-xs rounded-xl transition disabled:opacity-60"
          >
            {#if order.isReviewed}
              <svg class="w-4 h-4 text-emerald-500" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
              </svg>
              <span>Avaliado</span>
            {:else}
              <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                <path d="M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z" />
              </svg>
              <span>Avaliar Pedido</span>
            {/if}
          </button>
        </div>
      </div>
    {/each}
  </div>
</section>
