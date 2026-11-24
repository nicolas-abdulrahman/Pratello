<script lang="ts">
  import type { Snippet } from 'svelte';
  import type { HTMLSelectAttributes } from 'svelte/elements';

  interface Props extends HTMLSelectAttributes {
    value?: any;
    className?: string;
    children?: Snippet;
  }

  let { value = $bindable(''), className = '', children, ...restProps }: Props = $props();
</script>

<select
  bind:value
  class={`w-full rounded-xl border border-zinc-200 bg-white px-3.5 py-2.5 text-sm text-zinc-900 focus:border-yellow-400 focus:outline-none focus:ring-2 focus:ring-yellow-400/40 ${className}`}
  {...restProps}
>
  {#if children}
    {@render children()}
  {:else}
    <slot />
  {/if}
</select>
