<script>
  import ReviewElement from './ReviewElement.svelte';

  let { rating = null, count = null, theme = 'light' } = $props();
</script>

<ReviewElement {rating} {count} {theme} />
