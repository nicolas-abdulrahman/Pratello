<script>
  import Footer from './common/footer.svelte';
  import Navbar from './common/navbar.svelte';
  import OrderCart from './order/OrderCart.svelte';
  import OrderSection from './order/OrderSection.svelte';
  import CheckoutSection from './payment/CheckoutSection.svelte';

  import { cart } from './lib/stores/cart.js';
  import { account } from './lib/stores/account.js';

  import Restaurant from './home/FrontPage.svelte';
  import MenuPage from './menu/Cardapio.svelte';
  import Login from './auth/Auth.svelte';
  import RestaurantsPage from './restaurants/RestaurantsPage.svelte';
  import DeliverySection from './delivery/DeliverySection.svelte';

  // --- CART (store → derived count for the navbar badge) ---
  let cartItemCount = $derived(
    $cart.reduce((total, item) => total + (item.quantity || 1), 0)
  );

  // --- ROUTING ---
  const path = typeof window !== 'undefined' ? window.location.pathname : '/';
  let currentView = $state('home');
  let slug = $state('');

  if (path === '/login') currentView = 'login';
  else if (path === '/cardapio') currentView = 'cardapio';
  else if (path === '/restaurantes') currentView = 'restaurantes';
  else if (path === '/order') currentView = 'order';
  else if (path === '/checkout') currentView = 'checkout';
  else if (path === '/home' || path === '/frontpage' || path.startsWith('/r/') || path.startsWith('/restaurant/')) {
    currentView = 'restaurant';
    slug = path.startsWith('/r/')
      ? path.slice(3).replace(/\/$/, '')
      : path.startsWith('/restaurant/')
      ? path.slice(12).replace(/\/$/, '')
      : '';
  }

  function goToHome() {
    currentView = 'home';
    window.history.pushState({}, '', '/');
  }
  function goToCardapio() {
    currentView = 'cardapio';
    window.history.pushState({}, '', '/cardapio');
  }
  function goToRestaurantes() {
    currentView = 'restaurantes';
    window.history.pushState({}, '', '/restaurantes');
  }
  function goToLogin() {
    currentView = 'login';
    window.history.pushState({}, '', '/login');
  }
  function goToOrder() {
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
      const newPath = window.location.pathname;
      if (newPath === '/login') currentView = 'login';
      else if (newPath === '/cardapio') currentView = 'cardapio';
      else if (newPath === '/restaurantes') currentView = 'restaurantes';
      else if (newPath === '/order') currentView = 'order';
      else if (newPath === '/checkout') currentView = 'checkout';
      else if (
        newPath === '/home' ||
        newPath === '/frontpage' ||
        newPath.startsWith('/r/') ||
        newPath.startsWith('/restaurant/')
      ) {
        currentView = 'restaurant';
        slug = newPath.startsWith('/r/')
          ? newPath.slice(3).replace(/\/$/, '')
          : newPath.startsWith('/restaurant/')
          ? newPath.slice(12).replace(/\/$/, '')
          : '';
      } else {
        currentView = 'home';
      }
    });
  }

  // --- NAVBAR BEHAVIOUR ---
  // Início / Cardápio / Delivery appear ONLY on /r/... (restaurant detail)
  let showNavLinks = $derived(currentView === 'restaurant');

  // Brand theme for RestaurantsPage, restaurantTheme for Restaurant
  let brandTheme = $derived(currentTheme === 'brand');

  // Hide global chrome on full-screen flows
  let hideGlobalUI = $derived(
    currentView === 'login' ||
    currentView === 'order' ||
    currentView === 'checkout'
  );

  // Dynamic theme: brand for RestaurantsPage, restaurantTheme for Restaurant
  let currentTheme = $derived(
    currentView === 'restaurant' ? 'restaurantTheme' : 'brand'
  );

  // Redirect to home if the user lands on /login while already authenticated.
  // (Optional but useful — remove if you'd rather let logged-in users hit /login.)
  $effect(() => {
    if (currentView === 'login' && $account.isLoggedIn) {
      goToHome();
    }
  });
</script>

<div data-theme={currentTheme} class="relative w-full min-h-screen bg-background text-text">
  {#if !hideGlobalUI}
    <Navbar
      {cartItemCount}
      onOpenMenu={goToCardapio}
      onGoHome={goToHome}
      onGoLogin={goToLogin}
      onGoRestaurantes={goToRestaurantes}
      {showNavLinks}
      brand={brandTheme}
    />
    <OrderCart goToOrder={goToOrder} />
  {/if}

  <div class="relative overflow-hidden w-full">
    {#key currentView}
      <div class="animate-slide-in w-full">
        {#if currentView === 'home'}
          <RestaurantsPage />
        {:else if currentView === 'restaurantes'}
          <RestaurantsPage />
        {:else if currentView === 'restaurant'}
          <Restaurant />
        {:else if currentView === 'cardapio'}
          <MenuPage />
        {:else if currentView === 'delivery'}
          <DeliverySection />
        {:else if currentView === 'order'}
          <OrderSection goToCheckout={goToCheckout} onCancel={goToHome} />
        {:else if currentView === 'checkout'}
          <CheckoutSection onCancel={goToOrder} total={150.0} />
        {:else if currentView === 'login'}
          <Login />
        {/if}
      </div>
    {/key}
  </div>

  {#if !hideGlobalUI}
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
