<script lang="ts">
  import { onMount } from 'svelte';
  import { transactions, visibleOrders, refreshOrder, rememberSlug, updateOrder, isPaid, type SavedOrder } from './lib/stores/Transaction';
  import OrderPopup from './order/order_review/OrderPopup.svelte';
  import OrderSuccededPopup from './order/order_review/OrderSuccededPopup.svelte';
  import Footer from './common/Footer.svelte';
  import { restaurantDark, restaurantThemeVars, setRestaurantDisplayConfig } from './lib/stores/restaurant_style';
  import { brandDark } from './lib/stores/brand_theme';
  import Navbar from './common/Navbar.svelte';
  import OrderCart from './order/OrderCart.svelte';
  import OrderSection from './order/OrderSection.svelte';
  import CheckoutPage from './payment/CheckoutPage.svelte';
  import OrderStatusPage from './order/orders_page/OrderStatusPage.svelte';
  import { dineSession, enterDineIn, leaveDineIn } from './lib/stores/session';

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
  import MenuPreview from './admin/MenuPreview.svelte';
  import OrderPage from './order/orders_page/OrderPage.svelte';
  import ReviewPage from './order/order_review/ReviewPage.svelte';
  import type { HistoricalOrder } from './order/orders_page/HistoryOrderSection.svelte';

  // --- CART (derived store from cart.ts) ---
  // $cart_count already sums `quantity` across lines.
  let cartItemCount = $derived($cart_count);

  function extractReviewOrderId(path: string): string {
    const match2 = path.match(/^\/me\/pedidos\/([^/]+)\/([^/]+)\/review\/?$/);
    if (match2) return match2[2];
    const match1 = path.match(/^\/me\/pedidos\/([^/]+)\/review\/?$/);
    if (match1) return match1[1];
    return '';
  }

  // --- ROUTE PARSER ---
  function parseRoute(path) {
    if (path === '/admin' || path === '/admin/') return { view: 'admin', slug: '' };
    const previewMatch = path.match(/^\/me\/my-restaurant\/([^/]+)\/?$/);
    if (previewMatch) return { view: 'menu-preview', slug: previewMatch[1] };
    if (path === '/login') return { view: 'login', slug: '' };
    if (path === '/account' || path === '/account/') return { view: 'account', slug: '' };
    if (path === '/me/my-restaurant' || path === '/me/my-restaurant/') return { view: 'my-restaurants', slug: '' };
    if (path === '/me/pedidos' || path === '/me/pedidos/') return { view: 'orders', slug: '' };
    const reviewWithSlugMatch = path.match(/^\/me\/pedidos\/([^/]+)\/([^/]+)\/review\/?$/);
    if (reviewWithSlugMatch) {
      return { view: 'order-review', slug: reviewWithSlugMatch[1], orderId: reviewWithSlugMatch[2] };
    }
    if (/^\/me\/pedidos\/[^/]+\/review\/?$/.test(path) || path === '/me/pedidos/review' || path === '/me/pedidos/review/' || path === '/order/review' || path === '/order/review/') return { view: 'order-review', slug: '' };
    const orderStatusMatch = path.match(/^\/me\/pedidos\/([^/]+)\/([^/]+)\/?$/);
    if (orderStatusMatch && orderStatusMatch[2] !== 'review') {
      return { view: 'order-status', slug: '', restaurantId: orderStatusMatch[1], orderId: orderStatusMatch[2] };
    }
    if (path === '/order') return { view: 'order', slug: '' };
    if (path === '/checkout') return { view: 'checkout', slug: '' };
    const whatsappMatch = path.match(/^\/pedido-whatsapp\/([^/]+)\/?$/);
    if (whatsappMatch) return { view: 'checkout', slug: '', orderId: whatsappMatch[1] };
    if (path === '/restaurantes') return { view: 'restaurantes', slug: '' };
    if (path === '/cardapio') return { view: 'cardapio', slug: '' };

    const rMatch = path.match(/^\/(?:r|restaurant)\/([^/]+)(?:\/([^/]+))?\/?$/);
    if (rMatch) {
      const [, rSlug, sub] = rMatch;
      if (sub === 'admin') return { view: 'admin', slug: rSlug };
        if (sub === 'dine-in') return { view: 'restaurant', slug: rSlug, dineIn: true };   // NEW
      if (sub === 'cardapio') return { view: 'cardapio', slug: rSlug };
      if (sub === 'delivery') return { view: 'delivery', slug: rSlug };
      return { view: 'restaurant', slug: rSlug };
    }

    return { view: 'home', slug: '' };
  }

  // --- INITIAL ROUTING ---
  const path = typeof window !== 'undefined' ? window.location.pathname : '/';
  const initial = parseRoute(path);
  if (initial.dineIn) enterDineIn(initial.slug);
  let currentView = $state(initial.view);
  let slug = $state(initial.slug);
  let activeRestaurantId = $state(initial.restaurantId ?? '');
  let activeOrderId = $state(initial.orderId ?? '');

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
        leaveDineIn();   // NEW
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
  let reviewPathId = $state(extractReviewOrderId(window.location.pathname));
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
    const isDone = saved.order.status === 'delivered' || saved.order.status === 'ready' || !!saved.pickupAcknowledgedAt || isPaid(saved.order);
    return {
      id: String(saved.order.id), orderNumber: `#${saved.order.id}`,
      date: new Date(saved.order.created_at).toLocaleString('pt-BR'),
      restaurantName: saved.restaurantName,
      status: saved.order.status === 'cancelled' ? 'Cancelado' : isDone ? 'Entregue' : '',
      total: saved.order.total_cents / 100,
      isReviewed: !!saved.isReviewed || !!saved.serviceReviewed,
          items: (saved.order.items || [])
  .filter((item) => !item.voided_at)
  .map((item) => ({
    id: String(item.id),
    name: item.item_name,
    quantity: item.quantity,
    price: item.unit_price_cents / 100,
    imageUrl: saved.itemImages?.[item.item_id] ?? '',
  })),
    };
  }
  let reviewOrder = $derived.by(() => {
    const match = availableOrders.find((o) => String(o.order.id) === reviewPathId);
    return match ? historical(match) : selectedOrderForReview;
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
    const timer = window.setInterval(poll, import.meta.env.TEST_ENV === 'true' || import.meta.env.TEST_ENV === '1' ? 4000 : 15000);
    const notificationTimer = window.setInterval(() => { notificationTime = Date.now(); }, 1000);
    return () => { window.clearInterval(timer); window.clearInterval(notificationTimer); };
  });

  function goToOrders() {
    currentView = 'orders';
    window.history.pushState({}, '', '/me/pedidos');
    window.scrollTo({ top: 0, behavior: 'instant' });
  }

  function goToOrderReview(order: HistoricalOrder, restaurantSlug?: string) {
    selectedOrderForReview = order;
    reviewPathId = order.id;
    currentView = 'order-review';
    const targetSlug = restaurantSlug || slug || 'restaurante';
    window.history.pushState({}, '', `/me/pedidos/${encodeURIComponent(targetSlug)}/${encodeURIComponent(order.id)}/review`);
    window.scrollTo({ top: 0, behavior: 'instant' });
  }

  function goToOrderStatus(restaurantId: string | number, orderId: string | number) {
    activeRestaurantId = String(restaurantId);
    activeOrderId = String(orderId);
    currentView = 'order-status';
    window.history.pushState({}, '', `/me/pedidos/${restaurantId}/${orderId}`);
    window.scrollTo({ top: 0, behavior: 'instant' });
  }

  // --- Browser back/forward ---
  if (typeof window !== 'undefined') {
    window.addEventListener('popstate', () => {
      const parsed = parseRoute(window.location.pathname);
      currentView = parsed.view;
      slug = parsed.slug;
      if (parsed.dineIn) enterDineIn(parsed.slug);
      if (parsed.restaurantId) activeRestaurantId = parsed.restaurantId;
      if (parsed.orderId) activeOrderId = parsed.orderId;
      reviewPathId = extractReviewOrderId(window.location.pathname);
    });
  }

  // --- NAVBAR + THEME BEHAVIOUR ---
  let showNavLinks = $derived(
    !!slug && ['restaurant', 'cardapio', 'delivery', 'menu-preview'].includes(currentView)
  );

  let showOrderCart = $derived(showNavLinks);

  type RestaurantBrand = { name: string; icon_url?: string | null; favicon_url?: string | null; cover_url?: string | null; description?: string | null };
  let restaurantBrand = $state<RestaurantBrand | null>(null);
