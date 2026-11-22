<script lang="ts">
  import LeftSidebarSection from './LeftSidebarSection.svelte';
  import AddressSection from './AddressSection.svelte';
  import { account } from '../lib/stores/account';
  import type { Snippet } from 'svelte';
  import { restaurantCssVars, restaurantDark, restaurantNavbar, toggleRestaurantDark } from '../lib/stores/restaurant_style';
  import AccountAvatar from './AccountAvatar.svelte';
  import { addressLabel } from '../lib/stores/address.js';

  let {
    cartItemCount = 0,
    slug = '',
    onOpenMenu = () => {},
    onGoHome = () => {},
    onGoRestaurantes = () => {},
    onGoLogin = () => {},
    onGoAccount = undefined as (() => void) | undefined,
    onOpenCart = () => {},
    showNavLinks = false,
    cartControls = undefined as Snippet | undefined,
    brand = false
  } = $props();

  let isSidebarOpen = $state(false);
  let isAddressOpen = $state(false);
</script>

<nav
  class="w-full border-b border-navbar-border bg-navbar-bg shadow-xs relative z-10 transition-colors"
  style={showNavLinks && $restaurantNavbar ? $restaurantCssVars : undefined}
  style:background-color={showNavLinks ? $restaurantNavbar?.background : undefined}
  style:border-bottom-style={showNavLinks ? $restaurantNavbar?.borderStyle : undefined}
  style:border-bottom-color={showNavLinks ? $restaurantNavbar?.borderColor : undefined}
>
  <div class="max-w-7xl mx-auto px-4 md:px-8 h-20 flex items-center justify-between gap-4 relative">
    <div class="flex items-center gap-3">
      {#if $account.isLoggedIn}
        <button
          onclick={() => (isSidebarOpen = true)}
          class="flex items-center gap-2 p-1 pr-3 rounded-full border border-navbar-border bg-white/90 hover:bg-white transition shadow-2xs text-left"
        >
          <div class="relative">
            <AccountAvatar class="size-7 rounded-full text-xs ring-2 ring-surface shadow-xs" />
            <span class="absolute bottom-0 right-0 w-2 h-2 bg-emerald-500 border-2 border-white rounded-full"></span>
          </div>
          <div class="leading-none">
            <p class="text-[10px] text-stone-500 leading-none">Bem-vindo,</p>
            <p class="text-xs font-bold text-stone-900 mt-0.5">
              Olá {$account.name.split(' ')[0] || 'Gregório'}
            </p>
          </div>
          <svg class="w-3.5 h-3.5 text-stone-400 ml-1" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
            <path d="M19 9l-7 7-7-7" stroke-linecap="round" stroke-linejoin="round" />
          </svg>
        </button>
      {:else}
        <button
          onclick={() => (isSidebarOpen = true)}
          class={brand
            ? "flex items-center gap-2 font-bold transition px-4 py-2 rounded-full bg-primary-soft hover:bg-white text-primary border border-primary/25 shadow-2xs hover:shadow-xs"
            : "flex items-center gap-2 font-bold transition px-4 py-2 rounded-full bg-white hover:bg-white/90 text-[#f05a1f] border border-white/40 shadow-2xs hover:shadow-xs"}
        >
          <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M19 21v-2a4 4 0 0 0-4-4H9a4 4 0 0 0-4 4v2" />
            <circle cx="12" cy="7" r="4" />
          </svg>
          Entrar
        </button>
      {/if}

      <!-- Address chip + popover wrapper -->
      <div class="relative">
        <button
          type="button"
          onclick={() => (isAddressOpen = !isAddressOpen)}
          class="inline-flex items-center gap-2 px-4 py-2 rounded-full text-sm font-medium shadow-2xs transition cursor-pointer bg-white/90 border border-navbar-border text-stone-800 hover:bg-white"
        >
          <svg class="w-4 h-4 text-primary" fill="currentColor" viewBox="0 0 20 20">
            <path clip-rule="evenodd" d="M5.05 4.05a7 7 0 119.9 9.9L10 18.9l-4.95-4.95a7 7 0 010-9.9zM10 11a2 2 0 100-4 2 2 0 000 4z" fill-rule="evenodd" />
          </svg>
          <span class="font-bold">{$addressLabel}</span>
          <svg
            class={`w-3.5 h-3.5 text-stone-500 transition-transform ${isAddressOpen ? 'rotate-180' : ''}`}
            fill="none"
            stroke="currentColor"
            stroke-width="2.5"
            viewBox="0 0 24 24"
          >
            <path d="M19 9l-7 7-7-7" stroke-linecap="round" stroke-linejoin="round" />
          </svg>
        </button>

        <AddressSection
          open={isAddressOpen}
          onClose={() => (isAddressOpen = false)}
        />
      </div>
    </div>

    {#if showNavLinks}
      <div class="hidden md:flex items-center gap-4 lg:gap-8 font-medium">
        <a
          href={slug ? `/r/${slug}` : '/'}
          onclick={(e) => { e.preventDefault(); onGoHome(); }}
          class={brand ? "text-stone-700 hover:text-primary transition-colors" : "text-white/95 hover:text-white transition-colors font-semibold drop-shadow-xs"}
        >
          Início
        </a>
        <a
          href={slug ? `/r/${slug}/cardapio` : '/cardapio'}
          onclick={(e) => { e.preventDefault(); onOpenMenu(); }}
          class={brand ? "text-stone-700 hover:text-primary transition-colors cursor-pointer" : "text-white/95 hover:text-white transition-colors cursor-pointer font-semibold drop-shadow-xs"}
        >
          Cardápio
        </a>
        <a
          href={slug ? `/r/${slug}/delivery` : '/restaurantes'}
          onclick={(e) => { e.preventDefault(); onGoRestaurantes(); }}
          class={brand ? "text-stone-700 hover:text-primary transition-colors cursor-pointer" : "text-white/95 hover:text-white transition-colors cursor-pointer font-semibold drop-shadow-xs"}
        >
          Delivery
        </a>
      </div>
      <div class="ml-auto flex shrink-0 items-center gap-space-2xs">
        <button type="button" onclick={toggleRestaurantDark} aria-label="Modo escuro" aria-pressed={$restaurantDark} title={$restaurantDark ? 'Ativar modo claro' : 'Ativar modo escuro'} class="grid size-10 place-items-center rounded-full border border-border bg-surface text-text shadow-sm hover:bg-surface-hover">
          <svg aria-hidden="true" viewBox="0 0 24 24" class="size-5" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            {#if $restaurantDark}
              <circle cx="12" cy="12" r="4" /><path d="M12 2v2m0 16v2M2 12h2m16 0h2M5 5l1.5 1.5m11 11L19 19M5 19l1.5-1.5m11-11L19 5" />
            {:else}
              <path d="M20.9 13.35A9 9 0 0 1 10.65 3.1 9 9 0 1 0 20.9 13.35Z" />
            {/if}
          </svg>
        </button>
        {@render cartControls?.()}
      </div>
    {/if}
  </div>
</nav>

<LeftSidebarSection
  open={isSidebarOpen}
  onClose={() => (isSidebarOpen = false)}
  {onGoLogin}
  {onGoAccount}
  {onGoHome}
/>
