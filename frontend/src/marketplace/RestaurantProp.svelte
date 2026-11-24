<script module lang="ts">
  export type RestaurantReview = {
    id: number;
    rating: number;
    comment?: string;
    created_at: string;
    author_name?: string;
    author_avatar_url?: string;
  };
</script>

<script lang="ts">
  import { onMount } from 'svelte';
  import { deliveryEstimate, type Coordinates } from '../lib/stores/location';
  import RestaurantLocationMap from './RestaurantLocationMap.svelte';

  let { name, faviconSrc, description, rating, reviewCount, reviews, address, position, customerPosition, menuHref, onclose }: {
    name: string;
    faviconSrc?: string;
    description?: string;
    rating?: number | null;
    reviewCount?: number;
    reviews?: RestaurantReview[];
    address?: string;
    position?: Coordinates;
    customerPosition?: Coordinates | null;
    menuHref: string;
    onclose: () => void;
  } = $props();

  let dialog: HTMLDialogElement;
  let pressedBackdrop = false;
  let failedAvatars = $state<Record<number, string>>({});
  let comments = $derived((reviews ?? []).filter((review) => review.comment?.trim())
    .sort((a, b) => Date.parse(b.created_at) - Date.parse(a.created_at)).slice(0, 5));

  onMount(() => {
    dialog.showModal();
    return () => { if (dialog.open) dialog.close(); };
  });

  function initials(name?: string) {
    return name?.trim().split(/\s+/).slice(0, 2).map((part) => part[0]).join('').toLocaleUpperCase('pt-BR') || '?';
  }
</script>

<dialog
  bind:this={dialog}
  {onclose}
  onpointerdown={(event) => { pressedBackdrop = event.target === dialog; }}
  onclick={(event) => {
    if (pressedBackdrop && event.target === dialog) dialog.close();
    pressedBackdrop = false;
  }}
  aria-labelledby="restaurant-preview-title"
  class="restaurant-preview m-auto h-[min(90dvh,48rem)] max-h-[calc(100dvh-1rem)] w-[calc(100%-2rem)] max-w-6xl overflow-hidden rounded-3xl border border-border bg-surface text-text shadow-xl backdrop:bg-black/45"
