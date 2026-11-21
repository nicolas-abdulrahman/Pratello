<script lang="ts">
  type TagIcon = 'chef' | 'pin' | 'leaf';
  type Tag = { icon: TagIcon; label?: string };
  type GalleryImage = { src?: string; alt: string };

  type Props = {
    name?: string;
    description?: string;
    rating?: number;
    deliveryTime?: string;
    logoSrc?: string;
    logoAlt?: string;
    images?: GalleryImage[];
    tags?: Tag[];
    ctaLabel?: string;
    onmenu?: () => void;
    scrollSpeed?: number; // px per second
  };

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

  const defaultTags: Tag[] = [
    { icon: 'chef' },
    { icon: 'pin', label: 'local' },
    { icon: 'leaf', label: 'fresh' }
  ];

  let {
    name = 'Restaurant Name',
    description,
    rating,
    deliveryTime,
    logoSrc = placeholderLogo,
    logoAlt = 'Restaurant logo',
    images = [],
    tags = defaultTags,
    ctaLabel = 'Ver Cardápio',
    onmenu,
    scrollSpeed = 20
  }: Props = $props();

  let track = $state<HTMLDivElement>();
  let failedLogo = $state<string>();
  let rafId: number | undefined;
  let offset = 0;
  let lastTime = 0;
  let oneSetWidth = 0;

  const loopedImages = $derived(images.length > 1 ? [...images, ...images, ...images] : images);

  function measure() {
    if (!track) return;
    const first = track.children[0] as HTMLElement | undefined;
    const secondSetStart = track.children[images.length] as HTMLElement | undefined;
    oneSetWidth = first && secondSetStart ? secondSetStart.offsetLeft - first.offsetLeft : 0;
  }

  function tick(time: number) {
    if (oneSetWidth > 0) {
      const dt = lastTime ? (time - lastTime) / 1000 : 0;
      offset = (offset + scrollSpeed * dt) % oneSetWidth;
      if (track) track.style.transform = `translateX(-${offset}px)`;
    }
    lastTime = time;
    rafId = requestAnimationFrame(tick);
  }

  $effect(() => {
    if (!track || images.length < 2) return;
    offset = 0;
    lastTime = 0;
    measure();
    rafId = requestAnimationFrame(tick);

    const ro = new ResizeObserver(measure);
    if (track) ro.observe(track);

    return () => {
      if (rafId) cancelAnimationFrame(rafId);
      ro.disconnect();
    };
  });
</script>

