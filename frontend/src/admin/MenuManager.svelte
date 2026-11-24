<script lang="ts">
  import { Pencil, Plus, Search, Trash2, X, ImageOff } from 'lucide-svelte';
  import {
    discountPercent,
    formatPrice,
    parsePrice,
    type Category,
    type CategoryId,
    type Product,
    type ProductInput,
    type VisualCategory
  } from './admin';

  // --- Abstracted State & Dependencies ---
  export let products: Product[] = [];
  export let categories: Category[] = [
    { id: 'lanches', name: 'Lanches' },
    { id: 'pratos', name: 'Pratos' },
    { id: 'bebidas', name: 'Bebidas' },
    { id: 'sobremesas', name: 'Sobremesas' },
  ];
  export let visualCategories: VisualCategory[] = [
    { id: 'destaque', name: 'Mais pedidos' },
    { id: 'vegetariano', name: 'Vegetariano' },
    { id: 'sobremesas', name: 'Doces' },
  ];

  export let onSaveProduct: (p: ProductInput) => void = () => {};
  export let onDeleteProduct: (id: string) => void = () => {};
  export let onToggleAvailability: (id: string) => void = () => {};

  interface FormState {
    id?: string;
    name: string;
    description: string;
    price: string;
    originalPrice: string;
    prepTime: string;
    category: CategoryId;
    tags: string[];
    image: string;
    ingredients: string;
    available: boolean;
    popular: boolean;
  }

  const emptyForm: FormState = {
    name: '',
    description: '',
    price: '',
    originalPrice: '',
    prepTime: '20',
    category: 'lanches',
    tags: [],
    image: '',
    ingredients: '',
    available: true,
    popular: false,
  };

  const money = (v?: number) => (v ? v.toFixed(2).replace('.', ',') : '');

  const toForm = (p: Product): FormState => ({
    ...p,
    price: money(p.price),
    originalPrice: money(p.originalPrice),
    prepTime: String(p.prepTime),
    ingredients: p.ingredients ? p.ingredients.join(', ') : '',
  });

  let form: FormState | null = null;
  let error: string = '';
  let query: string = '';
  let selectedCategory: string = 'todas';

  $: filteredList = products.filter(
    (p) =>
      (selectedCategory === 'todas' || p.category === selectedCategory) &&
      p.name.toLowerCase().includes(query.trim().toLowerCase())
  );

  const catName = (id: CategoryId) => categories.find((c) => c.id === id)?.name ?? id;

  function openModal(p?: Product) {
    error = '';
    form = p ? toForm(p) : { ...emptyForm };
  }

  function closeModal() {
    form = null;
    error = '';
  }

  function handleSubmit() {
    if (!form) return;
    const price = parsePrice(form.price);
    const original = form.originalPrice.trim() ? parsePrice(form.originalPrice) : undefined;
    const prepTime = Number(form.prepTime);

    if (!form.name.trim()) {
      error = 'Informe o nome do prato.';
      return;
    }
    if (!Number.isFinite(price) || price <= 0) {
      error = 'Informe um preço válido (ex: 29,90).';
      return;
    }
    if (original !== undefined && (!Number.isFinite(original) || original <= price)) {
      error = 'O preço "de" (promoção) deve ser maior que o preço atual.';
      return;
    }
    if (!Number.isInteger(prepTime) || prepTime <= 0) {
      error = 'Informe o tempo de preparo em minutos.';
      return;
    }

    const input: ProductInput = {
      id: form.id,
      name: form.name.trim(),
      description: form.description.trim(),
      price: Math.round(price * 100) / 100,
      originalPrice: original !== undefined ? Math.round(original * 100) / 100 : undefined,
      prepTime,
      tags: form.tags,
      category: form.category,
      image: form.image.trim(),
      ingredients: form.ingredients
        .split(',')
        .map((s) => s.trim())
        .filter(Boolean),
      available: form.available,
      popular: form.popular,
    };

    onSaveProduct(input);
    closeModal();
  }

  function handleRemove(p: Product) {
    if (confirm(`Excluir "${p.name}" do cardápio?`)) {
      onDeleteProduct(p.id);
    }
  }

  function toggleTag(tagId: string) {
    if (!form) return;
    if (form.tags.includes(tagId)) {
      form.tags = form.tags.filter((t) => t !== tagId);
    } else {
      form.tags = [...form.tags, tagId];
    }
  }
</script>

