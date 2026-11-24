import { loadMenuDraft, saveMenuDraft } from './menu';

export type MenuColors = {
  bg: string;
  primary: string;
  secondary: string;
  bg_dark: string;
  primary_dark: string;
  secondary_dark: string;
};

export type MenuSection = {
  id: string;
  type?: 'section' | 'promotion_banner';
  name: string;
  category: string;
  filter_by?: 'category' | 'tag' | 'collection';
  border: string;
  border_color?: string;
  layout: string;
  bg?: string;
  dark?: boolean | null;
  banner_id?: string;
  title_color?: string;
  gradient_enabled?: boolean;
  gradient_first?: string;
  gradient_second?: string;
  gradient_end?: number;
  gradient_smooth?: boolean;
  item_style_bg?: string;
  item_style_title?: string;
  item_style_price?: string;
  item_style_dark?: boolean | null;
};

export type MenuConfig = {
  name: string;
  picture: string;
  tags: string[];
  dark?: boolean;
  typography?: string;
  colors: MenuColors;
  sections: MenuSection[];
  hero_type?: 'header1' | 'header2';
  hero_bg?: string;
  hero_dark?: boolean | null;
  welcome_text?: string;
  description?: string;
  button1_text?: string;
  button1_bg?: string;
  button1_href?: string;
  button1_dark_text?: boolean | null;
  button2_text?: string;
  button2_bg?: string;
  button2_href?: string;
  button2_dark_text?: boolean | null;
  navbar_dark?: boolean | null;
  navbar_bg?: string;
  navbar_border?: boolean | null;
  navbar_border_style?: 'none' | 'solid' | 'dotted';
  navbar_border_color?: string;
  menu_container?: 'grid' | 'horizontal';
};

export const defaultConfig: MenuConfig = {
  name: '',
  picture: '',
  tags: [],
  dark: false,
  typography: 'Inter',
  colors: {
    bg: '#FFF9DB',
    primary: '#EA580C',
    secondary: '#FDE68A',
    bg_dark: '#431407',
    primary_dark: '#FACC15',
    secondary_dark: '#9A3412',
  },
  sections: [],
  hero_type: 'header1',
  hero_bg: 'bg',
  hero_dark: null,
  welcome_text: 'Olá!',
  description: 'Bem-vindo ao nosso restaurante',
  button1_text: 'Ver Cardápio',
  button1_bg: 'primary_color',
  button1_href: '#cardapio',
  button1_dark_text: true,
  button2_text: 'Promoções',
  button2_bg: 'secondary_color',
  button2_href: '#precos',
  button2_dark_text: false,
  navbar_dark: null,
  navbar_bg: 'primary_color',
  navbar_border: false,
  navbar_border_style: 'none',
  menu_container: 'grid',
};

export function readConfig(slug = ''): MenuConfig {
  try {
    const saved = slug ? (loadMenuDraft(slug) ?? {}) : JSON.parse(localStorage.getItem('pratello-menu-config') || '{}');
    return {
      ...defaultConfig,
      ...saved,
      colors: { ...defaultConfig.colors, ...(saved?.colors || {}) },
      sections: Array.isArray(saved?.sections) ? saved.sections : [],
    };
  } catch {
    return { ...defaultConfig };
  }
}

export function saveConfig(config: MenuConfig, slug = '') {
  if (slug) saveMenuDraft(slug, config);
  else localStorage.setItem('pratello-menu-config', JSON.stringify(config));
}

export function restaurantSlug(name: string) {
  return name
    .normalize('NFD')
    .replace(/[\u0300-\u036f]/g, '')
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-|-$/g, '');
}
