<script lang="ts">
  import OrderItem, { type Order } from './OrderItem.svelte';

  interface Props {
    order: Order;
    onAdvance?: () => void;
    onMarkPaid?: () => void;
  }

  let { order, onAdvance, onMarkPaid }: Props = $props();
</script>

<OrderItem {order} {onAdvance} {onMarkPaid} />
