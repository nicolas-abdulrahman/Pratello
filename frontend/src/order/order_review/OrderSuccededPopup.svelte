<script lang="ts">
  let {
    orderCode = '',
    restaurantName = 'o restaurante',
    onConfirm = () => {},
  }: {
    orderCode?: string;
    restaurantName?: string;
    onConfirm?: () => void;
  } = $props();
</script>

<div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4 backdrop-blur-xs">
  <div role="dialog" aria-modal="true" aria-labelledby="pickup-title" class="w-full max-w-md rounded-3xl bg-surface p-8 shadow-2xl space-y-5 border border-border">
    <div class="text-center space-y-2">
      <div class="mx-auto w-16 h-16 rounded-full bg-emerald-500/10 text-emerald-600 flex items-center justify-center mb-3">
        <svg class="w-8 h-8" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
        </svg>
      </div>
      <h2 id="pickup-title" class="text-2xl font-bold text-text">Seu pedido está pronto!</h2>
      <p class="text-sm text-text-muted">Vá buscar seu pedido em <b class="text-text">{restaurantName}</b>.</p>
    </div>

    <div class="bg-surface-hover border border-border rounded-2xl py-4 px-4 text-center space-y-1">
      <span class="text-xs uppercase tracking-wider text-text-muted font-bold block">Código de retirada</span>
      <span class="text-3xl font-black text-primary tracking-widest font-mono">{orderCode}</span>
      <p class="text-[11px] text-text-muted mt-1">Apresente este código no balcão</p>
    </div>

    <button
      type="button"
      onclick={onConfirm}
      class="w-full rounded-xl bg-primary hover:bg-primary-hover px-5 py-3.5 font-bold text-white transition cursor-pointer shadow-lg shadow-primary/20 text-base"
    >
      OK
    </button>
  </div>
</div>
