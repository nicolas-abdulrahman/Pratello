import { account } from '../lib/stores/account';
import type { MenuConfig } from './menuConfig';

// ==========================================
// Types according to Pratello OpenAPI 2.2.16
// ==========================================

export type BaseColorKey =
  | 'bg'
  | 'primary_color'
  | 'secondary_color'
  | 'bg_dark'
  | 'primary_color_dark'
  | 'secondary_color_dark';

export type ItemStyleToken = BaseColorKey | 'text' | 'text_muted';

export type DarkerDisplayToken =
  | 'darker_bg'
  | 'darker_primary_color'
  | 'darker_secondary_color'
  | 'darker_bg_dark'
  | 'darker_primary_color_dark'
  | 'darker_secondary_color_dark';

export type LighterDisplayToken =
  | 'lighter_bg'
  | 'lighter_primary_color'
  | 'lighter_secondary_color'
  | 'lighter_bg_dark'
  | 'lighter_primary_color_dark'
  | 'lighter_secondary_color_dark';

export type DisplayToken = ItemStyleToken | DarkerDisplayToken | LighterDisplayToken;

export interface ColorPalette {
  bg: string;
  primary_color: string;
  secondary_color: string;
  bg_dark: string;
  primary_color_dark: string;
  secondary_color_dark: string;
}

export interface Gradient {
  end: number;
  smooth: boolean;
  first_color: DisplayToken;
  second_color: DisplayToken;
  dark?: boolean | null;
}

export interface StyledText {
  text: string;
  color?: DisplayToken;
  dark?: boolean | null;
  gradient?: Gradient | null;
}

export interface HeroDisplayButton {
  bg?: DisplayToken | null;
  href?: string;
  dark?: boolean | null;
  text?: { dark?: boolean | null } | null;
}

export interface Header1Block {
  id: string;
  type: 'header1';
  dark?: boolean | null;
  bg?: DisplayToken;
  img?: string;
  welcome?: StyledText;
  title?: StyledText;
  desc?: StyledText;
  button1?: HeroDisplayButton | null;
  button2?: HeroDisplayButton | null;
}

export interface Header2Block {
  id: string;
  type: 'header2';
  dark?: boolean | null;
  bg?: DisplayToken;
  img?: string;
}

export interface ItemDisplayStyle {
  bg?: ItemStyleToken;
  title_color?: ItemStyleToken;
  price_color?: ItemStyleToken;
  dark?: boolean | null;
}

export interface SectionBlock {
  id: string;
  type: 'section';
  dark?: boolean | null;
  bg?: DisplayToken;
  border: 'none' | 'solid' | 'dotted';
  border_color?: DisplayToken;
  container: 'grid' | 'horizontal';
  title: StyledText;
  display_items: {
    by: 'category' | 'tag' | 'collection';
    value: string;
  };
  item_style?: ItemDisplayStyle;
}

export interface PromotionBannerBlock {
  id: string;
  type: 'promotion_banner';
  banner_id: string;
  dark?: boolean | null;
}

export type FirstPageBlock = Header1Block | Header2Block | SectionBlock | PromotionBannerBlock;

export interface NavbarConfig {
  dark?: boolean | null;
  bg_color: ItemStyleToken;
  border?: boolean | null;
  border_style?: 'none' | 'solid' | 'dotted';
  border_color?: DisplayToken;
}

export interface DisplayConfig {
  dark: boolean;
  colors: ColorPalette;
  navbar?: NavbarConfig;
  menu: {
    container: 'grid' | 'horizontal';
  };
  first_page: FirstPageBlock[];
}

export interface StorefrontMenuPayload {
  name: string;
  typography?: string;
  display_config: DisplayConfig;
}

// Banner record from GET /api/restaurants/{slug}/menu/banners
export interface PromotionalBanner {
  id: string;
  code: string;
  title: string;
  description: string;
  background_image_url: string;
}

