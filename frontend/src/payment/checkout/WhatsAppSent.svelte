<script lang="ts">
  import { onMount } from 'svelte';
  import { ArrowLeft, CheckCheck, CheckCircle2, Info, MessageCircle, MoreVertical, Phone, Video } from 'lucide-svelte';
  import { formatPrice, type ServiceLocation } from './checkout';

  export interface OrderItem {
    productId: string;
    name: string;
    price: number;
    quantity: number;
  }

  export interface Order {
    id: string;
    number: number;
    clientId: string;
    customerName: string;
    phone: string;
    location: ServiceLocation;
    items: OrderItem[];
    notes: string;
    subtotal: number;
    deliveryFee: number;
    discount?: number;
    coupons?: string[];
    channel?: 'app' | 'whatsapp';
    shipping?: 'padrao' | 'expressa';
    total: number;
    payment: {
      method: string;
      cardType?: string;
      installments?: number;
      needsMachine?: boolean;
      changeFor?: number;
      status: 'pendente' | 'pago';
    };
    status: 'pendente' | 'preparo' | 'pronto' | 'entregue';
    createdAt: number;
  }

  // --- Abstracted State & Dependencies ---
  export let order: Order | null = null;
  export let restaurantName: string = 'Pratello Restaurante';
  export let onNavigate: (path: string) => void = (path) => (window.location.href = path);

  type Stage = 'sending' | 'sent' | 'read' | 'typing' | 'replied';
  let stage: Stage = 'sending';

  onMount(() => {
    const steps: [Stage, number][] = [
      ['sent', 700],
      ['read', 1600],
      ['typing', 2300],
      ['replied', 4200],
    ];
    const timers = steps.map(([s, ms]) => setTimeout(() => (stage = s), ms));
    return () => timers.forEach(clearTimeout);
  });

  const formatLocationFull = (loc: ServiceLocation): string => {
    if (loc.mode === 'mesa') return `Mesa ${loc.table}`;
    const a = loc.address;
    return `${a.street}, ${a.number}${a.complement ? ` - ${a.complement}` : ''} · ${a.district} · CEP ${a.cep}`;
  };

  const formatPayment = (p: Order['payment']): string => {
    if (p.method === 'cartao') {
      const type = p.cardType === 'debito' ? 'Débito' : 'Crédito';
      const parcelas = p.cardType !== 'debito' && p.installments && p.installments > 1 ? ` ${p.installments}x` : '';
      return `Cartão de ${type}${parcelas}${p.needsMachine ? ' · levar maquininha' : ''}`;
    }
    if (p.method === 'dinheiro') return p.changeFor ? `Dinheiro · troco para ${formatPrice(p.changeFor)}` : 'Dinheiro · sem troco';
    return 'Pix';
  };

  const formatTime = (timestamp: number) =>
    new Date(timestamp).toLocaleTimeString('pt-BR', { hour: '2-digit', minute: '2-digit' });

  function buildMessageLines(o: Order): string[] {
    const lines = [
      `*Novo pedido #${o.number}* 🍽️`,
      '',
      ...o.items.map((it) => `${it.quantity}x ${it.name} — ${formatPrice(it.price * it.quantity)}`),
      '',
      `Subtotal: ${formatPrice(o.subtotal)}`,
      `Frete: ${o.deliveryFee > 0 ? formatPrice(o.deliveryFee) : 'Grátis'}`,
    ];
    if (o.discount) lines.push(`Desconto: -${formatPrice(o.discount)}`);
    if (o.coupons?.length) lines.push(`Cupons: ${o.coupons.join(', ')}`);
    lines.push(
      `*Total: ${formatPrice(o.total)}*`,
      '',
      `👤 ${o.customerName}`,
      `📍 ${formatLocationFull(o.location)}`,
      `💳 ${formatPayment(o.payment)}`
    );
    if (o.notes) lines.push(`📝 ${o.notes}`);
    return lines;
  }

  $: time = order ? formatTime(order.createdAt) : '';
  $: firstName = order?.customerName ? order.customerName.split(' ')[0] : 'Cliente';
  $: eta = order?.shipping === 'expressa' ? '20 a 35 minutos' : '40 a 60 minutos';
  $: done = stage === 'replied';
</script>

