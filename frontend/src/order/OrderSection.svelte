<script>
  import Restaurants from '../common/Restaurants.svelte';
  import OrderCard from './OrderCard.svelte';

  let { goToCheckout = () => {}, onCancel = () => {} } = $props();

  let selectedRestaurant = $state(1);
  
  let orderItems = $state([
    {
      id: 'item-1',
      title: "Hambúrguer Duplo Especial",
      price: 22.90,
      image: "https://images.unsplash.com/photo-1568901346375-23c9450c58cd?auto=format&fit=crop&q=80&w=200&h=200",
      isCombo: false,
      quantity: 1,
      options: [
        { id: 'pao', name: 'Pão Brioche', type: 'remove' },
        { id: 'carne', name: 'Carne Bovina 180g', type: 'remove' },
        { id: 'queijo', name: 'Queijo Cheddar', type: 'remove' },
        { id: 'alface', name: 'Alface Americana', type: 'remove' },
        { id: 'tomate', name: 'Tomate', type: 'remove' },
        { id: 'bacon', name: 'Bacon', type: 'remove' },
        { id: 'molho', name: 'Molho Especial', type: 'remove' }
      ]
    },
    {
      id: 'item-2',
      title: "Combo Casal",
      originalPrice: 70.00,
      price: 55.90,
      image: "https://images.unsplash.com/photo-1594212699903-ec8a3eca50f5?auto=format&fit=crop&q=80&w=200&h=200",
      isCombo: true,
      quantity: 1,
      sections: [
        { 
          id: 'sec-burger', 
          name: 'Hambúrgueres',
          options: [{ id: 'b-pao', name: 'Pão Brioche' }, { id: 'b-carne', name: 'Carne Bovina' }] 
        },
        { 
          id: 'sec-fries', 
          name: 'Batata Grande',
          addOns: [{ id: 'f-mayo', name: 'Maionese Gourmet', price: 2.00 }] 
        },
        { 
          id: 'sec-drink', 
          name: 'Bebidas',
          choices: [{ id: 'd-coke', name: 'Coca-Cola 350ml' }, { id: 'd-pepsi', name: 'Pepsi 350ml' }] 
        }
      ]
    }
  ]);

  let removedState1 = $state({});
  let removedState2 = $state({});

  let comboOptions = $state({
    extraMayo: false,
    isChained: true,
    extraMeat1: false,
    extraMeat2: false,
    isDrinkChained: true,
    selectedDrink1: 'd-coke',
    selectedDrink2: 'd-coke'
  });

  let finalTotal = $derived.by(() => {
    let total = 0;
    for (let item of orderItems) {
      total += item.price * item.quantity;
    }

    if (comboOptions.extraMayo) total += 2.00;
    
    if (comboOptions.isChained) {
      if (comboOptions.extraMeat1) total += 4.00;
    } else {
      if (comboOptions.extraMeat1) total += 2.00;
      if (comboOptions.extraMeat2) total += 2.00;
    }

    return total;
  });

  // Calculate the original total without sale discounts
  let originalTotal = $derived.by(() => {
    let total = 0;
    for (let item of orderItems) {
      total += (item.originalPrice || item.price) * item.quantity;
    }
    // Add the same extras to the original total
    if (comboOptions.extraMayo) total += 2.00;
    if (comboOptions.isChained) {
      if (comboOptions.extraMeat1) total += 4.00;
    } else {
      if (comboOptions.extraMeat1) total += 2.00;
      if (comboOptions.extraMeat2) total += 2.00;
    }
    return total;
  });

  let totalSaved = $derived(originalTotal - finalTotal);
</script>

<div class="min-h-screen bg-[#fff9f5] font-sans selection:bg-[#f05a1f] selection:text-white pb-12 pt-6 px-4 md:px-8">
  <div class="max-w-3xl mx-auto space-y-6">
    
    <div class="flex items-center gap-4 mb-4">
      <button onclick={onCancel} class="p-2.5 bg-white border border-gray-200 rounded-full shadow-sm hover:text-[#ff5722]">
        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="m15 18-6-6 6-6"/></svg>
      </button>
      <h1 class="text-2xl font-black text-gray-900">Seu Pedido</h1>
    </div>

    <Restaurants selectedId={selectedRestaurant} onSelect={(id) => selectedRestaurant = id} />

    <div class="space-y-4">
      {#each orderItems as item}
        {#if item.isCombo}
          <OrderCard 
            {item} 
            bind:removedState={removedState2}
            bind:extraMayo={comboOptions.extraMayo}
            bind:isChained={comboOptions.isChained}
            bind:extraMeat1={comboOptions.extraMeat1}
            bind:extraMeat2={comboOptions.extraMeat2}
            bind:isDrinkChained={comboOptions.isDrinkChained}
            bind:selectedDrink1={comboOptions.selectedDrink1}
            bind:selectedDrink2={comboOptions.selectedDrink2}
          />
        {:else}
          <OrderCard 
            {item} 
            bind:removedState={removedState1}
          />
        {/if}
      {/each}
    </div>

    <div class="bg-white rounded-3xl p-4 md:p-6 shadow-xl shadow-gray-200/50 border border-gray-100 flex flex-col md:flex-row items-center justify-between gap-4 sticky bottom-4 z-30">
      
      <!-- Price Breakdown -->
      <div class="flex flex-col flex-shrink-0 w-full md:w-[40%]">
        {#if totalSaved > 0}
          <span class="text-[10px] text-gray-400 line-through">Total original: R$ {originalTotal.toFixed(2).replace('.', ',')}</span>
          <span class="text-xs text-emerald-600 font-bold mb-1">Você economizou R$ {totalSaved.toFixed(2).replace('.', ',')}!</span>
        {/if}
        <div class="flex items-end gap-2">
          <span class="text-2xl md:text-3xl font-black text-[#ff5722] leading-none">
            R$ {finalTotal.toFixed(2).replace('.', ',')}
          </span>
          <span class="text-[10px] text-gray-400 font-medium mb-1">no total</span>
        </div>
      </div>
      
      <!-- Checkout Button -->
      <button 
        onclick={goToCheckout}
        class="w-full md:flex-1 bg-[#ff5722] hover:bg-[#e64a19] text-white font-bold py-3.5 md:py-4 px-6 rounded-full shadow-lg hover:shadow-orange-500/30 transition-all flex items-center justify-center gap-2 text-sm md:text-base"
      >
        Ir para o Pagamento
        <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12h14"/><path d="m12 5 7 7-7 7"/></svg>
      </button>
    </div>

  </div>
</div>
