<script lang="ts">
  import { onMount } from 'svelte';
  import { account } from '../lib/stores/account';
  import { getMyRestaurants, type UserRestaurant } from './restaurants';

  let restaurants = $state<UserRestaurant[]>([]);
  let loading = $state(true);
  let showCreate = $state(false);
  let name = $state('');

  onMount(async () => {
    try {
      if ($account.isLoggedIn) restaurants = await getMyRestaurants();
    } finally {
      loading = false;
    }
  });
</script>

<main class="mx-auto max-w-4xl px-6 py-12">
  <a href="/" class="text-sm font-semibold text-amber-700 hover:underline">← Voltar ao início</a>
  <div class="mt-6 flex flex-wrap items-start justify-between gap-4">
    <div>
      <h1 class="font-serif text-3xl font-bold text-stone-900">Meus restaurantes</h1>
      <p class="mt-2 text-stone-500">Restaurantes em que você trabalha e seu status em cada um.</p>
    </div>
    {#if $account.isLoggedIn}
      <button class="rounded-xl bg-amber-500 px-4 py-3 font-semibold text-stone-900 hover:bg-amber-400" onclick={() => showCreate = !showCreate}>Criar restaurante</button>
    {/if}
  </div>

  {#if !$account.isLoggedIn}
    <p class="mt-8 text-stone-600">Entre na sua conta para ver seus restaurantes. <a class="font-semibold text-amber-700 underline" href="/login?next=/me/my-restaurant">Entrar</a></p>
  {:else}
    {#if showCreate}
      <form class="mt-8 rounded-2xl border border-stone-200 bg-white p-6" onsubmit={(event) => event.preventDefault()}>
        <h2 class="text-lg font-bold">Novo restaurante</h2>
        <p class="mt-1 text-sm text-stone-500">Modelo de cadastro: a criação estará disponível em breve.</p>
        <label for="restaurant-name" class="mt-4 block text-sm font-semibold">Nome do restaurante</label>
        <input id="restaurant-name" bind:value={name} class="mt-2 w-full rounded-xl border border-stone-300 px-4 py-2" placeholder="Nome do restaurante" />
        <button type="button" disabled class="mt-4 rounded-xl bg-stone-200 px-4 py-2 font-semibold text-stone-500">Criar em breve</button>
      </form>
    {/if}
    {#if loading}
      <p class="mt-8 text-stone-500">Carregando restaurantes...</p>
    {:else if restaurants.length === 0}
      <p class="mt-8 rounded-2xl border border-stone-200 bg-white p-6 text-stone-500">Nenhum restaurante disponível no momento.</p>
    {:else}
      <div class="mt-8 grid gap-4 sm:grid-cols-2">
        {#each restaurants as restaurant (restaurant.slug)}
          <a href="/r/{encodeURIComponent(restaurant.slug)}/admin" class="rounded-2xl border border-stone-200 bg-white p-6 shadow-sm transition hover:border-amber-400">
            <h2 class="text-lg font-bold text-stone-900">{restaurant.name}</h2>
            <p class="mt-2 text-sm text-stone-600">Função: {restaurant.role}</p>
            <p class="mt-1 text-sm text-stone-600">Status: {restaurant.status}</p>
            <span class="mt-4 inline-block text-sm font-semibold text-amber-700">Acessar painel →</span>
          </a>
        {/each}
      </div>
    {/if}
  {/if}
</main>
