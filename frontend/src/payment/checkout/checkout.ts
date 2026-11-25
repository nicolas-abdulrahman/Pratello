/**
 * Checkout logic, calculation rules, coupon checks, and dependency contracts.
 */

export type CheckoutMode = 'mesa' | 'delivery' | 'whatsapp';
export type ShippingId = 'padrao' | 'expressa';
export type PaymentMethod = 'cartao' | 'pix' | 'dinheiro';

export interface Address {
  cep: string;
  street: string;
  number: string;
  district: string;
  complement?: string;
}

export type ServiceLocation =
  | { mode: 'mesa'; table: string }
  | { mode: 'delivery'; address: Address };

export interface Product {
  id: string;
  name: string;
  price: number;
  originalPrice?: number;
  image?: string;
  [key: string]: any;
}

export interface CartLine {
  product: Product;
  quantity: number;
}

export interface PaymentDraft {
  method: PaymentMethod | null;
  cardType: 'credito' | 'debito';
  installments: number;
  needsMachine: boolean;
  changeFor: string;
  pixPaid: boolean;
}

export interface ShippingOption {
  id: ShippingId;
  name: string;
  eta: string;
  fee: number;
  freeByMinimum: boolean;
}

export interface Coupon {
  code: string;
  kind: 'desconto' | 'frete';
  title: string;
  description: string;
  minSubtotal: number;
  percent?: number;
  amount?: number;
  maxDiscount?: number;
  modes?: CheckoutMode[];
  payment?: PaymentMethod;
  validUntil: string;
}

export const FREE_DELIVERY_FROM = 80;

export const COUPONS: Coupon[] = [
  { code: 'FRETEGRATIS', kind: 'frete', title: 'Frete grátis', description: 'Até R$ 12 de desconto no frete', minSubtotal: 30, amount: 12, modes: ['delivery', 'whatsapp'], validUntil: '30/11' },
  { code: 'FRETE50', kind: 'frete', title: '50% off no frete', description: 'Metade do frete, sem valor mínimo', minSubtotal: 0, percent: 50, modes: ['delivery', 'whatsapp'], validUntil: '15/12' },
  { code: 'BEMVINDO10', kind: 'desconto', title: '10% OFF', description: 'Desconto máximo de R$ 20', minSubtotal: 40, percent: 10, maxDiscount: 20, validUntil: '31/12' },
  { code: 'PRATELLO15', kind: 'desconto', title: 'R$ 15 OFF', description: 'Em pedidos acima de R$ 80', minSubtotal: 80, amount: 15, validUntil: '10/12' },
  { code: 'PIX7', kind: 'desconto', title: '7% OFF no Pix', description: 'Pagando com Pix · máx. R$ 15', minSubtotal: 25, percent: 7, maxDiscount: 15, payment: 'pix', validUntil: '31/12' },
  { code: 'MESA5', kind: 'desconto', title: 'R$ 5 OFF na mesa', description: 'Só para pedidos feitos no salão', minSubtotal: 30, amount: 5, modes: ['mesa'], validUntil: '31/12' },
  { code: 'ZAP10', kind: 'desconto', title: 'R$ 10 OFF no WhatsApp', description: 'Pedidos enviados pelo WhatsApp', minSubtotal: 50, amount: 10, modes: ['whatsapp'], validUntil: '20/12' },
];

export const findCoupon = (code: string) => COUPONS.find((c) => c.code === code.trim().toUpperCase());

export function distanceFromCep(cep: string): number {
  const d = cep.replace(/\D/g, '');
  if (d.length !== 8) return 0;
  const hash = [...d].reduce((h, c) => (h * 31 + Number(c)) % 997, 7);
  return Math.round((1.5 + (hash % 81) / 10) * 10) / 10;
}

export function shippingOptions(km: number, subtotal: number): ShippingOption[] {
  const base = Math.round((4.9 + km * 0.9) * 100) / 100;
  const freeStd = subtotal >= FREE_DELIVERY_FROM;
  const extra = Math.ceil(km * 2);
  return [
    { id: 'padrao', name: 'Padrão', eta: `${35 + extra}-${50 + extra} min`, fee: freeStd ? 0 : base, freeByMinimum: freeStd },
    { id: 'expressa', name: 'Expressa', eta: `${18 + extra}-${28 + extra} min`, fee: Math.round((base + 6) * 100) / 100, freeByMinimum: false },
  ];
}

