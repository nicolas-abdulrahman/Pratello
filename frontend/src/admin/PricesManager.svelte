<script lang="ts">
  import { onMount } from 'svelte';
  import { Tag, Image, Sparkles, Check, AlertCircle, ArrowRight, Plus, ExternalLink } from 'lucide-svelte';
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
    updateStorefrontLayout,
    buildMenuPayload,
    type PromotionalBanner,
    type StorefrontMenuPayload,
    saveRestaurantLayout,
    loadRestaurantLayout
  } from './menu';

  export let slug = '';
  export let products: Product[] = [];
  export let categories: Category[] = [];

  let config: MenuConfig = readConfig(slug);
  let banners: PromotionalBanner[] = [];
  let loadingBanners = false;
  let statusMessage = '';
  let statusType: 'success' | 'error' | '' = '';
  let isDeploying = false;

  // Manual banner input for quick insertion
  let customBannerCode = '';
  let customBannerDark: boolean | null = null;

  $: activeSlug = slug || restaurantSlug(config.name) || 'joao-burger';

  $: discountedProducts = products.filter(
    (p) => (p.originalPrice && p.originalPrice > p.price) ||
           (p.effective_price_cents && p.base_price_cents && p.effective_price_cents < p.base_price_cents)
  );

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

  function addPromotionBanner(bannerId: string, darkOverride: boolean | null = null) {
    if (!bannerId.trim()) return;
    const newSection: MenuSection = {
      id: `promo_${crypto.randomUUID()}`,
      type: 'promotion_banner',
      name: `Banner Promocional (${bannerId.trim()})`,
      category: bannerId.trim(),
      banner_id: bannerId.trim(),
      border: 'none',
      layout: 'grid',
      dark: darkOverride,
    };
    config.sections = [...config.sections, newSection];
    saveConfig(config, activeSlug);
    statusType = 'success';
    statusMessage = `Banner "${bannerId}" adicionado ao layout do cardápio!`;
    customBannerCode = '';
  }

  function addSalesSection() {
    const existing = config.sections.find((s) => s.type === 'section' && s.category === 'sales');
    if (existing) {
      statusType = 'error';
      statusMessage = 'A seção de ofertas (Promoções) já está adicionada ao cardápio.';
      return;
    }
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
    statusType = 'success';
    statusMessage = 'Seção "Promoções" (sales) adicionada com sucesso ao cardápio!';
  }

  async function deployTest() {
    isDeploying = true;
    statusMessage = '';
    try {
      const payload: StorefrontMenuPayload = buildMenuPayload(config);
      // Save locally in persistent storage dictionary
      saveRestaurantLayout(activeSlug, payload);
      // Persist in localStorage for test preview
      localStorage.setItem(`pratello-menu-preview:${activeSlug}`, JSON.stringify({ config, products }));
      // Push to API PUT /api/restaurants/{slug}/menu
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
  <div class="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
    <div>
      <h1 class="text-2xl font-bold text-zinc-900">Preços & Promoções</h1>
      <p class="mt-1 text-sm text-zinc-600">
        Gerencie banners promocionais, cupons e seções de ofertas especiais para o storefront.
      </p>
    </div>
    <div class="flex items-center gap-3">
      <button
        type="button"
        class="inline-flex items-center gap-2 rounded-xl bg-yellow-400 px-4 py-2 font-semibold text-zinc-900 shadow-sm transition hover:bg-yellow-300 disabled:opacity-50"
        on:click={deployTest}
        disabled={isDeploying}
      >
        <Sparkles size={16} />
        {isDeploying ? 'Enviando...' : 'Deploy test'}
      </button>
    </div>
  </div>

  {#if statusMessage}
    <div
      role="alert"
      class="flex items-center gap-2 rounded-xl p-4 text-sm font-medium {statusType === 'success'
        ? 'bg-emerald-50 text-emerald-800 border border-emerald-200'
        : 'bg-rose-50 text-rose-800 border border-rose-200'}"
    >
      {#if statusType === 'success'}
        <Check size={18} class="text-emerald-600" />
      {:else}
        <AlertCircle size={18} class="text-rose-600" />
      {/if}
      <span>{statusMessage}</span>
    </div>
  {/if}

  <!-- Seção de Ofertas (Sales) -->
  <div class="rounded-2xl border border-zinc-200 bg-white p-5 shadow-sm">
    <div class="flex flex-col justify-between gap-3 sm:flex-row sm:items-center">
      <div>
        <h2 class="text-lg font-bold text-zinc-900 flex items-center gap-2">
          <Tag class="text-amber-500" size={20} />
          Seção de Descontos & Ofertas (Sales)
        </h2>
        <p class="mt-0.5 text-xs text-zinc-500">
          Renderiza automaticamente produtos com preço promocional ativo (effective_price_cents &lt; base_price_cents).
        </p>
      </div>
      <button
        type="button"
        class="inline-flex items-center gap-1.5 rounded-xl border border-zinc-300 bg-zinc-50 px-3 py-2 text-sm font-semibold text-zinc-700 hover:bg-zinc-100"
        on:click={addSalesSection}
      >
        <Plus size={16} />
        Adicionar Seção "Promoções" ao Cardápio
      </button>
    </div>

    <div class="mt-4">
      <h3 class="text-xs font-semibold uppercase tracking-wider text-zinc-500">
        Itens Atualmente em Promoção ({discountedProducts.length})
      </h3>
      {#if discountedProducts.length > 0}
        <div class="mt-2 grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {#each discountedProducts as item}
            <div class="flex items-center justify-between rounded-xl border border-amber-200 bg-amber-50/50 p-3">
              <div>
                <p class="font-semibold text-zinc-900">{item.name}</p>
                <div class="mt-1 flex items-center gap-2 text-xs">
                  <span class="text-zinc-400 line-through">
                    R$ {(item.originalPrice ?? ((item.base_price_cents ?? 0) / 100)).toFixed(2)}
                  </span>
                  <span class="font-bold text-emerald-600">
                    R$ {(item.price ?? ((item.effective_price_cents ?? 0) / 100)).toFixed(2)}
                  </span>
                </div>
              </div>
              <span class="rounded-lg bg-amber-100 px-2 py-1 text-xs font-bold text-amber-800">
                PROMO
              </span>
            </div>
          {/each}
        </div>
      {:else}
        <p class="mt-2 text-xs text-zinc-400">
          Nenhum prato com desconto no momento. Edite um prato em "Pratos" e preencha o preço original ("de") maior que o preço atual.
        </p>
      {/if}
    </div>
  </div>

  <!-- Banners Promocionais da Campanha -->
  <div class="rounded-2xl border border-zinc-200 bg-white p-5 shadow-sm space-y-4">
    <div class="flex flex-col justify-between gap-3 sm:flex-row sm:items-center">
      <div>
        <h2 class="text-lg font-bold text-zinc-900 flex items-center gap-2">
          <Image class="text-indigo-500" size={20} />
          Banners Promocionais Cadastrados
        </h2>
        <p class="mt-0.5 text-xs text-zinc-500">
          Obtidos via GET /api/restaurants/{activeSlug}/menu/banners.
        </p>
      </div>
      <button
        type="button"
        class="text-xs font-medium text-zinc-600 underline hover:text-zinc-900"
        on:click={fetchBanners}
        disabled={loadingBanners}
      >
        {loadingBanners ? 'Recarregando...' : 'Recarregar banners da API'}
      </button>
    </div>

    <!-- Adicionar Banner Personalizado ou Código -->
    <div class="flex flex-wrap items-end gap-3 rounded-xl border border-zinc-200 bg-zinc-50 p-3">
      <div class="flex-1 min-w-[200px]">
        <label for="custom-banner-input" class="block text-xs font-medium text-zinc-700">Código ou ID do Banner</label>
        <input
          id="custom-banner-input"
          class="mt-1 w-full rounded-lg border border-zinc-300 bg-white p-2 text-sm"
          placeholder="Ex: DUPLA8 ou UUID"
          bind:value={customBannerCode}
        />
      </div>
      <div>
        <label for="custom-banner-theme" class="block text-xs font-medium text-zinc-700">Modo Escuro</label>
        <select
          id="custom-banner-theme"
          class="mt-1 rounded-lg border border-zinc-300 bg-white p-2 text-sm"
          bind:value={customBannerDark}
        >
          <option value={null}>Herdar do tema</option>
          <option value={false}>Claro (dark: false)</option>
          <option value={true}>Escuro (dark: true)</option>
        </select>
      </div>
      <button
        type="button"
        class="inline-flex items-center gap-1 rounded-xl bg-zinc-900 px-3 py-2 text-sm font-semibold text-white hover:bg-zinc-800 disabled:opacity-50"
        disabled={!customBannerCode.trim()}
        on:click={() => addPromotionBanner(customBannerCode, customBannerDark)}
      >
        <Plus size={16} />
        Inserir Banner no Cardápio
      </button>
    </div>

    {#if loadingBanners}
      <p class="text-sm text-zinc-500">Carregando banners da API...</p>
    {:else if banners.length > 0}
      <div class="grid gap-4 sm:grid-cols-2">
        {#each banners as banner}
          <div class="relative overflow-hidden rounded-xl border border-zinc-200 bg-white p-4 shadow-xs">
            <div class="flex items-start justify-between gap-2">
              <div>
                <span class="inline-block rounded-md bg-indigo-100 px-2 py-0.5 text-xs font-mono font-bold text-indigo-700">
                  {banner.code || banner.id}
                </span>
                <h4 class="mt-2 font-bold text-zinc-900">{banner.title}</h4>
                <p class="mt-1 text-xs text-zinc-600 line-clamp-2">{banner.description}</p>
              </div>
              {#if banner.background_image_url}
                <img
                  src={banner.background_image_url}
                  alt={banner.title}
                  class="h-16 w-16 rounded-lg object-cover"
                />
              {/if}
            </div>
            <div class="mt-3 flex justify-end">
              <button
                type="button"
                class="inline-flex items-center gap-1 rounded-lg border border-zinc-300 px-2.5 py-1.5 text-xs font-semibold text-zinc-700 hover:bg-zinc-100"
                on:click={() => addPromotionBanner(banner.code || banner.id)}
              >
                <Plus size={14} />
                Usar no Cardápio
              </button>
            </div>
          </div>
        {/each}
      </div>
    {:else}
      <div class="rounded-xl border border-dashed border-zinc-200 p-6 text-center text-xs text-zinc-400">
        Nenhum banner retornado pela API para este restaurante. Você pode inserir um código de cupom/banner manualmente acima (ex: <code>DUPLA8</code>).
      </div>
    {/if}
  </div>
</div>
