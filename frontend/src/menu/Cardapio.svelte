<script>
  import { untrack } from 'svelte';
  import { get_description } from '../lib/stores/Item';
  import MenuCategoryCard from './MenuCategoryCard.svelte';
  import ItemCard from '../restaurant/section/ItemCard.svelte';
  import { resolveItemColors } from '../restaurant/section/section';
  import { restaurantDark } from '../lib/stores/restaurant_style';
  import ItemModal from '../dish/ItemModal.svelte';

  let { slug = '' } = $props();
  let categories = $state([]);
  let openCategory = $state(null);
  let selectedId = $state(null);
  let reviews = $state({});
  let loading = $state(true);
  let failed = $state(false);
  let categoryFailed = $state(false);
  const money = new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' });
  const fallbackImage = '/media/default.png';
  let selectedCategory = $derived(categories.find((category) => category.id === openCategory));
  let container = $state('horizontal');
  let selectedDish = $derived(selectedCategory?.dishes?.find((dish) => dish.id === selectedId));

  async function getJSON(url, signal) {
    const response = await fetch(url, { signal });
    if (!response.ok) throw new Error('Não foi possível carregar os dados.');
    return response.json();
  }

  $effect(() => {
    const root = slug && `/api/restaurants/${encodeURIComponent(slug)}`;
    if (!root) return;
    const controller = new AbortController();
    const { signal } = controller;
    void getJSON(`${root}/menu`, signal).then(({ display_config }) => {
      if (!signal.aborted) container = display_config?.menu.container ?? 'horizontal';
    }).catch(() => {});
    void getJSON(`${root}/categories`, signal).then((groups) => {
      if (!signal.aborted) categories = groups.map((group) => ({ id: group.id, title: group.name, image: group.image_url }));
    }).catch(() => {
      if (!signal.aborted) failed = true;
    }).finally(() => {
      if (!signal.aborted) loading = false;
    });
    return () => controller.abort();
  });

  $effect(() => {
    const categoryId = openCategory;
    if (!slug || categoryId == null) return;
    if (untrack(() => categories.find((group) => group.id === categoryId)?.dishes)) return;
    const root = `/api/restaurants/${encodeURIComponent(slug)}`;
    const controller = new AbortController();
    const { signal } = controller;
    categoryFailed = false;
    async function loadItems() {
      const items = [];
      for (let page = 1; ; page++) {
        const result = await getJSON(`${root}/items?category_id=${categoryId}&page=${page}&page_size=100`, signal);
        items.push(...result.data);
        if (!result.data.length || items.length >= result.total) return items;
      }
    }

    void loadItems().then(async (items) => {
      const dishes = await Promise.all(items.filter((item) => item.is_available).map(async (item) => ({
          id: item.id, title: item.name, description: await get_description(slug, item, signal).catch(() => item.description),
          image: item.photo_url || fallbackImage, combo: item.is_combo,
          basePrice: (item.effective_price_cents ?? item.base_price_cents) / 100,
          price: money.format((item.effective_price_cents ?? item.base_price_cents) / 100)
      })));
      if (!signal.aborted) categories = categories.map((group) => group.id === categoryId ? { ...group, dishes } : group);
    }).catch(() => {
      if (!signal.aborted) categoryFailed = true;
    });
    return () => controller.abort();
  });

  $effect(() => {
    if (!slug || !selectedCategory?.dishes) return;
    const root = `/api/restaurants/${encodeURIComponent(slug)}`;
    const controller = new AbortController();
    const { signal } = controller;
    for (const dish of selectedCategory.dishes) {
      if (untrack(() => reviews[dish.id])) continue;
      void getJSON(`${root}/items/${dish.id}/reviews?page_size=5`, signal).then((result) => {
        if (!signal.aborted) reviews[dish.id] = result;
      }).catch(() => {});
    }
    return () => controller.abort();
  });
</script>

<main class="min-h-screen bg-background text-text">
  <section class="mx-auto max-w-7xl px-space-sm py-space-xl sm:px-space-lg">
    <h1 style:color={`var(--darker_primary_color${$restaurantDark ? '_dark' : ''}, var(--color-primary))`} class="mb-space-lg text-center text-4xl font-black text-primary">Cardápio</h1>
    {#if !slug}
      <p class="text-center text-text-muted">Escolha um restaurante para ver o cardápio. <a href="/restaurantes" class="font-bold text-primary underline">Ver restaurantes</a></p>
    {:else if loading}
      <p role="status" class="text-center text-text-muted">Carregando cardápio…</p>
    {:else if failed}
      <p role="alert" class="text-center text-error">Não foi possível carregar o cardápio.</p>
    {:else if !categories.length}
      <p class="text-center text-text-muted">Nenhum prato disponível no momento.</p>
    {:else}
           <div class={container === 'grid' ? 'grid grid-cols-1 gap-space-md sm:grid-cols-2 lg:grid-cols-3' : 'flex flex-nowrap gap-space-md overflow-x-auto pb-space-xs'}>
        {#each categories as category (category.id)}
          <button type="button" aria-expanded={openCategory === category.id} onclick={() => { selectedId = null; openCategory = openCategory === category.id ? null : category.id; }} class="h-48 w-64 shrink-0 rounded-3xl text-left">
             <MenuCategoryCard title={category.title} image={category.image || category.dishes?.[0]?.image || fallbackImage} description={category.dishes ? `${category.dishes.length} ${category.dishes.length === 1 ? 'prato' : 'pratos'}` : 'Ver pratos'} icon="" />
          </button>
        {/each}
      </div>

      {#if selectedCategory}
        <section class="mt-space-lg rounded-3xl border border-border bg-surface p-space-md">
          <div class="mb-space-md flex items-center justify-between gap-space-sm border-b border-border pb-space-sm">
            <h2 class="text-2xl font-black">{selectedCategory.title}</h2>
            <button type="button" class="text-sm font-bold text-primary" onclick={() => { openCategory = null; selectedId = null; }}>Fechar</button>
          </div>
          {#if categoryFailed && !selectedCategory.dishes}
            <p role="alert" class="text-error">Não foi possível carregar os pratos desta categoria.</p>
          {:else if !selectedCategory.dishes}
            <p role="status" class="text-text-muted">Carregando pratos…</p>
          {:else if !selectedCategory.dishes.length}
            <p class="text-text-muted">Nenhum prato disponível nesta categoria.</p>
          {:else}
          <div class="flex flex-nowrap gap-space-md overflow-x-auto pb-space-xs">
            {#each selectedCategory.dishes as dish (dish.id)}
               <div class="w-80 shrink-0">
                <ItemCard colors={resolveItemColors({}, $restaurantDark, 'bg')}
                  item={{ id: dish.id, name: dish.title, description: dish.description, imageUrl: dish.image, basePrice: dish.basePrice, type: dish.combo ? 'combo' : 'simple', ingredients: [], extras: [], rating: reviews[dish.id]?.summary.average_rating, reviewCount: reviews[dish.id]?.summary.review_count }}
                  onOpenDetails={() => { selectedId = dish.id; }} />
              </div>
            {/each}
          </div>
          {/if}
        </section>
      {/if}
    {/if}
  </section>
</main>

{#if selectedDish}
  <ItemModal isOpen item={{ ...selectedDish, newPrice: selectedDish.price, rating: reviews[selectedDish.id]?.summary.average_rating,
    count: reviews[selectedDish.id]?.summary.review_count, reviews: reviews[selectedDish.id]?.data }} onClose={() => { selectedId = null; }} />
{/if}
