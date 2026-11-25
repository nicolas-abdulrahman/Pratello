<script lang="ts">
  import { X } from 'lucide-svelte';
  import VoucherCard from './VoucherCard.svelte';
  import {
    COUPONS,
    findCoupon,
    couponProblem,
    type Coupon,
    type CouponContext
  } from '../checkout';

  interface Props {
    open?: boolean;
    selectedCodes?: string[];
    ctx: CouponContext;
    onClose?: () => void;
    onToggleCoupon?: (coupon: Coupon) => void;
  }

  let {
    open = false,
    selectedCodes = [],
    ctx,
    onClose = () => {},
    onToggleCoupon = () => {}
  }: Props = $props();

  let codeInput = $state('');
  let message = $state<{ ok: boolean; text: string } | null>(null);

  function applyCode() {
    const c = findCoupon(codeInput);
    if (!c) {
      message = { ok: false, text: 'Cupom não encontrado. Confira o código.' };
      return;
    }
    const problem = couponProblem(c, ctx);
    if (problem) {
      message = { ok: false, text: `${c.code}: ${problem}.` };
      return;
    }
    if (!selectedCodes.includes(c.code)) {
      onToggleCoupon(c);
    }
    message = { ok: true, text: `Cupom ${c.code} aplicado!` };
    codeInput = '';
  }

  let freteCoupons = $derived(COUPONS.filter((c) => c.kind === 'frete'));
  let descontoCoupons = $derived(COUPONS.filter((c) => c.kind === 'desconto'));
</script>

{#if open}
  <div
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4"
    role="presentation"
    onclick={onClose}
  >
    <div
      role="dialog"
      aria-modal="true"
      aria-label="Cupons Pratello"
      class="max-h-[90vh] w-full max-w-lg overflow-y-auto rounded-2xl bg-white p-5 shadow-xl"
      onclick={(e) => e.stopPropagation()}
    >
      <!-- Header -->
      <div class="mb-4 flex items-center justify-between">
        <h2 class="text-lg font-bold text-zinc-900">Cupons Pratello</h2>
        <button
          type="button"
          onclick={onClose}
          class="rounded-lg p-1 text-zinc-400 hover:bg-zinc-100"
          aria-label="Fechar"
        >
          <X size={20} />
        </button>
      </div>

      <div class="space-y-5">
        <!-- Form to apply code manually -->
        <form
          class="flex gap-2"
          onsubmit={(e) => {
            e.preventDefault();
            applyCode();
          }}
        >
          <input
            type="text"
            bind:value={codeInput}
            placeholder="Digite o código do cupom"
            class="w-full rounded-xl border border-zinc-200 bg-white px-3.5 py-2.5 text-sm uppercase text-zinc-900 placeholder:text-zinc-400 focus:border-yellow-400 focus:outline-none focus:ring-2 focus:ring-yellow-400/40"
          />
          <button
            type="submit"
            disabled={!codeInput.trim()}
            class="rounded-xl bg-zinc-900 px-4 py-2.5 text-sm font-semibold text-white hover:bg-zinc-800 disabled:opacity-50"
          >
            Aplicar
          </button>
        </form>

        {#if message}
          <p class={`-mt-3 text-sm font-medium ${message.ok ? 'text-green-600' : 'text-red-600'}`}>
            {message.text}
          </p>
        {/if}

        <!-- Cupons de frete -->
        <div>
          <p class="mb-2 text-xs font-semibold uppercase tracking-wide text-zinc-500">Cupons de frete</p>
          <p class="-mt-1 mb-2 text-[11px] text-zinc-400">Escolha 1 cupom de frete</p>
          <div class="space-y-2.5">
            {#each freteCoupons as c (c.code)}
              <VoucherCard
                coupon={c}
                active={selectedCodes.includes(c.code)}
                problem={couponProblem(c, ctx)}
                onToggle={onToggleCoupon}
              />
            {/each}
          </div>
        </div>

        <!-- Cupons de desconto -->
        <div>
          <p class="mb-2 text-xs font-semibold uppercase tracking-wide text-zinc-500">Cupons de desconto</p>
          <p class="-mt-1 mb-2 text-[11px] text-zinc-400">Escolha 1 cupom de desconto</p>
          <div class="space-y-2.5">
            {#each descontoCoupons as c (c.code)}
              <VoucherCard
                coupon={c}
                active={selectedCodes.includes(c.code)}
                problem={couponProblem(c, ctx)}
                onToggle={onToggleCoupon}
              />
            {/each}
          </div>
        </div>

        <button
          type="button"
          onclick={onClose}
          class="w-full rounded-xl bg-yellow-400 px-5 py-3 font-semibold text-zinc-900 hover:bg-yellow-500 transition-colors"
        >
          Confirmar
        </button>
      </div>
    </div>
  </div>
{/if}
