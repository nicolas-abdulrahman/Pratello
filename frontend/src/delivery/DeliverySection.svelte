
<script>
  import DeliveryMap from './DeliveryMap.svelte';
  import DeliveryCalc from './DeliveryCalc.svelte';
  import DeliveryOrder from './DeliveryOrder.svelte';
  import DeliveryForm from './DeliveryForm.svelte';
  import Restaurants from './Restaurants.svelte'; // Imported the new component

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

<section class="min-h-screen bg-[#fff9f5] py-12 px-4 md:px-8">
  
  <div class="max-w-7xl mx-auto">
    
    <!-- Page Header -->
    <div class="mb-6">
      <h1 class="text-3xl font-black text-gray-900 mb-2">Área de Delivery</h1>
      <p class="text-gray-500">Selecione a unidade mais próxima e confira a área de cobertura.</p>
    </div>

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
