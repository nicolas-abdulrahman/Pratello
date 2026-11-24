<script lang="ts">
  import type { HTMLInputAttributes } from 'svelte/elements';

  interface Props extends HTMLInputAttributes {
    value?: any;
    className?: string;
  }

  let { value = $bindable(''), className = '', ...restProps }: Props = $props();
</script>

<input
  bind:value
  class={`w-full rounded-xl border border-zinc-200 bg-white px-3.5 py-2.5 text-sm text-zinc-900 placeholder:text-zinc-400 focus:border-yellow-400 focus:outline-none focus:ring-2 focus:ring-yellow-400/40 ${className}`}
  {...restProps}
/>
