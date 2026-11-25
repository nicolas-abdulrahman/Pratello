<script lang="ts">
  import { onMount } from 'svelte';
  import { transactions, refreshOrder } from '../../lib/stores/Transaction';
  import RateOrder from '../RateOrder.svelte';

  let {
    restaurantId = '',
    orderId = '',
    onGoOrders = () => {},
    onGoHome = () => {}
  }: {
    restaurantId?: string | number;
    orderId?: string | number;
    onGoOrders?: () => void;
    onGoHome?: () => void;
  } = $props();

  let saved = $derived($transactions.find((entry) => String(entry.order.id) === String(orderId)) ?? null);

  // Poll current order status
  onMount(() => {
    const poll = async () => {
      const current = $transactions.find((entry) => String(entry.order.id) === String(orderId));
      if (current && current.order.status !== 'cancelled' && current.order.status !== 'delivered') {
        try {
          await refreshOrder(current);
        } catch {}
      }
    };
    void poll();
    const interval = setInterval(poll, import.meta.env.TEST_ENV === 'true' || import.meta.env.TEST_ENV === '1' ? 3000 : 10000);

    return () => clearInterval(interval);
  });

  const money = (cents: number) => (cents / 100).toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' });
</script>

<main class="min-h-screen bg-background text-text py-8 px-4 sm:px-6 lg:px-8 transition-colors duration-200">
  <div class="max-w-2xl mx-auto space-y-6">
    <!-- Breadcrumb header -->
    <div class="flex items-center justify-between gap-2 text-sm text-text-muted">
      <div class="flex items-center gap-2">
        <button
          type="button"
          onclick={onGoOrders}
          class="hover:text-primary transition flex items-center gap-1 font-medium cursor-pointer"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" d="M15 19l-7-7 7-7" />
          </svg>
          Meus Pedidos
        </button>
        <span>/</span>
        <span class="text-text font-medium">Acompanhamento</span>
      </div>

      <span class="text-xs text-text-muted font-mono bg-surface-hover px-2.5 py-1 rounded-lg border border-border">
        Pedido #{orderId}
      </span>
    </div>

    <!-- Main Order Status Card -->
    <div class="bg-surface rounded-3xl border border-border p-6 sm:p-8 shadow-xs space-y-6 transition-colors duration-200">
      <!-- Status pill & Icon -->
      <div class="flex flex-col items-center text-center space-y-3">
        <div class="relative">
          <div class="w-20 h-20 bg-primary-soft text-primary rounded-3xl flex items-center justify-center shadow-inner">
            <svg class="w-10 h-10 animate-pulse" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
          <span class="absolute -bottom-1 -right-1 flex h-4 w-4">
            <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-primary opacity-75"></span>
            <span class="relative inline-flex rounded-full h-4 w-4 bg-primary"></span>
          </span>
        </div>

        <div class="space-y-1">
          <span class="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-bold uppercase tracking-wider bg-primary-soft text-primary border border-primary/20">
            {saved?.order.status === 'ready' ? 'Pedido Pronto!' : saved?.order.status === 'delivered' ? 'Entregue' : 'Em Preparação'}
          </span>
          <h1 class="text-2xl sm:text-3xl font-extrabold text-text tracking-tight">
            O pedido está sendo preparado!
          </h1>
          <p class="text-xs sm:text-sm text-text-muted max-w-md mx-auto">
            {saved?.restaurantName ? `${saved.restaurantName} já` : 'A cozinha já'} confirmou sua transação e os itens estão em preparo.
          </p>
        </div>
      </div>

      <!-- Prominent Code Display -->
      <div class="bg-surface-hover border border-border rounded-2xl p-5 sm:p-6 text-center space-y-2">
        <p class="text-xs font-bold uppercase tracking-wider text-text-muted">
          Código de retirada / identificação
        </p>
        {#if saved?.order.display_code}
          <p class="text-4xl sm:text-5xl font-black text-primary tracking-widest font-mono">
            {saved.order.display_code}
          </p>
        {:else}
          <p class="text-3xl sm:text-4xl font-extrabold text-primary font-mono">
            #{orderId}
          </p>
        {/if}
        <p class="text-[11px] text-text-muted">Apresente este código quando for retirar seu pedido no balcão</p>
      </div>

      <!-- Items Summary Accordion / List if available -->
      {#if saved?.order.items && saved.order.items.length > 0}
        <div class="pt-2 border-t border-border">
          <h3 class="text-xs font-bold text-text-muted uppercase tracking-wider mb-3">Itens do pedido</h3>
          <div class="divide-y divide-border">
            {#each saved.order.items.filter((item) => !item.voided_at) as item}
              <div class="py-2.5 flex items-center justify-between text-xs sm:text-sm">
                <span class="text-text font-medium">{item.quantity}x {item.item_name}</span>
                <span class="text-text-muted font-mono">{money(item.unit_price_cents * item.quantity)}</span>
              </div>
            {/each}
          </div>
          <div class="pt-3 mt-2 border-t border-border flex items-center justify-between text-sm font-bold text-text">
            <span>Total</span>
            <span class="font-mono text-base text-primary">{money(saved.order.total_cents)}</span>
          </div>
        </div>
      {/if}
    </div>

    <!-- Rating Section: RateOrder component under it -->
    <RateOrder
      {restaurantId}
      {orderId}
      restaurantName={saved?.restaurantName}
      guestToken={saved?.guestToken}
    />

    <!-- Bottom Actions -->
    <div class="flex items-center justify-between gap-4 pt-2">
      <button
        type="button"
        onclick={onGoOrders}
        class="text-xs sm:text-sm font-bold text-text-muted hover:text-text transition flex items-center gap-1.5 cursor-pointer"
      >
        ← Ver todos os meus pedidos
      </button>

      <button
        type="button"
        onclick={onGoHome}
        class="text-xs sm:text-sm font-bold text-primary hover:text-primary-hover transition cursor-pointer"
      >
        Ir para a página inicial →
      </button>
    </div>
  </div>
</main>