<article class="restaurant-card relative">
  <!-- Logo panel -->
  <div class="logo-frame">
    <div class="logo-inner">
      {#key logoSrc}
        {@const source = logoSrc}
        <img
          src={source === failedLogo ? placeholderLogo : source}
          alt={logoAlt}
          class="logo-img"
          loading="lazy"
          onerror={() => { failedLogo = source; }}
        />
      {/key}
      <div class="logo-inner-rule"></div>
    </div>
  </div>

  <div class="p-card">
    <h2 class="restaurant-name">{name}</h2>
    {#if description}<p class="text-sm text-text-muted">{description}</p>{/if}

    <!-- Meta row -->
    <div class="restaurant-meta">
      <span class="flex items-center gap-1.5">
        <svg viewBox="0 0 24 24" class="star-icon h-5 w-5" aria-hidden="true">
          <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z" />
        </svg>
        <span class="text-base font-medium">{rating == null ? '—' : rating.toLocaleString('pt-BR', { minimumFractionDigits: 1, maximumFractionDigits: 1 })}</span>
      </span>
      <span class="meta-divider"></span>
      <span class="clock-icon flex items-center gap-1.5" title="Tempo aproximado em linha reta, com base na cidade selecionada">
        <svg viewBox="0 0 24 24" class="h-4 w-4 shrink-0" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
          <circle cx="12" cy="12" r="9" />
          <path d="M12 7v5l3 2" />
        </svg>
        <span class="text-base font-medium text-inherit">{deliveryTime ?? '—'}</span>
      </span>
    </div>

    {#if images.length}
    <div class="carousel-mask -mx-5 mt-5">
      <div bind:this={track} class="carousel-track">
        {#each loopedImages as img, i (i)}
          <div class="carousel-item relative bg-surface" aria-hidden={i >= images.length ? 'true' : undefined}>
            {#if img.src}<img src={img.src} alt="" loading="lazy" class="h-full w-full object-cover" />{/if}
            <span class="absolute inset-x-0 bottom-0 truncate bg-surface/90 px-space-2xs py-space-3xs text-xs font-semibold text-text" title={img.alt}>{img.alt}</span>
          </div>
        {/each}
      </div>
    </div>
    {/if}

    <!-- Tags -->
    <ul class="mt-4 flex items-center justify-center gap-2.5">
      {#each tags as tag}
        <li class="tag-chip">
          <svg viewBox="0 0 24 24" class="h-4 w-4 shrink-0" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            {#if tag.icon === 'chef'}
              <path d="M6 13.87A4 4 0 0 1 7.41 6a5.11 5.11 0 0 1 1.05-1.54 5 5 0 0 1 7.08 0A5.11 5.11 0 0 1 16.59 6 4 4 0 0 1 18 13.87V21H6Z" />
              <line x1="6" y1="17" x2="18" y2="17" />
            {:else if tag.icon === 'pin'}
              <path d="M20 10c0 6-8 12-8 12s-8-6-8-12a8 8 0 0 1 16 0Z" />
              <circle cx="12" cy="10" r="3" />
            {:else if tag.icon === 'leaf'}
              <path d="M11 20A7 7 0 0 1 9.8 6.1C15.5 5 17 4.48 19 2c1 2 2 4.18 2 8 0 5.5-4.78 10-10 10Z" />
              <path d="M2 21c0-3 1.85-5.36 5.08-6C9.5 14.52 12 13 13 12" />
            {/if}
          </svg>
          {#if tag.label}<span>{tag.label}</span>{/if}
        </li>
      {/each}
    </ul>

    <button type="button" onclick={() => onmenu?.()} class="cta-menu after:absolute after:inset-0" aria-label={`${ctaLabel}: ${name}`}>
      {ctaLabel}
    </button>
  </div>
</article>

<style>
  @reference "../app.css";

  .restaurant-card {
    --rc-gold-light:  #f3dc92;
    --rc-gold:        #b98a3a;
    --rc-gold-dark:   #a77a31;
    --rc-gold-accent: #d4a84f;
    --rc-cream:       #FAF4EA;
    --rc-cream-dark:  #faf3e3;
    --rc-brown:       #4a2514;
    --rc-brown-mid:   #6b3a22;
    --rc-orange:      #9a5532;
    --rc-orange-dark: #84452a;
    --rc-star:        #f0b232;
    --rc-tag-bg:      #f1e4cc;

    @apply w-full max-w-[25rem] overflow-hidden rounded-3xl;
    background-color: var(--rc-cream-dark);
    box-shadow: 0 20px 45px -15px rgba(60, 40, 20, 0.35);
  }

  .logo-frame {
    @apply rounded-[1.75rem] p-2.5;
    background: linear-gradient(135deg, var(--rc-gold-light), var(--rc-gold) 40%, var(--rc-gold-dark) 60%, var(--rc-gold));
  }
  .logo-inner {
    @apply relative flex aspect-[5/4] items-center justify-center overflow-hidden rounded-[1.1rem];
    background-color: var(--rc-cream);
  }
  .logo-img {
    @apply h-full w-full object-contain p-space-md;
    object-position: center;
  }
  .logo-inner-rule {
    @apply pointer-events-none absolute inset-4 rounded-[0.9rem] border;
    border-color: var(--rc-gold-accent);
  }

  .restaurant-name {
    @apply text-center font-serif text-4xl font-bold tracking-tight;
    color: var(--rc-brown);
  }
  .restaurant-meta {
    @apply mt-2 flex items-center justify-center gap-4;
    color: var(--rc-brown);
  }
  .star-icon  { fill: var(--rc-star); }
  .clock-icon { color: var(--rc-orange); }
  .meta-divider {
    @apply h-5 w-px;
    background-color: color-mix(in oklab, var(--rc-brown) 20%, transparent);
  }

  .carousel-mask { @apply overflow-hidden; }
  .carousel-track {
    @apply flex gap-3 py-2 will-change-transform;
    transform: translateX(0);
  }
  .carousel-item {
    @apply h-28 w-28 shrink-0 overflow-hidden rounded-xl shadow-md;
  }

  .tag-chip {
    @apply flex h-7 items-center gap-1.5 rounded-full px-2.5 text-xs font-medium;
    background-color: var(--rc-tag-bg);
    color: var(--rc-brown-mid);
  }

  .cta-menu {
    @apply mt-5 w-full rounded-full py-3.5 text-base font-medium shadow-md transition-colors;
    background-color: var(--rc-orange);
    color: #fdf6ea;
  }
  .cta-menu:hover         { background-color: var(--rc-orange-dark); }
  .cta-menu:active        { transform: scale(0.99); }
  .cta-menu:focus-visible { outline: 2px solid var(--rc-orange); outline-offset: 2px; }
</style>
