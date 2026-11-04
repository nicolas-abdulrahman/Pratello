<script>
  let { item, onRemove = () => {} } = $props();

  // Safely extract the sub-items no matter how they are formatted in the cart
  let subItems = $derived(
    item.combo?.items || 
    item.items || 
    (Array.isArray(item.combo) ? item.combo : [])
  );
</script>

<div class="relative flex gap-4 p-3 bg-white border border-gray-200 rounded-2xl shadow-sm hover:shadow-md transition-shadow">
  
  <!-- Remove Button -->
  <button 
    onclick={() => onRemove(item.id)}
    class="absolute -top-2.5 -right-2.5 w-6 h-6 bg-gray-100 hover:bg-red-100 text-gray-500 hover:text-red-500 rounded-full flex items-center justify-center transition-colors shadow-sm z-10"
    aria-label="Remover item"
  >
    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M18 6 6 18"/><path d="m6 6 12 12"/></svg>
  </button>

  <!-- Image -->
  <div class="w-20 h-20 shrink-0 rounded-xl overflow-hidden bg-gray-100 relative">
    <img src={item.icon || item.image} alt={item.name || item.title} class="w-full h-full object-cover" />
    
    <!-- Quantity Badge -->
    {#if item.quantity > 1}
      <div class="absolute bottom-0 right-0 bg-[#ff5722] text-white text-[10px] font-bold px-1.5 py-0.5 rounded-tl-md">
        x{item.quantity}
      </div>
    {/if}
  </div>

  <!-- Content -->
  <div class="flex flex-col flex-1 min-w-0 py-1">
    <div class="flex justify-between items-start mb-1 gap-2">
      <div>
        <h3 class="font-bold text-gray-900 text-base leading-tight truncate">{item.name || item.title}</h3>
        
        <!-- Optional: Label showing it's a combo -->
        {#if item.combo}
          <span class="inline-block bg-orange-100 text-[#d94e17] text-[9px] font-bold px-1.5 py-0.5 rounded mt-1 uppercase tracking-wider">
            Combo
          </span>
        {/if}
      </div>
      
      <span class="font-bold text-[#ff5722] text-sm whitespace-nowrap">
        R$ {(item.price * (item.quantity || 1)).toFixed(2).replace('.', ',')}
      </span>
    </div>
    
    <!-- Sub-items (Combo details) -->
    {#if subItems && subItems.length > 0}
      <ul class="space-y-0.5 mt-1.5">
        {#each subItems as subItem}
          <li class="text-xs text-gray-500 flex items-center gap-1.5">
            <span class="w-1 h-1 rounded-full bg-gray-300 shrink-0"></span>
            <span class="truncate">{subItem}</span>
          </li>
        {/each}
      </ul>
    {/if}
  </div>
</div>
