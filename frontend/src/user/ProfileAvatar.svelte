<script lang="ts">
  import { account } from '../lib/stores/account';
  import ProfileSection from './ProfileSection.svelte';

  let url = $state('');
  let busy = $state(false);
  let error = $state('');
  let saved = $state(false);

  async function save(value: string) {
    if (busy) return;
    busy = true;
    error = '';
    saved = false;
    try {
      await account.setAvatar(value);
      url = '';
      saved = true;
    } catch (err) {
      error = err instanceof Error ? err.message : 'Não foi possível salvar a foto.';
    } finally {
      busy = false;
    }
  }
</script>

<ProfileSection title="Foto de perfil" description="Cole o link direto para a imagem que deseja usar.">
  <form onsubmit={(event) => { event.preventDefault(); void save(url.trim()); }} class="space-y-space-sm">
    <label class="profile-field">
      Link da imagem
      <input type="text" inputmode="url" placeholder="https://exemplo.com/foto.jpg" maxlength="2048" bind:value={url} required />
    </label>
    {#if error}<p role="alert" class="text-sm text-error">{error}</p>{/if}
    {#if saved}<p role="status" class="text-sm text-text-muted">Foto atualizada.</p>{/if}
    <div class="flex flex-wrap items-center gap-space-sm">
      <button type="submit" class="profile-button" disabled={busy || !url.trim()}>Salvar foto</button>
      {#if $account.avatar_url}
        <button type="button" onclick={() => save('')} disabled={busy} class="text-sm font-semibold text-brand-accent">Remover foto</button>
      {/if}
    </div>
  </form>
</ProfileSection>
