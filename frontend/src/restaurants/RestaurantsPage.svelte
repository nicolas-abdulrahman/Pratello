<script lang="ts">
  import RestaurantsSection from './RestaurantsSection.svelte';

  const heroImage =
    'https://images.unsplash.com/photo-1568901346375-23c9450c58cd?w=900&h=900&fit=crop&auto=format&q=80';

  type Tag = { icon: 'chef' | 'pin' | 'leaf'; label?: string };
  type GalleryImage = { src: string; alt: string };
  type Banner = {
    id: string;
    eyebrow?: string;
    title: string;
    subtitle?: string;
    cta?: string;
    imageSrc?: string;
    imageAlt?: string;
  };
  type Restaurant = {
    key: string;
    name: string;
    rating: number;
    deliveryTime: string;
    logoSrc?: string;
    images?: GalleryImage[];
    tags?: Tag[];
    banners: Banner[];
  };

  /** Turn "João Burger" → "joao-burger" for safe URLs. */
  function slugify(value: string): string {
    return value
      .toLowerCase()
      .normalize('NFD')
      .replace(/[\u0300-\u036f]/g, '')
      .replace(/[^a-z0-9]+/g, '-')
      .replace(/(^-|-$)/g, '');
  }

  /** Works in plain Svelte/Vite. Swap for `goto(path)` from `$app/navigation` on SvelteKit. */
  function navigateTo(path: string) {
    window.location.assign(path);
  }

  const categories = [
    { emoji: '🍔', label: 'Hambúrguer', active: true },
    { emoji: '🍝', label: 'Italiana', active: false },
    { emoji: '🍱', label: 'Japonesa', active: false },
    { emoji: '🍕', label: 'Pizza', active: false },
    { emoji: '🍲', label: 'Brasileira', active: false },
    { emoji: '🥙', label: 'Árabe', active: false },
    { emoji: '🥡', label: 'Chinesa', active: false },
    { emoji: '🌱', label: 'Vegana', active: false },
    { emoji: '🧁', label: 'Doces', active: false }
  ];

  const joaoBurgerImages: GalleryImage[] = [
    { src: 'https://images.unsplash.com/photo-1568901346375-23c9450c58cd?w=240&h=240&fit=crop&auto=format&q=80', alt: 'Cheeseburger clássico' },
    { src: 'https://images.unsplash.com/photo-1571091718767-18b5b1457add?w=240&h=240&fit=crop&auto=format&q=80', alt: 'Bacon burger' },
    { src: 'https://images.unsplash.com/photo-1573080496219-bb080dd4f877?w=240&h=240&fit=crop&auto=format&q=80', alt: 'Batatas fritas' },
    { src: 'https://images.unsplash.com/photo-1572802419224-296b0aeee0d9?w=240&h=240&fit=crop&auto=format&q=80', alt: 'Combo burger + fritas' },
    { src: 'https://images.unsplash.com/photo-1629203851122-3726ecdf080e?w=240&h=240&fit=crop&auto=format&q=80', alt: 'Refrigerante gelado' }
  ];

  const restaurants: Restaurant[] = [
    {
      key: 'joao-burger',
      name: 'João Burger',
      rating: 4.9,
      deliveryTime: '25-35 min',
      logoSrc:
        'https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcSNazKnntMiV0xmffkq5Zii4PUMFnhCOm_2wpvehU3zP4_wh-sxgWqGfja6&s=10',
      images: joaoBurgerImages,
      tags: [
        { icon: 'chef' },
        { icon: 'pin', label: 'local' },
        { icon: 'leaf', label: 'artesanal' }
      ],
      banners: [
        {
          id: 'jb-a',
          eyebrow: 'Oferta da semana',
          title: 'Combo duplo João Burger',
          subtitle: 'Dois burgers + fritas + refri por R$ 59,90',
          cta: 'Pedir agora',
          imageSrc:
            'https://images.unsplash.com/photo-1572802419224-296b0aeee0d9?w=1200&h=600&fit=crop&auto=format&q=80',
          imageAlt: 'Combo duplo de hambúrguer com fritas'
        },
        {
          id: 'jb-b',
          eyebrow: 'Novidade',
          title: 'Smash Bacon Supremo',
          subtitle: 'Blend 180g, cheddar, bacon crocante e molho da casa',
          cta: 'Conhecer',
          imageSrc:
            'https://images.unsplash.com/photo-1571091718767-18b5b1457add?w=1200&h=600&fit=crop&auto=format&q=80',
          imageAlt: 'Smash bacon burger'
        },
        {
          id: 'jb-c',
          eyebrow: 'Fritas',
          title: 'Fritas rústicas com cheddar',
          subtitle: 'Porção generosa pra dividir (ou não)',
          imageSrc:
            'https://images.unsplash.com/photo-1573080496219-bb080dd4f877?w=1200&h=600&fit=crop&auto=format&q=80',
          imageAlt: 'Batatas fritas'
        },
        {
          id: 'jb-d',
          eyebrow: 'Bebida',
          title: 'Refrigerante gelado',
          subtitle: 'Lata 350ml pra acompanhar seu burger',
          imageSrc:
            'https://images.unsplash.com/photo-1629203851122-3726ecdf080e?w=1200&h=600&fit=crop&auto=format&q=80',
          imageAlt: 'Refrigerante gelado'
        }
      ]
    },
    {
      // Kept as restaurant 2, but with empty content.
      key: 'restaurante-dois',
      name: 'Restaurante Dois',
      rating: 0,
      deliveryTime: '—',
      images: [],
      tags: [],
      banners: []
    }
  ];
