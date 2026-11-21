<script>
  import CombosCard from './CombosCard.svelte';

  let { title, items, loading, failed, onOpenDetails } = $props();
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
    if (items.length) updateArrows();
  });

  function scroll(direction) {
    const [first, second] = track.children;
    if (!second) return;
    const step = second.getBoundingClientRect().left - first.getBoundingClientRect().left;
    const reducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
    track.scrollBy({ left: step * direction, behavior: reducedMotion ? 'instant' : 'smooth' });
  }
</script>

<section class="mx-auto max-w-7xl space-y-space-sm px-page-x py-space-lg">
  <h2 class="text-2xl font-black text-text md:text-3xl">{title}</h2>
  {#if loading}
    <p role="status" class="text-text-muted">Carregando itens…</p>
  {:else if failed}
    <p role="alert" class="text-error">Não foi possível carregar os itens.</p>
  {:else if !items.length}
    <p class="text-text-muted">Nenhum item disponível no momento.</p>
  {:else}
    <div class="relative">
    <div bind:this={track} bind:clientWidth={trackWidth} onscroll={updateArrows} class="flex snap-x gap-space-sm overflow-x-auto pb-space-xs">
      {#each items as item (item.id)}
        <div class="shrink-0 snap-start">
          <CombosCard image={item.image} title={item.title} desc={item.description} price={item.newPrice}
            originalPrice={item.oldPrice} combo={item.combo} rating={item.rating} count={item.count}
            onOpenDetails={() => onOpenDetails(item.id)} />
        </div>
      {/each}
    </div>
    {#if canGoLeft}
      <button type="button" onclick={() => scroll(-1)} aria-label="Combos anteriores" class="absolute start-0 top-1/2 z-10 grid size-10 -translate-y-1/2 place-items-center rounded-full border border-border bg-surface text-primary shadow-md hover:bg-surface-hover">
        <svg aria-hidden="true" viewBox="0 0 24 24" class="size-5" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="m15 18-6-6 6-6" /></svg>
      </button>
    {/if}
    {#if canGoRight}
      <button type="button" onclick={() => scroll(1)} aria-label="Próximos combos" class="absolute end-0 top-1/2 z-10 grid size-10 -translate-y-1/2 place-items-center rounded-full border border-border bg-surface text-primary shadow-md hover:bg-surface-hover">
        <svg aria-hidden="true" viewBox="0 0 24 24" class="size-5" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="m9 18 6-6-6-6" /></svg>
      </button>
    {/if}
    </div>
  {/if}
</section>
