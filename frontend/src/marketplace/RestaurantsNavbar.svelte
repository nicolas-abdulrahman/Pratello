<script lang="ts">
  let { onOpenProfile }: { onOpenProfile: () => void } = $props();
</script>

<header class="flex items-center justify-between gap-4">
  <div class="flex items-center gap-3">
    <!-- Profile pill (opens sidebar) -->
    <div class="relative">
      <button
        class="flex items-center gap-2 p-1.5 pr-3 rounded-full bg-[#F4EDE2] hover:bg-[#EDE3D4] transition border border-[#E7DDCE] text-left group shadow-2xs"
        onclick={onOpenProfile}
      >
        <div class="relative">
          <div
            class="w-8 h-8 rounded-full bg-amber-500 text-white font-semibold flex items-center justify-center ring-2 ring-white shadow-xs overflow-hidden"
          >
            <span class="text-xs font-bold tracking-wider">G</span>
          </div>
          <span
            class="absolute bottom-0 right-0 w-2.5 h-2.5 bg-emerald-500 border-2 border-white rounded-full"
          ></span>
        </div>
        <div class="leading-none">
          <p class="text-[10px] text-stone-500 leading-none">Bem-vindo,</p>
          <p class="text-xs font-bold text-stone-900 font-sans tracking-tight mt-0.5">
            Olá Gregório
          </p>
        </div>
        <svg
          class="w-3.5 h-3.5 text-stone-400 group-hover:text-stone-700 transition-transform group-hover:translate-y-0.5 ml-1"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          viewBox="0 0 24 24"
        >
          <path d="M19 9l-7 7-7-7" stroke-linecap="round" stroke-linejoin="round" />
        </svg>
      </button>
    </div>

    <!-- Location pill -->
    <div
      class="inline-flex items-center gap-2 bg-[#F3ECE1] border border-[#E6DCCF] px-4 py-2 rounded-full text-stone-800 text-sm font-medium shadow-2xs hover:bg-[#EFE5D7] transition cursor-pointer"
    >
      <svg class="w-4 h-4 text-amber-600" fill="currentColor" viewBox="0 0 20 20">
        <path
          clip-rule="evenodd"
          d="M5.05 4.05a7 7 0 119.9 9.9L10 18.9l-4.95-4.95a7 7 0 010-9.9zM10 11a2 2 0 100-4 2 2 0 000 4z"
          fill-rule="evenodd"
        ></path>
      </svg>
      <span class="font-bold">São Paulo, SP</span>
      <svg class="w-3.5 h-3.5 text-stone-500" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
        <path d="M19 9l-7 7-7-7" stroke-linecap="round" stroke-linejoin="round" />
      </svg>
    </div>

    <!-- Distance hint -->
    <div class="hidden md:flex items-center gap-1.5 text-xs text-stone-500 font-medium">
      <span class="w-2 h-2 rounded-full bg-amber-500 inline-block"></span>
      Você está a aprox. <span class="font-bold text-stone-800">1.8 km</span> de diversos restaurantes
    </div>
  </div>

  <div class="flex items-center gap-3">
    <!-- Search -->
    <div class="relative w-80">
      <svg
        class="w-4 h-4 text-stone-400 absolute left-3.5 top-1/2 -translate-y-1/2"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        viewBox="0 0 24 24"
      >
        <circle cx="11" cy="11" r="8"></circle>
        <path d="m21 21-4.35-4.35" stroke-linecap="round"></path>
      </svg>
      <input
        class="w-full bg-[#F3ECE1] border border-[#E6DCCF] rounded-full pl-10 pr-4 py-2 text-xs font-medium text-stone-800 placeholder-stone-400 focus:outline-none focus:ring-2 focus:ring-amber-500 focus:bg-white transition"
        placeholder="Buscar restaurantes ou pratos..."
        type="text"
      />
    </div>

    <!-- Notifications -->
    <button
      aria-label="Notificações"
      class="w-10 h-10 rounded-full bg-[#F3ECE1] border border-[#E6DCCF] flex items-center justify-center text-stone-700 hover:bg-[#EFE5D7] transition shadow-2xs"
    >
      <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
        <path
          d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"
          stroke-linecap="round"
          stroke-linejoin="round"
        ></path>
      </svg>
    </button>
  </div>
</header>
