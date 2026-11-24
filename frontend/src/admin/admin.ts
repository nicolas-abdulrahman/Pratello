/**
 * Admin types, contracts and abstracted dependency interfaces.
 */

export type Role = 'admin' | 'funcionario';

export interface StaffUser {
  id: string;
  name: string;
  role: Role;
  login: string;
  password?: string;
}

export type CategoryId = 'lanches' | 'pratos' | 'bebidas' | 'sobremesas' | string;
export type TagId = string;

export interface Category {
  id: CategoryId;
  name: string;
  description?: string;
}

export interface VisualCategory {
  id: TagId;
  name: string;
  icon?: string;
}

export interface Product {
  id: string;
  name: string;
  description: string;
  price: number;
  originalPrice?: number;
  prepTime: number;
  category: CategoryId;
  tags: TagId[];
  image: string;
  ingredients: string[];
  available: boolean;
  popular: boolean;
  base_price_cents?: number;
  effective_price_cents?: number;
  base_cost_cents?: number | null;
  saleEnd?: string;
  sale_ends_at?: string;
  is_combo?: boolean;
  combo_items?: Array<{ item_id: string; quantity: number }>;
  display_style?: {
    bg?: string;
    title_color?: string;
    price_color?: string;
    dark?: boolean | null;
  } | null;
}

export interface ProductInput extends Omit<Product, 'id'> {
  id?: string;
}

export interface OrderItem {
  productId: string;
  name: string;
  price: number;
  quantity: number;
}

export interface Order {
  id: string;
  number: number;
  customerName: string;
  items: OrderItem[];
  total: number;
  status: 'pendente' | 'preparo' | 'pronto' | 'entregue';
  createdAt: number;
  [key: string]: any;
}

export interface AccentColor {
  id: string;
  name: string;
  shades: Record<200 | 400 | 600, string>;
}

export const DEFAULT_ACCENTS: AccentColor[] = [
  { id: 'yellow', name: 'Amarelo Mostarda', shades: { 200: '#fef08a', 400: '#facc15', 600: '#ca8a04' } },
  { id: 'orange', name: 'Laranja Fogo', shades: { 200: '#fed7aa', 400: '#fb923c', 600: '#ea580c' } },
  { id: 'red', name: 'Vermelho Brasa', shades: { 200: '#fecaca', 400: '#f87171', 600: '#dc2626' } },
  { id: 'emerald', name: 'Verde Manjericão', shades: { 200: '#a7f3d0', 400: '#34d399', 600: '#059669' } },
  { id: 'teal', name: 'Teal Marítimo', shades: { 200: '#99f6e4', 400: '#2dd4bf', 600: '#0d9488' } },
  { id: 'violet', name: 'Roxo Artesanal', shades: { 200: '#ddd6fe', 400: '#a78bfa', 600: '#7c3aed' } },
  { id: 'zinc', name: 'Preto & Grafite', shades: { 200: '#e4e4e7', 400: '#a1a1aa', 600: '#52525b' } },
];

export const formatPrice = (value: number) =>
  new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(value);

export const parsePrice = (raw: string) => {
  const s = raw.trim();
  if (!s) return NaN;
  return Number(s.includes(',') ? s.replace(/\./g, '').replace(',', '.') : s);
};

export const formatTime = (timestamp: number) =>
  new Date(timestamp).toLocaleTimeString('pt-BR', { hour: '2-digit', minute: '2-digit' });

export const isToday = (timestamp: number) =>
  new Date(timestamp).toDateString() === new Date().toDateString();

export const discountPercent = (price: number, original?: number) =>
  original && original > price ? Math.round((1 - price / original) * 100) : 0;

/**
 * Abstracted Admin State & Dependencies
 * (To be implemented/filled by consumer)
 */
export interface AdminDependencies {
  // Authentication & session
  session?: { name?: string; role?: Role; userId?: string } | null;
  logout?: () => void;
  resetDemoData?: () => void;

