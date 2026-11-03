<script>
  export let image;
  export let title;
  export let description;
  export let oldPrice;
  export let newPrice;
  export let isPopular = false;
</script>

<!-- Height set to 180px to fit the bigger prices comfortably -->
<div class="relative flex w-full min-w-[280px] h-[180px] bg-white rounded-2xl overflow-hidden shadow-sm border border-gray-100 snap-start group hover:shadow-md transition-shadow">
  
  <!-- Left Side: Image -->
  <div class="w-[42%] h-full relative shrink-0">
    <img src={image} alt={title} class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500" />
    
    <!-- Popular Badge -->
    {#if isPopular}
      <div class="absolute top-0 left-0 bg-[#ff5722] text-white text-[10px] font-bold px-2 py-1 rounded-br-lg flex items-center gap-1 shadow-md z-10">
        <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="currentColor" stroke="none"><path d="M8.5 14.5A2.5 2.5 0 0 0 11 12c0-1.38-.5-2-1-3-1.072-2.143-.224-4.054 2-6 .5 2.5 2 4.9 4 6.5 2 1.6 3 3.5 3 5.5a7 7 0 1 1-14 0c0-1.153.433-2.294 1-3a2.5 2.5 0 0 0 2.5 2.5z"/></svg>
        Mais pedido!
      </div>
    {/if}
  </div>

  <!-- Right Side: Content -->
  <div class="flex flex-col flex-1 p-3 pl-4 relative z-10 bg-gradient-to-r from-white via-white to-transparent">
    
    <!-- Reverted Title back to 15px -->
    <h3 class="text-[#a82b00] font-extrabold text-[15px] leading-tight mb-1 pr-1">{title}</h3>
    
    <!-- Reverted Description back to 11px -->
    <p class="text-gray-500 text-[11px] leading-snug line-clamp-2 mb-2 flex-1 pr-1">
      {description}
    </p>
    
    <!-- Pricing & Button -->
    <div class="flex items-end justify-between mt-auto pr-1">
      <div class="flex flex-col">
        {#if oldPrice}
          <!-- Made old price bigger (text-[15px]) -->
          <span class="text-gray-400 text-[15px] line-through font-medium leading-none mb-1">{oldPrice}</span>
        {/if}
        <!-- Made new price much bigger (text-2xl) -->
        <span class="text-[#ff5722] font-black text-2xl leading-none">{newPrice}</span>
      </div>
      
      <!-- Slightly bumped the + button size so it balances nicely with the big price -->
      <button class="bg-[#ff5722]/10 hover:bg-[#ff5722] text-[#ff5722] hover:text-white p-2 rounded-full flex items-center justify-center transition-all duration-300 group/btn shrink-0 mb-0.5" aria-label="Adicionar ao carrinho">
        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" class="group-hover/btn:rotate-90 transition-transform duration-300"><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg>
      </button>
    </div>
  </div>
</div>
