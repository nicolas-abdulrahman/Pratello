<script>
  import Footer from './common/footer.svelte';
  import Navbar from './common/navbar.svelte';
  import OrderCart from './order/OrderCart.svelte'; 
  import OrderSection from './order/OrderSection.svelte'; 
  import CheckoutSection from './payment/CheckoutSection.svelte'; // 👈 Import Checkout
  
  import { cart } from './lib/stores/cart.js';

  import Restaurant from './home/FrontPage.svelte';
  import MenuPage from './menu/Cardapio.svelte'; 
  import Login from './auth/Auth.svelte'; 
  import RestaurantOld from './restaurantPage.svelte';
  import DeliverySection from './delivery/DeliverySection.svelte'; 

  // --- GLOBAL CART STATE ---
  let cartItemCount = $derived($cart.reduce((total, item) => total + (item.quantity || 1), 0));

  // --- ROUTING STATE ---
  const path = typeof window !== 'undefined' ? window.location.pathname : '/';
  let currentView = $state('home');
  let slug = $state('');

  if (path === '/login') currentView = 'login';
  else if (path === '/cardapio') currentView = 'cardapio';
  else if (path === '/restaurantes') currentView = 'restaurantes'; 
  else if (path === '/order') currentView = 'order'; 
  else if (path === '/checkout') currentView = 'checkout'; // 👈 Added checkout route
  else if (path.startsWith('/r/') || path.startsWith('/restaurant/')) {
    currentView = 'restaurant';
    slug = path.startsWith('/r/') ? path.slice(3).replace(/\/$/, '') : path.slice(12).replace(/\/$/, '');
  }

  // --- Navigation Functions ---
  function goToHome() { currentView = 'home'; window.history.pushState({}, '', '/'); }
  function goToCardapio() { currentView = 'cardapio'; window.history.pushState({}, '', '/cardapio'); }
  function goToRestaurantes() { currentView = 'restaurantes'; window.history.pushState({}, '', '/restaurantes'); }
  function goToLogin() { currentView = 'login'; window.history.pushState({}, '', '/login'); }
  
  function goToOrder() { 
    currentView = 'order'; 
    window.history.pushState({}, '', '/order'); 
  }

  // 👈 Fix: Now it actually navigates to checkout
  function goToCheckout() {
    currentView = 'checkout';
    window.history.pushState({}, '', '/checkout');
  }

  // Handle browser Back/Forward
  if (typeof window !== 'undefined') {
    window.addEventListener('popstate', () => {
      const newPath = window.location.pathname;
      if (newPath === '/login') currentView = 'login';
      else if (newPath === '/cardapio') currentView = 'cardapio';
      else if (newPath === '/restaurantes') currentView = 'restaurantes';
      else if (newPath === '/order') currentView = 'order';
      else if (newPath === '/checkout') currentView = 'checkout'; // 👈
      else if (newPath.startsWith('/r/') || newPath.startsWith('/restaurant/')) {
        currentView = 'restaurant';
        slug = newPath.startsWith('/r/') ? newPath.slice(3).replace(/\/$/, '') : newPath.slice(12).replace(/\/$/, '');
      } else {
        currentView = 'home';
      }
    });
  }

  // Helper to hide global UI on specific pages
  let hideGlobalUI = $derived(
    currentView === 'login' || 
    currentView === 'order' || 
    currentView === 'checkout' // 👈 Hide navbar on checkout too
  );
</script>

<div class="relative w-full min-h-screen bg-[#fff9f5]">
  
  {#if !hideGlobalUI}
    <Navbar 
      {cartItemCount}
      onOpenMenu={goToCardapio} 
      onGoHome={goToHome} 
      onGoLogin={goToLogin} 
      onGoRestaurantes={goToRestaurantes} 
    />
    <OrderCart goToOrder={goToOrder} />
  {/if}

  <div class="relative overflow-hidden w-full">
    {#key currentView}
      <div class="animate-slide-in w-full">
        {#if currentView === 'home'}
          <Restaurant />
        {:else if currentView === 'cardapio'}
          <MenuPage />
        {:else if currentView === 'restaurantes'}
          <DeliverySection />
        {:else if currentView === 'order'}
          <OrderSection 
            goToCheckout={goToCheckout} 
            onCancel={goToHome} 
          />
        {:else if currentView === 'checkout'}
          <CheckoutSection 
            onCancel={goToOrder} 
            total={150.00} 
          />
        {:else if currentView === 'login'}
          <Login />
        {:else if currentView === 'restaurant'}
          <RestaurantOld restaurantName={slug || 'mamma-mia-mario'} />
        {/if}
      </div>
    {/key}
  </div>
  
  {#if !hideGlobalUI}
    <Footer/>
  {/if}
</div>

<style>
  .animate-slide-in { animation: slideIn 0.4s cubic-bezier(0.25, 1, 0.5, 1) forwards; }
  @keyframes slideIn {
    0% { opacity: 0; transform: translateY(20px); }
    100% { opacity: 1; transform: translateY(0); }
  }
</style>
