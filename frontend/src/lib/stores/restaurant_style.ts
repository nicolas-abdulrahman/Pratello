import { derived, get, writable } from 'svelte/store';

const baseColors = ['bg', 'primary_color', 'secondary_color', 'bg_dark', 'primary_color_dark', 'secondary_color_dark'] as const;
export type BaseColorKey = typeof baseColors[number];
export type ColorToken = BaseColorKey | `${'darker' | 'lighter'}_${BaseColorKey}` | 'text' | 'text_muted';
export type ColorPalette = Record<BaseColorKey, string>;

export interface StyledText {
  text: string;
  color?: ColorToken;
  dark?: boolean | null;
  gradient?: {
    dark?: boolean | null;
    end: number;
    smooth: boolean;
    first_color: ColorToken;
    second_color: ColorToken;
  } | null;
}

interface HeroButton {
  dark?: boolean | null;
  text?: { dark?: boolean | null } | null;
  bg?: ColorToken | null;
  href?: string;
}

interface Header1Block {
  id: string;
  type: 'header1';
  dark?: boolean | null;
  bg?: ColorToken;
  welcome?: StyledText;
  title?: StyledText;
  desc?: StyledText;
  button1?: HeroButton | null;
  button2?: HeroButton | null;
  img?: string;
}

export interface DisplayConfig {
  dark: boolean;
  colors: ColorPalette;
  navbar?: { dark?: boolean | null; bg_color?: ColorToken; border?: boolean; border_style?: 'solid' | 'dotted' | 'none'; border_color?: ColorToken } | null;
  menu: { container: 'grid' | 'horizontal' };
  first_page: [Header1Block | { id: string; type: 'header2'; dark?: boolean | null; img?: string }, ...Record<string, unknown>[]];
}

export type ResolvedText =
  | { mode: 'flat'; text: string; color: string }
  | { mode: 'rough'; text: string; first: string; second: string; firstColor: string; secondColor: string }
  | { mode: 'smooth'; text: string; gradientCss: string };

export function resolveColor(token: ColorToken, isDark: boolean): string {
  if (token === 'text') return isDark ? '#FFFFFF' : '#111111';
  if (token === 'text_muted') return isDark ? '#A1A1AA' : '#6B7280';
  return `var(--${isDark && !token.endsWith('_dark') ? `${token}_dark` : token})`;
}

