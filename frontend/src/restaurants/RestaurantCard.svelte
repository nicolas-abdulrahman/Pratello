<script lang="ts">
  import { onMount } from 'svelte';

  type TagIcon = 'chef' | 'pin' | 'leaf';
  type Tag = { icon: TagIcon; label?: string };
  type GalleryImage = { src: string; alt: string };

  type Props = {
    name?: string;
    rating?: number;
    deliveryTime?: string;
    logoSrc?: string;
    logoAlt?: string;
    images?: GalleryImage[];
    tags?: Tag[];
    ctaLabel?: string;
    onmenu?: () => void;
  };

  // Generic placeholder logo. Transparent background, so it sits on the navy panel.
  const placeholderLogo =
    'data:image/svg+xml;utf8,' +
    encodeURIComponent(
      `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 400 500">
        <g fill="none" stroke="#d4a84f" stroke-width="10" stroke-linecap="round" stroke-linejoin="round">
          <rect x="120" y="170" width="160" height="140" rx="18"/>
          <circle cx="165" cy="215" r="18"/>
          <path d="M135 290l50-50 35 35 25-25 45 45"/>
        </g>
        <text x="200" y="385" text-anchor="middle" font-family="Georgia, serif" font-size="30" letter-spacing="4" fill="#d4a84f">YOUR LOGO</text>
      </svg>`
    );

  const defaultImages: GalleryImage[] = [
    { src: 'https://picsum.photos/seed/dish1/240/240', alt: 'Dish 1' },
    { src: 'https://picsum.photos/seed/dish2/240/240', alt: 'Dish 2' },
    { src: 'https://picsum.photos/seed/dish3/240/240', alt: 'Dish 3' },
    { src: 'https://picsum.photos/seed/dish4/240/240', alt: 'Dish 4' },
    { src: 'https://picsum.photos/seed/dish5/240/240', alt: 'Dish 5' }
  ];

  const defaultTags: Tag[] = [
    { icon: 'chef' },
    { icon: 'pin', label: 'local' },
    { icon: 'leaf', label: 'fresh' }
  ];

  let {
    name = 'Restaurant Name',
    rating = 4.8,
    deliveryTime = '30-45 min',
    logoSrc = placeholderLogo,
    logoAlt = 'Restaurant logo',
    images = defaultImages,
    tags = defaultTags,
    ctaLabel = 'Ver Cardápio',
    onmenu
  }: Props = $props();

  let scroller: HTMLUListElement | undefined = $state();
  let active = $state(0);

  // Scroll the item at index i so its center sits in the middle of the strip.
  function centerOn(i: number, behavior: ScrollBehavior = 'auto') {
    const list = scroller;
    const item = list?.children[i] as HTMLElement | undefined;
    if (!list || !item) return;
    const listBox = list.getBoundingClientRect();
    const itemBox = item.getBoundingClientRect();
    const itemCenter = itemBox.left - listBox.left + list.scrollLeft + itemBox.width / 2;
    list.scrollTo({ left: itemCenter - list.clientWidth / 2, behavior });
  }

  // Whichever item is closest to the strip's center is the active one.
  function onScroll() {
    if (!scroller) return;
    const listBox = scroller.getBoundingClientRect();
    const mid = listBox.left + listBox.width / 2;
    let best = 0;
    let bestDist = Infinity;
    Array.from(scroller.children).forEach((el, i) => {
      const box = el.getBoundingClientRect();
      const dist = Math.abs(box.left + box.width / 2 - mid);
      if (dist < bestDist) {
        bestDist = dist;
        best = i;
      }
    });
    active = best;
  }

  function select(i: number) {
    active = i;
    centerOn(i, 'smooth');
  }

  onMount(() => {
    const start = Math.floor(images.length / 2);
    active = start;
    centerOn(start);
  });
</script>

