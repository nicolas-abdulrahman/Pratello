<script>
  import MenuCategoryCard from './MenuCategoryCard.svelte';
  import MenuDishCard from './MenuDishCard.svelte';

  // State to track which category is open
  let openCategory = $state(null);

  // Toggle function
  function toggleCategory(title) {
    if (openCategory === title) {
      openCategory = null;
    } else {
      openCategory = title;
    }
  }

  const categories = [
    {
      title: "Hambúrgueres",
      description: "O hambúrguer perfeito para você!",
      icon: `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#f05a1f" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 11h18a2 2 0 0 1 0 4H3a2 2 0 0 1 0-4Z"/><path d="M5 11V7a3 3 0 0 1 3-3h8a3 3 0 0 1 3 3v4"/><path d="M3 15h18v3a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-3Z"/></svg>`,
      image: "https://images.unsplash.com/photo-1568901346375-23c9450c58cd?ixlib=rb-4.0.3&auto=format&fit=crop&w=800&q=80",
      dishes: [
        { id: 1, title: "Cheese Burger", description: "Pão brioche, carne 180g, queijo cheddar e molho especial.", price: "R$ 24,90", image: "https://images.unsplash.com/photo-1568901346375-23c9450c58cd?auto=format&fit=crop&w=300&h=300", rating: 4.8, count: "1k+" },
        { id: 2, title: "Bacon Burger", description: "Carne 180g, muito bacon crocante e queijo prato.", price: "R$ 29,90", image: "https://images.unsplash.com/photo-1594212699903-ec8a3eca50f5?auto=format&fit=crop&w=300&h=300", rating: 4.9, count: "500+" },
        { id: 3, title: "Veggie Burger", description: "Hambúrguer de grão de bico, alface e tomate.", price: "R$ 22,90", image: "https://images.unsplash.com/photo-1520072959219-c595dc870360?auto=format&fit=crop&w=300&h=300" }
      ]
    },
    {
      title: "Pizzas",
      description: "Muito mais sabor em cada fatia!",
      icon: `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#f05a1f" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M15 14c.2-1 .7-1.7 1.5-2.5 1-.9 1.5-2.2 1.5-3.5A6 6 0 0 0 6 8c0 1 .2 2.2 1.5 3.5.7.9 1.3 1.5 1.5 2.5"/><path d="M9 18h6"/><path d="M10 22h4"/></svg>`,
      image: "https://images.unsplash.com/photo-1513104890138-7c749659a591?ixlib=rb-4.0.3&auto=format&fit=crop&w=800&q=80",
      dishes: [
        { id: 4, title: "Pizza Margherita", description: "Molho de tomate, muçarela e manjericão.", price: "R$ 45,00", image: "https://images.unsplash.com/photo-1574071318508-1cdbab80d002?auto=format&fit=crop&w=300&h=300", rating: 4.7, count: "800+" },
        { id: 5, title: "Pizza Pepperoni", description: "Molho de tomate, muçarela e pepperoni.", price: "R$ 49,90", image: "https://images.unsplash.com/photo-1628840042765-356cda07504e?auto=format&fit=crop&w=300&h=300" }
      ]
    },
    {
      title: "Porções",
      description: "Para compartilhar ou se deliciar sozinho!",
      icon: `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#f05a1f" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 9h18v10a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V9Z"/><path d="m3 9 2.45-4.9A2 2 0 0 1 7.24 3h9.52a2 2 0 0 1 1.8 1.1L21 9"/><path d="M12 3v6"/></svg>`,
      image: "https://images.unsplash.com/photo-1573080496219-bb080dd4f877?ixlib=rb-4.0.3&auto=format&fit=crop&w=800&q=80",
      dishes: [
        { id: 6, title: "Batata Frita", description: "Porção grande de batata crocante.", price: "R$ 18,90", image: "https://images.unsplash.com/photo-1573080496219-bb080dd4f877?auto=format&fit=crop&w=300&h=300", rating: 4.5, count: "2k+" },
        { id: 7, title: "Onion Rings", description: "Anéis de cebola empanados.", price: "R$ 22,00", image: "https://images.unsplash.com/photo-1639024471283-03518883512d?auto=format&fit=crop&w=300&h=300" }
      ]
    },
    {
      title: "Bebidas",
      description: "Refresque seus momentos!",
      icon: `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#f05a1f" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M6 8h12l-2 14H8L6 8Z"/><path d="M6 8V6a2 2 0 0 1 2-2h8a2 2 0 0 1 2 2v2"/></svg>`,
      image: "https://images.unsplash.com/photo-1513558161293-cdaf765ed2fd?ixlib=rb-4.0.3&auto=format&fit=crop&w=800&q=80",
      dishes: [
        { id: 8, title: "Coca-Cola Lata", description: "350ml bem gelada.", price: "R$ 6,00", image: "https://images.unsplash.com/photo-1622483767028-3f66f32aef97?auto=format&fit=crop&w=300&h=300" }
      ]
    }
  ];
