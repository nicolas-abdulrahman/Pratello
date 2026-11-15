<script lang="ts">
  import type { CartItem } from '../lib/stores/Cart';
  import {
    calc_line_price,
    toggle_ingredient,
    toggle_extra,
    update_quantity,
    remove_item,
  } from '../lib/stores/Cart';
  import IngredientItem from './IngredientItem.svelte';

  let { line }: { line: CartItem } = $props();

  const menu = $derived(line.item);
  const unitPrice = $derived(calc_line_price({ ...line, quantity: 1 }));
  const linePrice = $derived(calc_line_price(line));
  const isCombo = $derived(menu.type === 'combo');

  function formatBRL(v: number): string {
    return new Intl.NumberFormat('pt-BR', {
      style: 'currency',
      currency: 'BRL',
    }).format(v);
  }
</script>

<article class="order-card">
  <!-- Header: image · name · price -->
  <div class="card-header">
    <div class="thumb">
      <img src={menu.imageUrl} alt={menu.name} />
    </div>
    <div class="min-w-0 flex-1">
      <h2 class="item-name">{menu.name}</h2>
      {#if isCombo}
        <span class="combo-badge">Combo</span>
      {/if}
      <p class="unit-price">
        {formatBRL(unitPrice)}{line.quantity > 1 ? ` × ${line.quantity}` : ''}
      </p>
    </div>
    <div class="line-price">{formatBRL(linePrice)}</div>
  </div>

  <!-- Combo composition (informational) -->
  {#if isCombo && menu.components?.length}
    <section class="section">
      <h3 class="section-title">Composição</h3>
      <ul class="component-list">
        {#each menu.components as comp}
          <li>
            <span class="component-qty">{comp.quantity}×</span>
            <span class="truncate">{comp.name}</span>
          </li>
        {/each}
      </ul>
    </section>
  {/if}

  <!-- Ingredients (toggle to remove) -->
  {#if menu.ingredients.length > 0}
    <section class="section">
      <h3 class="section-title">Ingredientes</h3>
      <div class="chip-list">
        {#each menu.ingredients as ing (ing.id)}
          <IngredientItem
            label={ing.name}
            isRemoved={!ing.selected}
            onToggle={() => toggle_ingredient(line.id, ing.id)}
          />
        {/each}
      </div>
    </section>
  {/if}

  <!-- Extras (checkbox + price) -->
  {#if menu.extras.length > 0}
    <section class="section">
      <h3 class="section-title">Adicionar extras</h3>
      <ul class="extra-list">
        {#each menu.extras as ex (ex.id)}
          <li>
            <label class="extra" class:is-selected={ex.selected}>
              <input
                type="checkbox"
                checked={ex.selected}
                onchange={() => toggle_extra(line.id, ex.id)}
                class="extra-checkbox"
              />
              <span class="extra-name">{ex.name}</span>
              <span class="extra-price">+ {formatBRL(ex.price)}</span>
            </label>
          </li>
        {/each}
      </ul>
    </section>
  {/if}

  {#if line.notes}
    <p class="notes">"{line.notes}"</p>
  {/if}

  <!-- Footer: quantity · remove -->
  <div class="card-footer">
    <div class="qty-controls">
      <button
        type="button"
        onclick={() => update_quantity(line.id, line.quantity - 1)}
        class="qty-btn"
        aria-label="Diminuir quantidade"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round">
          <line x1="5" y1="12" x2="19" y2="12" />
        </svg>
      </button>
      <span class="qty-value">{line.quantity}</span>
      <button
        type="button"
        onclick={() => update_quantity(line.id, line.quantity + 1)}
        class="qty-btn"
        aria-label="Aumentar quantidade"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round">
          <line x1="12" y1="5" x2="12" y2="19" />
          <line x1="5" y1="12" x2="19" y2="12" />
        </svg>
      </button>
    </div>

    <button
      type="button"
      onclick={() => remove_item(line.id)}
      class="remove-btn"
      aria-label="Remover item"
    >
      <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
        <path d="M3 6h18" />
        <path d="M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
        <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6" />
        <line x1="10" y1="11" x2="10" y2="17" />
        <line x1="14" y1="11" x2="14" y2="17" />
      </svg>
      Remover
    </button>
  </div>
</article>

<style>
  @reference "../app.css";

  .order-card {
    @apply bg-surface rounded-3xl p-5 shadow-sm border border-border;
  }

  /* Header */
  .card-header {
    @apply flex gap-4 items-start mb-4 pb-4 border-b border-border;
  }
  .thumb {
    @apply w-16 h-16 rounded-xl overflow-hidden shrink-0 bg-surface-hover;
  }
  .thumb img {
    @apply w-full h-full object-cover;
  }
  .item-name {
    @apply text-lg font-bold text-text leading-tight;
  }
  .combo-badge {
    @apply inline-block text-[9px] font-bold px-1.5 py-0.5 rounded mt-1 uppercase tracking-wider bg-primary-soft text-primary-hover;
  }
  .unit-price {
    @apply text-xs text-text-muted mt-1;
  }
  .line-price {
    @apply text-lg font-black text-primary whitespace-nowrap;
  }

  /* Sections */
  .section {
    @apply mb-4;
  }
  .section-title {
    @apply text-xs font-bold uppercase tracking-wider text-text-muted mb-2;
  }

  /* Combo composition */
  .component-list {
    @apply space-y-1;
  }
  .component-list li {
    @apply text-sm text-text flex items-center gap-2;
  }
  .component-qty {
    @apply text-xs font-bold text-primary shrink-0;
  }

  /* Ingredient chips */
  .chip-list {
    @apply flex flex-wrap gap-2;
  }

  /* Extras */
  .extra-list {
    @apply space-y-1.5;
  }
  .extra {
    @apply flex items-center gap-2 p-2 rounded-lg border border-border bg-surface cursor-pointer hover:bg-surface-hover transition-colors;
  }
  .extra.is-selected {
    @apply border-primary/30 bg-primary-soft/40;
  }
  .extra-checkbox {
    @apply w-4 h-4 rounded text-primary focus:ring-primary cursor-pointer;
  }
  .extra-name {
    @apply flex-1 text-sm font-medium text-text truncate;
  }
  .extra-price {
    @apply text-sm font-bold text-primary whitespace-nowrap;
  }

  /* Notes */
  .notes {
    @apply text-xs italic text-text-muted mb-4 p-2 rounded-lg bg-surface-hover;
  }

  /* Footer */
  .card-footer {
    @apply flex items-center justify-between gap-3 pt-3 border-t border-border;
  }
  .qty-controls {
    @apply flex items-center gap-2 bg-surface-hover rounded-full p-1;
  }
  .qty-btn {
    @apply w-7 h-7 rounded-full bg-surface text-text hover:bg-primary hover:text-white transition-colors flex items-center justify-center;
  }
  .qty-value {
    @apply w-6 text-center text-sm font-bold text-text;
  }
  .remove-btn {
    @apply flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold text-red-600 hover:bg-red-50 transition-colors;
  }
</style>
