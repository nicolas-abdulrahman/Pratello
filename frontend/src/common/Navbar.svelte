<script lang="ts">
  import LeftSidebarSection from './LeftSidebarSection.svelte';
  import AddressSection from './AddressSection.svelte';
  import { account } from '../lib/stores/account';
  import type { Snippet } from 'svelte';
  import { restaurantCssVars, restaurantNavbar } from '../lib/stores/restaurant_style';
  import { brandDark, toggleBrandDark } from '../lib/stores/brand_theme';
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
    onGoOrders = undefined as (() => void) | undefined,
    onOpenCart = () => {},
    showNavLinks = false,
    cartControls = undefined as Snippet | undefined,
    brand = false
  } = $props();

  let isSidebarOpen = $state(false);
  let isAddressOpen = $state(false);

  // --- Floating cart state ---
  let cartSlotEl = $state<HTMLDivElement | undefined>(undefined);
  let cartFloatEl = $state<HTMLDivElement | undefined>(undefined);
  let cartReady = $state(false);
  let cartSize = $state({ width: 0, height: 0 });
  let cartPos = $state({ top: 0, left: 0 });

  $effect(() => {
    if (!cartFloatEl || !cartSlotEl) {
      cartReady = false;
      return;
    }

    const update = () => {
      if (!cartFloatEl || !cartSlotEl) return;
      const w = cartFloatEl.offsetWidth;
      const h = cartFloatEl.offsetHeight;
      const slotRect = cartSlotEl.getBoundingClientRect();
      // Document-space top of the slot (i.e. viewport top when scrollY = 0)
      const docTop = slotRect.top + window.scrollY;

      cartSize = { width: w, height: h };
      cartPos = { top: docTop, left: slotRect.left };
      cartReady = true;
    };

    update();

    const ro = new ResizeObserver(update);
    ro.observe(cartFloatEl);
    ro.observe(cartSlotEl);
    window.addEventListener('resize', update);

    return () => {
      ro.disconnect();
      window.removeEventListener('resize', update);
    };
  });
</script>

<nav
  class="w-full border-b border-navbar-border bg-navbar-bg shadow-xs relative z-10 transition-colors"
  style={showNavLinks && $restaurantNavbar ? $restaurantCssVars : undefined}
  style:--navbar-item-bg={$brandDark ? 'var(--lighter_primary_color)' : 'var(--lighter_primary_color)'}
  style:--navbar-item-text="#000000"
  style:background-color={$brandDark ? 'var(--darker_primary_color_dark, var(--color-primary))' : 'var(--lighter_primary_color, var(--color-primary))'}
  style:border-bottom-style={showNavLinks ? $restaurantNavbar?.borderStyle : undefined}
  style:border-bottom-color={showNavLinks ? $restaurantNavbar?.borderColor : undefined}
