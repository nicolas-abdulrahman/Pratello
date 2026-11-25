<script lang="ts">
  import { Banknote, Check, CreditCard, QrCode, Wallet } from 'lucide-svelte';
  import PixCard from './PixCard.svelte';
  import {
    maxInstallments,
    formatPrice,
    type PaymentDraft
  } from '../checkout';

  type RowId = 'pix' | 'credito' | 'debito' | 'dinheiro';

  interface Props {
    payment?: PaymentDraft;
    total: number;
    isDelivery?: boolean;
    pixKey?: string;
    pixMerchant?: string;
    onPaymentChange?: (payment: PaymentDraft) => void;
  }

  let {
    payment = $bindable({
      method: null,
      cardType: 'credito',
      installments: 1,
      needsMachine: true,
      changeFor: '',
      pixPaid: false,
    }),
    total,
    isDelivery = true,
    pixKey = '12.345.678/0001-90',
    pixMerchant = 'PRATELLO RESTAURANTE LTDA',
    onPaymentChange = () => {}
  }: Props = $props();

  const ROWS: { id: RowId; label: string; sub: string; icon: any; badge?: string }[] = [
    { id: 'pix', label: 'Pix', sub: 'Aprovação na hora', icon: QrCode, badge: 'Cupom PIX7' },
    // { id: 'credito', label: 'Cartão de crédito', sub: 'Até 3x sem juros', icon: CreditCard },
    // { id: 'debito', label: 'Cartão de débito', sub: 'Na maquininha', icon: Wallet },
    // { id: 'dinheiro', label: 'Dinheiro', sub: 'Troco opcional', icon: Banknote },
  ];

  let current = $derived(
    payment.method === 'cartao' ? payment.cardType : payment.method
  );

  let maxParc = $derived(maxInstallments(total));

  function pick(id: RowId) {
    if (id === 'credito' || id === 'debito') {
      payment.method = 'cartao';
      payment.cardType = id;
    } else {
      payment.method = id;
    }
    onPaymentChange(payment);
  }
</script>

<div role="radiogroup" aria-label="Forma de pagamento" class="divide-y divide-zinc-100">
  {#each ROWS as { id, label, sub, icon: Icon, badge } (id)}
    {@const active = current === id}
    <div>
      <button
        type="button"
        role="radio"
        aria-checked={active}
        onclick={() => pick(id)}
        class="flex w-full items-center gap-3 py-3 text-left"
      >
        <span
          class={`grid h-9 w-9 shrink-0 place-items-center rounded-lg ${
            active ? 'bg-yellow-400 text-zinc-900' : 'bg-zinc-100 text-zinc-500'
          }`}
        >
          <svelte:component this={Icon} size={18} />
        </span>
        <span class="min-w-0 flex-1">
          <span class="flex items-center gap-2 text-sm font-semibold text-zinc-900">
            {label}
            {#if badge}
              <span class="rounded border border-yellow-500 px-1 text-[10px] font-bold text-yellow-700">
                {badge}
              </span>
            {/if}
          </span>
          <span class="block text-xs text-zinc-500">{sub}</span>
        </span>
        <span
          class={`grid h-5 w-5 shrink-0 place-items-center rounded-full border-2 ${
            active ? 'border-yellow-500 bg-yellow-400' : 'border-zinc-300'
          }`}
        >
          {#if active}
            <Check size={12} strokeWidth={3} class="text-zinc-900" />
          {/if}
        </span>
      </button>

      {#if active}
        <div class="mb-3">
          {#if id === 'pix'}
            <PixCard
              {pixKey}
              {pixMerchant}
              {total}
              bind:paid={payment.pixPaid}
              onTogglePaid={(paid) => onPaymentChange(payment)}
            />
          {:else if id === 'credito'}
            <div class="space-y-3 rounded-xl bg-zinc-50 p-3 ring-1 ring-zinc-200">
              <div>
                <p class="mb-1.5 text-xs font-semibold text-zinc-600">Parcelas</p>
                <div class="grid grid-cols-3 gap-2">
                  {#each [1, 2, 3] as n}
                    {@const disabled = n > maxParc}
                    {@const isCurrent = Math.min(payment.installments, maxParc) === n}
                    <button
                      type="button"
                      {disabled}
                      onclick={() => {
                        payment.installments = n;
                        onPaymentChange(payment);
                      }}
                      class={`rounded-lg border px-2 py-2 text-left text-xs transition-colors disabled:cursor-not-allowed disabled:opacity-40 ${
                        isCurrent
                          ? 'border-yellow-400 bg-yellow-50'
                          : 'border-zinc-200 bg-white hover:border-zinc-300'
                      }`}
                    >
                      <b class="block text-zinc-900">{n}x {formatPrice(total / n)}</b>
                      <span class="text-zinc-500">{n === 1 ? 'à vista' : 'sem juros'}</span>
                    </button>
                  {/each}
                </div>
                {#if maxParc < 3}
                  <p class="mt-1 text-[11px] text-zinc-500">Parcela mínima de R$ 15,00.</p>
                {/if}
              </div>
              <label class="flex items-center gap-2 text-sm text-zinc-700 cursor-pointer">
                <input
                  type="checkbox"
                  bind:checked={payment.needsMachine}
                  onchange={() => onPaymentChange(payment)}
                  class="h-4 w-4 rounded accent-yellow-400"
                />
                <span>{isDelivery ? 'Entregador deve levar a maquininha' : 'Preciso da maquininha na mesa'}</span>
              </label>
            </div>
          {:else if id === 'debito'}
            <div class="rounded-xl bg-zinc-50 p-3 ring-1 ring-zinc-200">
              <label class="flex items-center gap-2 text-sm text-zinc-700 cursor-pointer">
                <input
                  type="checkbox"
                  bind:checked={payment.needsMachine}
                  onchange={() => onPaymentChange(payment)}
                  class="h-4 w-4 rounded accent-yellow-400"
                />
                <span>{isDelivery ? 'Entregador deve levar a maquininha' : 'Preciso da maquininha na mesa'}</span>
              </label>
            </div>
          {:else if id === 'dinheiro'}
            <div class="rounded-xl bg-zinc-50 p-3 ring-1 ring-zinc-200">
              <label class="block">
                <span class="mb-1 block text-sm text-zinc-700">
                  Troco para quanto? <span class="text-zinc-400">(opcional)</span>
                </span>
                <input
                  type="text"
                  inputmode="decimal"
                  bind:value={payment.changeFor}
                  oninput={() => onPaymentChange(payment)}
                  placeholder={`Ex: ${Math.ceil((total + 1) / 50) * 50}`}
                  class="w-full rounded-xl border border-zinc-200 bg-white px-3.5 py-2.5 text-sm text-zinc-900 focus:border-yellow-400 focus:outline-none"
                />
              </label>
            </div>
          {/if}
        </div>
      {/if}
    </div>
  {/each}
</div>
