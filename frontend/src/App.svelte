<script lang="ts">
  import Footer from './common/Footer.svelte';
  import { restaurantDark, restaurantThemeVars, setRestaurantDisplayConfig } from './lib/stores/restaurant_style';
  import Navbar from './common/Navbar.svelte';
  import OrderCart from './order/OrderCart.svelte';
  import OrderSection from './order/OrderSection.svelte';
  import CheckoutSection from './payment/CheckoutSection.svelte';

  import { cart, cart_count, cart_total } from './lib/stores/Cart';
  import { account } from './lib/stores/account';

  import Restaurant from './restaurant/FrontPage.svelte';
  import MenuPage from './menu/Cardapio.svelte';
  import Login from './auth/Auth.svelte';
  import RestaurantsPage from './marketplace/RestaurantsPage.svelte';
  import DeliverySection from './delivery/DeliverySection.svelte';
  import ProfilePage from './user/ProfilePage.svelte';

  // --- CART (derived store from cart.ts) ---
  // $cart_count already sums `quantity` across lines.
  let cartItemCount = $derived($cart_count);

  // --- ROUTE PARSER ---
  function parseRoute(path) {
    if (path === '/login') return { view: 'login', slug: '' };
    if (path === '/account' || path === '/account/') return { view: 'account', slug: '' };
    if (path === '/order') return { view: 'order', slug: '' };
    if (path === '/checkout') return { view: 'checkout', slug: '' };
    if (path === '/restaurantes') return { view: 'restaurantes', slug: '' };
    if (path === '/cardapio') return { view: 'cardapio', slug: '' };

    const rMatch = path.match(/^\/(?:r|restaurant)\/([^/]+)(?:\/([^/]+))?\/?$/);
    if (rMatch) {
      const [, rSlug, sub] = rMatch;
      if (sub === 'cardapio') return { view: 'cardapio', slug: rSlug };
      if (sub === 'delivery') return { view: 'delivery', slug: rSlug };
      return { view: 'restaurant', slug: rSlug };
    }

    return { view: 'home', slug: '' };
  }

  // --- INITIAL ROUTING ---
  const path = typeof window !== 'undefined' ? window.location.pathname : '/';
  const initial = parseRoute(path);
  let currentView = $state(initial.view);
  let slug = $state(initial.slug);

  // --- NAVIGATION ---
  function goToHome() {
    if (slug) {
      currentView = 'restaurant';
      window.history.pushState({}, '', `/r/${slug}`);
    } else {
      currentView = 'home';
      window.history.pushState({}, '', '/');
    }
  }

  function goToCardapio() {
    if (slug) {
      currentView = 'cardapio';
      window.history.pushState({}, '', `/r/${slug}/cardapio`);
    } else {
      currentView = 'cardapio';
      window.history.pushState({}, '', '/cardapio');
    }
  }

  function goToRestaurantes() {
    const path = slug ? `/r/${slug}/delivery` : '/restaurantes';
    if (window.location.pathname !== path) window.history.pushState({}, '', path);
    const route = parseRoute(path);
    currentView = route.view;
    slug = route.slug;
    window.scrollTo({ top: 0, behavior: 'instant' });
  }

  function goToLogin() {
    currentView = 'login';
    window.history.pushState({}, '', '/login');
  }

  function goToAccountLogin() {
    currentView = 'login';
    window.history.pushState({}, '', '/login?next=/account');
  }

  function goToAccount() {
    window.history.pushState({ fromAccountLink: true }, '', '/account');
    slug = '';
    currentView = 'account';
  }

  function goBackFromAccount() {
    if (window.history.state?.fromAccountLink) {
      window.history.back();
    } else {
      goToHome();
    }
  }

  function goToOrder(mode) {
    if (mode === 'delivery') {
      goToRestaurantes();
      return;
    }
    currentView = 'order';
    window.history.pushState({}, '', '/order');
  }

  function goToCheckout() {
    currentView = 'checkout';
    window.history.pushState({}, '', '/checkout');
  }

  // --- Browser back/forward ---
  if (typeof window !== 'undefined') {
    window.addEventListener('popstate', () => {
      const parsed = parseRoute(window.location.pathname);
      currentView = parsed.view;
      slug = parsed.slug;
    });
  }

  // --- NAVBAR + THEME BEHAVIOUR ---
  let showNavLinks = $derived(
    !!slug && ['restaurant', 'cardapio', 'delivery'].includes(currentView)
  );

  let showOrderCart = $derived(showNavLinks);

  type RestaurantBrand = { name: string; icon_url?: string | null; favicon_url?: string | null; cover_url?: string | null; description?: string | null };
  let restaurantBrand = $state<RestaurantBrand | null>(null);

  $effect(() => {
    restaurantBrand = null;
    if (!showNavLinks) return;
    const controller = new AbortController();
    void fetch(`/api/restaurants/${encodeURIComponent(slug)}/menu`, { signal: controller.signal })
      .then((response) => response.ok ? response.json() : null)
      .then((menu) => {
        if (!controller.signal.aborted) setRestaurantDisplayConfig(menu?.display_config ?? null);
      })
      .catch(() => {});
    const originalTitle = document.title;
    const favicon = document.querySelector<HTMLLinkElement>('link[rel="icon"]');
    const original = favicon?.getAttribute('href');
    const originalType = favicon?.getAttribute('type');
    void fetch(`/api/restaurants/${encodeURIComponent(slug)}`, { signal: controller.signal })
      .then((response) => response.ok ? response.json() as Promise<RestaurantBrand> : null)
      .then((brand) => {
        if (!brand || controller.signal.aborted) return;
        restaurantBrand = brand;
        document.title = brand.name;
        if (favicon && brand.favicon_url) {
          favicon.href = brand.favicon_url;
          favicon.removeAttribute('type');
        }
      })
      .catch(() => {});
    return () => {
      controller.abort();
      setRestaurantDisplayConfig(null);
      document.title = originalTitle;
      if (favicon && original) favicon.setAttribute('href', original);
      if (favicon && originalType) favicon.setAttribute('type', originalType);
    };
  });

  let currentTheme = $derived(showNavLinks ? 'restaurantTheme' : 'brand');
  let brandTheme = $derived(currentTheme === 'brand');

  let hideGlobalUI = $derived(
    currentView === 'login' ||
    currentView === 'account' ||
    currentView === 'order' ||
    currentView === 'checkout'
  );

  // Return to the profile after signing in from /account.
  $effect(() => {
    if (currentView === 'login' && $account.isLoggedIn) {
      if (new URLSearchParams(window.location.search).get('next') === '/account') {
        slug = '';
        currentView = 'account';
        window.history.replaceState({}, '', '/account');
      } else {
        goToHome();
      }
    }
  });
