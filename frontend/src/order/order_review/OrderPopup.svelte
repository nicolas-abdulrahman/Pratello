<script lang="ts">
  let { restaurantName, onContinue, onDismiss }: {
    restaurantName: string;
    onContinue: () => void;
    onDismiss: () => void;
  } = $props();
</script>

<div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
  <div role="dialog" aria-modal="true" aria-labelledby="order-popup-title" class="w-full max-w-md rounded-3xl bg-surface p-8 shadow-xl space-y-4">
    <h2 id="order-popup-title" class="text-2xl font-bold">Rate your experience</h2>
    <p>Seu pedido de {restaurantName} foi entregue. Conte como foi sua experiência!</p>
    <div class="flex gap-3">
      <button type="button" onclick={onContinue} class="rounded-xl bg-primary px-5 py-3 font-bold text-white">Continue</button>
      <button type="button" onclick={onDismiss} class="rounded-xl border border-border px-5 py-3">Depois</button>
    </div>
  </div>
</div>
