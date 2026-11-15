<script lang="ts">
  import type { CartItem } from '../lib/stores/Cart';
  import { calc_line_price } from '../lib/stores/Cart';

  let {
    line,
    onRemove = () => {},
  }: {
    line: CartItem;
    onRemove?: (id: number) => void;
  } = $props();

  const menu = $derived(line.item);
  const unitPrice = $derived(calc_line_price({ ...line, quantity: 1 }));
  const linePrice = $derived(calc_line_price(line));
  const isCombo = $derived(menu.type === 'combo');

  const selectedExtras = $derived(menu.extras.filter((e) => e.selected));
  const removedIngredients = $derived(menu.ingredients.filter((i) => !i.selected));

  function formatBRL(v: number): string {
    return new Intl.NumberFormat('pt-BR', {
      style: 'currency',
      currency: 'BRL',
    }).format(v);
  }
</script>

<div class="sidebar-item">
  <!-- Remove button -->
  <button
    onclick={() => onRemove(line.id)}
    class="remove-btn"
    aria-label="Remover item"
  >
    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
      <path d="M18 6 6 18" />
      <path d="m6 6 12 12" />
    </svg>
  </button>

  <!-- Image -->
  <div class="thumb">
    <img
      src={menu.imageUrl}
      alt={menu.name}
      class="w-full h-full object-cover"
    />

    {#if line.quantity > 1}
      <div class="qty-badge">x{line.quantity}</div>
    {/if}
  </div>

  <!-- Content -->
  <div class="content">
    <div class="flex justify-between items-start gap-2 mb-1">
      <div class="min-w-0">
        <h3 class="title">{menu.name}</h3>

        {#if isCombo}
          <span class="combo-badge">Combo</span>
        {/if}

        <p class="unit-price">
          {formatBRL(unitPrice)}{line.quantity > 1 ? ` × ${line.quantity}` : ''}
        </p>
      </div>

      <span class="line-price">{formatBRL(linePrice)}</span>
    </div>

    <!-- Combo components -->
    {#if isCombo && menu.components?.length}
      <ul class="sub-list">
        {#each menu.components as comp}
          <li class="sub-item">
            <span class="dot"></span>
            <span class="truncate">
              {comp.quantity > 1 ? `${comp.quantity}× ` : ''}{comp.name}
            </span>
          </li>
        {/each}
      </ul>
    {/if}

    <!-- Selected extras -->
    {#if selectedExtras.length > 0}
      <ul class="sub-list">
        {#each selectedExtras as ex}
          <li class="sub-item">
            <span class="dot dot--accent"></span>
            <span class="truncate">
              + {ex.name}
              <span class="text-text-muted">({formatBRL(ex.price)})</span>
            </span>
          </li>
        {/each}
      </ul>
    {/if}

    <!-- Removed ingredients (shown as strikethrough) -->
    {#if removedIngredients.length > 0}
      <ul class="sub-list">
        {#each removedIngredients as ing}
          <li class="sub-item sub-item--removed">
            <span class="dot dot--removed"></span>
            <span class="truncate">Sem {ing.name}</span>
          </li>
        {/each}
      </ul>
    {/if}

    {#if line.notes}
      <p class="notes">“{line.notes}”</p>
    {/if}
  </div>
</div>

<style>
  @reference "../app.css";

  .sidebar-item {
    @apply relative flex gap-4 p-3 rounded-2xl border transition-shadow;
    background-color: var(--color-surface);
    border-color: var(--color-border);
    box-shadow: 0 1px 2px 0 rgb(0 0 0 / 0.05);
  }
  .sidebar-item:hover {
    box-shadow: 0 4px 6px -1px rgb(0 0 0 / 0.08);
  }

  .remove-btn {
    @apply absolute -top-2.5 -right-2.5 w-6 h-6 rounded-full flex items-center justify-center transition-colors z-10;
    background-color: var(--color-surface-hover);
    color: var(--color-text-muted);
    box-shadow: 0 1px 2px 0 rgb(0 0 0 / 0.08);
  }
  .remove-btn:hover {
    background-color: color-mix(in oklab, var(--color-primary) 12%, transparent);
    color: var(--color-primary);
  }

  .thumb {
    @apply w-20 h-20 shrink-0 rounded-xl overflow-hidden relative;
    background-color: var(--color-surface-hover);
  }

  .qty-badge {
    @apply absolute bottom-0 right-0 text-[10px] font-bold px-1.5 py-0.5 rounded-tl-md text-white;
    background-color: var(--color-primary);
  }

  .content {
    @apply flex flex-col flex-1 min-w-0 py-1;
  }

  .title {
    @apply font-bold text-base leading-tight truncate;
    color: var(--color-text);
  }

  .combo-badge {
    @apply inline-block text-[9px] font-bold px-1.5 py-0.5 rounded mt-1 uppercase tracking-wider;
    background-color: var(--color-primary-soft);
    color: var(--color-primary-hover);
  }

  .unit-price {
    @apply text-[11px] mt-1;
    color: var(--color-text-muted);
  }

  .line-price {
    @apply font-bold text-sm whitespace-nowrap;
    color: var(--color-primary);
  }

  .sub-list {
    @apply space-y-0.5 mt-1.5;
  }

  .sub-item {
    @apply text-xs flex items-center gap-1.5;
    color: var(--color-text-muted);
  }
  .sub-item--removed {
    @apply line-through opacity-70;
  }

  .dot {
    @apply w-1 h-1 rounded-full shrink-0;
    background-color: var(--color-border);
  }
  .dot--accent {
    background-color: var(--color-primary);
  }
  .dot--removed {
    background-color: var(--color-text-muted);
  }

  .notes {
    @apply text-[11px] italic mt-1.5;
    color: var(--color-text-muted);
  }
</style>
