<script>
  import ReviewProp from '../review/ReviewProp.svelte';
  import { add_item} from '../lib/stores/Cart';

  let { 
    isOpen = false, 
    onClose = () => {},
    item = {}
  } = $props();

  const mockReviews = [
    { name: "Gregory the III", text: "Best dish I ever eaten." },
    { name: "Mr Puzzles", text: "It shatters your mind." }
  ];

  let isAdding = $state(false);

  // Lock body scroll
  $effect(() => {
    if (isOpen) {
      document.body.style.overflow = 'hidden';
    } else {
      document.body.style.overflow = '';
    }
    return () => {
      document.body.style.overflow = '';
    };
  });

  function teleport(node) {
    document.body.appendChild(node);
    return {
      destroy() {
        node.remove();
      }
    };
  }

  function parsePrice(value) {
    if (typeof value === 'number') return value;
    if (!value) return 0;
    return Number(
      String(value)
        .replace(/[^\d,.-]/g, '')
        .replace(/\./g, '')
        .replace(',', '.')
    ) || 0;
  }

  let discountPercent = $derived.by(() => {
    if (!item.oldPrice || !item.newPrice) return null;
    const oldP = parsePrice(item.oldPrice);
    const newP = parsePrice(item.newPrice);
    if (oldP <= newP || oldP === 0) return null;
    return Math.round(((oldP - newP) / oldP) * 100);
  });

  function handleAdd() {
    if (isAdding) return;
    isAdding = true;

    add_item({
      name: item.title,
      icon: item.image,
      price: parsePrice(item.newPrice),
      combo: item.combo,
    });

    setTimeout(() => {
      isAdding = false;
    }, 800);
  }
</script>

