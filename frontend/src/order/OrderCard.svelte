<script>
  import IngredientItem from './IngredientItem.svelte';

  let { 
    item,
    removedState = $bindable({}),
    extraMayo = $bindable(false),
    isChained = $bindable(true),
    extraMeat1 = $bindable(false),
    extraMeat2 = $bindable(false),
    isDrinkChained = $bindable(true),
    selectedDrink1 = $bindable('d-coke'),
    selectedDrink2 = $bindable('d-coke')
  } = $props();

  // Ensure all keys exist before render
  function initKeys() {
    if (!item.isCombo) {
      item.options.forEach(opt => {
        if (removedState[opt.id] === undefined) removedState[opt.id] = false;
      });
    } else {
      item.sections[0].options.forEach(opt => {
        if (removedState[opt.id] === undefined) removedState[opt.id] = false;
        if (removedState[`${opt.id}-1`] === undefined) removedState[`${opt.id}-1`] = false;
        if (removedState[`${opt.id}-2`] === undefined) removedState[`${opt.id}-2`] = false;
      });
    }
  }
  
  initKeys();
  
  $effect(() => {
    initKeys();
  });

  // Calculate discount if originalPrice is provided
  let discount = $derived.by(() => {
    if (!item.originalPrice) return null;
    const diff = item.originalPrice - item.price;
    const percent = Math.round((diff / item.originalPrice) * 100);
    return { diff, percent };
  });
</script>

