<script lang="ts">
  import { cart, cart_total, remove_item } from '../lib/stores/Cart';
  import OrderSidebarItem from './OrderSidebarItem.svelte';

  let {
    isOpen = false,
    onClose = () => {},
    goToOrder = () => {},
  }: {
    isOpen?: boolean;
    onClose?: () => void;
    goToOrder?: (mode: 'dine-in' | 'delivery') => void;
  } = $props();

  function handleCheckout(mode: 'dine-in' | 'delivery') {
    goToOrder(mode);
    onClose();
  }

  function formatBRL(v: number): string {
    return new Intl.NumberFormat('pt-BR', {
      style: 'currency',
      currency: 'BRL',
    }).format(v);
  }
</script>

{#if isOpen}
  <div
    class="fixed inset-0 bg-black/40 backdrop-blur-sm z-40 transition-opacity duration-300"
    onclick={onClose}
    role="button"
    tabindex="0"
    aria-label="Fechar carrinho"
  ></div>
{/if}

<div
  class="fixed top-0 right-0 h-full w-full max-w-md bg-surface shadow-2xl z-50 transform transition-transform duration-300 ease-in-out flex flex-col
  {isOpen ? 'translate-x-0' : 'translate-x-full'}"
>
  <div class="flex items-center justify-between p-6 border-b border-border">
    <h2 class="text-2xl font-black text-text flex items-center gap-2">
      <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-primary">
        <circle cx="8" cy="21" r="1" />
        <circle cx="19" cy="21" r="1" />
        <path d="M2.05 2.05h2l2.66 12.42a2 2 0 0 0 2 1.58h9.78a2 2 0 0 0 1.95-1.57l1.65-7.43H5.12" />
      </svg>
      Seu Pedido
    </h2>
    <button
      onclick={onClose}
      class="p-2 text-text-muted hover:text-text bg-surface-hover hover:bg-border rounded-full transition-colors"
      aria-label="Fechar"
    >
      <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M18 6 6 18" />
        <path d="m6 6 12 12" />
      </svg>
    </button>
  </div>

  <div class="flex-1 overflow-y-auto p-6 space-y-4">
    {#if $cart.length === 0}
      <div class="text-center py-10">
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
    {/if}

    {#each $cart as line (line.id)}
      <OrderSidebarItem {line} onRemove={remove_item} />
    {/each}
  </div>

  <div class="p-6 bg-surface-hover border-t border-border">
    <div class="flex justify-between items-center mb-4">
      <span class="text-text-muted font-medium">Total</span>
      <span class="text-3xl font-black text-text">{$cart_total ? formatBRL($cart_total) : 'R$ 0,00'}</span>
    </div>

    <div class="flex gap-3 w-full">
      <button
        onclick={() => handleCheckout('dine-in')}
        class="w-[70%] bg-primary hover:bg-primary-hover text-white font-bold text-sm py-3 rounded-xl transition-colors shadow-lg shadow-primary/20 flex items-center justify-center gap-2"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M3 2v7c0 1.1.9 2 2 2h4a2 2 0 0 0 2-2V2" />
          <path d="M7 2v20" />
          <path d="M21 15V2v0a5 5 0 0 0-5 5v6c0 1.1.9 2 2 2h3Zm0 0v7" />
        </svg>
        No Restaurante
      </button>

      <button
        onclick={() => handleCheckout('delivery')}
        class="w-[30%] bg-primary-soft hover:bg-border text-primary-hover font-bold text-sm py-3 rounded-xl transition-colors flex items-center justify-center gap-1.5"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M5 18H3c-.6 0-1-.4-1-1V7c0-.6.4-1 1-1h10c.6 0 1 .4 1 1v11" />
          <path d="M14 9h4l4 4v4c0 .6-.4 1-1 1h-2" />
          <circle cx="7" cy="18" r="2" />
          <circle cx="17" cy="18" r="2" />
          <path d="M9 18h6" />
        </svg>
        Deliver
      </button>
    </div>
  </div>
</div>
