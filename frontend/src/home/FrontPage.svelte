<script>
  import CombosSection from './CombosSection.svelte';
  import MelhoresPratosSection from './PromotionSection.svelte';
</script>

<div class="min-h-screen bg-[#fff9f5] font-sans selection:bg-[#f05a1f] selection:text-white">

  <!-- Hero Section (Made Smaller) -->
  <section class="relative pt-20 pb-10 px-8 overflow-hidden">
    <!-- Background Decorative Shapes -->
    <div class="absolute top-0 right-0 w-[600px] h-[600px] bg-[#f05a1f] rounded-full blur-[150px] opacity-10 -z-10 translate-x-1/3 -translate-y-1/4"></div>
    <div class="absolute bottom-0 left-0 w-[400px] h-[400px] bg-[#f05a1f] rounded-full blur-[100px] opacity-10 -z-10 -translate-x-1/2 translate-y-1/4"></div>

    <div class="max-w-7xl mx-auto flex flex-col lg:flex-row items-center justify-between gap-8">
      
      <!-- Text Content -->
      <div class="lg:w-1/2 z-10 text-center lg:text-left">
        <span class="text-[#f05a1f] font-bold text-lg tracking-wide uppercase">Bem-vindo ao</span>
        <!-- Reduced text size slightly for a tighter fit -->
        <h1 class="text-6xl md:text-8xl font-black text-gray-900 mt-2 mb-6 tracking-tighter relative">
          Pratello
          <!-- Decorative underline/swoosh -->
          <svg class="absolute -bottom-4 left-0 w-full h-8 text-[#f05a1f] opacity-80" viewBox="0 0 300 30" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M5 25C50 10 150 5 295 20" stroke="currentColor" stroke-width="8" stroke-linecap="round"/>
          </svg>
        </h1>
        <p class="text-lg md:text-xl text-gray-600 mb-8 max-w-lg mx-auto lg:mx-0">
          Sabor e qualidade em cada detalhe!
        </p>
        
        <!-- New Dual Action Buttons -->
        <div class="flex flex-col sm:flex-row items-center justify-center lg:justify-start gap-4">
          
          <!-- Order in Restaurant -->
          <button class="w-full sm:w-auto bg-[#f05a1f] text-white px-6 py-3 rounded-full font-bold text-base flex items-center justify-center gap-2 hover:bg-[#d94e17] transition-all hover:scale-105 shadow-xl shadow-orange-500/30">
            <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 2v7c0 1.1.9 2 2 2h4a2 2 0 0 0 2-2V2"/><path d="M7 2v20"/><path d="M21 15V2v0a5 5 0 0 0-5 5v6c0 1.1.9 2 2 2h3Zm0 0v7"/></svg>
            Pedir no Restaurante
          </button>

          <!-- Order Delivery -->
          <button class="w-full sm:w-auto bg-white text-[#f05a1f] border-2 border-[#f05a1f] px-6 py-3 rounded-full font-bold text-base flex items-center justify-center gap-2 hover:bg-orange-50 transition-all hover:scale-105 shadow-lg">
            <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M5 18H3c-.6 0-1-.4-1-1V7c0-.6.4-1 1-1h10c.6 0 1 .4 1 1v11"/><path d="M14 9h4l4 4v4c0 .6-.4 1-1 1h-2"/><circle cx="7" cy="18" r="2"/><circle cx="17" cy="18" r="2"/><path d="M9 18h6"/></svg>
            Delivery
          </button>

        </div>
      </div>

      <!-- Hero Image (Made slightly smaller) -->
      <div class="lg:w-1/2 relative flex justify-center">
        <!-- Floating text element -->
        <div class="absolute top-0 right-0 md:-right-8 z-20 transform rotate-6 bg-white px-4 py-2 rounded-xl shadow-lg border border-orange-100">
           <p class="text-[#f05a1f] font-bold italic font-serif text-base leading-tight">Peça agora<br/>e aproveite!</p>
        </div>

        <!-- Main Burger Image with organic masking -->
        <div class="relative w-[300px] h-[300px] md:w-[450px] md:h-[450px] rounded-full overflow-hidden shadow-2xl border-8 border-white z-10">
           <img 
            src="https://images.unsplash.com/photo-1568901346375-23c9450c58cd?ixlib=rb-4.0.3&auto=format&fit=crop&w=1000&q=80" 
            alt="Delicious Burger" 
            class="w-full h-full object-cover scale-110"
          />
        </div>
        
        <!-- Decorative dashes behind burger -->
        <div class="absolute -top-4 -left-4 text-[#f05a1f]">
          <svg width="60" height="60" viewBox="0 0 100 100" fill="none" stroke="currentColor" stroke-width="12" stroke-linecap="round">
            <line x1="20" y1="20" x2="40" y2="10"/>
            <line x1="10" y1="40" x2="30" y2="30"/>
            <line x1="20" y1="60" x2="40" y2="50"/>
          </svg>
        </div>
      </div>

    </div>
  </section>

  <MelhoresPratosSection />

  <!-- ========================================== -->
  <!-- PROMOTIONAL BANNER                         -->
  <!-- ========================================== -->
  <section class="py-8 px-4 md:px-8 w-full max-w-7xl mx-auto">
    <div class="relative rounded-3xl overflow-hidden shadow-2xl group">
      <!-- Background Image -->
      <img 
        src="https://images.unsplash.com/photo-1550547660-d9450f859349?ixlib=rb-4.0.3&auto=format&fit=crop&w=1600&q=80" 
        alt="Promo Banner" 
        class="w-full h-[300px] md:h-[400px] object-cover group-hover:scale-105 transition-transform duration-700"
      />
      <!-- Dark Gradient Overlay for text readability -->
      <div class="absolute inset-0 bg-gradient-to-r from-black/80 via-black/50 to-transparent"></div>
      
      <!-- Banner Content -->
      <div class="absolute inset-0 flex flex-col justify-center p-8 md:p-16 max-w-2xl">
        <span class="text-[#ff5722] font-bold text-lg md:text-xl uppercase tracking-wider mb-2">Oferta Especial</span>
        <h2 class="text-3xl md:text-5xl font-black text-white leading-tight mb-4">
          Sexta do Hambúrguer: <br/>
          <span class="text-[#ff5722]">20% OFF</span> em qualquer combo!
        </h2>
        <p class="text-gray-200 text-sm md:text-base mb-8 max-w-md">
          Válido todas as sextas-feiras no delivery. Aproveite a combinação perfeita de sabor e economia.
        </p>
        <button class="bg-[#ff5722] hover:bg-[#e64a19] text-white font-bold py-3 px-8 rounded-full w-max transition-colors shadow-lg shadow-orange-500/30 flex items-center gap-2">
          Aproveitar Oferta
          <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m9 18 6-6-6-6"/></svg>
        </button>
      </div>
    </div>
  </section>

  <CombosSection />

</div>
