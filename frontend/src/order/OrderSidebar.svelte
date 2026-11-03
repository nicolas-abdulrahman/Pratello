<script>
  // Svelte 5 Props
  let { 
    isOpen = false, 
    onClose = () => {} 
  } = $props();
  
  // Example Data
  const cartItems = [
    {
      id: 1,
      type: 'combo',
      title: 'Combo Delícia',
      price: 45.90,
      items: ['Hambúrguer Artesanal', 'Batata Frita Média']
    },
    {
      id: 2,
      type: 'single',
      title: 'Brownie com Sorvete',
      price: 18.50,
      items: []
    }
  ];

  // Calculate total dynamically
  let total = $derived(cartItems.reduce((acc, item) => acc + item.price, 0));
</script>

<!-- Backdrop -->
{#if isOpen}
  <div 
    class="fixed inset-0 bg-black/40 backdrop-blur-sm z-40 transition-opacity duration-300"
    onclick={onClose}
    onkeydown={(e) => e.key === 'Escape' && onClose()}
    role="button"
    tabindex="0"
    aria-label="Fechar carrinho"
  ></div>
{/if}

<!-- Sidebar Container -->
<div 
  class="fixed top-0 right-0 h-full w-full max-w-md bg-white shadow-2xl z-50 transform transition-transform duration-300 ease-in-out flex flex-col
  {isOpen ? 'translate-x-0' : 'translate-x-full'}"
>
  <!-- Header -->
  <div class="flex items-center justify-between p-6 border-b border-gray-100">
    <h2 class="text-2xl font-black text-gray-900 flex items-center gap-2">
      <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#ff5722" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="8" cy="21" r="1"/><circle cx="19" cy="21" r="1"/><path d="M2.05 2.05h2l2.66 12.42a2 2 0 0 0 2 1.58h9.78a2 2 0 0 0 1.95-1.57l1.65-7.43H5.12"/></svg>
      Seu Pedido
    </h2>
    <button onclick={onClose} class="p-2 text-gray-400 hover:text-gray-900 bg-gray-100 hover:bg-gray-200 rounded-full transition-colors">
      <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 6 6 18"/><path d="m6 6 12 12"/></svg>
    </button>
  </div>

  <!-- Cart Items Area -->
  <div class="flex-1 overflow-y-auto p-6 space-y-6">
    {#each cartItems as item}
      <div class="flex flex-col">
        <div class="flex justify-between items-start mb-1">
          <h3 class="font-bold text-gray-900 text-lg">{item.title}</h3>
          <span class="font-bold text-[#ff5722]">R$ {item.price.toFixed(2).replace('.', ',')}</span>
        </div>
        
        {#if item.items.length > 0}
          <ul class="ml-2 border-l-2 border-gray-200 pl-3 mt-1 space-y-1">
            {#each item.items as subItem}
              <li class="text-sm text-gray-500 flex items-center gap-2">
                <span class="w-1.5 h-1.5 rounded-full bg-gray-300"></span>
                {subItem}
              </li>
            {/each}
          </ul>
        {/if}
      </div>
    {/each}
  </div>

  <!-- Footer -->
  <div class="p-6 bg-gray-50 border-t border-gray-100">
    <div class="flex justify-between items-center mb-4">
      <span class="text-gray-500 font-medium">Total</span>
      <span class="text-3xl font-black text-gray-900">R$ {total.toFixed(2).replace('.', ',')}</span>
    </div>
    
    <!-- Action Buttons (70% / 30% Split) -->
    <div class="flex gap-3 w-full">
      <!-- Pedir no Restaurante (70% width) -->
      <button class="w-[70%] bg-[#ff5722] hover:bg-[#e64a19] text-white font-bold text-sm py-3 rounded-xl transition-colors shadow-lg shadow-orange-500/20 flex items-center justify-center gap-2">
        <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 2v7c0 1.1.9 2 2 2h4a2 2 0 0 0 2-2V2"/><path d="M7 2v20"/><path d="M21 15V2v0a5 5 0 0 0-5 5v6c0 1.1.9 2 2 2h3Zm0 0v7"/></svg>
        No Restaurante
      </button>

      <!-- Deliver (30% width) -->
      <button class="w-[30%] bg-orange-100 hover:bg-orange-200 text-[#d94e17] font-bold text-sm py-3 rounded-xl transition-colors flex items-center justify-center gap-1.5">
        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M5 18H3c-.6 0-1-.4-1-1V7c0-.6.4-1 1-1h10c.6 0 1 .4 1 1v11"/><path d="M14 9h4l4 4v4c0 .6-.4 1-1 1h-2"/><circle cx="7" cy="18" r="2"/><circle cx="17" cy="18" r="2"/><path d="M9 18h6"/></svg>
        Deliver
      </button>
    </div>
  </div>
</div>
