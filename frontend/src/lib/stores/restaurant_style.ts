import { derived, writable } from 'svelte/store';

const baseColors = ['bg', 'primary_color', 'secondary_color', 'bg_dark', 'primary_color_dark', 'secondary_color_dark'] as const;
export type BaseColorKey = typeof baseColors[number];
export type ColorToken = BaseColorKey | `${'darker' | 'lighter'}_${BaseColorKey}` | 'text' | 'text_muted';
export type ColorPalette = Record<BaseColorKey, string>;

export interface StyledText {
  text: string;
  color?: ColorToken;
  text_white?: boolean | null;
  gradient?: {
    type: 'rough' | 'smooth';
    start_char: number;
    first_color: ColorToken;
    second_color: ColorToken;
  } | null;
}

interface HeroButton {
  text: string;
  bg?: ColorToken | null;
  text_color?: ColorToken | null;
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
  const bg = usePrimary ? 'primary_color' : button?.bg;
  return {
    href,
    bgCss: bg ? resolveColor(bg, isDark) : 'white',
    textCss: button?.text_color ? resolveColor(button.text_color, isDark) : bg && !isDark ? 'white' : 'black',
    borderCss: usePrimary ? resolveColor('darker_primary_color', isDark) : bg ? resolveColor(bg, isDark) : 'black',
  };
}

export const fallbackHeroButtons = [resolveButton(null, false), resolveButton(null, false)] as const;

function resolveText(node: StyledText | undefined, isDark: boolean, fallback: ColorToken): ResolvedText {
  const text = node?.text ?? '';
  const gradient = node?.gradient;
  if (!gradient || node?.text_white) {
    return { mode: 'flat', text, color: node?.text_white ? '#FFFFFF' : resolveColor(node?.color ?? fallback, isDark) };
  }
  // The API counts Unicode code points, not UTF-16 code units.
  const chars = Array.from(text);
  const start = Math.max(0, Math.min(chars.length, gradient.start_char));
  const firstColor = resolveColor(gradient.first_color, isDark);
  const secondColor = resolveColor(gradient.second_color, isDark);
  if (gradient.type === 'rough') {
    return { mode: 'rough', text, first: chars.slice(0, start).join(''), second: chars.slice(start).join(''), firstColor, secondColor };
  }
  const stop = start / Math.max(1, chars.length) * 100;
  return { mode: 'smooth', text, gradientCss: `linear-gradient(90deg, ${firstColor} 0%, ${firstColor} ${stop}%, ${secondColor} 100%)` };
}

export const rawDisplayConfig = writable<DisplayConfig | null>(null);
export function setRestaurantDisplayConfig(config: DisplayConfig | null) {
  rawDisplayConfig.set(config);
}

export const restaurantCssVars = derived(rawDisplayConfig, (config) => {
  if (!config?.colors) return '';
  return baseColors.map((key) => {
    const hex = config.colors[key];
    if (!/^#[0-9a-f]{6}$/i.test(hex)) return '';
    const weight = key.startsWith('bg') ? 88 : 85;
    return `--${key}: ${hex}; --darker_${key}: color-mix(in srgb, var(--${key}) ${weight}%, black); --lighter_${key}: color-mix(in srgb, var(--${key}) ${weight}%, white);`;
  }).join(' ');
});

export const header1Rules = derived(rawDisplayConfig, (config) => {
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
