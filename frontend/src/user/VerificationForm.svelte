<script lang="ts">
  import type { Snippet } from 'svelte';

  let { details, requestCode, verify }: {
    details: Snippet;
    requestCode: () => Promise<void>;
    verify: (code: string) => Promise<void>;
  } = $props();

  let awaitingCode = $state(false);
  let busy = $state(false);
  let code = $state('');
  let error = $state('');
  let message = $state('');

  async function run(action: () => Promise<void>) {
    if (busy) return;
    busy = true;
    error = '';
    message = '';
    try {
      await action();
    } catch (err) {
      error = err instanceof Error ? err.message : 'Não foi possível concluir a solicitação.';
    } finally {
      busy = false;
    }
  }

  function sendCode() {
    return run(async () => {
      await requestCode();
      awaitingCode = true;
      code = '';
      message = 'Código solicitado. Confira o canal de verificação.';
    });
  }

  function submit(event: SubmitEvent) {
    event.preventDefault();
    if (!awaitingCode) return sendCode();
    return run(async () => {
      await verify(code);
      awaitingCode = false;
      code = '';
      message = 'Alteração salva.';
    });
  }

  function back() {
    awaitingCode = false;
    code = '';
    error = '';
    message = '';
  }
</script>

<form onsubmit={submit} class="space-y-space-sm">
  {#if awaitingCode}
    <label class="profile-field">
      Código de verificação
      <input type="text" inputmode="numeric" autocomplete="one-time-code" pattern={'[0-9]{6}'} maxlength="6" bind:value={code} required />
    </label>
  {:else}
    <fieldset disabled={busy} class="space-y-space-sm">
      {@render details()}
    </fieldset>
  {/if}

  {#if error}<p role="alert" class="text-sm text-error">{error}</p>{/if}
  {#if message}<p role="status" class="text-sm text-text-muted">{message}</p>{/if}

  <div class="flex flex-wrap items-center gap-space-sm">
    <button type="submit" class="profile-button" disabled={busy}>
      {busy ? 'Aguarde…' : awaitingCode ? 'Confirmar código' : 'Solicitar código'}
    </button>
    {#if awaitingCode}
      <button type="button" onclick={sendCode} disabled={busy} class="text-sm font-semibold text-brand-accent">Reenviar código</button>
      <button type="button" onclick={back} disabled={busy} class="text-sm font-semibold text-text">Voltar</button>
    {/if}
  </div>
</form>
