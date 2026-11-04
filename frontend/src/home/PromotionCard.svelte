<script>
  import { addToCart } from '../lib/stores/cart.js';
  import ReviewProp from '../review/ReviewProp.svelte';

  let {
    image,
    title,
    description,
    oldPrice,
    newPrice,
    isPopular = false,
    combo = null,
    rating = null,
    count = null,
    onOpenDetails = () => {} // 👈 New prop
  } = $props();

  let isPopping = $state(false);
  let justAdded = $state(false);

  function parsePrice(value) {
    if (typeof value === 'number') return value;
    if (!value) return 0;
    return Number(
      String(value)
        .replace(/[^\d,.-]/g, '')
        .replace(/\./g, '')
        .replace(',', '.')
    ) || 0;
  }

  function handleAdd(e) {
    // 👈 Stop click from opening the modal!
    e.stopPropagation(); 
    
    if (isPopping) return;
    isPopping = true;
    justAdded = true;

    addToCart({
      name: title,
      icon: image,
      price: parsePrice(newPrice),
      combo: combo,
    });

    setTimeout(() => { isPopping = false; }, 400); 
    setTimeout(() => { justAdded = false; }, 1200); 
  }
</script>

<!-- Card Container (Click opens modal) -->
<div 
  onclick={onOpenDetails}
  class="relative flex w-full min-w-[280px] h-[185px] bg-white rounded-2xl overflow-hidden shadow-sm border border-gray-100 snap-start group hover:shadow-md transition-shadow cursor-pointer
  {isPopping ? 'animate-card-pop' : ''}"
>
  
  <div class="w-[42%] h-full relative shrink-0">
    <img src={image} alt={title} class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500" />
    
    {#if isPopular}
      <div class="absolute top-0 left-0 bg-[#ff5722] text-white text-[10px] font-bold px-2 py-1 rounded-br-lg flex items-center gap-1 shadow-md z-10">
        <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="currentColor" stroke="none"><path d="M8.5 14.5A2.5 2.5 0 0 0 11 12c0-1.38-.5-2-1-3-1.072-2.143-.224-4.054 2-6 .5 2.5 2 4.9 4 6.5 2 1.6 3 3.5 3 5.5a7 7 0 1 1-14 0c0-1.153.433-2.294 1-3a2.5 2.5 0 0 0 2.5 2.5z"/></svg>
        Mais pedido!
      </div>
    {/if}
  </div>

  <div class="flex flex-col flex-1 p-3 pl-4 relative z-10 bg-gradient-to-r from-white via-white to-transparent">
    
    <h3 class="text-[#a82b00] font-extrabold text-[15px] leading-tight mb-1 pr-1">{title}</h3>
    
    <p class="text-gray-500 text-[11px] leading-snug line-clamp-2 mb-2 flex-1 pr-1">
      {description}
    </p>
    
    <div class="flex items-end justify-between mt-auto pr-1">
      
      <div class="flex flex-col gap-1">
        <ReviewProp {rating} {count} />
        
        <div class="flex flex-col">
          {#if oldPrice}
            <span class="text-gray-400 text-[11px] line-through font-medium leading-none mb-0.5">{oldPrice}</span>
          {/if}
          <span class="text-[#ff5722] font-black text-xl leading-none">{newPrice}</span>
        </div>
      </div>
      
      <!-- Add Button -->
      <button
        onclick={handleAdd}
        disabled={isPopping}
        class="p-2 rounded-full flex items-center justify-center transition-all duration-300 group/btn shrink-0 mb-0.5 min-w-[36px] min-h-[36px]
        {justAdded 
          ? 'bg-emerald-500 text-white shadow-[0_0_15px_rgba(16,185,129,0.5)]' 
          : 'bg-[#ff5722]/10 hover:bg-[#ff5722] text-[#ff5722] hover:text-white'}"
        aria-label="Adicionar ao carrinho"
      >
        {#if justAdded}
          <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg>
        {:else}
          <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" class="group-hover/btn:rotate-90 transition-transform duration-300"><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg>
        {/if}
      </button>
    </div>
  </div>
</div>

<style>
  @keyframes card-pop {
    0% { transform: scale(1); }
    40% { transform: scale(1.04) translateY(-2px); box-shadow: 0 10px 25px -5px rgba(255, 87, 34, 0.4); }
    100% { transform: scale(1) translateY(0); }
  }

  .animate-card-pop {
    animation: card-pop 0.4s cubic-bezier(0.34, 1.56, 0.64, 1);
    z-index: 10;
    border-color: #ff5722;
  }
</style>