// ==========================================
// Persistent Storage (Dictionary per Restaurant)
// ==========================================

const DRAFT_STORAGE_KEY = 'pratello-menu-layouts';
const LAYOUT_STORAGE_KEY = 'pratello-chosen-menu-layouts';

const memoryDrafts: Record<string, MenuConfig> = {};
const memoryLayouts: Record<string, StorefrontMenuPayload> = {};

function getStorage(): Storage | null {
  if (typeof window !== 'undefined' && window.localStorage) return window.localStorage;
  if (typeof localStorage !== 'undefined') return localStorage;
  return null;
}

/**
 * Returns all saved drafts as a dictionary { [restaurantSlug]: MenuConfig }
 */
export function getAllMenuDrafts(): Record<string, MenuConfig> {
  const store = getStorage();
  if (!store) return { ...memoryDrafts };
  try {
    return JSON.parse(store.getItem(DRAFT_STORAGE_KEY) || '{}');
  } catch {
    return { ...memoryDrafts };
  }
}

/**
 * Load draft MenuConfig for a specific restaurant slug.
 */
export function loadMenuDraft(slug: string): MenuConfig | null {
  if (!slug) return null;
  const store = getStorage();
  if (!store) return memoryDrafts[slug] ?? null;
  try {
    const all = JSON.parse(store.getItem(DRAFT_STORAGE_KEY) || '{}');
    return all[slug] ?? null;
  } catch {
    return memoryDrafts[slug] ?? null;
  }
}

/**
 * Save draft MenuConfig in the dictionary under restaurant slug.
 */
export function saveMenuDraft(slug: string, config: MenuConfig): void {
  if (!slug) throw new Error('Selecione um restaurante antes de salvar o cardápio.');
  memoryDrafts[slug] = config;
  const store = getStorage();
  if (!store) return;
  try {
    const all = JSON.parse(store.getItem(DRAFT_STORAGE_KEY) || '{}');
    all[slug] = config;
    store.setItem(DRAFT_STORAGE_KEY, JSON.stringify(all));
  } catch {
    throw new Error('Não foi possível ler os rascunhos salvos.');
  }
}

/**
 * Returns all chosen menu layouts as a dictionary { [restaurantSlug]: StorefrontMenuPayload }
 */
export function getAllRestaurantLayouts(): Record<string, StorefrontMenuPayload> {
  const store = getStorage();
  if (!store) return { ...memoryLayouts };
  try {
    return JSON.parse(store.getItem(LAYOUT_STORAGE_KEY) || '{}');
  } catch {
    return { ...memoryLayouts };
  }
}

/**
 * Loads the chosen StorefrontMenuPayload for a restaurant slug from persistent storage.
 */
export function loadRestaurantLayout(slug: string): StorefrontMenuPayload | null {
  if (!slug) return null;
  const store = getStorage();
  if (!store) return memoryLayouts[slug] ?? null;
  try {
    const dict = JSON.parse(store.getItem(LAYOUT_STORAGE_KEY) || '{}');
    return dict[slug] ?? null;
  } catch {
    return memoryLayouts[slug] ?? null;
  }
}

/**
 * Stores the chosen StorefrontMenuPayload in persistent storage keyed by restaurant slug.
 */
export function saveRestaurantLayout(slug: string, layout: StorefrontMenuPayload): void {
  if (!slug) throw new Error('Informe o slug do restaurante para persistir o layout.');
  memoryLayouts[slug] = layout;
  const store = getStorage();
  if (!store) return;
  try {
    const dict = JSON.parse(store.getItem(LAYOUT_STORAGE_KEY) || '{}');
    dict[slug] = layout;
    store.setItem(LAYOUT_STORAGE_KEY, JSON.stringify(dict));
  } catch {
    throw new Error('Não foi possível salvar o layout no armazenamento persistente.');
  }
}

/**
 * Removes a restaurant's layout from persistent storage dictionary.
 */
