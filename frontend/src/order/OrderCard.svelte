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

  let {
    line,
    item: legacyItem,
    onRemove
  }: {
    line?: CartItem;
    item?: any;
    onRemove?: (id: number | string) => void;
  } = $props();

  const activeLine = $derived(
    line ||
      (legacyItem
        ? {
            id: legacyItem.id || 1,
            item: legacyItem.item || legacyItem,
            quantity: legacyItem.quantity || 1,
            notes: legacyItem.notes,
          }
        : null)
  );

  const menu = $derived(activeLine?.item || {});
  const isCombo = $derived(menu.type === 'combo' || menu.isCombo || Boolean(menu.combo));

  // ─────────────────────────────────────────────────────────────
  // Customization & Chain States (from original OrderCard)
  // ─────────────────────────────────────────────────────────────
  let isChained = $state(true);
  let isDrinkChained = $state(true);

  let extraMayo = $state(false);
  let extraMeat1 = $state(false);
  let extraMeat2 = $state(false);

  let selectedDrink1 = $state('d-coke');
  let selectedDrink2 = $state('d-coke');

  let removedState = $state<Record<string, boolean>>({});

  // Default burger options if not given in menu
  const defaultBurgerIngredients = [
    { id: 'pao', name: 'Pão Brioche' },
    { id: 'carne', name: 'Carne Bovina 180g' },
    { id: 'queijo', name: 'Queijo Cheddar' },
    { id: 'alface', name: 'Alface Americana' },
    { id: 'tomate', name: 'Tomate' },
    { id: 'bacon', name: 'Bacon' },
    { id: 'molho', name: 'Molho Especial' }
  ];

  const burgerIngredients = $derived(
    menu.ingredients && menu.ingredients.length > 0
      ? menu.ingredients.map((i: any) => ({ id: String(i.id), name: i.name }))
      : defaultBurgerIngredients
  );

  const drinkChoices = [
    { id: 'd-coke', name: 'Coca-Cola 350ml' },
    { id: 'd-guarana', name: 'Guaraná Antarctica 350ml' },
    { id: 'd-pepsi', name: 'Pepsi 350ml' },
    { id: 'd-sprite', name: 'Sprite 350ml' }
  ];

  function toggleRemoved(key: string) {
    removedState[key] = !removedState[key];
  }

  // Calculate extras specific to combo options
  const comboExtras = $derived.by(() => {
    let extra = 0;
    if (isCombo) {
      if (extraMayo) extra += 2.00;
      if (isChained) {
        if (extraMeat1) extra += 4.00;
      } else {
        if (extraMeat1) extra += 2.00;
        if (extraMeat2) extra += 2.00;
      }
    }
    return extra;
  });

  const baseUnitPrice = $derived(
    activeLine ? calc_line_price({ ...activeLine, quantity: 1 }) : (Number(menu.basePrice || menu.price) || 0)
  );
  const unitPrice = $derived(baseUnitPrice + comboExtras);
  const linePrice = $derived(unitPrice * (activeLine?.quantity || 1));

  // Original price & discount calculation
  const originalPrice = $derived(Number(menu.originalPrice || menu.oldPrice) || 0);
  const discount = $derived.by(() => {
    if (!originalPrice || originalPrice <= baseUnitPrice) return null;
    const diff = originalPrice - baseUnitPrice;
    const percent = Math.round((diff / originalPrice) * 100);
    return { diff, percent };
  });

  function formatBRL(v: number): string {
    return new Intl.NumberFormat('pt-BR', {
      style: 'currency',
      currency: 'BRL',
    }).format(v);
  }

  function handleDecrease() {
    if (!activeLine) return;
    if (activeLine.quantity <= 1) {
      handleRemove();
    } else {
      update_quantity(activeLine.id, activeLine.quantity - 1);
    }
  }

  function handleIncrease() {
    if (!activeLine) return;
    update_quantity(activeLine.id, activeLine.quantity + 1);
  }

  function handleRemove() {
    if (!activeLine) return;
    if (onRemove) {
      onRemove(activeLine.id);
    } else {
      remove_item(activeLine.id);
    }
  }
</script>

