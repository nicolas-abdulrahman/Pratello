<script>
  // Svelte 5 Props
  let { 
    selectedId = 1, 
    onSelect = () => {} 
  } = $props();

  // The two specific locations requested
  const restaurants = [
    {
      id: 1,
      name: "Petrello Boa Viagem",
      address: "Rua Petros de Silva Barreto, Boa Viagem",
      time: "30-40 min",
    },
    {
      id: 2,
      name: "Petrello Candeias",
      address: "Rua Marciano, Candeias",
      time: "45-55 min",
    }
  ];
</script>

<div class="w-full mb-8">
  <h3 class="text-sm font-bold text-gray-400 uppercase tracking-wider mb-3">Escolha uma unidade</h3>
  
  <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
    {#each restaurants as rest}
      <button 
        onclick={() => onSelect(rest.id)}
        class="flex items-center text-left p-4 rounded-2xl border-2 transition-all duration-200 
        {selectedId === rest.id 
          ? 'border-[#ea4519] bg-orange-50/50 shadow-sm' 
          : 'border-gray-200 bg-white hover:border-orange-200 hover:bg-gray-50'}"
      >
        <!-- Icon -->
        <div class="w-10 h-10 rounded-full flex items-center justify-center shrink-0 mr-4 
          {selectedId === rest.id ? 'bg-[#ea4519] text-white' : 'bg-gray-100 text-gray-500'}"
        >
          <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 10c0 6-8 12-8 12s-8-6-8-12a8 8 0 0 1 16 0Z"/><circle cx="12" cy="10" r="3"/></svg>
        </div>
        
        <!-- Info -->
        <div class="flex-1">
          <h4 class="text-base font-bold text-gray-900">{rest.name}</h4>
          <p class="text-xs text-gray-500 mb-1">{rest.address}</p>
          <div class="flex items-center gap-3 text-xs font-semibold text-gray-700">
            <span class="flex items-center gap-1">
              <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
              {rest.time}
            </span>
            <span class="text-[#ea4519]">{rest.fee}</span>
          </div>
        </div>

        <!-- Checkmark (only if selected) -->
        {#if selectedId === rest.id}
          <div class="ml-4 text-[#ea4519]">
            <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M20 6 9 17l-5-5"/></svg>
          </div>
        {/if}
      </button>
    {/each}
  </div>
</div>