export function deleteRestaurantLayout(slug: string): void {
  if (!slug) return;
  delete memoryLayouts[slug];
  const store = getStorage();
  if (!store) return;
  try {
    const dict = JSON.parse(store.getItem(LAYOUT_STORAGE_KEY) || '{}');
    delete dict[slug];
    store.setItem(LAYOUT_STORAGE_KEY, JSON.stringify(dict));
  } catch {
    // Ignore cleanup errors
  }
}

// ==========================================
// API Client Connections
// ==========================================

export async function restaurantRequest(slug: string, path: string, init: RequestInit = {}) {
  if (!slug) throw new Error('Abra o admin de um restaurante para acessar a API.');
  const response = await account.authenticatedFetch(`/restaurants/${encodeURIComponent(slug)}${path}`, init);
  if (!response.ok) {
    const body = await response.text();
    throw new Error(`API (${response.status}): ${body || response.statusText}`);
  }
  if (response.status === 204) return null;
  return response.json();
}

/**
 * Fetch current storefront layout from GET /api/restaurants/{slug}/menu
 */
export async function getStorefrontLayout(slug: string): Promise<StorefrontMenuPayload> {
  return restaurantRequest(slug, '/menu');
}

/**
 * Save/Deploy storefront layout via PUT /api/restaurants/{slug}/menu
 */
