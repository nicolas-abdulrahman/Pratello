<script lang="ts">
  import {
    AlertCircle,
    ArrowLeft,
    Bike,
    ChevronRight,
    MessageCircle,
    PiggyBank,
    ShoppingBag,
    Ticket,
    Truck,
    X,
    Zap
  } from 'lucide-svelte';

  import LocationSelector from './components/LocationSelector.svelte';
  import PaymentSelector from './components/PaymentSelector.svelte';
  import CouponModal from './components/CouponModal.svelte';

  import {
    COUPONS,
    computeCheckout,
    couponProblem,
    findCoupon,
    formatPrice,
    type CheckoutMode,
    type Coupon,
    type CartLine,
    type ShippingId,
    type Address,
    type ServiceLocation,
    type PaymentDraft,
    type RestaurantDependencies
  } from './checkout';

  // --- Abstracted State & Dependencies ---
  interface Props {
    cartLines?: CartLine[];
    subtotal?: number;
    clientSession?: { clientId?: string; name?: string; phone?: string; location?: ServiceLocation } | null;
    restaurantInfo?: {
      name: string;
      pixKey: string;
      pixMerchant: string;
      tableCount: number;
      phone: string;
    };
    dependencies?: Partial<RestaurantDependencies>;
    onPlaceOrder?: (payload: any) => Promise<{ id: string; number: number } | null>;
    onNavigate?: (path: string, state?: any) => void;
  }

  let {
    cartLines = [],
    subtotal = 0,
    clientSession = null,
    restaurantInfo = {
      name: 'Pratello Restaurante',
      pixKey: '12.345.678/0001-90',
      pixMerchant: 'PRATELLO RESTAURANTE LTDA',
      tableCount: 15,
      phone: '(11) 98765-4321',
    },
    dependencies = {},
    onPlaceOrder = dependencies.placeOrder ?? (async () => null),
    onNavigate = dependencies.navigateTo ?? ((p) => (window.location.href = p)),
  }: Props = $props();

  // --- Internal Page State ---
  let mode: CheckoutMode = $state(clientSession?.location?.mode ?? 'mesa');
  let table: string = $state(clientSession?.location?.mode === 'mesa' ? clientSession.location.table : '01');
  let address: Address = $state(
    clientSession?.location?.mode === 'delivery'
      ? { ...clientSession.location.address }
      : { cep: '', street: '', number: '', district: '', complement: '' }
  );
  let phone: string = $state(clientSession?.phone ?? '');
  let shipping: ShippingId = $state('padrao');
  let couponCodes: string[] = $state([]);
  let notes: string = $state('');
  let couponsOpen: boolean = $state(false);
  let error: string = $state('');

  let payment: PaymentDraft = $state({
    method: null,
    cardType: 'credito',
    installments: 1,
    needsMachine: true,
    changeFor: '',
    pixPaid: false,
  });

  let isShipped = $derived(mode !== 'mesa');
  let itemCount = $derived(cartLines.reduce((s, l) => s + l.quantity, 0));

  let summary = $derived(
    computeCheckout({
      subtotal,
      mode,
      payment: payment.method,
      cep: address.cep,
      shipping,
      couponCodes,
    })
  );

  let ctx = $derived({ subtotal, mode, payment: payment.method });
  let availableCouponsCount = $derived(COUPONS.filter((c) => !couponProblem(c, ctx)).length);
  let appliedProduct = $derived(summary.applied.find((c) => c.kind === 'desconto'));
  let appliedShipping = $derived(summary.applied.find((c) => c.kind === 'frete'));
  let submitLabel = $derived(mode === 'whatsapp' ? 'Enviar pelo WhatsApp' : 'Fazer pedido');

  function clearError() {
    error = '';
  }

  function toggleCoupon(c: Coupon) {
    clearError();
    if (couponCodes.includes(c.code)) {
      couponCodes = couponCodes.filter((x) => x !== c.code);
    } else {
      couponCodes = [...couponCodes.filter((x) => findCoupon(x)?.kind !== c.kind), c.code];
    }
  }

  async function submit() {
    if (mode === 'mesa' && !table) {
      error = 'Selecione o número da mesa.';
      return;
    }
    if (mode !== 'mesa') {
      if (address.cep.replace(/\D/g, '').length !== 8) {
        error = 'Informe um CEP válido.';
        return;
      }
      if (!address.street.trim() || !address.number.trim() || !address.district.trim()) {
        error = 'Preencha rua, número e bairro.';
        return;
      }
      if (phone.replace(/\D/g, '').length < 10) {
        error = 'Informe um telefone com DDD para contato.';
        return;
      }
    }
    if (summary.invalid.length) {
      error = 'Remova os cupons que não valem para este pedido.';
      return;
    }
    if (!payment.method) {
      error = 'Escolha a forma de pagamento.';
      return;
    }

    const location: ServiceLocation =
      mode === 'mesa' ? { mode: 'mesa', table } : { mode: 'delivery', address };

    const payload = {
      location,
      channel: mode === 'whatsapp' ? ('whatsapp' as const) : ('app' as const),
      phone,
      notes,
      payment,
      shipping,
      couponCodes,
      summary,
    };

    const order = await onPlaceOrder(payload);
    if (!order) {
      error = 'Não foi possível enviar o pedido. Tente novamente.';
      return;
    }

    if (mode === 'whatsapp') {
      onNavigate(`/pedido-whatsapp/${order.id}`);
    } else {
      onNavigate('/cardapio', { placedOrderId: order.id });
    }
  }