{#if !order}
  <div class="grid min-h-screen place-items-center bg-zinc-100 p-4">
    <div class="rounded-xl bg-white p-6 text-center shadow ring-1 ring-zinc-200">
      <p class="font-bold text-zinc-900">Pedido não encontrado</p>
      <button
        on:click={() => onNavigate('/cardapio')}
        class="mt-4 rounded-xl bg-yellow-400 px-4 py-2 font-semibold text-zinc-900 hover:bg-yellow-500"
      >
        Voltar ao cardápio
      </button>
    </div>
  </div>
{:else}
  <div class="flex min-h-screen flex-col items-center bg-zinc-200 sm:py-6">
    <p class="flex w-full max-w-md items-center justify-center gap-1.5 bg-zinc-900 px-3 py-1.5 text-center text-[11px] font-medium text-zinc-300 sm:mb-2 sm:rounded-full">
      <Info size={12} /> Simulação para demonstração — nenhuma mensagem real foi enviada
    </p>

    <div class="flex w-full max-w-md flex-1 flex-col overflow-hidden bg-[#efeae2] shadow-xl sm:flex-none sm:rounded-2xl">
      <!-- Cabeçalho do chat -->
      <header class="flex items-center gap-3 bg-[#008069] px-3 py-2.5 text-white">
        <button on:click={() => onNavigate('/cardapio')} class="rounded-full p-1 hover:bg-white/10" aria-label="Voltar ao cardápio">
          <ArrowLeft size={20} />
        </button>
        <div class="grid h-10 w-10 shrink-0 place-items-center rounded-full bg-yellow-400 text-sm font-extrabold text-zinc-900">P</div>
        <div class="min-w-0 flex-1 leading-tight">
          <p class="flex items-center gap-1 font-semibold">
            {restaurantName} <CheckCircle2 size={14} class="fill-white text-[#008069]" />
          </p>
          <p class="text-xs text-white/80">{stage === 'typing' ? 'digitando…' : 'online'}</p>
        </div>
        <Video size={20} class="opacity-90" />
        <Phone size={18} class="opacity-90" />
        <MoreVertical size={18} class="opacity-90" />
      </header>

      <!-- Conversa -->
      <div
        class="flex-1 space-y-2 overflow-y-auto px-3 py-4"
        style="background-image: radial-gradient(rgba(0,0,0,0.04) 1px, transparent 1px); background-size: 14px 14px;"
      >
        <p class="mx-auto w-fit rounded-lg bg-white/90 px-3 py-1 text-[11px] font-medium uppercase text-zinc-500 shadow-sm">Hoje</p>
        <p class="mx-auto max-w-[85%] rounded-lg bg-[#fff5c4] px-3 py-1.5 text-center text-[11px] text-zinc-600 shadow-sm">
          🔒 Esta é uma conta comercial. Pedidos enviados por aqui entram direto na cozinha.
        </p>

        <!-- Mensagem do cliente -->
        <div class="flex justify-end">
          <div class="relative max-w-[85%] rounded-lg rounded-tr-none bg-[#d9fdd3] px-2.5 pb-1 pt-1.5 text-sm text-zinc-900 shadow-sm">
            <div class="whitespace-pre-wrap">
              {#each buildMessageLines(order) as line}
                {#if line.startsWith('*') && line.endsWith('*')}
                  <b>{line.slice(1, -1)}</b><br />
                {:else}
                  <span>{line}</span><br />
                {/if}
              {/each}
            </div>
            <p class="mt-0.5 flex items-center justify-end gap-1 text-[10px] text-zinc-500">
              {time}
              {#if stage === 'sending'}
                <span class="inline-block h-2.5 w-2.5 rounded-full border border-zinc-400" />
              {:else}
                <CheckCheck size={15} class={stage === 'sent' ? 'text-zinc-400' : 'text-sky-500'} />
              {/if}
            </p>
          </div>
        </div>

        {#if stage === 'typing'}
          <div class="flex">
            <div class="flex gap-1 rounded-lg rounded-tl-none bg-white px-3 py-3 shadow-sm">
              <span class="h-2 w-2 animate-bounce rounded-full bg-zinc-400" />
              <span class="h-2 w-2 animate-bounce rounded-full bg-zinc-400" style="animation-delay: 150ms;" />
              <span class="h-2 w-2 animate-bounce rounded-full bg-zinc-400" style="animation-delay: 300ms;" />
            </div>
          </div>
        {/if}

        {#if done}
          <div class="flex">
            <div class="max-w-[85%] rounded-lg rounded-tl-none bg-white px-2.5 pb-1 pt-1.5 text-sm text-zinc-900 shadow-sm">
              <p class="whitespace-pre-wrap">
                Olá, {firstName}! 👋<br /><br />
                <b>Pedido #{order.number} recebido</b> ✅<br />
                Já foi enviado para a cozinha.<br /><br />
                🛵 Previsão de entrega: {eta}.<br />
                {order.payment.status === 'pago' ? 'Pagamento confirmado. ' : ''}Obrigado por pedir conosco! 💛
              </p>
              <p class="mt-0.5 text-right text-[10px] text-zinc-500">{time}</p>
            </div>
          </div>
        {/if}
      </div>

      <!-- Rodapé / Resultado -->
      <div class="border-t border-black/5 bg-white p-4">
        {#if done}
          <div class="text-center">
            <div class="mx-auto mb-2 grid h-14 w-14 place-items-center rounded-full bg-green-100 text-green-600">
              <CheckCircle2 size={30} />
            </div>
            <p class="text-lg font-extrabold text-zinc-900">Pedido enviado com sucesso!</p>
            <p class="mt-1 text-sm text-zinc-600">
              O restaurante confirmou o pedido <b>#{order.number}</b> de <b>{formatPrice(order.total)}</b>.
            </p>
            <button
              on:click={() => onNavigate('/cardapio')}
              class="mt-4 inline-flex w-full items-center justify-center gap-2 rounded-xl bg-[#008069] px-4 py-3 font-semibold text-white hover:bg-[#006e5a]"
            >
              Voltar ao cardápio
            </button>
          </div>
        {:else}
          <p class="flex items-center justify-center gap-2 py-3 text-sm text-zinc-500">
            <MessageCircle size={16} class="animate-pulse text-[#008069]" /> Enviando pedido pelo WhatsApp…
          </p>
        {/if}
      </div>
    </div>
  </div>
{/if}
