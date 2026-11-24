import { resolveColor, resolveText, type ColorToken, type DisplayConfig, type StyledText } from '../../lib/stores/restaurant_style';

export const missingText = '<Not Found>';

export type SectionBlock = {
  dark?: boolean | null;
  bg?: ColorToken | null;
  border?: 'solid' | 'dotted' | 'none' | null;
  border_color?: ColorToken | null;
  container?: 'grid' | 'horizontal' | null;
  title?: StyledText | null;
  item_style?: ItemColors | null;
};

export function findSection(config: DisplayConfig | null, category: string): SectionBlock | undefined {
  return config?.first_page.find((block) => block.type === 'section' &&
    'display_items' in block && (block.display_items as { by?: string; value?: string } | undefined)?.by === 'category' &&
    (block.display_items as { value?: string } | undefined)?.value?.trim().toLowerCase() === category) as SectionBlock | undefined;
}

export function sectionText(text?: string | null): string {
  return text?.trim() ? text : missingText;
}

function withFallback(css: string, fallback: string): string {
  return css.startsWith('var(') ? `${css.slice(0, -1)}, ${fallback})` : css;
}

export function resolveSection(block: SectionBlock, config: Pick<DisplayConfig, 'dark'>) {
  const isDark = block.dark ?? config.dark;
  const border = block.border === 'solid' || block.border === 'dotted' ? block.border : 'none';

  return {
    isDark,
    background: withFallback(resolveColor(block.bg ?? 'bg', isDark), 'var(--color-background)'),
    border,
    container: block.container === 'grid' ? 'grid' : 'horizontal',
    borderColor: block.border_color ? resolveColor(block.border_color, isDark) : 'black',
    title: resolveText({ ...block.title, text: sectionText(block.title?.text ?? (block as any).name) }, isDark, 'text'),
  };
}

export type ItemColors = {
  dark?: boolean | null;
  bg?: ColorToken | null;
  title_color?: ColorToken | null;
  price_color?: ColorToken | null;
};

function shade(color: ColorToken, variant: 'darker' | 'lighter', isDark: boolean): string {
  if (color === 'text' || color === 'text_muted') {
    return `color-mix(in srgb, ${resolveColor(color, isDark)} 85%, ${variant === 'darker' ? 'black' : 'white'})`;
  }
  const base = color.replace(/^(darker|lighter)_/, '');
  return resolveColor(`${variant}_${base}` as ColorToken, isDark);
}

export function resolveItemColors(colors: ItemColors, isDark: boolean, sectionBg: ColorToken = 'bg') {
  isDark = colors.dark ?? isDark;
  const background = colors.bg ?? sectionBg;
  const price = withFallback(shade(colors.price_color ?? 'primary_color', 'lighter', isDark), 'color-mix(in srgb, var(--color-primary) 85%, white)');
  return {
    background: withFallback(resolveColor(background, isDark), 'var(--color-surface)'),
    title: withFallback(resolveColor(colors.title_color ?? 'text', isDark), 'var(--color-text)'),
    description: resolveColor('text_muted', isDark),
    price,
    addBackground: withFallback(shade(background, 'darker', isDark), 'color-mix(in srgb, var(--color-surface) 88%, black)'),
    addForeground: price,
  };
}