</script>

<div data-theme={currentTheme} data-restaurant-dark={showNavLinks && $restaurantDark ? '' : undefined} style={showNavLinks ? $restaurantThemeVars : undefined} class="relative w-full min-h-screen bg-background text-text">
  {#if !hideGlobalUI}
    <Navbar
      {cartItemCount}
      {slug}
      onOpenMenu={goToCardapio}
      onGoHome={goToHome}
      onGoLogin={goToLogin}
      onGoAccount={goToAccount}
      onGoRestaurantes={goToRestaurantes}
      {showNavLinks}
      brand={brandTheme}
    >
      {#snippet cartControls()}
        {#if showOrderCart}<OrderCart goToOrder={goToOrder} />{/if}
      {/snippet}
    </Navbar>
  {/if}

  <div class="relative overflow-hidden w-full">
    {#key currentView + ':' + slug}
      <div class="animate-slide-in w-full">
        {#if currentView === 'home'}
          <RestaurantsPage />
        {:else if currentView === 'restaurantes'}
          <RestaurantsPage />
        {:else if currentView === 'restaurant'}
          <Restaurant {slug} name={restaurantBrand?.name} coverURL={restaurantBrand?.cover_url} description={restaurantBrand?.description} />
        {:else if currentView === 'cardapio'}
          <MenuPage {slug} />
        {:else if currentView === 'delivery'}
          <DeliverySection />
        {:else if currentView === 'order'}
          <OrderSection goToCheckout={goToCheckout} onCancel={goToHome} />
        {:else if currentView === 'checkout'}
          <CheckoutSection onCancel={goToOrder} total={$cart_total} />
        {:else if currentView === 'login'}
          <Login onGoHome={() => { if (!$account.isLoggedIn) goToHome(); }} />
        {:else if currentView === 'account'}
          <ProfilePage onBack={goBackFromAccount} onGoLogin={goToAccountLogin} />
        {/if}
      </div>
    {/key}
  </div>

  {#if currentView === 'home' || currentView === 'restaurantes'}
    <Footer />
  {/if}
</div>

<style>
  .animate-slide-in {
    animation: slideIn 0.4s cubic-bezier(0.25, 1, 0.5, 1) forwards;
  }
  @keyframes slideIn {
    0% {
      opacity: 0;
      transform: translateY(20px);
    }
    100% {
      opacity: 1;
      transform: translateY(0);
    }
  }
</style>
