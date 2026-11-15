<script lang="ts">
  import { cart, cart_total, calc_line_price } from '../lib/stores/Cart';

  let {
    onCancel = () => {},
    total: subtotalProp,
    deliveryFee = 8.5,
  }: {
    onCancel?: () => void;
    /** Cart subtotal, before delivery. Falls back to $cart_total. */
    total?: number;
    deliveryFee?: number;
  } = $props();

  let paymentMethod = $state<'card' | 'pix'>('card');
  let cardNumber = $state('');
  let cardName = $state('');
  let expiry = $state('');
  let cvv = $state('');
  let isProcessing = $state(false);

  // Subtotal: prop wins, otherwise live from the store
  const subtotal = $derived(subtotalProp ?? $cart_total);
  const grandTotal = $derived(subtotal + deliveryFee);

  function formatBRL(v: number): string {
    return new Intl.NumberFormat('pt-BR', {
      style: 'currency',
      currency: 'BRL',
    }).format(v);
  }

  function formatCardNumber(e: Event) {
    let value = (e.target as HTMLInputElement).value.replace(/\D/g, '');
    value = value.replace(/(\d{4})(?=\d)/g, '$1 ');
    cardNumber = value;
  }

  function formatExpiry(e: Event) {
    let value = (e.target as HTMLInputElement).value.replace(/\D/g, '');
    if (value.length > 2) value = value.replace(/^(\d{2})(\d)/, '$1/$2');
    expiry = value;
  }

  async function copyPix() {
    try {
      await navigator.clipboard.writeText(
        '00020126580014br.gov.bcb.pix.placeholder',
      );
    } catch {
      // ignore — user can select manually
    }
  }

  function handleSubmit(e: Event) {
    e.preventDefault();
    isProcessing = true;
    setTimeout(() => {
      isProcessing = false;
      alert(
        `Pagamento via ${
          paymentMethod === 'card' ? 'Cartão de Crédito' : 'PIX'
        } aprovado!`,
      );
    }, 2000);
  }
</script>