$effect(() => { if (slug) rememberSlug(slug); });
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
    currentView === 'cardapio' ||
    currentView === 'account' ||
    currentView === 'my-restaurants' ||
    currentView === 'admin' ||
    currentView === 'menu-preview' ||
    currentView === 'orders' ||
    currentView === 'order-review' ||
    currentView === 'order-status' ||
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
              <MenuPage {slug} onBack={goToHome} />
        {:else if currentView === 'delivery'}
          <DeliverySection />
        {:else if currentView === 'order'}
          <OrderSection goToCheckout={goToCheckout} onCancel={goToHome} />
        {:else if currentView === 'checkout'}
          <CheckoutPage onCancel={goToOrder} total={$cart_total} {slug} onOrderConfirmed={goToOrderStatus} />
        {:else if currentView === 'login'}
          <Login onGoHome={() => { if (!$account.isLoggedIn) goToHome(); }} />
        {:else if currentView === 'account'}
          <ProfilePage onBack={goBackFromAccount} onGoLogin={goToAccountLogin} />
        {:else if currentView === 'my-restaurants'}
          <MyRestaurants />
        {:else if currentView === 'menu-preview'}
          <MenuPreview {slug} />
        {:else if currentView === 'orders'}
          <OrderPage onOpenReview={goToOrderReview} />
        {:else if currentView === 'order-status'}
          <OrderStatusPage
            restaurantId={activeRestaurantId}
            orderId={activeOrderId}
            onGoOrders={goToOrders}
            onGoHome={goToHome}
          />
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
  {#if pickupOrder}
    <OrderSuccededPopup
      orderCode={pickupOrder.order.display_code || `#${pickupOrder.order.id}`}
      restaurantName={pickupOrder.restaurantName}
      onConfirm={() => {
        const target = pickupOrder!;
        const slugOrRest = target.slug || target.restaurantId || 'restaurante';
        updateOrder(target.order.id, { pickupAcknowledgedAt: Date.now() });
        goToOrderReview(historical(target), slugOrRest);
      }}
    />
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
