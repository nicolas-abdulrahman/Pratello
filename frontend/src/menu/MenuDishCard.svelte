<script>
  import ReviewProp from '../review/ReviewProp.svelte';
  import { addToCart } from '../lib/stores/cart.js';

  let { 
    image, 
    title, 
    description, 
    price, 
    rating = null, 
    count = null 
  } = $props();

  let isAdding = $state(false);

  function parsePrice(value) {
    if (typeof value === 'number') return value;
    if (!value) return 0;
    return Number(String(value).replace(/[^\d,.-]/g, '').replace(/\./g, '').replace(',', '.')) || 0;
  }

  function handleAdd() {
    if (isAdding) return;
    isAdding = true;
    addToCart({ name: title, icon: image, price: parsePrice(price) });
    setTimeout(() => { isAdding = false; }, 800);
  }
</script>

<div class="bg-white rounded-2xl p-3 flex flex-col sm:flex-row gap-4 shadow-sm border border-gray-100 hover:shadow-md transition-shadow group relative">
  
  <!-- Image -->
  <div class="w-full sm:w-32 h-32 shrink-0 rounded-xl overflow-hidden bg-gray-100">
    <img src={image} alt={title} class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500" />
  </div>

  <!-- Content -->
  <div class="flex flex-col flex-1 min-w-0 py-1">
    <h3 class="text-base font-extrabold text-[#a82b00] leading-tight mb-1 truncate">{title}</h3>
    <p class="text-xs text-gray-500 leading-snug line-clamp-2 mb-2 flex-1">{description}</p>
    
    <div class="flex items-end justify-between mt-auto">
      <div class="flex flex-col gap-1">
        <ReviewProp {rating} {count} />
        <span class="text-xl font-black text-[#ff5722] leading-none">{price}</span>
      </div>
      
      <button 
        onclick={handleAdd}
        class="bg-[#ff5722]/10 hover:bg-[#ff5722] text-[#ff5722] hover:text-white p-2 rounded-full flex items-center justify-center transition-all duration-300 shrink-0
        {isAdding ? 'bg-emerald-500 text-white scale-95' : ''}"
      >
        {#if isAdding}
          <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg>
        {:else}
          <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg>
        {/if}
      </button>
    </div>
  </div>
</div>
