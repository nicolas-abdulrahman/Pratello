<script>
  let { 
    rating = null, 
    count = null,
    theme = 'light'
  } = $props();

  // Convert the 1-5 rating to a percentage
  let percentage = $derived(rating ? Math.round((rating / 5) * 100) : 0);
</script>

{#if rating && rating > 0}
  <!-- 
    Light mode: soft gray background + border
    Dark mode: glassmorphism background + border
  -->
  <div 
    class="flex items-center gap-2 border rounded-full px-3 py-1.5 shadow-sm
    {theme === 'dark' 
      ? 'bg-black/40 backdrop-blur-md border-white/20' 
      : 'bg-gray-50 border-gray-200'}"
  >
    <!-- Percentage and Happy Face -->
    <div class="flex items-center gap-1.5">
      <span class="text-[13px] font-bold {theme === 'dark' ? 'text-white' : 'text-gray-800'}">
        {percentage}%
      </span>
      
      <!-- Happy Face Icon -->
      <svg 
        xmlns="http://www.w3.org/2000/svg" 
        width="16" 
        height="16" 
        viewBox="0 0 24 24" 
        fill="none" 
        stroke={theme === 'dark' ? '#FFD54F' : '#D97706'} 
        stroke-width="2.5" 
        stroke-linecap="round" 
        stroke-linejoin="round"
        class="shrink-0"
      >
        <circle cx="12" cy="12" r="10"></circle>
        <path d="M8 14s1.5 2 4 2 4-2 4-2"></path>
        <line x1="9" x2="9.01" y1="9" y2="9"></line>
        <line x1="15" x2="15.01" y1="9" y2="9"></line>
      </svg>
    </div>

    <!-- Order Count Pill -->
    {#if count}
      <div class="rounded-full px-2 py-[2px] flex items-center justify-center border
        {theme === 'dark' ? 'bg-black/20 border-white/10' : 'bg-white border-gray-200'}"
      >
        <span class="text-[10px] {theme === 'dark' ? 'text-white/90' : 'text-gray-500'} font-medium whitespace-nowrap">
          {count} orders
        </span>
      </div>
    {/if}
  </div>
{/if}
