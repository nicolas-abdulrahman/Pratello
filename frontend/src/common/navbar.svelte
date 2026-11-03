<script>
  // Svelte 5 Props
  let {
    cartItemCount = 0,
    onOpenMenu = () => {},
    onGoHome = () => {},
    onGoRestaurantes = () => {},
    onGoLogin = () => {},
    onOpenCart = () => {}
  } = $props();

  // State to control the Left Menu Sidebar
  let isLeftMenuOpen = $state(false);

  function toggleLeftMenu() {
    isLeftMenuOpen = !isLeftMenuOpen;
  }
</script>

<nav class="w-full bg-white shadow-sm relative z-10">
  <div class="max-w-7xl mx-auto px-4 md:px-8 h-20 flex items-center justify-between">
    
    <!-- LEFT SIDE: Logo & Enter Button -->
    <div class="flex items-center gap-6">
      <!-- Logo -->

      <!-- Enter / Login Button -->
      <button 
        onclick={toggleLeftMenu}
        class="hidden md:flex items-center gap-2 text-gray-600 hover:text-[#ff5722] font-semibold transition-colors bg-gray-50 hover:bg-orange-50 px-3 py-1.5 rounded-full border border-gray-200"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M19 21v-2a4 4 0 0 0-4-4H9a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
        Entrar
      </button>
    </div>

    <!-- CENTER: Desktop Links (Cleaned up) -->
    <div class="hidden md:flex items-center gap-8 font-medium text-gray-600">
      <a href="/" onclick={(e) => { e.preventDefault(); onGoHome(); }} class="hover:text-[#ff5722] transition-colors">Início</a>
      
      <a href="/cardapio" onclick={(e) => { e.preventDefault(); onOpenMenu(); }} class="hover:text-[#ff5722] transition-colors cursor-pointer">Cardápio</a>
      
      <!-- New Restaurantes Link -->
      <a href="/restaurantes" onclick={(e) => { e.preventDefault(); onGoRestaurantes(); }} class="hover:text-[#ff5722] transition-colors cursor-pointer">Delivery</a>
    </div>

    <!-- RIGHT SIDE: Mobile Menu Toggle -->
    <div class="flex items-center gap-4">
      <button class="md:hidden p-2 text-gray-600">
        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="4" x2="20" y1="12" y2="12"/><line x1="4" x2="20" y1="6" y2="6"/><line x1="4" x2="20" y1="18" y2="18"/></svg>
      </button>
    </div>

  </div>
</nav>

<!-- ========================================== -->
<!-- LEFT SIDEBAR MENU (Enter/Login Menu)       -->
<!-- ========================================== -->

<!-- Backdrop Overlay -->
{#if isLeftMenuOpen}
  <div 
    class="fixed inset-0 bg-black/40 backdrop-blur-sm z-40 transition-opacity"
    onclick={() => isLeftMenuOpen = false}
    role="button"
    tabindex="0"
    onkeydown={(e) => e.key === 'Escape' && (isLeftMenuOpen = false)}
    aria-label="Fechar menu"
  ></div>
{/if}

<!-- Sidebar Container -->
<div 
  class="fixed top-0 left-0 h-full w-full max-w-sm bg-white shadow-2xl z-50 transform transition-transform duration-300 ease-in-out flex flex-col
  {isLeftMenuOpen ? 'translate-x-0' : '-translate-x-full'}"
>
  <!-- Sidebar Header -->
  <div class="flex items-center justify-between p-6 border-b border-gray-100 bg-gray-50/50">
    <h2 class="text-2xl font-black text-gray-900">Acesso</h2>
    <button onclick={() => isLeftMenuOpen = false} class="p-2 text-gray-400 hover:text-gray-900 bg-gray-100 hover:bg-gray-200 rounded-full transition-colors">
      <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 6 6 18"/><path d="m6 6 12 12"/></svg>
    </button>
  </div>

  <!-- Sidebar Content -->
  <div class="flex-1 overflow-y-auto p-6 space-y-8">
    
    <!-- Login Options -->
    <div class="space-y-3">
      <h3 class="text-xs font-bold text-gray-400 uppercase tracking-wider mb-2">Entrar como</h3>
      
      <!-- Customer Login -->
      <button 
        onclick={() => { onGoLogin(); isLeftMenuOpen = false; }}
        class="w-full flex items-center justify-between p-4 bg-orange-50 hover:bg-orange-100 border border-orange-100 rounded-xl transition-colors group text-left"
      >
        <div class="flex items-center gap-3">
          <div class="p-2 bg-white rounded-full text-[#ff5722] shadow-sm">
            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M19 21v-2a4 4 0 0 0-4-4H9a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
          </div>
          <div>
            <p class="font-bold text-gray-900">Cliente</p>
            <p class="text-xs text-gray-500">Faça pedidos e veja seu histórico</p>
          </div>
        </div>
        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-gray-300 group-hover:text-[#ff5722] transition-colors"><path d="m9 18 6-6-6-6"/></svg>
      </button>

      <!-- Employee Login -->
      <button 
        onclick={() => { onGoLogin(); isLeftMenuOpen = false; }}
        class="w-full flex items-center justify-between p-4 bg-white hover:bg-gray-50 border border-gray-200 rounded-xl transition-colors group text-left"
      >
        <div class="flex items-center gap-3">
          <div class="p-2 bg-gray-100 rounded-full text-gray-600 group-hover:text-[#ff5722] transition-colors">
            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="20" height="14" x="2" y="7" rx="2" ry="2"/><path d="M16 21V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v16"/></svg>
          </div>
          <div>
            <p class="font-bold text-gray-900">Funcionário</p>
            <p class="text-xs text-gray-500">Acesse o painel de gestão</p>
          </div>
        </div>
        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-gray-300 group-hover:text-[#ff5722] transition-colors"><path d="m9 18 6-6-6-6"/></svg>
      </button>
    </div>

    <!-- Extra Links -->
    <div class="space-y-2 pt-4 border-t border-gray-100">
      <h3 class="text-xs font-bold text-gray-400 uppercase tracking-wider mb-3">Mais opções</h3>
      
      <a href="#history" class="flex items-center gap-3 p-3 text-gray-600 hover:text-[#ff5722] hover:bg-orange-50 rounded-lg transition-colors font-medium">
        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8"/><path d="M3 3v5h5"/><path d="M12 7v5l4 2"/></svg>
        Histórico de Pedidos
      </a>

      <a href="#contact" class="flex items-center gap-3 p-3 text-gray-600 hover:text-[#ff5722] hover:bg-orange-50 rounded-lg transition-colors font-medium">
        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.5 19.5 0 0 1-6-6 19.79 19.79 0 0 1-3.07-8.67A2 2 0 0 1 4.11 2h3a2 2 0 0 1 2 1.72 12.84 12.84 0 0 0 .7 2.81 2 2 0 0 1-.45 2.11L8.09 9.91a16 16 0 0 0 6 6l1.27-1.27a2 2 0 0 1 2.11-.45 12.84 12.84 0 0 0 2.81.7A2 2 0 0 1 22 16.92z"/></svg>
        Fale Conosco
      </a>
    </div>

  </div>
</div>
