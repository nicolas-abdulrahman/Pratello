<script>
  import MelhoresPratosCard from './MelhoresPratosCard.svelte';

  const pratos = [
    {
      id: 1,
      title: "Hambúrguer Duplo Especial",
      description: "Dois hambúrgueres, queijo derretido, bacon crocante e muito sabor!",
      oldPrice: "R$ 34,90",
      newPrice: "R$ 22,90",
      isPopular: true,
      image: "https://images.unsplash.com/photo-1568901346375-23c9450c58cd?auto=format&fit=crop&q=80&w=300&h=300"
    },
    {
      id: 2,
      title: "Pizza Pepperoni",
      description: "Massa artesanal, molho de tomate e pepperoni selecionado.",
      oldPrice: "R$ 39,90",
      newPrice: "R$ 29,90",
      isPopular: false,
      image: "https://images.unsplash.com/photo-1628840042765-356cda07504e?auto=format&fit=crop&q=80&w=300&h=300"
    },
    {
      id: 3,
      title: "Porção de Batata Frita",
      description: "Batata crocante e sequinha, perfeita para compartilhar.",
      oldPrice: "R$ 19,90",
      newPrice: "R$ 12,90",
      isPopular: false,
      image: "https://images.unsplash.com/photo-1573080496219-bb080dd4f877?auto=format&fit=crop&q=80&w=300&h=300"
    },
    {
      id: 4,
      title: "Sanduíche de Frango",
      description: "Frango grelhado, alface e molho especial da casa.",
      oldPrice: "R$ 25,90",
      newPrice: "R$ 18,90",
      isPopular: false,
      image: "https://images.unsplash.com/photo-1521390188846-e2a3a97453a0?auto=format&fit=crop&q=80&w=300&h=300"
    },
    {
      id: 5,
      title: "Onion Rings Crocantes",
      description: "Anéis de cebola empanados na hora, super crocantes e saborosos.",
      oldPrice: "R$ 22,00",
      newPrice: "R$ 15,90",
      isPopular: false,
      image: "https://images.unsplash.com/photo-1639024471283-03518883512d?auto=format&fit=crop&q=80&w=300&h=300"
    },
    {
      id: 6,
      title: "Combo Casal",
      description: "2 Hambúrgueres, 1 Batata Grande e 2 Refrigerantes.",
      oldPrice: "R$ 75,00",
      newPrice: "R$ 55,90",
      isPopular: true,
      image: "https://images.unsplash.com/photo-1594212699903-ec8a3eca50f5?auto=format&fit=crop&q=80&w=300&h=300"
    },
    {
      id: 7,
      title: "Brownie com Sorvete",
      description: "Brownie quentinho acompanhado de sorvete de creme.",
      oldPrice: "R$ 18,00",
      newPrice: "R$ 12,50",
      isPopular: false,
      image: "https://images.unsplash.com/photo-1606313564200-e75d5e30476c?auto=format&fit=crop&q=80&w=300&h=300"
    }
  ];
</script>

<section class="py-8 w-full px-4 md:px-8">
  
    <div class="flex items-center mb-8">
      <h2 class="text-3xl md:text-4xl font-black text-gray-900 tracking-tight">
        Promo<span class="text-[#ff5722]">ções</span>
      </h2>
    </div>
  <!-- The Orange Circular Edge Container -->
  <div class="relative border-[4px] border-[#ff5722] rounded-[3rem] p-6 md:p-8 bg-orange-50/40 shadow-xl overflow-hidden">
    
    <!-- Title Section (Centered, no icon) -->

    <!-- Content Container -->
    <div class="w-full">
      
      <!-- Mobile: flex scroll. Desktop: Grid to fill width -->
      <div class="flex md:grid md:grid-cols-3 lg:grid-cols-4 md:grid-rows-2 gap-4 overflow-x-auto md:overflow-visible pb-6 md:pb-0 snap-x snap-mandatory hide-scrollbar">
        {#each pratos as prato}
          <!-- Added w-full so the wrapper fills the grid cell -->
          <div class="snap-start min-w-[280px] w-full flex-shrink-0">
             <MelhoresPratosCard {...prato} />
          </div>
        {/each}
      </div>

    </div>
  </div>
</section>

<style>
  .hide-scrollbar::-webkit-scrollbar {
    display: none;
  }
  .hide-scrollbar {
    -ms-overflow-style: none;
    scrollbar-width: none;
  }
</style>