<main class="checkout-page">
  <div class="checkout-container">
    <!-- Header -->
    <div class="checkout-header">
      <button
        type="button"
        onclick={onCancel}
        class="back-btn"
        aria-label="Voltar"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="m15 18-6-6 6-6" />
        </svg>
      </button>
      <div>
        <h1 class="checkout-title">Finalizar Pagamento</h1>
        <p class="checkout-subtitle">Escolha como deseja pagar o seu pedido.</p>
      </div>
    </div>

    {#if $cart.length === 0}
      <!-- Empty cart guard -->
      <div class="empty-card">
        <h2 class="empty-title">Seu carrinho está vazio</h2>
        <p class="empty-msg">Adicione itens antes de finalizar o pedido.</p>
        <button type="button" onclick={onCancel} class="btn-primary mt-4">
          Voltar ao cardápio
        </button>
      </div>
    {:else}
      <div class="checkout-grid">
        <!-- LEFT: payment methods + form -->
        <div class="checkout-main">
          <!-- Method selector -->
          <div class="method-tabs" role="tablist" aria-label="Método de pagamento">
            <button
              type="button"
              role="tab"
              aria-selected={paymentMethod === 'card'}
              class="method-tab"
              class:active={paymentMethod === 'card'}
              onclick={() => (paymentMethod = 'card')}
            >
              <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <rect width="20" height="14" x="2" y="5" rx="2" />
                <line x1="2" x2="22" y1="10" y2="10" />
              </svg>
              Cartão de Crédito
            </button>

            <button
              type="button"
              role="tab"
              aria-selected={paymentMethod === 'pix'}
              class="method-tab"
              class:active={paymentMethod === 'pix'}
              onclick={() => (paymentMethod = 'pix')}
            >
              <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M12 2v20" />
                <path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6" />
              </svg>
              PIX
            </button>
          </div>

          <!-- Form card -->
          <div class="form-card">
            {#if paymentMethod === 'card'}
              <form onsubmit={handleSubmit} class="form-body">
                <h2 class="form-title">Dados do Cartão</h2>

                <div>
                  <label class="form-label" for="card-number">Número do Cartão</label>
                  <div class="relative">
                    <input
                      id="card-number"
                      type="text"
                      bind:value={cardNumber}
                      oninput={formatCardNumber}
                      placeholder="0000 0000 0000 0000"
                      maxlength="19"
                      class="form-input pl-10"
                      required
                    />
                    <div class="form-input-icon">
                      <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <rect width="20" height="14" x="2" y="5" rx="2" />
                        <line x1="2" x2="22" y1="10" y2="10" />
                      </svg>
                    </div>
                  </div>
                </div>

                <div>
                  <label class="form-label" for="card-name">Nome impresso no Cartão</label>
                  <input
                    id="card-name"
                    type="text"
                    bind:value={cardName}
                    placeholder="Como está no cartão"
                    class="form-input"
                    required
                  />
                </div>

                <div class="grid grid-cols-2 gap-4">
                  <div>
                    <label class="form-label" for="expiry">Validade</label>
                    <input
                      id="expiry"
                      type="text"
                      bind:value={expiry}
                      oninput={formatExpiry}
                      placeholder="MM/AA"
                      maxlength="5"
                      class="form-input"
                      required
                    />
                  </div>
                  <div>
                    <label class="form-label" for="cvv">CVV</label>
                    <input
                      id="cvv"
                      type="text"
                      bind:value={cvv}
                      placeholder="123"
                      maxlength="4"
                      class="form-input"
                      required
                    />
                  </div>
                </div>

                <div class="pt-2">
                  <button type="submit" disabled={isProcessing} class="btn-primary w-full">
                    {#if isProcessing}
                      <svg class="animate-spin h-5 w-5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                      </svg>
                      Processando…
                    {:else}
                      <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <rect x="3" y="11" width="18" height="11" rx="2" ry="2" />
                        <path d="M7 11V7a5 5 0 0 1 10 0v4" />
                      </svg>
                      Pagar {formatBRL(grandTotal)}
                    {/if}
                  </button>
                </div>
              </form>
            {:else}
              <div class="pix-body">
                <h2 class="form-title text-center">Pague com PIX</h2>
                <p class="pix-hint">
                  Escaneie o código QR abaixo com o aplicativo do seu banco para finalizar o pagamento.
                </p>

                <div class="qr-frame">
                  <div class="qr-inner">
                    <svg class="w-full h-full text-text" viewBox="0 0 100 100" fill="currentColor">
                      <rect x="10" y="10" width="20" height="20" />
                      <rect x="15" y="15" width="10" height="10" fill="white" />
                      <rect x="70" y="10" width="20" height="20" />
                      <rect x="75" y="15" width="10" height="10" fill="white" />
                      <rect x="10" y="70" width="20" height="20" />
                      <rect x="15" y="75" width="10" height="10" fill="white" />
                      <rect x="40" y="10" width="10" height="10" />
                      <rect x="50" y="20" width="10" height="10" />
                      <rect x="40" y="40" width="20" height="20" />
                      <rect x="10" y="40" width="10" height="20" />
                      <rect x="70" y="40" width="20" height="10" />
                      <rect x="40" y="70" width="10" height="10" />
                      <rect x="60" y="60" width="20" height="20" />
                      <rect x="80" y="80" width="10" height="10" />
                    </svg>
                    <div class="qr-center">
                      <div class="qr-logo">P</div>
                    </div>
                  </div>
                </div>

                <div class="pix-copy">
                  <label class="form-label">Ou copie e cole o código:</label>
                  <div class="flex gap-2">
                    <input
                      type="text"
                      readonly
                      value="00020126580014br.gov.bcb.pix…"
                      class="form-input text-xs"
                    />
                    <button type="button" onclick={copyPix} class="btn-secondary shrink-0">
                      Copiar
                    </button>
                  </div>
                </div>

                <button type="button" class="btn-primary w-full mt-4" onclick={() => alert('Pagamento PIX confirmado!')}>
                  Já realizei o pagamento
                </button>
              </div>
            {/if}

            <div class="security-badge">
              <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <rect x="3" y="11" width="18" height="11" rx="2" ry="2" />
                <path d="M7 11V7a5 5 0 0 1 10 0v4" />
              </svg>
              Pagamento 100% seguro e criptografado
            </div>
          </div>
        </div>

        <!-- RIGHT: order summary -->
        <aside class="checkout-aside">
          <div class="summary-card">
            <h3 class="summary-title">Resumo do Pedido</h3>

            <div class="summary-items">
              {#each $cart as line (line.id)}
                <div class="summary-row">
                  <div class="summary-left">
                    <span class="summary-qty">{line.quantity}×</span>
                    <span class="truncate">{line.item.name}</span>
                  </div>
                  <span class="summary-price">{formatBRL(calc_line_price(line))}</span>
                </div>
              {/each}
            </div>

            <div class="summary-totals">
              <div class="summary-total-row">
                <span>Subtotal</span>
                <span>{formatBRL(subtotal)}</span>
              </div>
              <div class="summary-total-row">
                <span>Taxa de entrega</span>
                <span>{formatBRL(deliveryFee)}</span>
              </div>
            </div>

            <div class="summary-grand-total">
              <span class="summary-grand-label">Total</span>
              <span class="summary-grand-value">{formatBRL(grandTotal)}</span>
            </div>
          </div>
        </aside>
      </div>
    {/if}
  </div>
</main>

<style>
  @reference "../app.css";

  /* ── Page shell ─────────────────────────────────────────── */
  .checkout-page {
    @apply min-h-screen bg-background py-12 px-4 md:px-8 font-sans;
  }
  .checkout-container {
    @apply max-w-6xl mx-auto;
  }

  .checkout-header {
    @apply flex items-start gap-4 mb-10;
  }
  .back-btn {
    @apply p-2.5 bg-surface border border-border rounded-full shadow-sm text-text-muted hover:text-primary transition-colors shrink-0 mt-1;
  }
  .checkout-title {
    @apply text-3xl font-black text-text mb-2;
  }
  .checkout-subtitle {
    @apply text-text-muted;
  }

  /* ── Grid ──────────────────────────────────────────────── */
  .checkout-grid {
    @apply grid grid-cols-1 lg:grid-cols-3 gap-8;
  }
  .checkout-main {
    @apply lg:col-span-2 space-y-6;
  }
  .checkout-aside {
    @apply lg:col-span-1;
  }

  /* ── Method tabs ───────────────────────────────────────── */
  .method-tabs {
    @apply bg-surface rounded-2xl p-2 shadow-sm border border-border flex gap-2;
  }
  .method-tab {
    @apply flex-1 py-3 px-4 rounded-xl font-bold text-sm flex items-center justify-center gap-2 transition-all text-text-muted;
  }
  .method-tab:hover:not(.active) {
    @apply bg-surface-hover;
  }
  .method-tab.active {
    @apply bg-primary-soft text-primary border border-primary/25 shadow-sm;
  }

  /* ── Form card ─────────────────────────────────────────── */
  .form-card {
    @apply bg-surface rounded-2xl p-6 md:p-8 shadow-sm border border-border;
  }
  .form-body {
    @apply space-y-5;
  }
  .form-title {
    @apply text-xl font-bold text-text mb-4;
  }
  .form-label {
    @apply block text-xs font-semibold text-text mb-1.5;
  }
  .form-input {
    @apply w-full text-sm text-text px-4 py-3.5 bg-surface-hover border border-border rounded-xl focus:outline-none focus:border-primary transition;
  }
  .form-input-icon {
    @apply absolute left-3 top-1/2 -translate-y-1/2 text-text-muted pointer-events-none;
  }

  /* ── Buttons ───────────────────────────────────────────── */
  .btn-primary {
    @apply py-4 bg-primary hover:bg-primary-hover text-white font-bold text-base rounded-xl shadow-lg shadow-primary/30 transition-all flex items-center justify-center gap-2 disabled:opacity-70 disabled:cursor-not-allowed;
  }
  .btn-secondary {
    @apply px-4 bg-surface-hover hover:bg-border text-text font-bold text-xs rounded-xl transition flex items-center justify-center;
  }

  /* ── PIX block ─────────────────────────────────────────── */
  .pix-body {
    @apply flex flex-col items-center text-center py-4;
  }
  .pix-hint {
    @apply text-sm text-text-muted mb-6 max-w-sm;
  }
  .qr-frame {
    @apply p-4 bg-surface border-2 border-dashed border-primary/40 rounded-3xl mb-6 shadow-sm;
  }
  .qr-inner {
    @apply w-48 h-48 bg-surface-hover rounded-xl flex items-center justify-center relative overflow-hidden;
  }
  .qr-center {
    @apply absolute inset-0 flex items-center justify-center pointer-events-none;
  }
  .qr-logo {
    @apply w-10 h-10 bg-surface rounded-lg flex items-center justify-center shadow-md text-primary font-black text-lg;
  }
  .pix-copy {
    @apply w-full max-w-xs mb-6;
  }

  /* ── Security badge ────────────────────────────────────── */
  .security-badge {
    @apply mt-8 pt-6 border-t border-border flex items-center justify-center gap-2 text-xs text-text-muted;
  }

  /* ── Summary ───────────────────────────────────────────── */
  .summary-card {
    @apply bg-surface rounded-2xl p-6 shadow-sm border border-border sticky top-8;
  }
  .summary-title {
    @apply text-lg font-bold text-text mb-4;
  }
  .summary-items {
    @apply space-y-3 mb-6;
  }
  .summary-row {
    @apply flex justify-between items-start text-sm gap-2;
  }
  .summary-left {
    @apply flex items-start gap-2 text-text-muted min-w-0;
  }
  .summary-qty {
    @apply font-bold text-primary shrink-0;
  }
  .summary-price {
    @apply font-semibold text-text whitespace-nowrap;
  }
  .summary-totals {
    @apply space-y-2 pt-4 border-t border-border text-sm;
  }
  .summary-total-row {
    @apply flex justify-between text-text-muted;
  }
  .summary-grand-total {
    @apply flex justify-between items-center pt-4 mt-4 border-t border-border;
  }
  .summary-grand-label {
    @apply text-base font-bold text-text;
  }
  .summary-grand-value {
    @apply text-2xl font-black text-primary;
  }

  /* ── Empty cart ────────────────────────────────────────── */
  .empty-card {
    @apply bg-surface rounded-3xl p-10 text-center border border-border shadow-sm max-w-md mx-auto;
  }
  .empty-title {
    @apply text-xl font-bold text-text mb-2;
  }
  .empty-msg {
    @apply text-sm text-text-muted;
  }
</style>
