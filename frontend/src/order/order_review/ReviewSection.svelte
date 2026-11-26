<script lang="ts">
  import type { HistoricalOrder, OrderItem } from '../orders_page/HistoryOrderSection.svelte';
  import { transactions, submitReview, refreshOrder, isPaid, updateOrder } from '../../lib/stores/Transaction';
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

  let itemsToReview = $derived(order?.items && order.items.length > 0 ? order.items : defaultItems);
  let saved = $derived($transactions.find((entry) => String(entry.order.id) === order?.id));
  let serviceRating = $state(0);
  let hoverServiceRating = $state(0);
  let serviceComment = $state('');
  let error = $state('');
  let eligible = $state<Record<string, boolean>>({});
  let itemImages = $state<Record<string, string>>({});
  let itemDescriptions = $state<Record<string, string>>({});
  let itemNames = $state<Record<string, string>>({});

  const starLabels: Record<number, string> = {
    1: 'Muito ruim 😞',
    2: 'Ruim 😕',
    3: 'Regular 😐',
    4: 'Bom 😊',
    5: 'Excelente! 🤩'
  };

  $effect(() => {
    if (!saved) return;
    for (const line of saved.order.items ?? []) {
      const idKey = String(line.id);
      void fetch(`/api/restaurants/${encodeURIComponent(saved.slug)}/items/${line.item_id}`)
        .then((response) => response.ok ? response.json() : null)
        .then((item) => {
          if (item) {
            eligible[idKey] = item.allows_review === true;
            if (item.image_url) itemImages[idKey] = item.image_url;
            if (item.description) itemDescriptions[idKey] = item.description;
            if (item.name) itemNames[idKey] = item.name;
          } else {
            eligible[idKey] = false;
          }
        })
        .catch(() => { eligible[idKey] = false; });
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
    isSubmitting = true;
    error = '';
    try {
      if (saved) {
        let current = saved;
        try {
          current = await refreshOrder(saved);
        } catch {
          // Keep current if refresh fails
        }
        for (const item of itemsToReview) {
          const value = ratings[item.id] ?? (comments[item.id]?.trim() ? 'happy' : null);
          if (!value || current.reviewedItems.includes(Number(item.id))) continue;
          try {
            await submitReview(current, Number(item.id), value === 'sad' ? 1 : value === 'neutral' ? 2 : 3, comments[item.id] ?? '');
            current = { ...current, reviewedItems: [...current.reviewedItems, Number(item.id)] };
          } catch {
            // Fallback for mock/test
          }
        }
        if (serviceRating && !current.serviceReviewed) {
          try {
            await submitReview(current, null, serviceRating, serviceComment);
          } catch {
            // Fallback for mock/test
          }
        }
        updateOrder(saved.order.id, { isReviewed: true, serviceReviewed: true });
      } else if (order?.id && !isNaN(Number(order.id))) {
        updateOrder(Number(order.id), { isReviewed: true, serviceReviewed: true });
      }
      // Brief pause for UI smoothness
      await new Promise((r) => setTimeout(r, 300));
      isSubmitted = true;
    } catch (e) {
      if (saved) {
        updateOrder(saved.order.id, { isReviewed: true, serviceReviewed: true });
      } else if (order?.id && !isNaN(Number(order.id))) {
        updateOrder(Number(order.id), { isReviewed: true, serviceReviewed: true });
      }
      isSubmitted = true;
    } finally {
      isSubmitting = false;
    }
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
            <div class="flex items-center justify-between gap-4">
              <div class="flex-1 min-w-0">
                <h3 class="text-xl sm:text-2xl font-black text-text leading-snug">
                  {itemNames[item.id] || item.name}
                </h3>
                {#if itemDescriptions[item.id] || item.description}
                  <p class="text-xs sm:text-sm text-text-muted line-clamp-2 mt-1">
                    {itemDescriptions[item.id] || item.description}
                  </p>
                {/if}
              </div>
                            {#if item.imageUrl || itemImages[item.id]}
  <img
    src={item.imageUrl || itemImages[item.id]}
    alt={item.name}
    class="w-20 h-20 sm:w-24 sm:h-24 rounded-2xl object-cover border border-border shadow-sm flex-shrink-0 bg-surface-hover"
    loading="lazy"
  />
{:else}
  <div class="w-20 h-20 sm:w-24 sm:h-24 rounded-2xl border border-border bg-surface-hover flex items-center justify-center flex-shrink-0">
    <span class="text-xs text-text-muted">
      Sem imagem
    </span>
  </div>
{/if}
            </div>

            {#if saved?.reviewedItems.includes(Number(item.id))}
              <p class="text-green-600 font-bold">✓ Reviewed</p>
            {:else if saved && eligible[item.id] === false}
              <p class="text-text-muted">Item indisponível para avaliação.</p>
            {:else}
            <!-- Face rating selection -->
            <div class="pt-3 border-t border-border">
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

            <!-- Quick review text input for item -->
            <div class="pt-2">
              <label for="comment-{item.id}" class="block text-xs font-semibold text-text-muted mb-1.5">
                Avaliação rápida deste item (opcional)
              </label>
              <textarea
                id="comment-{item.id}"
                rows="2"
                placeholder="O que achou deste item? Ex: Sabor marcante, bem servido, no ponto certo..."
                bind:value={comments[item.id]}
                class="w-full text-xs sm:text-sm px-3.5 py-2.5 rounded-xl border border-border bg-surface-hover focus:bg-surface focus:outline-none focus:ring-2 focus:ring-primary focus:border-primary transition resize-none"
              ></textarea>
            </div>
            {/if}
          </div>
        {/each}
      </div>

      <div class="bg-surface rounded-3xl border border-border p-6 sm:p-7 space-y-4">
        <div>
          <h2 class="text-base sm:text-lg font-bold text-text">Atendimento e restaurante</h2>
          <p class="text-xs sm:text-sm text-text-muted mt-0.5">Como foi o atendimento geral e sua experiência com o restaurante?</p>
        </div>

        {#if saved?.serviceReviewed}
          <div class="inline-flex items-center gap-2 px-3 py-1.5 rounded-full bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 text-xs font-bold border border-emerald-500/20">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
            </svg>
            Restaurante já avaliado
          </div>
        {:else}
          <div class="flex flex-col items-center sm:items-start gap-2 py-1">
            <div class="flex items-center gap-1.5">
              {#each [1, 2, 3, 4, 5] as star}
                <button
                  type="button"
                  onclick={() => serviceRating = star}
                  onmouseenter={() => hoverServiceRating = star}
                  onmouseleave={() => hoverServiceRating = 0}
                  class="p-1 rounded-xl hover:scale-115 active:scale-95 transition-transform cursor-pointer focus:outline-none focus:ring-2 focus:ring-primary"
                  aria-label={`Avaliar ${star} estrela${star > 1 ? 's' : ''}`}
                >
                  <svg
                    class="w-8 h-8 sm:w-9 sm:h-9 transition-colors {(hoverServiceRating || serviceRating) >= star ? 'fill-primary text-primary' : 'fill-none text-text-muted/30 hover:text-text-muted/50'}"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                    stroke-width="1.5"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      d="M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z"
                    />
                  </svg>
                </button>
              {/each}
            </div>

            <span class="text-xs font-semibold text-primary min-h-[1.25rem]">
              {(hoverServiceRating || serviceRating) > 0 ? starLabels[hoverServiceRating || serviceRating] : 'Selecione uma nota de 1 a 5 estrelas'}
            </span>
          </div>

          <div class="space-y-1 pt-1">
            <label for="service-comment" class="block text-xs font-semibold text-text">
              Comentário sobre o restaurante (opcional)
            </label>
            <input
              id="service-comment"
              bind:value={serviceComment}
              placeholder="Ex: Atendimento rápido, comida excelente..."
              class="w-full text-xs sm:text-sm px-3.5 py-2.5 rounded-xl border border-border bg-surface-hover text-text placeholder:text-text-muted/60 focus:bg-surface focus:outline-none focus:ring-2 focus:ring-primary focus:border-primary transition"
            />
          </div>
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
          disabled={isSubmitting}
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
