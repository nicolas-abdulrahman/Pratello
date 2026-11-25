<script lang="ts">
  import { Check, Ticket, Truck } from 'lucide-svelte';
  import type { Coupon } from '../checkout';

  interface Props {
    coupon: Coupon;
    active?: boolean;
    problem?: string | null;
    formatPrice?: (value: number) => string;
    onToggle?: (coupon: Coupon) => void;
  }

  let {
    coupon,
    active = false,
    problem = null,
    formatPrice = (v: number) =>
      new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(v),
    onToggle = () => {}
  }: Props = $props();

  let disabled = $derived(!!problem && !active);
  let isShip = $derived(coupon.kind === 'frete');
</script>

<button
  type="button"
  disabled={disabled}
  aria-pressed={active}
  onclick={() => onToggle(coupon)}
  class={`relative flex w-full overflow-hidden rounded-xl text-left ring-1 transition-shadow ${
    active ? 'ring-2 ring-yellow-500' : 'ring-zinc-200 hover:shadow-md'
  } ${disabled ? 'cursor-not-allowed opacity-60' : ''}`}
>
  <!-- Canhoto / Perforated ticket stub -->
  <div
    class={`relative flex w-24 shrink-0 flex-col items-center justify-center gap-1 border-r-2 border-dashed border-white px-2 py-3 ${
      disabled ? 'bg-zinc-300 text-zinc-600' : isShip ? 'bg-teal-500 text-white' : 'bg-yellow-400 text-zinc-900'
    }`}
  >
    {#if isShip}
      <Truck size={22} />
      <span class="text-[10px] font-extrabold uppercase tracking-wide">Frete</span>
    {:else}
      <Ticket size={22} />
      <span class="text-[10px] font-extrabold uppercase tracking-wide">Pratello</span>
    {/if}

    <!-- Semicircular cutout notches at perforation points -->
    <span class="absolute -right-2 -top-2 h-4 w-4 rounded-full bg-white shadow-inner" />
    <span class="absolute -bottom-2 -right-2 h-4 w-4 rounded-full bg-white shadow-inner" />
  </div>

  <!-- Ticket details -->
  <div class="flex min-w-0 flex-1 items-center gap-2 bg-white p-3">
    <div class="min-w-0 flex-1">
      <p class="font-bold text-zinc-900">{coupon.title}</p>
      <p class="text-xs text-zinc-600">{coupon.description}</p>
      <p class="mt-1 text-[11px] text-zinc-400">
        {#if coupon.minSubtotal > 0}
          Mín. {formatPrice(coupon.minSubtotal)} ·{' '}
        {/if}
        Válido até {coupon.validUntil} ·{' '}
        <span class="font-mono font-semibold text-zinc-500">{coupon.code}</span>
      </p>
      {#if problem}
        <p class="mt-1 text-[11px] font-semibold text-red-600">{problem}</p>
      {/if}
    </div>

    <!-- Active radio/checkbox checkmark indicator -->
    <span
      class={`grid h-5 w-5 shrink-0 place-items-center rounded-full border-2 ${
        active ? 'border-yellow-500 bg-yellow-400' : 'border-zinc-300'
      }`}
    >
      {#if active}
        <Check size={12} strokeWidth={3} class="text-zinc-900" />
      {/if}
    </span>
  </div>
</button>
