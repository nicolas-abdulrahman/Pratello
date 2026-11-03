<script>
  import Footer from './common/footer.svelte';
  import Navbar from './common/navbar.svelte';
  import OrderCart from './order/OrderCart.svelte'; // The new self-contained cart

  
  import Restaurant from './home/restaurantFrontPage.svelte';
  import MenuPage from './home/Cardapio.svelte'; 
  import Login from './auth/Auth.svelte'; 
  import RestaurantOld from './restaurantPage.svelte';
  import DeliverySection from './delivery/DeliverySection.svelte'; 

  // --- Svelte 5 State ---
  let cartItemCount = $state(2); 
  
  const path = typeof window !== 'undefined' ? window.location.pathname : '/';
  
  // Use $state for the current view
  let currentView = $state('home');
  let slug = $state('');

  // Initial route logic
  if (path === '/login') currentView = 'login';
  else if (path === '/cardapio') currentView = 'cardapio';
  else if (path === '/restaurantes') currentView = 'restaurantes'; 
  else if (path.startsWith('/r/') || path.startsWith('/restaurant/')) {
    currentView = 'restaurant';
    slug = path.startsWith('/r/') ? path.slice(3).replace(/\/$/, '') : path.slice(12).replace(/\/$/, '');
  }

  // --- Navigation Functions ---
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

  // Handle browser Back/Forward
  if (typeof window !== 'undefined') {
    window.addEventListener('popstate', () => {
      const newPath = window.location.pathname;
      if (newPath === '/login') currentView = 'login';
      else if (newPath === '/cardapio') currentView = 'cardapio';
      else if (newPath === '/restaurantes') currentView = 'restaurantes';
      else if (newPath.startsWith('/r/') || newPath.startsWith('/restaurant/')) {
        currentView = 'restaurant';
        slug = newPath.startsWith('/r/') ? newPath.slice(3).replace(/\/$/, '') : newPath.slice(12).replace(/\/$/, '');
      } else {
        currentView = 'home';
      }
    });
  }
</script>

<div class="relative w-full min-h-screen bg-[#fff9f5]">
  
  <!-- GLOBAL NAVBAR -->
  <Navbar 
    {cartItemCount}
    onOpenMenu={goToCardapio} 
    onGoHome={goToHome} 
    onGoLogin={goToLogin} 
    onGoRestaurantes={goToRestaurantes} 
  />

  <!-- GLOBAL CART (Handles its own Sidebar now) -->
  {#if currentView !== 'login'}
    <OrderCart cartItemCount={cartItemCount} />
  {/if}

  <!-- PAGE CONTENT SWAPPER -->
  <div class="relative overflow-hidden w-full">
    {#key currentView}
      <div class="animate-slide-in w-full">
        
        {#if currentView === 'home'}
          <Restaurant />
        {:else if currentView === 'cardapio'}
          <MenuPage />
        {:else if currentView === 'restaurantes'}
          <DeliverySection />
        {:else if currentView === 'login'}
          <Login />
        {:else if currentView === 'restaurant'}
          <RestaurantOld restaurantName={slug || 'mamma-mia-mario'} />
        {/if}
        
      </div>
    {/key}
  </div>
  
  <Footer/>

</div>

<style>
  .animate-slide-in {
    animation: slideIn 0.4s cubic-bezier(0.25, 1, 0.5, 1) forwards;
  }
  @keyframes slideIn {
    0% { opacity: 0; transform: translateY(20px); }
    100% { opacity: 1; transform: translateY(0); }
  }
</style>
