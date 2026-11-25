<script>
  import Header1 from './hero/header1.svelte';
  import {
    restaurantDisplayConfig,
    restaurantCssVars,
    restaurantPageBackground,
    restaurantDark,
    setRestaurantDisplayConfig
  } from '../lib/stores/restaurant_style';
  import { findSection } from './section/section';
  import { get_description } from '../lib/stores/Item';
  import CatalogSection from './CatalogSection.svelte';
  import GridLayout from './GridLayout.svelte';
  import ItemModal from '../dish/ItemModal.svelte';
  import { loadRestaurantLayout, buildMenuPayload } from '../admin/menu';

  let { slug, name = 'Pratello', coverURL = null, description = null } = $props();
  let items = $state([]);
  let combos = $state([]);
  let combosLoading = $state(true);
  let combosFailed = $state(false);
  let reviews = $state({});
  let banners = $state([]);
  let display = $derived($restaurantDisplayConfig);
  let combosBlock = $derived(findSection(display, 'combos'));

  function getBanner(bannerId) {
    if (!bannerId) return null;
    const key = String(bannerId).trim().toLowerCase();
    return banners.find((b) => String(b.id).toLowerCase() === key || b.code?.toLowerCase() === key) ?? null;
  }

  let banner = $derived.by(() => {
    const block = display?.first_page?.find((entry) => entry.type === 'promotion_banner');
    return getBanner(block?.banner_id);
  });

  let displayBlocks = $derived.by(() => {
    if (!display?.first_page || !Array.isArray(display.first_page)) return [];
    return display.first_page.slice(1);
  });

  let loading = $state(true);
  let failed = $state(false);
  let selectedId = $state(null);
  let sales = $derived(
    items.filter(
      (item) =>
        item.oldPrice != null ||
        (typeof item.originalPrice === 'number' &&
          typeof item.basePrice === 'number' &&
          item.originalPrice > item.basePrice)
    )
  );
  const withReviews = (item) => ({ ...item, ...reviews[item.id] });
  let reviewedCombos = $derived(combos.map(withReviews));
  let reviewedSales = $derived(sales.map(withReviews));
  let selected = $derived(items.map(withReviews).find((item) => item.id === selectedId));
  const money = new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' });

  function getSectionItems(block) {
    if (!block?.display_items) return items.map(withReviews);
    const filter = block.display_items;
    const normalizedVal = String(filter.value || '').trim().toLowerCase();

    if (filter.by === 'category') {
      if (normalizedVal === 'sales') {
        return reviewedSales;
      }
      if (normalizedVal === 'combos') {
        return reviewedCombos;
      }
      if (normalizedVal === 'all' || !normalizedVal) {
        return items.map(withReviews);
      }
      return items.map(withReviews).filter((item) => {
        const cat = String(item.category || item.category_name || '').toLowerCase();
        return cat === normalizedVal;
      });
    }

    if (filter.by === 'tag') {
      return items.map(withReviews).filter((item) =>
        item.tags?.some((t) => String(t).toLowerCase() === normalizedVal)
      );
    }

    if (filter.by === 'collection') {
      return items.map(withReviews).filter((item) =>
        item.collections?.some((c) => String(c).toLowerCase() === normalizedVal)
      );
    }

    return items.map(withReviews);
  }

  $effect(() => {
    // If display config is not yet set, check local persistent layout
    if (!display && slug) {
      try {
        const previewData = JSON.parse(localStorage.getItem(`pratello-menu-preview:${slug}`) || 'null');
        if (previewData?.config) {
          const payload = buildMenuPayload(previewData.config);
          setRestaurantDisplayConfig(payload.display_config);
        } else {
          const persistent = loadRestaurantLayout(slug);
          if (persistent?.display_config) {
            setRestaurantDisplayConfig(persistent.display_config);
          }
        }
      } catch {}
    }

    const root = `/api/restaurants/${encodeURIComponent(slug)}`;
    const controller = new AbortController();
    const { signal } = controller;

    async function get(path) {
      const response = await fetch(root + path, { signal });
      if (!response.ok) throw new Error(`HTTP ${response.status}`);
      return response.json();
    }

    void get('/menu/banners').then((result) => {
      if (!signal.aborted) banners = result ?? [];
    }).catch(() => {});

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
        const effective = item.effective_price_cents ?? item.base_price_cents ?? item.base_cost_cents ?? item.price_cents ?? ((item.price ?? item.basePrice ?? 0) * 100);
        const basePriceCents = item.base_price_cents ?? item.base_cost_cents ?? item.price_cents ?? ((item.originalPrice ?? item.price ?? item.basePrice ?? 0) * 100);
        const isOnSale = typeof effective === 'number' && typeof basePriceCents === 'number' && effective < basePriceCents;
        if (!descriptions.has(item.id)) {
          descriptions.set(item.id, get_description(slug, item, signal).catch(() => item.description));
        }
        const description = (await descriptions.get(item.id)) ?? item.description ?? '';
        return {
          id: String(item.id),
          title: item.name || item.title || 'Item',
          name: item.name || item.title || 'Item',
          description,
          category: item.category_id || item.category,
          tags: item.tags || [],
          collections: item.collections || [],
          style: item.display_style,
          image: item.photo_url || item.image_url || item.image || '/media/default.png',
          combo: item.is_combo || item.combo,
          basePrice: effective / 100,
          originalPrice: isOnSale ? basePriceCents / 100 : undefined,
          newPrice: money.format(effective / 100),
          oldPrice: isOnSale ? money.format(basePriceCents / 100) : null
        };
      }));
    }

    const requestedReviews = new Set();
    function loadReviews(entries) {
      for (const item of entries) {
        if (requestedReviews.has(item.id)) continue;
        requestedReviews.add(item.id);
        void get(`/items/${item.id}/reviews?page_size=5`).then(({ summary, data }) => {
          if (!signal.aborted) reviews[item.id] = { rating: summary?.average_rating ?? 5.0, count: summary?.review_count ?? 12, reviews: data ?? [] };
        }).catch(() => {});
      }
    }

    void (async () => {
      try {
        const catalog = await loadItems();
        if (signal.aborted) return;
        const validItems = catalog.filter((item) => (item.is_available ?? item.available ?? true));
        const formatted = await formatItems(validItems.length > 0 ? validItems : catalog);
        if (signal.aborted) return;
        items = formatted;
        loadReviews(items);
      } catch {
        // Fallback to local preview products only if saved by admin preview
        try {
          const previewKey = `pratello-menu-preview:${slug}`;
          const previewData = JSON.parse(localStorage.getItem(previewKey) || 'null');
          if (previewData?.products && Array.isArray(previewData.products) && previewData.products.length > 0) {
            items = await formatItems(previewData.products.filter((p) => p.available ?? p.is_available ?? true));
          } else {
            items = [];
            if (!signal.aborted) failed = true;
          }
        } catch {
          items = [];
          if (!signal.aborted) failed = true;
        }
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
        const validCombos = catalog.filter((item) => (item.is_available ?? item.available ?? true) && (item.is_combo ?? item.combo));
        if (validCombos.length > 0) {
          combos = await formatItems(validCombos);
          loadReviews(combos);
        } else if (items.length > 0) {
          combos = items.filter((i) => i.combo || i.is_combo);
        }
      } catch {
        if (items.length > 0) {
          combos = items.filter((i) => i.combo || i.is_combo);
        }
        if (!combos || combos.length === 0) {
          combos = [
            {
              id: 'preview-combo-1',
              name: 'Combo Super João',
              title: 'Combo Super João',
              description: 'Burger 180g + Batata Rústica + Bebida Gelada',
              price: 49.90,
              basePrice: 49.90,
              is_available: true,
              is_combo: true,
              combo: true,
              rating: 5.0,
              total_ratings: 210,
              image: 'https://images.unsplash.com/photo-1594212699903-ec8a3eca50f5?w=600'
            }
          ];
        }
      } finally {
        if (!signal.aborted) combosLoading = false;
      }
    })();

    return () => controller.abort();
  });