<div class="bg-white rounded-3xl p-6 shadow-sm border border-gray-100 overflow-hidden relative">
  
  <div class="flex gap-4 items-center mb-6 border-b border-gray-50 pb-4">
    <div class="w-16 h-16 rounded-xl overflow-hidden shrink-0 bg-gray-50">
      <img src={item.image} alt={item.title} class="w-full h-full object-cover" />
    </div>
    <div>
      <h2 class="text-xl font-bold text-gray-900">{item.title}</h2>
      <div class="flex items-center gap-2 mt-1">
        <!-- Sale Price Logic -->
        {#if item.originalPrice}
          <span class="text-sm text-gray-400 line-through font-medium">R$ {item.originalPrice.toFixed(2).replace('.', ',')}</span>
          <span class="text-xl font-black text-[#ff5722]">R$ {item.price.toFixed(2).replace('.', ',')}</span>
          <span class="bg-green-100 text-green-700 text-[10px] font-bold px-1.5 py-0.5 rounded uppercase tracking-wider">
            {discount.percent}% OFF
          </span>
        {:else}
          <span class="text-xl font-black text-[#ff5722]">R$ {item.price.toFixed(2).replace('.', ',')}</span>
        {/if}
      </div>
    </div>
  </div>

  {#if !item.isCombo}
    <div class="bg-orange-50/50 p-4 rounded-2xl border border-orange-100">
      <h3 class="text-sm font-bold text-gray-800 mb-2">Remover Ingredientes</h3>
      <div class="flex flex-wrap gap-2">
        {#each item.options as ing}
          <IngredientItem 
            label={ing.name} 
            bind:isRemoved={removedState[ing.id]} 
          />
        {/each}
      </div>
    </div>
  {/if}

  {#if item.isCombo}
    <div class="space-y-4">
      
      <!-- BURGERS -->
      <div class="pl-4 border-l-4 border-blue-200 bg-blue-50/30 p-4 rounded-r-2xl">
        <div class="flex items-center justify-between mb-3">
          <h3 class="text-base font-bold text-gray-800">
            {isChained ? '2x Hambúrgueres Ciri Cascudo' : 'Hambúrgueres Ciri Cascudo'}
          </h3>
          <label class="flex items-center gap-2 cursor-pointer text-xs font-bold text-blue-600 bg-white px-2 py-1 rounded-lg border border-blue-100 shadow-sm hover:border-blue-300 transition-colors">
            <input type="checkbox" bind:checked={isChained} class="w-3.5 h-3.5 text-blue-600 rounded focus:ring-0 cursor-pointer" />
            
            {#if isChained}
              <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/></svg>
            {:else}
              <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/><line x1="2" y1="2" x2="22" y2="22" stroke-width="3"/></svg>
            {/if}
            
            Iguais
          </label>
        </div>

        {#if isChained}
          <div class="space-y-3">
            <div class="flex flex-wrap gap-2">
              {#each item.sections[0].options as ing}
                <IngredientItem 
                  label={ing.name} 
                  bind:isRemoved={removedState[ing.id]} 
                />
              {/each}
            </div>
            <label class="flex items-center justify-between p-3 rounded-xl border border-blue-100 bg-white cursor-pointer hover:border-blue-300 mt-2">
              <div class="flex items-center gap-3">
                <input type="checkbox" bind:checked={extraMeat1} class="w-4 h-4 text-blue-600 rounded focus:ring-blue-500" />
                <span class="font-semibold text-gray-700 text-sm">Carne Extra (para os dois)</span>
              </div>
              <span class="font-bold text-blue-600 text-sm">+ R$ 4,00</span>
            </label>
          </div>
        {:else}
          <div class="space-y-3">
            <div class="bg-white p-3 rounded-xl border border-gray-100">
              <h4 class="text-[10px] font-bold text-gray-400 uppercase mb-2">Hambúrguer 1</h4>
              <div class="flex flex-wrap gap-2 mb-3">
                {#each item.sections[0].options as ing}
                  <IngredientItem 
                    label={ing.name} 
                    bind:isRemoved={removedState[`${ing.id}-1`]} 
                  />
                {/each}
              </div>
              <label class="flex items-center justify-between cursor-pointer">
                <div class="flex items-center gap-3">
                  <input type="checkbox" bind:checked={extraMeat1} class="w-4 h-4 text-blue-600 rounded focus:ring-blue-500" />
                  <span class="font-medium text-gray-700 text-xs">Carne Extra</span>
                </div>
                <span class="font-bold text-blue-600 text-xs">+ R$ 2,00</span>
              </label>
            </div>

            <div class="bg-white p-3 rounded-xl border border-gray-100">
              <h4 class="text-[10px] font-bold text-gray-400 uppercase mb-2">Hambúrguer 2</h4>
              <div class="flex flex-wrap gap-2 mb-3">
                {#each item.sections[0].options as ing}
                  <IngredientItem 
                    label={ing.name} 
                    bind:isRemoved={removedState[`${ing.id}-2`]} 
                  />
                {/each}
              </div>
              <label class="flex items-center justify-between cursor-pointer">
                <div class="flex items-center gap-3">
                  <input type="checkbox" bind:checked={extraMeat2} class="w-4 h-4 text-blue-600 rounded focus:ring-blue-500" />
                  <span class="font-medium text-gray-700 text-xs">Carne Extra</span>
                </div>
                <span class="font-bold text-blue-600 text-xs">+ R$ 2,00</span>
              </label>
            </div>
          </div>
        {/if}
      </div>

      <!-- FRIES -->
      <div class="pl-4 border-l-4 border-yellow-200 bg-yellow-50/30 p-4 rounded-r-2xl">
        <h3 class="text-base font-bold text-gray-800 mb-3">Batata Grande</h3>
        <div class="space-y-2">
          {#each item.sections[1].addOns as addon}
            <label class="flex items-center justify-between p-3 rounded-xl border border-yellow-100 bg-white cursor-pointer hover:border-yellow-300">
              <div class="flex items-center gap-3">
                <input type="checkbox" bind:checked={extraMayo} class="w-4 h-4 text-yellow-600 rounded focus:ring-yellow-500" />
                <span class="font-semibold text-gray-700 text-sm">{addon.name}</span>
              </div>
              <span class="font-bold text-yellow-600 text-sm">+ R$ {addon.price.toFixed(2).replace('.', ',')}</span>
            </label>
          {/each}
        </div>
      </div>

      <!-- DRINKS -->
      <div class="pl-4 border-l-4 border-purple-200 bg-purple-50/30 p-4 rounded-r-2xl">
        <div class="flex items-center justify-between mb-3">
          <h3 class="text-base font-bold text-gray-800">
            {isDrinkChained ? '2x Bebidas' : 'Bebidas'}
          </h3>
          <label class="flex items-center gap-2 cursor-pointer text-xs font-bold text-purple-600 bg-white px-2 py-1 rounded-lg border border-purple-100 shadow-sm hover:border-purple-300 transition-colors">
            <input type="checkbox" bind:checked={isDrinkChained} class="w-3.5 h-3.5 text-purple-600 rounded focus:ring-0 cursor-pointer" />
            
            {#if isDrinkChained}
              <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/></svg>
            {:else}
              <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/><line x1="2" y1="2" x2="22" y2="22" stroke-width="3"/></svg>
            {/if}
            
            Iguais
          </label>
        </div>

        {#if isDrinkChained}
          <div class="flex flex-col gap-2">
            {#each item.sections[2].choices as choice}
              <button
                onclick={() => selectedDrink1 = choice.id}
                class="w-full text-left px-4 py-3 rounded-xl text-sm font-bold border transition-all flex items-center justify-between
                {selectedDrink1 === choice.id 
                  ? 'bg-purple-600 text-white border-purple-600 shadow-md' 
                  : 'bg-white text-gray-600 border-gray-200 hover:border-purple-200'}"
              >
                {choice.name}
              </button>
            {/each}
          </div>
        {:else}
          <div class="space-y-4">
            <div class="bg-white p-3 rounded-xl border border-gray-100">
              <h4 class="text-[10px] font-bold text-gray-400 uppercase mb-2">Bebida 1</h4>
              <div class="flex flex-col gap-2">
                {#each item.sections[2].choices as choice}
                  <button
                    onclick={() => selectedDrink1 = choice.id}
                    class="w-full text-left px-3 py-2 rounded-lg text-xs font-bold border transition-all
                    {selectedDrink1 === choice.id ? 'bg-purple-100 text-purple-700 border-purple-200' : 'bg-gray-50 text-gray-600 border-gray-100'}"
                  >
                    {choice.name}
                  </button>
                {/each}
              </div>
            </div>
            <div class="bg-white p-3 rounded-xl border border-gray-100">
              <h4 class="text-[10px] font-bold text-gray-400 uppercase mb-2">Bebida 2</h4>
              <div class="flex flex-col gap-2">
                {#each item.sections[2].choices as choice}
                  <button
                    onclick={() => selectedDrink2 = choice.id}
                    class="w-full text-left px-3 py-2 rounded-lg text-xs font-bold border transition-all
                    {selectedDrink2 === choice.id ? 'bg-purple-100 text-purple-700 border-purple-200' : 'bg-gray-50 text-gray-600 border-gray-100'}"
                  >
                    {choice.name}
                  </button>
                {/each}
              </div>
            </div>
          </div>
        {/if}
      </div>

    </div>
  {/if}
</div>
