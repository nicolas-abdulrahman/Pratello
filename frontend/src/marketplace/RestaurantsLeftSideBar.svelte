<script lang="ts">
  let { show = false, onClose }: { show?: boolean; onClose: () => void } = $props();
</script>

{#if show}
  <!-- Overlay -->
  <div
    class="fixed inset-0 bg-black/40 z-50 transition-opacity"
    role="presentation"
    onclick={onClose}
  ></div>

  <!-- Sidebar -->
  <aside
    class="fixed top-0 left-0 bottom-0 w-80 bg-[#FAF6F0] z-50 shadow-2xl p-6 flex flex-col gap-6 overflow-y-auto transition-transform duration-300 translate-x-0"
  >
    <!-- Brand + close -->
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-3">
        <div
          class="w-11 h-11 rounded-2xl bg-amber-500/10 flex items-center justify-center text-amber-600 shadow-inner"
        >
          <svg
            class="w-7 h-7"
            fill="none"
            stroke="currentColor"
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            viewBox="0 0 24 24"
          >
            <path
              d="M6 13.87A4 4 0 0 1 7.41 6a5.11 5.11 0 0 1 1.05-1.54 5 5 0 0 1 7.08 0A5.11 5.11 0 0 1 16.59 6 4 4 0 0 1 18 13.87V21H6Z"
            ></path>
            <line x1="6" x2="18" y1="17" y2="17"></line>
          </svg>
        </div>
        <span class="text-2xl font-black tracking-tight text-stone-900 font-serif">Pratello</span>
      </div>
      <button
        class="w-8 h-8 rounded-full flex items-center justify-center text-stone-600 hover:bg-stone-200 transition"
        onclick={onClose}
        aria-label="Fechar"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
          <path d="M6 18L18 6M6 6l12 12" stroke-linecap="round" stroke-linejoin="round" />
        </svg>
      </button>
    </div>

    <!-- User info -->
    <div class="flex items-center gap-3 pb-4 border-b border-stone-200">
      <div class="w-12 h-12 rounded-2xl bg-amber-500/15 flex items-center justify-center text-amber-600">
        <svg class="w-6 h-6" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
          <path d="M16 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path>
          <circle cx="9" cy="7" r="4"></circle>
        </svg>
      </div>
      <div>
        <p class="text-sm font-bold text-stone-900 leading-tight">Pratello Gourmet</p>
        <p class="text-[11px] text-stone-400">gregorio@pratello.com.br</p>
      </div>
    </div>

    <!-- Nav links -->
    <nav class="space-y-1.5 font-medium text-stone-600 text-sm">
      <a
        class="flex items-center gap-3.5 px-4 py-3 rounded-2xl bg-amber-200/60 text-stone-900 font-semibold transition shadow-xs"
        href="#"
      >
        <svg class="w-5 h-5 text-amber-700" fill="currentColor" viewBox="0 0 24 24">
          <path d="M10 20v-6h4v6h5v-8h3L12 3 2 12h3v8z"></path>
        </svg>
        Início
      </a>
      <a
        class="flex items-center gap-3.5 px-4 py-3 rounded-2xl hover:bg-[#F2EADC] text-stone-600 hover:text-stone-900 transition"
        href="#restaurantes-proximos"
      >
        <svg class="w-5 h-5 text-stone-500" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
          <circle cx="11" cy="11" r="8"></circle>
          <path d="m21 21-4.35-4.35" stroke-linecap="round"></path>
        </svg>
        Buscar
      </a>
      <a
        class="flex items-center gap-3.5 px-4 py-3 rounded-2xl hover:bg-[#F2EADC] text-stone-600 hover:text-stone-900 transition"
        href="#"
      >
        <svg class="w-5 h-5 text-stone-500" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
          <path
            d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z"
            stroke-linecap="round"
            stroke-linejoin="round"
          ></path>
        </svg>
        Favoritos
      </a>
      <a
        class="flex items-center gap-3.5 px-4 py-3 rounded-2xl hover:bg-[#F2EADC] text-stone-600 hover:text-stone-900 transition"
        href="#"
      >
        <svg class="w-5 h-5 text-stone-500" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
          <path
            d="M7 7h.01M7 3h5a1.99 1.99 0 011.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.99 1.99 0 013 12V5a2 2 0 012-2z"
            stroke-linecap="round"
            stroke-linejoin="round"
          ></path>
        </svg>
        Promoções
      </a>
      <a
        class="flex items-center gap-3.5 px-4 py-3 rounded-2xl hover:bg-[#F2EADC] text-stone-600 hover:text-stone-900 transition"
        href="#"
      >
        <svg class="w-5 h-5 text-stone-500" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
          <circle cx="12" cy="12" r="10"></circle>
          <polyline points="12 6 12 12 16 14"></polyline>
        </svg>
        Pedidos
      </a>
      <a
        class="flex items-center gap-3.5 px-4 py-3 rounded-2xl hover:bg-[#F2EADC] text-stone-600 hover:text-stone-900 transition"
        href="#"
      >
        <svg class="w-5 h-5 text-stone-500" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
          <circle cx="12" cy="12" r="10"></circle>
          <path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3" stroke-linecap="round"></path>
          <line x1="12" x2="12.01" y1="17" y2="17"></line>
        </svg>
        Ajuda
      </a>
    </nav>

    <!-- Logout -->
    <div class="mt-auto">
      <button
        class="w-full flex items-center justify-center gap-2 px-3 py-2.5 text-xs font-semibold text-red-600 rounded-xl border border-red-200 hover:bg-red-50 transition"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
          <path
            d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"
            stroke-linecap="round"
            stroke-linejoin="round"
          ></path>
        </svg>
        Fazer logout
      </button>
    </div>
  </aside>
{/if}
