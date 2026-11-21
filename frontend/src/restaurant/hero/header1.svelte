<script lang="ts">
  import { fallbackHeroButtons, header1Rules, restaurantCssVars, type ResolvedText } from '../../lib/stores/restaurant_style';

  let { slug, fallbackName, fallbackDescription, fallbackImage }: {
    slug: string; fallbackName?: string; fallbackDescription?: string | null; fallbackImage?: string | null;
  } = $props();
  let hero = $derived($header1Rules);
  let menuHref = $derived(`/r/${slug}/cardapio`);
  let deliveryHref = $derived(`/r/${slug}/delivery`);
  let dineIn = $derived(hero?.button1 ?? fallbackHeroButtons[0]);
  let delivery = $derived(hero?.button2 ?? fallbackHeroButtons[1]);
  let image = $derived(hero?.img || fallbackImage);
</script>

{#snippet styled(node: ResolvedText)}
  {#if node.mode === 'rough'}
    <span style:color={node.firstColor}>{node.first}</span><span style:color={node.secondColor}>{node.second}</span>
  {:else if node.mode === 'smooth'}
    <span class="gradient-text" style:background-image={node.gradientCss}>{node.text}</span>
  {:else}
    <span style:color={node.color}>{node.text}</span>
  {/if}
{/snippet}

{#if hero || fallbackName}
  <header class="overflow-hidden px-page-x pb-10 pt-20" style={`${$restaurantCssVars} background-color: ${hero?.bgCss ?? 'var(--color-background)'}; color: ${hero?.textCss ?? 'var(--color-text)'}; --hero-accent: ${hero?.accentCss ?? 'var(--color-primary)'};`}>
    <div class="mx-auto flex max-w-7xl flex-col items-center justify-between gap-space-lg lg:flex-row">
      <div class="min-w-0 text-center lg:w-1/2 lg:text-left">
        <p class="hero-welcome text-lg font-bold uppercase tracking-wide">
          {#if hero?.welcome.text}{@render styled(hero.welcome)}{:else}Bem-vindo ao{/if}
        </p>
        <h1 class="hero-title mt-space-2xs text-6xl font-black tracking-tighter md:text-8xl">
          {#if hero?.title.text}{@render styled(hero.title)}{:else}{fallbackName}{/if}
        </h1>
        {#if hero?.desc.text || fallbackDescription}
          <p class="mx-auto mt-space-lg max-w-lg text-lg md:text-xl lg:mx-0" style:color={hero?.mutedTextCss ?? 'var(--color-text-muted)'}>
            {#if hero?.desc.text}{@render styled(hero.desc)}{:else}{fallbackDescription}{/if}
          </p>
        {/if}
        <div class="mt-space-lg flex flex-wrap justify-center gap-space-sm lg:justify-start">
          <a class="hero-button items-center gap-space-2xs border-2 shadow-lg" href={menuHref} style:background-color={dineIn.bgCss} style:border-color={dineIn.borderCss} style:color={dineIn.textCss}>
            <svg aria-hidden="true" viewBox="0 0 24 24" class="size-5" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="5" y="3" width="14" height="18" rx="2"/><path d="M9 8h6m-6 4h6m-6 4h4"/></svg>
            Ver cardápio
          </a>
          <a class="hero-button items-center gap-space-2xs border-2 shadow-lg" href={delivery.href ?? deliveryHref} style:background-color={delivery.bgCss} style:border-color={delivery.borderCss} style:color={delivery.textCss}>
            <svg aria-hidden="true" viewBox="0 0 24 24" class="size-5" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 17H2V6h12v11h-4m4-8h4l4 4v4h-2m-4 0h-2"/><circle cx="7" cy="17" r="2"/><circle cx="18" cy="17" r="2"/></svg>
            Delivery
          </a>
        </div>
      </div>
      {#if image}
        <div class="flex w-full justify-center lg:w-1/2">
          <div class="aspect-square w-[300px] max-w-full overflow-hidden rounded-full shadow-2xl md:w-[450px]">
            <img src={image} alt={hero?.title.text || fallbackName || ''} class="h-full w-full scale-110 object-cover" />
          </div>
        </div>
      {/if}
    </div>
  </header>
{/if}

<style>
  .hero-welcome { color: var(--hero-accent); }

  .hero-title::after {
    content: '';
    display: block;
    height: var(--spacing-space-2xs);
    margin-block-start: var(--spacing-space-2xs);
    border-top: var(--spacing-space-2xs) solid var(--hero-accent);
    border-radius: 50%;
    transform: rotate(-2deg);
  }

  .gradient-text {
    display: inline-block;
    background-clip: text;
    color: transparent;
  }

  .hero-button {
    display: inline-flex;
    justify-content: center;
    border-radius: var(--radius-full);
    padding: var(--spacing-space-sm) var(--spacing-space-lg);
    font-weight: 700;
  }
</style>
