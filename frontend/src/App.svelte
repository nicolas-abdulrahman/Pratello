<script lang="ts">
  import { onMount } from 'svelte';
  import { transactions, visibleOrders, refreshOrder, updateOrder, isPaid, type SavedOrder } from './lib/stores/Transaction';
  import OrderPopup from './order/order_review/OrderPopup.svelte';
  import Footer from './common/Footer.svelte';
  import { restaurantDark, restaurantThemeVars, setRestaurantDisplayConfig } from './lib/stores/restaurant_style';
  import { brandDark } from './lib/stores/brand_theme';
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
  import MyRestaurants from './user/MyRestaurants.svelte';
  import Admin from './admin/Admin.svelte';
  import OrderPage from './order/orders_page/OrderPage.svelte';
  import ReviewPage from './order/order_review/ReviewPage.svelte';
  import type { HistoricalOrder } from './order/orders_page/HistoryOrderSection.svelte';

  // --- CART (derived store from cart.ts) ---
  // $cart_count already sums `quantity` across lines.
  let cartItemCount = $derived($cart_count);

  // --- ROUTE PARSER ---
  function parseRoute(path) {
    if (path === '/login') return { view: 'login', slug: '' };
    if (path === '/account' || path === '/account/') return { view: 'account', slug: '' };
    if (path === '/me/my-restaurant' || path === '/me/my-restaurant/') return { view: 'my-restaurants', slug: '' };
    if (path === '/me/pedidos' || path === '/me/pedidos/') return { view: 'orders', slug: '' };
    if (/^\/me\/pedidos\/[^/]+\/review\/?$/.test(path) || path === '/me/pedidos/review' || path === '/me/pedidos/review/' || path === '/order/review' || path === '/order/review/') return { view: 'order-review', slug: '' };
    if (path === '/order') return { view: 'order', slug: '' };
    if (path === '/checkout') return { view: 'checkout', slug: '' };
    if (path === '/restaurantes') return { view: 'restaurantes', slug: '' };
    if (path === '/cardapio') return { view: 'cardapio', slug: '' };

    const rMatch = path.match(/^\/(?:r|restaurant)\/([^/]+)(?:\/([^/]+))?\/?$/);
    if (rMatch) {
      const [, rSlug, sub] = rMatch;
      if (sub === 'admin') return { view: 'admin', slug: rSlug };
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

  let selectedOrderForReview = $state<HistoricalOrder | null>(null);
  let reviewPathId = $state(window.location.pathname.match(/^\/me\/pedidos\/([^/]+)\/review\/?$/)?.[1] ?? '');
  let notificationTime = $state(Date.now());
  let availableOrders = $derived($transactions.filter((o) => $account.isLoggedIn ? o.userId === $account.id : o.userId === null));
  let pickupOrder = $derived(currentView === 'order-review' ? undefined : availableOrders.find((o) =>
    (o.order.status === 'ready' || o.order.status === 'delivered') && !o.pickupAcknowledgedAt && !o.popupDismissed
  ));
  let preparingOrder = $derived(availableOrders.find((o) =>
    isPaid(o.order) && (o.order.status === 'pending' || o.order.status === 'preparing')
  ));
  let popupOrder = $derived(currentView === 'order-review' || pickupOrder || preparingOrder ? undefined : availableOrders.find((o) =>
    o.order.status === 'delivered' && isPaid(o.order) && !o.popupDismissed &&
    !!o.pickupAcknowledgedAt && notificationTime >= o.pickupAcknowledgedAt + 20000
  ));

  function historical(saved: SavedOrder): HistoricalOrder {
    return {
      id: String(saved.order.id), orderNumber: `#${saved.order.id}`,
      date: new Date(saved.order.created_at).toLocaleString('pt-BR'),
      restaurantName: saved.restaurantName,
      status: saved.order.status === 'delivered' ? 'Entregue' : saved.order.status === 'cancelled' ? 'Cancelado' : 'Em andamento',
      total: saved.order.total_cents / 100,
      items: (saved.order.items || []).filter((item) => !item.voided_at).map((item) => ({
        id: String(item.id), name: item.item_name, quantity: item.quantity,
        price: item.unit_price_cents / 100, imageUrl: '',
      })),
    };
  }
  let reviewOrder = $derived.by(() => {
    const match = availableOrders.find((o) => String(o.order.id) === reviewPathId);
    return match ? historical(match) : null;
  });

  onMount(() => {
    let running = false;
    const poll = async () => {
      if (running) return;
      running = true;
      try {
        await Promise.allSettled(visibleOrders().filter((o) => o.order.status !== 'cancelled' && (o.order.status !== 'delivered' || !isPaid(o.order))).map(refreshOrder));
      } finally { running = false; }
    };
    void poll();
    const timer = window.setInterval(poll, import.meta.env.TEST_ENV === 'true' || import.meta.env.TEST_ENV === '1' ? 15000 : 30000);
    const notificationTimer = window.setInterval(() => { notificationTime = Date.now(); }, 1000);
    return () => { window.clearInterval(timer); window.clearInterval(notificationTimer); };
  });

  function goToOrders() {
    currentView = 'orders';
    window.history.pushState({}, '', '/me/pedidos');
    window.scrollTo({ top: 0, behavior: 'instant' });
  }

  function goToOrderReview(order: HistoricalOrder) {
    selectedOrderForReview = order;
    reviewPathId = order.id;
    currentView = 'order-review';
    window.history.pushState({}, '', `/me/pedidos/${encodeURIComponent(order.id)}/review`);
    window.scrollTo({ top: 0, behavior: 'instant' });
  }

  // --- Browser back/forward ---
  if (typeof window !== 'undefined') {
    window.addEventListener('popstate', () => {
      const parsed = parseRoute(window.location.pathname);
      currentView = parsed.view;
      slug = parsed.slug;
      reviewPathId = window.location.pathname.match(/^\/me\/pedidos\/([^/]+)\/review\/?$/)?.[1] ?? '';
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
    currentView === 'my-restaurants' ||
    currentView === 'admin' ||
    currentView === 'orders' ||
    currentView === 'order-review' ||
    currentView === 'order' ||
    currentView === 'checkout'
  );

  // Return to the profile after signing in from /account.
  $effect(() => {
    if (currentView === 'login' && $account.isLoggedIn) {
      const next = new URLSearchParams(window.location.search).get('next');
      if (next === '/account' || next === '/me/my-restaurant') {
        slug = '';
        currentView = next === '/account' ? 'account' : 'my-restaurants';
        window.history.replaceState({}, '', next);
      } else {
        goToHome();
      }
    }
  });
</script>

<div data-theme={currentTheme} data-brand-dark={brandTheme && $brandDark ? '' : undefined} data-restaurant-dark={showNavLinks && $restaurantDark ? '' : undefined} style={showNavLinks ? $restaurantThemeVars : undefined} class="relative w-full min-h-screen bg-background text-text">
  {#if !hideGlobalUI}
    <Navbar
      {cartItemCount}
      {slug}
      onOpenMenu={goToCardapio}
      onGoHome={goToHome}
      onGoLogin={goToLogin}
      onGoAccount={goToAccount}
      onGoOrders={goToOrders}
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
          <CheckoutSection onCancel={goToOrder} total={$cart_total} {slug} />
        {:else if currentView === 'login'}
          <Login onGoHome={() => { if (!$account.isLoggedIn) goToHome(); }} />
        {:else if currentView === 'account'}
          <ProfilePage onBack={goBackFromAccount} onGoLogin={goToAccountLogin} />
        {:else if currentView === 'my-restaurants'}
          <MyRestaurants />
        {:else if currentView === 'orders'}
          <OrderPage onOpenReview={goToOrderReview} />
        {:else if currentView === 'order-review'}
          <ReviewPage order={reviewOrder} onBack={goToOrders} />
        {:else if currentView === 'admin'}
          {#if $account.isLoggedIn}
            <Admin {slug} session={{ name: $account.name, role: 'admin' }} onLogout={() => { account.logout(); window.location.href = '/'; }} />
          {:else}
            <main class="mx-auto max-w-4xl px-6 py-12">
              <p>Entre na sua conta para acessar o painel.</p>
              <a href="/login" class="text-amber-700 underline">Entrar</a>
            </main>
          {/if}
        {/if}
      </div>
    {/key}
  </div>

  {#if currentView === 'home' || currentView === 'restaurantes'}
    <Footer />
  {/if}
  {#if preparingOrder}
    <div class="fixed inset-0 z-50 flex items-start justify-center bg-black/50 p-4 pt-8">
      <div role="dialog" aria-modal="true" aria-labelledby="preparing-title" class="w-full max-w-md rounded-3xl bg-surface p-8 text-center shadow-xl space-y-4">
        <h2 id="preparing-title" class="text-xl font-bold">Transação realizada código:</h2>
        {#if preparingOrder.order.display_code}
          <p class="text-5xl font-black tracking-widest">{preparingOrder.order.display_code}</p>
        {:else}
          <p class="text-text-muted" role="status">Aguardando código do pedido…</p>
        {/if}
      </div>
    </div>
  {:else if pickupOrder}
    <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div role="dialog" aria-modal="true" aria-labelledby="pickup-title" class="w-full max-w-md rounded-3xl bg-surface p-8 shadow-xl space-y-4">
        <h2 id="pickup-title" class="text-2xl font-bold">Seu pedido está pronto!</h2>
        <p>Vá buscar seu pedido em {pickupOrder.restaurantName}.</p>
        <button type="button" onclick={() => updateOrder(pickupOrder!.order.id, { pickupAcknowledgedAt: Date.now() })} class="rounded-xl bg-primary px-5 py-3 font-bold text-white">OK</button>
      </div>
    </div>
  {:else if popupOrder}
    <OrderPopup restaurantName={popupOrder.restaurantName}
      onDismiss={() => updateOrder(popupOrder!.order.id, { popupDismissed: true })}
      onContinue={() => { const target = popupOrder!; updateOrder(target.order.id, { popupDismissed: true }); goToOrderReview(historical(target)); }} />
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
