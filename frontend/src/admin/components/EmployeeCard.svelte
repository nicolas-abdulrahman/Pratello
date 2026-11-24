<script lang="ts">
  import { Pencil, ShieldCheck, Trash2, UserRound } from 'lucide-svelte';

  export type Role = 'admin' | 'funcionario';

  export interface StaffUser {
    id: string;
    name: string;
    role: Role;
    login: string;
    password?: string;
  }

  interface Props {
    user: StaffUser;
    isCurrentUser?: boolean;
    onEdit?: (user: StaffUser) => void;
    onDelete?: (user: StaffUser) => void;
  }

  let {
    user,
    isCurrentUser = false,
    onEdit = () => {},
    onDelete = () => {}
  }: Props = $props();

  const ROLE_LABEL: Record<string, string> = {
    admin: 'Administrador',
    funcionario: 'Funcionário',
  };
</script>

<article class="flex flex-col rounded-2xl border border-zinc-200 bg-white p-5 shadow-sm">
  <div class="flex items-start gap-3">
    <span
      class={`grid h-11 w-11 shrink-0 place-items-center rounded-xl ${
        user.role === 'admin' ? 'bg-zinc-800 text-yellow-400' : 'bg-yellow-400 text-zinc-900'
      }`}
    >
      {#if user.role === 'admin'}
        <ShieldCheck size={20} />
      {:else}
        <UserRound size={20} />
      {/if}
    </span>
    <div class="min-w-0">
      <p class="truncate font-semibold text-zinc-900">
        {user.name}
        {#if isCurrentUser}
          <span class="text-xs font-normal text-zinc-500">(você)</span>
        {/if}
      </p>
      <p class="text-sm text-zinc-600">{ROLE_LABEL[user.role] || user.role}</p>
    </div>
  </div>

  <p class="mt-4 truncate rounded-lg bg-zinc-100 px-3 py-2 font-mono text-xs text-zinc-700">
    {user.login}
  </p>

  <div class="mt-4 flex justify-end gap-1">
    <button
      type="button"
      onclick={() => onEdit(user)}
      class="inline-flex items-center gap-1 rounded-xl px-3 py-1.5 text-sm font-semibold text-zinc-600 hover:bg-zinc-100"
      aria-label={`Editar ${user.name}`}
    >
      <Pencil size={16} /> Editar
    </button>
    <button
      type="button"
      disabled={isCurrentUser}
      onclick={() => onDelete(user)}
      class="inline-flex items-center gap-1 rounded-xl bg-red-50 px-3 py-1.5 text-sm font-semibold text-red-600 hover:bg-red-100 disabled:opacity-50 disabled:cursor-not-allowed"
      aria-label={`Excluir ${user.name}`}
    >
      <Trash2 size={16} /> Excluir
    </button>
  </div>
</article>
