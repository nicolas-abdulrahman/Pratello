<script lang="ts">
  import { DollarSign, Receipt, ShoppingBag, Trophy } from 'lucide-svelte';
  import { formatPrice, formatTime, isToday, type Order, type Product, type Category } from './admin';

  // --- Abstracted State & Dependencies ---
  export let orders: Order[] = [];
  export let products: Product[] = [];
  export let categories: Category[] = [
    { id: 'lanches', name: 'Lanches' },
    { id: 'pratos', name: 'Pratos Principais' },
    { id: 'bebidas', name: 'Bebidas' },
    { id: 'sobremesas', name: 'Sobremesas' },
  ];

  const statusStyle: Record<string, string> = {
    pendente: 'bg-zinc-100 text-zinc-700 ring-zinc-300',
    preparo: 'bg-yellow-100 text-yellow-800 ring-yellow-300',
    pronto: 'bg-green-100 text-green-800 ring-green-300',
    entregue: 'bg-zinc-800 text-zinc-100 ring-zinc-700',
  };

  const statusLabel: Record<string, string> = {
    pendente: 'Pendente',
    preparo: 'Em Preparo',
    pronto: 'Pronto',
    entregue: 'Entregue',
  };

  $: todayOrders = (orders || []).filter((o) => isToday(o.createdAt));
  $: revenue = todayOrders.reduce((s, o) => s + (o.total || 0), 0);
  $: count = todayOrders.length;
  $: ticket = count ? revenue / count : 0;

  $: stats = (() => {
    const qtyByProduct = new Map<string, { name: string; qty: number }>();
    const revenueByCategory = new Map<string, number>();

    for (const o of todayOrders) {
      for (const it of o.items || []) {
        const cur = qtyByProduct.get(it.productId) ?? { name: it.name, qty: 0 };
        cur.qty += it.quantity;
        qtyByProduct.set(it.productId, cur);

        const cat = products.find((p) => p.id === it.productId)?.category ?? 'outros';
        revenueByCategory.set(cat, (revenueByCategory.get(cat) ?? 0) + it.price * it.quantity);
      }
    }
    const ranking = [...qtyByProduct.values()].sort((a, b) => b.qty - a.qty);
    return {
      ranking: ranking.slice(0, 5),
      top: ranking[0],
      revenueByCategory,
    };
  })();

  $: cards = [
    { label: 'Faturamento do dia', value: formatPrice(revenue), icon: DollarSign },
    { label: 'Total de pedidos', value: String(count), icon: ShoppingBag },
    { label: 'Ticket médio', value: formatPrice(ticket), icon: Receipt },
    {
      label: 'Prato mais vendido',
      value: stats.top?.name ?? '—',
      sub: stats.top ? `${stats.top.qty} unidades` : undefined,
      icon: Trophy,
    },
  ];

  $: maxRank = stats.ranking[0]?.qty ?? 1;
  $: maxCat = Math.max(1, ...Array.from(stats.revenueByCategory.values()));
  $: todayFormatted = new Date().toLocaleDateString('pt-BR', { day: '2-digit', month: 'long' });
  $: recentOrders = [...todayOrders].sort((a, b) => b.createdAt - a.createdAt).slice(0, 6);
</script>

<div>
  <h1 class="text-2xl font-bold text-zinc-900">Dashboard</h1>
  <p class="text-sm text-zinc-600">Resumo das vendas de hoje, {todayFormatted}.</p>

  <!-- Metric Cards -->
  <div class="mt-6 grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
    {#each cards as { label, value, sub, icon: Icon }}
      <div class="rounded-2xl border border-zinc-200 bg-white p-5 shadow-sm">
        <div class="flex items-center justify-between">
          <p class="text-sm font-medium text-zinc-600">{label}</p>
          <span class="grid h-9 w-9 place-items-center rounded-xl bg-yellow-400 text-zinc-900">
            <svelte:component this={Icon} size={18} />
          </span>
        </div>
        <p class="mt-3 truncate text-2xl font-extrabold text-zinc-900" title={value}>
          {value}
        </p>
        {#if sub}
          <p class="text-xs text-zinc-500">{sub}</p>
        {/if}
      </div>
    {/each}
  </div>

  <div class="mt-6 grid gap-4 xl:grid-cols-3">
    <!-- Mais vendidos -->
    <section class="rounded-2xl border border-zinc-200 bg-white p-5 shadow-sm">
      <h2 class="font-bold text-zinc-900">Mais vendidos hoje</h2>
      <ul class="mt-4 space-y-3">
        {#if stats.ranking.length === 0}
          <li class="text-sm text-zinc-500">Sem vendas ainda.</li>
        {/if}
        {#each stats.ranking as r (r.name)}
          <li>
            <div class="flex justify-between text-sm">
              <span class="truncate text-zinc-700">{r.name}</span>
              <span class="font-semibold text-zinc-900">{r.qty}</span>
            </div>
            <div class="mt-1 h-2 rounded-full bg-zinc-100">
              <div class="h-2 rounded-full bg-yellow-400" style={`width: ${(r.qty / maxRank) * 100}%`} />
            </div>
          </li>
        {/each}
      </ul>
    </section>

    <!-- Faturamento por categoria -->
    <section class="rounded-2xl border border-zinc-200 bg-white p-5 shadow-sm">
      <h2 class="font-bold text-zinc-900">Faturamento por categoria</h2>
      <ul class="mt-4 space-y-3">
        {#each categories as c (c.id)}
          {@const v = stats.revenueByCategory.get(c.id) ?? 0}
          <li>
            <div class="flex justify-between text-sm">
              <span class="text-zinc-700">{c.name}</span>
              <span class="font-semibold text-zinc-900">{formatPrice(v)}</span>
            </div>
            <div class="mt-1 h-2 rounded-full bg-zinc-100">
              <div class="h-2 rounded-full bg-zinc-800" style={`width: ${(v / maxCat) * 100}%`} />
            </div>
          </li>
        {/each}
      </ul>
    </section>

    <!-- Últimos pedidos -->
    <section class="rounded-2xl border border-zinc-200 bg-white p-5 shadow-sm">
      <h2 class="font-bold text-zinc-900">Últimos pedidos</h2>
      <ul class="mt-4 divide-y divide-zinc-100">
        {#if recentOrders.length === 0}
          <li class="py-2.5 text-sm text-zinc-500">Nenhum pedido hoje ainda.</li>
        {/if}
        {#each recentOrders as o (o.id)}
          <li class="flex items-center justify-between gap-2 py-2.5 text-sm">
            <div class="min-w-0">
              <p class="font-semibold text-zinc-900">
                #{o.number} · <span class="font-normal text-zinc-600">{o.customerName}</span>
              </p>
              <p class="text-xs text-zinc-500">
                {formatTime(o.createdAt)} · {formatPrice(o.total)}
              </p>
            </div>
            <span
              class={`inline-flex rounded-full px-2.5 py-0.5 text-xs font-semibold ring-1 ring-inset ${
                statusStyle[o.status] || 'bg-zinc-100 text-zinc-700 ring-zinc-300'
              }`}
            >
              {statusLabel[o.status] || o.status}
            </span>
          </li>
        {/each}
      </ul>
    </section>
  </div>
</div>
