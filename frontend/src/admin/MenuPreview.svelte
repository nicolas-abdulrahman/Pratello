<script lang="ts">
  import type { Product } from './admin';
  import { readConfig, type MenuConfig } from './menuConfig';
  import { loadRestaurantLayout } from './menu';

  export let slug = '';
  let preview: { config: MenuConfig; products: Product[] } | null = null;
  let dark = false;

  try {
    preview = JSON.parse(localStorage.getItem(`pratello-menu-preview:${slug}`) || 'null');
    if (!preview) {
      const cfg = readConfig(slug);
      if (cfg && cfg.name) {
        preview = { config: cfg, products: [] };
      }
    }
  } catch {
    /* Invalid or unavailable storage. */
  }

  function getProductsForSection(section: any, allProducts: Product[]) {
    if (section.type === 'promotion_banner') return [];
    if (section.category === 'sales') {
      return allProducts.filter(
        (p) => p.available && (
          (p.originalPrice && p.originalPrice > p.price) ||
          (p.effective_price_cents && p.base_price_cents && p.effective_price_cents < p.base_price_cents)
        )
      );
    }
    return allProducts.filter(
      (product) => product.available && (!section.category || section.category === 'all' || product.category === section.category)
    );
  }
</script>

{#if preview}
  {@const colors = preview.config.colors}
  <main
    class="min-h-screen p-6 sm:p-10 transition-colors"
    style:background-color={dark ? colors.bg_dark : colors.bg}
    style:color={dark ? colors.secondary_dark : colors.secondary}
  >
    <div class="mx-auto max-w-6xl">
      <!-- Navbar / Header bar -->
      <div class="flex items-center justify-between gap-4 border-b pb-4 border-current/20">
        <div>
          <span class="rounded bg-yellow-400 px-2 py-0.5 text-xs font-bold text-zinc-900">Prévia de teste</span>
          <h1 class="text-3xl font-bold mt-1">{preview.config.name}</h1>
          {#if preview.config.description}
            <p class="text-sm opacity-80 mt-1">{preview.config.description}</p>
          {/if}
        </div>
        <button
          type="button"
          class="rounded-xl px-4 py-2 font-semibold shadow-xs"
          style:background-color={dark ? colors.primary_dark : colors.primary}
          style:color={dark ? '#111111' : '#ffffff'}
          on:click={() => (dark = !dark)}
        >
          {dark ? 'Tema claro' : 'Tema escuro'}
        </button>
      </div>

      <!-- Hero preview if picture present -->
      {#if preview.config.picture}
        <div class="mt-6 overflow-hidden rounded-2xl relative shadow-md max-h-80">
          <img src={preview.config.picture} alt={preview.config.name} class="w-full h-full object-cover" />
          {#if preview.config.welcome_text}
            <div class="absolute bottom-4 left-4 bg-black/60 backdrop-blur-xs text-white p-4 rounded-xl">
              <span class="text-xs uppercase tracking-wider font-bold text-yellow-400">{preview.config.welcome_text}</span>
              <h2 class="text-xl font-bold">{preview.config.name}</h2>
            </div>
          {/if}
        </div>
      {/if}

      <!-- Sections & Banners -->
      {#each preview.config.sections as section (section.id)}
        {#if section.type === 'promotion_banner'}
          <!-- Banner Promocional -->
          <div class="mt-8 rounded-2xl border-2 border-dashed border-indigo-400 bg-indigo-50/20 p-6">
            <div class="flex items-center justify-between">
              <div>
                <span class="rounded bg-indigo-600 px-2 py-0.5 text-xs font-mono font-bold text-white">
                  CUPOM: {section.banner_id || section.category}
                </span>
                <h3 class="text-lg font-bold mt-1 text-zinc-900">Campanha Promocional Especial</h3>
                <p class="text-xs opacity-75">Aproveite os descontos usando o cupom no checkout!</p>
              </div>
              <span class="rounded-xl bg-indigo-600 px-4 py-2 text-xs font-bold text-white">
                Ativo no Cardápio
              </span>
            </div>
          </div>
        {:else}
          <!-- Catalog or Sales Section -->
          <section class="mt-10">
            <div class="mb-4 flex items-center justify-between border-b border-current/10 pb-2">
              <h2 class="text-xl font-bold">
                {section.name || (section.category === 'sales' ? 'Promoções' : 'Pratos')}
              </h2>
              {#if section.category === 'sales'}
                <span class="rounded-full bg-amber-400 px-2.5 py-0.5 text-xs font-bold text-zinc-900">
                  OFERTAS
                </span>
              {/if}
            </div>

            <div
              class={section.layout === 'horizontal'
                ? 'flex gap-4 overflow-x-auto pb-4'
                : 'grid gap-4 sm:grid-cols-2 lg:grid-cols-3'}
            >
              {#each getProductsForSection(section, preview.products) as product (product.id)}
                <article
                  class:preview-border={section.border === 'border' || section.border === 'solid'}
                  class:preview-dotted={section.border === 'dotted'}
                  class="min-w-0 rounded-xl p-4 bg-white/50 backdrop-blur-xs shadow-xs"
                  style:min-width={section.layout === 'horizontal' ? '240px' : undefined}
                >
                  {#if product.image}
                    <img src={product.image} alt="" class="mb-3 h-36 w-full rounded-lg object-cover" />
                  {/if}
                  <h3 class="font-bold text-base">{product.name}</h3>
                  {#if product.description}
                    <p class="mt-1 text-sm opacity-75 line-clamp-2">{product.description}</p>
                  {/if}
                  <div class="mt-3 flex items-center gap-2">
                    {#if product.originalPrice && product.originalPrice > product.price}
                      <span class="text-xs opacity-50 line-through">
                        {product.originalPrice.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })}
                      </span>
                    {/if}
                    <p class="font-bold text-lg">
                      {product.price.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })}
                    </p>
                  </div>
                </article>
              {:else}
                <p class="text-sm opacity-70 py-4">Nenhum prato disponível nesta seção.</p>
              {/each}
            </div>
          </section>
        {/if}
      {/each}
    </div>
  </main>
{:else}
  <main class="mx-auto max-w-4xl p-8">
    <h1 class="text-2xl font-bold">Prévia não encontrada</h1>
    <p class="mt-2 text-zinc-600">Faça o deploy de teste no admin neste navegador para gerar a prévia.</p>
    <a class="mt-4 inline-block underline font-semibold text-yellow-600" href="/admin">
      Ir para o admin
    </a>
  </main>
{/if}

<style>
  .preview-border {
    border: 1px solid currentColor;
  }
  .preview-dotted {
    border: 1px dotted currentColor;
  }
</style>
