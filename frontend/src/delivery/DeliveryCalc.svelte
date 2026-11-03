<script>
  export let restaurant = {
    address: "Av. Barreto de Menezes, 799",
    neighborhood: "Shopping Guararapes - Jaboatão",
    distance: "2.5 km",
    deliveryFee: 8.50,
    deliveryTime: "30-45 min"
  };

  let cepInput = "";
  let calculatingShipping = false;
  let deliveryCalculated = false;

  function formatCep(e) {
    let value = e.target.value.replace(/\D/g, "");
    if (value.length > 5) {
      value = value.replace(/^(\d{5})(\d)/, "$1-$2");
    }
    cepInput = value;
  }

  function handleCalculateShipping() {
    if (cepInput.replace(/\D/g, '').length < 8) return;
    
    calculatingShipping = true;
    
    // Simulate API request delay
    setTimeout(() => {
      calculatingShipping = false;
      deliveryCalculated = true;
    }, 1000);
  }
</script>

<div class="bg-white border border-gray-200 rounded-2xl p-5 shadow-sm">
  <h3 class="text-sm font-bold text-gray-900 mb-3">Logo na esquina</h3>
  
  <!-- Location Info -->
  <div class="mb-4">
    <div class="text-sm font-bold text-gray-900">{restaurant.address}</div>
    <div class="text-xs text-gray-500 mt-0.5">{restaurant.neighborhood} · {restaurant.distance} away</div>
  </div>

  <!-- Calcular Frete Section -->
  <div class="pt-4 border-t border-gray-100">
    <div class="text-xs font-semibold text-gray-800 mb-2 flex items-center justify-between">
      <span class="flex items-center gap-1.5">
        <svg class="w-4 h-4 text-[#ea4519]" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="18.5" cy="17.5" r="3.5"/>
          <circle cx="5.5" cy="17.5" r="3.5"/>
          <circle cx="15" cy="5" r="1"/>
          <path d="M12 17.5V14l-3-3 4-3 2 3h2"/>
        </svg>
        Calcular frete e prazo
      </span>
      {#if deliveryCalculated}
        <span class="text-[10px] text-emerald-700 bg-emerald-50 px-2 py-0.5 rounded-full font-semibold border border-emerald-200">
          Calculado
        </span>
      {/if}
    </div>

    <form on:submit|preventDefault={handleCalculateShipping} class="flex gap-2">
      <input
        type="text"
        bind:value={cepInput}
        on:input={formatCep}
        placeholder="00000-000 (somente CEP)"
        maxlength="9"
        class="flex-1 text-sm text-gray-900 px-3 py-2 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:border-[#ea4519] transition placeholder-gray-400 font-medium"
      />
      <button
        type="submit"
        disabled={calculatingShipping || cepInput.replace(/\D/g, '').length < 8}
        class="px-3.5 py-2 bg-[#ea4519] hover:bg-[#d63c12] text-white text-xs font-bold rounded-xl transition disabled:opacity-50 disabled:cursor-not-allowed whitespace-nowrap shadow-sm"
      >
        {calculatingShipping ? 'Calculando...' : 'Calcular'}
      </button>
    </form>

    {#if deliveryCalculated}
      <div class="mt-3 p-3 bg-gray-50 rounded-xl space-y-1.5 text-xs border border-gray-100">
        <div class="flex justify-between items-center text-gray-700">
          <span class="text-gray-500">Valor do Frete:</span>
          <span class="font-bold text-gray-900">R${restaurant.deliveryFee.toFixed(2).replace('.', ',')}</span>
        </div>
        <div class="flex justify-between items-center text-gray-700">
          <span class="text-gray-500">Prazo estimado:</span>
          <span class="font-bold text-gray-900">{restaurant.deliveryTime}</span>
        </div>
      </div>
    {/if}
  </div>
</div>
