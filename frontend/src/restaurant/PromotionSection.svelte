<script lang="ts">
  import MelhoresPratosCard from './PromotionCard.svelte';
  import ItemModal from '../dish/ItemModal.svelte';
  import type { Item } from '../lib/stores/Item';

  /** Promotion-only display fields layered on top of the base Item type. */
  type PromotionItem = Item & {
    originalPrice?: number;
    isPopular?: boolean;
  };

  const pratos: PromotionItem[] = [
    {
      id: 1,
      name: 'Hambúrguer Duplo Especial',
      description: 'Dois hambúrgueres, queijo derretido, bacon crocante e muito sabor!',
      imageUrl:
        'https://images.unsplash.com/photo-1568901346375-23c9450c58cd?auto=format&fit=crop&q=80&w=300&h=300',
      basePrice: 22.9,
      originalPrice: 34.9,
      isPopular: true,
      type: 'simple',
      ingredients: [
        { id: 1, name: 'Pão brioche artesanal', selected: true },
        { id: 2, name: 'Hambúrguer 180g (1º)', selected: true },
        { id: 3, name: 'Hambúrguer 180g (2º)', selected: true },
        { id: 4, name: 'Queijo cheddar inglês', selected: true },
        { id: 5, name: 'Bacon defumado', selected: true },
        { id: 6, name: 'Alface americana', selected: true },
        { id: 7, name: 'Tomate', selected: true },
        { id: 8, name: 'Molho secreto da casa', selected: true },
      ],
      extras: [
        { id: 1, name: 'Bacon extra', price: 6.0, selected: false },
        { id: 2, name: 'Ovo frito', price: 3.5, selected: false },
        { id: 3, name: 'Cheddar duplo', price: 4.0, selected: false },
      ],
    },
    {
      id: 2,
      name: 'Pizza Pepperoni',
      description: 'Massa artesanal, molho de tomate e pepperoni selecionado.',
      imageUrl:
        'https://images.unsplash.com/photo-1628840042765-356cda07504e?auto=format&fit=crop&q=80&w=300&h=300',
      basePrice: 29.9,
      originalPrice: 39.9,
      type: 'simple',
      ingredients: [
        { id: 1, name: 'Massa de fermentação natural (48h)', selected: true },
        { id: 2, name: 'Molho de tomate San Marzano', selected: true },
        { id: 3, name: 'Muçarela de búfala', selected: true },
        { id: 4, name: 'Pepperoni artesanal', selected: true },
        { id: 5, name: 'Manjericão fresco', selected: true },
      ],
      extras: [
        { id: 1, name: 'Muçarela extra', price: 5.0, selected: false },
        { id: 2, name: 'Pepperoni extra', price: 6.5, selected: false },
        { id: 3, name: 'Azeitonas pretas', price: 3.0, selected: false },
      ],
    },
    {
      id: 3,
      name: 'Porção de Batata Frita',
      description: 'Batata crocante e sequinha, perfeita para compartilhar.',
      imageUrl:
        'https://images.unsplash.com/photo-1573080496219-bb080dd4f877?auto=format&fit=crop&q=80&w=300&h=300',
      basePrice: 12.9,
      originalPrice: 19.9,
      type: 'simple',
      ingredients: [
        { id: 1, name: 'Batata inglesa', selected: true },
        { id: 2, name: 'Sal marinho', selected: true },
        { id: 3, name: 'Alecrim', selected: true },
      ],
      extras: [
        { id: 1, name: 'Cheddar cremoso', price: 4.5, selected: false },
        { id: 2, name: 'Bacon em cubos', price: 5.0, selected: false },
        { id: 3, name: 'Azeite trufado', price: 7.0, selected: false },
      ],
    },
    {
      id: 4,
      name: 'Sanduíche de Frango',
      description: 'Frango grelhado, alface e molho especial da casa.',
      imageUrl:
        'https://images.unsplash.com/photo-1521390188846-e2a3a97453a0?auto=format&fit=crop&q=80&w=300&h=300',
      basePrice: 18.9,
      originalPrice: 25.9,
      type: 'simple',
      ingredients: [
        { id: 1, name: 'Pão ciabatta', selected: true },
        { id: 2, name: 'Peito de frango grelhado', selected: true },
        { id: 3, name: 'Queijo suíço', selected: true },
        { id: 4, name: 'Alface crocante', selected: true },
        { id: 5, name: 'Tomate', selected: true },
        { id: 6, name: 'Maionese de limão siciliano', selected: true },
      ],
      extras: [
        { id: 1, name: 'Queijo extra', price: 4.0, selected: false },
        { id: 2, name: 'Bacon', price: 5.5, selected: false },
      ],
    },
    {
      id: 5,
      name: 'Onion Rings Crocantes',
      description: 'Anéis de cebola empanados na hora, super crocantes e saborosos.',
      imageUrl:
        'https://images.unsplash.com/photo-1639024471283-03518883512d?auto=format&fit=crop&q=80&w=300&h=300',
      basePrice: 15.9,
      originalPrice: 22.0,
      type: 'simple',
      ingredients: [
        { id: 1, name: 'Cebola roxa', selected: true },
        { id: 2, name: 'Empanado panko', selected: true },
      ],
      extras: [
        { id: 1, name: 'Molho ranch', price: 3.0, selected: false },
        { id: 2, name: 'Molho cheddar', price: 4.0, selected: false },
      ],
    },
    {
      id: 6,
      name: 'Combo Casal',
      description: '2 Hambúrgueres, 1 Batata Grande e 2 Refrigerantes.',
      imageUrl:
        'https://images.unsplash.com/photo-1594212699903-ec8a3eca50f5?auto=format&fit=crop&q=80&w=300&h=300',
      basePrice: 55.9,
      originalPrice: 75.0,
      isPopular: true,
      type: 'combo',
      components: [
        {
          name: 'Hambúrguer Duplo',
          quantity: 2,
          imageUrl:
            'https://images.unsplash.com/photo-1568901346375-23c9450c58cd?auto=format&fit=crop&q=80&w=120&h=120',
        },
        { name: 'Batata Frita Grande', quantity: 1 },
        { name: 'Refrigerante Lata', quantity: 2 },
      ],
      ingredients: [],
      extras: [
        { id: 1, name: 'Batata extra', price: 8.0, selected: false },
        { id: 2, name: 'Upgrade p/ refrigerante 2L', price: 6.0, selected: false },
        { id: 3, name: 'Brownie extra', price: 9.0, selected: false },
      ],
    },
    {
      id: 7,
      name: 'Brownie com Sorvete',
      description: 'Brownie quentinho acompanhado de sorvete de creme.',
      imageUrl:
        'https://images.unsplash.com/photo-1606313564200-e75d5e30476c?auto=format&fit=crop&q=80&w=300&h=300',
      basePrice: 12.5,
      originalPrice: 18.0,
      type: 'simple',
      ingredients: [
        { id: 1, name: 'Brownie de chocolate belga', selected: true },
        { id: 2, name: 'Sorvete de creme', selected: true },
        { id: 3, name: 'Calda de chocolate', selected: true },
      ],
      extras: [
        { id: 1, name: 'Bola extra de sorvete', price: 3.5, selected: false },
        { id: 2, name: 'Chantilly', price: 2.0, selected: false },
        { id: 3, name: 'Castanhas caramelizadas', price: 4.0, selected: false },
      ],
    },
  ];

  let isModalOpen = $state(false);
  let selectedItem = $state<PromotionItem | null>(null);

  function openModal(item: PromotionItem) {
    selectedItem = item;
    isModalOpen = true;
  }
