<script lang="ts">
  import { onMount } from 'svelte';
  import {
    Plus,
    Trash2,
    Sparkles,
    Check,
    AlertCircle,
    Palette,
    Layers,
    Tag,
    Image,
    ChevronDown,
    ChevronUp,
    ExternalLink,
    Sliders,
    Eye
  } from 'lucide-svelte';
  import type { Product, Category } from './admin';
  import {
    readConfig,
    saveConfig,
    restaurantSlug,
    type MenuConfig,
    type MenuSection,
    defaultConfig
  } from './menuConfig';
  import {
    buildMenuPayload,
    updateStorefrontLayout,
    getStorefrontLayout,
    getPromotionalBanners,
    getMenuCategories,
    saveRestaurantLayout,
    loadRestaurantLayout,
    type StorefrontMenuPayload,
    type PromotionalBanner,
    type ItemStyleToken,
    type DisplayToken
  } from './menu';

  export let slug = '';
  export let products: Product[] = [];
  export let categories: Category[] = [];

  // Available token definitions according to manual
  const BASE_TOKENS: { value: ItemStyleToken; label: string }[] = [
    { value: 'bg', label: 'bg (Fundo)' },
    { value: 'primary_color', label: 'primary_color (Destaque Principal)' },
    { value: 'secondary_color', label: 'secondary_color (Secundário)' },
    { value: 'bg_dark', label: 'bg_dark (Fundo Escuro)' },
    { value: 'primary_color_dark', label: 'primary_color_dark (Destaque Escuro)' },
    { value: 'secondary_color_dark', label: 'secondary_color_dark (Secundário Escuro)' },
    { value: 'text', label: 'text (#111111 / #FFFFFF)' },
    { value: 'text_muted', label: 'text_muted (#6B7280 / #A1A1AA)' },
  ];

  const DISPLAY_TOKENS: { value: DisplayToken; label: string }[] = [
    ...BASE_TOKENS,
    { value: 'darker_bg', label: 'darker_bg' },
    { value: 'darker_primary_color', label: 'darker_primary_color' },
    { value: 'darker_secondary_color', label: 'darker_secondary_color' },
    { value: 'darker_bg_dark', label: 'darker_bg_dark' },
    { value: 'darker_primary_color_dark', label: 'darker_primary_color_dark' },
    { value: 'darker_secondary_color_dark', label: 'darker_secondary_color_dark' },
    { value: 'lighter_bg', label: 'lighter_bg' },
    { value: 'lighter_primary_color', label: 'lighter_primary_color' },
    { value: 'lighter_secondary_color', label: 'lighter_secondary_color' },
    { value: 'lighter_bg_dark', label: 'lighter_bg_dark' },
    { value: 'lighter_primary_color_dark', label: 'lighter_primary_color_dark' },
    { value: 'lighter_secondary_color_dark', label: 'lighter_secondary_color_dark' },
  ];

  let config: MenuConfig = readConfig(slug);
  let banners: PromotionalBanner[] = [];
  let message = '';
  let messageType: 'success' | 'error' | '' = '';
  let isDeploying = false;
  let activeTab: 'general' | 'hero' | 'precos' | 'sections' | 'preview_payload' = 'general';

  $: activeSlug = slug || restaurantSlug(config.name) || 'joao-burger';

  onMount(async () => {
    config = readConfig(activeSlug);
    if (!config.name && slug) {
      config.name = slug.replace(/-/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase());
    }

    // Try fetching existing server layout
    if (activeSlug) {
      try {
        const serverLayout = await getStorefrontLayout(activeSlug);
        if (serverLayout && serverLayout.display_config) {
          syncFromPayload(serverLayout);
        }
      } catch {
        // If not found on server, fallback to persistent draft
        const persistent = loadRestaurantLayout(activeSlug);
        if (persistent) syncFromPayload(persistent);
      }

      try {
        banners = await getPromotionalBanners(activeSlug);
      } catch {
        banners = [];
      }

      if (categories.length === 0) {
        try {
          const apiCats = await getMenuCategories(activeSlug);
          if (apiCats.length > 0) {
            categories = apiCats.map((c) => ({ id: c.id, name: c.name }));
          }
        } catch {
          // ignore
        }
      }
    }
  });

  function syncFromPayload(payload: StorefrontMenuPayload) {
    if (!payload?.display_config) return;
    const dc = payload.display_config;
    config.name = payload.name || config.name;
    config.typography = payload.typography || 'Inter';
    config.dark = dc.dark;

    if (dc.colors) {
      config.colors = {
        bg: dc.colors.bg || config.colors.bg,
        primary: dc.colors.primary_color || config.colors.primary,
        secondary: dc.colors.secondary_color || config.colors.secondary,
        bg_dark: dc.colors.bg_dark || config.colors.bg_dark,
        primary_dark: dc.colors.primary_color_dark || config.colors.primary_dark,
        secondary_dark: dc.colors.secondary_color_dark || config.colors.secondary_dark,
      };
    }

    if (dc.navbar) {
      config.navbar_bg = dc.navbar.bg_color;
      config.navbar_border = dc.navbar.border ?? false;
      config.navbar_border_style = dc.navbar.border_style || 'none';
      config.navbar_border_color = dc.navbar.border_color;
      config.navbar_dark = dc.navbar.dark;
    }

    if (dc.menu) {
      config.menu_container = dc.menu.container;
    }

    if (Array.isArray(dc.first_page) && dc.first_page.length > 0) {
      const first = dc.first_page[0];
      if (first.type === 'header2') {
        config.hero_type = 'header2';
        config.hero_bg = first.bg || 'bg';
        config.hero_dark = first.dark;
        config.picture = first.img || '';
      } else if (first.type === 'header1') {
        config.hero_type = 'header1';
        config.hero_bg = first.bg || 'bg';
        config.hero_dark = first.dark;
        config.picture = first.img || '';
        config.welcome_text = first.welcome?.text || '';
        config.description = first.desc?.text || '';
        if (first.button1) {
          config.button1_text = 'Botão 1';
          config.button1_bg = first.button1.bg || 'primary_color';
          config.button1_href = first.button1.href || '';
          config.button1_dark_text = first.button1.text?.dark ?? true;
        }
        if (first.button2) {
          config.button2_text = 'Botão 2';
          config.button2_bg = first.button2.bg || 'secondary_color';
          config.button2_href = first.button2.href || '';
          config.button2_dark_text = first.button2.text?.dark ?? false;
        }
      }

      // Remaining blocks are sections or banners
      const sections: MenuSection[] = [];
      for (let i = 1; i < dc.first_page.length; i++) {
        const b = dc.first_page[i];
        if (b.type === 'promotion_banner') {
          sections.push({
            id: b.id,
            type: 'promotion_banner',
            name: `Banner (${b.banner_id})`,
            category: b.banner_id,
            banner_id: b.banner_id,
            border: 'none',
            layout: 'grid',
            dark: b.dark,
          });
        } else if (b.type === 'section') {
          sections.push({
            id: b.id,
            type: 'section',
            name: b.title?.text || 'Seção',
            category: b.display_items?.value || '',
            filter_by: b.display_items?.by || 'category',
            border: b.border || 'none',
            border_color: b.border_color,
            layout: b.container || 'grid',
            bg: b.bg,
            dark: b.dark,
            title_color: b.title?.color,
            gradient_enabled: !!b.title?.gradient,
            gradient_first: b.title?.gradient?.first_color,
            gradient_second: b.title?.gradient?.second_color,
            gradient_end: b.title?.gradient?.end,
            gradient_smooth: b.title?.gradient?.smooth,
            item_style_bg: b.item_style?.bg,
            item_style_title: b.item_style?.title_color,
            item_style_price: b.item_style?.price_color,
            item_style_dark: b.item_style?.dark,
          });
        }
      }
      config.sections = sections;
    }
  }

  function persist() {
    saveConfig(config, activeSlug);
  }

  function addCatalogSection() {
    config.sections = [
      ...config.sections,
      {
        id: `sec_${crypto.randomUUID()}`,
        type: 'section',
        name: '',
        category: categories[0]?.id ?? '',
        filter_by: 'category',
        border: 'solid',
        border_color: 'primary_color',
        layout: 'grid',
        bg: 'bg',
        dark: null,
      },
    ];
    persist();
  }

  function addPromotionBannerBlock(bannerId = 'DUPLA8') {
    config.sections = [
      ...config.sections,
      {
        id: `promo_${crypto.randomUUID()}`,
        type: 'promotion_banner',
        name: `Banner Promocional (${bannerId})`,
        category: bannerId,
        banner_id: bannerId,
        border: 'none',
        layout: 'grid',
        dark: null,
      },
    ];
    persist();
    messageType = 'success';
    message = `Banner promocional "${bannerId}" inserido no layout.`;
  }

  function addSalesSectionBlock() {
    config.sections = [
      ...config.sections,
      {
        id: `sec_promo_${crypto.randomUUID()}`,
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
      },
    ];
    persist();
    messageType = 'success';
    message = 'Seção de ofertas (sales) inserida no layout.';
  }

  function removeSection(id: string) {
    config.sections = config.sections.filter((s) => s.id !== id);
    persist();
  }

  async function deploy() {
    const slg = activeSlug;
    if (!slg) {
      messageType = 'error';
      message = 'Informe um nome válido para o restaurante.';
      return;
    }

    persist();
    isDeploying = true;
    message = '';

    try {
      // 1. Build and validate layout payload according to OpenAPI manual rules
      const payload: StorefrontMenuPayload = buildMenuPayload(config);

      // 2. Save in dictionary persistent storage for this restaurant slug
      saveRestaurantLayout(slg, payload);

      // 3. Save local preview data for testing
      localStorage.setItem(`pratello-menu-preview:${slg}`, JSON.stringify({ config, products }));

      // 4. Send API PUT /api/restaurants/{slug}/menu
      await updateStorefrontLayout(slg, payload);

      messageType = 'success';
      message = `Cardápio publicado com sucesso na API para "${slg}" e salvo no armazenamento persistente!`;
      window.open(`/me/my-restaurant/${slg}`, '_blank', 'noopener');
    } catch (err: any) {
      messageType = 'error';
      message = err?.message || 'Erro ao realizar deploy do cardápio na API.';
    } finally {
      isDeploying = false;
    }
  }

  let generatedPayloadJson = '';
  $: {
    try {
      generatedPayloadJson = JSON.stringify(buildMenuPayload(config), null, 2);
    } catch (e: any) {
      generatedPayloadJson = `// Erro de validação:\n// ${e.message}`;
    }
  }
