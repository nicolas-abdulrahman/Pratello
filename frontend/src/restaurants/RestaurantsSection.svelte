<script >
  import RestaurantCard from './RestaurantCard.svelte';
  import RestaurantsNavbar from './RestaurantsNavbar.svelte';
  import RestaurantsLeftSideBar from './RestaurantsLeftSideBar.svelte';

  let showProfileSidebar = $state(false);

  const restaurants = [
    {
      name: 'Pai João',
      tagline: 'Hambúrguer Artesanal',
      cuisine: 'Lanches & Burgers',
      distance: '0.6 km',
      rating: 4.8,
      reviews: '2.3k',
      category: 'Lanches & Burgers',
      badge: 'Destaque da semana',
      badgeClass: 'bg-amber-500/20 text-amber-300',
      deliveryTime: '30-45 min',
      deliveryFee: 'Grátis',
      feeClass: 'text-emerald-700 bg-emerald-50',
      imageUrl:
        'https://lh3.googleusercontent.com/aida-public/AB6AXuCFmhePNwlSmB1YMgPzKwDCZj-s8i0PFk__wiHGFFC85DdcjxRKbGbiqkpd4iOs4lEGMOBrbxWJ5hncUO4i3O3XJfowTDv_VIXcSD0e3HcsjoJazcXB1h6LU2bNt3l5MJDaZzA7RYqkVn4uM1Y86lYlX9RqCNkLqmF-YTYEy3TKmju2xNphN-q2JmojJuCkifhTCAFao_XM1XiPwetyFQF0VhVP6KAAqO0Ebg1CXdnmG00NsCbrcwLO',
      bgFrom: '#211611',
      bgTo: '#3A2419',
      titleClass: 'text-amber-400 font-mono',
      subtitleClass: 'text-stone-300',
      emojis: ['🍔', '🍟', '🥤']
    },
    {
      name: "Armario Mario's",
      tagline: 'Italian Ristorante',
      cuisine: 'Massas & Pizza',
      distance: '1.2 km',
      rating: 4.7,
      reviews: '1.1k',
      category: 'Massas & Pizza',
      badge: 'Massas caseiras',
      badgeClass: 'bg-emerald-500/20 text-emerald-300',
      deliveryTime: '25-40 min',
      deliveryFee: 'R$ 4,90',
      feeClass: 'text-amber-800 bg-amber-50',
      imageUrl:
        'https://lh3.googleusercontent.com/aida-public/AB6AXuCFmhePNwlSmB1YMgPzKwDCZj-s8i0PFk__wiHGFFC85DdcjxRKbGbiqkpd4iOs4lEGMOBrbxWJ5hncUO4i3O3XJfowTDv_VIXcSD0e3HcsjoJazcXB1h6LU2bNt3l5MJDaZzA7RYqkVn4uM1Y86lYlX9RqCNkLqmF-YTYEy3TKmju2xNphN-q2JmojJuCkifhTCAFao_XM1XiPwetyFQF0VhVP6KAAqO0Ebg1CXdnmG00NsCbrcwLO',
      bgFrom: '#1E281F',
      bgTo: '#2B3B2B',
      titleClass: 'text-amber-300 font-serif',
      subtitleClass: 'text-emerald-300',
      emojis: ['🍕', '🍝', '🍷']
    },
    {
      name: 'Sakura',
      tagline: 'Contemporary Sushi',
      cuisine: 'Culinária Japonesa',
      distance: '1.5 km',
      rating: 4.8,
      reviews: '892',
      category: 'Culinária Japonesa',
      badge: 'Peixes frescos',
      badgeClass: 'bg-rose-500/20 text-rose-300',
      deliveryTime: '35-50 min',
      deliveryFee: 'R$ 6,50',
      feeClass: 'text-stone-700 bg-stone-100',
      imageUrl:
        'https://lh3.googleusercontent.com/aida-public/AB6AXuCFmhePNwlSmB1YMgPzKwDCZj-s8i0PFk__wiHGFFC85DdcjxRKbGbiqkpd4iOs4lEGMOBrbxWJ5hncUO4i3O3XJfowTDv_VIXcSD0e3HcsjoJazcXB1h6LU2bNt3l5MJDaZzA7RYqkVn4uM1Y86lYlX9RqCNkLqmF-YTYEy3TKmju2xNphN-q2JmojJuCkifhTCAFao_XM1XiPwetyFQF0VhVP6KAAqO0Ebg1CXdnmG00NsCbrcwLO',
      bgFrom: '#2B171A',
      bgTo: '#401F25',
      titleClass: 'text-rose-400 font-sans',
      subtitleClass: 'text-stone-300',
      emojis: ['🍣', '🍱', '🥢']
    }
  ];

  const trending = [
    {
      name: 'Cantina da Nonna',
      tagline: 'da Nonna',
      cuisine: 'Italiana',
      distance: '1.1 km',
      rating: 4.9,
      reviews: '',
      category: 'Italiana • R$ 35 - 80',
      badge: '',
      badgeClass: '',
      deliveryTime: '30-45 min',
      deliveryFee: 'Grátis acima de R$50',
      feeClass: 'text-emerald-700 bg-emerald-50',
      imageUrl:
        'https://lh3.googleusercontent.com/aida-public/AB6AXuCFmhePNwlSmB1YMgPzKwDCZj-s8i0PFk__wiHGFFC85DdcjxRKbGbiqkpd4iOs4lEGMOBrbxWJ5hncUO4i3O3XJfowTDv_VIXcSD0e3HcsjoJazcXB1h6LU2bNt3l5MJDaZzA7RYqkVn4uM1Y86lYlX9RqCNkLqmF-YTYEy3TKmju2xNphN-q2JmojJuCkifhTCAFao_XM1XiPwetyFQF0VhVP6KAAqO0Ebg1CXdnmG00NsCbrcwLO',
      bgFrom: '#291812',
      bgTo: '#3C241B',
      titleClass: 'text-amber-300',
      subtitleClass: 'text-stone-300',
      emojis: ['🍝']
    },
    {
      name: 'Taco Feliz',
      tagline: 'Autêntico Mexicano',
      cuisine: 'Mexicana',
      distance: '1.9 km',
      rating: 4.6,
      reviews: '',
      category: 'Mexicana • 20-35 min',
      badge: '',
      badgeClass: '',
      deliveryTime: '20-35 min',
      deliveryFee: 'Cupom 15% OFF',
      feeClass: 'text-amber-800 bg-amber-50',
      imageUrl:
        'https://lh3.googleusercontent.com/aida-public/AB6AXuCFmhePNwlSmB1YMgPzKwDCZj-s8i0PFk__wiHGFFC85DdcjxRKbGbiqkpd4iOs4lEGMOBrbxWJ5hncUO4i3O3XJfowTDv_VIXcSD0e3HcsjoJazcXB1h6LU2bNt3l5MJDaZzA7RYqkVn4uM1Y86lYlX9RqCNkLqmF-YTYEy3TKmju2xNphN-q2JmojJuCkifhTCAFao_XM1XiPwetyFQF0VhVP6KAAqO0Ebg1CXdnmG00NsCbrcwLO',
      bgFrom: '#2E2011',
      bgTo: '#453118',
      titleClass: 'text-amber-400',
      subtitleClass: 'text-stone-300',
      emojis: ['🌮']
    },
    {
      name: 'Pizzaria Bella Notte',
      tagline: 'Bella Notte',
      cuisine: 'Pizza artesanal',
      distance: '2.0 km',
      rating: 4.8,
      reviews: '',
      category: 'Pizza artesanal • Forno a lenha',
      badge: '',
      badgeClass: '',
      deliveryTime: '40-55 min',
      deliveryFee: 'R$ 6,00',
      feeClass: 'text-stone-700 bg-stone-100',
      imageUrl:
        'https://lh3.googleusercontent.com/aida-public/AB6AXuCFmhePNwlSmB1YMgPzKwDCZj-s8i0PFk__wiHGFFC85DdcjxRKbGbiqkpd4iOs4lEGMOBrbxWJ5hncUO4i3O3XJfowTDv_VIXcSD0e3HcsjoJazcXB1h6LU2bNt3l5MJDaZzA7RYqkVn4uM1Y86lYlX9RqCNkLqmF-YTYEy3TKmju2xNphN-q2JmojJuCkifhTCAFao_XM1XiPwetyFQF0VhVP6KAAqO0Ebg1CXdnmG00NsCbrcwLO',
      bgFrom: '#201511',
      bgTo: '#39241E',
      titleClass: 'text-amber-300',
      subtitleClass: 'text-stone-300',
      emojis: ['🍕']
    }
  ];

  const categories = [
    { emoji: '🍔', label: 'Hambúrguer', active: true },
    { emoji: '🍝', label: 'Italiana', active: false },
    { emoji: '🍱', label: 'Japonesa', active: false },
    { emoji: '🍕', label: 'Pizza', active: false },
    { emoji: '🍲', label: 'Brasileira', active: false },
    { emoji: '🥙', label: 'Árabe', active: false },
    { emoji: '🥡', label: 'Chinesa', active: false },
    { emoji: '🌱', label: 'Vegana', active: false },
    { emoji: '🧁', label: 'Doces', active: false }
  ];