</script>

<section class="py-8 w-full px-4 md:px-8">
  <div class="flex items-center mb-8">
    <h2 class="text-3xl md:text-4xl font-black text-gray-900 tracking-tight">
      Promo<span class="text-[#ff5722]">ções</span>
    </h2>
  </div>

  <div class="relative border-[4px] border-[#ff5722] rounded-[3rem] p-6 md:p-8 bg-orange-50/40 shadow-xl overflow-hidden">
    <div class="w-full">
      <div class="flex md:grid md:grid-cols-3 lg:grid-cols-4 md:grid-rows-2 gap-4 overflow-x-auto md:overflow-visible pb-6 md:pb-0 snap-x snap-mandatory hide-scrollbar">
        {#each pratos as prato (prato.id)}
          <div class="snap-start min-w-[280px] w-full flex-shrink-0">
            <MelhoresPratosCard
              item={prato}
              onOpenDetails={() => openModal(prato)}
            />
          </div>
        {/each}
      </div>
    </div>
  </div>
</section>

{#if selectedItem}
  <ItemModal
    isOpen={isModalOpen}
    onClose={() => (isModalOpen = false)}
    item={selectedItem}
  />
{/if}

<style>
  .hide-scrollbar::-webkit-scrollbar { display: none; }
  .hide-scrollbar { -ms-overflow-style: none; scrollbar-width: none; }
</style>
