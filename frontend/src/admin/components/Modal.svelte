<script lang="ts">
  import { X } from 'lucide-svelte';
  import type { Snippet } from 'svelte';

  interface Props {
    open?: boolean;
    title: string;
    onClose?: () => void;
    children?: Snippet;
  }

  let { open = false, title, onClose = () => {}, children }: Props = $props();

  $effect(() => {
    if (!open) return;
    const onKey = (e: KeyboardEvent) => e.key === 'Escape' && onClose();
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  });
</script>

{#if open}
  <div
    class="fixed inset-0 z-50 flex items-end justify-center bg-zinc-900/60 p-0 sm:items-center sm:p-4"
    role="presentation"
    onclick={onClose}
  >
    <div
      role="dialog"
      aria-modal="true"
      aria-label={title}
      class="max-h-[92vh] w-full overflow-y-auto rounded-t-2xl bg-white shadow-xl sm:max-w-lg sm:rounded-2xl"
      onclick={(e) => e.stopPropagation()}
    >
      <div class="sticky top-0 flex items-center justify-between border-b border-zinc-200 bg-white px-5 py-4">
        <h2 class="text-lg font-bold text-zinc-900">{title}</h2>
        <button
          type="button"
          onclick={onClose}
          class="rounded-lg p-1.5 text-zinc-500 hover:bg-zinc-100"
          aria-label="Fechar"
        >
          <X size={20} />
        </button>
      </div>
      <div class="p-5">
        {#if children}
          {@render children()}
        {:else}
          <slot />
        {/if}
      </div>
    </div>
  </div>
{/if}
