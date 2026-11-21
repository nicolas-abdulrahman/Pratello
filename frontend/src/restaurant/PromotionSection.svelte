<script>
  import PromotionCard from './PromotionCard.svelte';

  let { items, loading, failed, onOpenDetails } = $props();
</script>

<section class="w-full px-page-x py-space-lg">
  <h2 class="mb-space-lg text-3xl font-black tracking-tight md:text-4xl">Promo<span class="text-primary">ções</span></h2>
  <div class="rounded-[3rem] border-4 border-primary bg-primary-soft/40 p-space-md shadow-xl md:p-space-lg">
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
            <PromotionCard item={{ id: item.id, name: item.title, description: item.description,
              imageUrl: item.image, basePrice: item.basePrice, originalPrice: item.originalPrice,
              type: item.combo ? 'combo' : 'simple', ingredients: [], extras: [], rating: item.rating,
              reviewCount: item.count }} onOpenDetails={() => onOpenDetails(item.id)} />
          </div>
        {/each}
      </div>
    {/if}
  </div>
</section>
