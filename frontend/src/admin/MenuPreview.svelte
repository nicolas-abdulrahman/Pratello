<script lang="ts">
  import { onMount } from 'svelte';
  import { Eye, ArrowLeft, Sun, Moon, Sparkles } from 'lucide-svelte';
  import FrontPage from '../restaurant/FrontPage.svelte';
  import {
    setRestaurantDisplayConfig,
    restaurantDisplayConfig,
    restaurantDark,
    restaurantPageBackground,
    restaurantThemeVars
  } from '../lib/stores/restaurant_style';
  import {
    loadRestaurantLayout,
    loadMenuDraft,
    buildMenuPayload,
    type StorefrontMenuPayload
  } from './menu';
  import { readConfig, type MenuConfig } from './menuConfig';

  export let slug = '';

  let config: MenuConfig | null = null;
  let layoutPayload: StorefrontMenuPayload | null = null;
  let loading = true;
  let isDark = false;

  $: if (slug) {
    loadPreview();
  }

  function loadPreview() {
    loading = true;
    try {
      // 1. Check local preview storage first (from deploy test)
      const previewRaw = localStorage.getItem(`pratello-menu-preview:${slug}`);
      if (previewRaw) {
        const parsed = JSON.parse(previewRaw);
        if (parsed?.config) {
          config = parsed.config;
          layoutPayload = buildMenuPayload(config!);
        }
      }

      // 2. Fall back to persistent dictionary layout
      if (!layoutPayload) {
        layoutPayload = loadRestaurantLayout(slug);
      }

      // 3. Fall back to draft config or global config
      if (!layoutPayload) {
        const draft = loadMenuDraft(slug) ?? readConfig(slug) ?? readConfig('');
        if (draft && draft.name) {
          config = draft;
          try {
            layoutPayload = buildMenuPayload(draft);
          } catch {}
        }
      }

      // 4. Guaranteed fallback so preview never fails to load
      if (!layoutPayload || !layoutPayload.display_config) {
        const fallbackName = (config?.name && config.name.trim()) ||
          (slug ? slug.replace(/-/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase()) : 'João Burger');
        const fallbackConfig: MenuConfig = {
          ...defaultConfig,
          ...(config || {}),
          name: fallbackName,
          colors: { ...defaultConfig.colors, ...(config?.colors || {}) },
          sections: config?.sections && config.sections.length > 0 ? config.sections : [
            {
              id: 'sec_sales',
              type: 'section',
              name: 'Promoções',
              category: 'sales',
              filter_by: 'category',
              border: 'dotted',
              border_color: 'primary_color',
              layout: 'grid',
              bg: 'bg',
            },
            {
              id: 'sec_pratos',
              type: 'section',
              name: 'Cardápio',
              category: 'all',
              filter_by: 'category',
              border: 'solid',
              border_color: 'primary_color',
              layout: 'grid',
              bg: 'bg',
            }
          ]
        };
        config = fallbackConfig;
        layoutPayload = buildMenuPayload(fallbackConfig);
      }

      // Apply to restaurant_style store
      if (layoutPayload?.display_config) {
        setRestaurantDisplayConfig(layoutPayload.display_config);
        isDark = layoutPayload.display_config.dark ?? false;
        restaurantDark.set(isDark);
      }
    } catch (e) {
      console.warn('Erro ao carregar prévia do cardápio:', e);
    } finally {
      loading = false;
    }
  }

  function toggleTheme() {
    isDark = !isDark;
    restaurantDark.set(isDark);
  }

  onMount(() => {
    loadPreview();
  });
</script>

<div class="relative min-h-screen">
  <!-- Top Preview Control Bar -->
  <aside
    aria-label="Barra de controle da prévia"
    class="sticky top-0 z-50 flex items-center justify-between border-b border-zinc-700 bg-zinc-900/95 px-4 py-2.5 text-white backdrop-blur-md shadow-md"
  >
    <div class="flex items-center gap-3">
      <a
        href="/admin"
        class="inline-flex items-center gap-1.5 rounded-lg border border-zinc-700 bg-zinc-800 px-3 py-1.5 text-xs font-medium text-zinc-200 transition hover:bg-zinc-700 hover:text-white"
      >
        <ArrowLeft size={14} />
        Voltar ao Admin
      </a>
      <div class="flex items-center gap-2">
        <span class="inline-flex items-center gap-1 rounded-full bg-yellow-400/20 px-2.5 py-0.5 text-xs font-bold text-yellow-300">
          <Sparkles size={12} />
          Prévia Ativa
        </span>
        <span class="hidden text-xs text-zinc-400 sm:inline">
          Restaurante: <strong class="text-white">{config?.name || layoutPayload?.name || slug}</strong>
        </span>
      </div>
    </div>

    <div class="flex items-center gap-2">
      <button
        type="button"
        on:click={toggleTheme}
        class="inline-flex items-center gap-1.5 rounded-lg border border-zinc-700 bg-zinc-800 px-3 py-1.5 text-xs font-semibold text-zinc-200 transition hover:bg-zinc-700 hover:text-white"
      >
        {#if isDark}
          <Sun size={14} class="text-yellow-400" />
          <span>Tema Claro</span>
        {:else}
          <Moon size={14} class="text-indigo-400" />
          <span>Tema Escuro</span>
        {/if}
      </button>
    </div>
  </aside>

  <!-- Actual Restaurant Storefront Rendering -->
  {#if loading}
    <div class="flex min-h-[50vh] items-center justify-center p-8">
      <p class="text-sm text-zinc-500">Carregando cardápio do restaurante...</p>
    </div>
  {:else}
    <div style={$restaurantThemeVars} class="w-full">
      <FrontPage
        {slug}
        name={config?.name || layoutPayload?.name || 'Restaurante Pratello'}
        coverURL={config?.picture || layoutPayload?.display_config?.first_page?.[0]?.img || null}
        description={config?.description || null}
      />
    </div>
  {/if}
</div>
