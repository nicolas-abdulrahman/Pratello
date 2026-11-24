<script lang="ts">
  export type OrderStatus = 'pendente' | 'preparo' | 'pronto' | 'entregue' | string;

  interface Props {
    status: OrderStatus;
  }

  let { status }: Props = $props();

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
</script>

<span
  class={`inline-flex rounded-full px-2.5 py-0.5 text-xs font-semibold ring-1 ring-inset ${
    statusStyle[status] || 'bg-zinc-100 text-zinc-700 ring-zinc-300'
  }`}
>
  {statusLabel[status] || status}
</span>
