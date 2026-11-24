<script lang="ts">
  import AdminLayout from './AdminLayout.svelte';
  import Dashboard from './Dashboard.svelte';
  import MenuManager from './MenuManager.svelte';
  import StaffManager from './StaffManager.svelte';
  import Settings from './Settings.svelte';
  import MenuBuilder from './MenuBuilder.svelte';
  import PricesManager from './PricesManager.svelte';
  import RestaurantSettings from './RestaurantSettings.svelte';
  import OrderCard from './components/OrderCard.svelte';
  import type { Order as PanelOrder } from './components/OrderItem.svelte';
  import { createAdminData, type AdminDependencies, type Order, type Product, type ProductInput, type StaffUser, type Category, type VisualCategory } from './admin';

  import { onMount } from 'svelte';
  import {
    listMenuItems,
    createMenuItem,
    updateMenuItem,
    deleteMenuItem,
    patchMenuItem,
    getMenuCategories
  } from './menu';

  // --- Abstracted State & Dependencies ---
  export let dependencies: Partial<AdminDependencies> = {};
  export let slug = '';
  const data = createAdminData(slug);
  const initial = data.initialData();

  export let session = dependencies.session ?? { name: 'Administrador Pratello', role: 'admin' as const };
  export let activeTab: string = 'dashboard';
  export let orders: Order[] = initial.orders;
  export let products: Product[] = initial.products;
  export let categories: Category[] = [
    { id: 'lanches', name: 'Lanches' },
    { id: 'pratos', name: 'Pratos' },
    { id: 'bebidas', name: 'Bebidas' },
    { id: 'sobremesas', name: 'Sobremesas' },
  ];
  export let visualCategories: VisualCategory[] = [
    { id: 'destaque', name: 'Mais pedidos' },
    { id: 'vegetariano', name: 'Vegetariano' },
    { id: 'sobremesas', name: 'Doces' },
  ];
  export let staffUsers: StaffUser[] = initial.staff;
  export let panelOrders: PanelOrder[] = initial.orders;
  export let currentAccentId: string = 'yellow';

  function mapBackendItemToProduct(item: any): Product {
    const basePrice = (item.base_price_cents ?? 0) / 100;
    const effectivePrice = (item.effective_price_cents ?? item.base_price_cents ?? 0) / 100;
    const isSale = item.effective_price_cents !== undefined && item.effective_price_cents < item.base_price_cents;
    return {
      id: String(item.id),
      name: item.name || '',
      description: item.description || '',
      price: effectivePrice || basePrice,
      originalPrice: isSale ? basePrice : undefined,
      prepTime: item.preparation_time_minutes ?? 20,
      category: item.category_id || 'lanches',
      tags: [],
      image: item.image_url || '',
      ingredients: [],
      available: item.is_available ?? true,
      popular: false,
      base_price_cents: item.base_price_cents,
      effective_price_cents: item.effective_price_cents,
      base_cost_cents: item.base_cost_cents,
      is_combo: item.is_combo,
      combo_items: item.combo_items,
      display_style: item.display_style,
    };
  }

  onMount(async () => {
    if (slug) {
      try {
        const [apiItems, apiCats] = await Promise.all([
          listMenuItems(slug),
          getMenuCategories(slug),
        ]);
        if (Array.isArray(apiItems) && apiItems.length > 0) {
          products = apiItems.map(mapBackendItemToProduct);
        }
        if (Array.isArray(apiCats) && apiCats.length > 0) {
          categories = apiCats.map((c: any) => ({ id: String(c.id), name: c.name }));
        }
      } catch (e) {
        console.warn('Could not load menu items or categories from API:', e);
      }
    }
  });

  // Callbacks
  export let onLogout: () => void = dependencies.logout ?? (() => {});
  export let onResetDemoData: () => void = dependencies.resetDemoData ?? (() => {
    if (!data.testMode) return;
    const fresh = data.initialData();
    products = fresh.products; staffUsers = fresh.staff; orders = fresh.orders; panelOrders = fresh.orders;
  });
  export let onSaveProduct = dependencies.saveProduct ?? (async (product: ProductInput) => {
    if (slug) {
      try {
        const payload: any = {
          name: product.name,
          description: product.description || null,
          category_id: product.category || null,
          base_price_cents: product.base_price_cents ?? Math.round((product.originalPrice ?? product.price) * 100),
          effective_price_cents: product.effective_price_cents ?? Math.round(product.price * 100),
          base_cost_cents: product.base_cost_cents ?? null,
          image_url: product.image || null,
          is_available: product.available ?? true,
          is_combo: product.is_combo ?? false,
          preparation_time_minutes: product.prepTime ?? null,
          display_style: product.display_style ?? null,
        };
        if (product.id && !product.id.startsWith('product-')) {
          await updateMenuItem(slug, product.id, payload);
        } else {
          const created = await createMenuItem(slug, payload);
          if (created?.id) product.id = String(created.id);
        }
      } catch (e) {
        console.warn('Could not sync item with API:', e);
      }
    }
    products = data.saveProduct(products, product);
  });
  export let onDeleteProduct = dependencies.deleteProduct ?? (async (id: string) => {
    if (slug && !id.startsWith('product-')) {
      try {
        await deleteMenuItem(slug, id);
      } catch (e) {
        console.warn('Could not delete item on API:', e);
      }
    }
    products = data.deleteProduct(products, id);
  });
  export let onToggleProductAvailability = dependencies.toggleProductAvailability ?? (async (id: string) => {
    const item = products.find((p) => p.id === id);
    if (slug && item && !id.startsWith('product-')) {
      try {
        await patchMenuItem(slug, id, { is_available: !item.available });
      } catch (e) {
        console.warn('Could not patch item on API:', e);
      }
    }
    products = data.toggleProduct(products, id);
  });
  export let onSaveStaffUser = dependencies.saveStaffUser ?? (async (user: StaffUser) => { staffUsers = data.saveStaff(staffUsers, user); return undefined; });
  export let onDeleteStaffUser = dependencies.deleteStaffUser ?? (async (id: string) => { staffUsers = data.deleteStaff(staffUsers, id); return undefined; });
  export let onChangeAccent = dependencies.setAccent ?? (() => {});
