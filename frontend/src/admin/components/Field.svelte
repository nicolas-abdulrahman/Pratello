<script lang="ts">
  import type { Snippet } from 'svelte';

  interface Props {
    label: string;
    hint?: string;
    className?: string;
    children?: Snippet;
  }

  let { label, hint, className = '', children }: Props = $props();
</script>

<label class={`block ${className}`}>
  <span class="mb-1.5 block text-sm font-medium text-zinc-700">{label}</span>
  {#if children}
    {@render children()}
  {:else}
    <slot />
  {/if}
  {#if hint}
    <span class="mt-1 block text-xs text-zinc-500">{hint}</span>
  {/if}
</label>