<div {@attach teleport}>
  {#if isOpen}
    <!-- Backdrop -->
    <div 
      class="fixed inset-0 bg-black/60 backdrop-blur-sm z-[99998] transition-opacity duration-300"
      onclick={onClose}
      role="button"
      tabindex="0"
      onkeydown={(e) => e.key === 'Escape' && onClose()}
      aria-label="Fechar detalhes do prato"
    ></div>

    <!-- Modal Wrapper -->
    <div class="fixed inset-0 z-[99999] flex items-center justify-center p-4 md:p-8 pointer-events-none">
      
      <!-- Modal Body -->
      <div 
        class="bg-white w-full h-full md:h-auto md:max-h-[85vh] md:rounded-3xl shadow-[0_20px_60px_-15px_rgba(0,0,0,0.5)] border border-white/20 overflow-hidden pointer-events-auto flex flex-col relative animate-slide-up max-w-4xl"
        onclick={(e) => e.stopPropagation()}
        onkeydown={(e) => e.stopPropagation()}
        role="dialog"
        aria-modal="true"
        aria-labelledby="modal-title"
        tabindex="-1" 
      >
        
        <!-- TOP SECTION: Scrollable Content Area -->
        <div class="flex-1 overflow-y-auto relative bg-gray-50/30">
          
          <!-- Back Button -->
          <button 
            onclick={onClose}
            class="absolute top-4 left-4 z-20 bg-white/80 backdrop-blur-md text-gray-800 p-2.5 rounded-full shadow-lg hover:bg-white hover:scale-105 transition-all"
            aria-label="Voltar"
          >
            <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="m15 18-6-6 6-6"/></svg>
          </button>

          <!-- Hero Image -->
          <div class="w-full h-[280px] md:h-[380px] relative shrink-0">
            <img src={item.image} alt={item.title} class="w-full h-full object-cover" />
            <div class="absolute inset-0 bg-gradient-to-t from-black/90 via-black/40 to-transparent"></div>
            
            <!-- Header info over image -->
            <div class="absolute bottom-0 left-0 p-6 md:p-10 w-full flex flex-col items-start gap-3">
              <h1 id="modal-title" class="text-3xl md:text-5xl font-black text-white">{item.title}</h1>
              
              <!-- 👇 ReviewProp is now rendered directly (it handles its own pill bg) -->
              {#if item.rating}
                <ReviewProp rating={item.rating} count={item.count} theme="dark" />
              {/if}
            </div>
          </div>

          <!-- Content Area -->
          <div class="p-6 md:p-10 flex flex-col md:flex-row gap-8 pb-10">
            
            <!-- Left: Description & Ingredients -->
            <div class="flex-1 space-y-6">
              <div>
                <h2 class="text-lg font-bold text-gray-900 mb-2">Sobre o prato</h2>
                <p class="text-gray-600 leading-relaxed">{item.description}</p>
              </div>
              
              <div>
                <h2 class="text-lg font-bold text-gray-900 mb-2">Ingredientes</h2>
                <p class="text-gray-500 text-sm leading-relaxed">
                  {item.ingredients || "Ingredientes selecionados com cuidado para garantir o melhor sabor. Contém glúten e lactose. Pode conter traços de oleaginosas. Preparado em ambiente que manipula diversos tipos de alimentos."}
                </p>
              </div>
            </div>

            <!-- Right: Reviews -->
            <div class="flex-1">
              <h2 class="text-lg font-bold text-gray-900 mb-4 flex items-center gap-2">
                Avaliações
                {#if item.rating}
                  <span class="bg-orange-100 text-[#ff5722] text-xs px-2 py-0.5 rounded-full">{mockReviews.length}</span>
                {/if}
              </h2>

              {#if item.rating}
                <div class="space-y-4">
                  {#each mockReviews as review}
                    <div class="bg-white p-4 rounded-2xl border border-gray-100 shadow-sm">
                      <div class="flex items-center gap-2 mb-1">
                        <div class="w-8 h-8 rounded-full bg-gray-100 flex items-center justify-center text-gray-500">
                          <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M19 21v-2a4 4 0 0 0-4-4H9a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
                        </div>
                        <span class="font-bold text-sm text-gray-800">{review.name}</span>
                      </div>
                      <p class="text-gray-600 text-sm">"{review.text}"</p>
                    </div>
                  {/each}
                </div>
              {:else}
                <div class="bg-white p-8 rounded-2xl border border-gray-100 text-center shadow-sm">
                  <div class="text-gray-300 mb-2 flex justify-center">
                    <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>
                  </div>
                  <p class="text-gray-500 text-sm font-medium">Nenhuma avaliação ainda.</p>
                </div>
              {/if}
            </div>
          </div>
        </div> <!-- End Scrollable Container -->

        <!-- BOTTOM SECTION: Fixed Footer -->
        <div class="shrink-0 bg-white border-t border-gray-100 p-4 md:p-6 z-20 w-full">
          <div class="max-w-4xl mx-auto flex items-center justify-between gap-4">
            
            <!-- Price Section -->
            <div class="flex flex-col flex-shrink-0 w-[35%]">
              {#if item.oldPrice}
                <div class="flex items-center gap-2 mb-1 flex-wrap">
                  <span class="text-xs text-gray-400 line-through font-medium">{item.oldPrice}</span>
                  {#if discountPercent}
                    <span class="bg-green-100 text-green-700 text-[10px] font-bold px-1.5 py-0.5 rounded uppercase tracking-wider">
                      {discountPercent}% OFF
                    </span>
                  {/if}
                </div>
              {/if}
              <span class="text-3xl md:text-4xl font-black text-[#ff5722] leading-none">{item.newPrice}</span>
            </div>
            
            <!-- Add Button -->
            <button 
              onclick={handleAdd}
              disabled={isAdding}
              class="flex-1 bg-[#ff5722] hover:bg-[#e64a19] text-white font-bold py-3.5 px-6 rounded-full flex items-center justify-center gap-2 transition-all shadow-lg hover:shadow-orange-500/30 text-sm md:text-base
              {isAdding ? 'bg-emerald-500 hover:bg-emerald-500 shadow-emerald-500/30 scale-[0.98]' : ''}"
            >
              {#if isAdding}
                <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg>
                Adicionado!
              {:else}
                <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="8" cy="21" r="1"/><circle cx="19" cy="21" r="1"/><path d="M2.05 2.05h2l2.66 12.42a2 2 0 0 0 2 1.58h9.78a2 2 0 0 0 1.95-1.57l1.65-7.43H5.12"/></svg>
                Adicionar ao Carrinho
              {/if}
            </button>

          </div>
        </div>

      </div>
    </div>
  {/if}
</div>

<style>
  .animate-slide-up {
    animation: slideUp 0.4s cubic-bezier(0.16, 1, 0.3, 1) forwards;
  }
  
  @keyframes slideUp {
    0% { opacity: 0; transform: translateY(40px) scale(0.98); }
    100% { opacity: 1; transform: translateY(0) scale(1); }
  }
</style>
