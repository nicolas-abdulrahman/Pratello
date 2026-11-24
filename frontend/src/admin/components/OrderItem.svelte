<script lang="ts">
  import {
    Armchair,
    ArrowRight,
    Banknote,
    Bike,
    Check,
    Clock,
    CreditCard,
    MessageCircle,
    MessageSquareText,
    Phone,
    QrCode,
    Ticket,
    User,
    Zap
  } from 'lucide-svelte';

  export interface OrderProductItem {
    productId: string;
    name: string;
    price: number;
    quantity: number;
  }

  export interface Order {
    id: string;
    number: number;
    customerName: string;
    phone?: string;
    location:
      | { mode: 'mesa'; table: string }
      | { mode: 'delivery'; address: { street: string; number: string; complement?: string; district: string; cep: string } };
    items: OrderProductItem[];
    notes?: string;
    subtotal: number;
    deliveryFee: number;
    discount?: number;
    coupons?: string[];
    channel?: 'app' | 'whatsapp';
    shipping?: 'padrao' | 'expressa';
    total: number;
    payment: {
      method: 'cartao' | 'pix' | 'dinheiro' | string;
      cardType?: 'credito' | 'debito' | string;
      installments?: number;
      needsMachine?: boolean;
      changeFor?: number;
      status: 'pendente' | 'pago';
    };
    status: 'pendente' | 'preparo' | 'pronto' | 'entregue';
    createdAt: number;
  }

  interface Props {
    order: Order;
    onAdvance?: () => void;
    onMarkPaid?: () => void;
  }

  let { order, onAdvance = () => {}, onMarkPaid = () => {} }: Props = $props();

  const NEXT_ACTION_LABEL: Record<string, string | null> = {
    pendente: 'Iniciar preparo',
    preparo: 'Marcar como pronto',
    pronto: 'Marcar como entregue',
    entregue: null,
  };

  const formatPrice = (value: number) =>
    new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(value);

  const formatTime = (timestamp: number) =>
    new Date(timestamp).toLocaleTimeString('pt-BR', { hour: '2-digit', minute: '2-digit' });

  const minutesSince = (timestamp: number) =>
    Math.max(0, Math.floor((Date.now() - timestamp) / 60000));

  function formatPayment(p: Order['payment']): string {
    if (p.method === 'cartao') {
      const type = p.cardType === 'debito' ? 'Débito' : 'Crédito';
      const parcelas = p.cardType !== 'debito' && p.installments && p.installments > 1 ? ` ${p.installments}x` : '';
      return `Cartão de ${type}${parcelas}${p.needsMachine ? ' · levar maquininha' : ''}`;
    }
    if (p.method === 'dinheiro') {
      return p.changeFor ? `Dinheiro · troco para ${formatPrice(p.changeFor)}` : 'Dinheiro · sem troco';
    }
    return 'Pix';
  }

  let action = $derived(NEXT_ACTION_LABEL[order.status] ?? null);
  let waiting = $derived(minutesSince(order.createdAt));
  let late = $derived(order.status !== 'entregue' && waiting >= 20);
  let loc = $derived(order.location);
  let paid = $derived(order.payment.status === 'pago');
</script>

