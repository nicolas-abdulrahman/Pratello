<script lang="ts">
  import { ImageOff } from 'lucide-svelte';

  interface Props {
    src?: string;
    alt: string;
    className?: string;
  }

  let { src = '', alt, className = '' }: Props = $props();

  let failed = $state(false);

  $effect(() => {
    // Reset failed state whenever src changes
    void src;
    failed = false;
  });
</script>

{#if !src || failed}
  <div
    class={`grid place-items-center bg-zinc-200 text-zinc-400 ${className}`}
    aria-label={alt}
  >
    <ImageOff size={24} />
  </div>
{:else}
  <img
    {src}
    {alt}
    loading="lazy"
    onerror={() => (failed = true)}
    class={`object-cover ${className}`}
  />
{/if}