const MODE_LABEL: Record<CheckoutMode, string> = { mesa: 'na mesa', delivery: 'no delivery', whatsapp: 'no WhatsApp' };
const PAY_LABEL: Record<PaymentMethod, string> = { pix: 'Pix', cartao: 'cartão', dinheiro: 'dinheiro' };

export interface CouponContext {
  subtotal: number;
  mode: CheckoutMode;
  payment: PaymentMethod | null;
}

export function couponProblem(c: Coupon, ctx: CouponContext): string | null {
  if (c.modes && !c.modes.includes(ctx.mode)) return `Válido apenas ${c.modes.map((m) => MODE_LABEL[m]).join(' ou ')}`;
  if (c.payment && ctx.payment !== c.payment) return `Pague com ${PAY_LABEL[c.payment]} para usar`;
  if (ctx.subtotal < c.minSubtotal) {
    const diff = new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(c.minSubtotal - ctx.subtotal);
    return `Faltam ${diff} para usar`;
  }
  return null;
}

function couponValue(c: Coupon, base: number) {
  let v = c.percent ? (base * c.percent) / 100 : c.amount ?? 0;
  if (c.maxDiscount) v = Math.min(v, c.maxDiscount);
  return Math.round(Math.min(v, base) * 100) / 100;
}

export interface CheckoutInput extends CouponContext {
  cep: string;
  shipping: ShippingId;
  couponCodes: string[];
}

export interface CheckoutSummary {
  subtotal: number;
  km: number;
  options: ShippingOption[];
  shippingFee: number;
  shippingDiscount: number;
  discount: number;
  total: number;
  savings: number;
  applied: Coupon[];
  invalid: { coupon: Coupon; reason: string }[];
}

export function computeCheckout(i: CheckoutInput): CheckoutSummary {
  const isShipped = i.mode !== 'mesa';
  const km = isShipped ? distanceFromCep(i.cep) : 0;
  const options = isShipped && km > 0 ? shippingOptions(km, i.subtotal) : [];
  const option = options.find((o) => o.id === i.shipping) ?? options[0];
  const shippingFee = option?.fee ?? 0;

  const applied: Coupon[] = [];
  const invalid: CheckoutSummary['invalid'] = [];
  for (const code of i.couponCodes) {
    const c = findCoupon(code);
    if (!c) continue;
    const problem = couponProblem(c, i);
    if (problem) invalid.push({ coupon: c, reason: problem });
    else applied.push(c);
  }

  const productCoupon = applied.find((c) => c.kind === 'desconto');
  const shippingCoupon = applied.find((c) => c.kind === 'frete');
  const discount = productCoupon ? couponValue(productCoupon, i.subtotal) : 0;
  const shippingDiscount = shippingCoupon ? couponValue(shippingCoupon, shippingFee) : 0;
  const total = Math.max(0, Math.round((i.subtotal - discount + shippingFee - shippingDiscount) * 100) / 100);
  const freeByMin = option?.freeByMinimum ? shippingOptions(km, 0)[0].fee : 0;

  return {
    subtotal: i.subtotal,
    km,
    options,
    shippingFee,
    shippingDiscount,
    discount,
    total,
    savings: discount + shippingDiscount + freeByMin,
    applied,
    invalid,
  };
}

export const maxInstallments = (total: number) => Math.max(1, Math.min(3, Math.floor(total / 15)));

export const formatPrice = (value: number) =>
  new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(value);

/**
 * Abstracted Restaurant & Checkout Dependency Contract
 * (To be implemented/filled by consumer)
 */
export interface RestaurantDependencies {
  getRestaurantInfo: () => Promise<{ name: string; pixKey?: string; phone?: string; tableCount?: number }>;
  placeOrder: (payload: {
    location: ServiceLocation;
    channel: 'app' | 'whatsapp';
    phone: string;
    notes: string;
    payment: any;
    shipping: ShippingId;
    couponCodes: string[];
  }) => Promise<{ id: string; number: number } | null>;
  navigateTo: (route: string, state?: any) => void;
  fetchAddressFromCep?: (cep: string) => Promise<{ street?: string; district?: string } | null>;
}