</script>

<div class="space-y-6">
  <!-- Header Bar -->
  <div class="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
    <div>
      <h1 class="text-2xl font-bold text-zinc-900">Preparar cardápio (Storefront Layout)</h1>
      <p class="mt-1 text-sm text-zinc-600">
        Configure todos os parâmetros visuais, seções, cabeçalho e banners promocionais conforme a API OpenAPI 2.2.16.
      </p>
    </div>
    <div class="flex items-center gap-3">
      <button
        type="button"
        class="inline-flex items-center gap-2 rounded-xl bg-yellow-400 px-5 py-2.5 font-bold text-zinc-900 shadow-sm transition hover:bg-yellow-300 disabled:opacity-50"
        on:click={deploy}
        disabled={isDeploying}
      >
        <Sparkles size={18} />
        {isDeploying ? 'Publicando...' : 'Deploy test'}
      </button>
    </div>
  </div>

  {#if message}
    <div
      role="alert"
      class="flex items-center gap-2 rounded-xl p-4 text-sm font-medium {messageType === 'success'
        ? 'bg-emerald-50 text-emerald-800 border border-emerald-200'
        : 'bg-rose-50 text-rose-800 border border-rose-200'}"
    >
      {#if messageType === 'success'}
        <Check size={18} class="text-emerald-600" />
      {:else}
        <AlertCircle size={18} class="text-rose-600" />
      {/if}
      <span>{message}</span>
    </div>
  {/if}

  <!-- Navigation Tabs -->
  <div class="flex flex-wrap gap-2 border-b border-zinc-200 pb-2">
    <button
      type="button"
      class="inline-flex items-center gap-2 rounded-lg px-3.5 py-2 text-sm font-semibold transition {activeTab === 'general' ? 'bg-zinc-900 text-white' : 'text-zinc-600 hover:bg-zinc-100'}"
      on:click={() => (activeTab = 'general')}
    >
      <Palette size={16} />
      Tema & Cores
    </button>
    <button
      type="button"
      class="inline-flex items-center gap-2 rounded-lg px-3.5 py-2 text-sm font-semibold transition {activeTab === 'hero' ? 'bg-zinc-900 text-white' : 'text-zinc-600 hover:bg-zinc-100'}"
      on:click={() => (activeTab = 'hero')}
    >
      <Image size={16} />
      Cabeçalho (Hero & Navbar)
    </button>
    <button
      type="button"
      class="inline-flex items-center gap-2 rounded-lg px-3.5 py-2 text-sm font-semibold transition {activeTab === 'precos' ? 'bg-zinc-900 text-white' : 'text-zinc-600 hover:bg-zinc-100'}"
      on:click={() => (activeTab = 'precos')}
    >
      <Tag size={16} />
      Preços & Promoções
    </button>
    <button
      type="button"
      class="inline-flex items-center gap-2 rounded-lg px-3.5 py-2 text-sm font-semibold transition {activeTab === 'sections' ? 'bg-zinc-900 text-white' : 'text-zinc-600 hover:bg-zinc-100'}"
      on:click={() => (activeTab = 'sections')}
    >
      <Layers size={16} />
      Seções do Cardápio ({config.sections.length})
    </button>
    <button
      type="button"
      class="inline-flex items-center gap-2 rounded-lg px-3.5 py-2 text-sm font-semibold transition {activeTab === 'preview_payload' ? 'bg-zinc-900 text-white' : 'text-zinc-600 hover:bg-zinc-100'}"
      on:click={() => (activeTab = 'preview_payload')}
    >
      <Sliders size={16} />
      JSON OpenAPI
    </button>
  </div>

  <!-- TAB 1: TEMA & CORES -->
  {#if activeTab === 'general'}
    <div class="space-y-5 rounded-2xl border border-zinc-200 bg-white p-6 shadow-sm">
      <h2 class="text-lg font-bold text-zinc-900">Configurações Gerais & Cores Base (6 Chaves Hex)</h2>
      <p class="text-xs text-zinc-500">
        Todo storefront define exatamente as 6 cores hexadecimais obrigatórias (^#[0-9A-Fa-f]&#123;6&#125;$).
      </p>

      <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <label class="block text-sm">
          <span class="font-medium text-zinc-700">Nome do Cardápio</span>
          <input
            class="mt-1 w-full rounded-lg border border-zinc-300 p-2 text-sm"
            bind:value={config.name}
            on:input={persist}
            placeholder="Ex: João Burger"
          />
        </label>
        <label class="block text-sm">
          <span class="font-medium text-zinc-700">Tipografia</span>
          <input
            class="mt-1 w-full rounded-lg border border-zinc-300 p-2 text-sm"
            bind:value={config.typography}
            on:input={persist}
            placeholder="Inter"
          />
        </label>
        <label class="block text-sm">
          <span class="font-medium text-zinc-700">Tema Padrão Raiz</span>
          <select
            class="mt-1 w-full rounded-lg border border-zinc-300 p-2 text-sm"
            bind:value={config.dark}
            on:change={persist}
          >
            <option value={false}>Claro (dark: false)</option>
            <option value={true}>Escuro (dark: true)</option>
          </select>
        </label>
        <label class="block text-sm">
          <span class="font-medium text-zinc-700">Layout Padrão do Cardápio (/cardapio)</span>
          <select
            class="mt-1 w-full rounded-lg border border-zinc-300 p-2 text-sm"
            bind:value={config.menu_container}
            on:change={persist}
          >
            <option value="grid">Grid (Colunas Responsivas)</option>
            <option value="horizontal">Horizontal (Carrossel Snap)</option>
          </select>
        </label>
      </div>

      <div class="mt-6 border-t border-zinc-200 pt-4">
        <h3 class="text-sm font-bold text-zinc-900 mb-3">Paleta de Cores Obrigatória</h3>
        <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          <label class="block text-xs">
            <span class="font-semibold text-zinc-700">bg (Fundo Claro)</span>
            <div class="mt-1 flex items-center gap-2">
              <input type="color" bind:value={config.colors.bg} on:input={persist} class="h-8 w-10 cursor-pointer rounded border" />
              <input class="w-full rounded-lg border border-zinc-300 p-1.5 text-xs font-mono" bind:value={config.colors.bg} on:input={persist} />
            </div>
          </label>
          <label class="block text-xs">
            <span class="font-semibold text-zinc-700">primary_color (Destaque Claro)</span>
            <div class="mt-1 flex items-center gap-2">
              <input type="color" bind:value={config.colors.primary} on:input={persist} class="h-8 w-10 cursor-pointer rounded border" />
              <input class="w-full rounded-lg border border-zinc-300 p-1.5 text-xs font-mono" bind:value={config.colors.primary} on:input={persist} />
            </div>
          </label>
          <label class="block text-xs">
            <span class="font-semibold text-zinc-700">secondary_color (Secundário Claro)</span>
            <div class="mt-1 flex items-center gap-2">
              <input type="color" bind:value={config.colors.secondary} on:input={persist} class="h-8 w-10 cursor-pointer rounded border" />
              <input class="w-full rounded-lg border border-zinc-300 p-1.5 text-xs font-mono" bind:value={config.colors.secondary} on:input={persist} />
            </div>
          </label>
          <label class="block text-xs">
            <span class="font-semibold text-zinc-700">bg_dark (Fundo Escuro)</span>
            <div class="mt-1 flex items-center gap-2">
              <input type="color" bind:value={config.colors.bg_dark} on:input={persist} class="h-8 w-10 cursor-pointer rounded border" />
              <input class="w-full rounded-lg border border-zinc-300 p-1.5 text-xs font-mono" bind:value={config.colors.bg_dark} on:input={persist} />
            </div>
          </label>
          <label class="block text-xs">
            <span class="font-semibold text-zinc-700">primary_color_dark (Destaque Escuro)</span>
            <div class="mt-1 flex items-center gap-2">
              <input type="color" bind:value={config.colors.primary_dark} on:input={persist} class="h-8 w-10 cursor-pointer rounded border" />
              <input class="w-full rounded-lg border border-zinc-300 p-1.5 text-xs font-mono" bind:value={config.colors.primary_dark} on:input={persist} />
            </div>
          </label>
          <label class="block text-xs">
            <span class="font-semibold text-zinc-700">secondary_color_dark (Secundário Escuro)</span>
            <div class="mt-1 flex items-center gap-2">
              <input type="color" bind:value={config.colors.secondary_dark} on:input={persist} class="h-8 w-10 cursor-pointer rounded border" />
              <input class="w-full rounded-lg border border-zinc-300 p-1.5 text-xs font-mono" bind:value={config.colors.secondary_dark} on:input={persist} />
            </div>
          </label>
        </div>
      </div>
    </div>
  {/if}

  <!-- TAB 2: HERO & NAVBAR -->
  {#if activeTab === 'hero'}
    <div class="space-y-6">
      <!-- Navbar Settings -->
      <div class="rounded-2xl border border-zinc-200 bg-white p-6 shadow-sm space-y-4">
        <h2 class="text-lg font-bold text-zinc-900">Barra de Navegação Superior (Navbar)</h2>
        <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <label class="block text-sm">
            <span class="font-medium text-zinc-700">Cor de Fundo (Base Token)</span>
            <select
              class="mt-1 w-full rounded-lg border border-zinc-300 p-2 text-sm"
              bind:value={config.navbar_bg}
              on:change={persist}
            >
              {#each BASE_TOKENS as token}
                <option value={token.value}>{token.label}</option>
              {/each}
            </select>
          </label>
          <label class="block text-sm">
            <span class="font-medium text-zinc-700">Estilo de Borda Inferior</span>
            <select
              class="mt-1 w-full rounded-lg border border-zinc-300 p-2 text-sm"
              bind:value={config.navbar_border_style}
              on:change={persist}
            >
              <option value="none">Nenhuma (none)</option>
              <option value="solid">Sólida (solid)</option>
              <option value="dotted">Pontilhada (dotted)</option>
            </select>
          </label>
          <label class="block text-sm">
            <span class="font-medium text-zinc-700">Cor da Borda (Display Token)</span>
            <select
              class="mt-1 w-full rounded-lg border border-zinc-300 p-2 text-sm"
              bind:value={config.navbar_border_color}
              on:change={persist}
            >
              <option value="">Padrão</option>
              {#each DISPLAY_TOKENS as token}
                <option value={token.value}>{token.label}</option>
              {/each}
            </select>
          </label>
          <label class="block text-sm">
            <span class="font-medium text-zinc-700">Tema da Navbar</span>
            <select
              class="mt-1 w-full rounded-lg border border-zinc-300 p-2 text-sm"
              bind:value={config.navbar_dark}
              on:change={persist}
            >
              <option value={null}>Herdar do root</option>
              <option value={false}>Forçar Claro</option>
              <option value={true}>Forçar Escuro</option>
            </select>
          </label>
        </div>
      </div>

      <!-- Hero Header Settings -->
      <div class="rounded-2xl border border-zinc-200 bg-white p-6 shadow-sm space-y-4">
        <h2 class="text-lg font-bold text-zinc-900">Bloco de Cabeçalho (first_page[0])</h2>
        <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          <label class="block text-sm">
            <span class="font-medium text-zinc-700">Tipo de Hero</span>
            <select
              class="mt-1 w-full rounded-lg border border-zinc-300 p-2 text-sm font-semibold"
              bind:value={config.hero_type}
              on:change={persist}
            >
              <option value="header1">Header 1 (Hero Completo com Botões)</option>
              <option value="header2">Header 2 (Capa Minimalista)</option>
            </select>
          </label>
          <label class="block text-sm">
            <span class="font-medium text-zinc-700">Cor de Fundo do Hero</span>
            <select
              class="mt-1 w-full rounded-lg border border-zinc-300 p-2 text-sm"
              bind:value={config.hero_bg}
              on:change={persist}
            >
              {#each DISPLAY_TOKENS as token}
                <option value={token.value}>{token.label}</option>
              {/each}
            </select>
          </label>
          <label class="block text-sm">
            <span class="font-medium text-zinc-700">Tema do Hero</span>
            <select
              class="mt-1 w-full rounded-lg border border-zinc-300 p-2 text-sm"
              bind:value={config.hero_dark}
              on:change={persist}
            >
              <option value={null}>Herdar do root</option>
              <option value={false}>Claro (dark: false)</option>
              <option value={true}>Escuro (dark: true)</option>
            </select>
          </label>
          <label class="block text-sm sm:col-span-2 lg:col-span-3">
            <span class="font-medium text-zinc-700">URL da Imagem de Capa / Hero</span>
            <input
              class="mt-1 w-full rounded-lg border border-zinc-300 p-2 text-sm"
              bind:value={config.picture}
              on:input={persist}
              placeholder="https://cdn.exemplo.com/hero.jpg"
            />
          </label>
        </div>

        {#if config.hero_type === 'header1'}
          <div class="mt-4 border-t border-zinc-200 pt-4 space-y-4">
            <h3 class="text-sm font-bold text-zinc-900">Textos & Botões do Hero 1</h3>
            <div class="grid gap-4 sm:grid-cols-2">
              <label class="block text-sm">
                <span class="font-medium text-zinc-700">Texto de Boas-vindas (welcome)</span>
                <input
                  class="mt-1 w-full rounded-lg border border-zinc-300 p-2 text-sm"
                  bind:value={config.welcome_text}
                  on:input={persist}
                  placeholder="Ex: Olá!"
                />
              </label>
              <label class="block text-sm">
                <span class="font-medium text-zinc-700">Descrição / Slogan (desc)</span>
                <input
                  class="mt-1 w-full rounded-lg border border-zinc-300 p-2 text-sm"
                  bind:value={config.description}
                  on:input={persist}
                  placeholder="Ex: Temos os melhores hambúrgueres artesanais"
                />
              </label>
            </div>

            <!-- Botões CTA -->
            <div class="grid gap-4 sm:grid-cols-2">
              <div class="rounded-xl border border-zinc-200 p-3 space-y-2 bg-zinc-50">
                <span class="text-xs font-bold uppercase text-zinc-600">Botão Primário (button1)</span>
                <input
                  class="w-full rounded-lg border border-zinc-300 p-1.5 text-xs bg-white"
                  placeholder="Link (#cardapio)"
                  bind:value={config.button1_href}
                  on:input={persist}
                />
                <div class="flex items-center gap-2">
                  <select
                    class="w-full rounded-lg border border-zinc-300 p-1.5 text-xs bg-white"
                    bind:value={config.button1_bg}
                    on:change={persist}
                  >
                    {#each DISPLAY_TOKENS as token}
                      <option value={token.value}>{token.label}</option>
                    {/each}
                  </select>
                  <label class="flex items-center gap-1 text-xs whitespace-nowrap">
                    <input
                      type="checkbox"
                      bind:checked={config.button1_dark_text}
                      on:change={persist}
                    />
                    Texto Branco
                  </label>
                </div>
              </div>

              <div class="rounded-xl border border-zinc-200 p-3 space-y-2 bg-zinc-50">
                <span class="text-xs font-bold uppercase text-zinc-600">Botão Secundário (button2)</span>
                <input
                  class="w-full rounded-lg border border-zinc-300 p-1.5 text-xs bg-white"
                  placeholder="Link (#precos)"
                  bind:value={config.button2_href}
                  on:input={persist}
                />
                <div class="flex items-center gap-2">
                  <select
                    class="w-full rounded-lg border border-zinc-300 p-1.5 text-xs bg-white"
                    bind:value={config.button2_bg}
                    on:change={persist}
                  >
                    {#each DISPLAY_TOKENS as token}
                      <option value={token.value}>{token.label}</option>
                    {/each}
                  </select>
                  <label class="flex items-center gap-1 text-xs whitespace-nowrap">
                    <input
                      type="checkbox"
                      bind:checked={config.button2_dark_text}
                      on:change={persist}
                    />
                    Texto Branco
                  </label>
                </div>
              </div>
            </div>
          </div>
        {/if}
      </div>
    </div>
  {/if}

  <!-- TAB 3: SEÇÃO PREÇOS & PROMOÇÕES -->
  {#if activeTab === 'precos'}
    <div class="space-y-6">
      <div class="rounded-2xl border border-zinc-200 bg-white p-6 shadow-sm space-y-4">
        <div class="flex flex-col justify-between gap-3 sm:flex-row sm:items-center">
          <div>
            <h2 class="text-lg font-bold text-zinc-900 flex items-center gap-2">
              <Tag class="text-amber-500" size={20} />
              Configurar Elementos de Preços & Promoções
            </h2>
            <p class="mt-1 text-xs text-zinc-500">
              Adicione banners promocionais de cupons ou seções dinâmicas de itens em promoção (sales).
            </p>
          </div>
          <div class="flex items-center gap-2">
            <button
              type="button"
              class="inline-flex items-center gap-1.5 rounded-xl border border-zinc-300 bg-zinc-50 px-3 py-2 text-xs font-semibold hover:bg-zinc-100"
              on:click={addSalesSectionBlock}
            >
              <Plus size={14} />
              + Seção de Promoções (Sales)
            </button>
            <button
              type="button"
              class="inline-flex items-center gap-1.5 rounded-xl bg-zinc-900 px-3 py-2 text-xs font-semibold text-white hover:bg-zinc-800"
              on:click={() => addPromotionBannerBlock('DUPLA8')}
            >
              <Plus size={14} />
              + Banner Promocional
            </button>
          </div>
        </div>

        <!-- Banners cadastrados na API -->
        <div class="mt-4 border-t border-zinc-100 pt-4">
          <h3 class="text-xs font-semibold uppercase tracking-wider text-zinc-500 mb-2">
            Banners Disponíveis da API ({banners.length})
          </h3>
          {#if banners.length > 0}
            <div class="grid gap-3 sm:grid-cols-2">
              {#each banners as b}
                <div class="flex items-center justify-between rounded-xl border border-zinc-200 p-3 bg-zinc-50/50">
                  <div>
                    <span class="font-mono text-xs font-bold text-indigo-700 bg-indigo-50 px-1.5 py-0.5 rounded">
                      {b.code || b.id}
                    </span>
                    <p class="text-sm font-semibold text-zinc-900 mt-1">{b.title}</p>
                    <p class="text-xs text-zinc-500 line-clamp-1">{b.description}</p>
                  </div>
                  <button
                    type="button"
                    class="rounded-lg border border-zinc-300 px-2.5 py-1 text-xs font-semibold hover:bg-white"
                    on:click={() => addPromotionBannerBlock(b.code || b.id)}
                  >
                    Usar
                  </button>
                </div>
              {/each}
            </div>
          {:else}
            <p class="text-xs text-zinc-400">
              Nenhum banner cadastrado no momento no backend. Você pode inserir qualquer código de banner livremente (ex: DUPLA8).
            </p>
          {/if}
        </div>
      </div>
    </div>
  {/if}

  <!-- TAB 4: SEÇÕES DO CARDÁPIO -->
  {#if activeTab === 'sections' || activeTab === 'precos'}
    <div class="space-y-4 rounded-2xl border border-zinc-200 bg-white p-6 shadow-sm">
      <div class="flex items-center justify-between">
        <h2 class="text-lg font-bold text-zinc-900">Blocos do Cardápio (first_page[1..N])</h2>
        <div class="flex items-center gap-2">
          <button
            type="button"
            class="rounded-xl border border-zinc-300 px-3 py-1.5 text-xs font-semibold hover:bg-zinc-50"
            on:click={addCatalogSection}
          >
            + Adicionar Seção Normal
          </button>
          <button
            type="button"
            class="rounded-xl bg-amber-400 px-3 py-1.5 text-xs font-semibold text-zinc-900 hover:bg-amber-300"
            on:click={addSalesSectionBlock}
          >
            + Seção Promoções (Sales)
          </button>
        </div>
      </div>

      {#each config.sections as section, index (section.id)}
        <div class="space-y-3 rounded-xl border {section.type === 'promotion_banner' ? 'border-indigo-300 bg-indigo-50/20' : section.category === 'sales' ? 'border-amber-300 bg-amber-50/20' : 'border-zinc-200'} p-4">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2">
              <span class="rounded-md bg-zinc-200 px-2 py-0.5 text-xs font-bold text-zinc-700">
                Bloco {index + 1}
              </span>
              {#if section.type === 'promotion_banner'}
                <span class="rounded-md bg-indigo-100 px-2 py-0.5 text-xs font-semibold text-indigo-800">
                  Banner Promocional
                </span>
              {:else if section.category === 'sales'}
                <span class="rounded-md bg-amber-100 px-2 py-0.5 text-xs font-semibold text-amber-800">
                  Ofertas (Sales)
                </span>
              {:else}
                <span class="rounded-md bg-zinc-100 px-2 py-0.5 text-xs font-semibold text-zinc-800">
                  Catálogo
                </span>
              {/if}
              <h3 class="font-semibold text-zinc-900 text-sm">{section.name || '(Sem título)'}</h3>
            </div>
            <button
              type="button"
              class="text-xs font-semibold text-red-600 hover:underline"
              on:click={() => removeSection(section.id)}
            >
              Remover
            </button>
          </div>

          {#if section.type === 'promotion_banner'}
            <!-- Promotional Banner Block fields -->
            <div class="grid gap-3 sm:grid-cols-2">
              <label class="text-xs">
                <span class="font-medium text-zinc-700">Código / ID do Banner (banner_id)</span>
                <input
                  class="mt-1 w-full rounded-lg border border-zinc-300 p-2 text-sm font-mono bg-white"
                  bind:value={section.banner_id}
                  on:input={() => {
                    section.category = section.banner_id || '';
                    persist();
                  }}
                  placeholder="DUPLA8"
                />
              </label>
              <label class="text-xs">
                <span class="font-medium text-zinc-700">Tema do Banner</span>
                <select
                  class="mt-1 w-full rounded-lg border border-zinc-300 p-2 text-sm bg-white"
                  bind:value={section.dark}
                  on:change={persist}
                >
                  <option value={null}>Herdar do root</option>
                  <option value={false}>Claro (dark: false)</option>
                  <option value={true}>Escuro (dark: true)</option>
                </select>
              </label>
            </div>
          {:else}
            <!-- Catalog Section fields -->
            <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
              <label class="text-xs">
                <span class="font-medium text-zinc-700">Nome da Seção</span>
                <input
                  class="mt-1 w-full rounded-lg border border-zinc-300 p-2 text-sm bg-white"
                  bind:value={section.name}
                  on:input={persist}
                  placeholder="Ex: Hambúrgueres Artesanais"
                />
              </label>

              <label class="text-xs">
                <span class="font-medium text-zinc-700">Filtro de Itens (display_items)</span>
                <div class="mt-1 flex gap-1">
                  <select
                    class="w-24 rounded-lg border border-zinc-300 p-2 text-xs bg-white"
                    bind:value={section.filter_by}
                    on:change={persist}
                  >
                    <option value="category">Categoria</option>
                    <option value="tag">Tag</option>
                    <option value="collection">Coleção</option>
                  </select>
                  {#if section.filter_by === 'category'}
                    <select
                      class="flex-1 rounded-lg border border-zinc-300 p-2 text-xs bg-white"
                      bind:value={section.category}
                      on:change={persist}
                    >
                      <option value="sales">Promoções (sales)</option>
                      <option value="combos">Combos</option>
                      <option value="all">Todos os itens</option>
                      {#each categories as cat}
                        <option value={cat.id}>{cat.name}</option>
                      {/each}
                    </select>
                  {:else}
                    <input
                      class="flex-1 rounded-lg border border-zinc-300 p-2 text-xs bg-white"
                      bind:value={section.category}
                      on:input={persist}
                      placeholder={section.filter_by === 'tag' ? 'Ex: vegetariano' : 'Ex: chef-specials'}
                    />
                  {/if}
                </div>
              </label>

              <label class="text-xs">
                <span class="font-medium text-zinc-700">Borda (border)</span>
                <div class="mt-1 flex gap-1">
                  <select
                    class="flex-1 rounded-lg border border-zinc-300 p-2 text-xs bg-white"
                    bind:value={section.border}
                    on:change={persist}
                  >
                    <option value="solid">Sólida (solid)</option>
                    <option value="none">Nenhuma (none)</option>
                    <option value="dotted">Pontilhada (dotted)</option>
                  </select>
                  <select
                    class="flex-1 rounded-lg border border-zinc-300 p-2 text-xs bg-white"
                    bind:value={section.border_color}
                    on:change={persist}
                  >
                    <option value="">Cor da Borda</option>
                    {#each DISPLAY_TOKENS as t}
                      <option value={t.value}>{t.value}</option>
                    {/each}
                  </select>
                </div>
              </label>

              <label class="text-xs">
                <span class="font-medium text-zinc-700">Container & Fundo</span>
                <div class="mt-1 flex gap-1">
                  <select
                    class="flex-1 rounded-lg border border-zinc-300 p-2 text-xs bg-white"
                    bind:value={section.layout}
                    on:change={persist}
                  >
                    <option value="grid">Grid</option>
                    <option value="horizontal">Horizontal</option>
                  </select>
                  <select
                    class="flex-1 rounded-lg border border-zinc-300 p-2 text-xs bg-white"
                    bind:value={section.bg}
                    on:change={persist}
                  >
                    {#each DISPLAY_TOKENS as t}
                      <option value={t.value}>{t.value}</option>
                    {/each}
                  </select>
                </div>
              </label>
            </div>

            <!-- Gradient & Item Style collapsible options -->
            <div class="border-t border-zinc-200/60 pt-3 text-xs space-y-2">
              <div class="flex flex-wrap items-center gap-4">
                <label class="flex items-center gap-1.5 cursor-pointer font-medium text-zinc-700">
                  <input
                    type="checkbox"
                    bind:checked={section.gradient_enabled}
                    on:change={persist}
                  />
                  Gradiente no Título
                </label>
                <label class="flex items-center gap-1.5 text-zinc-600">
                  <span>Tema do Card:</span>
                  <select
                    class="rounded border border-zinc-300 p-1 text-xs bg-white"
                    bind:value={section.item_style_dark}
                    on:change={persist}
                  >
                    <option value={null}>Padrão</option>
                    <option value={false}>Claro</option>
                    <option value={true}>Escuro</option>
                  </select>
                </label>
              </div>

              {#if section.gradient_enabled}
                <div class="flex flex-wrap items-center gap-3 rounded-lg bg-zinc-100 p-2.5">
                  <span class="font-bold text-zinc-700">Config Gradiente:</span>
                  <label class="flex items-center gap-1">
                    <span>Corte (runas):</span>
                    <input
                      type="number"
                      min="0"
                      class="w-16 rounded border border-zinc-300 p-1 text-xs bg-white"
                      bind:value={section.gradient_end}
                      on:input={persist}
                    />
                  </label>
                  <label class="flex items-center gap-1">
                    <span>Cor 1:</span>
                    <select
                      class="rounded border border-zinc-300 p-1 text-xs bg-white"
                      bind:value={section.gradient_first}
                      on:change={persist}
                    >
                      {#each DISPLAY_TOKENS as t}
                        <option value={t.value}>{t.value}</option>
                      {/each}
                    </select>
                  </label>
                  <label class="flex items-center gap-1">
                    <span>Cor 2:</span>
                    <select
                      class="rounded border border-zinc-300 p-1 text-xs bg-white"
                      bind:value={section.gradient_second}
                      on:change={persist}
                    >
                      {#each DISPLAY_TOKENS as t}
                        <option value={t.value}>{t.value}</option>
                      {/each}
                    </select>
                  </label>
                  <label class="flex items-center gap-1">
                    <input
                      type="checkbox"
                      bind:checked={section.gradient_smooth}
                      on:change={persist}
                    />
                    Suave (smooth)
                  </label>
                </div>
              {/if}
            </div>
          {/if}
        </div>
      {:else}
        <p class="text-sm text-zinc-500 py-4 text-center">
          Nenhuma seção adicionada. Adicione uma seção ou banner acima.
        </p>
      {/each}
    </div>
  {/if}

  <!-- TAB 5: JSON OPENAPI PAYLOAD PREVIEW -->
  {#if activeTab === 'preview_payload'}
    <div class="space-y-4 rounded-2xl border border-zinc-200 bg-white p-6 shadow-sm">
      <div class="flex items-center justify-between">
        <h2 class="text-lg font-bold text-zinc-900">Payload Gerado para PUT /api/restaurants/{activeSlug}/menu</h2>
        <span class="rounded-lg bg-zinc-100 px-2 py-1 text-xs font-mono text-zinc-600">OpenAPI 2.2.16</span>
      </div>
      <pre class="max-h-[500px] overflow-auto rounded-xl bg-zinc-900 p-4 font-mono text-xs text-emerald-400">{generatedPayloadJson}</pre>
    </div>
  {/if}

  <!-- Bottom Action Bar -->
  <div class="flex flex-wrap items-center justify-between gap-4 rounded-2xl border border-zinc-200 bg-white p-4 shadow-xs">
    <div class="flex items-center gap-2">
      <button
        type="button"
        class="inline-flex items-center gap-2 rounded-xl bg-yellow-400 px-5 py-2.5 font-bold text-zinc-900 shadow-sm transition hover:bg-yellow-300 disabled:opacity-50"
        on:click={deploy}
        disabled={isDeploying}
      >
        <Sparkles size={18} />
        {isDeploying ? 'Realizando deploy...' : 'Deploy test'}
      </button>
      <a
        href={`/me/my-restaurant/${activeSlug}`}
        target="_blank"
        rel="noopener"
        class="inline-flex items-center gap-1 rounded-xl border border-zinc-300 px-4 py-2.5 text-sm font-semibold text-zinc-700 hover:bg-zinc-50"
      >
        <Eye size={16} />
        Ver Prévia
      </a>
    </div>
    <p class="text-xs text-zinc-500">
      Ao clicar em "Deploy test", o layout é validado, salvo no armazenamento persistente e sincronizado via PUT /api/restaurants/{activeSlug}/menu.
    </p>
  </div>
</div>
