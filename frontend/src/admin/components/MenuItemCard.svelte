<script lang="ts">
  import { ImageOff, Pencil, Trash2 } from 'lucide-svelte';

  export interface Product {
    id: string;
    name: string;
    description?: string;
    price: number;
    originalPrice?: number;
    category: string;
    image?: string;
    available: boolean;
    popular?: boolean;
    [key: string]: any;
  }

  interface Props {
    product: Product;
    categoryName?: string;
    onToggleAvailability?: (productId: string) => void;
    onEdit?: (product: Product) => void;
    onDelete?: (product: Product) => void;
    formatPrice?: (price: number) => string;
    discountPercent?: (price: number, original?: number) => number;
  }

  let {
    product,
    categoryName = '',
    onToggleAvailability = () => {},
    onEdit = () => {},
    onDelete = () => {},
    formatPrice = (p: number) =>
      new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(p),
    discountPercent = (price: number, original?: number) =>
      original && original > price ? Math.round((1 - price / original) * 100) : 0,
  }: Props = $props();

  let failedImage = $state(false);

  let discount = $derived(discountPercent(product.price, product.originalPrice));
</script>

<li class={`flex flex-wrap items-center gap-4 p-4 sm:flex-nowrap ${product.available ? '' : 'bg-zinc-50'}`}>
  {#if !product.image || failedImage}
    <div class="grid h-16 w-16 shrink-0 place-items-center rounded-xl bg-zinc-200 text-zinc-400">
      <ImageOff size={24} />
    </div>
  {:else}
    <img
      src={product.image}
      alt={product.name}
      loading="lazy"
      onerror={() => (failedImage = true)}
      class={`h-16 w-16 shrink-0 rounded-xl object-cover ${product.available ? '' : 'opacity-50 grayscale'}`}
    />
  {/if}

  <div class="min-w-0 flex-1">
    <p class="flex items-center gap-2 font-semibold text-zinc-900">
      <span class="truncate">{product.name}</span>
      {#if product.popular}
        <span class="shrink-0 rounded-full bg-yellow-400 px-2 py-0.5 text-[10px] font-bold uppercase">
          Destaque
        </span>
      {/if}
      {#if discount > 0}
        <span class="shrink-0 rounded-full bg-zinc-900 px-2 py-0.5 text-[10px] font-bold uppercase text-yellow-400">
          -{discount}%
        </span>
      {/if}
    </p>
    <p class="text-sm text-zinc-500">
      {categoryName || product.category} · <b class="text-zinc-800">{formatPrice(product.price)}</b>
    </p>
  </div>

  <div class="flex w-full items-center justify-between gap-2 sm:w-auto sm:justify-end">
    <!-- Switch de disponibilidade -->
    <label class="flex cursor-pointer items-center gap-2 text-sm text-zinc-600">
      <button
        type="button"
        role="switch"
        aria-checked={product.available}
        onclick={() => onToggleAvailability(product.id)}
        class={`relative h-6 w-11 rounded-full transition-colors ${
          product.available ? 'bg-yellow-400' : 'bg-zinc-300'
        }`}
      >
        <span
          class={`absolute top-0.5 h-5 w-5 rounded-full bg-white shadow transition-all ${
            product.available ? 'left-[22px]' : 'left-0.5'
          }`}
        />
      </button>
      <span class="w-20">{product.available ? 'Disponível' : 'Indisponível'}</span>
    </label>

    <!-- Ações -->
    <div class="flex gap-1">
      <button
        type="button"
        onclick={() => onEdit(product)}
        class="rounded-lg p-2 text-zinc-600 hover:bg-zinc-100"
        aria-label={`Editar ${product.name}`}
      >
        <Pencil size={16} />
      </button>
      <button
        type="button"
        onclick={() => onDelete(product)}
        class="rounded-lg p-2 text-red-600 hover:bg-red-50"
        aria-label={`Excluir ${product.name}`}
      >
        <Trash2 size={16} />
      </button>
    </div>
  </div>
</li>
