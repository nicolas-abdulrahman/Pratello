<script>
  import { createEventDispatcher } from 'svelte';

  export let cartCount = 0;
  export let deliveryAddress = 'Rua Padre Carapuceiro 1234 , Recife';
  export let activeNav = 'Restaurantes';
  export let searchQuery = '';

  const dispatch = createEventDispatcher();

  function handleSearch(e) {
    dispatch('search', searchQuery);
  }

  function handleLogin() {
    dispatch('login');
  }

  function handleCartClick() {
    dispatch('openCart');
  }

  function handleLocationClick() {
    dispatch('changeLocation');
  }
</script>

<header class="bg-white border-b border-gray-200 sticky top-0 z-50">
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
    <!-- Top header row -->
    <div class="flex items-center justify-between h-20 gap-4">
      <!-- Left: Brand Logo & Delivery Address -->
      <div class="flex items-center space-x-8">
        <!-- Brand Logo -->
        <a class="flex items-center gap-2 group text-[#ea4519]" href="/">
          <div class="flex items-center">
            <svg class="w-8 h-8 fill-current text-[#ea4519]" viewBox="0 0 24 24">
              <path d="M11 2a1 1 0 0 1 1 1v7a1 1 0 0 1-2 0V3a1 1 0 0 1 1-1zm-4 0a1 1 0 0 1 1 1v5a2 2 0 0 0 2 2h0a2 2 0 0 0 2-2V3a1 1 0 1 1 2 0v5a4 4 0 0 1-3 3.874V21a1 1 0 0 1-2 0v-9.126A4 4 0 0 1 6 8V3a1 1 0 0 1 1-1zm10 0a1 1 0 0 1 1 1v6a3 3 0 0 1-2 2.83V21a1 1 0 0 1-2 0v-9.17A3 3 0 0 1 17 9V3a1 1 0 0 1 1-1z"></path>
            </svg>
            <span class="text-2xl font-black tracking-tight text-[#ea4519] ml-1">pratello<span class="text-3xl font-black leading-none">.</span></span>
          </div>
        </a>

        <!-- Location Selector -->
        <button
          on:click={handleLocationClick}
          type="button"
          class="hidden md:flex items-center text-left py-1.5 px-3 rounded-xl hover:bg-gray-50 transition border border-transparent hover:border-gray-200"
        >
          <svg class="w-4 h-4 text-[#ea4519] mr-2.5 flex-shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M20 10c0 6-8 12-8 12s-8-6-8-12a8 8 0 0 1 16 0Z"/>
            <circle cx="12" cy="10" r="3"/>
          </svg>
          <div>
            <div class="text-[10px] uppercase font-bold tracking-wider text-gray-400">Entregar para:</div>
            <div class="text-xs font-semibold text-gray-800 flex items-center gap-1">
              {deliveryAddress}
              <svg class="w-3.5 h-3.5 text-gray-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="m6 9 6 6 6-6"/>
              </svg>
            </div>
          </div>
        </button>
      </div>

      <!-- Center: Search Bar -->
      <div class="flex-1 max-w-lg mx-4 hidden sm:block">
        <div class="relative">
          <div class="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none text-gray-400">
            <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="11" cy="11" r="8"/>
              <line x1="21" y1="21" x2="16.65" y2="16.65"/>
            </svg>
          </div>
          <input
            bind:value={searchQuery}
            on:input={handleSearch}
            class="w-full pl-10 pr-4 py-2.5 bg-gray-100/80 hover:bg-gray-100 focus:bg-white text-sm border-transparent focus:border-gray-300 rounded-xl focus:ring-0 transition placeholder-gray-400"
            placeholder="Restaurante, prato ou culinária"
            type="text"
          />
        </div>
      </div>

      <!-- Right: Auth and Cart -->
      <div class="flex items-center gap-4">
        <button
          on:click={handleLogin}
          type="button"
          class="flex items-center gap-2 text-sm font-medium text-gray-700 hover:text-black py-2 px-3 rounded-lg hover:bg-gray-100 transition"
        >
          <svg class="w-4 h-4 text-gray-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M19 21v-2a4 4 0 0 0-4-4H9a4 4 0 0 0-4 4v2"/>
            <circle cx="12" cy="7" r="4"/>
          </svg>
          <span>Entrar</span>
        </button>

        <button
          on:click={handleCartClick}
          type="button"
          class="flex items-center gap-2 text-sm font-medium text-gray-700 hover:text-black border border-gray-200 px-3.5 py-2 rounded-xl hover:border-gray-300 bg-white transition shadow-sm"
        >
          <svg class="w-4 h-4 text-gray-600" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M6 2 3 6v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2V6l-3-4Z"/>
            <path d="M3 6h18"/>
            <path d="M16 10a4 4 0 0 1-8 0"/>
          </svg>
          <span class="font-bold text-gray-900">{cartCount}</span>
        </button>
      </div>
    </div>

    <!-- Sub-navigation row -->
    <div class="flex items-center justify-between border-t border-gray-100 text-sm">
      <nav class="flex space-x-8">
        <a
          class="py-3 font-semibold transition {activeNav === 'Restaurantes' ? 'text-[#ea4519] border-b-2 border-[#ea4519]' : 'text-gray-500 hover:text-gray-900'}"
          href="/"
          on:click={() => activeNav = 'Restaurantes'}
        >
          Restaurantes
        </a>
        <a
          class="py-3 font-medium transition {activeNav === 'Ofertas' ? 'text-[#ea4519] border-b-2 border-[#ea4519]' : 'text-gray-500 hover:text-gray-900'}"
          href="#ofertas"
          on:click={() => activeNav = 'Ofertas'}
        >
          Ofertas
        </a>
        <a
          class="py-3 font-medium transition {activeNav === 'Favoritos' ? 'text-[#ea4519] border-b-2 border-[#ea4519]' : 'text-gray-500 hover:text-gray-900'}"
          href="#favoritos"
          on:click={() => activeNav = 'Favoritos'}
        >
          Favoritos
        </a>
      </nav>
      <div>
        <a class="py-3 text-xs font-medium text-gray-500 hover:text-gray-700 flex items-center gap-1" href="#ajuda">
          <svg class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/>
            <path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3"/>
            <line x1="12" y1="17" x2="12.01" y2="17"/>
          </svg>
          Precisa de ajuda?
        </a>
      </div>
    </div>
  </div>
</header>
