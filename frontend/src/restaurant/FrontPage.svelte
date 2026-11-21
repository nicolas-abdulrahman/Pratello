<script>
  import { get_description } from '../lib/stores/Item';
  import CatalogSection from './CatalogSection.svelte';
  import PromotionSection from './PromotionSection.svelte';
  import ItemModal from '../dish/ItemModal.svelte';
  import FirstPage from './FirstPage.svelte';

  let { slug, name = 'Pratello', logoURL = null } = $props();
  let items = $state([]);
  let combos = $state([]);
  let combosLoading = $state(true);
  let combosFailed = $state(false);
  let reviews = $state({});
  let banner = $state(null);
  let banners = $state([]);
  let display = $state(null);
  let groups = $state([]);
  let loading = $state(true);
  let failed = $state(false);
  let selectedId = $state(null);
  let sales = $derived(items.filter((item) => item.oldPrice != null));
  const withReviews = (item) => ({ ...item, ...reviews[item.id] });
  let reviewedCombos = $derived(combos.map(withReviews));
  let reviewedSales = $derived(sales.map(withReviews));
  let selected = $derived(items.map(withReviews).find((item) => item.id === selectedId));
  const money = new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' });

  $effect(() => {
    const root = `/api/restaurants/${encodeURIComponent(slug)}`;
    const controller = new AbortController();
    const { signal } = controller;
    async function get(path) {
      const response = await fetch(root + path, { signal });
      if (!response.ok) throw new Error(`HTTP ${response.status}`);
      return response.json();
    }
    void get('/menu/banners').then((result) => {
      if (!signal.aborted) { banners = result ?? []; banner = banners[0] ?? null; }
    }).catch(() => {});
    void get('/menu').then((result) => { if (!signal.aborted) display = result.display_config ?? null; }).catch(() => {});
    async function loadItems(categoryId) {
        const catalog = [];
        for (let page = 1; ; page++) {
          const category = categoryId == null ? '' : `category_id=${encodeURIComponent(categoryId)}&`;
          const result = await get(`/items?${category}page=${page}&page_size=100`);
          const entries = result.data ?? [];
          catalog.push(...entries);
          if (!entries.length || catalog.length >= result.total) break;
        }
        return catalog;
    }

    const descriptions = new Map();
    function formatItems(catalog) {
        return Promise.all(catalog.map(async (item) => {
          const effective = item.effective_price_cents ?? item.base_price_cents;
          if (!descriptions.has(item.id)) {
            descriptions.set(item.id, get_description(slug, item, signal).catch(() => item.description));
          }
          const description = await descriptions.get(item.id);
          return {
             id: item.id, title: item.name, description, category: item.category_id, tags: item.tags, collections: item.collections,
            image: item.photo_url || '/media/default.png', combo: item.is_combo,
            basePrice: effective / 100, originalPrice: item.base_price_cents / 100,
            newPrice: money.format(effective / 100),
            oldPrice: effective < item.base_price_cents ? money.format(item.base_price_cents / 100) : null
          };
        }));
    }

    const requestedReviews = new Set();
    function loadReviews(entries) {
        for (const item of entries) {
          if (requestedReviews.has(item.id)) continue;
          requestedReviews.add(item.id);
          void get(`/items/${item.id}/reviews?page_size=5`).then(({ summary, data }) => {
            if (!signal.aborted) reviews[item.id] = { rating: summary.average_rating, count: summary.review_count, reviews: data };
          }).catch(() => {});
        }
    }

    void (async () => {
      try {
        const catalog = await loadItems();
        if (signal.aborted) return;
         const formatted = await formatItems(catalog.filter((item) => item.is_available));
         if (signal.aborted) return;
         items = formatted;
        loadReviews(items);
      } catch {
        if (!signal.aborted) failed = true;
      } finally {
        if (!signal.aborted) loading = false;
      }
    })();
    void (async () => {
      try {
         const categories = await get('/categories');
         if (!signal.aborted) groups = categories;
        const category = categories.find((entry) => entry.name.trim().toLowerCase() === 'combos');
        const catalog = category ? await loadItems(category.id) : [];
        if (signal.aborted) return;
        const formatted = await formatItems(catalog.filter((item) => item.is_available && item.is_combo));
        if (signal.aborted) return;
        combos = formatted;
        loadReviews(combos);
      } catch {
        if (!signal.aborted) combosFailed = true;
      } finally {
        if (!signal.aborted) combosLoading = false;
      }
    })();
    return () => controller.abort();
  });
</script>

