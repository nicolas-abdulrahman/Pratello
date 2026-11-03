<script>
  import MenuCategoryCard from './MenuCategoryCard.svelte';

  // ✅ YOU WERE MISSING THIS ARRAY
  const categories = [
    {
      title: "Hambúrgueres",
      description: "O hambúrguer perfeito para você!",
      icon: `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#f05a1f" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 11h18a2 2 0 0 1 0 4H3a2 2 0 0 1 0-4Z"/><path d="M5 11V7a3 3 0 0 1 3-3h8a3 3 0 0 1 3 3v4"/><path d="M3 15h18v3a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-3Z"/></svg>`,
      image: "https://images.unsplash.com/photo-1568901346375-23c9450c58cd?ixlib=rb-4.0.3&auto=format&fit=crop&w=800&q=80"
    },
    {
      title: "Pizzas",
      description: "Muito mais sabor em cada fatia!",
      icon: `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#f05a1f" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M15 14c.2-1 .7-1.7 1.5-2.5 1-.9 1.5-2.2 1.5-3.5A6 6 0 0 0 6 8c0 1 .2 2.2 1.5 3.5.7.9 1.3 1.5 1.5 2.5"/><path d="M9 18h6"/><path d="M10 22h4"/></svg>`,
      image: "https://images.unsplash.com/photo-1513104890138-7c749659a591?ixlib=rb-4.0.3&auto=format&fit=crop&w=800&q=80"
    },
    {
      title: "Porções",
      description: "Para compartilhar ou se deliciar sozinho!",
      icon: `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#f05a1f" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 9h18v10a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V9Z"/><path d="m3 9 2.45-4.9A2 2 0 0 1 7.24 3h9.52a2 2 0 0 1 1.8 1.1L21 9"/><path d="M12 3v6"/></svg>`,
      image: "https://images.unsplash.com/photo-1573080496219-bb080dd4f877?ixlib=rb-4.0.3&auto=format&fit=crop&w=800&q=80"
    },
    {
      title: "Bebidas",
      description: "Refresque seus momentos!",
      icon: `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#f05a1f" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M6 8h12l-2 14H8L6 8Z"/><path d="M6 8V6a2 2 0 0 1 2-2h8a2 2 0 0 1 2 2v2"/></svg>`,
      image: "https://images.unsplash.com/photo-1513558161293-cdaf765ed2fd?ixlib=rb-4.0.3&auto=format&fit=crop&w=800&q=80"
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

      <!-- Grid -->
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {#each categories as cat}
          <MenuCategoryCard 
            title={cat.title} 
            description={cat.description} 
            image={cat.image} 
            icon={cat.icon} 
          />
        {/each}
      </div>
    </section>
</div>