  // Product management
  getProducts?: () => Promise<Product[]>;
  saveProduct?: (product: ProductInput) => Promise<void>;
  deleteProduct?: (productId: string) => Promise<void>;
  toggleProductAvailability?: (productId: string) => Promise<void>;

  // Category management
  getCategories?: () => Promise<Category[]>;
  getVisualCategories?: () => Promise<VisualCategory[]>;

  // Staff management
  getStaffUsers?: () => Promise<StaffUser[]>;
  saveStaffUser?: (user: StaffUser) => Promise<string | undefined>;
  deleteStaffUser?: (userId: string) => Promise<string | undefined>;

  // Order management & stats
  getOrders?: () => Promise<Order[]>;

  // Theme / appearance
  getAccent?: () => string;
  setAccent?: (accentId: string) => void;
}

import type { Order as PanelOrder } from './components/OrderItem.svelte';

// This is the single boundary for future restaurant-scoped admin API calls.
export function createAdminData(_slug: string) {
  const testMode = import.meta.env.TEST_ENV === 'true' || import.meta.env.TEST_ENV === '1';
  function initialData() {
    const defaultProducts: Product[] = [
      {
        id: 'beef-burger',
        name: 'Classic Beef Burger',
        description: 'Hambúrguer de carne bovina artesanal, queijo cheddar, alface e molho especial.',
        price: 32.90,
        prepTime: 20,
        category: 'lanches',
        tags: ['destaque'],
        image: 'https://media.istockphoto.com/id/1302436326/photo/junk-food-homemade-beef-burgers-on-vintage-wooden-background.jpg?s=612x612&w=0&k=20&c=NsyDE31unoNd80wGfrkMOqvsnjeNOpHER-yL_8KwcRw=',
        ingredients: ['Pão Brioche', 'Blend Bovino 180g', 'Queijo Cheddar', 'Alface', 'Molho Especial'],
        available: true,
        popular: true,
        base_price_cents: 3290,
        effective_price_cents: 3290,
      },
      {
        id: 'bacon-cheeseburger',
        name: 'Bacon Cheeseburger',
        description: 'Pão tostado, hambúrguer suculento, queijo derretido e tiras crocantes de bacon.',
        price: 36.90,
        originalPrice: 42.00,
        prepTime: 20,
        category: 'lanches',
        tags: ['destaque', 'promo'],
        image: 'https://media.istockphoto.com/id/1398630614/photo/bacon-cheeseburger-on-a-toasted-bun.jpg?s=612x612&w=0&k=20&c=3HWrUVnS-FsJETFlCWnDH2-1ekJ0ic3T3XPrhLjpo98=',
        ingredients: ['Pão Brioche', 'Blend Bovino 180g', 'Queijo Cheddar', 'Bacon Crocante'],
        available: true,
        popular: true,
        base_price_cents: 4200,
        effective_price_cents: 3690,
      },
      {
        id: 'burger-galinha-ultra',
        name: 'Burger Galinha Ultra (Crispy Chicken Sandwich)',
        description: 'Sanduíche de frango super crocante temperado ao estilo sulista americano com picles e maionese.',
        price: 29.90,
        prepTime: 18,
        category: 'lanches',
        tags: [],
        image: 'https://media.istockphoto.com/id/521207406/photo/southern-country-fried-chicken-sandwich.jpg?s=612x612&w=0&k=20&c=XfttLSxEO2YAjop4Gy0CIb1L5N_OK1tTKxmkiPr3QD8=',
        ingredients: ['Pão com Gergelim', 'Sobrecoxa Empanada Crocante', 'Picles', 'Maionese Especial'],
        available: true,
        popular: true,
        base_price_cents: 2990,
        effective_price_cents: 2990,
      },
      {
        id: 'seasoned-fries',
        name: 'Golden Seasoned Fries',
        description: 'Batatas fritas douradas e crocantes servidas com ervas finas e tempero especial da casa.',
        price: 18.90,
        prepTime: 12,
        category: 'lanches',
        tags: [],
        image: 'https://media.istockphoto.com/id/2227232722/photo/golden-crispy-seasoned-french-fries-served-on-a-wooden-tray-sprinkled-with-herbs-and-spices.jpg?s=612x612&w=0&k=20&c=adoolW_zI9RYPuvaX9WK9dTHKkGXJJNDCqIIfw9A3nE=',
        ingredients: ['Batatas Especiais', 'Mix de Ervas', 'Sal Marinho'],
        available: true,
        popular: true,
        base_price_cents: 1890,
        effective_price_cents: 1890,
      },
      {
        id: 'onion-rings',
        name: 'Crispy Onion Rings',
        description: 'Anéis de cebola empanados e fritos na hora com textura ultra crocante.',
        price: 21.90,
        prepTime: 15,
        category: 'lanches',
        tags: [],
        image: 'https://media.istockphoto.com/id/2188707691/photo/homemade-onion-rings.jpg?s=612x612&w=0&k=20&c=CZLCSANsHGsIlzgOl6iX-BMnwT241l4HYCTes1CIiVg=',
        ingredients: ['Cebola Doce', 'Empanamento Crocante', 'Molho Barbecue'],
        available: true,
        popular: false,
        base_price_cents: 2190,
        effective_price_cents: 2190,
      },
      {
        id: 'pepperoni-pizza',
        name: 'Pepperoni Pizza',
        description: 'Pizza quentinha com fatias generosas de pepperoni defumado e queijo mozzarella derretido.',
        price: 49.90,
        originalPrice: 59.90,
        prepTime: 25,
        category: 'pratos',
        tags: ['destaque', 'promo'],
        image: 'https://media.istockphoto.com/id/521403691/pt/foto/quente-feito-em-casa-pizza-de-pepperoni.jpg?s=612x612&w=0&k=20&c=KWGroPbutn4ruhgedMeQ6vvPhFULRrnCMtV5coxr1AI=',
        ingredients: ['Massa de Fermentação Natural', 'Molho de Tomate', 'Mozzarella', 'Pepperoni'],
        available: true,
        popular: true,
        base_price_cents: 5990,
        effective_price_cents: 4990,
      },
      {
        id: 'margherita-pizza',
        name: 'Margherita Basil Pizza',
        description: 'Clássica pizza napolitana com queijo fior di latte, molho de tomate fresco e folhas de manjericão.',
        price: 44.90,
        prepTime: 25,
        category: 'pratos',
        tags: [],
        image: 'https://cdn.loveandlemons.com/wp-content/uploads/2023/07/margherita-pizza.jpg',
        ingredients: ['Massa Artesanal', 'Molho de Tomate Pelati', 'Mozzarella', 'Manjericão Fresco'],
        available: true,
        popular: false,
        base_price_cents: 4490,
        effective_price_cents: 4490,
      },
      {
        id: 'cola-soda',
        name: 'Ice Cold Cola Soda',
        description: 'Copo de refrigerante de cola super refrescante servido com gelo.',
        price: 8.00,
        prepTime: 3,
        category: 'bebidas',
        tags: [],
        image: 'https://media.istockphoto.com/id/530428650/photo/cola-glass-with-ice-cubes.jpg?s=612x612&w=0&k=20&c=keqH2KNWHO1sFxtsBfx5EZjyep1CRBHIqwe_ZwxszHc=',
        ingredients: ['Cola', 'Gelo', 'Limão opcional'],
        available: true,
        popular: true,
        base_price_cents: 800,
        effective_price_cents: 800,
      },
      {
        id: 'chocolate-milkshake',
        name: 'Chocolate Milkshake',
        description: 'Milkshake cremoso de chocolate com calda especial e chantilly.',
        price: 19.90,
        prepTime: 8,
        category: 'bebidas',
        tags: ['destaque'],
        image: 'https://media.istockphoto.com/id/1291175021/photo/chocolate-milkshake.jpg?s=612x612&w=0&k=20&c=LkPVfYUcIdM85m9FI6QrAjXyzvOxrdhBEkmcPmlbTYw=',
        ingredients: ['Sorvete de Chocolate', 'Leite Integral', 'Calda de Chocolate', 'Chantilly'],
        available: true,
        popular: true,
        base_price_cents: 1990,
        effective_price_cents: 1990,
      },
      {
        id: 'iced-caramel-latte',
        name: 'Iced Caramel Latte',
        description: 'Café gelado suave com leite cremoso e xarope de caramelo artesanal.',
        price: 16.50,
        prepTime: 5,
        category: 'bebidas',
        tags: [],
        image: 'https://media.istockphoto.com/id/1366672341/photo/glass-with-ice-and-coffee.jpg?s=612x612&w=0&k=20&c=lR7zj__gNDmO0aXfjV8N9Ti-WnOx2iuNxRRyEDvFCi4=',
        ingredients: ['Café Expresso', 'Leite Gelado', 'Calda de Caramelo', 'Gelo'],
        available: true,
        popular: false,
        base_price_cents: 1650,
        effective_price_cents: 1650,
      },
      {
        id: 'classic-hot-dog',
        name: 'Classic Hot Dog',
        description: 'Cachorro quente clássico americano com salsicha grelhada, ketchup e mostarda amarela.',
        price: 22.00,
        prepTime: 12,
        category: 'lanches',
        tags: [],
        image: 'https://media.istockphoto.com/id/1313385617/photo/hot-dog-with-ketchup-and-yellow-mustard.jpg?s=612x612&w=0&k=20&c=MgOEout7LUn9yD354y_WKwLXrh0hB-_S24KR5gwkrU0=',
        ingredients: ['Pão de Hot Dog', 'Salsicha Bovina', 'Mostarda', 'Ketchup'],
        available: true,
        popular: false,
        base_price_cents: 2200,
        effective_price_cents: 2200,
      },
      {
        id: 'chicken-wings',
        name: 'Sweet Chili Chicken Wings',
        description: 'Asinhas de frango crocantes glaceadas em molho sweet chili com gergelim torrado.',
        price: 34.90,
        prepTime: 20,
        category: 'lanches',
        tags: ['destaque'],
        image: 'https://media.istockphoto.com/id/835903320/photo/baked-chicken-wings-with-sesame-seeds-and-sweet-chili-sauce-on-white-wooden-board.jpg?s=612x612&w=0&k=20&c=SH8ZCkEKuWD_wxulpntIJ0uD4yRnUf9UXovQwSwrmmA=',
        ingredients: ['Asas de Frango', 'Molho Sweet Chili', 'Gergelim'],
        available: true,
        popular: true,
        base_price_cents: 3490,
        effective_price_cents: 3490,
      },
      {
        id: 'mozzarella-sticks',
        name: 'Fried Mozzarella Sticks',
        description: 'Palitos crocantes de mozzarella empanada servidos com molho marinara caseiro.',
        price: 26.90,
        prepTime: 15,
        category: 'lanches',
        tags: [],
        image: 'https://media.istockphoto.com/id/1315418985/photo/mozzarella-sticks.jpg?s=612x612&w=0&k=20&c=uM3OdjYu1jGSXj8DrUXpo9zNM0RuNd4i0NMlw8SWrRA=',
        ingredients: ['Queijo Mozzarella', 'Farinha Panko', 'Ervas', 'Molho Marinara'],
        available: true,
        popular: false,
        base_price_cents: 2690,
        effective_price_cents: 2690,
      },
      {
        id: 'loaded-nachos',
        name: 'Loaded Cheese Nachos',
        description: 'Tortilhas crocantes cobertas com molho de queijo cheddar cremoso, jalapeños e salsa fresca.',
        price: 38.00,
        prepTime: 15,
        category: 'lanches',
        tags: [],
        image: 'https://media.istockphoto.com/id/157642304/photo/nachos-with-cheese-sauce-jalapenos-and-salsa.jpg?s=612x612&w=0&k=20&c=sEowpZb83ccV0tTM_1yyzHL4CKEkSuMcb5g__eLSgHQ=',
        ingredients: ['Totopos de Milho', 'Cheddar Cremoso', 'Jalapeños', 'Salsa Mexicana'],
        available: true,
        popular: true,
        base_price_cents: 3800,
        effective_price_cents: 3800,
      },
      {
        id: 'birria-tacos',
        name: 'Birria Tacos Platter',
        description: 'Tacos dourados recheados com carne desfiada suculenta, queijo derretido e consomê para mergulhar.',
        price: 42.00,
        prepTime: 22,
        category: 'pratos',
        tags: ['destaque'],
        image: 'https://cdn.babyseo.ai/images/foodshot.ai/mexican-food-photography/mexican-food-photography-hero-spread-tacos-birria-salsa.webp',
        ingredients: ['Tortilhas de Milho', 'Carne Bovina Desfiada', 'Queijo Oaxaca', 'Consomê'],
        available: true,
        popular: true,
        base_price_cents: 4200,
        effective_price_cents: 4200,
      },
      {
        id: 'chicken-burrito-bowl',
        name: 'Chicken Burrito Bowl',
        description: 'Bowl completo com arroz temperado, feijão preto, frango grelhado, milho, abacate e molho chipotle.',
        price: 35.00,
        prepTime: 18,
        category: 'pratos',
        tags: [],
        image: 'https://media.istockphoto.com/id/1300212146/photo/homemade-healthy-chicken-burrito-bowl.jpg?s=612x612&w=0&k=20&c=IYNxJJU-fuZgmVhzauujV2Kt3S215HUua5RD7jwv6y8=',
        ingredients: ['Arroz com Coentro', 'Feijão Preto', 'Frango Grelhado', 'Abacate', 'Chipotle'],
        available: true,
        popular: false,
        base_price_cents: 3500,
        effective_price_cents: 3500,
      },
      {
        id: 'sushi-roll-platter',
        name: 'Sushi Roll Platter',
        description: 'Combinado especial com uramakis, hossomakis e nigiris frescos preparados pelo sushiman.',
        price: 58.00,
        prepTime: 25,
        category: 'pratos',
        tags: [],
        image: 'https://plus.unsplash.com/premium_photo-1668146927669-f2edf6e86f6f?fm=jpg&q=60&w=3000&auto=format&fit=crop&ixlib=rb-4.1.0&ixid=NHwxMjA3fDA%3D',
        ingredients: ['Salmão Fresco', 'Arroz Japonês', 'Nori', 'Cream Cheese'],
        available: true,
        popular: true,
        base_price_cents: 5800,
        effective_price_cents: 5800,
      },
      {
        id: 'chicken-caesar-salad',
        name: 'Grilled Chicken Caesar Salad',
        description: 'Folhas crocantes de alface romana, peito de frango grelhado, croutons dourados e queijo parmesão.',
        price: 31.00,
        prepTime: 15,
        category: 'pratos',
        tags: [],
        image: 'https://www.spendwithpennies.com/wp-content/uploads/2023/06/Grilled-Chicken-Caesar-Salad-SpendWithPennies-4.jpg',
        ingredients: ['Alface Romana', 'Peito de Frango', 'Croutons', 'Parmesão', 'Molho Caesar'],
        available: true,
        popular: false,
        base_price_cents: 3100,
        effective_price_cents: 3100,
      },
      {
        id: 'glazed-donuts',
        name: 'Glazed Donuts Assortment',
        description: 'Seleção sortida de donuts artesanais fofinhos com coberturas glaceadas e confeitos.',
        price: 24.00,
        prepTime: 5,
        category: 'sobremesas',
        tags: [],
        image: 'https://cdn.babyseo.ai/images/foodshot.ai/donut-photography/donut-photography-hero-assortment-raking-light.webp',
        ingredients: ['Massa Fermentada', 'Glaçagem de Baunilha', 'Chocolate', 'Confeitos'],
        available: true,
        popular: true,
        base_price_cents: 2400,
        effective_price_cents: 2400,
      },
      {
        id: 'banana-split',
        name: 'Banana Split Sundae',
        description: 'Sobremesa clássica com banana fresca fatiada, bolas de sorvete, calda quente, chantilly e castanhas.',
        price: 22.50,
        prepTime: 8,
        category: 'sobremesas',
        tags: ['destaque'],
        image: 'https://media.istockphoto.com/id/2033937470/photo/banana-split-with-cherries-and-nuts.jpg?s=612x612&w=0&k=20&c=X9-I6JhDMlxAhqt8TC5iZ7W54NcPzqbHjNx-O-Ooolo=',
        ingredients: ['Banana', 'Sorvete Napolitano', 'Calda de Chocolate', 'Chantilly', 'Cereja'],
        available: true,
        popular: false,
        base_price_cents: 2250,
        effective_price_cents: 2250,
      },
    ];
    const products: Product[] = defaultProducts;
    const staff: StaffUser[] = testMode ? [
      { id: 'joao', name: 'João Silva', login: 'joao', role: 'admin' },
      { id: 'maria', name: 'Maria Santos', login: 'maria', role: 'funcionario' },
    ] : [];
    const orders: PanelOrder[] = testMode ? [
      { id: 'order-1', number: 101, customerName: 'Ana Souza', location: { mode: 'mesa', table: '4' }, items: [{ productId: 'burger', name: 'João Burger', price: 32.90, quantity: 2 }], subtotal: 65.80, deliveryFee: 0, total: 65.80, payment: { method: 'pix', status: 'pago' }, status: 'preparo', createdAt: Date.now() - 10 * 60000 },
      { id: 'order-2', number: 102, customerName: 'Pedro Lima', location: { mode: 'delivery', address: { street: 'Rua das Flores', number: '123', district: 'Centro', cep: '50000-000' } }, items: [{ productId: 'burger', name: 'João Burger', price: 32.90, quantity: 1 }, { productId: 'batata', name: 'Batata frita', price: 18.90, quantity: 1 }], subtotal: 51.80, deliveryFee: 5, total: 56.80, payment: { method: 'cartao', status: 'pendente' }, status: 'pendente', createdAt: Date.now() - 5 * 60000 },
    ] : [];
    return { products, staff, orders };
  }
  return {
    testMode,
    initialData,
    saveProduct: (products: Product[], product: ProductInput) => {
      if (!testMode) return products;
      const saved = { ...product, id: product.id ?? `product-${Date.now()}` };
      return [...products.filter((p) => p.id !== saved.id), saved];
    },
    deleteProduct: (products: Product[], id: string) => testMode ? products.filter((p) => p.id !== id) : products,
    toggleProduct: (products: Product[], id: string) => testMode ? products.map((p) => p.id === id ? { ...p, available: !p.available } : p) : products,
    saveStaff: (staff: StaffUser[], user: StaffUser) => testMode ? [...staff.filter((u) => u.id !== user.id), user] : staff,
    deleteStaff: (staff: StaffUser[], id: string) => testMode ? staff.filter((u) => u.id !== id) : staff,
    advanceOrder: (orders: PanelOrder[], id: string) => {
      const next = { pendente: 'preparo', preparo: 'pronto', pronto: 'entregue', entregue: 'entregue' } as const;
      return testMode ? orders.map((o) => o.id === id ? { ...o, status: next[o.status] } : o) : orders;
    },
    markPaid: (orders: PanelOrder[], id: string): PanelOrder[] =>
      testMode ? orders.map((o) => o.id === id ? { ...o, payment: { ...o.payment, status: 'pago' } } : o) : orders,
  };
}
