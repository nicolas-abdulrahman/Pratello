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
    const products: Product[] = testMode ? [
      { id: 'burger', name: 'João Burger', description: 'Hambúrguer artesanal com queijo e molho da casa.', price: 32.90, prepTime: 20, category: 'lanches', tags: ['destaque'], image: '', ingredients: ['Pão', 'Carne', 'Queijo'], available: true, popular: true },
      { id: 'batata', name: 'Batata frita', description: 'Porção crocante.', price: 18.90, prepTime: 15, category: 'lanches', tags: [], image: '', ingredients: ['Batata'], available: true, popular: false },
      { id: 'refrigerante', name: 'Refrigerante', description: 'Lata 350 ml.', price: 7, prepTime: 1, category: 'bebidas', tags: [], image: '', ingredients: [], available: true, popular: false },
    ] : [];
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
