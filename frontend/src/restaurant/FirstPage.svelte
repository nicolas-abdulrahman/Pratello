<script>
  import StyledTextNode from './StyledTextNode.svelte';
  import { buildRootColorVars, resolveColor } from './display';

  let { config, items, categories, banners, onOpenDetails } = $props();
  const normalized = (value) => String(value).trim().toLowerCase();
  function entries(filter) {
    if (filter.by === 'category') {
      if (normalized(filter.value) === 'sales') return items.filter((i) => i.oldPrice);
      const ids = new Set(categories.filter((c) => String(c.id) === filter.value || normalized(c.name) === normalized(filter.value)).map((c) => c.id));
      return items.filter((i) => ids.has(i.category));
    }
    return items.filter((i) => (filter.by === 'tag' ? i.tags : i.collections)?.some((value) => normalized(value) === normalized(filter.value)));
  }
  function scroll(event, direction) {
    const row = event.currentTarget.closest('section').querySelector('.horizontal');
    row.scrollBy({ left: direction * row.clientWidth, behavior: 'smooth' });
  }
</script>

<div style={buildRootColorVars(config.colors)}>
  {#each config.first_page as block (block.id)}
    {@const isDark = block.dark ?? config.dark}
    {#if block.type === 'header1'}
      <header class="page-container py-12 grid items-center gap-8 md:grid-cols-2" style:background-color={resolveColor(block.bg ?? 'bg', isDark)} style:color={resolveColor('text', isDark)}>
        <div><p><StyledTextNode content={block.welcome} {isDark} /></p><h1 class="text-5xl font-black"><StyledTextNode content={block.title} {isDark} /></h1><p><StyledTextNode content={block.desc} {isDark} /></p></div>
        <img src={block.img} alt={block.title.text} class="max-h-96 w-full rounded-2xl object-cover" />
      </header>
    {:else if block.type === 'header2'}
      <header style:background-color={resolveColor('bg', isDark)}><img src={block.img} alt="Destaque do restaurante" class="max-h-96 w-full object-cover" /></header>
    {:else if block.type === 'section'}
      <section class="px-page-x py-8" style:background-color={resolveColor(block.bg ?? 'bg', isDark)} style:color={resolveColor('text', isDark)} style:border-bottom={block.border && block.border !== 'none' ? `2px ${block.border} ${resolveColor('primary_color', isDark)}` : 'none'}>
        <div class="mb-4 flex items-center justify-between gap-4">
          <h2 class="text-3xl font-bold"><StyledTextNode content={block.title} {isDark} /></h2>
          {#if block.container === 'horizontal'}
            <div class="flex gap-3"><button type="button" aria-label="Anterior" onclick={(e) => scroll(e, -1)}>←</button><button type="button" aria-label="Próximo" onclick={(e) => scroll(e, 1)}>→</button></div>
          {/if}
        </div>
        <div class={block.container === 'grid' ? 'catalog-grid' : 'horizontal'}>
          {#each entries(block.display_items) as item (item.id)}
            <button type="button" onclick={() => onOpenDetails(item.id)} class="overflow-hidden rounded-xl border border-current text-start">
              <img src={item.image} alt="" loading="lazy" class="aspect-[4/3] w-full object-cover" />
              <span class="block p-3"><strong>{item.title}</strong><br />{#if item.oldPrice}<s class="me-2 opacity-75">{item.oldPrice}</s>{/if}{item.newPrice}</span>
            </button>
          {/each}
        </div>
      </section>
    {:else if block.type === 'promotion_banner'}
      {@const banner = banners.find((b) => String(b.id) === block.banner_id || normalized(b.code) === normalized(block.banner_id))}
      {#if banner}
        <section class="px-page-x py-8" style:background-color={resolveColor('bg', isDark)} style:color={resolveColor('text', isDark)}>
          {#if banner.background_image_url}<img src={banner.background_image_url} alt="" loading="lazy" class="mb-4 max-h-64 w-full rounded-xl object-cover" />{/if}
          <h2 class="text-2xl font-bold">{banner.title || banner.code}</h2><p>{banner.description}</p><strong>Cupom: {banner.code}</strong>
        </section>
      {/if}
    {/if}
  {/each}
</div>

<style>
  .catalog-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(min(100%, 180px), 1fr)); gap: 1rem; }
  .horizontal { display: flex; gap: 1rem; overflow-x: auto; }
  .horizontal > button { flex: 0 0 min(240px, 80%); }
</style>
