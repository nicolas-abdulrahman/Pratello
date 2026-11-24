<script lang="ts">
  import HistoryOrderSection from './HistoryOrderSection.svelte';
  import type { HistoricalOrder } from './HistoryOrderSection.svelte';

  let {
    onOpenReview = (order: HistoricalOrder) => {}
  }: {
    onOpenReview?: (order: HistoricalOrder) => void;
  } = $props();

  function handleSelectOrder(order: HistoricalOrder) {
    onOpenReview(order);
  }
</script>

<main class="min-h-screen bg-background py-8 px-4 sm:px-6 lg:px-8">
  <div class="max-w-4xl mx-auto">
    <!-- Breadcrumb / Header navigation -->
    <div class="mb-6 flex items-center gap-2 text-sm text-text-muted">
      <a href="/" class="hover:text-primary transition" aria-label="Voltar para início">← Início</a>
      <span>/</span>
      <span class="text-text font-medium">Meus Pedidos</span>
    </div>

    <!-- History Order Section -->
    <HistoryOrderSection onSelectOrder={handleSelectOrder} />
  </div>
</main>
