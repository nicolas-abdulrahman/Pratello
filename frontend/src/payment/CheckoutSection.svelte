<script lang="ts">
  import { cart, cart_total, clear_cart } from '../lib/stores/Cart';
  import { visibleOrders, createOrder, preparePayment, confirmPayment, type SavedOrder } from '../lib/stores/Transaction';
  import { get } from 'svelte/store';

  let {
    onCancel = () => {},
    slug = '',
    onOrderConfirmed = undefined as ((restaurantId: string | number, orderId: string | number) => void) | undefined
  }: {
    onCancel?: () => void;
    slug?: string;
    total?: number;
    onOrderConfirmed?: (restaurantId: string | number, orderId: string | number) => void;
  } = $props();
  const resumeId = new URLSearchParams(window.location.search).get('resume');
  let saved = $state<SavedOrder | null>(visibleOrders().find((entry) => String(entry.order.id) === resumeId) ?? null);
  let busy = $state(false);
  let error = $state('');
  let paid = $state(false);

  async function start() {
    if (busy) return;
    busy = true;
    error = '';
    try {
      if (!saved) saved = await createOrder(slug, get(cart));
      saved = await preparePayment(saved);
    } catch (e) {
      error = e instanceof Error ? e.message : 'Não foi possível iniciar o pagamento.';
    } finally { busy = false; }
  }

  async function confirm() {
    if (!saved || busy) return;
    busy = true;
    error = '';
    try {
      if (import.meta.env.TEST_ENV !== 'true' && import.meta.env.TEST_ENV !== '1') throw new Error('A confirmação simulada PIX só está disponível no ambiente de teste.');
      saved = await confirmPayment(saved);
      paid = true;
      if (!resumeId) clear_cart();
      const restId = saved.restaurantId || saved.slug || slug || 'restaurante';
      const orderId = saved.order.id;
      if (onOrderConfirmed) {
        onOrderConfirmed(restId, orderId);
      } else {
        window.history.pushState({}, '', `/me/pedidos/${restId}/${orderId}`);
        window.dispatchEvent(new PopStateEvent('popstate'));
      }
    } catch (e) {
      error = e instanceof Error ? e.message : 'Não foi possível confirmar o pagamento.';
    } finally { busy = false; }
  }

  const money = (value: number) => (value / 100).toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' });
</script>

<main class="min-h-screen bg-background py-12 px-4">
  <div class="max-w-xl mx-auto bg-surface border border-border rounded-3xl p-8 space-y-6">
    <button type="button" onclick={onCancel} class="text-primary font-bold">← Voltar</button>
    <h1 class="text-3xl font-black">Finalizar pedido</h1>
    <p class="text-text-muted">Pedido no local · Pagamento PIX de teste</p>
    {#if paid}
      <p class="text-green-600 font-bold" role="status">Pagamento confirmado! Aguarde a entrega do pedido para avaliá-lo.</p>
      <a href="/me/pedidos" class="text-primary font-bold">Ver meus pedidos</a>
    {:else if saved?.pix}
      <p>Pedido #{saved.order.id} · Total {money(saved.order.total_cents)}</p>
      {#if saved.pix.qr_code_base64}
        <img class="w-48 h-48 mx-auto" src={`data:image/png;base64,${saved.pix.qr_code_base64}`} alt="Código QR PIX" />
      {/if}
      <label for="pix-code" class="block font-bold">PIX copia e cola</label>
      <input id="pix-code" class="w-full p-3 border border-border rounded-xl" readonly value={saved.pix.copy_paste_code} />
      <button type="button" class="text-primary font-bold" onclick={() => navigator.clipboard.writeText(saved!.pix!.copy_paste_code)}>Copiar código</button>
      <button type="button" disabled={busy} onclick={confirm} class="block w-full bg-primary text-white font-bold rounded-xl p-4 disabled:opacity-50">
        {busy ? 'Confirmando…' : 'Confirmar pagamento de teste'}
      </button>
    {:else}
      {#if $cart.length === 0 && !saved}
        <p>Seu carrinho está vazio.</p>
      {:else}
        <p>Subtotal estimado: {$cart_total.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })}. O total final será calculado pelo restaurante.</p>
        <button type="button" disabled={busy} onclick={start} class="w-full bg-primary text-white font-bold rounded-xl p-4 disabled:opacity-50">
          {busy ? 'Preparando pedido…' : saved ? 'Tentar novamente' : 'Criar pedido e gerar PIX'}
        </button>
      {/if}
    {/if}
    {#if error}<p role="alert" class="text-red-600">{error}</p>{/if}
  </div>
</main>
