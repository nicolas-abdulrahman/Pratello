<script lang="ts">
  import { onMount } from 'svelte';
  import { account } from '../lib/stores/account';
  import ProfileSection from './ProfileSection.svelte';
  import ProfileAvatar from './ProfileAvatar.svelte';
  import ProfileCredentials from './ProfileCredentials.svelte';
  import ProfilePhone from './ProfilePhone.svelte';

  let { onBack, onGoLogin }: { onBack: () => void; onGoLogin: () => void } = $props();
  const sections = [
    { id: 'avatar', label: 'Foto de perfil' },
    { id: 'email', label: 'E-mail' },
    { id: 'password', label: 'Senha' },
    { id: 'phone', label: 'Telefone' }
  ] as const;
  let active = $state<(typeof sections)[number]['id']>('avatar');
  let loading = $state(false);
  let failed = $state(false);
  let failedAvatar = $state('');

  async function loadProfile() {
    loading = true;
    failed = !(await account.fetchMe());
    loading = false;
  }

  onMount(() => {
    window.scrollTo(0, 0);
    if ($account.isLoggedIn) void loadProfile();
  });
</script>

<main class="account-page mx-auto min-h-screen max-w-5xl space-y-space-md px-space-sm py-space-lg sm:px-space-lg sm:py-space-xl">
  <button type="button" onclick={onBack} class="inline-flex items-center gap-space-2xs text-sm font-bold text-brand-accent hover:underline">
    <span aria-hidden="true">←</span> Voltar
  </button>
  <h1 class="text-3xl font-bold sm:text-4xl">Minha conta</h1>

  {#if !$account.isLoggedIn}
    <ProfileSection title="Acesse seu perfil" description="Entre para ver e atualizar seus dados.">
      <button type="button" class="profile-button" onclick={onGoLogin}>Entrar</button>
    </ProfileSection>
  {:else if loading}
    <p role="status">Carregando perfil…</p>
  {:else if failed}
    <p role="alert" class="text-error">Não foi possível carregar seu perfil.</p>
    <button type="button" class="profile-button" onclick={loadProfile}>Tentar novamente</button>
  {:else}
    <div class="grid items-start gap-space-md md:grid-cols-[minmax(0,1fr)_minmax(0,2fr)]">
      <ProfileSection title="Meu perfil">
        {#if $account.avatar_url && failedAvatar !== $account.avatar_url}
          <img src={$account.avatar_url} alt="Foto de perfil" class="mb-space-sm size-20 rounded-full object-cover" referrerpolicy="no-referrer" onerror={() => { failedAvatar = $account.avatar_url ?? ''; }} />
        {:else}
          <div aria-hidden="true" class="mb-space-sm grid size-20 place-items-center rounded-full bg-primary-soft text-2xl font-bold text-brand-accent">{$account.initials || 'U'}</div>
        {/if}
        <p class="break-words font-bold">{$account.name}</p>
        <dl class="mt-space-sm space-y-space-xs text-sm">
          <div><dt class="font-semibold">E-mail</dt><dd class="break-words text-text-muted">{$account.email || 'Não cadastrado'}</dd></div>
          <div><dt class="font-semibold">Telefone</dt><dd class="text-text-muted">{$account.phone || 'Não cadastrado'}</dd></div>
        </dl>
      </ProfileSection>

      <div class="space-y-space-sm">
        <div role="group" aria-label="Configurações do perfil" class="flex flex-wrap gap-space-2xs">
          {#each sections as section}
            <button type="button" aria-pressed={active === section.id} onclick={() => { active = section.id; }} class="rounded-lg px-space-xs py-space-2xs text-sm font-semibold {active === section.id ? 'bg-primary-soft text-brand-accent' : 'bg-surface text-text hover:bg-surface-hover'}">
              {section.label}
            </button>
          {/each}
        </div>
        {#key active}
          {#if active === 'avatar'}
            <ProfileAvatar />
          {:else if active === 'phone'}
            <ProfilePhone />
          {:else}
            <ProfileCredentials mode={active} />
          {/if}
        {/key}
      </div>
    </div>
  {/if}
</main>

<style>
  .account-page :global(.profile-field) {
    display: grid;
    gap: var(--spacing-space-3xs);
    font-size: var(--text-sm);
    font-weight: 600;
  }
  .account-page :global(.profile-field input) {
    width: 100%;
    border: 1px solid var(--color-border);
    border-radius: var(--radius-lg);
    background: var(--color-background);
    padding: var(--spacing-space-2xs) var(--spacing-space-xs);
    font-weight: 400;
  }
  .account-page :global(.profile-button) {
    border-radius: var(--radius-lg);
    padding: var(--spacing-space-2xs) var(--spacing-space-sm);
    background: var(--color-brand-accent);
    color: var(--color-brand-on-accent);
    font-weight: 700;
  }
  .account-page :global(.profile-button:hover:not(:disabled)) { background: var(--color-brand-accent-hover); }
  .account-page :global(.profile-button:disabled) { opacity: .6; }
</style>
