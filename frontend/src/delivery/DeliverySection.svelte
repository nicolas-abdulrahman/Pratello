
<script>
  import DeliveryMap from './DeliveryMap.svelte';
  import DeliveryCalc from './DeliveryCalc.svelte';
  import DeliveryOrder from './DeliveryOrder.svelte';
  import DeliveryForm from './DeliveryForm.svelte';
  import Restaurants from '../common/Restaurants.svelte'; // Imported the new component

  let { slug = '', onBack = () => {} } = $props();

  // State to toggle the checkout form under the map
  let showCheckoutForm = $state(false);
  
  // State for the selected restaurant (1 = Boa Viagem, 2 = Candeias)
  let selectedRestaurantId = $state(1);

  function handleCheckout() {
    showCheckoutForm = true;
    // Scroll down to show the form smoothly
    setTimeout(() => {
      window.scrollBy({ top: 400, behavior: 'smooth' });
    }, 100);
  }

  function handleFormSubmit(data) {
    console.log("Order placed with address:", data);
    alert("Pedido finalizado com sucesso!");
    showCheckoutForm = false;
  }
</script>

<section class="min-h-screen bg-background text-text py-12 px-4 md:px-8">
  
  <div class="max-w-7xl mx-auto">
    
    <!-- Page Header -->
    <header class="relative mb-8 flex min-h-12 items-center justify-center">
      <!-- Back button -->
      <button
        type="button"
        aria-label="Voltar"
        onclick={onBack}
        class="absolute left-0 flex h-11 w-11 items-center justify-center rounded-full border border-border bg-surface text-text shadow-sm transition-all hover:bg-surface-hover active:scale-95 cursor-pointer"
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="24"
          height="24"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
          aria-hidden="true"
        >
          <path d="m12 19-7-7 7-7" />
          <path d="M19 12H5" />
        </svg>
      </button>

      <div class="text-center px-12">
        <h1 class="text-2xl sm:text-3xl font-black text-text mb-1">Área de Delivery</h1>
        <p class="text-sm text-text-muted">Selecione a unidade mais próxima e confira a área de cobertura.</p>
      </div>
    </header>

    <!-- NEW: Restaurant Selection -->
    <Restaurants 
      selectedId={selectedRestaurantId} 
      onSelect={(id) => selectedRestaurantId = id} 
    />

    <!-- Top Grid (Map, Calc, Order) -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      
      <!-- Left Side: Map -->
      <div class="lg:col-span-2">
        <DeliveryMap />
      </div>

      <!-- Right Side: Calculator & Order Summary -->
      <div class="lg:col-span-1 space-y-6">
        <DeliveryCalc />
        <DeliveryOrder onCheckout={handleCheckout} />
      </div>

    </div>

    <!-- 
      BOTTOM SECTION: The Form 
      Placed outside the grid so it spans under everything with padding 
    -->
    {#if showCheckoutForm}
      <div class="mt-8 pt-6 border-t border-gray-200/50">
        <div class="max-w-4xl mx-auto">
          <DeliveryForm onSubmit={handleFormSubmit} />
        </div>
      </div>
    {/if}

  </div>

</section>