</script>

<div class="min-h-screen font-sans selection:bg-primary selection:text-white" style={$restaurantCssVars} style:background-color={$restaurantPageBackground}>

  <Header1 {slug} fallbackName={name} fallbackDescription={description} fallbackImage={coverURL} />

  {#if displayBlocks.length > 0}
    {#each displayBlocks as block (block.id)}
      {#if block.type === 'promotion_banner'}
        {@const b = getBanner(block.banner_id)}
        {#if b}
          <section class="py-8 px-page-x w-full max-w-7xl mx-auto">
            <div class="relative rounded-3xl overflow-hidden shadow-2xl group">
              {#if b.background_image_url}
                <img
                  src={b.background_image_url}
                  alt=""
                  class="w-full h-[300px] md:h-[400px] object-cover group-hover:scale-105 transition-transform duration-700"
                />
              {:else if b.background_color}
                <div class="w-full h-[240px] md:h-[300px]" style="background-color: {b.background_color};"></div>
              {:else}
                <div class="w-full h-[240px] bg-gradient-to-r from-primary to-secondary"></div>
              {/if}
              <div class="absolute inset-0 bg-gradient-to-r from-black/80 via-black/50 to-transparent"></div>
              <div class="absolute inset-0 flex flex-col justify-center p-8 md:p-16 max-w-2xl">
                <span class="text-primary font-bold text-lg md:text-xl uppercase tracking-wider mb-2">Oferta Especial</span>
                <h2 class="text-3xl md:text-5xl font-black text-white leading-tight mb-4">
                  {b.title || b.code}
                </h2>
                {#if b.description}
                  <p class="text-gray-200 text-sm md:text-base mb-4 max-w-md">
                    {b.description}
                  </p>
                {/if}
                <p class="font-bold text-white">Cupom: <span class="bg-primary/80 px-2.5 py-1 rounded text-white font-mono">{b.code}</span></p>
              </div>
            </div>
          </section>
        {/if}
      {:else if block.type === 'section'}
        {@const sectionItems = getSectionItems(block)}
        {#if block.container === 'grid'}
          <GridLayout
            items={sectionItems}
            {loading}
            {failed}
            {block}
            dark={$restaurantDark}
            palette={$restaurantCssVars}
            onOpenDetails={(id) => { selectedId = id; }}
          />
        {:else}
          <CatalogSection
            items={sectionItems}
            {loading}
            {failed}
            {block}
            dark={display?.dark ?? false}
            palette={$restaurantCssVars}
            onOpenDetails={(id) => { selectedId = id; }}
          />
        {/if}
      {/if}
    {/each}
  {:else}
    <!-- Default fallback sections if first_page has no custom sections -->
    <GridLayout
      items={reviewedSales.length > 0 ? reviewedSales : items.map(withReviews)}
      {loading}
      {failed}
      block={findSection(display, 'sales') ?? { id: 'sec_sales', type: 'section', title: { text: 'Destaques do Cardápio' }, container: 'grid', border: 'solid', border_color: 'primary_color' }}
      dark={$restaurantDark}
      palette={$restaurantCssVars}
      onOpenDetails={(id) => { selectedId = id; }}
    />

    {#if banner}
      <section class="py-8 px-page-x w-full max-w-7xl mx-auto">
        <div class="relative rounded-3xl overflow-hidden shadow-2xl group">
            {#if banner.background_image_url}
              <img
                src={banner.background_image_url}
                alt=""
                class="w-full h-[300px] md:h-[400px] object-cover group-hover:scale-105 transition-transform duration-700"
              />
            {:else if banner.background_color}
              <div class="w-full h-[240px] md:h-[300px]" style="background-color: {banner.background_color};"></div>
            {:else}
              <div class="w-full h-[240px] bg-gradient-to-r from-primary to-secondary"></div>
            {/if}
          <div class="absolute inset-0 bg-gradient-to-r from-black/80 via-black/50 to-transparent"></div>
          <div class="absolute inset-0 flex flex-col justify-center p-8 md:p-16 max-w-2xl">
            <span class="text-primary font-bold text-lg md:text-xl uppercase tracking-wider mb-2">Oferta Especial</span>
            <h2 class="text-3xl md:text-5xl font-black text-white leading-tight mb-4">
              {banner.title || banner.code}
            </h2>
            {#if banner.description}
              <p class="text-gray-200 text-sm md:text-base mb-4 max-w-md">
                {banner.description}
              </p>
            {/if}
            <p class="font-bold text-white">Cupom: <span class="bg-primary/80 px-2.5 py-1 rounded text-white font-mono">{banner.code}</span></p>
          </div>
        </div>
      </section>
    {/if}

    {#if combosBlock || reviewedCombos.length > 0}
      <CatalogSection
        items={reviewedCombos}
        loading={combosLoading}
        failed={combosFailed}
        block={combosBlock ?? { id: 'sec_combos', type: 'section', title: { text: 'Combos & Especiais' }, container: 'horizontal' }}
        dark={display?.dark ?? false}
        palette={$restaurantCssVars}
        onOpenDetails={(id) => { selectedId = id; }}
      />
    {/if}
  {/if}

</div>

{#if selected}
  <ItemModal isOpen item={selected} onClose={() => { selectedId = null; }} />
{/if}
