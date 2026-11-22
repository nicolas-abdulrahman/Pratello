<script>
  import CombosCard from './CombosCard.svelte';
  import SectionTitle from './section/SectionTitle.svelte';
  import { resolveItemColors, resolveSection } from './section/section';
  import { restaurantDarkOverride } from '../lib/stores/restaurant_style';

  let { items, loading, failed, onOpenDetails, block, dark = false, palette = '' } = $props();
  let styles = $derived(resolveSection(block ?? {}, { dark }));
  let horizontal = $derived(styles.container === 'horizontal');
  let track = $state();
  let trackWidth = $state(0);
  let canGoLeft = $state(false);
  let canGoRight = $state(false);

  function updateArrows() {
    const left = track?.scrollLeft ?? 0;
    canGoLeft = left > 1;
    canGoRight = !!track && track.scrollWidth - trackWidth - left > 1;
  }

  $effect(() => {
    if (horizontal && items.length) updateArrows();
  });

  function scroll(direction) {
    const [first, second] = track.children;
    if (!second) return;
    const step = second.getBoundingClientRect().left - first.getBoundingClientRect().left;
    const reducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
    track.scrollBy({ left: step * direction, behavior: reducedMotion ? 'instant' : 'smooth' });
  }
</script>

<section class="mx-auto max-w-7xl space-y-space-sm px-page-x py-space-lg" style={palette} style:background-color={styles.background} style:border-bottom-style={styles.border} style:border-bottom-color={styles.borderColor} style:border-bottom-width={styles.border === 'none' ? '0' : '2px'}>
  <h2 class="text-2xl font-black md:text-3xl"><SectionTitle title={styles.title} /></h2>
  {#if loading}
    <p role="status" class="text-text-muted">Carregando itens…</p>
  {:else if failed}
    <p role="alert" class="text-error">Não foi possível carregar os itens.</p>
  {:else if !items.length}
    <p class="text-text-muted">Nenhum item disponível no momento.</p>
  {:else}
    <div class="relative">
    <div bind:this={track} bind:clientWidth={trackWidth} onscroll={updateArrows} class={horizontal
      ? 'flex snap-x gap-space-sm overflow-x-auto pb-space-xs'
      : 'grid grid-cols-[repeat(auto-fill,minmax(min(100%,300px),1fr))] gap-space-sm'}>
      {#each items as item (item.id)}
        <div class={horizontal ? 'w-[300px] shrink-0 snap-start' : 'min-w-0'}>
          <CombosCard image={item.image} title={item.title} desc={item.description} price={item.newPrice}
            colors={resolveItemColors({ ...block?.item_style, ...item.style, dark: $restaurantDarkOverride ?? item.style?.dark ?? block?.item_style?.dark }, styles.isDark, block?.bg ?? 'bg')}
            originalPrice={item.oldPrice} combo={item.combo} rating={item.rating} count={item.count}
            onOpenDetails={() => onOpenDetails(item.id)} />
        </div>
      {/each}
    </div>
    {#if horizontal && canGoLeft}
      <button type="button" onclick={() => scroll(-1)} aria-label="Combos anteriores" class="absolute start-0 top-1/2 z-10 grid size-10 -translate-y-1/2 place-items-center rounded-full border border-border bg-surface text-primary shadow-md hover:bg-surface-hover">
        <svg aria-hidden="true" viewBox="0 0 24 24" class="size-5" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="m15 18-6-6 6-6" /></svg>
      </button>
    {/if}
    {#if horizontal && canGoRight}
      <button type="button" onclick={() => scroll(1)} aria-label="Próximos combos" class="absolute end-0 top-1/2 z-10 grid size-10 -translate-y-1/2 place-items-center rounded-full border border-border bg-surface text-primary shadow-md hover:bg-surface-hover">
        <svg aria-hidden="true" viewBox="0 0 24 24" class="size-5" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="m9 18 6-6-6-6" /></svg>
      </button>
    {/if}
    </div>
  {/if}
</section>
