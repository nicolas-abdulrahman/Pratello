<script>
  export let value = '';
  export let label = '';
</script>

<div class="flex items-center gap-3 px-2 pt-2 md:pt-0">
  <div class="p-2.5 bg-gray-50 text-gray-600 rounded-xl flex items-center justify-center">
    <slot name="icon" />
  </div>
  <div>
    <div class="text-xs font-bold text-gray-900">{value}</div>
    <div class="text-[11px] text-gray-500">{label}</div>
  </div>
</div>