>
  <div class="max-w-7xl mx-auto px-4 md:px-8 h-20 flex items-center justify-between gap-4 relative">
    <div class="flex items-center gap-3">
      {#if $account.isLoggedIn}
        <button
          onclick={() => (isSidebarOpen = true)}
          class="navbar-item flex items-center gap-2 p-1 pr-3 rounded-full border border-navbar-border transition shadow-2xs text-left"
        >
          <div class="relative">
            <AccountAvatar class="size-7 rounded-full text-xs ring-2 ring-surface shadow-xs" />
            <span class="absolute bottom-0 right-0 w-2 h-2 bg-emerald-500 border-2 border-white rounded-full"></span>
          </div>
          <div class="leading-none">
            <p class="text-[10px] opacity-70 leading-none">Bem-vindo,</p>
            <p class="text-xs font-bold mt-0.5">
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
          class="navbar-item flex items-center gap-2 font-bold transition px-4 py-2 rounded-full border border-navbar-border shadow-2xs hover:shadow-xs"
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
          class="navbar-item inline-flex items-center gap-2 px-4 py-2 rounded-full text-sm font-medium shadow-2xs transition cursor-pointer border border-navbar-border"
        >
          <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
            <path clip-rule="evenodd" d="M5.05 4.05a7 7 0 119.9 9.9L10 18.9l-4.95-4.95a7 7 0 010-9.9zM10 11a2 2 0 100-4 2 2 0 000 4z" fill-rule="evenodd" />
          </svg>
          <span class="font-bold">{$addressLabel}</span>
          <svg
            class={`w-3.5 h-3.5 opacity-70 transition-transform ${isAddressOpen ? 'rotate-180' : ''}`}
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
      <!-- <div class="hidden md:flex items-center gap-4 lg:gap-8 font-medium"> -->
      <!--   <a -->
      <!--     href={slug ? `/r/${slug}` : '/'} -->
      <!--     onclick={(e) => { e.preventDefault(); onGoHome(); }} -->
      <!--     class={brand ? "text-stone-700 hover:text-primary transition-colors" : "text-white/95 hover:text-white transition-colors font-semibold drop-shadow-xs"} -->
      <!--   > -->
      <!--     Início -->
      <!--   </a> -->
      <!--   <a -->
      <!--     href={slug ? `/r/${slug}/cardapio` : '/cardapio'} -->
      <!--     onclick={(e) => { e.preventDefault(); onOpenMenu(); }} -->
      <!--     class={brand ? "text-stone-700 hover:text-primary transition-colors cursor-pointer" : "text-white/95 hover:text-white transition-colors cursor-pointer font-semibold drop-shadow-xs"} -->
      <!--   > -->
      <!--     Cardápio -->
      <!--   </a> -->
      <!--   <a -->
      <!--     href={slug ? `/r/${slug}/delivery` : '/restaurantes'} -->
      <!--     onclick={(e) => { e.preventDefault(); onGoRestaurantes(); }} -->
      <!--     class={brand ? "text-stone-700 hover:text-primary transition-colors cursor-pointer" : "text-white/95 hover:text-white transition-colors cursor-pointer font-semibold drop-shadow-xs"} -->
      <!--   > -->
      <!--     Delivery -->
      <!--   </a> -->
      <!-- </div> -->
    {/if}
      <div class="ml-auto flex shrink-0 items-center gap-space-2xs">
        <button type="button" onclick={toggleBrandDark} aria-label={$brandDark ? 'Ativar modo claro' : 'Ativar modo escuro'} aria-pressed={$brandDark} title={$brandDark ? 'Ativar modo claro' : 'Ativar modo escuro'} class="navbar-item grid size-10 place-items-center rounded-full border border-border shadow-sm">
          <svg aria-hidden="true" viewBox="0 0 24 24" class="size-5" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            {#if $brandDark}
              <circle cx="12" cy="12" r="4" /><path d="M12 2v2m0 16v2M2 12h2m16 0h2M5 5l1.5 1.5m11 11L19 19M5 19l1.5-1.5m11-11L19 5" />
            {:else}
              <path d="M20.9 13.35A9 9 0 0 1 10.65 3.1 9 9 0 1 0 20.9 13.35Z" />
            {/if}
          </svg>
        </button>

        <!-- Placeholder that reserves the exact space the cart used to occupy -->
        {#if showNavLinks}<div
          bind:this={cartSlotEl}
          aria-hidden="true"
          class="shrink-0"
          style:width="{cartSize.width}px"
          style:height="{cartSize.height}px"
        ></div>{/if}
      </div>
  </div>
</nav>

<!-- The actual cart: fixed so it floats on scroll, positioned over its navbar slot -->
{#if cartControls && showNavLinks}
  <div
    bind:this={cartFloatEl}
    class="fixed z-50"
    style:top="{cartPos.top}px"
    style:left="{cartPos.left}px"
    style:visibility={cartReady ? 'visible' : 'hidden'}
  >
    {@render cartControls()}
  </div>
{/if}

<LeftSidebarSection
  open={isSidebarOpen}
  onClose={() => (isSidebarOpen = false)}
  {onGoLogin}
  {onGoAccount}
  {onGoHome}
  {onGoOrders}
/>

<style>
  .navbar-item {
    background-color: var(--navbar-item-bg);
    color: var(--navbar-item-text);
  }
  .navbar-item:hover {
    filter: brightness(0.9);
  }
</style>
