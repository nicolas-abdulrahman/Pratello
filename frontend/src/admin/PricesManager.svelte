<script lang="ts">
  import { onMount } from 'svelte';
  import {
    Tag,
    Image,
    Sparkles,
    Check,
    AlertCircle,
    Plus,
    Calendar,
    X,
    DollarSign,
    Percent,
    Trash2,
    Search,
    Clock,
    Palette,
    Layers,
    Utensils
  } from 'lucide-svelte';
  import type { Product, Category } from './admin';
  import {
    readConfig,
    saveConfig,
    restaurantSlug,
    type MenuConfig,
    type MenuSection
  } from './menuConfig';
  import {
    getPromotionalBanners,
    saveCustomBanner,
    deleteCustomBanner,
    setDishPrice,
    getStoredItemPricing,
    updateStorefrontLayout,
    buildMenuPayload,
    type PromotionalBanner,
    type StorefrontMenuPayload,
    saveRestaurantLayout,
  } from './menu';

  export let slug = '';
  export let products: Product[] = [];
  export let categories: Category[] = [];
  export let onSaveProduct: ((p: any) => Promise<any>) | undefined = undefined;

  let config: MenuConfig = readConfig(slug);
  let banners: PromotionalBanner[] = [];
  let loadingBanners = false;
  let statusMessage = '';
  let statusType: 'success' | 'error' | '' = '';
  let isDeploying = false;

  // Dish filtering
  let dishSearch = '';
  let selectedCategoryFilter = 'all';

  // Dish Price Modal State
  let editingProduct: Product | null = null;
  let editBasePrice: number = 0;
  let editPromoPrice: number = 0;
  let editSaleEnd: string = '';
  let isPromoActive: boolean = false;
  let isSavingPrice: boolean = false;

  // New Banner Modal State
  let showNewBannerModal = false;
  let newBannerTitle = '';
  let newBannerDescription = '';
  let newBannerCode = '';
  let bannerBackgroundType: 'image' | 'color' = 'color';
  let newBannerImageUrl = '';
  let newBannerColor = '#4f46e5';

  const COLOR_PRESETS = [
    { label: 'Indigo', value: '#4f46e5' },
    { label: 'Amber', value: '#d97706' },
    { label: 'Vermelho', value: '#e11d48' },
    { label: 'Esmeralda', value: '#059669' },
    { label: 'Escuro', value: '#18181b' },
    { label: 'Roxo', value: '#7c3aed' },
    { label: 'Azul', value: '#2563eb' },
    { label: 'Laranja', value: '#ea580c' },
  ];

  $: activeSlug = slug || restaurantSlug(config.name) || 'joao-burger';

  $: discountedProducts = products.filter(
    (p) =>
      (p.originalPrice && p.originalPrice > p.price) ||
      (p.effective_price_cents &&
        p.base_price_cents &&
        p.effective_price_cents < p.base_price_cents)
  );

  $: filteredProducts = products.filter((p) => {
    const matchesSearch =
      !dishSearch ||
      p.name.toLowerCase().includes(dishSearch.toLowerCase()) ||
      (p.description && p.description.toLowerCase().includes(dishSearch.toLowerCase()));
    const matchesCategory =
      selectedCategoryFilter === 'all' || p.category === selectedCategoryFilter;
    return matchesSearch && matchesCategory;
  });

  onMount(async () => {
    config = readConfig(activeSlug);
    await fetchBanners();
  });

  async function fetchBanners() {
    if (!activeSlug) return;
    loadingBanners = true;
    try {
      banners = await getPromotionalBanners(activeSlug);
    } catch {
      banners = [];
    } finally {
      loadingBanners = false;
    }
  }

  function ensureSalesSection() {
    const hasPromo = products.some(
      (p) =>
        (p.originalPrice && p.originalPrice > p.price) ||
        (p.effective_price_cents &&
          p.base_price_cents &&
          p.effective_price_cents < p.base_price_cents)
    );
    if (!hasPromo) return;

    const existing = config.sections.find(
      (s) => s.type === 'section' && s.category === 'sales'
    );
    if (!existing) {
      const salesSection: MenuSection = {
        id: `sec_promocoes_${crypto.randomUUID()}`,
        type: 'section',
        name: 'Promoções',
        category: 'sales',
        filter_by: 'category',
        border: 'dotted',
        border_color: 'primary_color',
        layout: 'grid',
        bg: 'bg',
        gradient_enabled: true,
        gradient_first: 'text',
        gradient_second: 'primary_color',
        gradient_end: 5,
        gradient_smooth: false,
        item_style_bg: 'bg',
        item_style_title: 'text',
        item_style_price: 'primary_color',
      };
      config.sections = [salesSection, ...config.sections];
      saveConfig(config, activeSlug);
    }
  }

  function selectProductForPriceEdit(prod: Product) {
    editingProduct = prod;
    const stored = getStoredItemPricing(activeSlug, prod.id);
    const baseInCents =
      prod.base_price_cents ??
      (prod.originalPrice ? Math.round(prod.originalPrice * 100) : Math.round(prod.price * 100));
    const effectiveInCents = prod.effective_price_cents ?? Math.round(prod.price * 100);

    editBasePrice = (stored?.base_price_cents ?? baseInCents) / 100;
    editPromoPrice = (stored?.effective_price_cents ?? effectiveInCents) / 100;
    editSaleEnd = prod.saleEnd ?? prod.sale_ends_at ?? stored?.saleEnd ?? stored?.sale_ends_at ?? '';
    isPromoActive = editPromoPrice < editBasePrice;
  }

  function applyDiscountPct(pct: number) {
    if (editBasePrice <= 0) return;
    const discounted = editBasePrice * (1 - pct / 100);
    editPromoPrice = Number(discounted.toFixed(2));
    isPromoActive = true;
  }

  function clearPromo() {
    editPromoPrice = editBasePrice;
    isPromoActive = false;
    editSaleEnd = '';
  }

  async function handleSavePrice() {
    if (!editingProduct) return;
    isSavingPrice = true;
    try {
      const baseCents = Math.round(editBasePrice * 100);
      const effectiveCents = isPromoActive ? Math.round(editPromoPrice * 100) : baseCents;

      const priceUpdate = {
        base_price_cents: baseCents,
        effective_price_cents: effectiveCents,
        saleEnd: isPromoActive && editSaleEnd ? editSaleEnd : null,
        sale_ends_at: isPromoActive && editSaleEnd ? editSaleEnd : null,
      };

      // Store via menu.ts
      await setDishPrice(activeSlug, editingProduct.id, priceUpdate);

      // Update local product reference
      const updatedProduct: Product = {
        ...editingProduct,
        price: effectiveCents / 100,
        originalPrice: isPromoActive ? baseCents / 100 : undefined,
        base_price_cents: baseCents,
        effective_price_cents: effectiveCents,
        saleEnd: isPromoActive && editSaleEnd ? editSaleEnd : undefined,
        sale_ends_at: isPromoActive && editSaleEnd ? editSaleEnd : undefined,
      };

      products = products.map((p) => (p.id === editingProduct!.id ? updatedProduct : p));

      if (onSaveProduct) {
        try {
          await onSaveProduct(updatedProduct);
        } catch (e) {
          console.warn('onSaveProduct sync warning:', e);
        }
      }

      // Automatically keep sales section in layout if promos exist
      ensureSalesSection();

      statusType = 'success';
      statusMessage = `Preço do prato "${updatedProduct.name}" atualizado com sucesso!`;
      editingProduct = null;
    } catch (err: any) {
      statusType = 'error';
      statusMessage = err?.message || 'Falha ao salvar preço.';
    } finally {
      isSavingPrice = false;
    }
  }

  function handleCreateBanner() {
    if (!newBannerTitle.trim() && !newBannerCode.trim()) {
      statusType = 'error';
      statusMessage = 'Informe ao menos um título ou código para o banner.';
      return;
    }

    const code =
      newBannerCode.trim() ||
      newBannerTitle
        .trim()
        .toUpperCase()
        .replace(/[^A-Z0-9]/g, '_')
        .slice(0, 10) ||
      'PROMO';

    const newBanner: PromotionalBanner = {
      id: `banner_${crypto.randomUUID().slice(0, 8)}`,
      code,
      title: newBannerTitle.trim(),
      description: newBannerDescription.trim(),
      background_image_url:
        bannerBackgroundType === 'image' && newBannerImageUrl.trim()
          ? newBannerImageUrl.trim()
          : undefined,
      background_color:
        bannerBackgroundType === 'color' ? newBannerColor : undefined,
    };

    // Store in menu.ts
    banners = saveCustomBanner(activeSlug, newBanner);

    // Automatically add banner section to cardapio sections if not already present
    const existingSection = config.sections.find(
      (s) =>
        s.type === 'promotion_banner' &&
        (s.banner_id === newBanner.id || s.banner_id === newBanner.code)
    );
    if (!existingSection) {
      const newSection: MenuSection = {
        id: `promo_${newBanner.id}`,
        type: 'promotion_banner',
        name: newBanner.title || `Banner Promocional (${newBanner.code})`,
        category: newBanner.code,
        banner_id: newBanner.code || newBanner.id,
        border: 'none',
        layout: 'grid',
      };
      config.sections = [...config.sections, newSection];
      saveConfig(config, activeSlug);
    }

    statusType = 'success';
    statusMessage = `Banner "${newBanner.title || newBanner.code}" cadastrado e adicionado ao cardápio!`;

    // Reset modal
    showNewBannerModal = false;
    newBannerTitle = '';
    newBannerDescription = '';
    newBannerCode = '';
    newBannerImageUrl = '';
    newBannerColor = '#4f46e5';
  }

  function handleDeleteBanner(banner: PromotionalBanner) {
    const bannerKey = banner.id || banner.code;
    banners = deleteCustomBanner(activeSlug, bannerKey);
    config.sections = config.sections.filter(
      (s) =>
        !(
          s.type === 'promotion_banner' &&
          (s.banner_id === banner.id ||
            s.banner_id === banner.code ||
            s.category === banner.code)
        )
    );
    saveConfig(config, activeSlug);
    statusType = 'success';
    statusMessage = `Banner "${banner.title || banner.code}" removido.`;
  }

  function formatSaleEnd(dateStr?: string | null) {
    if (!dateStr) return null;
    try {
      const d = new Date(dateStr);
      if (isNaN(d.getTime())) return dateStr;
      return d.toLocaleString('pt-BR', {
        day: '2-digit',
        month: '2-digit',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
      });
    } catch {
      return dateStr;
    }
  }

  async function deployTest() {
    isDeploying = true;
    statusMessage = '';
    try {
      const payload: StorefrontMenuPayload = buildMenuPayload(config);
      saveRestaurantLayout(activeSlug, payload);
      localStorage.setItem(
        `pratello-menu-preview:${activeSlug}`,
        JSON.stringify({ config, products })
      );
      await updateStorefrontLayout(activeSlug, payload);

      statusType = 'success';
      statusMessage = 'Deploy de teste realizado com sucesso na API!';
      window.open(`/me/my-restaurant/${activeSlug}`, '_blank', 'noopener');
    } catch (err: any) {
      statusType = 'error';
      statusMessage = err?.message || 'Falha ao realizar deploy de teste na API.';
    } finally {
      isDeploying = false;
    }
  }
