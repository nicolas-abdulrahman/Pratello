<script lang="ts">
  import AdminLayout from './AdminLayout.svelte';
  import Dashboard from './Dashboard.svelte';
  import MenuManager from './MenuManager.svelte';
  import StaffManager from './StaffManager.svelte';
  import Settings from './Settings.svelte';
  import OrderCard from './components/OrderCard.svelte';
  import type { Order as PanelOrder } from './components/OrderItem.svelte';
  import { createAdminData, type AdminDependencies, type Order, type Product, type ProductInput, type StaffUser, type Category, type VisualCategory } from './admin';

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

  // Callbacks
  export let onLogout: () => void = dependencies.logout ?? (() => {});
  export let onResetDemoData: () => void = dependencies.resetDemoData ?? (() => {
    if (!data.testMode) return;
    const fresh = data.initialData();
    products = fresh.products; staffUsers = fresh.staff; orders = fresh.orders; panelOrders = fresh.orders;
  });
  export let onSaveProduct = dependencies.saveProduct ?? (async (product: ProductInput) => {
    products = data.saveProduct(products, product);
  });
  export let onDeleteProduct = dependencies.deleteProduct ?? (async (id: string) => { products = data.deleteProduct(products, id); });
  export let onToggleProductAvailability = dependencies.toggleProductAvailability ?? (async (id: string) => { products = data.toggleProduct(products, id); });
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
  {#if activeTab === 'dashboard'}
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
  {:else if activeTab === 'configuracoes'}
    <Settings
      {currentAccentId}
      onChangeAccent={onChangeAccent}
    />
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
