<script lang="ts">
  import RestaurantsSection from './RestaurantsSection.svelte';

  const heroImage =
    'https://lh3.googleusercontent.com/aida-public/AB6AXuCFmhePNwlSmB1YMgPzKwDCZj-s8i0PFk__wiHGFFC85DdcjxRKbGbiqkpd4iOs4lEGMOBrbxWJ5hncUO4i3O3XJfowTDv_VIXcSD0e3HcsjoJazcXB1h6LU2bNt3l5MJDaZzA7RYqkVn4uM1Y86lYlX9RqCNkLqmF-YTYEy3TKmju2xNphN-q2JmojJuCkifhTCAFao_XM1XiPwetyFQF0VhVP6KAAqO0Ebg1CXdnmG00NsCbrcwLO';

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

  const restaurants = [
    {
      key: 'restaurante-um',
      name: 'Restaurante Um',
      rating: 4.8,
      deliveryTime: '30-45 min',
      banners: [
        { id: 'r1-a', eyebrow: 'Oferta da semana', title: 'Promoção do dia', subtitle: 'Economize 30% no prato principal', cta: 'Pedir agora' },
        { id: 'r1-b', title: 'Frete grátis acima de R$ 50', subtitle: 'Para pedidos selecionados' },
        { id: 'r1-c', title: 'Combo para duas pessoas', subtitle: 'Sobremesa por conta da casa' },
        { id: 'r1-d', title: 'Novidade no cardápio', subtitle: 'Experimente antes que acabe' }
      ]
    },
    {
      key: 'restaurante-dois',
      name: 'Restaurante Dois',
      rating: 4.6,
      deliveryTime: '25-40 min',
      banners: [
        { id: 'r2-a', eyebrow: 'Destaque', title: 'Pratos da casa', subtitle: 'Receitas de família todos os dias', cta: 'Ver cardápio' },
        { id: 'r2-b', title: 'Cupom de 15% OFF', subtitle: 'Válido até domingo' },
        { id: 'r2-c', title: 'Almoço executivo', subtitle: 'Das 11h às 15h' }
      ]
    },
    {
      key: 'restaurante-tres',
      name: 'Restaurante Três',
      rating: 4.9,
      deliveryTime: '35-50 min',
      banners: [
        { id: 'r3-a', eyebrow: 'Mais pedido', title: 'Massas artesanais', subtitle: 'Feitas na hora, todos os dias', cta: 'Pedir agora' },
        { id: 'r3-b', title: 'Entrega grátis no primeiro pedido' },
        { id: 'r3-c', title: 'Vinhos com desconto', subtitle: 'Escolha o seu' },
        { id: 'r3-d', title: 'Sobremesas da semana', subtitle: 'Experimente o petit gâteau' }
      ]
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
              alt="Hambúrguer artesanal Pratello"
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
          cardProps={{ name: r.name, rating: r.rating, deliveryTime: r.deliveryTime }}
          banners={r.banners}
        />
      {/each}
    </div>
  </section>
</main>

<style>
  /* Give @apply access to the theme from app.css without re-emitting it. */
  @reference "../app.css";

  /* ─────────────────────────────────────────────────────────
     Component-scoped styles for the chunky one-off visuals.
     ───────────────────────────────────────────────────────── */

  /* Hero card: three-stop gradient + border/padding composition.
     Responsive padding is in a plain @media so it doesn't rely on
     @apply with an `lg:` variant (edge-case in Tailwind v4). */
  .hero-banner {
    @apply relative overflow-hidden rounded-3xl border border-border p-8 shadow-sm;
    background-image: linear-gradient(to right, #F6EDE0, #FAF4EA, #F7EFE2);
  }
  @media (min-width: 1024px) {
    .hero-banner { @apply p-10; }
  }

  /* Warm halo behind the hero (bottom-right) */
  .hero-glow {
    @apply pointer-events-none absolute -bottom-20 -right-10 h-96 w-96 rounded-full blur-3xl;
    background-color: color-mix(in oklab, var(--color-primary) 20%, transparent);
  }

  /* Organic blob on the mid-right of the hero */
  .hero-blob-decorative {
    @apply pointer-events-none absolute right-60 top-1/2 h-64 w-64 -translate-y-1/2;
    background-color: color-mix(in oklab, var(--color-primary) 10%, transparent);
    border-radius: 42% 58% 70% 30% / 45% 45% 55% 55%;
  }

  /* Animated pulse halo behind the hero image */
  .hero-image-halo {
    @apply absolute inset-0 animate-pulse rounded-full opacity-30 blur-2xl;
    background-image: linear-gradient(to top right, #fbbf24, #fde047);
  }

  /* Primary CTA pill */
  .cta-primary {
    @apply inline-flex items-center gap-2.5 rounded-full bg-primary px-6 py-3.5 text-sm font-semibold text-white shadow-md shadow-primary/25 transition duration-200 hover:bg-primary-hover hover:shadow-lg;
  }

  /* Search + filter bar */
  .search-panel {
    @apply flex flex-col items-center gap-3 rounded-2xl border border-border p-2 shadow-xs sm:flex-row;
    background-color: color-mix(in oklab, var(--color-surface-hover) 80%, transparent);
  }

  /* Category chip: base + `is-active` modifier */
  .category-chip {
    @apply flex shrink-0 items-center gap-1.5 rounded-xl border px-3 py-1.5 text-xs shadow-2xs transition;
    @apply border-border bg-surface font-medium text-text-muted hover:bg-surface-hover;
  }
  .category-chip.is-active {
    @apply border-primary bg-primary-soft font-semibold text-text;
  }
</style>
