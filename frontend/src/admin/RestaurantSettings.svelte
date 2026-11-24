<script lang="ts">
  import { ImagePlus } from 'lucide-svelte';
  import { readConfig, saveConfig } from './menuConfig';
  let config = readConfig();
  let tags = config.tags.join(', ');
  let error = '';
  let imageError = false;
  let fileInput: HTMLInputElement;

  function persist() {
    try { saveConfig(config); error = ''; }
    catch { error = 'Não foi possível salvar os dados neste navegador.'; }
  }
  function updateTags() {
    config.tags = tags.split(',').map(tag => tag.trim()).filter(Boolean);
    persist();
  }
  function selectPicture(event: Event) {
    const file = (event.currentTarget as HTMLInputElement).files?.[0];
    if (!file) return;
    if (!file.type.startsWith('image/')) { error = 'Selecione uma imagem válida.'; return; }
    const reader = new FileReader();
    reader.onload = () => {
      config.picture = String(reader.result);
      imageError = false;
      persist();
    };
    reader.onerror = () => { error = 'Não foi possível ler a imagem.'; };
    reader.readAsDataURL(file);
  }
</script>

<h1 class="text-2xl font-bold text-zinc-900">Restaurante</h1>
<p class="mt-1 text-sm text-zinc-600">Configure as informações do seu restaurante.</p>
<section class="mt-6 max-w-3xl space-y-5 rounded-2xl border border-zinc-200 bg-white p-5 shadow-sm sm:p-6">
  <label class="block text-sm font-semibold text-zinc-700">Nome do restaurante
    <input class="mt-2 w-full rounded-xl border border-zinc-300 px-3 py-2" bind:value={config.name} on:input={persist} placeholder="Nome do restaurante" />
  </label>
  <div class="flex flex-col gap-4 sm:flex-row sm:items-start">
    <div class="grid h-40 w-full shrink-0 place-items-center overflow-hidden rounded-xl bg-zinc-100 text-zinc-400 sm:w-48">
      {#if config.picture && !imageError}
        <img src={config.picture} alt="Imagem do restaurante" class="h-full w-full object-cover" on:error={() => imageError = true} />
      {:else}
        <ImagePlus size={40} aria-label="Imagem padrão do restaurante" />
      {/if}
    </div>
    <div class="min-w-0 flex-1 space-y-3">
      <label class="block text-sm font-semibold text-zinc-700">Link da imagem
        <input type="url" class="mt-2 w-full rounded-xl border border-zinc-300 px-3 py-2 font-normal" value={config.picture.startsWith('data:') ? '' : config.picture} on:input={(event) => { config.picture = event.currentTarget.value; imageError = false; persist(); }} placeholder="https://exemplo.com/imagem.jpg" />
      </label>
      <input bind:this={fileInput} class="hidden" type="file" accept="image/*" on:change={selectPicture} />
      <button type="button" class="rounded-xl border border-zinc-300 px-4 py-2 text-sm font-semibold text-zinc-700" on:click={() => fileInput.click()}>Upload img</button>
      {#if config.picture}
        <button type="button" class="ml-2 text-sm text-red-600" on:click={() => { config.picture = ''; imageError = false; fileInput.value = ''; persist(); }}>Remover imagem</button>
      {/if}
      {#if imageError}<p class="text-sm text-red-600">Não foi possível carregar esta imagem.</p>{/if}
    </div>
  </div>
  <label class="block text-sm font-semibold text-zinc-700">Tags
    <input class="mt-2 w-full rounded-xl border border-zinc-300 px-3 py-2" bind:value={tags} on:input={updateTags} placeholder="Ex.: Italiana, Pizza, Delivery" />
    <span class="mt-1 block font-normal text-zinc-500">Separe as tags por vírgulas.</span>
  </label>
  {#if error}<p role="alert" class="text-sm text-red-600">{error}</p>{/if}
  <div class="border-t border-zinc-200 pt-5">
    <button type="button" disabled class="cursor-not-allowed rounded-xl bg-zinc-200 px-4 py-2 font-semibold text-zinc-500">Publicar no marketplace</button>
    <p class="mt-2 text-sm text-zinc-600">A publicação completa ainda depende de uma integração que envie imagem, tags e cardápio ao marketplace.</p>
  </div>
</section>