<div>
  <!-- Header -->
  <div class="flex flex-wrap items-end justify-between gap-3">
    <div>
      <h1 class="text-2xl font-bold text-zinc-900">Gerenciar Cardápio</h1>
      <p class="text-sm text-zinc-600">
        {products.length} pratos · {products.filter((p) => p.available).length} disponíveis
      </p>
    </div>
    <button
      on:click={() => openModal()}
      class="inline-flex items-center gap-2 rounded-xl bg-yellow-400 px-4 py-2.5 text-sm font-semibold text-zinc-900 hover:bg-yellow-500"
    >
      <Plus size={18} /> Novo prato
    </button>
  </div>

  <!-- Filters -->
  <div class="mt-6 flex flex-col gap-3 sm:flex-row">
    <div class="relative flex-1">
      <Search size={16} class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-zinc-400" />
      <input
        type="text"
        bind:value={query}
        placeholder="Buscar prato…"
        class="w-full rounded-xl border border-zinc-200 bg-white py-2.5 pl-9 pr-3.5 text-sm text-zinc-900 placeholder:text-zinc-400 focus:border-yellow-400 focus:outline-none"
      />
    </div>
    <select
      bind:value={selectedCategory}
      class="w-full rounded-xl border border-zinc-200 bg-white px-3.5 py-2.5 text-sm text-zinc-900 focus:border-yellow-400 focus:outline-none sm:w-56"
    >
      <option value="todas">Todas as categorias</option>
      {#each categories as c (c.id)}
        <option value={c.id}>{c.name}</option>
      {/each}
    </select>
  </div>

  <!-- Product list -->
  <div class="mt-4 overflow-hidden rounded-2xl border border-zinc-200 bg-white shadow-sm">
    {#if filteredList.length === 0}
      <p class="p-8 text-center text-sm text-zinc-500">Nenhum prato encontrado.</p>
    {:else}
      <ul class="divide-y divide-zinc-100">
        {#each filteredList as p (p.id)}
          <li class={`flex flex-wrap items-center gap-4 p-4 sm:flex-nowrap ${p.available ? '' : 'bg-zinc-50'}`}>
            {#if p.image}
              <img
                src={p.image}
                alt={p.name}
                class={`h-16 w-16 shrink-0 rounded-xl object-cover ${p.available ? '' : 'opacity-50 grayscale'}`}
              />
            {:else}
              <div class="grid h-16 w-16 shrink-0 place-items-center rounded-xl bg-zinc-200 text-zinc-400">
                <ImageOff size={24} />
              </div>
            {/if}

            <div class="min-w-0 flex-1">
              <p class="flex items-center gap-2 font-semibold text-zinc-900">
                <span class="truncate">{p.name}</span>
                {#if p.popular}
                  <span class="shrink-0 rounded-full bg-yellow-400 px-2 py-0.5 text-[10px] font-bold uppercase">
                    Destaque
                  </span>
                {/if}
                {#if discountPercent(p.price, p.originalPrice) > 0}
                  <span class="shrink-0 rounded-full bg-zinc-900 px-2 py-0.5 text-[10px] font-bold uppercase text-yellow-400">
                    -{discountPercent(p.price, p.originalPrice)}%
                  </span>
                {/if}
              </p>
              <p class="text-sm text-zinc-500">
                {catName(p.category)} · <b class="text-zinc-800">{formatPrice(p.price)}</b>
              </p>
            </div>

            <div class="flex w-full items-center justify-between gap-2 sm:w-auto sm:justify-end">
              <!-- Switch de disponibilidade -->
              <label class="flex cursor-pointer items-center gap-2 text-sm text-zinc-600">
                <button
                  type="button"
                  role="switch"
                  aria-checked={p.available}
                  on:click={() => onToggleAvailability(p.id)}
                  class={`relative h-6 w-11 rounded-full transition-colors ${
                    p.available ? 'bg-yellow-400' : 'bg-zinc-300'
                  }`}
                >
                  <span
                    class={`absolute top-0.5 h-5 w-5 rounded-full bg-white shadow transition-all ${
                      p.available ? 'left-[22px]' : 'left-0.5'
                    }`}
                  />
                </button>
                <span class="w-20">{p.available ? 'Disponível' : 'Indisponível'}</span>
              </label>

              <!-- Ações -->
              <div class="flex gap-1">
                <button
                  on:click={() => openModal(p)}
                  class="rounded-lg p-2 text-zinc-600 hover:bg-zinc-100"
                  aria-label={`Editar ${p.name}`}
                >
                  <Pencil size={16} />
                </button>
                <button
                  on:click={() => handleRemove(p)}
                  class="rounded-lg p-2 text-red-600 hover:bg-red-50"
                  aria-label={`Excluir ${p.name}`}
                >
                  <Trash2 size={16} />
                </button>
              </div>
            </div>
          </li>
        {/each}
      </ul>
    {/if}
  </div>

  <!-- Modal de Criação / Edição -->
  {#if form}
    <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div class="max-h-[90vh] w-full max-w-lg overflow-y-auto rounded-2xl bg-white p-6 shadow-xl">
        <div class="mb-4 flex items-center justify-between">
          <h2 class="text-lg font-bold text-zinc-900">{form.id ? 'Editar prato' : 'Novo prato'}</h2>
          <button on:click={closeModal} class="rounded-lg p-1 text-zinc-400 hover:bg-zinc-100">
            <X size={20} />
          </button>
        </div>

        <form on:submit|preventDefault={handleSubmit} class="space-y-4">
          <label class="block">
            <span class="mb-1.5 block text-sm font-medium text-zinc-700">Nome</span>
            <input
              type="text"
              bind:value={form.name}
              class="w-full rounded-xl border border-zinc-200 bg-white px-3.5 py-2.5 text-sm text-zinc-900 focus:border-yellow-400 focus:outline-none"
            />
          </label>

          <label class="block">
            <span class="mb-1.5 block text-sm font-medium text-zinc-700">Descrição</span>
            <textarea
              rows="2"
              bind:value={form.description}
              class="w-full resize-none rounded-xl border border-zinc-200 bg-white px-3.5 py-2.5 text-sm text-zinc-900 focus:border-yellow-400 focus:outline-none"
            />
          </label>

          <div class="grid grid-cols-2 gap-3">
            <label class="block">
              <span class="mb-1.5 block text-sm font-medium text-zinc-700">Preço (R$)</span>
              <input
                type="text"
                inputmode="decimal"
                bind:value={form.price}
                placeholder="29,90"
                class="w-full rounded-xl border border-zinc-200 bg-white px-3.5 py-2.5 text-sm text-zinc-900 focus:border-yellow-400 focus:outline-none"
              />
            </label>
            <label class="block">
              <span class="mb-1.5 block text-sm font-medium text-zinc-700">Categoria</span>
              <select
                bind:value={form.category}
                class="w-full rounded-xl border border-zinc-200 bg-white px-3.5 py-2.5 text-sm text-zinc-900 focus:border-yellow-400 focus:outline-none"
              >
                {#each categories as c (c.id)}
                  <option value={c.id}>{c.name}</option>
                {/each}
              </select>
            </label>
          </div>

          <div class="grid grid-cols-2 gap-3">
            <label class="block">
              <span class="mb-1.5 block text-sm font-medium text-zinc-700">Preço “de” (opcional)</span>
              <input
                type="text"
                inputmode="decimal"
                bind:value={form.originalPrice}
                placeholder="34,90"
                class="w-full rounded-xl border border-zinc-200 bg-white px-3.5 py-2.5 text-sm text-zinc-900 focus:border-yellow-400 focus:outline-none"
              />
            </label>
            <label class="block">
              <span class="mb-1.5 block text-sm font-medium text-zinc-700">Preparo (min)</span>
              <input
                type="text"
                inputmode="numeric"
                bind:value={form.prepTime}
                class="w-full rounded-xl border border-zinc-200 bg-white px-3.5 py-2.5 text-sm text-zinc-900 focus:border-yellow-400 focus:outline-none"
              />
            </label>
          </div>

          <div>
            <span class="mb-1.5 block text-sm font-medium text-zinc-700">Categorias da vitrine</span>
            <div class="flex flex-wrap gap-1.5">
              {#each visualCategories as c (c.id)}
                {@const on = form.tags.includes(c.id)}
                <button
                  type="button"
                  on:click={() => toggleTag(c.id)}
                  class={`rounded-full px-2.5 py-1 text-xs font-semibold ring-1 ${
                    on
                      ? 'bg-yellow-400 text-zinc-900 ring-yellow-400'
                      : 'bg-white text-zinc-600 ring-zinc-200 hover:ring-zinc-300'
                  }`}
                >
                  {c.name}
                </button>
              {/each}
            </div>
          </div>

          <label class="block">
            <span class="mb-1.5 block text-sm font-medium text-zinc-700">URL da imagem</span>
            <div class="flex gap-3">
              <input
                type="text"
                bind:value={form.image}
                placeholder="https://…"
                class="w-full rounded-xl border border-zinc-200 bg-white px-3.5 py-2.5 text-sm text-zinc-900 focus:border-yellow-400 focus:outline-none"
              />
              {#if form.image}
                <img src={form.image} alt="Prévia" class="h-11 w-11 shrink-0 rounded-lg object-cover" />
              {/if}
            </div>
          </label>

          <label class="block">
            <span class="mb-1.5 block text-sm font-medium text-zinc-700">Ingredientes</span>
            <input
              type="text"
              bind:value={form.ingredients}
              placeholder="queijo, tomate, manjericão"
              class="w-full rounded-xl border border-zinc-200 bg-white px-3.5 py-2.5 text-sm text-zinc-900 focus:border-yellow-400 focus:outline-none"
            />
          </label>

          <div class="flex flex-wrap gap-6">
            <label class="flex items-center gap-2 text-sm text-zinc-700">
              <input type="checkbox" bind:checked={form.available} class="h-4 w-4 accent-yellow-400" />
              Disponível
            </label>
            <label class="flex items-center gap-2 text-sm text-zinc-700">
              <input type="checkbox" bind:checked={form.popular} class="h-4 w-4 accent-yellow-400" />
              Mais pedido (destaque)
            </label>
          </div>

          {#if error}
            <p class="text-sm font-medium text-red-600">{error}</p>
          {/if}

          <div class="flex justify-end gap-2 pt-2">
            <button
              type="button"
              on:click={closeModal}
              class="rounded-xl border border-zinc-200 bg-white px-4 py-2.5 text-sm font-semibold text-zinc-700 hover:bg-zinc-50"
            >
              Cancelar
            </button>
            <button
              type="submit"
              class="rounded-xl bg-yellow-400 px-4 py-2.5 text-sm font-semibold text-zinc-900 hover:bg-yellow-500"
            >
              Salvar
            </button>
          </div>
        </form>
      </div>
    </div>
  {/if}
</div>
