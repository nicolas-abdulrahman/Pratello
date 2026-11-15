<script lang="ts">
  let {
    label,
    isRemoved = false,
    onToggle,
  }: {
    label: string;
    isRemoved?: boolean;
    onToggle?: () => void;
  } = $props();
</script>

<button
  type="button"
  onclick={() => onToggle?.()}
  class="ingredient-chip"
  class:is-removed={isRemoved}
  aria-pressed={isRemoved}
  aria-label={isRemoved ? `Restaurar ${label}` : `Remover ${label}`}
>
  <span class="truncate">{label}</span>
  {#if isRemoved}
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="12"
      height="12"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      stroke-width="3"
      stroke-linecap="round"
      stroke-linejoin="round"
    >
      <path d="M18 6 6 18" />
      <path d="m6 6 12 12" />
    </svg>
  {/if}
</button>

<style>
  @reference "../app.css";

  .ingredient-chip {
    @apply px-3 py-1.5 rounded-full text-xs font-semibold border transition-all flex items-center gap-1.5;
  }

  .ingredient-chip:not(.is-removed) {
    @apply bg-emerald-50 text-emerald-700 border-emerald-200 hover:bg-emerald-100;
  }

  .ingredient-chip.is-removed {
    @apply bg-surface-hover text-text-muted border-border line-through;
  }
</style>