export async function updateStorefrontLayout(slug: string, payload: StorefrontMenuPayload): Promise<StorefrontMenuPayload> {
  const errors = validateMenuPayload(payload);
  if (errors.length > 0) {
    throw new Error(`Layout inválido: ${errors.join('; ')}`);
  }
  const result = await restaurantRequest(slug, '/menu', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
  // Also keep persistence storage dictionary updated!
  saveRestaurantLayout(slug, payload);
  return result;
}

/**
 * Fetch promotional banners for restaurant from GET /api/restaurants/{slug}/menu/banners
 */
export async function getPromotionalBanners(slug: string): Promise<PromotionalBanner[]> {
  try {
    const data = await restaurantRequest(slug, '/menu/banners');
    return Array.isArray(data) ? data : [];
  } catch {
    return [];
  }
}

/**
 * Fetch categories for restaurant from GET /api/restaurants/{slug}/categories
 */
export async function getMenuCategories(slug: string): Promise<Array<{ id: string; name: string }>> {
  try {
    const data = await restaurantRequest(slug, '/categories');
    return Array.isArray(data) ? data : [];
  } catch {
    return [];
  }
}

// ==========================================
// Validation Helpers (OpenAPI & Manual Rules)
// ==========================================

const HEX_COLOR_REGEX = /^#[0-9A-Fa-f]{6}$/;

const BASE_TOKENS = new Set<string>([
  'bg',
  'primary_color',
  'secondary_color',
  'bg_dark',
  'primary_color_dark',
  'secondary_color_dark',
  'text',
  'text_muted',
]);

const DISPLAY_TOKENS = new Set<string>([
  ...BASE_TOKENS,
  'darker_bg',
  'darker_primary_color',
  'darker_secondary_color',
  'darker_bg_dark',
  'darker_primary_color_dark',
  'darker_secondary_color_dark',
  'lighter_bg',
  'lighter_primary_color',
  'lighter_secondary_color',
  'lighter_bg_dark',
  'lighter_primary_color_dark',
  'lighter_secondary_color_dark',
]);

function countRunes(str: string): number {
  return Array.from(str).length;
}

export function validateMenuPayload(payload: StorefrontMenuPayload): string[] {
  const errors: string[] = [];
  if (!payload || typeof payload !== 'object') {
    return ['Payload deve ser um objeto'];
  }
  if (!payload.name || !payload.name.trim()) {
    errors.push('Nome do cardápio/restaurante é obrigatório');
  }

  const dc = payload.display_config;
  if (!dc || typeof dc !== 'object') {
    errors.push('display_config é obrigatório');
    return errors;
  }

  if (typeof dc.dark !== 'boolean') {
    errors.push('display_config.dark deve ser booleano');
  }

  // 6 base colors
  const requiredColors: BaseColorKey[] = [
    'bg',
    'primary_color',
    'secondary_color',
    'bg_dark',
    'primary_color_dark',
    'secondary_color_dark',
  ];
  if (!dc.colors || typeof dc.colors !== 'object') {
    errors.push('display_config.colors é obrigatório com 6 chaves');
  } else {
    for (const key of requiredColors) {
      const val = dc.colors[key];
      if (!val || !HEX_COLOR_REGEX.test(val)) {
        errors.push(`Cor inválida para '${key}': deve ser hexadecimal no formato #RRGGBB`);
      }
    }
  }

  // Navbar validation if provided
  if (dc.navbar) {
    if (!dc.navbar.bg_color || !BASE_TOKENS.has(dc.navbar.bg_color)) {
      errors.push(`navbar.bg_color deve ser um itemStyleToken base válido (${Array.from(BASE_TOKENS).join(', ')})`);
    }
    if (dc.navbar.border_color && !DISPLAY_TOKENS.has(dc.navbar.border_color)) {
      errors.push(`navbar.border_color deve ser um displayToken válido`);
    }
    if (dc.navbar.border_style && !['none', 'solid', 'dotted'].includes(dc.navbar.border_style)) {
      errors.push(`navbar.border_style deve ser 'none', 'solid' ou 'dotted'`);
    }
  }

  // Menu layout container
  if (!dc.menu || !['grid', 'horizontal'].includes(dc.menu.container)) {
    errors.push("display_config.menu.container deve ser 'grid' ou 'horizontal'");
  }

  // First page blocks
  if (!Array.isArray(dc.first_page) || dc.first_page.length === 0) {
    errors.push('display_config.first_page deve conter pelo menos 1 bloco');
    return errors;
  }

  const firstBlock = dc.first_page[0];
  if (firstBlock.type !== 'header1' && firstBlock.type !== 'header2') {
    errors.push('first_page[0] deve ser um cabeçalho (header1 ou header2)');
  }

  const ids = new Set<string>();
  dc.first_page.forEach((block, idx) => {
    if (!block.id || !block.id.trim()) {
      errors.push(`Bloco no índice ${idx} precisa de um id não vazio`);
    } else if (ids.has(block.id)) {
      errors.push(`Id de bloco duplicado: '${block.id}'`);
    } else {
      ids.add(block.id);
    }

    if (idx > 0 && (block.type === 'header1' || block.type === 'header2')) {
      errors.push(`Bloco no índice ${idx} não pode ser um cabeçalho (apenas o primeiro bloco pode)`);
    }

    if (block.type === 'section') {
      if (!['none', 'solid', 'dotted'].includes(block.border)) {
        errors.push(`Seção '${block.id}' deve ter border 'none', 'solid' ou 'dotted'`);
      }
      if (!['grid', 'horizontal'].includes(block.container)) {
        errors.push(`Seção '${block.id}' deve ter container 'grid' ou 'horizontal'`);
      }
      if (!block.title || !block.title.text) {
        errors.push(`Seção '${block.id}' requer title.text`);
      } else if (block.title.gradient) {
        const runes = countRunes(block.title.text);
        if (block.title.gradient.end < 0 || block.title.gradient.end > runes) {
          errors.push(`Seção '${block.id}' gradiente end (${block.title.gradient.end}) fora do intervalo de runas 0..${runes}`);
        }
      }
      if (!block.display_items || !['category', 'tag', 'collection'].includes(block.display_items.by) || !block.display_items.value) {
        errors.push(`Seção '${block.id}' requer display_items válido com by ('category'|'tag'|'collection') e value`);
      }
      if (block.item_style) {
        if (block.item_style.bg && !BASE_TOKENS.has(block.item_style.bg)) {
          errors.push(`Seção '${block.id}' item_style.bg deve usar apenas base tokens`);
        }
        if (block.item_style.title_color && !BASE_TOKENS.has(block.item_style.title_color)) {
          errors.push(`Seção '${block.id}' item_style.title_color deve usar apenas base tokens`);
        }
        if (block.item_style.price_color && !BASE_TOKENS.has(block.item_style.price_color)) {
          errors.push(`Seção '${block.id}' item_style.price_color deve usar apenas base tokens`);
        }
      }
    } else if (block.type === 'promotion_banner') {
      if (!block.banner_id || !block.banner_id.trim()) {
        errors.push(`Banner '${block.id}' requer banner_id não vazio`);
      }
      // Rejection check: cannot contain border_color or item_style
      if ('border_color' in block || 'item_style' in block) {
        errors.push(`Bloco de banner promocional '${block.id}' não pode conter border_color nem item_style`);
      }
    }
  });

  return errors;
}

// ==========================================
// Payload Builder from MenuConfig
// ==========================================

export function buildMenuPayload(config: MenuConfig): StorefrontMenuPayload {
  const colors = config.colors;
  if (!config.name.trim()) throw new Error('Informe o nome do restaurante.');

  const colorPalette: ColorPalette = {
    bg: colors.bg,
    primary_color: colors.primary,
    secondary_color: colors.secondary,
    bg_dark: colors.bg_dark,
    primary_color_dark: colors.primary_dark,
    secondary_color_dark: colors.secondary_dark,
  };

  for (const [k, v] of Object.entries(colorPalette)) {
    if (!HEX_COLOR_REGEX.test(v)) {
      throw new Error(`A cor ${k} (${v}) não é um hexadecimal válido.`);
    }
  }

  // Build first_page blocks
  const firstPageBlocks: FirstPageBlock[] = [];

  // Hero header block (header1 or header2)
  if (config.hero_type === 'header2') {
    const h2: Header2Block = {
      id: 'hero_main',
      type: 'header2',
      dark: config.hero_dark ?? null,
      bg: (config.hero_bg as DisplayToken) || 'bg',
      img: config.picture || undefined,
    };
    firstPageBlocks.push(h2);
  } else {
    const h1: Header1Block = {
      id: 'hero_main',
      type: 'header1',
      dark: config.hero_dark ?? null,
      bg: (config.hero_bg as DisplayToken) || 'bg',
      img: config.picture || undefined,
      welcome: config.welcome_text ? { text: config.welcome_text, color: 'primary_color' } : undefined,
      title: { text: config.name },
      desc: config.description ? { text: config.description, color: 'text_muted' } : undefined,
      button1: config.button1_text
        ? {
            bg: (config.button1_bg as DisplayToken) || 'primary_color',
            href: config.button1_href || '#cardapio',
            text: { dark: config.button1_dark_text ?? true },
          }
        : null,
      button2: config.button2_text
        ? {
            bg: (config.button2_bg as DisplayToken) || 'secondary_color',
            href: config.button2_href || '#precos',
            text: { dark: config.button2_dark_text ?? false },
          }
        : null,
    };
    firstPageBlocks.push(h1);
  }

  // Preços / Banners / Sections
  if (Array.isArray(config.sections)) {
    for (const section of config.sections) {
      if (!section.name.trim()) continue;

      if (section.type === 'promotion_banner') {
        firstPageBlocks.push({
          id: section.id || `promo_${crypto.randomUUID()}`,
          type: 'promotion_banner',
          banner_id: section.banner_id || section.category || 'promo',
          dark: section.dark ?? null,
        });
      } else {
        const borderStyle: 'none' | 'solid' | 'dotted' =
          section.border === 'border' || section.border === 'solid'
            ? 'solid'
            : section.border === 'dotted'
              ? 'dotted'
              : 'none';

        const containerStyle: 'grid' | 'horizontal' = section.layout === 'horizontal' ? 'horizontal' : 'grid';

        const styledTitle: StyledText = {
          text: section.name,
          color: (section.title_color as DisplayToken) || undefined,
          dark: section.dark ?? null,
        };

        if (section.gradient_enabled && section.gradient_first && section.gradient_second) {
          styledTitle.gradient = {
            end: section.gradient_end ?? Math.floor(countRunes(section.name) / 2),
            smooth: !!section.gradient_smooth,
            first_color: section.gradient_first as DisplayToken,
            second_color: section.gradient_second as DisplayToken,
          };
        }

        const secBlock: SectionBlock = {
          id: section.id || `sec_${crypto.randomUUID()}`,
          type: 'section',
          dark: section.dark ?? null,
          bg: (section.bg as DisplayToken) || 'bg',
          border: borderStyle,
          border_color: section.border_color ? (section.border_color as DisplayToken) : undefined,
          container: containerStyle,
          title: styledTitle,
          display_items: {
            by: (section.filter_by as 'category' | 'tag' | 'collection') || 'category',
            value: section.category || 'all',
          },
        };

        if (section.item_style_bg || section.item_style_title || section.item_style_price) {
          secBlock.item_style = {
            bg: (section.item_style_bg as ItemStyleToken) || undefined,
            title_color: (section.item_style_title as ItemStyleToken) || undefined,
            price_color: (section.item_style_price as ItemStyleToken) || undefined,
            dark: section.item_style_dark ?? null,
          };
        }

        firstPageBlocks.push(secBlock);
      }
    }
  }

  const payload: StorefrontMenuPayload = {
    name: config.name,
    typography: config.typography || 'Inter',
    display_config: {
      dark: !!config.dark,
      colors: colorPalette,
      navbar: {
        dark: config.navbar_dark ?? null,
        bg_color: (config.navbar_bg as ItemStyleToken) || 'primary_color',
        border: config.navbar_border ?? false,
        border_style: config.navbar_border_style || 'none',
        border_color: config.navbar_border_color as DisplayToken | undefined,
      },
      menu: {
        container: config.menu_container || 'grid',
      },
      first_page: firstPageBlocks,
    },
  };

  const validationErrors = validateMenuPayload(payload);
  if (validationErrors.length > 0) {
    throw new Error(`Validação do layout falhou: ${validationErrors.join('; ')}`);
  }

  return payload;
}

// ==========================================
// Menu Items API (OpenAPI 2.2.16 Section 4)
// ==========================================

export interface ComboItemInput {
  item_id: string;
  quantity: number;
}

export interface MenuItemPayload {
  name: string;
  description?: string | null;
  category_id?: string | null;
  base_price_cents: number;
  effective_price_cents?: number;
  base_cost_cents?: number | null;
  image_url?: string | null;
  is_available?: boolean;
  is_combo?: boolean;
  combo_items?: ComboItemInput[];
  preparation_time_minutes?: number | null;
  display_order?: number;
  display_style?: ItemDisplayStyle | null;
}

export async function listMenuItems(slug: string): Promise<any[]> {
  try {
    const data = await restaurantRequest(slug, '/items');
    return Array.isArray(data) ? data : [];
  } catch {
    return [];
  }
}

export async function createMenuItem(slug: string, item: MenuItemPayload): Promise<any> {
  return restaurantRequest(slug, '/items', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(item),
  });
}

export async function updateMenuItem(slug: string, id: string, item: MenuItemPayload): Promise<any> {
  return restaurantRequest(slug, `/items/${encodeURIComponent(id)}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(item),
  });
}

export async function patchMenuItem(slug: string, id: string, patch: Partial<MenuItemPayload>): Promise<any> {
  return restaurantRequest(slug, `/items/${encodeURIComponent(id)}`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(patch),
  });
}

export async function deleteMenuItem(slug: string, id: string): Promise<any> {
  return restaurantRequest(slug, `/items/${encodeURIComponent(id)}`, {
    method: 'DELETE',
  });
}
