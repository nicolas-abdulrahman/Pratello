<script lang="ts">
  import { address } from '../lib/stores/address.js';

  type StateOpt = { id: number; sigla: string; nome: string };
  type CityOpt = { id: number; nome: string };

  let { open = false, onClose = () => {} } = $props();

  let states = $state<StateOpt[]>([]);
  let cities = $state<CityOpt[]>([]);
  let selectedState = $state($address.state);
  let selectedCity = $state($address.city);
  let cep = $state('');
  let loadingStates = $state(false);
  let loadingCities = $state(false);
  let loadingCep = $state(false);
  let errorMsg = $state('');

  let wrapper = $state<HTMLDivElement | null>(null);

  // Load states once on open
  $effect(() => {
    if (open) {
      selectedState = $address.state;
      selectedCity = $address.city;
      cep = '';
      errorMsg = '';
      if (states.length === 0) loadStates();
    }
  });

  // Reload cities when state changes
  $effect(() => {
    if (selectedState) loadCities(selectedState);
  });

  // Click outside to close
  $effect(() => {
    if (!open) return;
    function onDocClick(e: MouseEvent) {
      if (wrapper && !wrapper.contains(e.target as Node)) onClose();
    }
    const t = setTimeout(() => document.addEventListener('mousedown', onDocClick), 0);
    return () => {
      clearTimeout(t);
      document.removeEventListener('mousedown', onDocClick);
    };
  });

  async function loadStates() {
    loadingStates = true;
    try {
      const res = await fetch(
        'https://servicodados.ibge.gov.br/api/v1/localidades/estados?orderBy=nome'
      );
      states = await res.json();
    } catch {
      errorMsg = 'Não foi possível carregar os estados.';
    } finally {
      loadingStates = false;
    }
  }

  async function loadCities(uf: string) {
    loadingCities = true;
    try {
      const res = await fetch(
        `https://servicodados.ibge.gov.br/api/v1/localidades/estados/${uf}/municipios`
      );
      cities = await res.json();
      if (!cities.some((c) => c.nome === selectedCity)) {
        // don't auto-pick a city — wait for user to choose
        selectedCity = '';
      }
    } catch {
      errorMsg = 'Não foi possível carregar as cidades.';
    } finally {
      loadingCities = false;
    }
  }

  // Called when the user changes the city dropdown
  function onCityChange(e: Event) {
    const city = (e.currentTarget as HTMLSelectElement).value;
    selectedCity = city;
    if (!city) return;
    const stateObj = states.find((s) => s.sigla === selectedState);
    address.setLocation({
      state: selectedState,
      stateName: stateObj?.nome ?? selectedState,
      city
    });
    onClose();
  }

  // CEP lookup via ViaCEP — fills state + city, saves, closes
  async function lookupCep() {
    const digits = cep.replace(/\D/g, '');
    if (digits.length !== 8) {
      errorMsg = 'Digite um CEP válido (8 dígitos).';
      return;
    }
    errorMsg = '';
    loadingCep = true;
    try {
      const res = await fetch(`https://viacep.com.br/ws/${digits}/json/`);
      if (!res.ok) throw new Error();
      const data = await res.json();
      if (data.erro) throw new Error();
      const uf = data.uf;
      const city = data.localidade;
      if (!uf || !city) throw new Error();
      const stateObj = states.find((s) => s.sigla === uf);
      address.setLocation({
        state: uf,
        stateName: stateObj?.nome ?? uf,
        city
      });
      onClose();
    } catch {
      errorMsg = 'CEP não encontrado. Verifique e tente novamente.';
    } finally {
      loadingCep = false;
    }
  }

  function formatCep(v: string) {
    return v
      .replace(/\D/g, '')
      .slice(0, 8)
      .replace(/(\d{5})(\d{1,3})/, '$1-$2');
  }
</script>

<div bind:this={wrapper} class="relative">
  {#if open}
    <div
      class="absolute left-0 top-full mt-2 w-80 bg-white rounded-2xl shadow-2xl border border-stone-200 z-50 p-5 space-y-4"
      role="dialog"
      aria-label="Escolher endereço"
    >
      {#if errorMsg}
        <div class="p-2.5 rounded-lg bg-red-50 text-red-700 text-[11px] font-medium border border-red-100">
          {errorMsg}
        </div>
      {/if}

      <!-- Estado / Cidade -->
      <div class="space-y-3">
        <h3 class="text-[11px] font-bold text-stone-400 uppercase tracking-wider">
          Estado / Cidade
        </h3>

        <div class="space-y-1">
          <label class="text-[11px] font-semibold text-stone-500" for="uf-select">Estado</label>
          <select
            id="uf-select"
            bind:value={selectedState}
            disabled={loadingStates}
            class="w-full bg-white border border-stone-200 rounded-lg px-3 py-2 text-sm font-medium focus:outline-none focus:ring-2 focus:ring-amber-500"
          >
            {#if loadingStates}
              <option>Carregando…</option>
            {/if}
            {#each states as s (s.sigla)}
              <option value={s.sigla}>{s.nome} ({s.sigla})</option>
            {/each}
          </select>
        </div>

        <div class="space-y-1">
          <label class="text-[11px] font-semibold text-stone-500" for="city-select">Cidade</label>
          <select
            id="city-select"
            bind:value={selectedCity}
            onchange={onCityChange}
            disabled={loadingCities || !cities.length}
            class="w-full bg-white border border-stone-200 rounded-lg px-3 py-2 text-sm font-medium focus:outline-none focus:ring-2 focus:ring-amber-500"
          >
            {#if loadingCities}
              <option>Carregando…</option>
            {:else}
              <option value="">Selecione uma cidade</option>
              {#each cities as c (c.id)}
                <option value={c.nome}>{c.nome}</option>
              {/each}
            {/if}
          </select>
        </div>
      </div>

      <!-- CEP -->
      <div class="space-y-2 pt-2 border-t border-stone-100">
        <label class="text-[11px] font-bold text-stone-400 uppercase tracking-wider" for="cep-input">
          Buscar por CEP
        </label>
        <div class="flex gap-2">
          <input
            id="cep-input"
            value={cep}
            oninput={(e) => (cep = formatCep(e.currentTarget.value))}
            onkeydown={(e) => e.key === 'Enter' && lookupCep()}
            placeholder="00000-000"
            inputmode="numeric"
            class="flex-1 min-w-0 bg-white border border-stone-200 rounded-lg px-3 py-2 text-sm font-medium tracking-wide focus:outline-none focus:ring-2 focus:ring-amber-500"
          />
          <button
            onclick={lookupCep}
            disabled={loadingCep}
            class="shrink-0 px-3 py-2 rounded-lg bg-amber-500 hover:bg-amber-600 text-white font-bold text-xs transition disabled:opacity-50 shadow-sm"
          >
            {loadingCep ? '…' : 'Buscar'}
          </button>
        </div>
      </div>

      <!-- Current selection preview -->
      <div class="pt-3 border-t border-stone-100">
        <p class="text-[10px] font-bold text-stone-400 uppercase tracking-wider mb-1">Selecionado</p>
        <p class="text-sm font-semibold text-stone-900">
          {#if $address.city}{$address.city}{#if $address.state}, {$address.state}{/if}{:else}Nenhum endereço selecionado{/if}
        </p>
      </div>
    </div>
  {/if}
</div>