{#snippet icon(kind: TagIcon | 'clock')}
  <svg
    viewBox="0 0 24 24"
    class="h-4 w-4 shrink-0"
    fill="none"
    stroke="currentColor"
    stroke-width="2"
    stroke-linecap="round"
    stroke-linejoin="round"
    aria-hidden="true"
  >
    {#if kind === 'chef'}
      <path d="M6 13.87A4 4 0 0 1 7.41 6a5.11 5.11 0 0 1 1.05-1.54 5 5 0 0 1 7.08 0A5.11 5.11 0 0 1 16.59 6 4 4 0 0 1 18 13.87V21H6Z" />
      <line x1="6" y1="17" x2="18" y2="17" />
    {:else if kind === 'pin'}
      <path d="M20 10c0 6-8 12-8 12s-8-6-8-12a8 8 0 0 1 16 0Z" />
      <circle cx="12" cy="10" r="3" />
    {:else if kind === 'leaf'}
      <path d="M11 20A7 7 0 0 1 9.8 6.1C15.5 5 17 4.48 19 2c1 2 2 4.18 2 8 0 5.5-4.78 10-10 10Z" />
      <path d="M2 21c0-3 1.85-5.36 5.08-6C9.5 14.52 12 13 13 12" />
    {:else if kind === 'clock'}
      <circle cx="12" cy="12" r="9" />
      <path d="M12 7v5l3 2" />
    {/if}
  </svg>
{/snippet}

<article
  class="mx-auto w-full max-w-[27.5rem] overflow-hidden rounded-3xl bg-[#faf3e3] shadow-[0_20px_45px_-15px_rgba(60,40,20,0.35)]"
>
  <!-- Logo panel: gold frame, navy field, thin inner gold rule -->
  <div class="rounded-[1.75rem] bg-[linear-gradient(135deg,#f3dc92,#b98a3a_40%,#f0d487_60%,#a77a31)] p-2.5">
    <div class="relative aspect-[4/5] overflow-hidden rounded-[1.1rem] bg-[#1c3a5c]">
      <img src={logoSrc} alt={logoAlt} class="h-full w-full object-contain p-12" />
      <div class="pointer-events-none absolute inset-4 rounded-[0.9rem] border border-[#d4a84f]"></div>
    </div>
  </div>

  <div class="px-5 pb-5 pt-5">
    <h2 class="text-center font-serif text-4xl font-bold tracking-tight text-[#4a2514]">
      {name}
    </h2>

    <!-- Meta row -->
    <div class="mt-2 flex items-center justify-center gap-4 text-[#4a2514]">
      <span class="flex items-center gap-1.5">
        <svg viewBox="0 0 24 24" class="h-5 w-5 fill-[#f0b232]" aria-hidden="true">
          <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z" />
        </svg>
        <span class="text-base font-medium">{rating.toFixed(1)}</span>
      </span>
      <span class="h-5 w-px bg-[#4a2514]/20"></span>
      <span class="flex items-center gap-1.5 text-[#9a5532]">
        {@render icon('clock')}
        <span class="text-base font-medium text-[#4a2514]">{deliveryTime}</span>
      </span>
    </div>

    <!-- Carousel: bleeds to the card edges, centered item is larger -->
    <div class="-mx-5 mt-5">
      <ul
        bind:this={scroller}
        onscroll={onScroll}
        class="flex snap-x snap-mandatory items-center gap-2.5 overflow-x-auto px-[calc(50%-3rem)] py-2 [scrollbar-width:none] [&::-webkit-scrollbar]:hidden"
        aria-label="Dish gallery"
      >
        {#each images as img, i}
          <li class="shrink-0 snap-center">
            <button
              type="button"
              onclick={() => select(i)}
              aria-label={`Show ${img.alt}`}
              aria-current={i === active ? 'true' : undefined}
              class="block overflow-hidden rounded-xl transition-[width,height,opacity,box-shadow] duration-300 ease-out focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[#9a5532]
                {i === active ? 'h-[7.5rem] w-[7.5rem] shadow-lg' : 'h-24 w-24 opacity-85'}"
            >
              <img src={img.src} alt={img.alt} loading="lazy" class="h-full w-full object-cover" />
            </button>
          </li>
        {/each}
      </ul>
    </div>

    <!-- Tags -->
    <ul class="mt-4 flex items-center justify-center gap-2.5">
      {#each tags as tag}
        <li class="flex h-7 items-center gap-1.5 rounded-full bg-[#f1e4cc] px-2.5 text-xs font-medium text-[#6b3a22]">
          {@render icon(tag.icon)}
          {#if tag.label}<span>{tag.label}</span>{/if}
        </li>
      {/each}
    </ul>

    <button
      type="button"
      onclick={() => onmenu?.()}
      class="mt-5 w-full rounded-full bg-[#9a5532] py-3.5 text-base font-medium text-[#fdf6ea] shadow-md transition-colors hover:bg-[#84452a] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[#9a5532] active:scale-[0.99]"
    >
      {ctaLabel}
    </button>
  </div>
</article>
