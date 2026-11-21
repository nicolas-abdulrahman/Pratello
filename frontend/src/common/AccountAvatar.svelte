<script lang="ts">
  import { account } from '../lib/stores/account';

  let { class: className = '', label }: { class?: string; label?: string } = $props();
  let failedUrl = $state('');
</script>

<span class="grid shrink-0 place-items-center overflow-hidden bg-primary-soft font-bold text-brand-accent {className}" role={label ? 'img' : undefined} aria-label={label} aria-hidden={label ? undefined : true}>
  {#key $account.avatar_url}
    {@const url = $account.avatar_url}
    {#if url && url !== failedUrl}
      <img src={url} alt="" referrerpolicy="no-referrer" class="size-full object-cover" onerror={() => { failedUrl = url; }} />
    {:else}
      <span>{$account.initials || 'U'}</span>
    {/if}
  {/key}
</span>
