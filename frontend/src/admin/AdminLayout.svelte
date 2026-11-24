<script lang="ts">
  import {
    ClipboardList,
    ArrowLeft,
    LayoutDashboard,
    Palette,
    Tag,
    Users,
    UtensilsCrossed
  } from 'lucide-svelte';

  export interface NavItem {
    id: string;
    label: string;
    icon: any;
    to?: string;
  }

  // --- Abstracted State & Dependencies ---
  export let session: { name?: string; role?: string; userId?: string } | null = {
    name: 'Administrador Pratello',
    role: 'admin',
    userId: 'u1'
  };
  export let activeTab: string = 'dashboard';
  export let onLogout: () => void = () => {};
  export let onResetDemoData: () => void = () => {};
  export let onNavigate: (tabId: string) => void = (id) => (activeTab = id);

  const NAV: NavItem[] = [
    { id: 'restaurante', label: 'Restaurante', icon: UtensilsCrossed },
    { id: 'dashboard', label: 'Dashboard', icon: LayoutDashboard },
    { id: 'funcionarios', label: 'Funcionários', icon: Users },
    { id: 'cardapio', label: 'Pratos', icon: UtensilsCrossed },
    { id: 'cores', label: 'Cores', icon: Palette },
    { id: 'precos', label: 'Preços', icon: Tag },
    { id: 'preparar', label: 'Preparar cardápio', icon: UtensilsCrossed },
    { id: 'pedidos', label: 'Painel de Pedidos', icon: ClipboardList },
  ];

  function goBack() {
    if (document.referrer && new URL(document.referrer).origin === window.location.origin && window.history.length > 1) {
      window.history.back();
    } else {
      window.location.href = '/me/my-restaurant';
    }
  }
</script>

<div class="min-h-screen bg-zinc-100 lg:flex">
  <!-- Sidebar desktop / topo mobile -->
  <aside class="sticky top-0 z-30 bg-zinc-800 lg:h-screen lg:w-64 lg:shrink-0">
    <div class="flex items-center justify-between px-4 py-3 lg:block lg:px-5 lg:py-6">
      <button type="button" on:click={goBack} class="mb-3 flex items-center gap-2 rounded-lg text-sm text-zinc-300 hover:text-white" aria-label="Voltar">
        <ArrowLeft size={18} /> Voltar
      </button>
      <div class="flex items-center gap-2">
        <span class="grid h-8 w-8 place-items-center rounded-lg bg-yellow-400 font-extrabold text-zinc-900">P</span>
        <span class="font-bold text-white">Pratello Admin</span>
      </div>
    </div>

    <nav class="flex gap-1 overflow-x-auto px-3 pb-3 lg:flex-col lg:px-3 lg:pb-0">
      {#each NAV as { id, label, icon: Icon, to }}
        {#if id === 'cardapio'}
          <span class="self-center px-3 text-xs font-bold uppercase tracking-wider text-zinc-400 lg:mt-4 lg:self-start">Menu</span>
        {/if}
        {#if id === 'dashboard'}
          <span class="self-center px-3 text-xs font-bold uppercase tracking-wider text-zinc-400 lg:mt-4 lg:self-start">Dashboard</span>
        {/if}
        {@const isActive = activeTab === id}
        <button
          type="button"
          on:click={() => {
            if (to) {
              window.location.href = to;
            } else {
              onNavigate(id);
            }
          }}
          class={`flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-medium transition-colors whitespace-nowrap text-left ${
            isActive ? 'bg-yellow-400 text-zinc-900' : 'text-zinc-300 hover:bg-zinc-700/60 hover:text-white'
          }`}
        >
          <svelte:component this={Icon} size={18} />
          <span>{label}</span>
        </button>
      {/each}
    </nav>

    <div class="absolute inset-x-0 bottom-0 hidden space-y-1 border-t border-zinc-700 p-3 lg:block">
      <div class="px-3 py-2">
        <p class="text-sm font-semibold text-white">{session?.name || 'Administrador'}</p>
        <p class="text-xs text-zinc-400">Administrador</p>
      </div>
    </div>
  </aside>

  <!-- Conteúdo principal -->
  <main class="min-w-0 flex-1 px-4 py-6 sm:px-8 sm:py-8">
    <slot />
  </main>
</div>
