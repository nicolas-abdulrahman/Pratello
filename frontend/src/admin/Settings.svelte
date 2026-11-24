<script lang="ts">
  import { Check, Flame, Palette, Plus, RotateCcw, Star } from 'lucide-svelte';
  import { DEFAULT_ACCENTS, type AccentColor } from './admin';
  import { readConfig, saveConfig, type MenuColors } from './menuConfig';
  let colors: MenuColors = readConfig().colors;
  const colorFields: { key: keyof MenuColors; label: string }[] = [
    { key: 'bg', label: 'Background' }, { key: 'primary', label: 'Primary' },
    { key: 'secondary', label: 'Secondary' }, { key: 'bg_dark', label: 'Background dark' },
    { key: 'primary_dark', label: 'Primary dark' }, { key: 'secondary_dark', label: 'Secondary dark' },
  ];
  function updateColor(key: keyof MenuColors, value: string) {
    colors = { ...colors, [key]: value };
    saveConfig({ ...readConfig(), colors });
  }

  // --- Abstracted State & Dependencies ---
  export let accents: AccentColor[] = DEFAULT_ACCENTS;
  export let currentAccentId: string = 'yellow';
  export let defaultAccentId: string = 'yellow';
  export let onChangeAccent: (id: string) => void = () => {};

  function handleSelect(id: string) {
    currentAccentId = id;
    onChangeAccent(id);
  }

  function handleResetDefault() {
    handleSelect(defaultAccentId);
  }
</script>

<div>
  <h1 class="text-2xl font-bold text-zinc-900">Cores</h1>
  <p class="text-sm text-zinc-600">Personalize a aparência do restaurante.</p>

  <section class="mt-6 rounded-2xl border border-zinc-200 bg-white p-5 shadow-sm sm:p-6">
    <h2 class="font-bold text-zinc-900">Cores do cardápio</h2>
    <p class="mt-1 text-sm text-zinc-600">Defina as cores para os temas claro e escuro.</p>
    <div class="mt-4 grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
      {#each colorFields as field}
        <label class="flex items-center gap-3 rounded-xl border border-zinc-200 p-3 text-sm font-medium">
          <input type="color" value={colors[field.key]} on:input={(event) => updateColor(field.key, event.currentTarget.value)} class="h-10 w-12 cursor-pointer" />
          <span>{field.label}<br /><span class="font-normal text-zinc-500">{colors[field.key]}</span></span>
        </label>
      {/each}
    </div>
  </section>
  <section class="mt-6 rounded-2xl border border-zinc-200 bg-white p-5 shadow-sm sm:p-6">
    <div class="flex flex-wrap items-start justify-between gap-3">
      <div>
        <h2 class="flex items-center gap-2 font-bold text-zinc-900">
          <Palette size={18} /> Cor de destaque
        </h2>
        <p class="mt-1 max-w-xl text-sm text-zinc-600">
          Usada em botões, abas, selos, carrosséis e no carrinho. Personaliza a identidade visual em todas as telas.
        </p>
      </div>

      {#if currentAccentId !== defaultAccentId}
        <button
          type="button"
          on:click={handleResetDefault}
          class="inline-flex items-center gap-1.5 rounded-xl border border-zinc-200 bg-white px-3 py-1.5 text-sm font-semibold text-zinc-700 hover:bg-zinc-50"
        >
          <RotateCcw size={14} /> Restaurar padrão
        </button>
      {/if}
    </div>

    <!-- Accent color selection -->
    <div class="mt-5 grid grid-cols-2 gap-3 sm:grid-cols-4 xl:grid-cols-7" role="radiogroup" aria-label="Cor de destaque">
      {#each accents as a (a.id)}
        {@const active = currentAccentId === a.id}
        <button
          type="button"
          role="radio"
          aria-checked={active}
          on:click={() => handleSelect(a.id)}
          class={`rounded-xl border-2 p-3 text-left transition-colors ${
            active ? 'border-zinc-900' : 'border-zinc-200 hover:border-zinc-300'
          }`}
        >
          <div class="flex h-10 overflow-hidden rounded-lg">
            <span class="flex-1" style={`background-color: ${a.shades[200]}`} />
            <span class="flex-1" style={`background-color: ${a.shades[400]}`} />
            <span class="flex-1" style={`background-color: ${a.shades[600]}`} />
          </div>
          <span class="mt-2 flex items-center justify-between text-sm font-semibold text-zinc-900">
            {a.name}
            {#if active}
              <span class="grid h-5 w-5 place-items-center rounded-full bg-zinc-900 text-white">
                <Check size={12} strokeWidth={3} />
              </span>
            {/if}
          </span>
        </button>
      {/each}
    </div>

    <!-- Preview -->
    <div class="mt-6 rounded-xl bg-zinc-100 p-4">
      <p class="mb-3 text-xs font-semibold uppercase tracking-wide text-zinc-500">Prévia</p>
      <div class="flex flex-wrap items-center gap-3">
        <button class="rounded-xl bg-yellow-400 px-4 py-2.5 text-sm font-semibold text-zinc-900 hover:bg-yellow-500">
          Finalizar Pedido
        </button>
        <span class="grid h-10 w-10 place-items-center rounded-xl bg-yellow-400 text-zinc-900">
          <Plus size={20} strokeWidth={2.5} />
        </span>
        <span class="inline-flex items-center gap-1 rounded-full bg-yellow-400 px-2.5 py-1 text-xs font-bold text-zinc-900">
          <Flame size={12} /> Mais pedido
        </span>
        <span class="rounded-full bg-yellow-400 px-4 py-2 text-sm font-semibold text-zinc-900">Lanches</span>
        <span class="rounded-full border border-zinc-200 bg-white px-4 py-2 text-sm font-semibold text-zinc-600">Bebidas</span>
        <span class="flex gap-0.5 text-yellow-400">
          {#each Array.from({ length: 5 }) as _, i}
            <Star size={16} fill="currentColor" />
          {/each}
        </span>
        <span class="rounded-lg bg-zinc-900 px-3 py-2 text-sm font-bold text-white">
          Combo <span class="text-yellow-400">com 15% off</span>
        </span>
      </div>
    </div>
  </section>
</div>
