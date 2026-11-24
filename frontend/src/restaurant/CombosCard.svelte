<script>
  import { add_item} from '../lib/stores/Cart';
  import ReviewProp from '../review/ReviewProp.svelte';
  import { sectionText } from './section/section';

  let {
    image,
    title,
    desc,
    price,
    originalPrice,
    combo = null,
    rating = null,
    count = null,
    colors,
    onOpenDetails = () => {}
  } = $props();

  let isPopping = $state(false);
  let justAdded = $state(false);

  function parsePrice(value) {
    if (typeof value === 'number') return value;
    if (!value) return 0;
    return Number(String(value).replace(/[^\d,.-]/g, '').replace(/\./g, '').replace(',', '.')) || 0;
  }

  function handleAdd(e) {
    e.stopPropagation(); // Prevents the modal from opening
    if (isPopping) return;
    isPopping = true;
    justAdded = true;

    add_item({
      name: title,
      icon: image,
      price: parsePrice(price),
      combo: combo,
    });

    setTimeout(() => { isPopping = false; }, 400); 
    setTimeout(() => { justAdded = false; }, 1200); 
  }
</script>

<!-- Using a native button for the card wrapper to completely avoid a11y errors -->
<button 
  type="button"
  onclick={onOpenDetails}
  style:background-color={colors.background}
  class="text-left rounded-2xl overflow-hidden shadow-sm border border-gray-100 hover:shadow-xl transition-all duration-300 group flex flex-col w-full h-full cursor-pointer outline-none focus-visible:ring-2 focus-visible:ring-[#ff5722]
  {isPopping ? 'animate-card-pop' : ''}"
>
  <div class="relative h-48 overflow-hidden shrink-0 w-full">
    <img src={image} alt={title} class="w-full h-full object-cover group-hover:scale-110 transition-transform duration-500" />
  </div>

  <div class="p-5 flex flex-col flex-1 w-full">
    <h3 class="text-lg font-extrabold mb-1" style:color={colors.title}>{title?.trim() || 'Combo'}</h3>
    {#if desc}
      <p class="text-sm leading-snug mb-4 flex-1 line-clamp-2" style:color={colors.description}>{desc}</p>
    {/if}
    
    <div class="flex items-end justify-between mt-auto pt-4 border-t border-gray-50 w-full">
      <div class="flex flex-col gap-1 items-start">
        <ReviewProp {rating} {count} />

        {#if originalPrice}
          <span class="text-xs text-gray-400 line-through font-medium leading-none mb-1">{originalPrice}</span>
        {/if}
        <span class="text-xl font-black leading-none" style:color={colors.price}>{price?.trim() || ''}</span>
      </div>
      
      <!-- Inner Add Button: Using a div with role=button because you can't nest buttons -->
      <div 
        role="button"
        tabindex="0"
        onclick={handleAdd}
        onkeydown={(e) => { if(e.key === 'Enter') handleAdd(e); }}
        style:background-color={justAdded ? undefined : colors.addBackground}
        style:color={justAdded ? undefined : colors.addForeground}
        class="font-bold py-2 px-4 rounded-full text-sm transition-all shadow-md flex items-center gap-2
        {justAdded ? 'bg-emerald-500 text-white scale-95' : 'hover:brightness-95'}"
      >
        {#if justAdded}
          <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg>
          Adicionado
        {:else}
          <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="8" cy="21" r="1"/><circle cx="19" cy="21" r="1"/><path d="M2.05 2.05h2l2.66 12.42a2 2 0 0 0 2 1.58h9.78a2 2 0 0 0 1.95-1.57l1.65-7.43H5.12"/></svg>
          Adicionar
        {/if}
      </div>
    </div>
  </div>
</button>

<style>
  @keyframes card-pop {
    0% { transform: scale(1); }
    40% { transform: scale(1.04) translateY(-2px); box-shadow: 0 10px 25px -5px rgba(255, 87, 34, 0.4); }
    100% { transform: scale(1) translateY(0); }
  }
  .animate-card-pop {
    animation: card-pop 0.4s cubic-bezier(0.34, 1.56, 0.64, 1);
    z-index: 10;
    border-color: #ff5722;
  }
</style>