</script>

<div class="min-h-screen bg-[#fff9f5] font-sans selection:bg-[#f05a1f] selection:text-white pt-10">
  <section class="py-16 px-8 max-w-7xl mx-auto">
      
      <!-- Section Title -->
      <div class="flex items-center justify-center gap-4 mb-12">
        <div class="h-px bg-[#f05a1f]/30 flex-grow"></div>
        <div class="flex items-center gap-3">
          <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="#f05a1f" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 2v7c0 1.1.9 2 2 2h4a2 2 0 0 0 2-2V2"/><path d="M7 2v20"/><path d="M21 15V2v0a5 5 0 0 0-5 5v6c0 1.1.9 2 2 2h3Zm0 0v7"/></svg>
          <h2 class="text-3xl md:text-4xl font-black text-gray-900"><span class="text-[#f05a1f]">Cardápio</span></h2>
        </div>
        <div class="h-px bg-[#f05a1f]/30 flex-grow"></div>
      </div>

      <!-- Categories Grid -->
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        {#each categories as cat}
          <div 
            onclick={() => toggleCategory(cat.title)}
            class="cursor-pointer transition-transform duration-300 {openCategory === cat.title ? 'scale-95 opacity-90 ring-2 ring-[#f05a1f]/20 rounded-3xl' : 'hover:-translate-y-1'}"
          >
            <MenuCategoryCard 
              title={cat.title} 
              description={cat.description} 
              image={cat.image} 
              icon={cat.icon} 
            />
          </div>
        {/each}
      </div>

      <!-- FULL WIDTH EXPANDED SECTION (Below the grid) -->
      {#if openCategory}
        {@const selectedCat = categories.find(c => c.title === openCategory)}
        
        <div class="animate-expand bg-white/50 backdrop-blur-sm border border-orange-100 rounded-3xl p-6 md:p-8 shadow-inner mt-8">
          
          <div class="flex items-center justify-between mb-6 border-b border-gray-200/60 pb-4">
            <h3 class="text-2xl font-black text-gray-900">
              {openCategory}
            </h3>
            <button 
              onclick={() => openCategory = null}
              class="text-sm font-bold text-gray-500 hover:text-[#f05a1f] flex items-center gap-1 transition-colors"
            >
              Fechar
              <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 6 6 18"/><path d="m6 6 12 12"/></svg>
            </button>
          </div>

          <!-- Dishes Grid: 1 col on mobile, 2 on md, 3 on lg -->
          <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {#each selectedCat.dishes as dish}
              <MenuDishCard 
                title={dish.title}
                description={dish.description}
                image={dish.image}
                price={dish.price}
                rating={dish.rating}
                count={dish.count}
              />
            {/each}
          </div>
          
        </div>
      {/if}

    </section>
</div>

<style>
  .animate-expand {
    animation: expand 0.4s cubic-bezier(0.16, 1, 0.3, 1) forwards;
    transform-origin: top;
  }
  
  @keyframes expand {
    0% { opacity: 0; transform: scaleY(0.95) translateY(-20px); }
    100% { opacity: 1; transform: scaleY(1) translateY(0); }
  }
</style>
