<script>
  import { cart } from '../lib/stores/Cart';
  import OrderSidebar from './OrderSidebar.svelte';

  // Accept the function from App.svelte
  let { goToOrder = () => {} } = $props();

  let isCartOpen = $state(false);

  // Automatically calculate total items (sum of quantities)
  let cartItemCount = $derived($cart.reduce((total, item) => total + item.quantity, 0));
</script>

<button 
  onclick={() => isCartOpen = true}
  class="fixed top-6 right-4 md:right-8 z-40 flex items-center gap-2 bg-orange-50 hover:bg-orange-100 text-[#ff5722] border border-orange-100 px-4 py-2.5 rounded-full font-bold transition-all shadow-lg hover:scale-105"
>
  <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
    <circle cx="8" cy="21" r="1"/>
    <circle cx="19" cy="21" r="1"/>
    <path d="M2.05 2.05h2l2.66 12.42a2 2 0 0 0 2 1.58h9.78a2 2 0 0 0 1.95-1.57l1.65-7.43H5.12"/>
  </svg>
  
  {#if cartItemCount > 0}
    <span class="bg-[#ff5722] text-white text-xs w-5 h-5 flex items-center justify-center rounded-full absolute -top-1 -left-1 shadow-sm">
      {cartItemCount}
    </span>
  {/if}
  
  <span class="hidden sm:block">Carrinho</span>
</button>

<!-- Pass the function down to the Sidebar -->
<OrderSidebar 
  isOpen={isCartOpen} 
  onClose={() => isCartOpen = false} 
  {goToOrder} 
/>
