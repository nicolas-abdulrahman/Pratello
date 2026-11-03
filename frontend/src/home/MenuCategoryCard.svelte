<script>
  // Define the props we expect from the parent
  let { title, description, image, icon } = $props();
</script>

<div class="bg-white rounded-3xl p-6 shadow-[0_8px_30px_rgb(0,0,0,0.04)] hover:shadow-[0_8px_30px_rgb(240,90,31,0.1)] transition-all duration-300 group border border-orange-50 relative overflow-hidden flex flex-col h-full">
  
  <!-- Image positioned right -->
  <div class="absolute -right-4 top-10 w-32 h-32 md:w-40 md:h-40 opacity-90 group-hover:scale-110 transition-transform duration-500 pointer-events-none">
     <img 
       src={image} 
       alt={title} 
       class="w-full h-full object-cover rounded-full shadow-2xl" 
       style="-webkit-mask-image: radial-gradient(circle, black 60%, transparent 70%);" 
     />
  </div>

  <div class="relative z-10 flex flex-col h-full">
    <!-- Icon -->
    <div class="mb-4 bg-orange-50 w-12 h-12 rounded-xl flex items-center justify-center">
      {@html icon}
    </div>
    
    <!-- Text Content -->
    <h3 class="text-xl font-bold text-gray-900 mb-2">{title}</h3>
    <p class="text-sm text-gray-500 mb-8 max-w-[70%]">{description}</p>
    
    <!-- Button -->
    <div class="mt-auto">
      <button class="text-[#f05a1f] border border-[#f05a1f]/30 px-4 py-2 rounded-full text-xs font-bold flex items-center gap-2 hover:bg-[#f05a1f] hover:text-white transition-colors">
        Ver opções 
        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="m9 18 6-6-6-6"/>
        </svg>
      </button>
    </div>
  </div>
</div>
