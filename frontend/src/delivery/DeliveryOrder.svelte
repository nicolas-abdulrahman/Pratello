<script>
  
  // Prop to notify the parent section to open the form
  export let onCheckout = () => {};

  $: cartTotal = cart.reduce((total, item) => total + (item.price * item.quantity), 0);
</script>

<div class="bg-white border border-gray-200 rounded-2xl p-5 shadow-sm">
  <h2 class="text-lg font-bold text-gray-900 mb-4">Seu Pedido</h2>
  
  {#if cart.length === 0}
    <div class="text-center py-6 px-4">
      <div class="w-12 h-12 rounded-full bg-gray-50 flex items-center justify-center mx-auto mb-3 text-gray-400 border border-gray-100">
        <svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M6 2 3 6v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2V6l-3-4Z"/><path d="M3 6h18"/><path d="M16 10a4 4 0 0 1-8 0"/></svg>
      </div>
      <h3 class="text-sm font-bold text-gray-800 mb-1">Algo delicioso te espera</h3>
      <p class="text-xs text-gray-400 max-w-xs mx-auto">Adicione seus favoritos do cardápio e nós cuidamos do resto.</p>
    </div>
  {:else}
    <div class="space-y-3 mb-4">
      {#each cart as item}
        <div class="flex items-center justify-between text-sm">
          <div class="flex items-center gap-2">
            <span class="font-bold text-primary">{item.quantity}x</span>
            <span class="font-medium text-gray-800">{item.name}</span>
          </div>
          <span class="font-semibold text-gray-900">R${(item.price * item.quantity).toFixed(2).replace('.', ',')}</span>
        </div>
      {/each}
    </div>
  {/if}

  <div class="mt-4 pt-4 border-t border-gray-100">
    {#if cartTotal >= restaurant.minOrder}
      <button 
        onclick={onCheckout}
        class="w-full py-btn-y bg-primary hover:bg-primary-hover text-white font-semibold text-sm rounded-xl shadow-md text-center transition"
      >
        Finalizar Pedido (R${cartTotal.toFixed(2).replace('.', ',')})
      </button>
    {:else}
      <button class="w-full py-3 bg-gray-100 text-gray-400 font-semibold text-xs rounded-xl cursor-not-allowed text-center transition" disabled>
        Mínimo R${restaurant.minOrder.toFixed(2).replace('.', ',')} para entrega
      </button>
    {/if}
  </div>
</div>
