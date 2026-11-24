<script lang="ts">
  import type { HistoricalOrder, OrderItem } from '../orders_page/HistoryOrderSection.svelte';
  import { transactions, submitReview, refreshOrder, isPaid } from '../../lib/stores/Transaction';
  import { onMount } from 'svelte';

  let {
    order = null as HistoricalOrder | null,
    onBack = () => {}
  }: {
    order?: HistoricalOrder | null;
    onBack?: () => void;
  } = $props();

  // Default fallback items if order wasn't passed directly:
  const defaultItems: OrderItem[] = [
    {
      id: 'default-combo',
      name: 'Combo Smash Bacon Duplo',
      description: 'Pão brioche tostado na manteiga, 2x smash burger 90g, cheddar cremoso, tiras crocantes de bacon e batata frita.',
      quantity: 1,
      price: 39.9,
      imageUrl: 'https://images.unsplash.com/photo-1568901346375-23c9450c58cd?w=400&auto=format&fit=crop&q=80'
    },
    {
      id: 'default-icecream',
      name: 'Sorvete Artesanal de Baunilha com Calda',
      description: 'Taça de sorvete premium sabor baunilha com calda quente de chocolate e castanhas.',
      quantity: 1,
      price: 15.0,
      imageUrl: 'https://images.unsplash.com/photo-1570197788417-0e82375c9371?w=400&auto=format&fit=crop&q=80'
    }
  ];

  let itemsToReview = $derived(order?.items ?? []);
  let saved = $derived($transactions.find((entry) => String(entry.order.id) === order?.id));
  let serviceRating = $state(0);
  let serviceComment = $state('');
  let error = $state('');
  let eligible = $state<Record<string, boolean>>({});

  onMount(() => {
    if (!saved) return;
    for (const line of saved.order.items ?? []) {
      void fetch(`/api/restaurants/${encodeURIComponent(saved.slug)}/items/${line.item_id}`)
        .then((response) => response.ok ? response.json() : null)
        .then((item) => { eligible[String(line.id)] = item?.allows_review === true; })
        .catch(() => { eligible[String(line.id)] = false; });
    }
  });

  type Rating = 'sad' | 'neutral' | 'happy' | null;

  // Track ratings and comments per item
  let ratings = $state<Record<string, Rating>>({});
  let comments = $state<Record<string, string>>({});
  let isSubmitting = $state(false);
  let isSubmitted = $state(false);

  function setRating(itemId: string, rating: Rating) {
    ratings[itemId] = rating;
  }

  async function handleSubmit(e: SubmitEvent) {
    e.preventDefault();
    if (!saved) return;
    isSubmitting = true;
    error = '';
    try {
      let current = await refreshOrder(saved);
      if (current.order.status !== 'delivered' || !isPaid(current.order)) {
        throw new Error('A avaliação estará disponível após o pagamento e a entrega do pedido.');
      }
      let submitted = false;
      for (const item of itemsToReview) {
        const value = ratings[item.id];
        if (!value || !eligible[item.id] || current.reviewedItems.includes(Number(item.id))) continue;
        await submitReview(current, Number(item.id), value === 'sad' ? 1 : value === 'neutral' ? 2 : 3, comments[item.id] ?? '');
        current = { ...current, reviewedItems: [...current.reviewedItems, Number(item.id)] };
        submitted = true;
      }
      if (serviceRating && !current.serviceReviewed) {
        await submitReview(current, null, serviceRating, serviceComment);
        submitted = true;
      }
      if (!submitted) throw new Error('Selecione uma avaliação antes de enviar.');
      isSubmitted = true;
    } catch (e) {
      error = e instanceof Error ? e.message : 'Não foi possível enviar a avaliação.';
    } finally { isSubmitting = false; }
  }

  function handleReset() {
    isSubmitted = false;
    ratings = {};
    comments = {};
  }
</script>

