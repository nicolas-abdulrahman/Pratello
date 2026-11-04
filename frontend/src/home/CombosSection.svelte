<script>
  import { onMount } from 'svelte';
  import CombosCard from './CombosCard.svelte';
  import ItemModal from '../dish/ItemModal.svelte';

  let scrollContainer;
  let activeIndex = $state(0);

  let isModalOpen = $state(false);
  let selectedItem = $state(null);

  function openModal(item) {
    selectedItem = item;
    isModalOpen = true;
  }

  const combos = [
    {
      id: 1,
      title: "Combo Casal",
      desc: "2 Hambúrgueres + 1 Batata Grande + 2 Refrigerantes.",
      price: "R$ 55,90",
      originalPrice: "R$ 75,00",
      image: "https://images.unsplash.com/photo-1594212699903-ec8a3eca50f5?auto=format&fit=crop&q=80&w=400&h=300"
    },
    {
      id: 2,
      title: "Combo Solo",
      desc: "1 Hambúrguer + 1 Batata Média + 1 Refrigerante.",
      price: "R$ 32,90",
      originalPrice: "R$ 42,00",
      image: "https://images.unsplash.com/photo-1561758033-d89a9ad46330?auto=format&fit=crop&q=80&w=400&h=300"
    },
    {
      id: 3,
      title: "Combo Família",
      desc: "4 Hambúrgueres + 2 Batatas Grandes + 1 Refrigerante 2L.",
      price: "R$ 99,90",
      originalPrice: "R$ 135,00",
      image: "https://images.unsplash.com/photo-1550547660-d9450f859349?auto=format&fit=crop&q=80&w=400&h=300"
    },
    {
      id: 4,
      title: "Combo Veggie",
      desc: "1 Hambúrguer de grão de bico + 1 Batata + 1 Suco natural.",
      price: "R$ 38,90",
      originalPrice: "",
      image: "https://images.unsplash.com/photo-1520072959219-c595dc870360?auto=format&fit=crop&q=80&w=400&h=300"
    },
    {
      id: 5,
      title: "Combo Kids",
      desc: "1 Hambúrguer pequeno + 1 Batata pequena + 1 Suco.",
      price: "R$ 25,90",
      originalPrice: "",
      image: "https://images.unsplash.com/photo-1568901346375-23c9450c58cd?auto=format&fit=crop&q=80&w=400&h=300"
    },
    {
      id: 6,
      title: "Combo Fitness",
      desc: "1 Hambúrguer de frango grelhado + Salada + Água.",
      price: "R$ 29,90",
      originalPrice: "",
      image: "https://images.unsplash.com/photo-1521390188846-e2a3a97453a0?auto=format&fit=crop&q=80&w=400&h=300"
    },
    {
      id: 7,
      title: "Combo Duplo",
      desc: "2 Hambúrgueres + 1 Batata Média + 1 Refrigerante.",
      price: "R$ 45,90",
      originalPrice: "R$ 58,00",
      image: "https://images.unsplash.com/photo-1568901346375-23c9450c58cd?auto=format&fit=crop&q=80&w=400&h=300"
    },
    {
      id: 8,
      title: "Combo Bacon Lover",
      desc: "1 Hambúrguer com muito bacon + 1 Batata + 1 Refrigerante.",
      price: "R$ 41,90",
      originalPrice: "R$ 52,00",
      image: "https://images.unsplash.com/photo-1553979459-d2229ba7433b?auto=format&fit=crop&q=80&w=400&h=300"
    },
    {
      id: 9,
      title: "Combo Doce Final",
      desc: "1 Hambúrguer tradicional + 1 Brownie + 1 Refrigerante.",
      price: "R$ 36,90",
      originalPrice: "",
      image: "https://images.unsplash.com/photo-1568901346375-23c9450c58cd?auto=format&fit=crop&q=80&w=400&h=300"
    },
    {
      id: 10,
      title: "Combo Asiático",
      desc: "1 Sanduíche de frango teriyaki + 1 Batata rústica.",
      price: "R$ 34,90",
      originalPrice: "R$ 40,00",
      image: "https://images.unsplash.com/photo-1521390188846-e2a3a97453a0?auto=format&fit=crop&q=80&w=400&h=300"
    },
    {
      id: 11,
      title: "Combo Executivo",
      desc: "1 Hambúrguer + 1 Salada Caesar + 1 Suco de laranja.",
      price: "R$ 33,90",
      originalPrice: "",
      image: "https://images.unsplash.com/photo-1594212699903-ec8a3eca50f5?auto=format&fit=crop&q=80&w=400&h=300"
    },
    {
      id: 12,
      title: "Combo Master",
      desc: "3 Hambúrgueres + 2 Batatas Grandes + 3 Refrigerantes.",
      price: "R$ 85,90",
      originalPrice: "R$ 110,00",
      image: "https://images.unsplash.com/photo-1550547660-d9450f859349?auto=format&fit=crop&q=80&w=400&h=300"
    }
  ];

  let step = $state(324);     // card width + gap (measured at runtime)
  let visible = $state(3);    // how many full cards fit (measured at runtime)
  let atStart = $state(true);
  let atEnd = $state(false);

  const totalPages = $derived(Math.ceil(combos.length / visible));

  function getCards() {
    return scrollContainer ? Array.from(scrollContainer.children) : [];
  }

  function measure() {
    const cards = getCards();
    if (cards.length < 2) return;
    step = cards[1].offsetLeft - cards[0].offsetLeft;
    const gap = step - cards[0].offsetWidth;
    visible = Math.max(1, Math.floor((scrollContainer.clientWidth + gap) / step));
    handleScroll();
  }

  function scrollToCard(index) {
    const last = Math.max(0, combos.length - visible);
    const target = Math.min(Math.max(index, 0), last);
    scrollContainer.scrollTo({ left: target * step, behavior: 'smooth' });
  }

  function currentCard() {
    return Math.round(scrollContainer.scrollLeft / step);
  }

  function scrollRight() {
    if (scrollContainer) scrollToCard(currentCard() + visible);
  }

  function scrollLeft() {
    if (scrollContainer) scrollToCard(currentCard() - visible);
  }

  function handleScroll() {
    if (!scrollContainer) return;
    const max = scrollContainer.scrollWidth - scrollContainer.clientWidth;
    atStart = scrollContainer.scrollLeft <= 5;
    atEnd = scrollContainer.scrollLeft >= max - 5;

    activeIndex = atEnd
      ? totalPages - 1
      : Math.min(totalPages - 1, Math.round(currentCard() / visible));
  }

  onMount(() => {
    measure();
    window.addEventListener('resize', measure);
    return () => window.removeEventListener('resize', measure);
  });
