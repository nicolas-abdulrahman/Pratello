<script>
  import ItemCard from './section/ItemCard.svelte';
  import SectionTitle from './section/SectionTitle.svelte';
  import { resolveItemColors, resolveSection } from './section/section';
  import { restaurantDarkOverride } from '../lib/stores/restaurant_style';

  let {
    items,
    loading,
    failed,
    onOpenDetails,
    block,
    dark = false,
    palette = ''
  } = $props();

  let styles = $derived(resolveSection(block ?? {}, { dark }));

  const CARD_WIDTH = 280;
  const CARD_HEIGHT = 185;
  const MAX_GAP = 16;
  const MAX_PADDING = 32;
  const BORDER_WIDTH = 4;

  let sectionWidth = $state(0);

  let availableWidth = $derived(
    Math.max(0, sectionWidth - BORDER_WIDTH * 2)
  );

  let columns = $derived(
    availableWidth >= CARD_WIDTH * 4 ? 4 :
    availableWidth >= CARD_WIDTH * 3 ? 3 : 2
  );

  let availableSpace = $derived(
    Math.max(0, availableWidth - CARD_WIDTH * columns)
  );

  let pagePadding = $derived(
    Math.min(MAX_PADDING, availableSpace / 2)
  );

  let gap = $derived(
    Math.max(
      0,
      Math.min(
        MAX_GAP,
        (availableSpace - pagePadding * 2) / (columns - 1)
      )
    )
  );

  let cardWidth = $derived(
    Math.min(
      CARD_WIDTH,
      Math.max(
        0,
        (
          availableWidth -
          pagePadding * 2 -
          gap * (columns - 1)
        ) / columns
      )
    )
  );

  let scale = $derived(cardWidth / CARD_WIDTH);
</script>

<section
  class="section-container mx-auto w-full min-w-0 max-w-7xl py-space-lg"
  bind:clientWidth={sectionWidth}
  style={palette}
  style:padding-inline={`${pagePadding}px`}
  style:background-color={styles.background}
>
  <h2 class="mb-space-lg text-3xl font-black tracking-tight md:text-4xl">
    <SectionTitle title={styles.title} />
  </h2>

  <div
    class="section-content w-full min-w-0 rounded-[2rem] border-4 shadow-xl sm:rounded-[3rem]"
    style:border-style={styles.border}
    style:border-color={styles.borderColor}
  >
    {#if loading}
      <p role="status" class="text-text-muted">
        Carregando ofertas…
      </p>

    {:else if failed}
      <p role="alert" class="text-error">
        Não foi possível carregar as ofertas.
      </p>

    {:else if !items.length}
      <p class="text-text-muted">
        Nenhuma oferta disponível no momento.
      </p>

    {:else}
      <div
        class="items-grid"
        style:grid-template-columns={`repeat(${columns}, ${cardWidth}px)`}
        style:column-gap={`${gap}px`}
      >
        {#each items as item (item.id)}
          <div
            class="item-wrapper"
            style:width={`${cardWidth}px`}
            style:height={`${CARD_HEIGHT * scale}px`}
          >
            <div
              class="item-scale"
              style:transform={`scale(${scale})`}
            >
              <ItemCard
                colors={resolveItemColors({
                  ...block?.item_style,
                  ...item.style,
                  dark: $restaurantDarkOverride
                    ?? item.style?.dark
                    ?? block?.item_style?.dark
                }, styles.isDark, block?.bg ?? 'bg')}
                item={{
                  id: item.id,
                  name: item.title,
                  description: item.description,
                  imageUrl: item.image,
                  basePrice: item.basePrice,
                  originalPrice: item.originalPrice,
                  type: item.combo ? 'combo' : 'simple',
                  ingredients: [],
                  extras: [],
                  rating: item.rating,
                  reviewCount: item.count
                }}
                onOpenDetails={() => onOpenDetails(item.id)}
              />
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </div>
</section>

<style>
  .section-container {
    box-sizing: border-box;
  }

  .section-content {
    box-sizing: border-box;
    padding-inline: 0;
    padding-block: 16px;
  }

  .items-grid {
    display: grid;
    justify-content: center;
    row-gap: 16px;
    width: 100%;
    min-width: 0;
  }

  .item-wrapper {
    position: relative;
    min-width: 0;
  }

  .item-scale {
    position: absolute;
    top: 0;
    left: 0;
    width: 280px;
    height: 185px;
    transform-origin: top left;
  }
</style>
