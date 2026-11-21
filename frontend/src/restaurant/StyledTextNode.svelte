<script>
  import { resolveColor } from './display';
  let { content, isDark } = $props();
  let chars = $derived(Array.from(content.text));
  let gradient = $derived(content.text_white ? null : content.gradient);
</script>

{#if gradient?.type === 'rough'}
  <span style:color={resolveColor(gradient.first_color, isDark)}>{chars.slice(0, gradient.start_char).join('')}</span><span style:color={resolveColor(gradient.second_color, isDark)}>{chars.slice(gradient.start_char).join('')}</span>
{:else if gradient?.type === 'smooth'}
  <span class="gradient" style:background-image={`linear-gradient(90deg, ${resolveColor(gradient.first_color, isDark)} ${gradient.start_char / Math.max(1, chars.length) * 100}%, ${resolveColor(gradient.second_color, isDark)} 100%)`}>{content.text}</span>
{:else}
  <span style:color={content.text_white ? '#ffffff' : resolveColor(content.color ?? 'text', isDark)}>{content.text}</span>
{/if}

<style>
  .gradient { background-clip: text; color: transparent; }
</style>
