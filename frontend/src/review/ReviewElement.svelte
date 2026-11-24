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
    class="inline-flex items-center  border rounded-full px-3 py-1.5 shadow-sm
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

    <!-- Review count -->
    {#if count}
      <div class="rounded-full px-2 py-[2px] flex items-center justify-center border
        {theme === 'dark' ? 'bg-black/20 border-white/10' : 'bg-white border-gray-200'}"
      >
        <span class="inline-flex items-center gap-1 text-[10px] leading-none
          {theme === 'dark' ? 'text-white/90' : 'text-gray-500'} font-medium whitespace-nowrap">

          <svg
            width="12"
            height="12"
            viewBox="0 0 24 24"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
            class="shrink-0"
          >
            <g id="User / Users_Group">
              <path
                id="Vector"
                d="M17 20C17 18.3431 14.7614 17 12 17C9.23858 17 7 18.3431 7 20M21 17.0004C21 15.7702 19.7659 14.7129 18 14.25M3 17.0004C3 15.7702 4.2341 14.7129 6 14.25M18 10.2361C18.6137 9.68679 19 8.8885 19 8C19 6.34315 17.6569 5 16 5C15.2316 5 14.5308 5.28885 14 5.76389M6 10.2361C5.38625 9.68679 5 8.8885 5 8C5 6.34315 6.34315 5 8 5C8.76835 5 9.46924 5.28885 10 5.76389M12 14C10.3431 14 9 12.6569 9 11C9 9.34315 10.3431 8 12 8C13.6569 8 15 9.34315 15 11C15 12.6569 13.6569 14 12 14Z"
                stroke={theme === 'dark' ? '#ffffff' : '#000000'}
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
            </g>
          </svg>
          {count}
        </span>
      </div>
    {/if}
  </div>
{/if}