<div class="space-y-8">
  {#if isSubmitted}
    <!-- Fake Success State -->
    <div class="bg-surface rounded-3xl border border-border p-8 sm:p-12 text-center shadow-lg shadow-emerald-500/5 animate-fade-in">
      <div class="w-20 h-20 bg-surface-hover text-primary rounded-full flex items-center justify-center mx-auto mb-6 shadow-inner">
        <svg class="w-10 h-10" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
        </svg>
      </div>

      <span class="inline-block px-3.5 py-1 rounded-full bg-surface-hover text-primary text-xs font-bold uppercase tracking-wider mb-2 border border-border">
        Sucesso!
      </span>

      <h2 class="text-2xl sm:text-3xl font-extrabold text-text mb-2">Avaliação Enviada!</h2>
      <p class="text-text-muted text-sm sm:text-base max-w-md mx-auto mb-8">
        Muito obrigado pelo seu feedback. Sua opinião nos ajuda a preparar pratos e combos ainda melhores para você!
      </p>

      <div class="flex flex-col sm:flex-row items-center justify-center gap-3">
        <button
          type="button"
          onclick={onBack}
          class="w-full sm:w-auto px-6 py-3 bg-primary hover:bg-primary-hover active:bg-primary-hover text-white font-bold rounded-2xl shadow-sm transition"
        >
          Voltar para Meus Pedidos
        </button>
      </div>
    </div>
  {:else}
    <!-- Review Form -->
    <form onsubmit={handleSubmit} class="space-y-6">
      <div class="border-b border-border pb-4">
        <h1 class="text-2xl sm:text-3xl font-extrabold text-text tracking-tight">Avaliar Pedido</h1>
        <p class="text-sm text-text-muted mt-1">
          {order?.restaurantName ? `${order.restaurantName} • ` : ''}{order?.orderNumber ?? 'Pedido Recente'} — Como estava cada item do seu pedido?
        </p>
      </div>

      <!-- Item review cards -->
      <div class="space-y-6">
        {#each itemsToReview as item (item.id)}
          <div class="bg-surface rounded-3xl border border-border p-5 sm:p-7 shadow-xs space-y-5">
            <!-- Item header -->
            <div class="flex items-start sm:items-center gap-4">
              {#if item.imageUrl}<img
                src={item.imageUrl}
                alt={item.name}
                class="w-20 h-20 sm:w-24 sm:h-24 rounded-2xl object-cover border border-border shadow-inner flex-shrink-0 bg-surface-hover"
                loading="lazy"
              />{/if}
              <div class="flex-1 min-w-0">
                <span class="inline-block text-[11px] font-semibold uppercase tracking-wider text-primary bg-surface-hover px-2 py-0.5 rounded-md border border-border mb-1">
                  Item do Pedido
                </span>
                <h3 class="text-base sm:text-lg font-bold text-text leading-snug">{item.name}</h3>
                {#if item.description}
                  <p class="text-xs text-text-muted line-clamp-2 mt-1">{item.description}</p>
                {/if}
              </div>
            </div>

            {#if saved?.reviewedItems.includes(Number(item.id))}
              <p class="text-green-600 font-bold">✓ Reviewed</p>
            {:else if !eligible[item.id]}
              <p class="text-text-muted">Item indisponível para avaliação.</p>
            {:else}
            <!-- Face rating selection -->
            <div class="pt-3 border-t border-border">
              <p class="text-xs font-bold text-text-muted uppercase tracking-wider mb-3">
                O que você achou deste item?
              </p>

              <div class="mx-auto grid w-full max-w-md grid-cols-3 gap-4 sm:gap-6">
                <!-- Sad face -->
                <button
                  type="button"
                  onclick={() => setRating(item.id, 'sad')}
                  class="flex flex-col items-center justify-center p-3 sm:p-4 rounded-2xl border-2 transition-all cursor-pointer {ratings[item.id] === 'sad' ? 'border-rose-400 bg-rose-50 text-rose-700 shadow-xs ring-2 ring-rose-200 scale-102' : 'border-border hover:border-stone-300 hover:bg-surface-hover text-text-muted'}"
                  aria-label="Avaliar como ruim"
                >
                  <svg class="w-8 h-8 sm:w-10 sm:h-10 transition-transform {ratings[item.id] === 'sad' ? 'scale-110 text-rose-500' : 'text-text-muted'}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <circle cx="12" cy="12" r="10" />
                    <path d="M16 16s-1.5-2-4-2-4 2-4 2" />
                    <line x1="9" y1="9" x2="9.01" y2="9" />
                    <line x1="15" y1="9" x2="15.01" y2="9" />
                  </svg>
                  <span class="text-xs font-semibold mt-2 {ratings[item.id] === 'sad' ? 'text-rose-700' : 'text-text-muted'}">Ruim</span>
                </button>

                <!-- Neutral face -->
                <button
                  type="button"
                  onclick={() => setRating(item.id, 'neutral')}
                  class="flex flex-col items-center justify-center p-3 sm:p-4 rounded-2xl border-2 transition-all cursor-pointer {ratings[item.id] === 'neutral' ? 'border-primary bg-surface-hover text-primary shadow-xs ring-2 ring-amber-200 scale-102' : 'border-border hover:border-stone-300 hover:bg-surface-hover text-text-muted'}"
                  aria-label="Avaliar como regular"
                >
                  <svg class="w-8 h-8 sm:w-10 sm:h-10 transition-transform {ratings[item.id] === 'neutral' ? 'scale-110 text-amber-500' : 'text-text-muted'}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <circle cx="12" cy="12" r="10" />
                    <line x1="8" y1="15" x2="16" y2="15" />
                    <line x1="9" y1="9" x2="9.01" y2="9" />
                    <line x1="15" y1="9" x2="15.01" y2="9" />
                  </svg>
                  <span class="text-xs font-semibold mt-2 {ratings[item.id] === 'neutral' ? 'text-primary' : 'text-text-muted'}">Regular</span>
                </button>

                <!-- Happy face -->
                <button
                  type="button"
                  onclick={() => setRating(item.id, 'happy')}
                  class="flex flex-col items-center justify-center p-3 sm:p-4 rounded-2xl border-2 transition-all cursor-pointer {ratings[item.id] === 'happy' ? 'border-emerald-400 bg-surface-hover text-primary shadow-xs ring-2 ring-emerald-200 scale-102' : 'border-border hover:border-stone-300 hover:bg-surface-hover text-text-muted'}"
                  aria-label="Avaliar como muito bom"
                >
                  <svg class="w-8 h-8 sm:w-10 sm:h-10 transition-transform {ratings[item.id] === 'happy' ? 'scale-110 text-emerald-500' : 'text-text-muted'}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <circle cx="12" cy="12" r="10" />
                    <path d="M8 14s1.5 2 4 2 4-2 4-2" />
                    <line x1="9" y1="9" x2="9.01" y2="9" />
                    <line x1="15" y1="9" x2="15.01" y2="9" />
                  </svg>
                  <span class="text-xs font-semibold mt-2 {ratings[item.id] === 'happy' ? 'text-primary' : 'text-text-muted'}">Muito Bom</span>
                </button>
              </div>
            </div>

            <!-- Optional comment input for item -->
            {#if saved?.userId !== null}<div class="pt-2">
              <label for="comment-{item.id}" class="block text-xs text-text-muted mb-1">
                Deixe um comentário adicional (opcional)
              </label>
              <input
                id="comment-{item.id}"
                type="text"
                placeholder="Ex: Ponto da carne perfeito, batata quentinha..."
                bind:value={comments[item.id]}
                class="w-full text-xs sm:text-sm px-3.5 py-2.5 rounded-xl border border-border bg-surface-hover focus:bg-surface focus:outline-none focus:ring-2 focus:ring-primary focus:border-primary transition"
              />
            </div>{/if}
            {/if}
          </div>
        {/each}
      </div>

      <div class="bg-surface rounded-3xl border border-border p-6 space-y-3">
        <h2 class="font-bold">Atendimento e restaurante</h2>
        {#if saved?.serviceReviewed}
          <p class="text-green-600 font-bold">✓ Reviewed</p>
        {:else}
          <label for="service-rating">Avaliação (1–5)</label>
          <select id="service-rating" bind:value={serviceRating} class="block border border-border rounded-xl p-2">
            <option value={0}>Selecione</option>
            {#each [1, 2, 3, 4, 5] as value}<option {value}>{value}</option>{/each}
          </select>
          {#if saved?.userId !== null}
            <label for="service-comment">Comentário (opcional)</label>
            <input id="service-comment" bind:value={serviceComment} class="block w-full border border-border rounded-xl p-2" />
          {/if}
        {/if}
      </div>
      {#if error}<p role="alert" class="text-red-600">{error}</p>{/if}
      <!-- Action buttons -->
      <div class="flex flex-col sm:flex-row items-center justify-between gap-4 pt-4">
        <button
          type="button"
          onclick={onBack}
          class="w-full sm:w-auto px-5 py-3 text-text-muted hover:text-text hover:bg-surface-hover rounded-xl font-medium text-sm transition"
        >
          ← Voltar para Meus Pedidos
        </button>

        <button
          type="submit"
          disabled={isSubmitting || !saved || saved.order.status !== 'delivered'}
          class="w-full sm:w-auto inline-flex items-center justify-center gap-2 px-8 py-3.5 bg-primary hover:bg-primary-hover active:bg-primary-hover disabled:opacity-60 text-white font-bold text-sm rounded-2xl shadow-md shadow-primary/20 transition cursor-pointer"
        >
          {#if isSubmitting}
            <svg class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            Enviando...
          {:else}
            <span>Enviar Avaliação</span>
            <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" d="M14 5l7 7m0 0l-7 7m7-7H3" />
            </svg>
          {/if}
        </button>
      </div>
    </form>
  {/if}
</div>

<style>
  @keyframes fadeIn {
    from {
      opacity: 0;
      transform: translateY(10px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .animate-fade-in {
    animation: fadeIn 0.3s ease-out forwards;
  }
</style>
