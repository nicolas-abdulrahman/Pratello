<script lang="ts">
  import type { Snippet } from 'svelte';

  let { title, description, children }: { title: string; description?: string; children: Snippet } = $props();
</script>

<section class="rounded-2xl border border-border bg-surface p-space-sm shadow-sm sm:p-space-md">
  <h2 class="text-xl font-bold">{title}</h2>
  {#if description}<p class="mt-space-3xs text-sm text-text-muted">{description}</p>{/if}
  <div class="mt-space-md">
    {@render children()}
  </div>
</section>
