<script>
  import ItemCard from './section/ItemCard.svelte';
  import SectionTitle from './section/SectionTitle.svelte';
  import { resolveItemColors, resolveSection } from './section/section';
  import { restaurantDarkOverride } from '../lib/stores/restaurant_style';

  let { items, loading, failed, onOpenDetails, block, dark = false, palette = '' } = $props();
  let styles = $derived(resolveSection(block ?? {}, { dark }));
</script>

<section class="mx-auto w-full max-w-7xl px-page-x py-space-lg" style={palette} style:background-color={styles.background}>
  <h2 class="mb-space-lg text-3xl font-black tracking-tight md:text-4xl"><SectionTitle title={styles.title} /></h2>
  <div class="rounded-[3rem] border-4 p-space-md shadow-xl md:p-space-lg" style:border-style={styles.border} style:border-color={styles.borderColor}>
    {#if loading}
      <p role="status" class="text-text-muted">Carregando ofertas…</p>
    {:else if failed}
      <p role="alert" class="text-error">Não foi possível carregar as ofertas.</p>
    {:else if !items.length}
      <p class="text-text-muted">Nenhuma oferta disponível no momento.</p>
    {:else}
      <div class="flex snap-x gap-space-sm overflow-x-auto pb-space-xs md:grid md:grid-cols-3 lg:grid-cols-4">
        {#each items as item (item.id)}
          <div class="w-70 shrink-0 snap-start md:w-auto">
            <ItemCard colors={resolveItemColors({ ...block?.item_style, ...item.style, dark: $restaurantDarkOverride ?? item.style?.dark ?? block?.item_style?.dark }, styles.isDark, block?.bg ?? 'bg')}
              item={{ id: item.id, name: item.title, description: item.description,
              imageUrl: item.image, basePrice: item.basePrice, originalPrice: item.originalPrice,
              type: item.combo ? 'combo' : 'simple', ingredients: [], extras: [], rating: item.rating,
              reviewCount: item.count }} onOpenDetails={() => onOpenDetails(item.id)} />
          </div>
        {/each}
      </div>
    {/if}
  </div>
</section>
