<script lang="ts">
  import type { Snippet } from 'svelte';
  import type { HTMLButtonAttributes } from 'svelte/elements';

  type Variant = 'primary' | 'dark' | 'ghost' | 'outline' | 'danger';
  type Size = 'sm' | 'md' | 'lg';

  interface Props extends HTMLButtonAttributes {
    variant?: Variant;
    size?: Size;
    className?: string;
    children?: Snippet;
  }

  let {
    variant = 'primary',
    size = 'md',
    className = '',
    children,
    type = 'button',
    ...restProps
  }: Props = $props();

  const base =
    'inline-flex items-center justify-center gap-2 rounded-xl font-semibold transition-colors disabled:cursor-not-allowed disabled:opacity-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-yellow-400 focus-visible:ring-offset-2';

  const variants: Record<Variant, string> = {
    primary: 'bg-yellow-400 text-zinc-900 hover:bg-yellow-500',
    dark: 'bg-zinc-900 text-white hover:bg-zinc-800',
    ghost: 'bg-transparent text-zinc-600 hover:bg-zinc-100',
    outline: 'border border-zinc-200 bg-white text-zinc-700 hover:bg-zinc-50',
    danger: 'bg-red-50 text-red-600 hover:bg-red-100',
  };

  const sizes: Record<Size, string> = {
    sm: 'px-3 py-1.5 text-sm',
    md: 'px-4 py-2.5 text-sm',
    lg: 'px-5 py-3 text-base',
  };
</script>

<button
  {type}
  class={`${base} ${variants[variant]} ${sizes[size]} ${className}`}
  {...restProps}
>
  {#if children}
    {@render children()}
  {:else}
    <slot />
  {/if}
</button>