</script>

<section class="py-12 w-full px-4 md:px-8 overflow-hidden">
  
  <div class="flex items-center justify-center gap-4 mb-8 max-w-7xl mx-auto">
    <div class="h-px bg-[#f05a1f]/30 flex-grow"></div>
    <div class="flex items-center gap-3">
      <svg xmlns="http://www.w3.org/2000/svg" width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="#f05a1f" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 3h18v18H3z"/><path d="M3 9h18"/><path d="M9 21V9"/></svg>
      <h2 class="text-2xl md:text-3xl font-black text-gray-900">Combos <span class="text-[#f05a1f]">Rápidos</span></h2>
    </div>
    <div class="h-px bg-[#f05a1f]/30 flex-grow"></div>
  </div>

  <div class="relative group w-full max-w-[1920px] mx-auto">
    
    <div 
      bind:this={scrollContainer}
      on:scroll={handleScroll}
      class="flex overflow-x-auto gap-6 pb-6 pt-2 snap-x snap-mandatory hide-scrollbar scroll-smooth"
    >
      {#each combos as combo}
        <CombosCard {...combo} onOpenDetails={() => openModal(combo)} />
      {/each}
    </div>

    {#if !atStart}
      <button 
        on:click={scrollLeft}
        class="absolute left-0 top-[45%] -translate-y-1/2 bg-white text-[#ff5722] p-3 rounded-full shadow-[0_4px_12px_rgba(0,0,0,0.15)] hover:bg-gray-50 transition-all hover:scale-110 z-20 flex items-center justify-center border border-gray-100"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="m15 18-6-6 6-6"/></svg>
      </button>
    {/if}

    {#if !atEnd}
      <button 
        on:click={scrollRight}
        class="absolute right-0 top-[45%] -translate-y-1/2 bg-white text-[#ff5722] p-3 rounded-full shadow-[0_4px_12px_rgba(0,0,0,0.15)] hover:bg-gray-50 transition-all hover:scale-110 z-20 flex items-center justify-center border border-gray-100"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="m9 18 6-6-6-6"/></svg>
      </button>
    {/if}
  </div>

  <div class="flex justify-center gap-2 mt-4">
    {#each Array(totalPages) as _, i}
      <div 
        class="w-2.5 h-2.5 rounded-full transition-colors duration-300 
        {activeIndex === i ? 'bg-[#ff5722]' : 'bg-gray-300'}"
      ></div>
    {/each}
  </div>

</section>

<ItemModal 
  isOpen={isModalOpen} 
  onClose={() => isModalOpen = false} 
  item={selectedItem || {}} 
/>

<style>
  .hide-scrollbar::-webkit-scrollbar { display: none; }
  .hide-scrollbar { -ms-overflow-style: none; scrollbar-width: none; }
  .hide-scrollbar > :global(*) { scroll-snap-align: start; }
</style>