</script>

<AdminLayout
  {session}
  bind:activeTab
  {onLogout}
  {onResetDemoData}
  onNavigate={(tab) => (activeTab = tab)}
>
  {#if activeTab === 'restaurante'}
    <RestaurantSettings />
  {:else if activeTab === 'dashboard'}
    <Dashboard {orders} {products} {categories} />
  {:else if activeTab === 'cardapio'}
    <MenuManager
      {products}
      {categories}
      {visualCategories}
      onSaveProduct={onSaveProduct}
      onDeleteProduct={onDeleteProduct}
      onToggleAvailability={onToggleProductAvailability}
    />
  {:else if activeTab === 'funcionarios'}
    <StaffManager
      users={staffUsers}
      currentUserId={session?.userId ?? ''}
      onSaveUser={onSaveStaffUser}
      onDeleteUser={onDeleteStaffUser}
    />
  {:else if activeTab === 'cores' || activeTab === 'configuracoes'}
    <Settings
      {currentAccentId}
      onChangeAccent={onChangeAccent}
    />
  {:else if activeTab === 'precos'}
    <PricesManager {slug} bind:products {categories} {onSaveProduct} />
  {:else if activeTab === 'preparar'}
    <MenuBuilder {slug} {products} {categories} />
  {:else if activeTab === 'pedidos'}
    <h1 class="text-2xl font-bold text-zinc-900">Painel de Pedidos</h1>
    <div class="mt-6 grid gap-4 md:grid-cols-2 xl:grid-cols-3">
      {#each panelOrders as order (order.id)}
        <OrderCard {order}
          onAdvance={() => {
            panelOrders = data.advanceOrder(panelOrders, order.id);
            orders = panelOrders;
          }}
          onMarkPaid={() => {
            panelOrders = data.markPaid(panelOrders, order.id);
            orders = panelOrders;
          }} />
      {:else}
        <p class="text-zinc-500">Nenhum pedido disponível.</p>
      {/each}
    </div>
  {/if}
</AdminLayout>
