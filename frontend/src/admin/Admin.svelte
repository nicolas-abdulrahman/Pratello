<script lang="ts">
  import AdminLayout from './AdminLayout.svelte';
  import Dashboard from './Dashboard.svelte';
  import MenuManager from './MenuManager.svelte';
  import StaffManager from './StaffManager.svelte';
  import Settings from './Settings.svelte';
  import type { AdminDependencies, Order, Product, StaffUser, Category, VisualCategory } from './admin';

  // --- Abstracted State & Dependencies ---
  export let dependencies: Partial<AdminDependencies> = {};

  export let session = dependencies.session ?? { name: 'Administrador Pratello', role: 'admin' as const };
  export let activeTab: string = 'dashboard';
  export let orders: Order[] = [];
  export let products: Product[] = [];
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
  export let staffUsers: StaffUser[] = [];
  export let currentAccentId: string = 'yellow';

  // Callbacks
  export let onLogout: () => void = dependencies.logout ?? (() => {});
  export let onResetDemoData: () => void = dependencies.resetDemoData ?? (() => {});
  export let onSaveProduct = dependencies.saveProduct ?? (async () => {});
  export let onDeleteProduct = dependencies.deleteProduct ?? (async () => {});
  export let onToggleProductAvailability = dependencies.toggleProductAvailability ?? (async () => {});
  export let onSaveStaffUser = dependencies.saveStaffUser ?? (async () => undefined);
  export let onDeleteStaffUser = dependencies.deleteStaffUser ?? (async () => undefined);
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
  {/if}
</AdminLayout>
