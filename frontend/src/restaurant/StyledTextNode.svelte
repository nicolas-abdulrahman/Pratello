<script>
  import { resolveText } from '../lib/stores/restaurant_style';
  import SectionTitle from './section/SectionTitle.svelte';
  let { content, isDark } = $props();
  let title = $derived(resolveText(content, isDark, 'text'));
</script>

<SectionTitle {title} />
