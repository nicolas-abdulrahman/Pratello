<script>
  import { onMount } from 'svelte';

  export let slug = 'dogao_do_seu_joao';

  let dishes = [];
  let loading = true;
  let error = '';

  let newDishName = '';
  let newDishPrice = '';
  let adding = false;

  async function fetchDishes() {
    try {
      loading = true;
      error = '';
      const res = await fetch(`/api/restaurants/${slug}/dishes`);
      if (!res.ok) {
        throw new Error(`Erro ao carregar pratos: ${res.statusText}`);
      }
      dishes = await res.json();
    } catch (err) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  async function handleAddDish(e) {
    e.preventDefault();
    if (!newDishName.trim() || !newDishPrice) return;

    try {
      adding = true;
      const res = await fetch(`/api/restaurants/${slug}/dishes`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: newDishName.trim(),
          price: parseFloat(newDishPrice),
        }),
      });

      if (!res.ok) {
        throw new Error('Falha ao adicionar prato');
      }

      newDishName = '';
      newDishPrice = '';
      await fetchDishes();
    } catch (err) {
      alert(err.message);
    } finally {
      adding = false;
    }
  }

  onMount(() => {
    fetchDishes();
  });
</script>

<main class="min-h-screen bg-gray-50 text-gray-900 py-12 px-4 sm:px-6 lg:px-8">
  <div class="max-w-md mx-auto bg-white rounded-xl shadow-md p-6 border border-gray-100">
    <!-- Header -->
    <div class="mb-6 border-b pb-4">
      <h1 class="text-2xl font-bold text-gray-800">🌭 Aqui é o dogão do seu joão</h1>
      <p class="text-sm text-gray-500 mt-1">Cardápio oficial do restaurante</p>
    </div>

    <!-- Dish List -->
    <section class="mb-8">
      <h2 class="text-lg font-semibold text-gray-700 mb-3">Pratos do Cardápio</h2>
      {#if loading}
        <p class="text-gray-400 text-sm">Carregando pratos...</p>
      {:else if error}
        <div class="p-3 bg-red-50 text-red-600 rounded-lg text-sm">{error}</div>
      {:else if dishes.length === 0}
        <p class="text-gray-400 text-sm">Nenhum prato cadastrado ainda.</p>
      {:else}
        <ul class="divide-y divide-gray-100">
          {#each dishes as dish (dish.id)}
            <li class="py-2.5 flex justify-between items-center">
              <span class="font-medium text-gray-800 capitalize">{dish.name}</span>
              <span class="text-green-700 font-semibold bg-green-50 px-2 py-0.5 rounded text-sm">
                R$ {Number(dish.price).toFixed(2)}
              </span>
            </li>
          {/each}
        </ul>
      {/if}
    </section>

    <!-- Add Dish Form -->
    <section class="border-t pt-4">
      <h2 class="text-lg font-semibold text-gray-700 mb-3">Adicionar Novo Prato</h2>
      <form on:submit={handleAddDish} class="space-y-3">
        <div>
          <label for="name" class="block text-xs font-medium text-gray-600 mb-1">Nome do prato</label>
          <input
            id="name"
            type="text"
            bind:value={newDishName}
            placeholder="Ex: Pastel, X-Tudo..."
            required
            class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-amber-500"
          />
        </div>

        <div>
          <label for="price" class="block text-xs font-medium text-gray-600 mb-1">Preço (R$)</label>
          <input
            id="price"
            type="number"
            step="0.01"
            min="0"
            bind:value={newDishPrice}
            placeholder="0.00"
            required
            class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-amber-500"
          />
        </div>

        <button
          type="submit"
          disabled={adding}
          class="w-full bg-amber-500 hover:bg-amber-600 text-white font-medium py-2 rounded-lg text-sm transition-colors disabled:opacity-50"
        >
          {adding ? 'Adicionando...' : '+ Adicionar prato'}
        </button>
      </form>
    </section>
  </div>
</main>