<article class="order-card">
  <!-- Header: image · name · price · discount -->
  <div class="card-header">
    <div class="thumb">
      <img src={menu.imageUrl || menu.image || menu.icon || 'https://images.unsplash.com/photo-1568901346375-23c9450c58cd?auto=format&fit=crop&q=80&w=200&h=200'} alt={menu.name || menu.title || 'Item'} />
    </div>
    <div class="min-w-0 flex-1">
      <div class="flex items-center gap-2 flex-wrap">
        <h2 class="item-name">{menu.name || menu.title}</h2>
        {#if isCombo}
          <span class="combo-badge">Combo</span>
        {/if}
      </div>

      <div class="flex items-center gap-2 mt-1">
        {#if discount}
          <span class="text-xs text-text-muted line-through font-medium">{formatBRL(originalPrice)}</span>
          <span class="text-lg font-black text-primary">{formatBRL(unitPrice)}</span>
          <span class="bg-emerald-100 text-emerald-800 dark:bg-emerald-950/60 dark:text-emerald-300 text-[10px] font-bold px-1.5 py-0.5 rounded uppercase tracking-wider">
            {discount.percent}% OFF
          </span>
        {:else}
          <span class="text-lg font-black text-primary">{formatBRL(unitPrice)}</span>
        {/if}
        {#if activeLine && activeLine.quantity > 1}
          <span class="text-xs text-text-muted font-medium">× {activeLine.quantity}</span>
        {/if}
      </div>
    </div>
    <div class="line-price">{formatBRL(linePrice)}</div>
  </div>

  <!-- ─────────────────────────────────────────────────────────────
       NON-COMBO: Ingredients to remove & extras
       ───────────────────────────────────────────────────────────── -->
  {#if !isCombo}
    {#if menu.ingredients && menu.ingredients.length > 0}
      <section class="section">
        <h3 class="section-title">Remover Ingredientes</h3>
        <div class="chip-list">
          {#each menu.ingredients as ing (ing.id)}
            <IngredientItem
              label={ing.name}
              isRemoved={!ing.selected}
              onToggle={() => {
                if (activeLine?.id) {
                  toggle_ingredient(activeLine.id, ing.id);
                }
              }}
            />
          {/each}
        </div>
      </section>
    {:else if menu.options && menu.options.length > 0}
      <section class="section">
        <h3 class="section-title">Remover Ingredientes</h3>
        <div class="chip-list">
          {#each menu.options as ing (ing.id)}
            <IngredientItem
              label={ing.name}
              isRemoved={Boolean(removedState[ing.id])}
              onToggle={() => toggleRemoved(ing.id)}
            />
          {/each}
        </div>
      </section>
    {/if}

    {#if menu.extras && menu.extras.length > 0}
      <section class="section">
        <h3 class="section-title">Adicionar extras</h3>
        <ul class="extra-list">
          {#each menu.extras as ex (ex.id)}
            <li>
              <label class="extra" class:is-selected={ex.selected}>
                <input
                  type="checkbox"
                  checked={ex.selected}
                  onchange={() => {
                    if (activeLine?.id) {
                      toggle_extra(activeLine.id, ex.id);
                    }
                  }}
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
  {/if}

  <!-- ─────────────────────────────────────────────────────────────
       COMBO: Chain toggle, burgers, fries, drinks
       ───────────────────────────────────────────────────────────── -->
  {#if isCombo}
    <div class="space-y-4 my-4">
      
      <!-- 1. BURGERS SECTION -->
      <div class="pl-4 border-l-4 border-blue-400 bg-blue-50/40 dark:bg-blue-950/20 dark:border-blue-500 p-4 rounded-r-2xl border border-border/50">
        <div class="flex items-center justify-between mb-3 gap-2">
          <h3 class="text-base font-bold text-text">
            {isChained ? '2x Hambúrgueres' : 'Hambúrgueres'}
          </h3>
          <label class="chain-toggle-badge text-blue-600 dark:text-blue-400 border-blue-200 dark:border-blue-800">
            <input type="checkbox" bind:checked={isChained} class="sr-only" />
            
            {#if isChained}
              <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/>
                <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/>
              </svg>
            {:else}
              <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/>
                <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/>
                <line x1="2" y1="2" x2="22" y2="22" stroke-width="3"/>
              </svg>
            {/if}
            
            <span>Iguais</span>
          </label>
        </div>

        {#if isChained}
          <div class="space-y-3">
            <div class="chip-list">
              {#each burgerIngredients as ing (ing.id)}
                <IngredientItem 
                  label={ing.name} 
                  isRemoved={Boolean(removedState[ing.id])} 
                  onToggle={() => toggleRemoved(ing.id)}
                />
              {/each}
            </div>
            <label class="custom-card-label">
              <div class="flex items-center gap-3">
                <input type="checkbox" bind:checked={extraMeat1} class="extra-checkbox" />
                <span class="font-semibold text-text text-sm">Carne Extra (para os dois)</span>
              </div>
              <span class="font-bold text-blue-600 dark:text-blue-400 text-sm">+ R$ 4,00</span>
            </label>
          </div>
        {:else}
          <div class="space-y-3">
            <div class="sub-card">
              <h4 class="sub-card-title">Hambúrguer 1</h4>
              <div class="chip-list mb-3">
                {#each burgerIngredients as ing (ing.id)}
                  <IngredientItem 
                    label={ing.name} 
                    isRemoved={Boolean(removedState[`${ing.id}-1`])} 
                    onToggle={() => toggleRemoved(`${ing.id}-1`)}
                  />
                {/each}
              </div>
              <label class="flex items-center justify-between cursor-pointer">
                <div class="flex items-center gap-3">
                  <input type="checkbox" bind:checked={extraMeat1} class="extra-checkbox" />
                  <span class="font-medium text-text text-xs">Carne Extra</span>
                </div>
                <span class="font-bold text-blue-600 dark:text-blue-400 text-xs">+ R$ 2,00</span>
              </label>
            </div>

            <div class="sub-card">
              <h4 class="sub-card-title">Hambúrguer 2</h4>
              <div class="chip-list mb-3">
                {#each burgerIngredients as ing (ing.id)}
                  <IngredientItem 
                    label={ing.name} 
                    isRemoved={Boolean(removedState[`${ing.id}-2`])} 
                    onToggle={() => toggleRemoved(`${ing.id}-2`)}
                  />
                {/each}
              </div>
              <label class="flex items-center justify-between cursor-pointer">
                <div class="flex items-center gap-3">
                  <input type="checkbox" bind:checked={extraMeat2} class="extra-checkbox" />
                  <span class="font-medium text-text text-xs">Carne Extra</span>
                </div>
                <span class="font-bold text-blue-600 dark:text-blue-400 text-xs">+ R$ 2,00</span>
              </label>
            </div>
          </div>
        {/if}
      </div>

      <!-- 2. FRIES / SIDE SECTION -->
      <div class="pl-4 border-l-4 border-yellow-400 bg-yellow-50/40 dark:bg-yellow-950/20 dark:border-yellow-500 p-4 rounded-r-2xl border border-border/50">
        <h3 class="text-base font-bold text-text mb-3">Batata Grande</h3>
        <div class="space-y-2">
          <label class="custom-card-label">
            <div class="flex items-center gap-3">
              <input type="checkbox" bind:checked={extraMayo} class="extra-checkbox" />
              <span class="font-semibold text-text text-sm">Maionese Gourmet</span>
            </div>
            <span class="font-bold text-yellow-600 dark:text-yellow-400 text-sm">+ R$ 2,00</span>
          </label>
        </div>
      </div>

      <!-- 3. DRINKS SECTION -->
      <div class="pl-4 border-l-4 border-purple-400 bg-purple-50/40 dark:bg-purple-950/20 dark:border-purple-500 p-4 rounded-r-2xl border border-border/50">
        <div class="flex items-center justify-between mb-3 gap-2">
          <h3 class="text-base font-bold text-text">
            {isDrinkChained ? '2x Bebidas' : 'Bebidas'}
          </h3>
          <label class="chain-toggle-badge text-purple-600 dark:text-purple-400 border-purple-200 dark:border-purple-800">
            <input type="checkbox" bind:checked={isDrinkChained} class="sr-only" />
            
            {#if isDrinkChained}
              <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/>
                <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/>
              </svg>
            {:else}
              <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/>
                <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/>
                <line x1="2" y1="2" x2="22" y2="22" stroke-width="3"/>
              </svg>
            {/if}
            
            <span>Iguais</span>
          </label>
        </div>

        {#if isDrinkChained}
          <div class="flex flex-col gap-2">
            {#each drinkChoices as choice}
              <button
                type="button"
                onclick={() => selectedDrink1 = choice.id}
                class="w-full text-left px-4 py-3 rounded-xl text-sm font-bold border transition-all flex items-center justify-between
                {selectedDrink1 === choice.id 
                  ? 'bg-purple-600 text-white border-purple-600 shadow-md' 
                  : 'bg-surface text-text border-border hover:border-purple-300 dark:hover:border-purple-600'}"
              >
                <span>{choice.name}</span>
                {#if selectedDrink1 === choice.id}
                  <span class="text-xs bg-white/20 px-2 py-0.5 rounded-full">Selecionado</span>
                {/if}
              </button>
            {/each}
          </div>
        {:else}
          <div class="space-y-4">
            <div class="sub-card">
              <h4 class="sub-card-title">Bebida 1</h4>
              <div class="flex flex-col gap-2">
                {#each drinkChoices as choice}
                  <button
                    type="button"
                    onclick={() => selectedDrink1 = choice.id}
                    class="w-full text-left px-3 py-2 rounded-lg text-xs font-bold border transition-all
                    {selectedDrink1 === choice.id ? 'bg-purple-100 text-purple-800 border-purple-300 dark:bg-purple-950/80 dark:text-purple-300 dark:border-purple-700' : 'bg-surface text-text-muted border-border hover:border-border-hover'}"
                  >
                    {choice.name}
                  </button>
                {/each}
              </div>
            </div>

            <div class="sub-card">
              <h4 class="sub-card-title">Bebida 2</h4>
              <div class="flex flex-col gap-2">
                {#each drinkChoices as choice}
                  <button
                    type="button"
                    onclick={() => selectedDrink2 = choice.id}
                    class="w-full text-left px-3 py-2 rounded-lg text-xs font-bold border transition-all
                    {selectedDrink2 === choice.id ? 'bg-purple-100 text-purple-800 border-purple-300 dark:bg-purple-950/80 dark:text-purple-300 dark:border-purple-700' : 'bg-surface text-text-muted border-border hover:border-border-hover'}"
                  >
                    {choice.name}
                  </button>
                {/each}
              </div>
            </div>
          </div>
        {/if}
      </div>

    </div>
  {/if}

  {#if activeLine?.notes}
    <p class="notes">"{activeLine.notes}"</p>
  {/if}

  <!-- Footer: quantity · remove -->
  <div class="card-footer">
    <div class="qty-controls">
      <button
        type="button"
        onclick={handleDecrease}
        class="qty-btn"
        aria-label="Diminuir quantidade"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round">
          <line x1="5" y1="12" x2="19" y2="12" />
        </svg>
      </button>
      <span class="qty-value">{activeLine?.quantity || 1}</span>
      <button
        type="button"
        onclick={handleIncrease}
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
      onclick={handleRemove}
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
    @apply bg-surface rounded-3xl p-5 shadow-sm border border-border transition-colors;
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

  /* Custom badges & cards */
  .chain-toggle-badge {
    @apply flex items-center gap-1.5 cursor-pointer text-xs font-bold bg-surface px-2.5 py-1 rounded-lg border shadow-xs hover:opacity-90 transition-all select-none;
  }
  .custom-card-label {
    @apply flex items-center justify-between p-3 rounded-xl border border-border bg-surface cursor-pointer hover:bg-surface-hover transition-colors;
  }
  .sub-card {
    @apply bg-surface p-3 rounded-xl border border-border;
  }
  .sub-card-title {
    @apply text-[10px] font-bold text-text-muted uppercase mb-2;
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
    @apply w-4 h-4 rounded text-primary focus:ring-primary cursor-pointer accent-orange-500;
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
    @apply flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold text-red-600 hover:bg-red-50 dark:hover:bg-red-950/40 transition-colors;
  }
</style>
