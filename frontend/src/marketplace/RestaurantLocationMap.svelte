<script lang="ts">
  import { divIcon, latLng, latLngBounds, map, marker, polyline, tileLayer } from 'leaflet';
  import 'leaflet/dist/leaflet.css';
  import type { Coordinates } from '../lib/stores/location';

  let { restaurant, restaurantName, customer }: { restaurant: Coordinates; restaurantName: string; customer?: Coordinates | null } = $props();
  let container = $state<HTMLDivElement>();

  function pin(kind: 'restaurant' | 'customer') {
    const symbol = kind === 'customer'
      ? '<path d="m7 12 7-6 7 6h-2v8h-4v-5h-2v5H9v-8Z" fill="var(--color-brand-on-accent)"/>'
      : '<circle cx="14" cy="13" r="4" fill="var(--color-brand-on-accent)"/>';
    return divIcon({
      className: `location-pin location-pin-${kind}`,
      html: `<svg viewBox="0 0 28 36" width="28" height="36" aria-hidden="true"><path d="M14 1a12 12 0 0 0-12 12c0 9 12 21 12 21s12-12 12-21A12 12 0 0 0 14 1Z" fill="currentColor" stroke="var(--color-brand-on-accent)" stroke-width="2"/>${symbol}</svg>`,
      iconSize: [28, 36],
      iconAnchor: [14, 35],
      tooltipAnchor: [0, -33]
    });
  }

  $effect(() => {
    if (!container) return;
    const host = container;
    const restaurantPoint = latLng(restaurant.latitude, restaurant.longitude);
    const customerPoint = customer ? latLng(customer.latitude, customer.longitude) : null;
    const colors = getComputedStyle(host);
    const restaurantColor = colors.getPropertyValue('--color-brand-accent').trim();
    const view = map(host, { scrollWheelZoom: true });
    tileLayer('https://tile.openstreetmap.org/{z}/{x}/{y}.png', {
      attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
      maxZoom: 19
    }).addTo(view);

    if (customerPoint) {
      polyline([customerPoint, restaurantPoint], { color: restaurantColor, weight: 4, dashArray: '8 7', interactive: false }).addTo(view);
      marker(customerPoint, { icon: pin('customer'), title: 'Sua localização aproximada' })
        .addTo(view);
    }
    const label = document.createElement('span');
    label.textContent = restaurantName;
    marker(restaurantPoint, { icon: pin('restaurant'), title: restaurantName })
      .bindTooltip(label, { permanent: true, direction: 'top' }).addTo(view);

    let centered = false;
    const observer = new ResizeObserver(() => {
      if (!host.clientWidth || !host.clientHeight) return;
      view.invalidateSize();
      if (centered) return;
      if (customerPoint) view.fitBounds(latLngBounds([restaurantPoint, customerPoint]), { paddingTopLeft: [24, 64], paddingBottomRight: [24, 24], maxZoom: 16 });
      else view.setView(restaurantPoint, 16);
      centered = true;
    });
    observer.observe(host);
    return () => { observer.disconnect(); view.remove(); };
  });
</script>

<div bind:this={container} role="region" aria-label="Mapa de localização" class="isolate min-h-0 w-full flex-1 rounded-xl border border-border"></div>

<style>
  :global(.location-pin-restaurant) { color: var(--color-brand-accent); }
  :global(.location-pin-customer) { color: var(--color-brand-primary-hover); }
</style>