<article class="rounded-xl border border-zinc-200 bg-white p-4 shadow-sm">
  <header class="flex items-start justify-between gap-2">
    <div class="min-w-0">
      <p class="text-lg font-extrabold text-zinc-900">#{order.number}</p>
      <p class="flex items-center gap-1 truncate text-sm text-zinc-600">
        <User size={14} class="shrink-0" /> {order.customerName}
      </p>
    </div>
    <span
      class={`inline-flex shrink-0 items-center gap-1 rounded-lg px-2 py-1 text-xs font-semibold ${
        late ? 'bg-red-50 text-red-600' : 'bg-zinc-100 text-zinc-600'
      }`}
      title={`Pedido feito às ${formatTime(order.createdAt)}`}
    >
      <Clock size={12} /> {formatTime(order.createdAt)}
      {#if order.status !== 'entregue'}
        <span>· {waiting} min</span>
      {/if}
    </span>
  </header>

  <!-- Localização -->
  {#if loc.mode === 'mesa'}
    <div class="mt-3 flex items-center gap-2 rounded-lg bg-zinc-900 px-3 py-2 text-white">
      <Armchair size={16} class="text-yellow-400" />
      <span class="text-sm font-bold">Mesa {loc.table}</span>
      <span class="ml-auto text-xs text-zinc-400">Salão</span>
    </div>
  {:else}
    <div class="mt-3 rounded-lg bg-yellow-50 px-3 py-2 text-sm ring-1 ring-yellow-200">
      <p class="flex items-center gap-1.5 font-bold text-zinc-900">
        <Bike size={16} /> Delivery
        {#if order.shipping === 'expressa'}
          <span class="inline-flex items-center gap-0.5 rounded bg-zinc-900 px-1.5 py-0.5 text-[10px] text-yellow-400">
            <Zap size={10} /> Expressa
          </span>
        {/if}
        {#if order.channel === 'whatsapp'}
          <span class="ml-auto inline-flex items-center gap-1 rounded-full bg-green-600 px-2 py-0.5 text-[10px] font-bold text-white">
            <MessageCircle size={11} /> WhatsApp
          </span>
        {/if}
      </p>
      <p class="mt-0.5 text-zinc-800">
        {loc.address.street}, {loc.address.number}
        {#if loc.address.complement}
          - {loc.address.complement}
        {/if}
      </p>
      <p class="text-xs text-zinc-600">
        {loc.address.district} · CEP {loc.address.cep}
      </p>
      {#if order.phone}
        <p class="mt-0.5 flex items-center gap-1 text-xs text-zinc-600">
          <Phone size={12} /> {order.phone}
        </p>
      {/if}
    </div>
  {/if}

  <!-- Itens -->
  <ul class="mt-3 space-y-1 border-t border-zinc-100 pt-3 text-sm">
    {#each order.items as it (it.productId)}
      <li class="text-zinc-700">
        <b class="text-zinc-900">{it.quantity}x</b> {it.name}
      </li>
    {/each}
  </ul>

  {#if order.notes}
    <p class="mt-3 flex gap-1.5 rounded-lg bg-yellow-50 p-2 text-sm text-yellow-900 ring-1 ring-yellow-200">
      <MessageSquareText size={16} class="mt-0.5 shrink-0" /> {order.notes}
    </p>
  {/if}

  <!-- Pagamento -->
  <div class="mt-3 flex items-center gap-2 border-t border-zinc-100 pt-3 text-sm">
    {#if order.payment.method === 'cartao'}
      <CreditCard size={16} class="shrink-0 text-zinc-500" />
    {:else if order.payment.method === 'dinheiro'}
      <Banknote size={16} class="shrink-0 text-zinc-500" />
    {:else}
      <QrCode size={16} class="shrink-0 text-zinc-500" />
    {/if}

    <span class="min-w-0 flex-1 text-zinc-700">{formatPayment(order.payment)}</span>

    {#if paid}
      <span class="inline-flex shrink-0 items-center gap-1 rounded-full bg-green-100 px-2 py-0.5 text-xs font-semibold text-green-800">
        <Check size={12} /> Pago
      </span>
    {:else}
      <button
        type="button"
        onclick={onMarkPaid}
        title="Confirmar recebimento do pagamento"
        class="shrink-0 rounded-full bg-zinc-100 px-2 py-0.5 text-xs font-semibold text-zinc-700 ring-1 ring-zinc-300 hover:bg-yellow-100 hover:ring-yellow-400"
      >
        A receber · confirmar
      </button>
    {/if}
  </div>

  <footer class="mt-3 flex items-center justify-between gap-2">
    <div class="leading-tight">
      <span class="text-sm font-bold text-zinc-900">{formatPrice(order.total)}</span>
      {#if order.deliveryFee > 0}
        <p class="text-[11px] text-zinc-500">inclui entrega {formatPrice(order.deliveryFee)}</p>
      {/if}
      {#if order.coupons && order.coupons.length > 0}
        <p class="flex items-center gap-1 text-[11px] font-semibold text-green-700" title={order.coupons.join(', ')}>
          <Ticket size={11} /> {order.coupons.join(' + ')}
          {#if order.discount}
            · -{formatPrice(order.discount)}
          {/if}
        </p>
      {/if}
    </div>

    {#if action}
      <button
        type="button"
        onclick={onAdvance}
        class="inline-flex items-center gap-1 rounded-xl bg-yellow-400 px-3 py-1.5 text-sm font-semibold text-zinc-900 hover:bg-yellow-500"
      >
        {action} <ArrowRight size={14} />
      </button>
    {/if}
  </footer>
</article>
