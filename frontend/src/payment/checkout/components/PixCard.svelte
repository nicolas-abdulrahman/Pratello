<script lang="ts">
  import { Check, Copy, QrCode } from 'lucide-svelte';

  interface Props {
    pixKey?: string;
    pixMerchant?: string;
    total: number;
    paid?: boolean;
    formatPrice?: (value: number) => string;
    onTogglePaid?: (paid: boolean) => void;
  }

  let {
    pixKey = '12.345.678/0001-90',
    pixMerchant = 'PRATELLO RESTAURANTE LTDA',
    total,
    paid = $bindable(false),
    formatPrice = (v: number) =>
      new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(v),
    onTogglePaid = () => {}
  }: Props = $props();

  let copied = $state(false);

  async function copyKey() {
    try {
      await navigator.clipboard.writeText(pixKey);
      copied = true;
      setTimeout(() => (copied = false), 2000);
    } catch {
      // clipboard fallback
    }
  }

  function handleCheck(e: Event) {
    const target = e.target as HTMLInputElement;
    paid = target.checked;
    onTogglePaid(paid);
  }
</script>

<div class="rounded-xl bg-zinc-50 p-3 ring-1 ring-zinc-200">
  <div class="flex gap-3">
    <!-- QR Code visual container -->
    <div class="relative flex h-24 w-24 shrink-0 flex-col items-center justify-center rounded-lg bg-white p-2 ring-1 ring-zinc-200">
      <QrCode size={56} class="text-zinc-800" />
      <span class="mt-0.5 text-[9px] font-bold uppercase tracking-wider text-zinc-400">Pix QR</span>
    </div>

    <div class="min-w-0 flex-1 text-xs">
      <p class="font-semibold text-zinc-800">Chave Pix ({pixMerchant})</p>
      <div class="mt-1 flex items-center gap-1">
        <code class="truncate rounded bg-white px-1.5 py-1 text-zinc-700 ring-1 ring-zinc-200 font-mono text-[11px]">{pixKey}</code>
        <button
          type="button"
          onclick={copyKey}
          class="rounded p-1 text-zinc-500 hover:bg-zinc-200 transition-colors"
          aria-label="Copiar chave Pix"
        >
          {#if copied}
            <Check size={14} class="text-green-600" />
          {:else}
            <Copy size={14} />
          {/if}
        </button>
      </div>
      <p class="mt-1 text-zinc-500">
        Valor: <b class="text-zinc-800">{formatPrice(total)}</b>
      </p>
      <label class="mt-2 flex items-center gap-2 text-sm text-zinc-700 cursor-pointer">
        <input
          type="checkbox"
          checked={paid}
          onchange={handleCheck}
          class="h-4 w-4 rounded border-zinc-300 accent-yellow-400"
        />
        <span>Já fiz o pagamento (simulação)</span>
      </label>
    </div>
  </div>
</div>
