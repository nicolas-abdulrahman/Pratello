<script lang="ts">
  import { onMount } from 'svelte';
  import type { ComponentProps } from 'svelte';
  import RestaurantCard from './RestaurantCard.svelte';

  type Banner = {
    id: string;
    eyebrow?: string;
    title: string;
    subtitle?: string;
    cta?: string;
    imageSrc?: string;
    imageAlt?: string;
  };

  type Props = {
    banners?: Banner[];
    cardProps?: Partial<ComponentProps<typeof RestaurantCard>>;
    interval?: number;
    transitionMs?: number;
    height?: number;
    mainHeight?: number;
    sideHeight?: number;
    gap?: number;
    onselect?: (banner: Banner) => void;
  };

  const placeholderImage =
    'data:image/svg+xml;utf8,' +
    encodeURIComponent(
      `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 160">
        <g fill="none" stroke="#d4a84f" stroke-width="7" stroke-linecap="round" stroke-linejoin="round">
          <rect x="30" y="30" width="140" height="100" rx="14"/>
          <circle cx="72" cy="66" r="13"/>
          <path d="M44 118l38-38 27 27 18-18 29 29"/>
        </g>
      </svg>`
    );

  let {
    banners = [],
    cardProps = {},
    interval = 5000,
    transitionMs = 600,
    height = 608,
    mainHeight = 330,
    sideHeight = 150,
    gap = 14,
    onselect
  }: Props = $props();

  let step = $state(0);
  let paused = $state(false);

  const count = $derived(banners.length);

  const slots = $derived(
    count === 0
      ? []
      : (count === 1 ? [0] : count === 2 ? [0, 1] : [-2, -1, 0, 1, 2]).map((d) => {
          const k = step + d;
          return { k, d, banner: banners[((k % count) + count) % count] };
        })
  );

  function slotBox(d: number) {
    const mainTop = (height - mainHeight) / 2;
    const belowTop = mainTop + mainHeight + gap;
    const aboveTop = mainTop - gap - sideHeight;
    switch (d) {
      case 0:
        return { top: mainTop, h: mainHeight, opacity: 1, z: 20 };
      case 1:
        return { top: belowTop, h: sideHeight, opacity: 1, z: 10 };
      case -1:
        return { top: aboveTop, h: sideHeight, opacity: 1, z: 10 };
      case 2:
        return { top: belowTop + sideHeight + gap, h: sideHeight, opacity: 0, z: 0 };
      default:
        return { top: aboveTop - sideHeight - gap, h: sideHeight, opacity: 0, z: 0 };
    }
  }

  function go(d: number) {
    step += d;
  }

  onMount(() => {
    const id = setInterval(() => {
      if (!paused && count > 1) step += 1;
    }, interval);
    return () => clearInterval(id);
  });
</script>

<section class="restaurants-section">
  <div class="card-slot">
    <RestaurantCard {...cardProps} />
  </div>

  {#if count > 0}
  <div class="carousel-slot">
    <div
      class="relative min-w-0 flex-1 overflow-hidden rounded-3xl"
      style="height: {height}px"
      role="region"
      aria-roledescription="carousel"
      aria-label="Banners promocionais"
      onmouseenter={() => (paused = true)}
      onmouseleave={() => (paused = false)}
    >
      {#each slots as s (s.k)}
        {@const box = slotBox(s.d)}
        {@const main = s.d === 0}
        {@const faded = Math.abs(s.d) === 2}
        {@const hasImage = !!s.banner.imageSrc}
        <button
          type="button"
          disabled={faded || (main && !onselect)}
          aria-current={main ? 'true' : undefined}
          aria-label={s.banner.title}
          onclick={() => (main ? onselect?.(s.banner) : go(s.d))}
          class="absolute inset-x-0 overflow-hidden rounded-3xl text-left shadow-lg {hasImage
            ? 'text-white border border-black/10'
            : 'bg-gradient-to-br from-[#FAF3E0] via-[#FAF4EA] to-[#F7EFE2] text-[#2A2017] border border-[#EDE2D2]'}"
          style="top: {box.top}px; height: {box.h}px; opacity: {box.opacity}; z-index: {box.z}; transition: top {transitionMs}ms ease-in-out, height {transitionMs}ms ease-in-out, opacity {transitionMs}ms ease-in-out;"
        >
          {#if hasImage}
            <img
              src={s.banner.imageSrc}
              alt={s.banner.imageAlt ?? ''}
              class="absolute inset-0 h-full w-full object-cover"
              loading="lazy"
            />
            <span class="absolute inset-0 bg-gradient-to-r from-black/80 via-black/55 to-black/20"></span>
          {/if}

          <span class="relative z-10 flex h-full flex-col justify-center {main ? 'gap-3 p-8' : 'gap-1 p-5'}">
            {#if main && s.banner.eyebrow}
              <span class="text-xs font-semibold uppercase tracking-wider {hasImage ? 'text-amber-300' : 'text-amber-800'}">
                {s.banner.eyebrow}
              </span>
            {/if}
            <span class="block font-serif font-bold leading-tight {main ? 'text-3xl' : 'text-xl'}">
              {s.banner.title}
            </span>
            {#if main && s.banner.subtitle}
              <span class="block text-sm {hasImage ? 'text-white/85' : 'text-stone-600'}">
                {s.banner.subtitle}
              </span>
            {/if}
            {#if main}
              <span class="mt-2 inline-flex w-fit items-center rounded-full bg-amber-500 hover:bg-amber-600 px-5 py-2.5 text-sm font-semibold text-white shadow-md shadow-amber-500/20">
                {s.banner.cta ?? 'Saiba mais'}
              </span>
            {/if}
          </span>
        </button>
      {/each}
    </div>
  </div>
  {/if}
</section>

<style>
  @reference "../app.css";

  .restaurants-section {
    @apply grid w-full items-center gap-space-sm;
    margin: 0;
    padding: 0;
  }

  @media (min-width: 1024px) {
    .restaurants-section {
      @apply grid-cols-[minmax(auto,25rem)_1fr];
    }
  }

  .card-slot {
    @apply flex w-full min-w-0 justify-start;
  }

  .carousel-slot {
    @apply w-full min-w-0;
    margin: 0;
  }
</style>
