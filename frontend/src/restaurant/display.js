export function buildRootColorVars(colors) {
  return Object.entries(colors).flatMap(([key, hex]) => {
    const pct = key.startsWith('bg') ? 88 : 85;
    return [`--${key}:${hex}`, `--darker_${key}:color-mix(in srgb, ${hex} ${pct}%, black)`, `--lighter_${key}:color-mix(in srgb, ${hex} ${pct}%, white)`];
  }).join(';');
}

export function resolveColor(token, isDark) {
  if (token === 'text') return isDark ? '#ffffff' : '#111111';
  if (token === 'text_muted') return isDark ? '#a1a1aa' : '#6b7280';
  return `var(--${token}${isDark && !token.endsWith('_dark') ? '_dark' : ''})`;
}
