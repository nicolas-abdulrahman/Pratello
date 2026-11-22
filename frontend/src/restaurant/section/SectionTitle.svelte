<script lang="ts">
  import type { ResolvedText } from '../../lib/stores/restaurant_style';

  let { title }: { title: ResolvedText } = $props();
</script>

{#if title.mode === 'rough'}
  <span style:color={title.firstColor}>{title.first}</span><span style:color={title.secondColor}>{title.second}</span>
{:else if title.mode === 'smooth'}
  <span class="bg-clip-text text-transparent" style:background-image={title.gradientCss}>{title.text}</span>
{:else}
  <span style:color={title.color}>{title.text}</span>
{/if}
