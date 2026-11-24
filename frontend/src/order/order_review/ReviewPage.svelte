<script lang="ts">
  import ReviewSection from './ReviewSection.svelte';
  import type { HistoricalOrder } from '../orders_page/HistoryOrderSection.svelte';

  let {
    order = null as HistoricalOrder | null,
    onBack = () => {}
  }: {
    order?: HistoricalOrder | null;
    onBack?: () => void;
  } = $props();
</script>

<main class="min-h-screen bg-background py-8 px-4 sm:px-6 lg:px-8">
  <div class="max-w-3xl mx-auto">
    <!-- Breadcrumb navigation -->
    <div class="mb-6 flex items-center gap-2 text-sm text-text-muted">
      <button
        type="button"
        onclick={onBack}
        class="hover:text-primary transition flex items-center gap-1.5 font-medium cursor-pointer"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" d="M15 19l-7-7 7-7" />
        </svg>
        Meus Pedidos
      </button>
      <span>/</span>
      <span class="text-text font-medium">Avaliação</span>
    </div>

    <!-- Review Section Component -->
    <ReviewSection {order} {onBack} />
  </div>
</main>
