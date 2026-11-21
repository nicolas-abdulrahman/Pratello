<script>
  import Header1 from './hero/header1.svelte';
  import { setRestaurantDisplayConfig } from '../lib/stores/restaurant_style';
  import { get_description } from '../lib/stores/Item';
  import CatalogSection from './CatalogSection.svelte';
  import PromotionSection from './PromotionSection.svelte';
  import ItemModal from '../dish/ItemModal.svelte';

  let { slug, name = 'Pratello', coverURL = null, description = null } = $props();
  let items = $state([]);
  let combos = $state([]);
  let combosLoading = $state(true);
  let combosFailed = $state(false);
  let reviews = $state({});
  let banner = $state(null);
  let display = $state(null);
  $effect(() => {
    setRestaurantDisplayConfig(display);
    return () => setRestaurantDisplayConfig(null);
  });
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
      if (!signal.aborted) banner = result?.[0] ?? null;
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

<div data-theme="restaurantTheme" class="min-h-screen bg-background font-sans selection:bg-primary selection:text-white">

  <Header1 {slug} fallbackName={name} fallbackDescription={description} fallbackImage={coverURL} />

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

{#if selected}
  <ItemModal isOpen item={selected} onClose={() => { selectedId = null; }} />
{/if}