</script>

<svelte:head>
  <title>Pratello - Delivery de Comida</title>
</svelte:head>

<main data-theme="brand" class="min-h-screen bg-background py-space-md text-text antialiased">
  <!-- Hero + search band -->
  <div class="page-container flex flex-col gap-space-xl">
    <!-- Hero -->
    <section class="hero-banner">
      <div class="hero-glow"></div>
      <div class="hero-blob-decorative"></div>

      <div class="relative z-10 grid grid-cols-1 items-center gap-6 lg:grid-cols-12">
        <div class="space-y-4 lg:col-span-7">
          <div class="inline-flex items-center gap-2 rounded-full bg-primary/10 px-3 py-1 text-xs font-semibold uppercase tracking-wider text-primary">
            <span>✨ Praticidade &amp; Agilidade</span>
          </div>
          <h1 class="text-3xl font-extrabold leading-[1.15] tracking-tight text-text lg:text-4xl xl:text-5xl">
            Descubra os melhores <br class="hidden xl:block" />restaurantes perto de você!
          </h1>
          <p class="max-w-lg text-base font-normal text-text-muted">
            Variedade, sabor e praticidade garantidos, tudo em um só lugar no Pratello.
          </p>
          <div class="pt-2">
            <a href="#restaurantes-proximos" class="cta-primary group">
              <span>Explorar restaurantes</span>
              <svg class="h-4 w-4 transform transition-transform group-hover:translate-x-1" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
                <path d="M14 5l7 7m0 0l-7 7m7-7H3" stroke-linecap="round" stroke-linejoin="round" />
              </svg>
            </a>
          </div>
        </div>

        <div class="relative flex items-center justify-center lg:col-span-5">
          <div class="relative flex h-72 w-72 items-center justify-center sm:h-80 sm:w-80">
            <div class="hero-image-halo"></div>
            <img
              alt="Hambúrguer artesanal"
              src={heroImage}
              class="relative z-10 h-full w-full rounded-3xl object-cover shadow-xl transition duration-300 hover:scale-[1.02]"
            />
          </div>
        </div>
      </div>
    </section>

    <!-- Search + filters -->
    <div id="restaurantes-proximos" class="space-y-3">
      <div class="search-panel">
        <div class="relative w-full flex-1">
          <svg class="absolute left-3.5 top-1/2 h-5 w-5 -translate-y-1/2 text-text-muted" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
            <circle cx="11" cy="11" r="8"></circle>
            <path d="m21 21-4.35-4.35" stroke-linecap="round"></path>
          </svg>
          <input
            type="text"
            placeholder="Buscar restaurantes ou pratos próximos..."
            class="w-full rounded-xl border border-border bg-surface py-2.5 pl-11 pr-4 text-sm font-medium text-text placeholder-text-muted shadow-2xs focus:outline-none focus:ring-2 focus:ring-primary"
          />
        </div>
        <div class="flex w-full items-center justify-end gap-2 sm:w-auto">
          <button type="button" class="inline-flex items-center gap-2 rounded-xl border border-border bg-surface px-3.5 py-2.5 text-xs font-semibold text-text transition hover:bg-surface-hover">
            <svg class="h-3.5 w-3.5 text-text-muted" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
              <path d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z" stroke-linecap="round" stroke-linejoin="round" />
            </svg>
            Filtros
          </button>
          <button type="button" class="inline-flex items-center gap-1.5 rounded-xl border border-border bg-surface px-3 py-2.5 text-xs font-medium text-text-muted transition hover:bg-surface-hover">
            <span>Mais bem avaliados</span>
          </button>
        </div>
      </div>

      <div class="space-y-1.5 pt-3">
        <span class="block px-1 text-xs font-bold uppercase tracking-wider text-text-muted">Categorias populares</span>
        <div class="relative flex items-center">
          <div class="hide-scrollbar flex w-full items-center gap-2 overflow-x-auto py-1 pr-10">
            {#each categories as cat (cat.label)}
              <button type="button" class="category-chip {cat.active ? 'is-active' : ''}">
                <span class="text-sm">{cat.emoji}</span>
                {cat.label}
              </button>
            {/each}
          </div>
          <button
            type="button"
            aria-label="Ver mais categorias"
            class="absolute right-0 top-1/2 flex h-7 w-7 -translate-y-1/2 items-center justify-center rounded-full border border-border bg-surface text-text shadow-md transition hover:bg-surface-hover"
          >
            <svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
              <path d="M9 5l7 7-7 7" stroke-linecap="round" stroke-linejoin="round" />
            </svg>
          </button>
        </div>
      </div>
    </div>
  </div>

  <!-- Restaurants band -->
  <section class="page-container mt-space-xl space-y-space-md">
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-space-2xs">
        <span class="text-lg text-primary">✨</span>
        <h2 class="text-xl font-bold tracking-tight text-text">Restaurantes próximos</h2>
      </div>
      <a href="#restaurantes-proximos" class="inline-flex items-center gap-1 text-xs font-bold text-primary transition hover:text-primary-hover">
        Ver todos
        <svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
          <path d="M9 5l7 7-7 7" stroke-linecap="round" stroke-linejoin="round" />
        </svg>
      </a>
    </div>

    <div class="flex flex-col gap-y-space-xl py-space-sm">
      {#each restaurants.slice(0, 2) as r (r.key)}
        <RestaurantsSection
          cardProps={{
            name: r.name,
            rating: r.rating,
            deliveryTime: r.deliveryTime,
            logoSrc: r.logoSrc,
            logoAlt: `${r.name} logo`,
            images: r.images,
            tags: r.tags,
            onmenu: () => navigateTo(`/r/${slugify(r.name)}`)
          }}
          banners={r.banners}
        />
      {/each}
    </div>
  </section>
</main>

<style>
  @reference "../app.css";

  .hero-banner {
    @apply relative overflow-hidden rounded-3xl border border-border p-8 shadow-sm;
    background-image: linear-gradient(to right, #F6EDE0, #FAF4EA, #F7EFE2);
  }
  @media (min-width: 1024px) {
    .hero-banner { @apply p-10; }
  }

  .hero-glow {
    @apply pointer-events-none absolute -bottom-20 -right-10 h-96 w-96 rounded-full blur-3xl;
    background-color: color-mix(in oklab, var(--color-primary) 20%, transparent);
  }

  .hero-blob-decorative {
    @apply pointer-events-none absolute right-60 top-1/2 h-64 w-64 -translate-y-1/2;
    background-color: color-mix(in oklab, var(--color-primary) 10%, transparent);
    border-radius: 42% 58% 70% 30% / 45% 45% 55% 55%;
  }

  .hero-image-halo {
    @apply absolute inset-0 animate-pulse rounded-full opacity-30 blur-2xl;
    background-image: linear-gradient(to top right, #fbbf24, #fde047);
  }

  .cta-primary {
    @apply inline-flex items-center gap-2.5 rounded-full bg-primary px-6 py-3.5 text-sm font-semibold text-white shadow-md shadow-primary/25 transition duration-200 hover:bg-primary-hover hover:shadow-lg;
  }

  .search-panel {
    @apply flex flex-col items-center gap-3 rounded-2xl border border-border p-2 shadow-xs sm:flex-row;
    background-color: color-mix(in oklab, var(--color-surface-hover) 80%, transparent);
  }

  .category-chip {
    @apply flex shrink-0 items-center gap-1.5 rounded-xl border px-3 py-1.5 text-xs shadow-2xs transition;
    @apply border-border bg-surface font-medium text-text-muted hover:bg-surface-hover;
  }
  .category-chip.is-active {
    @apply border-primary bg-primary-soft font-semibold text-text;
  }
</style>