{#if display}
  <FirstPage config={display} {items} categories={groups} {banners} onOpenDetails={(id) => { selectedId = id; }} />
{:else}
<div data-theme="restaurantTheme" class="min-h-screen bg-background font-sans selection:bg-primary selection:text-white">

  <!-- Hero Section (Made Smaller) -->
  <section class="relative pt-20 pb-10 px-page-x overflow-hidden">
    <!-- Background Decorative Shapes -->
    <div class="absolute top-0 right-0 w-[600px] h-[600px] bg-primary rounded-full blur-[150px] opacity-10 -z-10 translate-x-1/3 -translate-y-1/4"></div>
    <div class="absolute bottom-0 left-0 w-[400px] h-[400px] bg-primary rounded-full blur-[100px] opacity-10 -z-10 -translate-x-1/2 translate-y-1/4"></div>

    <div class="max-w-7xl mx-auto flex flex-col lg:flex-row items-center justify-between gap-8">

      <!-- Text Content -->
      <div class="lg:w-1/2 z-10 text-center lg:text-left">
        {#if logoURL}
          <img src={logoURL} alt="" class="mx-auto mb-4 h-20 w-auto max-w-full object-contain lg:mx-0" />
        {/if}
        <span class="text-primary font-bold text-lg tracking-wide uppercase">Bem-vindo ao</span>
        <!-- Reduced text size slightly for a tighter fit -->
        <h1 class="text-6xl md:text-8xl font-black text-gray-900 mt-2 mb-6 tracking-tighter relative">
          {name}
          <!-- Decorative underline/swoosh -->
          <svg class="absolute -bottom-4 left-0 w-full h-8 text-primary opacity-80" viewBox="0 0 300 30" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M5 25C50 10 150 5 295 20" stroke="currentColor" stroke-width="8" stroke-linecap="round"/>
          </svg>
        </h1>
        <p class="text-lg md:text-xl text-muted mb-8 max-w-lg mx-auto lg:mx-0">
          Sabor e qualidade em cada detalhe!
        </p>

        <!-- New Dual Action Buttons -->
        <div class="flex flex-col sm:flex-row items-stretch sm:items-center justify-center lg:justify-start gap-4 sm:gap-5 mt-2">

          <!-- Order in Restaurant -->
          <button class="w-full sm:w-auto min-w-[220px] bg-primary text-white px-8 py-4 rounded-full font-bold text-base flex items-center justify-center gap-2.5 hover:bg-primary-hover transition-all hover:scale-[1.03] shadow-xl shadow-orange-500/30">
            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 2v7c0 1.1.9 2 2 2h4a2 2 0 0 0 2-2V2"/><path d="M7 2v20"/><path d="M21 15V2v0a5 5 0 0 0-5 5v6c0 1.1.9 2 2 2h3Zm0 0v7"/></svg>
            Pedir no Restaurante
          </button>

          <!-- Order Delivery -->
          <button class="w-full sm:w-auto min-w-[220px] bg-white text-primary border-2 border-primary px-8 py-4 rounded-full font-bold text-base flex items-center justify-center gap-2.5 hover:bg-primary-soft transition-all hover:scale-[1.03] shadow-lg">
            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M5 18H3c-.6 0-1-.4-1-1V7c0-.6.4-1 1-1h10c.6 0 1 .4 1 1v11"/><path d="M14 9h4l4 4v4c0 .6-.4 1-1 1h-2"/><circle cx="7" cy="18" r="2"/><circle cx="17" cy="18" r="2"/><path d="M9 18h6"/></svg>
            Delivery
          </button>

        </div>
      </div>

      <!-- Hero Image (Made slightly smaller) -->
      <div class="lg:w-1/2 relative flex justify-center">
        <!-- Floating text element -->
        <div class="absolute top-0 right-0 md:-right-8 z-20 transform rotate-6 bg-white px-4 py-2 rounded-xl shadow-lg border border-orange-100">
           <p class="text-primary font-bold italic font-serif text-base leading-tight">Peça agora<br/>e aproveite!</p>
        </div>

        <!-- Main Burger Image with organic masking -->
        <div class="relative w-[300px] h-[300px] md:w-[450px] md:h-[450px] rounded-full overflow-hidden shadow-2xl border-8 border-white z-10">
           <img
            src="https://images.unsplash.com/photo-1568901346375-23c9450c58cd?ixlib=rb-4.0.3&auto=format&fit=crop&w=1000&q=80"
            alt="Delicious Burger"
            class="w-full h-full object-cover scale-110"
          />
        </div>

        <!-- Decorative dashes behind burger -->
        <div class="absolute -top-4 -left-4 text-primary">
          <svg width="60" height="60" viewBox="0 0 100 100" fill="none" stroke="currentColor" stroke-width="12" stroke-linecap="round">
            <line x1="20" y1="20" x2="40" y2="10"/>
            <line x1="10" y1="40" x2="30" y2="30"/>
            <line x1="20" y1="60" x2="40" y2="50"/>
          </svg>
        </div>
      </div>

    </div>
  </section>

  <PromotionSection items={reviewedSales} {loading} {failed} onOpenDetails={(id) => { selectedId = id; }} />

  <!-- ========================================== -->
  <!-- PROMOTIONAL BANNER                         -->
  <!-- ========================================== -->
  {#if banner}
  <section class="py-8 px-page-x w-full max-w-7xl mx-auto">
    <div class="relative rounded-3xl overflow-hidden shadow-2xl group">
      <!-- Background Image -->
      <img
        src={banner.background_image_url}
        alt=""
        class="w-full h-[300px] md:h-[400px] object-cover group-hover:scale-105 transition-transform duration-700"
      />
      <!-- Dark Gradient Overlay for text readability -->
      <div class="absolute inset-0 bg-gradient-to-r from-black/80 via-black/50 to-transparent"></div>

      <!-- Banner Content -->
      <div class="absolute inset-0 flex flex-col justify-center p-8 md:p-16 max-w-2xl">
        <span class="text-primary font-bold text-lg md:text-xl uppercase tracking-wider mb-2">Oferta Especial</span>
        <h2 class="text-3xl md:text-5xl font-black text-white leading-tight mb-4">
           {banner.title || banner.code}
        </h2>
        <p class="text-gray-200 text-sm md:text-base mb-8 max-w-md">
           {banner.description || ''}
        </p>
         <p class="font-bold text-white">Cupom: {banner.code}</p>
      </div>
    </div>
  </section>
  {/if}

  <CatalogSection title="Combos" items={reviewedCombos} loading={combosLoading} failed={combosFailed} onOpenDetails={(id) => { selectedId = id; }} />

</div>
{/if}

{#if selected}
  <ItemModal isOpen item={selected} onClose={() => { selectedId = null; }} />
{/if}
