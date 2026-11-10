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

<main data-theme="brand" class="page min-h-screen bg-background py-6 text-text antialiased">
  <div class="page-container mx-auto flex max-w-7xl flex-col gap-12">
    <!-- Hero -->
    <section
      class="relative overflow-hidden rounded-3xl border border-[#EDE2D2] bg-gradient-to-r from-[#F6EDE0] via-[#FAF4EA] to-[#F7EFE2] p-8 shadow-sm lg:p-10"
    >
      <div class="pointer-events-none absolute -bottom-20 -right-10 h-96 w-96 rounded-full bg-amber-400/20 blur-3xl"></div>
      <div class="hero-blob pointer-events-none absolute right-60 top-1/2 h-64 w-64 -translate-y-1/2 bg-amber-200/40"></div>

      <div class="relative z-10 grid grid-cols-1 items-center gap-6 lg:grid-cols-12">
        <div class="space-y-4 lg:col-span-7">
          <div class="inline-flex items-center gap-2 rounded-full bg-amber-500/10 px-3 py-1 text-xs font-semibold uppercase tracking-wider text-amber-800">
            <span>✨ Praticidade &amp; Agilidade</span>
          </div>
          <h1 class="text-3xl font-extrabold leading-[1.15] tracking-tight text-stone-900 lg:text-4xl xl:text-5xl">
            Descubra os melhores <br class="hidden xl:block" />restaurantes perto de você!
          </h1>
          <p class="max-w-lg text-base font-normal text-stone-600">
            Variedade, sabor e praticidade garantidos, tudo em um só lugar no Pratello.
          </p>
          <div class="pt-2">
            <a
              href="#restaurantes-proximos"
              class="group inline-flex items-center gap-2.5 rounded-full bg-gradient-to-r from-amber-500 to-amber-600 px-6 py-3.5 text-sm font-semibold text-white shadow-md shadow-amber-500/25 transition duration-200 hover:from-amber-600 hover:to-amber-700 hover:shadow-lg"
            >
              <span>Explorar restaurantes</span>
              <svg class="h-4 w-4 transform transition-transform group-hover:translate-x-1" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
                <path d="M14 5l7 7m0 0l-7 7m7-7H3" stroke-linecap="round" stroke-linejoin="round" />
              </svg>
            </a>
          </div>
        </div>

        <div class="relative flex items-center justify-center lg:col-span-5">
          <div class="relative flex h-72 w-72 items-center justify-center sm:h-80 sm:w-80">
            <div class="absolute inset-0 animate-pulse rounded-full bg-gradient-to-tr from-amber-400 to-yellow-300 opacity-30 blur-2xl"></div>
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
      <div class="flex flex-col items-center gap-3 rounded-2xl border border-[#E6DDCE] bg-[#F4EDE2]/80 p-2 shadow-xs sm:flex-row">
        <div class="relative w-full flex-1">
          <svg class="absolute left-3.5 top-1/2 h-5 w-5 -translate-y-1/2 text-stone-400" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
            <circle cx="11" cy="11" r="8"></circle>
            <path d="m21 21-4.35-4.35" stroke-linecap="round"></path>
          </svg>
          <input
            type="text"
            placeholder="Buscar restaurantes ou pratos próximos..."
            class="w-full rounded-xl border border-[#E2D6C6] bg-white py-2.5 pl-11 pr-4 text-sm font-medium text-stone-800 placeholder-stone-400 shadow-2xs focus:outline-none focus:ring-2 focus:ring-amber-500"
          />
        </div>
        <div class="flex w-full items-center justify-end gap-2 sm:w-auto">
          <button type="button" class="inline-flex items-center gap-2 rounded-xl border border-[#E2D6C6] bg-white px-3.5 py-2.5 text-xs font-semibold text-stone-700 transition hover:bg-stone-50">
            <svg class="h-3.5 w-3.5 text-stone-500" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
              <path
                d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z"
                stroke-linecap="round"
                stroke-linejoin="round"
              ></path>
            </svg>
            Filtros
          </button>
          <button type="button" class="inline-flex items-center gap-1.5 rounded-xl border border-[#E2D6C6] bg-white px-3 py-2.5 text-xs font-medium text-stone-600 transition hover:bg-stone-50">
            <span>Mais bem avaliados</span>
          </button>
        </div>
      </div>

      <div class="space-y-1.5 pt-3">
        <span class="block px-1 text-[11px] font-bold uppercase tracking-wider text-stone-500">Categorias populares</span>
        <div class="relative flex items-center">
          <div class="hide-scrollbar flex w-full items-center gap-2 overflow-x-auto py-1 pr-10">
            {#each categories as cat (cat.label)}
              <button
                type="button"
                class="flex shrink-0 items-center gap-1.5 rounded-xl border px-3 py-1.5 text-xs shadow-2xs transition {cat.active
                  ? 'border-amber-300 bg-amber-200 font-semibold text-stone-900'
                  : 'border-[#E4DACB] bg-[#F4EDE2] font-medium text-stone-700 hover:bg-[#ECE4D6]'}"
              >
                <span class="text-sm">{cat.emoji}</span>
                {cat.label}
              </button>
            {/each}
          </div>
          <button
            type="button"
            aria-label="Ver mais categorias"
            class="absolute right-0 top-1/2 flex h-7 w-7 -translate-y-1/2 items-center justify-center rounded-full border border-[#E4DACB] bg-white text-stone-700 shadow-md transition hover:bg-stone-50"
          >
            <svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
              <path d="M9 5l7 7-7 7" stroke-linecap="round" stroke-linejoin="round" />
            </svg>
          </button>
        </div>
      </div>
    </div>

    <!-- Restaurants -->
    <section class="space-y-16">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2">
          <span class="text-lg text-amber-500">✨</span>
          <h2 class="text-xl font-bold tracking-tight text-stone-900">Restaurantes próximos</h2>
        </div>
        <a href="#restaurantes-proximos" class="inline-flex items-center gap-1 text-xs font-bold text-amber-600 transition hover:text-amber-700">
          Ver todos
          <svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
            <path d="M9 5l7 7-7 7" stroke-linecap="round" stroke-linejoin="round" />
          </svg>
        </a>
      </div>

      {#each restaurants as r (r.key)}
        <RestaurantsSection
          cardProps={{ name: r.name, rating: r.rating, deliveryTime: r.deliveryTime }}
          banners={r.banners}
        />
      {/each}
    </section>
  </div>
</main>