</script>

{#if cartLines.length === 0}
  <div class="grid min-h-screen place-items-center bg-zinc-100 p-4">
    <div class="max-w-sm rounded-2xl bg-white p-8 text-center shadow-sm ring-1 ring-zinc-200">
      <div class="mx-auto mb-3 grid h-14 w-14 place-items-center rounded-full bg-zinc-100 text-zinc-500">
        <ShoppingBag size={24} />
      </div>
      <p class="font-bold text-zinc-900">Seu carrinho está vazio</p>
      <p class="mt-1 text-sm text-zinc-600">Adicione pratos do cardápio para finalizar um pedido.</p>
      <button
        onclick={() => onNavigate('/cardapio')}
        class="mt-5 inline-flex rounded-xl bg-yellow-400 px-4 py-2.5 text-sm font-semibold text-zinc-900 hover:bg-yellow-500"
      >
        Ver cardápio
      </button>
    </div>
  </div>
{:else}
  <div class="min-h-screen bg-zinc-100 pb-28 lg:pb-10">
    <!-- Header -->
    <header class="sticky top-0 z-30 border-b border-zinc-200 bg-white/95 backdrop-blur">
      <div class="mx-auto flex max-w-6xl items-center gap-3 px-4 py-3 sm:px-6">
        <button
          onclick={() => onNavigate('/cardapio')}
          class="rounded-lg p-1.5 text-zinc-600 hover:bg-zinc-100"
          aria-label="Voltar ao cardápio"
        >
          <ArrowLeft size={20} />
        </button>
        <div class="flex items-center gap-2">
          <span class="grid h-8 w-8 place-items-center rounded-lg bg-yellow-400 font-extrabold text-zinc-900">P</span>
          <span class="font-bold text-zinc-900">{restaurantInfo.name}</span>
        </div>
        <span class="h-6 w-px bg-zinc-200" />
        <h1 class="text-lg font-bold text-zinc-900">Finalizar pedido</h1>
      </div>
    </header>

    <div class="mx-auto grid max-w-6xl gap-4 px-4 py-5 sm:px-6 lg:grid-cols-[1fr_360px] lg:items-start">
      <main class="min-w-0 space-y-4">
        <!-- 1. Como receber (LocationSelector component) -->
        <LocationSelector
          bind:mode
          bind:table
          bind:address
          bind:phone
          tableCount={restaurantInfo.tableCount}
          fetchAddressFromCep={dependencies.fetchAddressFromCep}
          onClearError={clearError}
        />

        <!-- 2. Itens do pedido -->
        <section class="rounded-2xl bg-white p-4 shadow-sm ring-1 ring-zinc-200 sm:p-5">
          <div class="mb-3 flex items-center justify-between gap-2">
            <h2 class="flex items-center gap-2 font-bold text-zinc-900">
              <span class="text-yellow-600"><ShoppingBag size={18} /></span> Seu pedido
            </h2>
            <button
              onclick={() => onNavigate('/cardapio')}
              class="text-xs font-bold text-yellow-600 hover:text-yellow-700"
            >
              Editar itens
            </button>
          </div>
          <ul class="divide-y divide-zinc-100">
            {#each cartLines as { product, quantity } (product.id)}
              <li class="flex items-center gap-3 py-3 first:pt-0">
                {#if product.image}
                  <img src={product.image} alt={product.name} class="h-14 w-14 shrink-0 rounded-xl object-cover" />
                {:else}
                  <div class="grid h-14 w-14 shrink-0 place-items-center rounded-xl bg-zinc-200 text-zinc-400">
                    <ShoppingBag size={20} />
                  </div>
                {/if}
                <div class="min-w-0 flex-1">
                  <p class="truncate text-sm font-semibold text-zinc-900">{product.name}</p>
                  <p class="text-xs text-zinc-500">
                    {formatPrice(product.price)} <span class="text-zinc-400">× {quantity}</span>
                    {#if product.originalPrice && product.originalPrice > product.price}
                      <span class="ml-1.5 text-zinc-400 line-through">{formatPrice(product.originalPrice)}</span>
                    {/if}
                  </p>
                </div>
                <span class="text-sm font-bold text-zinc-900">{formatPrice(product.price * quantity)}</span>
              </li>
            {/each}
          </ul>
          <label class="mt-2 block">
            <span class="mb-1.5 block text-sm font-medium text-zinc-700">Mensagem para o restaurante</span>
            <textarea
              rows="2"
              bind:value={notes}
              placeholder="Ex: sem cebola, ponto da carne…"
              class="w-full resize-none rounded-xl border border-zinc-200 bg-white px-3.5 py-2.5 text-sm text-zinc-900 placeholder:text-zinc-400 focus:border-yellow-400 focus:outline-none focus:ring-2 focus:ring-yellow-400/40"
            ></textarea>
          </label>
        </section>

        <!-- 3. Frete / Opções de entrega -->
        {#if isShipped}
          <section class="rounded-2xl bg-white p-4 shadow-sm ring-1 ring-zinc-200 sm:p-5">
            <h2 class="mb-3 flex items-center gap-2 font-bold text-zinc-900">
              <span class="text-yellow-600"><Truck size={18} /></span> Opções de entrega
            </h2>
            {#if summary.km === 0}
              <p class="rounded-xl bg-zinc-50 p-3 text-sm text-zinc-600 ring-1 ring-zinc-200">
                Informe um CEP válido para calcular o frete.
              </p>
            {:else}
              <p class="mb-3 text-xs text-zinc-500">
                Distância estimada: <b class="text-zinc-700">{summary.km.toLocaleString('pt-BR')} km</b> do restaurante
              </p>
              <div class="grid gap-2 sm:grid-cols-2" role="radiogroup" aria-label="Opção de entrega">
                {#each summary.options as o (o.id)}
                  {@const active = shipping === o.id}
                  <button
                    type="button"
                    role="radio"
                    aria-checked={active}
                    onclick={() => (shipping = o.id)}
                    class={`flex items-center gap-3 rounded-xl border-2 p-3 text-left transition-colors ${
                      active ? 'border-yellow-400 bg-yellow-50' : 'border-zinc-200 hover:border-zinc-300'
                    }`}
                  >
                    <span
                      class={`grid h-9 w-9 shrink-0 place-items-center rounded-lg ${
                        active ? 'bg-yellow-400' : 'bg-zinc-100 text-zinc-500'
                      }`}
                    >
                      {#if o.id === 'expressa'}
                        <Zap size={18} />
                      {:else}
                        <Bike size={18} />
                      {/if}
                    </span>
                    <span class="min-w-0 flex-1">
                      <span class="block text-sm font-bold text-zinc-900">Entrega {o.name}</span>
                      <span class="block text-xs text-zinc-500">Chega em {o.eta}</span>
                    </span>
                    <span class="text-right text-sm font-bold">
                      {#if o.fee === 0}
                        <span class="text-green-600">Grátis</span>
                      {:else}
                        {formatPrice(o.fee)}
                      {/if}
                    </span>
                  </button>
                {/each}
              </div>
            {/if}
          </section>
        {/if}

        <!-- 4. Cupons -->
        <section class="rounded-2xl bg-white shadow-sm ring-1 ring-zinc-200">
          <button
            type="button"
            onclick={() => (couponsOpen = true)}
            class="flex w-full items-center gap-3 p-4 text-left sm:p-5"
          >
            <Ticket size={20} class="shrink-0 text-yellow-600" />
            <span class="min-w-0 flex-1">
              <span class="block font-bold text-zinc-900">Cupons Pratello</span>
              {#if summary.applied.length === 0}
                <span class="block text-xs text-zinc-500">{availableCouponsCount} cupons disponíveis para este pedido</span>
              {:else}
                <span class="mt-1 flex flex-wrap gap-1.5">
                  {#each summary.applied as c (c.code)}
                    <span class="rounded border border-dashed border-yellow-500 bg-yellow-50 px-1.5 py-0.5 text-[11px] font-bold text-yellow-800">
                      {c.code}
                    </span>
                  {/each}
                </span>
              {/if}
            </span>
            {#if summary.discount + summary.shippingDiscount > 0}
              <span class="text-sm font-bold text-green-600">- {formatPrice(summary.discount + summary.shippingDiscount)}</span>
            {:else}
              <span class="text-sm text-zinc-500">Selecionar</span>
            {/if}
            <ChevronRight size={18} class="shrink-0 text-zinc-400" />
          </button>

          {#if summary.invalid.length > 0}
            <div class="space-y-1.5 border-t border-zinc-100 px-4 py-3 sm:px-5">
              {#each summary.invalid as { coupon, reason } (coupon.code)}
                <p class="flex items-center gap-2 text-xs text-red-600">
                  <AlertCircle size={14} class="shrink-0" />
                  <span class="flex-1"><b>{coupon.code}</b> não vale agora: {reason.toLowerCase()}</span>
                  <button
                    type="button"
                    onclick={() => toggleCoupon(coupon)}
                    class="rounded p-0.5 hover:bg-red-50"
                    aria-label={`Remover ${coupon.code}`}
                  >
                    <X size={14} />
                  </button>
                </p>
              {/each}
            </div>
          {/if}
        </section>

        <!-- 5. Forma de Pagamento (PaymentSelector component with PixCard) -->
        <section class="rounded-2xl bg-white p-4 shadow-sm ring-1 ring-zinc-200 sm:p-5">
          <h2 class="mb-3 flex items-center gap-2 font-bold text-zinc-900">
            <span class="text-yellow-600"><PiggyBank size={18} /></span> Forma de pagamento
          </h2>

          <PaymentSelector
            bind:payment
            total={summary.total}
            isDelivery={isShipped}
            pixKey={restaurantInfo.pixKey}
            pixMerchant={restaurantInfo.pixMerchant}
            onPaymentChange={clearError}
          />
        </section>

        <!-- Resumo mobile -->
        <section class="rounded-2xl bg-white p-4 shadow-sm ring-1 ring-zinc-200 lg:hidden">
          <h2 class="mb-3 font-bold text-zinc-900">Detalhes do pagamento</h2>
          <div class="space-y-2 text-sm">
            <div class="flex justify-between gap-3">
              <span class="text-zinc-600">Subtotal ({itemCount} {itemCount === 1 ? 'item' : 'itens'})</span>
              <span class="text-zinc-900">{formatPrice(summary.subtotal)}</span>
            </div>
            {#if isShipped}
              <div class="flex justify-between gap-3">
                <span class="text-zinc-600">Frete</span>
                {#if summary.km === 0}
                  <span class="text-zinc-400">informe o CEP</span>
                {:else if summary.shippingFee === 0}
                  <b class="text-green-600">Grátis</b>
                {:else}
                  <span class="text-zinc-900">{formatPrice(summary.shippingFee)}</span>
                {/if}
              </div>
            {/if}
            {#if summary.shippingDiscount > 0}
              <div class="flex justify-between gap-3">
                <span class="text-zinc-600">Cupom de frete ({appliedShipping?.code})</span>
                <span class="font-semibold text-green-600">- {formatPrice(summary.shippingDiscount)}</span>
              </div>
            {/if}
            {#if summary.discount > 0}
              <div class="flex justify-between gap-3">
                <span class="text-zinc-600">Cupom de desconto ({appliedProduct?.code})</span>
                <span class="font-semibold text-green-600">- {formatPrice(summary.discount)}</span>
              </div>
            {/if}
            <div class="flex items-baseline justify-between border-t border-zinc-100 pt-3">
              <span class="font-semibold text-zinc-900">Total</span>
              <span class="text-2xl font-extrabold text-zinc-900">{formatPrice(summary.total)}</span>
            </div>
            {#if summary.savings > 0}
              <p class="flex items-center justify-end gap-1 text-xs font-semibold text-green-700">
                <PiggyBank size={14} /> Você economizou {formatPrice(summary.savings)}
              </p>
            {/if}
          </div>
        </section>
      </main>

      <!-- Resumo desktop -->
      <aside class="sticky top-20 hidden rounded-2xl bg-white p-5 shadow-sm ring-1 ring-zinc-200 lg:block">
        <h2 class="mb-4 font-bold text-zinc-900">Detalhes do pagamento</h2>
        <div class="space-y-2 text-sm">
          <div class="flex justify-between gap-3">
            <span class="text-zinc-600">Subtotal ({itemCount} {itemCount === 1 ? 'item' : 'itens'})</span>
            <span class="text-zinc-900">{formatPrice(summary.subtotal)}</span>
          </div>
          {#if isShipped}
            <div class="flex justify-between gap-3">
              <span class="text-zinc-600">Frete</span>
              {#if summary.km === 0}
                <span class="text-zinc-400">informe o CEP</span>
              {:else if summary.shippingFee === 0}
                <b class="text-green-600">Grátis</b>
              {:else}
                <span class="text-zinc-900">{formatPrice(summary.shippingFee)}</span>
              {/if}
            </div>
          {/if}
          {#if summary.shippingDiscount > 0}
            <div class="flex justify-between gap-3">
              <span class="text-zinc-600">Cupom de frete ({appliedShipping?.code})</span>
              <span class="font-semibold text-green-600">- {formatPrice(summary.shippingDiscount)}</span>
            </div>
          {/if}
          {#if summary.discount > 0}
            <div class="flex justify-between gap-3">
              <span class="text-zinc-600">Cupom de desconto ({appliedProduct?.code})</span>
              <span class="font-semibold text-green-600">- {formatPrice(summary.discount)}</span>
            </div>
          {/if}
          <div class="flex items-baseline justify-between border-t border-zinc-100 pt-3">
            <span class="font-semibold text-zinc-900">Total</span>
            <span class="text-2xl font-extrabold text-zinc-900">{formatPrice(summary.total)}</span>
          </div>
          {#if summary.savings > 0}
            <p class="flex items-center justify-end gap-1 text-xs font-semibold text-green-700">
              <PiggyBank size={14} /> Você economizou {formatPrice(summary.savings)}
            </p>
          {/if}
        </div>

        {#if error}
          <p class="mt-3 text-sm font-medium text-red-600">{error}</p>
        {/if}

        <button
          type="button"
          onclick={submit}
          class={`mt-4 inline-flex w-full items-center justify-center gap-2 rounded-xl px-5 py-3 font-semibold transition-colors ${
            mode === 'whatsapp'
              ? 'bg-green-600 text-white hover:bg-green-700'
              : 'bg-yellow-400 text-zinc-900 hover:bg-yellow-500'
          }`}
        >
          {#if mode === 'whatsapp'}
            <MessageCircle size={18} />
          {/if}
          {submitLabel}
        </button>
      </aside>
    </div>

    <!-- Barra inferior mobile -->
    <div class="fixed inset-x-0 bottom-0 z-30 border-t border-zinc-200 bg-white lg:hidden">
      {#if error}
        <p class="bg-red-50 px-4 py-2 text-sm font-medium text-red-600">{error}</p>
      {/if}
      <div class="flex items-center gap-3 p-3">
        <div class="min-w-0 flex-1 text-right leading-tight">
          <p class="text-xs text-zinc-500">
            Total <b class="text-lg text-zinc-900">{formatPrice(summary.total)}</b>
          </p>
          {#if summary.savings > 0}
            <p class="text-[11px] font-semibold text-green-700">Economia de {formatPrice(summary.savings)}</p>
          {/if}
        </div>
        <button
          type="button"
          onclick={submit}
          class={`inline-flex items-center justify-center gap-2 rounded-xl px-5 py-3 font-semibold transition-colors ${
            mode === 'whatsapp'
              ? 'bg-green-600 text-white hover:bg-green-700'
              : 'bg-yellow-400 text-zinc-900 hover:bg-yellow-500'
          }`}
        >
          {#if mode === 'whatsapp'}
            <MessageCircle size={18} />
          {/if}
          {submitLabel}
        </button>
      </div>
    </div>
  </div>

  <!-- Coupon Selection Modal -->
  <CouponModal
    open={couponsOpen}
    selectedCodes={couponCodes}
    {ctx}
    onClose={() => (couponsOpen = false)}
    onToggleCoupon={toggleCoupon}
  />
{/if}