</script>

<div class="space-y-6">
  <!-- Header -->
  <div class="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
    <div>
      <h1 class="text-2xl font-bold text-zinc-900">Preços & Promoções</h1>
      <p class="mt-1 text-sm text-zinc-600">
        Clique em um prato para editar seu preço e prazo de promoção, veja itens com desconto e gerencie banners.
      </p>
    </div>
    <div class="flex items-center gap-3">
      <button
        type="button"
        class="inline-flex items-center gap-2 rounded-xl bg-yellow-400 px-4 py-2 font-semibold text-zinc-900 shadow-sm transition hover:bg-yellow-300 disabled:opacity-50 cursor-pointer"
        on:click={deployTest}
        disabled={isDeploying}
      >
        <Sparkles size={16} />
        {isDeploying ? 'Enviando...' : 'Deploy test'}
      </button>
    </div>
  </div>

  <!-- Status Alert -->
  {#if statusMessage}
    <div
      role="alert"
      class="flex items-center justify-between gap-2 rounded-xl p-4 text-sm font-medium {statusType === 'success'
        ? 'bg-emerald-50 text-emerald-800 border border-emerald-200'
        : 'bg-rose-50 text-rose-800 border border-rose-200'}"
    >
      <div class="flex items-center gap-2">
        {#if statusType === 'success'}
          <Check size={18} class="text-emerald-600 shrink-0" />
        {:else}
          <AlertCircle size={18} class="text-rose-600 shrink-0" />
        {/if}
        <span>{statusMessage}</span>
      </div>
      <button
        type="button"
        class="text-zinc-400 hover:text-zinc-600"
        on:click={() => (statusMessage = '')}
      >
        <X size={16} />
      </button>
    </div>
  {/if}

  <!-- SEÇÃO 1: Pratos (Clique para alterar o preço e definir promoção) -->
  <div class="rounded-2xl border border-zinc-200 bg-white p-5 shadow-sm space-y-4">
    <div class="flex flex-col justify-between gap-3 sm:flex-row sm:items-center">
      <div>
        <h2 class="text-lg font-bold text-zinc-900 flex items-center gap-2">
          <Utensils class="text-amber-500" size={20} />
          Pratos & Definição de Preços
        </h2>
        <p class="mt-0.5 text-xs text-zinc-500">
          Clique em qualquer prato cadastrado em "Pratos" para alterar o preço base, definir preço promocional e agendar término da oferta.
        </p>
      </div>

      <!-- Filtros rápidos -->
      <div class="flex flex-wrap items-center gap-2">
        <div class="relative">
          <Search size={14} class="absolute left-3 top-1/2 -translate-y-1/2 text-zinc-400" />
          <input
            type="text"
            placeholder="Buscar prato..."
            bind:value={dishSearch}
            class="rounded-lg border border-zinc-200 bg-zinc-50 pl-8 pr-3 py-1.5 text-xs text-zinc-900 focus:border-amber-400 focus:bg-white focus:outline-none"
          />
        </div>
        {#if categories.length > 0}
          <select
            bind:value={selectedCategoryFilter}
            class="rounded-lg border border-zinc-200 bg-zinc-50 px-2.5 py-1.5 text-xs text-zinc-800 focus:border-amber-400 focus:outline-none"
          >
            <option value="all">Todas categorias</option>
            {#each categories as cat}
              <option value={cat.id}>{cat.name}</option>
            {/each}
          </select>
        {/if}
      </div>
    </div>

    <!-- Lista de Pratos Clicáveis -->
    {#if filteredProducts.length > 0}
      <div class="grid gap-3 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
        {#each filteredProducts as prod}
          {@const isPromo =
            (prod.originalPrice && prod.originalPrice > prod.price) ||
            (prod.effective_price_cents &&
              prod.base_price_cents &&
              prod.effective_price_cents < prod.base_price_cents)}
          <button
            type="button"
            class="group relative flex flex-col justify-between text-left rounded-xl border p-3 transition text-zinc-900 cursor-pointer {isPromo
              ? 'border-amber-300 bg-amber-50/40 hover:border-amber-400 hover:shadow-md'
              : 'border-zinc-200 bg-white hover:border-zinc-300 hover:shadow-md'}"
            on:click={() => selectProductForPriceEdit(prod)}
          >
            <div class="flex items-start gap-3">
              {#if prod.image}
                <img
                  src={prod.image}
                  alt={prod.name}
                  class="h-12 w-12 rounded-lg object-cover shrink-0 border border-zinc-100"
                />
              {:else}
                <div class="flex h-12 w-12 items-center justify-center rounded-lg bg-zinc-100 text-zinc-400 shrink-0">
                  <Utensils size={18} />
                </div>
              {/if}
              <div class="min-w-0 flex-1">
                <p class="font-semibold text-sm text-zinc-900 truncate group-hover:text-amber-600 transition">
                  {prod.name}
                </p>
                {#if prod.description}
                  <p class="text-xs text-zinc-500 line-clamp-1 mt-0.5">{prod.description}</p>
                {/if}
                <div class="mt-2 flex items-center gap-2 text-xs">
                  {#if isPromo}
                    <span class="text-zinc-400 line-through">
                      R$ {(prod.originalPrice ?? ((prod.base_price_cents ?? 0) / 100)).toFixed(2)}
                    </span>
                    <span class="font-bold text-emerald-600">
                      R$ {(prod.price ?? ((prod.effective_price_cents ?? 0) / 100)).toFixed(2)}
                    </span>
                  {:else}
                    <span class="font-semibold text-zinc-800">
                      R$ {Number(prod.price ?? 0).toFixed(2)}
                    </span>
                  {/if}
                </div>
              </div>
            </div>

            <div class="mt-3 pt-2 border-t border-zinc-100 flex items-center justify-between text-[11px]">
              {#if isPromo}
                <span class="inline-flex items-center gap-1 rounded bg-amber-100 px-1.5 py-0.5 font-bold text-amber-800">
                  <Percent size={10} /> Em Promoção
                </span>
                {#if prod.saleEnd || prod.sale_ends_at}
                  <span class="text-zinc-500 flex items-center gap-1" title={prod.saleEnd || prod.sale_ends_at}>
                    <Clock size={11} /> Até {formatSaleEnd(prod.saleEnd || prod.sale_ends_at)}
                  </span>
                {/if}
              {:else}
                <span class="text-zinc-400 group-hover:text-amber-600 transition flex items-center gap-1">
                  <DollarSign size={11} /> Clique para alterar preço
                </span>
              {/if}
            </div>
          </button>
        {/each}
      </div>
    {:else}
      <div class="rounded-xl border border-dashed border-zinc-200 p-8 text-center text-xs text-zinc-400">
        Nenhum prato encontrado com os filtros aplicados.
      </div>
    {/if}
  </div>

  <!-- SEÇÃO 2: Pratos Atualmente em Promoção (Sales) -->
  <div class="rounded-2xl border border-zinc-200 bg-white p-5 shadow-sm space-y-4">
    <div class="flex flex-col justify-between gap-3 sm:flex-row sm:items-center">
      <div>
        <h2 class="text-lg font-bold text-zinc-900 flex items-center gap-2">
          <Tag class="text-amber-500" size={20} />
          Pratos em Oferta ({discountedProducts.length})
        </h2>
        <p class="mt-0.5 text-xs text-zinc-500">
          Itens com preço promocional ativo. Adicionados automaticamente na seção "Promoções" do cardápio público.
        </p>
      </div>
    </div>

    {#if discountedProducts.length > 0}
      <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
        {#each discountedProducts as item}
          {@const orig = item.originalPrice ?? ((item.base_price_cents ?? 0) / 100)}
          {@const curr = item.price ?? ((item.effective_price_cents ?? 0) / 100)}
          {@const discountPct = orig > curr ? Math.round(((orig - curr) / orig) * 100) : 0}
          <div class="relative flex flex-col justify-between rounded-xl border border-amber-200 bg-amber-50/50 p-4">
            <div class="flex items-start justify-between gap-3">
              <div class="flex items-start gap-3 min-w-0">
                {#if item.image}
                  <img
                    src={item.image}
                    alt={item.name}
                    class="h-12 w-12 rounded-lg object-cover shrink-0 border border-amber-100"
                  />
                {/if}
                <div class="min-w-0">
                  <p class="font-bold text-zinc-900 truncate">{item.name}</p>
                  <div class="mt-1 flex items-baseline gap-2 text-xs">
                    <span class="text-zinc-400 line-through">
                      R$ {orig.toFixed(2)}
                    </span>
                    <span class="font-extrabold text-emerald-600 text-sm">
                      R$ {curr.toFixed(2)}
                    </span>
                  </div>
                </div>
              </div>
              <span class="rounded-lg bg-amber-100 px-2 py-1 text-xs font-black text-amber-800 shrink-0">
                -{discountPct}%
              </span>
            </div>

            {#if item.saleEnd || item.sale_ends_at}
              <div class="mt-3 flex items-center gap-1 text-[11px] text-zinc-500">
                <Clock size={12} class="text-amber-600 shrink-0" />
                <span>Oferta válida até: <strong class="text-zinc-700">{formatSaleEnd(item.saleEnd || item.sale_ends_at)}</strong></span>
              </div>
            {/if}

            <div class="mt-3 pt-2.5 border-t border-amber-200/60 flex justify-end">
              <button
                type="button"
                class="text-xs font-semibold text-amber-700 hover:text-amber-900 underline cursor-pointer"
                on:click={() => selectProductForPriceEdit(item)}
              >
                Editar Preço / Promoção
              </button>
            </div>
          </div>
        {/each}
      </div>
    {:else}
      <div class="rounded-xl border border-dashed border-zinc-200 p-6 text-center text-xs text-zinc-400">
        Nenhum prato com desconto no momento. Clique em um prato acima para definir um novo preço promocional!
      </div>
    {/if}
  </div>

  <!-- SEÇÃO 3: Banners Promocionais Cadastrados -->
  <div class="rounded-2xl border border-zinc-200 bg-white p-5 shadow-sm space-y-4">
    <div class="flex flex-col justify-between gap-3 sm:flex-row sm:items-center">
      <div>
        <h2 class="text-lg font-bold text-zinc-900 flex items-center gap-2">
          <Image class="text-indigo-500" size={20} />
          Banners Promocionais Cadastrados
        </h2>
        <p class="mt-0.5 text-xs text-zinc-500">
          Todos os banners cadastrados para este restaurante. Aparecem no topo do cardápio e no layout do storefront.
        </p>
      </div>

      <div class="flex items-center gap-2">
        <button
          type="button"
          class="text-xs font-medium text-zinc-600 underline hover:text-zinc-900 cursor-pointer mr-2"
          on:click={fetchBanners}
          disabled={loadingBanners}
        >
          {loadingBanners ? 'Recarregando...' : 'Recarregar'}
        </button>

        <button
          type="button"
          class="inline-flex items-center gap-1.5 rounded-xl bg-indigo-600 px-3.5 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-700 transition cursor-pointer"
          on:click={() => (showNewBannerModal = true)}
        >
          <Plus size={16} />
          + Novo Banner
        </button>
      </div>
    </div>

    {#if loadingBanners}
      <p class="text-sm text-zinc-500 py-4">Carregando banners...</p>
    {:else if banners.length > 0}
      <div class="grid gap-4 sm:grid-cols-2">
        {#each banners as banner}
          <div class="relative overflow-hidden rounded-xl border border-zinc-200 bg-white shadow-xs flex flex-col justify-between">
            <!-- Banner Visual Preview -->
            {#if banner.background_image_url}
              <div class="h-28 w-full overflow-hidden bg-zinc-100 relative">
                <img
                  src={banner.background_image_url}
                  alt={banner.title}
                  class="h-full w-full object-cover"
                />
                <div class="absolute inset-0 bg-gradient-to-t from-black/70 via-black/20 to-transparent"></div>
                <div class="absolute bottom-2 left-3 right-3 text-white">
                  <span class="inline-block rounded bg-indigo-500/90 px-1.5 py-0.5 text-[10px] font-mono font-bold uppercase tracking-wider text-white">
                    {banner.code || banner.id}
                  </span>
                  <h4 class="font-bold text-sm text-white drop-shadow-sm truncate">{banner.title || 'Sem título'}</h4>
                </div>
              </div>
            {:else if banner.background_color}
              <div
                class="h-28 w-full p-3 flex flex-col justify-end text-white relative shadow-inner"
                style="background-color: {banner.background_color};"
              >
                <div class="absolute inset-0 bg-black/10"></div>
                <div class="relative z-10">
                  <span class="inline-block rounded bg-white/20 px-1.5 py-0.5 text-[10px] font-mono font-bold uppercase tracking-wider text-white">
                    {banner.code || banner.id} (Cor sólida)
                  </span>
                  <h4 class="font-bold text-sm text-white drop-shadow-sm truncate">{banner.title || 'Sem título'}</h4>
                </div>
              </div>
            {:else}
              <div class="h-20 w-full bg-gradient-to-r from-indigo-500 to-purple-600 p-3 flex flex-col justify-end text-white">
                <span class="inline-block rounded bg-white/20 px-1.5 py-0.5 text-[10px] font-mono font-bold uppercase tracking-wider text-white">
                  {banner.code || banner.id}
                </span>
                <h4 class="font-bold text-sm text-white truncate">{banner.title || 'Sem título'}</h4>
              </div>
            {/if}

            <div class="p-3 flex items-start justify-between gap-2">
              <div class="flex-1 min-w-0">
                <p class="text-xs text-zinc-600 line-clamp-2">
                  {banner.description || 'Sem descrição cadastrada.'}
                </p>
                {#if banner.background_color}
                  <div class="mt-1.5 flex items-center gap-1.5 text-[11px] text-zinc-400">
                    <span
                      class="h-3 w-3 rounded-full border border-zinc-300"
                      style="background-color: {banner.background_color};"
                    ></span>
                    <span>Cor: {banner.background_color}</span>
                  </div>
                {/if}
              </div>

              <button
                type="button"
                title="Remover banner"
                class="text-zinc-400 hover:text-rose-600 p-1 rounded-lg hover:bg-rose-50 transition cursor-pointer"
                on:click={() => handleDeleteBanner(banner)}
              >
                <Trash2 size={16} />
              </button>
            </div>
          </div>
        {/each}
      </div>
    {:else}
      <div class="rounded-xl border border-dashed border-zinc-200 p-8 text-center text-xs text-zinc-400">
        Nenhum banner cadastrado. Clique no botão <strong>"+ Novo Banner"</strong> acima para criar um banner com imagem ou cor sólida!
      </div>
    {/if}
  </div>
</div>

<!-- MODAL: Editar Preço & Promoção de Prato -->
{#if editingProduct}
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
    <div
      role="dialog"
      aria-modal="true"
      class="w-full max-w-lg rounded-2xl bg-white p-6 shadow-xl space-y-5 animate-in fade-in"
    >
      <div class="flex items-start justify-between">
        <div class="flex items-center gap-3">
          {#if editingProduct.image}
            <img
              src={editingProduct.image}
              alt={editingProduct.name}
              class="h-12 w-12 rounded-xl object-cover border border-zinc-200 shrink-0"
            />
          {:else}
            <div class="flex h-12 w-12 items-center justify-center rounded-xl bg-amber-50 text-amber-600 shrink-0">
              <Utensils size={20} />
            </div>
          {/if}
          <div>
            <h3 class="text-lg font-bold text-zinc-900">{editingProduct.name}</h3>
            <p class="text-xs text-zinc-500">Alterar preço e configurar promoção do prato</p>
          </div>
        </div>
        <button
          type="button"
          class="rounded-lg p-1 text-zinc-400 hover:bg-zinc-100 hover:text-zinc-700 cursor-pointer"
          on:click={() => (editingProduct = null)}
        >
          <X size={18} />
        </button>
      </div>

      <div class="space-y-4 text-sm">
        <!-- Preço Base / Normal -->
        <div>
          <label for="edit-base-price" class="block font-medium text-xs text-zinc-700 mb-1">
            Preço Normal / Preço Base ("De") (R$)
          </label>
          <div class="relative">
            <span class="absolute left-3 top-1/2 -translate-y-1/2 font-semibold text-zinc-400 text-xs">R$</span>
            <input
              id="edit-base-price"
              type="number"
              step="0.01"
              min="0"
              bind:value={editBasePrice}
              class="w-full rounded-xl border border-zinc-300 py-2 pl-9 pr-3 text-sm font-semibold text-zinc-900 focus:border-amber-500 focus:outline-none"
            />
          </div>
        </div>

        <!-- Preço Promocional -->
        <div class="rounded-xl border border-amber-200 bg-amber-50/40 p-4 space-y-3">
          <div class="flex items-center justify-between">
            <span class="font-bold text-xs text-zinc-900 flex items-center gap-1.5">
              <Tag size={14} class="text-amber-600" />
              Preço Promocional / Com Desconto ("Por") (R$)
            </span>
            {#if isPromoActive}
              <span class="rounded bg-emerald-100 px-2 py-0.5 text-[11px] font-bold text-emerald-800">
                Desconto Ativo
              </span>
            {/if}
          </div>

          <div class="relative">
            <span class="absolute left-3 top-1/2 -translate-y-1/2 font-semibold text-zinc-400 text-xs">R$</span>
            <input
              type="number"
              step="0.01"
              min="0"
              bind:value={editPromoPrice}
              on:input={() => {
                isPromoActive = editPromoPrice < editBasePrice;
              }}
              class="w-full rounded-xl border border-zinc-300 bg-white py-2 pl-9 pr-3 text-sm font-bold text-emerald-700 focus:border-amber-500 focus:outline-none"
            />
          </div>

          <!-- Atalhos rápidos de desconto -->
          <div class="flex items-center gap-2 pt-1">
            <span class="text-xs text-zinc-500">Descontos rápidos:</span>
            <button
              type="button"
              class="rounded-lg border border-amber-300 bg-white px-2 py-1 text-xs font-semibold text-amber-800 hover:bg-amber-100 cursor-pointer"
              on:click={() => applyDiscountPct(10)}
            >
              -10%
            </button>
            <button
              type="button"
              class="rounded-lg border border-amber-300 bg-white px-2 py-1 text-xs font-semibold text-amber-800 hover:bg-amber-100 cursor-pointer"
              on:click={() => applyDiscountPct(20)}
            >
              -20%
            </button>
            <button
              type="button"
              class="rounded-lg border border-amber-300 bg-white px-2 py-1 text-xs font-semibold text-amber-800 hover:bg-amber-100 cursor-pointer"
              on:click={() => applyDiscountPct(30)}
            >
              -30%
            </button>
            {#if isPromoActive}
              <button
                type="button"
                class="ml-auto text-xs font-semibold text-rose-600 hover:underline cursor-pointer"
                on:click={clearPromo}
              >
                Remover Desconto
              </button>
            {/if}
          </div>

          <!-- Data/Hora de Término da Oferta -->
          <div class="pt-2 border-t border-amber-200/60">
            <label for="edit-sale-end" class="block font-medium text-xs text-zinc-700 mb-1 flex items-center gap-1.5">
              <Calendar size={13} class="text-zinc-500" />
              Término da Promoção (Sale End)
            </label>
            <input
              id="edit-sale-end"
              type="datetime-local"
              bind:value={editSaleEnd}
              class="w-full rounded-lg border border-zinc-300 bg-white px-3 py-1.5 text-xs text-zinc-800 focus:border-amber-500 focus:outline-none"
            />
            <p class="mt-1 text-[11px] text-zinc-500">
              Opcional. Exibe até quando o preço promocional será válido no storefront.
            </p>
          </div>
        </div>
      </div>

      <div class="flex items-center justify-end gap-3 pt-2">
        <button
          type="button"
          class="rounded-xl border border-zinc-300 px-4 py-2 text-sm font-semibold text-zinc-700 hover:bg-zinc-50 cursor-pointer"
          on:click={() => (editingProduct = null)}
        >
          Cancelar
        </button>
        <button
          type="button"
          class="inline-flex items-center gap-1.5 rounded-xl bg-amber-500 px-4 py-2 text-sm font-semibold text-zinc-950 shadow-sm hover:bg-amber-400 disabled:opacity-50 cursor-pointer"
          disabled={isSavingPrice}
          on:click={handleSavePrice}
        >
          <Check size={16} />
          {isSavingPrice ? 'Salvando...' : 'Salvar Preço'}
        </button>
      </div>
    </div>
  </div>
{/if}

<!-- MODAL: + Novo Banner Promocional -->
{#if showNewBannerModal}
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
    <div
      role="dialog"
      aria-modal="true"
      class="w-full max-w-lg rounded-2xl bg-white p-6 shadow-xl space-y-5 animate-in fade-in"
    >
      <div class="flex items-start justify-between">
        <div>
          <h3 class="text-lg font-bold text-zinc-900 flex items-center gap-2">
            <Image class="text-indigo-600" size={20} />
            Cadastrar Novo Banner
          </h3>
          <p class="text-xs text-zinc-500">
            Crie um banner promocional com imagem ou com uma cor sólida.
          </p>
        </div>
        <button
          type="button"
          class="rounded-lg p-1 text-zinc-400 hover:bg-zinc-100 hover:text-zinc-700 cursor-pointer"
          on:click={() => (showNewBannerModal = false)}
        >
          <X size={18} />
        </button>
      </div>

      <div class="space-y-4 text-sm">
        <!-- Título -->
        <div>
          <label for="new-banner-title" class="block font-medium text-xs text-zinc-700 mb-1">
            Título do Banner *
          </label>
          <input
            id="new-banner-title"
            type="text"
            placeholder="Ex: Oferta Especial da Semana"
            bind:value={newBannerTitle}
            class="w-full rounded-xl border border-zinc-300 p-2.5 text-sm text-zinc-900 focus:border-indigo-500 focus:outline-none"
          />
        </div>

        <!-- Descrição -->
        <div>
          <label for="new-banner-desc" class="block font-medium text-xs text-zinc-700 mb-1">
            Descrição do Banner
          </label>
          <textarea
            id="new-banner-desc"
            rows="2"
            placeholder="Ex: Peça o Combo Duplo e ganhe refrigerante grátis."
            bind:value={newBannerDescription}
            class="w-full rounded-xl border border-zinc-300 p-2.5 text-sm text-zinc-900 focus:border-indigo-500 focus:outline-none"
          ></textarea>
        </div>

        <!-- Código / Cupom -->
        <div>
          <label for="new-banner-code" class="block font-medium text-xs text-zinc-700 mb-1">
            Código do Cupom / ID (Opcional)
          </label>
          <input
            id="new-banner-code"
            type="text"
            placeholder="Ex: DUPLA8 ou PROMO10"
            bind:value={newBannerCode}
            class="w-full rounded-xl border border-zinc-300 p-2.5 text-sm font-mono text-zinc-900 focus:border-indigo-500 focus:outline-none uppercase"
          />
        </div>

        <!-- Tipo de Fundo: Imagem ou Cor Sólida -->
        <div>
          <span class="block font-medium text-xs text-zinc-700 mb-2">
            Estilo de Fundo do Banner
          </span>
          <div class="grid grid-cols-2 gap-2">
            <button
              type="button"
              class="flex items-center justify-center gap-2 rounded-xl border p-2.5 text-xs font-semibold transition cursor-pointer {bannerBackgroundType === 'color'
                ? 'border-indigo-600 bg-indigo-50 text-indigo-700'
                : 'border-zinc-200 bg-zinc-50 text-zinc-600 hover:bg-zinc-100'}"
              on:click={() => (bannerBackgroundType = 'color')}
            >
              <Palette size={15} />
              Cor Sólida (Flat Color)
            </button>
            <button
              type="button"
              class="flex items-center justify-center gap-2 rounded-xl border p-2.5 text-xs font-semibold transition cursor-pointer {bannerBackgroundType === 'image'
                ? 'border-indigo-600 bg-indigo-50 text-indigo-700'
                : 'border-zinc-200 bg-zinc-50 text-zinc-600 hover:bg-zinc-100'}"
              on:click={() => (bannerBackgroundType = 'image')}
            >
              <Image size={15} />
              Imagem (URL)
            </button>
          </div>
        </div>

        <!-- Opções dependendo do tipo -->
        {#if bannerBackgroundType === 'color'}
          <div class="rounded-xl border border-zinc-200 bg-zinc-50 p-3.5 space-y-3">
            <div class="flex items-center justify-between">
              <label for="banner-color-input" class="text-xs font-medium text-zinc-700">
                Selecione a cor sólida:
              </label>
              <div class="flex items-center gap-2">
                <input
                  id="banner-color-input"
                  type="color"
                  bind:value={newBannerColor}
                  class="h-8 w-10 cursor-pointer rounded border border-zinc-300 p-0.5 bg-white"
                />
                <span class="text-xs font-mono font-semibold text-zinc-700">{newBannerColor}</span>
              </div>
            </div>

            <!-- Paleta de cores prontas -->
            <div class="flex flex-wrap items-center gap-2">
              {#each COLOR_PRESETS as preset}
                <button
                  type="button"
                  title={preset.label}
                  class="h-7 w-7 rounded-full border-2 transition cursor-pointer {newBannerColor === preset.value
                    ? 'border-zinc-900 scale-110 shadow-sm'
                    : 'border-white hover:scale-105'}"
                  style="background-color: {preset.value};"
                  on:click={() => (newBannerColor = preset.value)}
                ></button>
              {/each}
            </div>
          </div>
        {:else}
          <div>
            <label for="new-banner-img" class="block font-medium text-xs text-zinc-700 mb-1">
              URL da Imagem
            </label>
            <input
              id="new-banner-img"
              type="url"
              placeholder="https://exemplo.com/banner.jpg"
              bind:value={newBannerImageUrl}
              class="w-full rounded-xl border border-zinc-300 p-2.5 text-sm text-zinc-900 focus:border-indigo-500 focus:outline-none"
            />
          </div>
        {/if}

        <!-- Pré-visualização do Banner -->
        <div>
          <span class="block font-medium text-xs text-zinc-500 mb-1">Pré-visualização do Banner:</span>
          <div class="relative overflow-hidden rounded-xl shadow-md">
            {#if bannerBackgroundType === 'image' && newBannerImageUrl}
              <div class="h-28 w-full relative">
                <img
                  src={newBannerImageUrl}
                  alt="Prévia"
                  class="h-full w-full object-cover"
                />
                <div class="absolute inset-0 bg-gradient-to-t from-black/80 via-black/30 to-transparent"></div>
              </div>
            {:else}
              <div
                class="h-28 w-full flex flex-col justify-end p-4 text-white"
                style="background-color: {newBannerColor};"
              >
              </div>
            {/if}
            <div class="absolute bottom-3 left-3 right-3 text-white pointer-events-none drop-shadow">
              {#if newBannerCode}
                <span class="inline-block rounded bg-white/20 px-1.5 py-0.5 text-[10px] font-mono font-bold uppercase tracking-wider text-white">
                  {newBannerCode}
                </span>
              {/if}
              <h4 class="font-bold text-sm text-white truncate">
                {newBannerTitle || 'Título do seu Banner'}
              </h4>
              {#if newBannerDescription}
                <p class="text-xs text-zinc-200 line-clamp-1">{newBannerDescription}</p>
              {/if}
            </div>
          </div>
        </div>
      </div>

      <div class="flex items-center justify-end gap-3 pt-2">
        <button
          type="button"
          class="rounded-xl border border-zinc-300 px-4 py-2 text-sm font-semibold text-zinc-700 hover:bg-zinc-50 cursor-pointer"
          on:click={() => (showNewBannerModal = false)}
        >
          Cancelar
        </button>
        <button
          type="button"
          class="inline-flex items-center gap-1.5 rounded-xl bg-indigo-600 px-4 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-700 cursor-pointer"
          on:click={handleCreateBanner}
        >
          <Check size={16} />
          Salvar Banner
        </button>
      </div>
    </div>
  </div>
{/if}