function buttonHref(href?: string): string | undefined {
  const value = href?.trim();
  return value && /^(#|\/(?![\/\\])|https?:\/\/)/i.test(value) ? value : undefined;
}

function resolveButton(button: HeroButton | null | undefined, isDark: boolean, usePrimary = false) {
  const href = buttonHref(button?.href);
  isDark = button?.dark ?? isDark;
  const bg = button?.bg ?? (usePrimary ? 'primary_color' : null);
  return {
    href,
    bgCss: bg ? resolveColor(bg, isDark) : isDark ? 'var(--color-surface)' : 'white',
    textCss: resolveColor('text', button?.text?.dark ?? isDark),
    borderCss: bg === 'primary_color' ? resolveColor('darker_primary_color', isDark) : bg ? resolveColor(bg, isDark) : resolveColor('text', isDark),
  };
}

export const fallbackHeroButtons = [resolveButton(null, false), resolveButton(null, false)] as const;

export function resolveText(node: StyledText | undefined, isDark: boolean, fallback: ColorToken): ResolvedText {
  const text = node?.text ?? '';
  const gradient = node?.gradient;
  const end = gradient?.end;
  isDark = node?.dark ?? isDark;
  if (!gradient || typeof end !== 'number' || !Number.isFinite(end)) {
    return { mode: 'flat', text, color: resolveColor(node?.color ?? fallback, isDark) };
  }
  // The API counts Unicode code points, not UTF-16 code units.
  const chars = Array.from(text);
  const split = Math.max(0, Math.min(chars.length, Math.trunc(end)));
  const firstColor = resolveColor(gradient.first_color, gradient.dark ?? isDark);
  const secondColor = resolveColor(gradient.second_color, gradient.dark ?? isDark);
  if (!gradient.smooth || split === chars.length) {
    return { mode: 'rough', text, first: chars.slice(0, split).join(''), second: chars.slice(split).join(''), firstColor, secondColor };
  }
  const stop = split / Math.max(1, chars.length) * 100;
  return { mode: 'smooth', text, gradientCss: `linear-gradient(90deg, ${firstColor} 0%, ${firstColor} ${stop}%, ${secondColor} 100%)` };
}

const rawDisplayConfig = writable<DisplayConfig | null>(null);
const darkOverride = writable<boolean | null>(null);
export const restaurantDarkOverride = { subscribe: darkOverride.subscribe };

export const restaurantDark = derived([rawDisplayConfig, darkOverride], ([config, override]) => override ?? config?.dark ?? false);

export function toggleRestaurantDark() {
  darkOverride.set(!get(restaurantDark));
}

export function setRestaurantDisplayConfig(config: DisplayConfig | null) {
  rawDisplayConfig.set(config);
}

// A preview overrides local dark flags as well as the storefront default.
function overrideDarkFlags<T>(value: T, dark: boolean): T {
  if (Array.isArray(value)) return value.map((entry) => overrideDarkFlags(entry, dark)) as T;
  if (!value || typeof value !== 'object') return value;
  return Object.fromEntries(Object.entries(value).map(([key, entry]) => [key, key === 'dark' ? dark : overrideDarkFlags(entry, dark)])) as T;
}

export const restaurantDisplayConfig = derived([rawDisplayConfig, darkOverride], ([config, override]) =>
  config && override !== null ? overrideDarkFlags(config, override) : config,
);

export const restaurantCssVars = derived(rawDisplayConfig, (config) => {
  if (!config?.colors) return '';
  return baseColors.map((key) => {
    const hex = config.colors[key];
    if (!/^#[0-9a-f]{6}$/i.test(hex)) return '';
    const weight = key.startsWith('bg') ? 88 : 85;
    return `--${key}: ${hex}; --darker_${key}: color-mix(in srgb, var(--${key}) ${weight}%, black); --lighter_${key}: color-mix(in srgb, var(--${key}) ${weight}%, white);`;
  }).join(' ');
});

export const restaurantThemeVars = derived([restaurantCssVars, restaurantDark], ([palette, dark]) => {
  const suffix = dark ? '_dark' : '';
  return `${palette}
    color-scheme: ${dark ? 'dark' : 'light'};
    --color-background: var(--bg${suffix}, var(--restaurant-default-bg));
    --color-primary: var(--primary_color${suffix}, var(--restaurant-default-primary));
    --color-primary-hover: var(--darker_primary_color${suffix}, var(--restaurant-default-primary));
    --color-primary-soft: color-mix(in srgb, var(--color-primary) 18%, var(--color-background));
    --color-text: ${resolveColor('text', dark)};
    --color-text-muted: ${resolveColor('text_muted', dark)};
    --color-surface: ${dark ? 'var(--lighter_bg_dark, var(--restaurant-default-surface))' : 'var(--color-restaurant-surface)'};
    --color-surface-hover: color-mix(in srgb, var(--color-surface) 90%, var(--color-text));
    --color-border: color-mix(in srgb, var(--color-surface) 78%, var(--color-text));
    --color-navbar-bg: ${dark ? 'var(--lighter_bg_dark, var(--restaurant-default-surface))' : 'var(--color-restaurant-navbar-bg)'};
    --color-navbar-border: var(--color-border);`;
});

export const restaurantPageBackground = derived(restaurantDisplayConfig, (config) =>
  config ? `var(--${config.dark ? 'bg_dark' : 'bg'}, var(--color-background))` : 'var(--color-background)',
);

export const restaurantNavbar = derived(restaurantDisplayConfig, (config) => {
  const nav = config?.navbar;
  if (!nav || !config) return null;
  const dark = nav.dark ?? config.dark;
  const color = nav.bg_color ?? 'bg';
  const bg = color === 'text' || color === 'text_muted'
    ? `color-mix(in srgb, ${resolveColor(color, dark)} 85%, white)`
    : resolveColor(`lighter_${color.replace(/^(lighter|darker)_/, '')}` as ColorToken, dark);
  return {
    background: bg,
    borderStyle: !nav.border || nav.border_style === 'none' ? 'none' : nav.border_style === 'dotted' ? 'dotted' : 'solid',
    borderColor: nav.border_color ? resolveColor(nav.border_color, dark) : 'black',
  };
});

export const header1Rules = derived(restaurantDisplayConfig, (config) => {
  const hero = config?.first_page?.[0];
  if (!config || hero?.type !== 'header1') return null;
  const isDark = hero.dark ?? config.dark;
  return {
    id: hero.id, isDark, img: hero.img,
    bgCss: resolveColor(hero.bg ?? 'bg', isDark),
    accentCss: resolveColor('primary_color', isDark),
    textCss: resolveColor('text', isDark),
    mutedTextCss: resolveColor('text_muted', isDark),
    welcome: resolveText(hero.welcome, isDark, 'primary_color'),
    title: resolveText(hero.title, isDark, 'text'),
    desc: resolveText(hero.desc, isDark, 'text_muted'),
    button1: resolveButton(hero.button1, isDark, /^#[0-9a-f]{6}$/i.test(config.colors?.[isDark ? 'primary_color_dark' : 'primary_color'] ?? '')),
    button2: resolveButton(hero.button2, isDark),
  };
});
