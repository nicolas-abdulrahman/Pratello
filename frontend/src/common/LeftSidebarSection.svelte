<script lang="ts">
  import { account } from '../lib/stores/account';
  import AccountAvatar from './AccountAvatar.svelte';

  let {
    open = false,
    onClose = () => {},
    onGoLogin = () => {},
    onGoHome = () => {},
    onGoAccount = undefined as (() => void) | undefined,
    onGoOrders = undefined as (() => void) | undefined,
    restaurantSlug = '',
    restaurantName = undefined as string | undefined,
    restaurantLogoIconUrl = undefined as string | null | undefined
  } = $props();

  const close = () => onClose();

  function navigateTo(path: string, event: MouseEvent) {
    if (event.button !== 0 || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return;
    event.preventDefault();
    close();
    window.history.pushState({}, '', path);
    window.dispatchEvent(new PopStateEvent('popstate'));
  }

  function handleLogin() {
    onGoLogin();
    onClose();
  }

  function handleLogout() {
    account.logout();
    onClose();
  }

  function handleGoOrders(event: MouseEvent) {
    if (event.button !== 0 || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return;
    event.preventDefault();
    close();
    if (onGoOrders) {
      onGoOrders();
    } else {
      window.history.pushState({}, '', '/me/pedidos');
      window.dispatchEvent(new PopStateEvent('popstate'));
    }
  }
</script>

{#if open}
  <!-- Backdrop -->
  <div
    class="fixed inset-0 bg-black/60 backdrop-blur-sm z-40"
    onclick={close}
    role="button"
    tabindex="0"
    onkeydown={(e) => e.key === 'Escape' && close()}
    aria-label="Fechar menu"
  ></div>

  <!-- Sidebar -->
  <aside
    class="fixed top-0 left-0 bottom-0 w-80 max-w-full z-50 shadow-2xl p-5 sm:p-6 flex flex-col gap-6 overflow-y-auto border-r sidebar"
  >
    <!-- ===== Header: Restaurant or Pratello logo ===== -->
    <div class="flex items-center justify-between gap-2">
      {#if restaurantSlug}
        <a
          href={`/r/${encodeURIComponent(restaurantSlug)}`}
          onclick={(e) => navigateTo(`/r/${encodeURIComponent(restaurantSlug)}`, e)}
          class="flex items-center gap-3 min-w-0 rounded-xl p-1 -ml-1 theme-hover transition-colors"
        >
          {#if restaurantLogoIconUrl}
            <img
              src={restaurantLogoIconUrl}
              alt={restaurantName || restaurantSlug}
              class="w-11 h-11 object-contain shrink-0"
            />
          {:else}
            <div class="w-11 h-11 rounded-2xl flex items-center justify-center theme-accent shadow-inner shrink-0">
              <span class="text-lg font-black">{(restaurantName || restaurantSlug).charAt(0).toUpperCase()}</span>
            </div>
          {/if}
          <span class="text-2xl font-black tracking-tight theme-text font-serif truncate">
            {restaurantName || restaurantSlug}
          </span>
        </a>
      {:else}
        <div class="flex items-center gap-3">
          <div
            class="w-11 h-11 rounded-2xl flex items-center justify-center theme-accent shadow-inner"
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
          <span class="text-2xl font-black tracking-tight theme-text font-serif">Pratello</span>
        </div>
      {/if}
      <button
        onclick={close}
        class="w-8 h-8 rounded-full flex items-center justify-center theme-muted theme-hover transition"
        aria-label="Fechar"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
          <path d="M6 18L18 6M6 6l12 12" stroke-linecap="round" stroke-linejoin="round" />
        </svg>
      </button>
    </div>

    {#if $account.isLoggedIn}
      <!-- ===== Logged in: user card ===== -->
      <a
        href="/account"
        onclick={(event) => {
          if (!onGoAccount || event.button !== 0 || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return;
          event.preventDefault();
          close();
          onGoAccount();
        }}
        class="flex items-center gap-3 rounded-2xl border p-3 theme-card transition-colors"
        aria-label="Minha conta"
      >
        <AccountAvatar class="size-12 rounded-2xl shadow-sm" />
        <div class="min-w-0">
          <p class="text-sm font-bold theme-text leading-tight truncate">{$account.name}</p>
          <p class="text-xs theme-muted truncate">{$account.email}</p>
        </div>
      </a>

      <!-- ===== Logged in: nav links ===== -->
      <nav class="space-y-1.5 font-medium text-sm">
        <a
          href="#"
          onclick={(e) => { e.preventDefault(); onGoHome(); close(); }}
          class="flex items-center gap-3.5 px-4 py-3 rounded-2xl theme-active font-semibold transition border"
        >
          <svg class="w-5 h-5 shrink-0 theme-primary" fill="currentColor" viewBox="0 0 24 24">
            <path d="M10 20v-6h4v6h5v-8h3L12 3 2 12h3v8z"></path>
          </svg>
          Início
        </a>
        <a
          href="/me/my-restaurant"
          onclick={close}
          class="flex items-center gap-3.5 px-4 py-3 rounded-2xl theme-nav transition-colors"
        >
          <svg class="w-5 h-5 shrink-0 theme-icon" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
            <path d="M3 10h18M5 10v10h14V10M4 10l1-6h14l1 6M9 20v-6h6v6" stroke-linecap="round" stroke-linejoin="round"></path>
          </svg>
          Meus restaurantes
        </a>
        <a
          href="#restaurantes-proximos"
          onclick={close}
          class="flex items-center gap-3.5 px-4 py-3 rounded-2xl theme-nav transition-colors"
        >
          <svg class="w-5 h-5 shrink-0 theme-icon" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
            <circle cx="11" cy="11" r="8"></circle>
            <path d="m21 21-4.35-4.35" stroke-linecap="round"></path>
          </svg>
          Buscar
        </a>
        <a
          href="#"
          onclick={close}
          class="flex items-center gap-3.5 px-4 py-3 rounded-2xl theme-nav transition-colors"
        >
          <svg class="w-5 h-5 shrink-0 theme-icon" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
            <path
              d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z"
              stroke-linecap="round"
              stroke-linejoin="round"
            ></path>
          </svg>
          Favoritos
        </a>
        <a
          href="#"
          onclick={close}
          class="flex items-center gap-3.5 px-4 py-3 rounded-2xl theme-nav transition-colors"
        >
          <svg class="w-5 h-5 shrink-0 theme-icon" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
            <path
              d="M7 7h.01M7 3h5a1.99 1.99 0 011.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.99 1.99 0 013 12V5a2 2 0 012-2z"
              stroke-linecap="round"
              stroke-linejoin="round"
            ></path>
          </svg>
          Promoções
        </a>
        <a
          href="/me/pedidos"
          onclick={handleGoOrders}
          class="flex items-center gap-3.5 px-4 py-3 rounded-2xl theme-nav transition-colors"
        >
          <svg class="w-5 h-5 shrink-0 theme-icon" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
            <circle cx="12" cy="12" r="10"></circle>
            <polyline points="12 6 12 12 16 14"></polyline>
          </svg>
          Pedidos
        </a>
        <a
          href="#"
          onclick={close}
          class="flex items-center gap-3.5 px-4 py-3 rounded-2xl theme-nav transition-colors"
        >
          <svg class="w-5 h-5 shrink-0 theme-icon" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
            <circle cx="12" cy="12" r="10"></circle>
            <path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3" stroke-linecap="round"></path>
            <line x1="12" x2="12.01" y1="17" y2="17"></line>
          </svg>
          Ajuda
        </a>
      </nav>
    {:else}
      <!-- ===== Logged out: login options ===== -->
      <div class="space-y-3">
        <h3 class="text-xs font-bold theme-muted uppercase tracking-wider mb-2">Entrar como</h3>

        <button
          onclick={handleLogin}
          class="w-full flex items-center justify-between p-4 theme-login border rounded-2xl transition group text-left"
        >
          <div class="flex items-center gap-3">
            <div class="p-2 rounded-xl theme-login-icon shadow-sm">
              <svg class="w-5 h-5 shrink-0" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                <path d="M19 21v-2a4 4 0 0 0-4-4H9a4 4 0 0 0-4 4v2"></path>
                <circle cx="12" cy="7" r="4"></circle>
              </svg>
            </div>
            <div>
              <p class="font-bold theme-text">Login</p>
              <p class="text-xs theme-muted">Acesse sua conta</p>
            </div>
          </div>
          <svg
            class="w-5 h-5 shrink-0 theme-chevron transition-colors"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            viewBox="0 0 24 24"
          >
            <path d="m9 18 6-6-6-6" stroke-linecap="round" stroke-linejoin="round" />
          </svg>
        </button>
      </div>
    {/if}

    <!-- ===== Mais opções (always) ===== -->
    <div class="space-y-2 pt-4 border-t theme-divider">
      <h3 class="text-xs font-bold theme-muted uppercase tracking-wider mb-3">Mais opções</h3>

      <a
        href="/"
        onclick={(e) => navigateTo('/', e)}
        class="flex items-center gap-3 p-3 theme-nav rounded-xl transition-colors font-medium"
      >
        <svg class="w-5 h-5 shrink-0" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
          <circle cx="12" cy="12" r="10"></circle>
          <path d="M2 12h20M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"></path>
        </svg>
        Marketplace
      </a>

      <a
        href="/me/pedidos"
        onclick={handleGoOrders}
        class="flex items-center gap-3 p-3 theme-nav rounded-xl transition-colors font-medium"
      >
        <svg class="w-5 h-5 shrink-0" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
          <circle cx="12" cy="12" r="10"></circle>
          <polyline points="12 6 12 12 16 14"></polyline>
        </svg>
        Pedidos
      </a>

      <a
        href="#contact"
        onclick={close}
        class="flex items-center gap-3 p-3 theme-nav rounded-xl transition-colors font-medium"
      >
        <svg class="w-5 h-5 shrink-0" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
          <path
            d="M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.5 19.5 0 0 1-6-6 19.79 19.79 0 0 1-3.07-8.67A2 2 0 0 1 4.11 2h3a2 2 0 0 1 2 1.72 12.84 12.84 0 0 0 .7 2.81 2 2 0 0 1-.45 2.11L8.09 9.91a16 16 0 0 0 6 6l1.27-1.27a2 2 0 0 1 2.11-.45 12.84 12.84 0 0 0 2.81.7A2 2 0 0 1 22 16.92z"
          ></path>
        </svg>
        Fale Conosco
      </a>
    </div>

    <!-- ===== Logout (logged in only) ===== -->
    {#if $account.isLoggedIn}
      <div class="mt-auto">
        <button
          onclick={handleLogout}
          class="w-full flex items-center justify-center gap-2 px-3 py-2.5 text-sm font-semibold theme-logout rounded-xl border transition"
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
    {/if}
  </aside>
{/if}


<style>
  /* All colors follow the semantic variables of the parent theme. */
  .sidebar {
    background-color: #ffffff;
    color: var(--color-text);
    border-color: var(--color-border);
  }

  .theme-text {
    color: var(--color-text);
  }

  .theme-muted {
    color: var(--color-text-muted);
  }

  .theme-primary {
    color: var(--color-primary);
  }

  .theme-accent {
    background-color: var(--color-primary-soft);
    color: var(--color-primary);
  }

  .theme-hover:hover {
    background-color: var(--color-surface-hover);
    color: var(--color-text);
  }

  .theme-card {
    background-color: var(--color-surface);
    color: var(--color-text);
    border-color: var(--color-border);
  }

  .theme-card:hover {
    background-color: var(--color-surface-hover);
    border-color: var(--color-primary);
  }

  .theme-active {
    background-color: var(--color-primary-soft);
    color: var(--color-text);
    border-color: var(--color-border);
  }

  .theme-active:hover {
    border-color: var(--color-primary);
  }

  .theme-nav {
    color: var(--color-text);
  }

  .theme-icon {
    color: var(--color-text-muted);
  }

  .theme-nav:hover {
    background-color: var(--color-surface-hover);
    color: var(--color-text);
  }

  .theme-nav:hover .theme-icon,
  .theme-nav:hover > svg {
    color: var(--color-primary);
  }

  .theme-login {
    background-color: var(--color-primary-soft);
    color: var(--color-text);
    border-color: var(--color-border);
  }

  .theme-login:hover {
    border-color: var(--color-primary);
  }

  .theme-login-icon {
    background-color: var(--color-surface);
    color: var(--color-primary);
  }

  .theme-chevron {
    color: var(--color-text-muted);
  }

  .theme-login:hover .theme-chevron {
    color: var(--color-primary);
  }

  .theme-divider {
    border-color: var(--color-border);
  }

  .theme-logout {
    background-color: var(--color-surface);
    color: var(--color-text);
    border-color: var(--color-border);
  }

  .theme-logout:hover {
    background-color: var(--color-surface-hover);
    border-color: var(--color-primary);
  }

  .sidebar a:focus-visible,
  .sidebar button:focus-visible {
    outline: 2px solid var(--color-primary);
    outline-offset: 3px;
  }
</style>

