export type Coordinates = { latitude: number; longitude: number };

const geocodeCache = new Map<string, Promise<Coordinates | null>>();

export function geocodeBrazil(query: string): Promise<Coordinates | null> {
  const key = query.trim().toLocaleLowerCase('pt-BR');
  if (!key) return Promise.resolve(null);
  const cached = geocodeCache.get(key);
  if (cached) return cached;

  const params = new URLSearchParams({ q: `${query}, Brasil`, limit: '1' });
  const pending = fetch(`https://photon.komoot.io/api/?${params}`, { signal: AbortSignal.timeout(8000) })
    .then(async (response) => {
      if (!response.ok) throw new Error('Geocoding unavailable');
      const { features } = await response.json();
      const result = features?.[0];
      const [longitude, latitude] = result?.geometry?.coordinates ?? [];
      return result?.properties?.countrycode === 'BR' && Number.isFinite(latitude) && Number.isFinite(longitude)
        ? { latitude, longitude }
        : null;
    })
    .catch(() => null)
    .then((point) => {
      if (!point) geocodeCache.delete(key);
      return point;
    });
  geocodeCache.set(key, pending);
  return pending;
}

export function deliveryEstimate(from: Coordinates, to: Coordinates): string {
  const radians = Math.PI / 180;
  const latitudeDelta = (to.latitude - from.latitude) * radians;
  const longitudeDelta = (to.longitude - from.longitude) * radians;
  const a = Math.sin(latitudeDelta / 2) ** 2 +
    Math.cos(from.latitude * radians) * Math.cos(to.latitude * radians) * Math.sin(longitudeDelta / 2) ** 2;
  const meters = 2 * 6_371_000 * Math.asin(Math.sqrt(Math.min(1, a)));
  const seconds = Math.round(meters);
  return seconds < 60 ? `~${Math.max(1, seconds)} s` : `~${Math.ceil(seconds / 60)} min`;
}