</script>

<svelte:head>
  <title>Pratello - Delivery de Comida</title>
</svelte:head>

<div
  class="min-h-screen flex flex-row overflow-x-hidden antialiased select-none selection:bg-brand-primary selection:text-white bg-[#FAF6F0] text-[#2A2017]"
>
  <!-- Left sidebar (slide-in) -->
  <RestaurantsLeftSideBar
    show={showProfileSidebar}
    onClose={() => (showProfileSidebar = false)}
  />

  <!-- Main Content -->
  <main class="flex-1 min-h-screen px-8 py-6 space-y-9 max-w-7xl mx-auto">
    <!-- Top bar -->

    <!-- Hero Section -->
    <section
      class="relative rounded-3xl bg-gradient-to-r from-[#F6EDE0] via-[#FAF4EA] to-[#F7EFE2] border border-[#EDE2D2] overflow-hidden p-8 lg:p-10 shadow-sm"
    >
      <div
        class="absolute -right-10 -bottom-20 w-96 h-96 bg-amber-400/20 rounded-full blur-3xl pointer-events-none"
      ></div>
      <div
        class="absolute right-60 top-1/2 -translate-y-1/2 w-64 h-64 bg-amber-200/40 hero-blob pointer-events-none"
      ></div>
      <div class="grid grid-cols-1 lg:grid-cols-12 gap-6 items-center relative z-10">
        <div class="lg:col-span-7 space-y-4">
          <div
            class="inline-flex items-center gap-2 bg-amber-500/10 text-amber-800 px-3 py-1 rounded-full text-xs font-semibold uppercase tracking-wider"
          >
            <span>✨ Praticidade &amp; Agilidade</span>
          </div>
          <h1
            class="text-3xl lg:text-4xl xl:text-5xl font-extrabold text-stone-900 tracking-tight leading-[1.15]"
          >
            Descubra os melhores <br class="hidden xl:block" />restaurantes perto de você!
          </h1>
          <p class="text-stone-600 text-base max-w-lg font-normal">
            Variedade, sabor e praticidade garantidos, tudo em um só lugar no Pratello.
          </p>
          <div class="pt-2">
            <a
              class="inline-flex items-center gap-2.5 bg-gradient-to-r from-amber-500 to-amber-600 hover:from-amber-600 hover:to-amber-700 text-white font-semibold text-sm px-6 py-3.5 rounded-full shadow-md shadow-amber-500/25 hover:shadow-lg transition duration-200 group"
              href="#restaurantes-proximos"
            >
              <span>Explorar restaurantes</span>
              <svg
                class="w-4 h-4 transform group-hover:translate-x-1 transition-transform"
                fill="none"
                stroke="currentColor"
                stroke-width="2.5"
                viewBox="0 0 24 24"
              >
                <path d="M14 5l7 7m0 0l-7 7m7-7H3" stroke-linecap="round" stroke-linejoin="round" />
              </svg>
            </a>
          </div>
        </div>
        <div class="lg:col-span-5 relative flex justify-center items-center">
          <div class="relative w-72 h-72 sm:w-80 sm:h-80 flex items-center justify-center">
            <div
              class="absolute inset-0 bg-gradient-to-tr from-amber-400 to-yellow-300 rounded-full opacity-30 blur-2xl animate-pulse"
            ></div>
            <img
              alt="Hambúrguer artesanal Pratello"
              class="relative z-10 w-full h-full object-cover rounded-3xl shadow-xl hover:scale-102 transition duration-300"
              src="https://lh3.googleusercontent.com/aida-public/AB6AXuCFmhePNwlSmB1YMgPzKwDCZj-s8i0PFk__wiHGFFC85DdcjxRKbGbiqkpd4iOs4lEGMOBrbxWJ5hncUO4i3O3XJfowTDv_VIXcSD0e3HcsjoJazcXB1h6LU2bNt3l5MJDaZzA7RYqkVn4uM1Y86lYlX9RqCNkLqmF-YTYEy3TKmju2xNphN-q2JmojJuCkifhTCAFao_XM1XiPwetyFQF0VhVP6KAAqO0Ebg1CXdnmG00NsCbrcwLO"
            />
          </div>
        </div>
      </div>
    </section>

    <!-- Search + Filters -->
    <div class="pt-2" id="restaurantes-proximos">
      <div
        class="flex flex-col sm:flex-row items-center gap-3 bg-[#F4EDE2]/80 border border-[#E6DDCE] p-2 rounded-2xl shadow-xs"
      >
        <div class="relative flex-1 w-full">
          <svg
            class="w-5 h-5 text-stone-400 absolute left-3.5 top-1/2 -translate-y-1/2"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            viewBox="0 0 24 24"
          >
            <circle cx="11" cy="11" r="8"></circle>
            <path d="m21 21-4.35-4.35" stroke-linecap="round"></path>
          </svg>
          <input
            class="w-full bg-white border border-[#E2D6C6] rounded-xl pl-11 pr-4 py-2.5 text-sm text-stone-800 placeholder-stone-400 focus:outline-none focus:ring-2 focus:ring-amber-500 shadow-2xs font-medium"
            placeholder="Buscar restaurantes ou pratos próximos..."
            type="text"
          />
        </div>
        <div class="flex items-center gap-2 w-full sm:w-auto justify-end">
          <button
            class="inline-flex items-center gap-2 px-3.5 py-2.5 bg-white border border-[#E2D6C6] hover:bg-stone-50 rounded-xl text-xs font-semibold text-stone-700 transition"
          >
            <svg
              class="w-3.5 h-3.5 text-stone-500"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              viewBox="0 0 24 24"
            >
              <path
                d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z"
                stroke-linecap="round"
                stroke-linejoin="round"
              ></path>
            </svg>
            Filtros
          </button>
          <button
            class="inline-flex items-center gap-1.5 px-3 py-2.5 bg-white border border-[#E2D6C6] hover:bg-stone-50 rounded-xl text-xs font-medium text-stone-600 transition"
          >
            <span>Mais bem avaliados</span>
          </button>
        </div>
      </div>

      <div class="pt-3 space-y-1.5">
        <div class="flex items-center justify-between px-1">
          <span class="text-[11px] font-bold text-stone-500 uppercase tracking-wider">
            Categorias populares
          </span>
        </div>
        <div class="relative flex items-center">
          <div class="flex items-center gap-2 overflow-x-auto hide-scrollbar py-1 w-full pr-10">
            {#each categories as cat}
              <button
                class="shrink-0 flex items-center gap-1.5 px-3 py-1.5 rounded-xl border text-xs transition shadow-2xs"
                class:bg-amber-200={cat.active}
                class:border-amber-300={cat.active}
                class:text-stone-900={cat.active}
                class:font-semibold={cat.active}
                class:bg-[#F4EDE2]={!cat.active}
                class:border-[#E4DACB]={!cat.active}
                class:text-stone-700={!cat.active}
                class:font-medium={!cat.active}
                class:hover:bg-[#ECE4D6]={!cat.active}
              >
                <span class="text-sm">{cat.emoji}</span>
                {cat.label}
              </button>
            {/each}
          </div>
          <button
            aria-label="Ver mais categorias"
            class="absolute right-0 top-1/2 -translate-y-1/2 w-7 h-7 rounded-full bg-white border border-[#E4DACB] flex items-center justify-center text-stone-700 shadow-md hover:bg-stone-50 transition"
          >
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
              <path d="M9 5l7 7-7 7" stroke-linecap="round" stroke-linejoin="round" />
            </svg>
          </button>
        </div>
      </div>
    </div>

    <!-- Restaurants Section -->
    <section class="space-y-4">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2">
          <span class="text-amber-500 text-lg">✨</span>
          <h2 class="text-xl font-bold text-stone-900 tracking-tight">Restaurantes próximos</h2>
        </div>
        <a
          class="inline-flex items-center gap-1 text-xs font-bold text-amber-600 hover:text-amber-700 transition"
          href="#"
        >
          Ver todos
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
            <path d="M9 5l7 7-7 7" stroke-linecap="round" stroke-linejoin="round" />
          </svg>
        </a>
      </div>
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {#each restaurants as restaurant}
          <RestaurantCard {restaurant} />
        {/each}
      </div>
    </section>

    <!-- Trending Section -->
    <section class="space-y-4 pt-2 pb-12">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2">
          <span class="text-amber-500 text-lg">🔥</span>
          <h2 class="text-xl font-bold text-stone-900 tracking-tight">Em alta na sua região</h2>
        </div>
        <a
          class="inline-flex items-center gap-1 text-xs font-bold text-amber-600 hover:text-amber-700 transition"
          href="#"
        >
          Ver todos
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
            <path d="M9 5l7 7-7 7" stroke-linecap="round" stroke-linejoin="round" />
          </svg>
        </a>
      </div>
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
        {#each trending as item}
          <RestaurantCard restaurant={item} />
        {/each}
      </div>
    </section>
  </main>
</div>
