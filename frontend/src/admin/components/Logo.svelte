<script lang="ts">
  import { Utensils } from 'lucide-svelte';

  interface Props {
    light?: boolean;
    className?: string;
  }

  let { light = false, className = '' }: Props = $props();
</script>

<span class={`inline-flex items-center gap-2 font-extrabold tracking-tight ${className}`}>
  <span class="grid h-9 w-9 place-items-center rounded-xl bg-yellow-400 text-zinc-900">
    <Utensils size={20} strokeWidth={2.5} />
  </span>
  <span class={`text-2xl ${light ? 'text-white' : 'text-zinc-900'}`}>Pratello</span>
</span>
