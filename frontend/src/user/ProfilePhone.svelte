<script lang="ts">
  import { account } from '../lib/stores/account';
  import ProfileSection from './ProfileSection.svelte';
  import VerificationForm from './VerificationForm.svelte';

  let phone = $state('');
  let targetPhone = '';

  function requestCode() {
    const digits = phone.replace(/\D/g, '');
    if (digits.length !== 10 && digits.length !== 11) throw new Error('Informe um telefone com DDD e 8 ou 9 dígitos.');
    targetPhone = `+55${digits}`;
    return account.requestProfilePhoneCode(targetPhone);
  }

  async function verify(code: string) {
    await account.verifyProfilePhone(targetPhone, code);
    phone = '';
  }
</script>

<ProfileSection title="Alterar telefone" description="Informe o novo número com DDD. Você não precisa incluir +55.">
  <VerificationForm {requestCode} {verify}>
    {#snippet details()}
      <label class="profile-field">
        Novo telefone
        <input type="tel" autocomplete="tel-national" placeholder="(81) 90000-0000" maxlength="15" bind:value={phone} required />
      </label>
    {/snippet}
  </VerificationForm>
</ProfileSection>
