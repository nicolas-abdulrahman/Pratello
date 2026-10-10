<script>
  import Home from './Home.svelte';
  import Restaurant from './Restaurant.svelte';

  const path = typeof window !== 'undefined' ? window.location.pathname : '/';
  const isRestaurantRoute = path.startsWith('/r/');
  const slug = isRestaurantRoute ? path.slice(3).replace(/\/$/, '') : '';
</script>

{#if isRestaurantRoute}
  <Restaurant slug={slug || 'dogao_do_seu_joao'} />
{:else}
  <Home />
{/if}
