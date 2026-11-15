<script lang="ts">
  import { cart, cart_total, remove_item } from '../lib/stores/Cart';
  import OrderSidebarItem from './OrderSidebarItem.svelte';

  let {
    goToCheckout = () => {},
    onCancel = () => {},
  }: {
    goToCheckout?: () => void;
    onCancel?: () => void;
  } = $props();

  function formatBRL(v: number): string {
    return new Intl.NumberFormat('pt-BR', {
      style: 'currency',
      currency: 'BRL',
    }).format(v);
  }
</script>

<div class="min-h-screen bg-background font-sans selection:bg-primary selection:text-white pb-12 pt-6 px-4 md:px-8">
  <div class="max-w-3xl mx-auto space-y-6">
    <div class="flex items-center gap-4 mb-4">
      <button
        onclick={onCancel}
        class="p-2.5 bg-surface border border-border rounded-full shadow-sm hover:text-primary transition-colors"
        aria-label="Voltar"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="m15 18-6-6 6-6" />
        </svg>
      </button>
      <h1 class="text-2xl font-black text-text">Seu Pedido</h1>
    </div>

    {#if $cart.length === 0}
      <div class="bg-surface rounded-3xl p-10 text-center border border-border">
        <div class="w-16 h-16 rounded-full flex items-center justify-center mx-auto mb-4 bg-surface-hover text-text-muted">
          <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="8" cy="21" r="1" />
            <circle cx="19" cy="21" r="1" />
            <path d="M2.05 2.05h2l2.66 12.42a2 2 0 0 0 2 1.58h9.78a2 2 0 0 0 1.95-1.57l1.65-7.43H5.12" />
          </svg>
        </div>
        <h3 class="text-text font-bold mb-1">Seu carrinho está vazio</h3>
        <p class="text-sm text-text-muted">Adicione itens para continuar.</p>
      </div>
    {:else}
      <div class="space-y-4">
        {#each $cart as line (line.id)}
          <OrderSidebarItem {line} onRemove={remove_item} />
        {/each}
      </div>

      <div class="bg-surface rounded-3xl p-4 md:p-6 shadow-xl shadow-black/5 border border-border flex flex-col md:flex-row items-center justify-between gap-4 sticky bottom-4 z-30">
        <div class="flex flex-col flex-shrink-0 w-full md:w-[40%]">
          <div class="flex items-end gap-2">
            <span class="text-2xl md:text-3xl font-black text-primary leading-none">
              {formatBRL($cart_total)}
            </span>
            <span class="text-[10px] text-text-muted font-medium mb-1">no total</span>
          </div>
        </div>

        <button
          onclick={goToCheckout}
          class="w-full md:flex-1 bg-primary hover:bg-primary-hover text-white font-bold py-3.5 md:py-4 px-6 rounded-full shadow-lg hover:shadow-primary/30 transition-all flex items-center justify-center gap-2 text-sm md:text-base"
        >
          Ir para o Pagamento
          <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <path d="M5 12h14" />
            <path d="m12 5 7 7-7 7" />
          </svg>
        </button>
      </div>
    {/if}
  </div>
</div>
