<script>
  import { onMount } from 'svelte';
  import Header from './Header.svelte';
  import Stat from './Stat.svelte';

  export let restaurantName = 'mamma-mia-mario';

  let loading = true;
  let error = null;

  let deliveryMode = 'delivery'; // 'delivery' | 'pickup'
  let activeCategory = 'Mais Pedidos';
  let searchQuery = '';

  let restaurant = {
    name: 'Mamma mia Mario',
    badge: 'LN',
    est: 'EST. 2018',
    isOpen: true,
    closingTime: '22:00',
    description: 'Um pedacinho de Napoli. Pizza no fogão a lenha, feita com amor atrás do armário.',
    rating: 4.9,
    reviewsCount: 324,
    cuisine: 'Italiana',
    category: 'Pizza',
    priceRange: 'R$',
    deliveryTime: '20–30 min',
    deliveryFee: 1.99,
    minOrder: 12.00,
    distance: '0.8 km',
    address: '746 Valencia St, San Francisco, CA',
    neighborhood: 'Mission District',
    phone: '(415) 555-0148',
    bannerUrl: 'https://lh3.googleusercontent.com/aida-public/AB6AXuD8i7X0qDoCvkouDq-uUv7UEDtkKSTah3gKO68ElZOQTiDH0KjKi_fyaIcGAqaF7mZ_ZgqueJ6j_U-Z-tt30B1fuhs28n6b5ZhcHjaFmjEgmxSSCJChcNGK27S-R-GsleUn0DChdITkkqd3lz-H6R4XMb81BP7tfPtAvQa2d2wjv4dqHh-zPRTB2cfC81OAkENDitgwJzfayVROlNm2Y_cpClw84FTUl8AaFP--O1M'
  };

  let categories = ['Mais Pedidos', 'Pizzas', 'Acompanhamentos', 'Sobremesas', 'Bebidas'];

  let dishes = [
    {
      id: 'margherita',
      name: 'Margherita',
      tag: 'MOST LOVED',
      description: 'San Marzano tomato, fior di latte mozzarella, fresh basil.',
      price: 16.00,
      category: 'Pizzas',
      imageUrl: 'https://lh3.googleusercontent.com/aida-public/AB6AXuCb-eU4XxPHWKrxNYHn7-TPGZZiDJ7J43_wyHW35GGvB4ZYni8tfx_WXwM5FpIXR5sPzrKxs7IsbzM8We9VCULm7jAj3tnmjxGjgDsOIjDnBTMIWNwM3hGed30aeoT2VhHS9PkWMmiaJALNAfdW7EcUR2jaBsZXviPHVg-BLQ_Yf-JaqHsdxRAb3o6FWddiPqt9vxwkUpiL3darOkhD4kKfXX9-ff_dhL3G24HQmLw'
    },
    {
      id: 'spicy-diavola',
      name: 'Spicy Diavola',
      tag: 'MOST LOVED',
      description: 'Spicy salami, mozzarella, tomato and a drizzle of chili honey.',
      price: 19.00,
      category: 'Pizzas',
      imageUrl: 'https://lh3.googleusercontent.com/aida-public/AB6AXuCTmOctI0RewSPwuz-BWjt1KjY06cnghRQd_5aD3yZ1mdACQ5v5UPrVcEaIeltyls3yjVEikrahYf-2fyZK4fwkm_VkaoGT7Rf2YHIePG4HfGnJJBV-j7xf-amYSPa3IDHVXFqKEhwbdbJomOR_fOEgiJJV31ru_FWjl8K31Wfai5rpjXVPa8DMj8HxJ_xQ4WWAwo0ZsO10gZXIHGP9fjCJZqNz6-vn7YPrM4fqcM8'
    },
    {
      id: 'burrata-tomatoes',
      name: 'Burrata & tomatoes',
      tag: null,
      description: 'Creamy burrata, heirloom tomatoes, basil and extra virgin olive oil.',
      price: 14.00,
      category: 'Acompanhamentos',
      imageUrl: 'https://lh3.googleusercontent.com/aida-public/AB6AXuBFETatsb6xiRzdMPM-1wMApk2RxLzpPFaJRKZ7yhdMVwciJo0Y53nDdCJM2yXu6qiqF4MZhPHNMAMwt_xVzRCnp2QM2M1YnoXM0bZks8dQ9-rrvPxY2LrYt1WwMYvUKvwgdw4FMn-O9G0clEClprlRhqIKNsE95Xpbulp8jfHSz4g_3Snl6Isi9kNw_rLR0ZtjJf8FsS4esS6B-XfCxNztdT6WBZKjFfkPHG95JYM'
    },
    {
      id: 'truffle-mushroom',
      name: 'Truffle Mushroom',
      tag: null,
      description: 'Roasted mushrooms, mozzarella, truffle cream and thyme.',
      price: 21.00,
      category: 'Pizzas',
      imageUrl: 'https://lh3.googleusercontent.com/aida-public/AB6AXuBgr4efN2zG2j5YdM4zBLIxePdIXevZq6BqNr273gvcuro9rXupd5oJaumsiQdZWIvLQAmvRpgdx85dyQqVn8PN50wzujrML8REWF96ZjCmsthj7QUrn4MIxisp7MMuSBQQWOL5JS2acb1mKF5rvuynj0qTiF1D1OVrpJeosxBLpnkKe8AcjAZssWraaTd70DXGmAVWrAFsiTR7auZsTqpJEMGzhk200QXJv4B2Tm4'
    },
    {
      id: 'crispy-arancini',
      name: 'Crispy Arancini',
      tag: null,
      description: 'Three golden risotto bites with mozzarella and house marinara.',
      price: 10.00,
      category: 'Acompanhamentos',
      imageUrl: 'https://lh3.googleusercontent.com/aida-public/AB6AXuA5M0wIG0STOjSlXO8KlzmDCmlJk4xNYmBVlb1wvHwjQJFOfHQgHgrQozsShpiTvjfx5PqIGqXPLqhqREqBdMQ4Ktm3b3114QqfxXvZgmUDYHb5MwXTkGITtp4VAbzXXk5iC-35fcFlWQLN6aXYpBkVXFEOnk0qjHsMpvoKmMuBIfamWVh_TRVW2Za1bLx8yrQ_P7tsPN3J_Xrw2VO8vjLn4ePchVL8DLJec2zaVj4'
    },
    {
      id: 'classic-tiramisu',
      name: 'Classic Tiramisù',
      tag: null,
      description: 'Espresso-soaked ladyfingers, mascarpone and a dusting of cocoa.',
      price: 8.00,
      category: 'Sobremesas',
      imageUrl: 'https://lh3.googleusercontent.com/aida-public/AB6AXuBEYrCpYgj2C2Uwt4Cs1myULdkmW_Rmh2yKrUw-_k7-tBFNS4Fl_ARM5OFNEprKyTnd0noJEtBjX9mgUNa0Macr0jt4Ci7iC1M5kWVCBq4FYwJ3xSPfa1EakfWraqlc5JYjtVCGDQ10swEnhKggEKkqycFOr_8zVWj8LyQjsToN4o6mkW_i70woK-300ZoysMTwFxVSaypKBjVx1LGMzmPMVv3JjPTFSd43fVUkm1M'
    }
  ];

  let cart = [];
  let isFavorite = false;

  $: cartTotal = cart.reduce((sum, item) => sum + (item.price * item.quantity), 0);
  $: totalCartItems = cart.reduce((sum, item) => sum + item.quantity, 0);
  $: filteredDishes = dishes.filter(d => {
    const matchesSearch = !searchQuery || d.name.toLowerCase().includes(searchQuery.toLowerCase()) || d.description.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesCat = activeCategory === 'Mais Pedidos' || d.category === activeCategory;
    return matchesSearch && matchesCat;
  });

  async function loadRestaurantData() {
    loading = true;
    error = null;
    try {
      // API call: /api/r/{restaurant_name}
      const res = await fetch(`/api/r/${restaurantName}`);
      if (res.ok) {
        const data = await res.json();
        if (data.restaurant) restaurant = { ...restaurant, ...data.restaurant };
        if (data.dishes) dishes = data.dishes;
      }
    } catch (e) {
      console.warn('Could not fetch from API, using fallback data:', e);
    } finally {
      loading = false;
    }
  }

  async function handleAddToCart(dish) {
    try {
      // API call: /api/r/{restaurant_name}/{dish}
      await fetch(`/api/r/${restaurantName}/${dish.id}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ quantity: 1 })
      });
    } catch (e) {
      console.warn('API call failed:', e);
    }

    const existing = cart.find(item => item.id === dish.id);
    if (existing) {
      existing.quantity += 1;
      cart = [...cart];
    } else {
      cart = [...cart, { ...dish, quantity: 1 }];
    }
  }

  async function toggleFavorite() {
    isFavorite = !isFavorite;
    try {
      await fetch(`/api/r/${restaurantName}/favorite`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ favorite: isFavorite })
      });
    } catch (e) {
      console.warn('Failed to update favorite status:', e);
    }
  }

  let cepInput = '';
  let calculatingShipping = false;
  let deliveryCalculated = false;

  function formatCep(e) {
    let value = e.target.value.replace(/\D/g, '').slice(0, 8);
    if (value.length > 5) {
      value = value.replace(/^(\d{5})(\d)/, '$1-$2');
    }
    cepInput = value;
  }

  async function handleCalculateShipping() {
    const rawCep = cepInput.replace(/\D/g, '');
    if (rawCep.length < 8) return;

    calculatingShipping = true;
    try {
      // API call: /api/r/{restaurantName}/shipping?cep={rawCep}
      const res = await fetch(`/api/r/${restaurantName}/shipping?cep=${rawCep}`);
      if (res.ok) {
        const data = await res.json();
        if (data.deliveryFee !== undefined) restaurant.deliveryFee = data.deliveryFee;
        if (data.deliveryTime) restaurant.deliveryTime = data.deliveryTime;
        if (data.distance) restaurant.distance = data.distance;
      } else {
        // Fallback simulated calculation
        restaurant.deliveryFee = 3.50;
        restaurant.deliveryTime = '25–35 min';
        restaurant.distance = '1.4 km';
      }
      deliveryCalculated = true;
    } catch (err) {
      restaurant.deliveryFee = 3.50;
      restaurant.deliveryTime = '25–35 min';
      restaurant.distance = '1.4 km';
      deliveryCalculated = true;
    } finally {
      calculatingShipping = false;
    }
  }

  onMount(() => {
    loadRestaurantData();
  });
</script>

<Header cartCount={totalCartItems} />

<main class="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-6 font-sans text-gray-800">
  <!-- Breadcrumb Navigation -->
  <nav class="flex items-center text-xs font-medium text-gray-500 mb-5 gap-2">
    <a class="hover:text-gray-800 transition" href="#restaurants">Restaurantes</a>
    <svg class="w-3 h-3 text-gray-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
      <path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
    </svg>
    <a class="hover:text-gray-800 transition" href="#category">{restaurant.cuisine}</a>
    <svg class="w-3 h-3 text-gray-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
      <path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
    </svg>
    <span class="text-gray-900 font-semibold">{restaurant.name}</span>
  </nav>

  <!-- Hero Image Banner -->
  <div class="relative w-full h-[320px] md:h-[400px] rounded-2xl overflow-hidden shadow-sm mb-6 bg-gray-100">
    <img
      alt="Interior do Restaurante {restaurant.name}"
      class="w-full h-full object-cover"
      src={restaurant.bannerUrl}
    />
    <!-- "Ver todas as fotos" Floating Action Button -->
    <button class="absolute bottom-4 right-4 bg-white/95 hover:bg-white text-gray-800 text-xs font-semibold px-4 py-2.5 rounded-xl shadow-md backdrop-blur-sm flex items-center gap-2 transition hover:scale-[1.02]">
      <svg class="w-4 h-4 text-gray-600" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <rect width="18" height="18" x="3" y="3" rx="2" ry="2"/>
        <circle cx="9" cy="9" r="2"/>
        <path d="m21 15-3.086-3.086a2 2 0 0 0-2.828 0L6 21"/>
      </svg>
      <span>Ver todas as fotos</span>
    </button>
  </div>

  <!-- Restaurant Header Card -->
  <section class="flex flex-col md:flex-row md:items-start justify-between gap-6 mb-6">
    <div class="flex items-start gap-4">
      <!-- Badge Avatar -->
      <div class="w-16 h-16 rounded-xl bg-[#faebe6] border border-[#f5d5cc] flex flex-col items-center justify-center text-[#ea4519] flex-shrink-0 shadow-sm">
        <span class="text-xl font-bold tracking-tight">{restaurant.badge}</span>
        <span class="text-[9px] uppercase tracking-wider font-semibold opacity-80">{restaurant.est}</span>
      </div>
      <!-- Details -->
      <div>
        <div class="flex flex-wrap items-center gap-3">
          <h1 class="text-2xl sm:text-3xl font-extrabold text-gray-900 tracking-tight">{restaurant.name}</h1>
          <span class="bg-emerald-50 text-emerald-700 text-xs px-2.5 py-0.5 rounded-full font-semibold border border-emerald-200">
            {restaurant.isOpen ? `Aberto até as ${restaurant.closingTime}` : 'Fechado'}
          </span>
        </div>
        <p class="text-sm text-gray-500 mt-1 max-w-xl">
          {restaurant.description}
        </p>
        <div class="flex items-center text-xs text-gray-600 mt-2 gap-1.5 flex-wrap">
          <span class="flex items-center text-amber-500 font-bold">
            <svg class="w-3.5 h-3.5 fill-current mr-1" viewBox="0 0 24 24">
              <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/>
            </svg>
            {restaurant.rating}
          </span>
          <span class="text-gray-500">({restaurant.reviewsCount} avaliações)</span>
          <span class="text-gray-300">·</span>
          <span class="text-gray-600">{restaurant.cuisine}</span>
          <span class="text-gray-300">·</span>
          <span class="text-gray-600">{restaurant.category}</span>
          <span class="text-gray-300">·</span>
          <span class="text-gray-600">{restaurant.priceRange}</span>
        </div>
      </div>
    </div>
    <!-- Action Button -->
    <div>
      <button
        on:click={toggleFavorite}
        class="w-full md:w-auto inline-flex items-center justify-center gap-2 border border-gray-200 hover:border-gray-300 bg-white {isFavorite ? 'text-red-500' : 'text-gray-700'} hover:text-black font-semibold text-xs py-2.5 px-4 rounded-xl shadow-sm transition"
      >
        <svg class="w-4 h-4 {isFavorite ? 'fill-current text-red-500' : 'text-gray-500'}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M19 14c1.49-1.46 3-3.21 3-5.5A5.5 5.5 0 0 0 16.5 3c-1.76 0-3 .5-4.5 2-1.5-1.5-2.74-2-4.5-2A5.5 5.5 0 0 0 2 8.5c0 2.3 1.5 4.05 3 5.5l7 7Z"/>
        </svg>
        <span>{isFavorite ? 'Restaurante Favoritado' : 'Favoritar restaurante'}</span>
      </button>
    </div>
  </section>

  <!-- Quick Info Stats Bar -->

  <!-- Layout Grid: Menu Items vs. Sidebar Cart -->
  <div class="grid grid-cols-1 lg:grid-cols-3 gap-8 items-start">
    <!-- BEGIN: MenuSection (Left 2 cols) -->
    <div class="lg:col-span-2">
      <!-- Category Navigation & Search -->
      <div class="border-b border-gray-200 flex items-center justify-between pb-1 mb-8 overflow-x-auto gap-4">
        <div class="flex gap-6 whitespace-nowrap">
          {#each categories as category}
            <button
              on:click={() => activeCategory = category}
              class="text-xs font-semibold pb-3 -mb-[1px] transition {activeCategory === category ? 'text-[#ea4519] border-b-2 border-[#ea4519]' : 'text-gray-500 hover:text-gray-900'}"
            >
              {category}
            </button>
          {/each}
        </div>
        <div class="relative flex items-center mb-2">
          <input
            bind:value={searchQuery}
            placeholder="Buscar pratos..."
            class="text-xs pl-7 pr-2 py-1 bg-gray-50 border border-gray-200 rounded-lg focus:outline-none focus:border-[#ea4519]"
          />
          <svg class="w-3.5 h-3.5 text-gray-400 absolute left-2 pointer-events-none" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="11" cy="11" r="8"/>
            <line x1="21" y1="21" x2="16.65" y2="16.65"/>
          </svg>
        </div>
      </div>

      <!-- Section Title -->
      <div class="mb-6">
        <h2 class="text-xl font-bold text-gray-900">{activeCategory}</h2>
        <p class="text-xs text-gray-500 mt-0.5">Os favoritos da casa. Ótimo para experimentar.</p>
      </div>

      <!-- Food Cards Grid -->
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        {#each filteredDishes as dish (dish.id)}
          <article class="bg-gray-400 rounded-2xl border border-gray-200/80 p-4 flex justify-between gap-3 hover:border-gray-300 hover:shadow-sm transition group cursor-pointer">
            <div class="flex flex-col justify-between flex-1">
              <div>
                {#if dish.tag}
                  <span class="inline-block text-[9px] font-bold tracking-wide uppercase text-red-600 bg-red-50 px-1.5 py-0.5 rounded mb-1">
                    {dish.tag}
                  </span>
                {/if}
                <h3 class="font-bold text-sm text-gray-900 group-hover:text-[#ea4519] transition">
                  {dish.name}
                </h3>
                <p class="text-xs text-gray-500 line-clamp-2 mt-1">
                  {dish.description}
                </p>
              </div>
              <div class="font-bold text-sm text-gray-900 mt-4">
                ${dish.price.toFixed(2)}
              </div>
            </div>
            <div class="relative w-28 h-28 flex-shrink-0 rounded-xl overflow-hidden bg-gray-100">
              <img alt={dish.name} class="w-full h-full object-cover" src={dish.imageUrl} />
              <button
                on:click|stopPropagation={() => handleAddToCart(dish)}
                class="absolute bottom-2 right-2 w-7 h-7 bg-white text-gray-800 rounded-full flex items-center justify-center shadow-md hover:bg-[#ea4519] hover:text-white transition"
                title="Adicionar ao pedido"
              >
                <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <line x1="12" y1="5" x2="12" y2="19"/>
                  <line x1="5" y1="12" x2="19" y2="12"/>
                </svg>
              </button>
            </div>
          </article>
        {/each}
      </div>

      <!-- Bottom Dietary / Allergy Notice -->
      <div class="mt-8 flex items-center gap-2 text-xs text-gray-400">
        <svg class="w-4 h-4 flex-shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="12" cy="12" r="10"/>
          <line x1="12" y1="16" x2="12" y2="12"/>
          <line x1="12" y1="8" x2="12.01" y2="8"/>
        </svg>
        <p>Allergies or dietary needs? Please call the restaurant at {restaurant.phone}.</p>
      </div>
    </div>
    <!-- END: MenuSection -->

    <!-- BEGIN: Sidebar (Right 1 col) -->
    <aside class="space-y-6 lg:sticky lg:top-28">
      <!-- Location Box ("Logo na esquina" on top) -->
      <div class="bg-white border border-gray-200 rounded-2xl p-5 shadow-sm">
        <h3 class="text-sm font-bold text-gray-900 mb-3">Logo na esquina</h3>
        <!-- Schematic Street Map Preview -->
        <div class="relative w-full h-36 bg-[#edece8] rounded-xl overflow-hidden mb-3 border border-gray-200 flex items-center justify-center">
          <div class="absolute inset-0 p-2 grid grid-cols-3 gap-2">
            <div class="bg-[#e4e2dd] rounded"></div>
            <div class="bg-[#e4e2dd] rounded relative">
              <span class="absolute top-1 left-1/2 -translate-x-1/2 text-[8px] font-bold text-gray-400 tracking-wider">18TH ST</span>
            </div>
            <div class="bg-[#e4e2dd] rounded"></div>
            <div class="bg-[#e4e2dd] rounded"></div>
            <div class="bg-[#dad7d0] rounded flex items-center justify-center">
              <!-- Location Pin Marker -->
              <div class="w-8 h-8 rounded-full bg-[#ea4519] flex items-center justify-center shadow-lg border-2 border-white">
                <svg class="w-4 h-4 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M18 2v6a3 3 0 0 1-3 3 3 3 0 0 1-3-3V2"/>
                  <path d="M15 2v20"/>
                  <path d="M6 2v7a2 2 0 0 0 2 2h0a2 2 0 0 0 2-2V2"/>
                  <path d="M8 11v11"/>
                </svg>
              </div>
            </div>
            <div class="bg-[#e4e2dd] rounded"></div>
            <div class="bg-[#e4e2dd] rounded"></div>
            <div class="bg-[#e4e2dd] rounded relative">
              <span class="absolute bottom-1 left-1/2 -translate-x-1/2 text-[8px] font-bold text-gray-400 tracking-wider">VALENCIA ST</span>
            </div>
            <div class="bg-[#e4e2dd] rounded"></div>
          </div>
        </div>

        <!-- Location Info -->
        <div>
          <div class="text-xs font-bold text-gray-900">{restaurant.address}</div>
          <div class="text-[11px] text-gray-500 mt-0.5">{restaurant.neighborhood} · {restaurant.distance} away</div>
          <a class="inline-flex items-center gap-1 text-xs font-bold text-[#ea4519] hover:underline mt-2" href="#directions">
            <span>Obter rotas</span>
            <svg class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="7" y1="17" x2="17" y2="7"/>
              <polyline points="7 7 17 7 17 17"/>
            </svg>
          </a>
        </div>

        <!-- Calcular Frete Section -->
        <div class="mt-4 pt-4 border-t border-gray-100">
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
              class="flex-1 text-xs px-3 py-2 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:border-[#ea4519] transition placeholder-gray-400"
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
                <span class="font-bold text-gray-900">R${restaurant.deliveryFee.toFixed(2)}</span>
              </div>
              <div class="flex justify-between items-center text-gray-700">
                <span class="text-gray-500">Prazo estimado:</span>
                <span class="font-bold text-gray-900">{restaurant.deliveryTime}</span>
              </div>
              <div class="flex justify-between items-center text-gray-700">
                <span class="text-gray-500">Distância calculada:</span>
                <span class="font-bold text-gray-900">{restaurant.distance}</span>
              </div>
            </div>
          {/if}
        </div>
      </div>

      <!-- Cart Box ("Seu Pedido" below) -->
      <div class="bg-white border border-gray-200 rounded-2xl p-5 shadow-sm">
        <h2 class="text-lg font-bold text-gray-900 mb-4">Seu Pedido</h2>
        <!-- Delivery / Pickup Toggle -->
        <div class="grid grid-cols-2 p-1 bg-gray-100/90 rounded-xl mb-6">
          <button
            on:click={() => deliveryMode = 'delivery'}
            class="py-1.5 text-xs font-semibold rounded-lg transition {deliveryMode === 'delivery' ? 'bg-white text-[#ea4519] shadow-sm' : 'text-gray-500 hover:text-gray-800'}"
          >
            Entrega
          </button>
          <button
            on:click={() => deliveryMode = 'pickup'}
            class="py-1.5 text-xs font-semibold rounded-lg transition {deliveryMode === 'pickup' ? 'bg-white text-[#ea4519] shadow-sm' : 'text-gray-500 hover:text-gray-800'}"
          >
            Pickup
          </button>
        </div>

        {#if cart.length === 0}
          <!-- Empty Cart State -->
          <div class="text-center py-6 px-4">
            <div class="w-12 h-12 rounded-full bg-gray-50 flex items-center justify-center mx-auto mb-3 text-gray-400 border border-gray-100">
              <svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M6 2 3 6v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2V6l-3-4Z"/>
                <path d="M3 6h18"/>
                <path d="M16 10a4 4 0 0 1-8 0"/>
              </svg>
            </div>
            <h3 class="text-sm font-bold text-gray-800 mb-1">Algo delicioso te espera</h3>
            <p class="text-xs text-gray-400 max-w-xs mx-auto">
              Adicione seus favoritos do cardápio e nós cuidamos do resto.
            </p>
          </div>
        {:else}
          <!-- Cart Items List -->
          <div class="space-y-3 mb-4">
            {#each cart as item}
              <div class="flex items-center justify-between text-xs">
                <div class="flex items-center gap-2">
                  <span class="font-bold text-[#ea4519]">{item.quantity}x</span>
                  <span class="font-medium text-gray-800">{item.name}</span>
                </div>
                <span class="font-semibold text-gray-900">${(item.price * item.quantity).toFixed(2)}</span>
              </div>
            {/each}
          </div>
        {/if}

        <!-- Minimum checkout button -->
        <div class="mt-4 pt-4 border-t border-gray-100">
          {#if cartTotal >= restaurant.minOrder}
            <button class="w-full py-3 bg-[#ea4519] hover:bg-[#d63c12] text-white font-semibold text-xs rounded-xl shadow-md text-center transition">
              Finalizar Pedido (${cartTotal.toFixed(2)})
            </button>
          {:else}
            <button class="w-full py-3 bg-gray-100 text-gray-400 font-semibold text-xs rounded-xl cursor-not-allowed text-center transition" disabled>
              Mínimo R${restaurant.minOrder.toFixed(2)} para entrega
            </button>
          {/if}
        </div>
      </div>
    </aside>
    <!-- END: Sidebar -->
  </div>
</main>
