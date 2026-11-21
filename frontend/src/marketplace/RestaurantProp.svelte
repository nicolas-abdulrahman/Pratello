<script lang="ts">
  import { onMount } from 'svelte';
  import type { Coordinates } from './location';

  let { name, description, rating, reviewCount, reviews, address, position, menuHref, onclose }: {
    name: string;
    description?: string;
    rating?: number | null;
    reviewCount?: number;
    reviews?: { id: number; rating: number; comment?: string; created_at: string }[];
    address?: string;
    position?: Coordinates;
    menuHref: string;
    onclose: () => void;
  } = $props();

  let dialog: HTMLDialogElement;
  let comments = $derived((reviews ?? []).filter((review) => review.comment?.trim())
    .sort((a, b) => Date.parse(b.created_at) - Date.parse(a.created_at)).slice(0, 5));
  let mapUrl = $derived.by(() => {
    if (!position) return '';
    const { latitude, longitude } = position;
    const bounds = [longitude - 0.006, latitude - 0.004, longitude + 0.006, latitude + 0.004];
    const query = new URLSearchParams({ bbox: bounds.join(','), layer: 'mapnik', marker: `${latitude},${longitude}` });
    return `https://www.openstreetmap.org/export/embed.html?${query}`;
  });

  onMount(() => {
    dialog.showModal();
    return () => { if (dialog.open) dialog.close(); };
  });
</script>

<dialog bind:this={dialog} {onclose} aria-labelledby="restaurant-preview-title" class="m-auto max-h-[90dvh] w-[calc(100%-2rem)] max-w-2xl overflow-y-auto rounded-3xl border border-border bg-surface text-text shadow-xl backdrop:bg-black/45">
  <div class="space-y-space-md p-space-md sm:p-space-lg">
    <div class="flex items-start justify-between gap-space-sm">
      <div>
        <h2 id="restaurant-preview-title" class="text-2xl font-bold">{name}</h2>
        {#if description}<p class="mt-space-3xs text-sm text-text-muted">{description}</p>{/if}
        {#if rating != null}
          <p class="mt-space-2xs font-semibold text-brand-accent"><span aria-hidden="true">★</span> {rating.toFixed(1).replace('.', ',')}/5 <span class="text-sm font-normal text-text-muted">({reviewCount} avaliações)</span></p>
        {/if}
      </div>
      <button type="button" class="rounded-lg px-space-xs py-space-2xs text-xl hover:bg-surface-hover" aria-label="Fechar detalhes" onclick={() => dialog.close()}>×</button>
    </div>

    <section class="space-y-space-2xs">
      <h3 class="font-bold">Localização</h3>
      {#if address}<p class="text-sm text-text-muted">{address}</p>{/if}
      {#if mapUrl}
        <iframe title={`Mapa de ${name}`} src={mapUrl} class="h-52 w-full rounded-xl border border-border" loading="lazy" referrerpolicy="no-referrer"></iframe>
      {:else}
        <p class="text-sm text-text-muted">Mapa indisponível no momento.</p>
      {/if}
    </section>

    <section class="space-y-space-2xs">
      <h3 class="font-bold">Comentários recentes</h3>
      {#if comments.length}
        <ul class="divide-y divide-border">
          {#each comments as review (review.id)}
            <li class="space-y-space-3xs py-space-2xs">
              <div class="flex items-center justify-between gap-space-sm text-sm">
                <span class="font-semibold text-brand-accent"><span aria-hidden="true">★</span> {review.rating}/5</span>
                <time datetime={review.created_at} class="text-text-muted">{new Date(review.created_at).toLocaleDateString('pt-BR')}</time>
              </div>
              <p class="wrap-break-word text-sm">{review.comment}</p>
            </li>
          {/each}
        </ul>
      {:else}
        <p class="text-sm text-text-muted">{reviews ? 'Nenhum comentário recente.' : 'Comentários indisponíveis no momento.'}</p>
      {/if}
    </section>

    <a href={menuHref} class="inline-flex rounded-full bg-brand-accent px-space-md py-space-xs font-bold text-brand-on-accent hover:bg-brand-accent-hover">Ver cardápio →</a>
  </div>
</dialog>
