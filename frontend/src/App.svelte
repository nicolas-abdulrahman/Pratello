<script>
  import Home from './Home.svelte';
  import Restaurant from './Restaurant.svelte';
  import RestaurantPage from './restaurantPage.svelte';

  const path = typeof window !== 'undefined' ? window.location.pathname : '/';
  const isRestaurantRoute = path.startsWith('/r/') || path.startsWith('/restaurant');
  const slug = path.startsWith('/r/')
    ? path.slice(3).replace(/\/$/, '')
    : path.startsWith('/restaurant/')
    ? path.slice(12).replace(/\/$/, '')
    : '';
</script>
{#if isRestaurantRoute}
  <RestaurantPage restaurantName={slug || 'mamma-mia-mario'} />
{:else}
  <Home />
{/if}