>
  <div class="grid h-full grid-rows-[auto_minmax(0,1fr)_auto] gap-space-xs p-space-sm sm:gap-space-md sm:p-space-md">
    <div class="flex items-start justify-between gap-space-sm">
      <div>
        <h2 id="restaurant-preview-title" class="flex items-center gap-space-2xs text-2xl font-bold">
          {#if faviconSrc}<img src={faviconSrc} alt="" class="size-7 shrink-0 object-contain" />{/if}
          <span class="min-w-0 wrap-break-word">{name}</span>
        </h2>
        {#if description}<p class="mt-space-3xs line-clamp-2 text-sm text-text-muted">{description}</p>{/if}
        {#if rating != null}
          <p class="mt-space-2xs font-semibold text-brand-accent"><span aria-hidden="true">★</span> {rating.toFixed(1).replace('.', ',')}/5 <span class="text-sm font-normal text-text-muted">({reviewCount} avaliações)</span></p>
        {/if}
      </div>
      <button type="button" class="shrink-0 rounded-lg px-space-xs py-space-2xs text-sm font-semibold text-brand-accent hover:bg-surface-hover" onclick={() => dialog.close()}><span aria-hidden="true">←</span> Voltar</button>
    </div>

    <div class="grid min-h-0 grid-rows-[minmax(0,3fr)_minmax(0,2fr)] gap-space-xs md:grid-cols-[minmax(0,3fr)_minmax(0,2fr)] md:grid-rows-[minmax(0,1fr)] md:gap-space-md">
    <section class="flex min-h-0 min-w-0 flex-col gap-space-2xs">
      <h3 class="font-bold">Localização</h3>
      {#if address}<p class="truncate text-sm text-text-muted" title={address}>{address}</p>{/if}
      {#if position}
        <RestaurantLocationMap restaurant={position} restaurantName={name} customer={customerPosition} />
        {#if customerPosition}
          <div class="flex shrink-0 items-center gap-space-sm text-sm font-semibold text-text" aria-label="Tempos aproximados em linha reta">
            <span class="inline-flex items-center gap-space-2xs" title="De moto">
              <svg aria-hidden="true" viewBox="0 0 24 24" class="size-5 text-brand-accent" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round">
                <circle cx="5" cy="17" r="3" /><circle cx="19" cy="17" r="3" /><path d="m5 17 4-6h6l-3 6H5m14 0-4-10h-3M9 11H5" />
              </svg>
              <span class="sr-only">De moto: </span>{deliveryEstimate(customerPosition, position)}
            </span>
            <span class="inline-flex items-center gap-space-2xs" title="A pé">
              <svg aria-hidden="true" viewBox="0 0 24 24" class="size-5 text-brand-accent" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round">
                <circle cx="13" cy="4" r="2" /><path d="m6 22 4-5 1-5m5 10v-5l-4-4 1-5m-8 5 3-1 3-4h2l3 4h3" />
              </svg>
              <span class="sr-only">A pé: </span>{deliveryEstimate(customerPosition, position, 'walking')}
            </span>
          </div>
        {/if}
      {:else}
        <p class="text-sm text-text-muted">Mapa indisponível no momento.</p>
      {/if}
    </section>

    <section class="flex min-h-0 min-w-0 flex-col gap-space-2xs rounded-xl border border-border bg-background p-space-sm">
      <h3 id="restaurant-reviews-title" class="font-bold">Comentários recentes</h3>
      <div role="region" aria-labelledby="restaurant-reviews-title" tabindex="0" class="min-h-0 flex-1 overflow-y-auto overscroll-contain pr-space-2xs">
      {#if comments.length}
        <ul class="divide-y divide-border">
          {#each comments as review (review.id)}
            {@const avatar = review.author_avatar_url}
            <li class="space-y-space-3xs py-space-2xs">
              <div class="flex flex-wrap items-center justify-between gap-space-2xs text-sm">
                <div class="flex min-w-0 items-center gap-space-2xs">
                  <span aria-hidden="true" class="grid size-9 shrink-0 place-items-center overflow-hidden rounded-full bg-primary-soft text-xs font-bold text-brand-accent">
                    {#if avatar && failedAvatars[review.id] !== avatar}
                      <img src={avatar} alt="" class="size-full object-cover" loading="lazy" referrerpolicy="no-referrer" onerror={() => { failedAvatars[review.id] = avatar; }} />
                    {:else}
                      {initials(review.author_name)}
                    {/if}
                  </span>
                  <span class="wrap-break-word font-semibold">{review.author_name?.trim() || 'Cliente anônimo'}</span>
                </div>
                <div class="flex items-center gap-space-xs">
                  <span class="font-semibold text-brand-accent"><span aria-hidden="true">★</span> {review.rating}/5</span>
                  <time datetime={review.created_at} class="text-text-muted">{new Date(review.created_at).toLocaleDateString('pt-BR')}</time>
                </div>
              </div>
              <p class="wrap-break-word text-sm">{review.comment}</p>
            </li>
          {/each}
        </ul>
      {:else}
        <p class="text-sm text-text-muted">{reviews ? 'Nenhum comentário recente.' : 'Comentários indisponíveis no momento.'}</p>
      {/if}
      </div>
    </section>

    </div>

    <div class="flex justify-end border-t border-border pt-space-xs">
      <a href={menuHref} class="inline-flex rounded-full bg-brand-accent px-space-md py-space-xs font-bold text-brand-on-accent hover:bg-brand-accent-hover">Ver cardápio →</a>
    </div>
  </div>
</dialog>

<style>
  :global(body:has(.restaurant-preview[open])) { overflow: hidden; }
</style>
