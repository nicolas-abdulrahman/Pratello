<script lang="ts">
  import { Pencil, Plus, ShieldCheck, Trash2, UserRound, X } from 'lucide-svelte';
  import type { Role, StaffUser } from './admin';

  // --- Abstracted State & Dependencies ---
  export let users: StaffUser[] = [];
  export let currentUserId: string = '';
  export let onSaveUser: (u: StaffUser) => Promise<string | undefined> | string | undefined = () => undefined;
  export let onDeleteUser: (id: string) => Promise<string | undefined> | string | undefined = () => undefined;

  interface FormState {
    id?: string;
    name: string;
    role: Role;
    login: string;
    password: string;
  }

  const ROLE_LABEL: Record<Role, string> = {
    admin: 'Administrador',
    funcionario: 'Funcionário',
  };

  let form: FormState | null = null;
  let error: string = '';
  let listError: string = '';

  function openModal(u?: StaffUser) {
    error = '';
    form = u
      ? { id: u.id, name: u.name, role: u.role, login: u.login, password: u.password ?? '' }
      : { name: '', role: 'funcionario', login: '', password: '' };
  }

  function closeModal() {
    form = null;
    error = '';
  }

  async function handleSubmit() {
    if (!form) return;
    if (!form.name.trim() || !form.login.trim() || !form.password) {
      error = 'Preencha todos os campos.';
      return;
    }

    const res = await onSaveUser({
      id: form.id ?? `staff-${Date.now()}`,
      name: form.name.trim(),
      role: form.role,
      login: form.login.trim(),
      password: form.password,
    });

    if (res) {
      error = res;
      return;
    }
    closeModal();
  }

  async function handleRemove(u: StaffUser) {
    listError = '';
    if (!confirm(`Excluir o acesso de ${u.name}?`)) return;
    const res = await onDeleteUser(u.id);
    if (res) {
      listError = res;
    }
  }
</script>

<div>
  <!-- Header -->
  <div class="flex flex-wrap items-end justify-between gap-3">
    <div>
      <h1 class="text-2xl font-bold text-zinc-900">Gerenciar Funcionários</h1>
      <p class="text-sm text-zinc-600">{users.length} pessoas com acesso à área da equipe</p>
    </div>
    <button
      on:click={() => openModal()}
      class="inline-flex items-center gap-2 rounded-xl bg-yellow-400 px-4 py-2.5 text-sm font-semibold text-zinc-900 hover:bg-yellow-500"
    >
      <Plus size={18} /> Novo funcionário
    </button>
  </div>

  {#if listError}
    <p class="mt-4 rounded-xl bg-red-50 p-3 text-sm font-medium text-red-600">{listError}</p>
  {/if}

  <!-- Users Grid -->
  <div class="mt-6 grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
    {#each users as u (u.id)}
      <article class="flex flex-col rounded-2xl border border-zinc-200 bg-white p-5 shadow-sm">
        <div class="flex items-start gap-3">
          <span
            class={`grid h-11 w-11 shrink-0 place-items-center rounded-xl ${
              u.role === 'admin' ? 'bg-zinc-800 text-yellow-400' : 'bg-yellow-400 text-zinc-900'
            }`}
          >
            {#if u.role === 'admin'}
              <ShieldCheck size={20} />
            {:else}
              <UserRound size={20} />
            {/if}
          </span>
          <div class="min-w-0">
            <p class="truncate font-semibold text-zinc-900">
              {u.name}
              {#if u.id === currentUserId}
                <span class="text-xs font-normal text-zinc-500">(você)</span>
              {/if}
            </p>
            <p class="text-sm text-zinc-600">{ROLE_LABEL[u.role] || u.role}</p>
          </div>
        </div>

        <p class="mt-4 truncate rounded-lg bg-zinc-100 px-3 py-2 font-mono text-xs text-zinc-700">
          {u.login}
        </p>

        <div class="mt-4 flex justify-end gap-1">
          <button
            on:click={() => openModal(u)}
            class="inline-flex items-center gap-1 rounded-xl px-3 py-1.5 text-sm font-semibold text-zinc-600 hover:bg-zinc-100"
          >
            <Pencil size={16} /> Editar
          </button>
          <button
            disabled={u.id === currentUserId}
            on:click={() => handleRemove(u)}
            class="inline-flex items-center gap-1 rounded-xl bg-red-50 px-3 py-1.5 text-sm font-semibold text-red-600 hover:bg-red-100 disabled:opacity-50"
          >
            <Trash2 size={16} /> Excluir
          </button>
        </div>
      </article>
    {/each}
  </div>

  <!-- Form Modal -->
  {#if form}
    <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div class="w-full max-w-md rounded-2xl bg-white p-6 shadow-xl">
        <div class="mb-4 flex items-center justify-between">
          <h2 class="text-lg font-bold text-zinc-900">
            {form.id ? 'Editar funcionário' : 'Novo funcionário'}
          </h2>
          <button on:click={closeModal} class="rounded-lg p-1 text-zinc-400 hover:bg-zinc-100">
            <X size={20} />
          </button>
        </div>

        <form on:submit|preventDefault={handleSubmit} class="space-y-4">
          <label class="block">
            <span class="mb-1.5 block text-sm font-medium text-zinc-700">Nome</span>
            <input
              type="text"
              bind:value={form.name}
              class="w-full rounded-xl border border-zinc-200 bg-white px-3.5 py-2.5 text-sm text-zinc-900 focus:border-yellow-400 focus:outline-none"
            />
          </label>

          <label class="block">
            <span class="mb-1.5 block text-sm font-medium text-zinc-700">Cargo</span>
            <select
              bind:value={form.role}
              class="w-full rounded-xl border border-zinc-200 bg-white px-3.5 py-2.5 text-sm text-zinc-900 focus:border-yellow-400 focus:outline-none"
            >
              <option value="funcionario">Funcionário</option>
              <option value="admin">Administrador</option>
            </select>
          </label>

          <label class="block">
            <span class="mb-1.5 block text-sm font-medium text-zinc-700">Login (usuário ou e-mail)</span>
            <input
              type="text"
              bind:value={form.login}
              placeholder="nome@pratello.com"
              class="w-full rounded-xl border border-zinc-200 bg-white px-3.5 py-2.5 text-sm text-zinc-900 focus:border-yellow-400 focus:outline-none"
            />
          </label>

          <label class="block">
            <span class="mb-1.5 block text-sm font-medium text-zinc-700">Senha</span>
            <input
              type="password"
              bind:value={form.password}
              class="w-full rounded-xl border border-zinc-200 bg-white px-3.5 py-2.5 text-sm text-zinc-900 focus:border-yellow-400 focus:outline-none"
            />
          </label>

          {#if error}
            <p class="text-sm font-medium text-red-600">{error}</p>
          {/if}

          <div class="flex justify-end gap-2 pt-2">
            <button
              type="button"
              on:click={closeModal}
              class="rounded-xl border border-zinc-200 bg-white px-4 py-2.5 text-sm font-semibold text-zinc-700 hover:bg-zinc-50"
            >
              Cancelar
            </button>
            <button
              type="submit"
              class="rounded-xl bg-yellow-400 px-4 py-2.5 text-sm font-semibold text-zinc-900 hover:bg-yellow-500"
            >
              Salvar
            </button>
          </div>
        </form>
      </div>
    </div>
  {/if}
</div>
