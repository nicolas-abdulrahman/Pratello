<script>
  import { onMount } from 'svelte';


  let currentPath = window.location.pathname;
  let restaurantSlug = currentPath.replace(/^\/+|\/+$/g, '');
  let restaurantData = null;
  let loading = false;
  let error = null;

  let customSlugInput = '';

  async function fetchRestaurant(slug) {
    if (!slug) {
      restaurantData = null;
      return;
    }
    loading = true;
    error = null;
    try {
      const res = await fetch(`/api/restaurants/${slug}`);
      if (!res.ok) throw new Error('Failed to fetch restaurant');
      restaurantData = await res.json();
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  }

  function navigateTo(path) {
    window.history.pushState({}, '', path);
    currentPath = window.location.pathname;
    restaurantSlug = currentPath.replace(/^\/+|\/+$/g, '');
    fetchRestaurant(restaurantSlug);
  }

  function handleCustomSearch() {
    if (customSlugInput.trim()) {
      navigateTo('/' + customSlugInput.trim().toLowerCase());
    }
  }

  onMount(() => {
    window.addEventListener('popstate', () => {
      currentPath = window.location.pathname;
      restaurantSlug = currentPath.replace(/^\/+|\/+$/g, '');
      fetchRestaurant(restaurantSlug);
    });

    if (restaurantSlug) {
      fetchRestaurant(restaurantSlug);
    }
  });
</script>

<div class="min-h-screen bg-slate-900 text-slate-100 flex flex-col">
  <!-- Header / Navigation Bar -->
  <header class="border-b border-slate-800 bg-slate-950/60 backdrop-blur sticky top-0 z-50">
    <div class="max-w-5xl mx-auto px-4 h-16 flex items-center justify-between">
      <div 
        class="flex items-center gap-2 cursor-pointer font-bold text-xl tracking-tight text-white hover:opacity-90 transition"
        onclick={() => navigateTo('/')}
        onkeydown={(e) => e.key === 'Enter' && navigateTo('/')}
        tabindex="0"
        role="button"
      >
        <span class="text-2xl">🍽️</span>
        <span class="bg-gradient-to-r from-amber-400 to-orange-500 bg-clip-text text-transparent">Pratello</span>
      </div>

      <nav class="flex items-center gap-3">
        <button
          onclick={() => navigateTo('/bobpizzeria')}
          class="px-3 py-1.5 text-sm rounded-lg border border-slate-700 bg-slate-800/80 hover:bg-slate-700 transition"
        >
          🍕 Bob's Pizzeria
        </button>
        <button
          onclick={() => navigateTo('/deliciadeprazeres')}
          class="px-3 py-1.5 text-sm rounded-lg border border-slate-700 bg-slate-800/80 hover:bg-slate-700 transition"
        >
          🧁 Delícia de Prazeres
        </button>
      </nav>
    </div>
  </header>

  <!-- Main Content Area -->
  <main class="flex-1  max-w-5xl mx-auto w-full px-4 py-8">
    {#if !restaurantSlug}
      <!-- Home Landing View (at /) -->
      <section class="text-center py-12">
        <div class="inline-block p-2 px-4 bg-amber-500/10 border border-amber-500/30 rounded-full text-amber-300 text-sm font-medium mb-6">
          Powered by Svelte + Tailwind + Go net/http
        </div>
        <h1 class="text-4xl sm:text-5xl font-extrabold tracking-tight mb-4">
          Discover Places on <span class="text-amber-400">Pratello</span>
        </h1>
        <p class="text-slate-400 max-w-xl mx-auto text-lg mb-8">
          A lightning-fast platform built with Go backend routing and a responsive Svelte frontend.
        </p>

        <!-- Jump to restaurant input -->
        <div class="max-w-md mx-auto flex gap-2 mb-12">
          <input
            type="text"
            bind:value={customSlugInput}
            onkeydown={(e) => e.key === 'Enter' && handleCustomSearch()}
            placeholder="e.g. bobpizzeria, sushi-central"
            class="flex-1 px-4 py-2.5 rounded-lg bg-slate-800 border border-slate-700 text-white placeholder-slate-500 focus:outline-none focus:border-amber-400 transition"
          />
          <button
            onclick={handleCustomSearch}
            class="px-5 py-2.5 bg-amber-500 hover:bg-amber-400 text-slate-950 font-semibold rounded-lg transition"
          >
            Visit
          </button>
        </div>

        <!-- Featured Cards -->
        <div class="grid md:grid-cols-2 gap-6 text-left max-w-3xl mx-auto">
          <div
            onclick={() => navigateTo('/bobpizzeria')}
            onkeydown={(e) => e.key === 'Enter' && navigateTo('/bobpizzeria')}
            tabindex="0"
            role="button"
            class="p-6 bg-slate-800/60 rounded-xl border border-slate-700 hover:border-amber-500/50 cursor-pointer transition group"
          >
            <div class="text-3xl mb-2">🍕</div>
            <h3 class="text-xl font-bold group-hover:text-amber-400 transition">Bob's Pizzeria</h3>
            <p class="text-sm text-slate-400 mt-1">Crisp wood-fired pizza & artisanal ingredients.</p>
            <span class="inline-block mt-4 text-xs font-semibold text-amber-400">pratello.org/bobpizzeria →</span>
          </div>

          <div
            onclick={() => navigateTo('/deliciadeprazeres')}
            onkeydown={(e) => e.key === 'Enter' && navigateTo('/deliciadeprazeres')}
            tabindex="0"
            role="button"
            class="p-6 bg-slate-800/60 rounded-xl border border-slate-700 hover:border-amber-500/50 cursor-pointer transition group"
          >
            <div class="text-3xl mb-2">🧁</div>
            <h3 class="text-xl font-bold group-hover:text-amber-400 transition">Delícia de Prazeres</h3>
            <p class="text-sm text-slate-400 mt-1">Authentic sweets, pastries, and savory delights.</p>
            <span class="inline-block mt-4 text-xs font-semibold text-amber-400">pratello.org/deliciadeprazeres →</span>
          </div>
        </div>
      </section>
    {:else}
      <!-- Restaurant Page (at /{restaurant}) -->
      <div class="mb-6">
        <button
          onclick={() => navigateTo('/')}
          class="text-sm text-slate-400 hover:text-white flex items-center gap-1 transition"
        >
          ← Back to Pratello Home
        </button>
      </div>

      {#if loading}
        <div class="py-20 text-center text-slate-400 animate-pulse">
          Loading restaurant details...
        </div>
      {:else if error}
        <div class="p-4 bg-red-900/30 border border-red-800 rounded-lg text-red-200">
          Error loading restaurant: {error}
        </div>
      {:else if restaurantData}
        <div class="bg-slate-800/70 border border-slate-700 rounded-2xl p-6 sm:p-8">
          <div class="flex flex-wrap items-center justify-between gap-4 border-b border-slate-700 pb-6 mb-6">
            <div>
              <span class="text-xs uppercase font-bold tracking-wider text-amber-400 bg-amber-400/10 px-2.5 py-1 rounded-md">
                {restaurantData.cuisine}
              </span>
              <h1 class="text-3xl sm:text-4xl font-extrabold mt-3">{restaurantData.name}</h1>
              <p class="text-slate-400 mt-1">{restaurantData.description}</p>
            </div>
            <div class="text-right">
              <span class="text-xs text-slate-400 font-mono">pratello.org/{restaurantData.slug}</span>
            </div>
          </div>

          <!-- Menu items -->
          <h2 class="text-xl font-bold mb-4 flex items-center gap-2">
            <span>📋</span> Featured Menu
          </h2>
          <div class="grid sm:grid-cols-2 gap-3">
            {#each restaurantData.menu_items as item}
              <div class="p-3 bg-slate-900/60 border border-slate-700/60 rounded-lg flex items-center justify-between">
                <span class="font-medium text-slate-200">{item}</span>
                <span class="text-xs font-semibold text-emerald-400 bg-emerald-500/10 px-2 py-0.5 rounded">Available</span>
              </div>
            {/each}
          </div>
        </div>
      {/if}
    {/if}
  </main>

  <!-- Footer -->
  <footer class="border-t border-slate-800 py-6 text-center text-xs text-slate-500">
    Pratello Architecture Demo • Go Standard Library net/http & Svelte
  </footer>
</div>
