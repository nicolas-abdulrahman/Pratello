<script lang="ts">
  import { account } from '../lib/stores/account';
  import ProfileSection from './ProfileSection.svelte';
  import VerificationForm from './VerificationForm.svelte';

  let { mode }: { mode: 'email' | 'password' } = $props();
  let email = $state('');
  let currentPassword = $state('');
  let password = $state('');
  let targetEmail = '';

  async function requestCode() {
    targetEmail = mode === 'email' ? email.trim() : $account.email;
    await account.requestProfileEmailCode(targetEmail);
  }

  async function verify(code: string) {
    await account.verifyProfileEmail({ email: targetEmail, current_password: currentPassword, password, code });
    email = '';
    currentPassword = '';
    password = '';
  }
</script>

<ProfileSection
  title={mode === 'email' ? 'Alterar e-mail' : 'Alterar senha'}
  description={mode === 'email' ? 'Confirme o novo endereço com um código e defina a senha que usará com ele.' : 'Confirme a alteração com um código enviado ao seu e-mail atual.'}
>
  {#if mode === 'password' && !$account.email}
    <p class="text-sm text-text-muted">Adicione um e-mail para configurar uma senha.</p>
  {:else}
    <VerificationForm {requestCode} {verify}>
      {#snippet details()}
        {#if mode === 'email'}
          <label class="profile-field">Novo e-mail <input type="email" autocomplete="email" bind:value={email} required /></label>
        {/if}
        {#if $account.email}
          <label class="profile-field">Senha atual <input type="password" autocomplete="current-password" bind:value={currentPassword} required /></label>
        {/if}
        <label class="profile-field">
          {mode === 'email' ? 'Senha para o novo e-mail' : 'Nova senha'}
          <input type="password" minlength="8" autocomplete="new-password" bind:value={password} required />
          <span class="text-xs font-normal text-text-muted">Use pelo menos 8 caracteres.</span>
        </label>
      {/snippet}
    </VerificationForm>
  {/if}
</ProfileSection>
