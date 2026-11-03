<script>
  export let image;
  export let title;
  export let desc;
  export let price;
  export let originalPrice;
</script>

<div class="bg-white rounded-2xl overflow-hidden shadow-sm border border-gray-100 hover:shadow-xl transition-all duration-300 group flex flex-col w-[300px] min-w-[300px] snap-start h-full">
  
  <!-- Image Container -->
  <div class="relative h-48 overflow-hidden shrink-0">
    <img 
      src={image} 
      alt={title} 
      class="w-full h-full object-cover group-hover:scale-110 transition-transform duration-500"
    />
  </div>

  <!-- Content -->
  <div class="p-5 flex flex-col flex-1">
    <h3 class="text-lg font-extrabold text-gray-900 mb-1">{title}</h3>
    <p class="text-sm text-gray-500 leading-snug mb-4 flex-1 line-clamp-2">{desc}</p>
    
    <div class="flex items-end justify-between mt-auto pt-4 border-t border-gray-50">
      <div class="flex flex-col">
        {#if originalPrice}
          <span class="text-xs text-gray-400 line-through font-medium leading-none mb-1">{originalPrice}</span>
        {/if}
        <span class="text-xl font-black text-[#ff5722] leading-none">{price}</span>
      </div>
      
      <button class="bg-[#ff5722] hover:bg-[#e64a19] text-white font-bold py-2 px-4 rounded-full text-sm transition-colors shadow-md flex items-center gap-2">
        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="8" cy="21" r="1"/><circle cx="19" cy="21" r="1"/><path d="M2.05 2.05h2l2.66 12.42a2 2 0 0 0 2 1.58h9.78a2 2 0 0 0 1.95-1.57l1.65-7.43H5.12"/></svg>
        Adicionar
      </button>
    </div>
  </div>
  
</div>
