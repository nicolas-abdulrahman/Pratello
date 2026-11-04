<script>
  let { 
    label, 
    isRemoved = $bindable(false) 
  } = $props();

  function toggle() {
    isRemoved = !isRemoved;
  }
</script>

<button
  type="button"
  onclick={toggle}
  class="px-3 py-1.5 rounded-full text-xs font-semibold border transition-all flex items-center gap-1.5
  {isRemoved 
    ? 'bg-gray-100 text-gray-400 border-gray-200 line-through decoration-gray-400' 
    : 'bg-emerald-50 text-emerald-700 border-emerald-200 hover:bg-emerald-100'}"
>
  {label}
  {#if isRemoved}
    <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><path d="M18 6 6 18"/><path d="m6 6 12 12"/></svg>
  {/if}
</button>
