<script lang="ts">
  import CheckoutSection from './CheckoutSection.svelte';

  let {
    onCancel = () => {},
    slug = '',
    total = 0,
    onOrderConfirmed = undefined as ((restaurantId: string | number, orderId: string | number) => void) | undefined
  }: {
    onCancel?: () => void;
    slug?: string;
    total?: number;
    onOrderConfirmed?: (restaurantId: string | number, orderId: string | number) => void;
  } = $props();
</script>

<main class="min-h-screen bg-background text-text py-8 px-4 sm:px-6 lg:px-8 transition-colors duration-200">
  <div class="max-w-3xl mx-auto space-y-6">
    <!-- Breadcrumb Header -->
    <div class="flex items-center gap-2 text-sm text-text-muted">
      <button
        type="button"
        onclick={onCancel}
        class="hover:text-primary transition flex items-center gap-1 font-medium cursor-pointer"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" d="M15 19l-7-7 7-7" />
        </svg>
        Voltar para o carrinho
      </button>
      <span>/</span>
      <span class="text-text font-medium">Finalizar Pedido</span>
    </div>

    <!-- Checkout Section Container -->
    <CheckoutSection {onCancel} {slug} {total} {onOrderConfirmed} />
  </div>
</main>
