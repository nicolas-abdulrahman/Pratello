<script lang="ts">
  import {
    Armchair,
    Bike,
    Check,
    Loader2,
    MapPin,
    MessageCircle
  } from 'lucide-svelte';
  import type { Address, CheckoutMode } from '../checkout';

  interface Props {
    mode?: CheckoutMode;
    table?: string;
    address?: Address;
    phone?: string;
    tableCount?: number;
    /** true when the customer came through /r/{slug}/dine-in: only the table is asked. */
    dineIn?: boolean;
    fetchAddressFromCep?: (cep: string) => Promise<{ street?: string; district?: string } | null>;
    onModeChange?: (mode: CheckoutMode) => void;
    onAddressChange?: (address: Address) => void;
    onPhoneChange?: (phone: string) => void;
    onTableChange?: (table: string) => void;
    onClearError?: () => void;
  }

  let {
    mode = $bindable('mesa'),
    table = $bindable(''),
    address = $bindable({ cep: '', street: '', number: '', district: '', complement: '' }),
    phone = $bindable(''),
    tableCount = 15,
    dineIn = false,
    fetchAddressFromCep,
    onModeChange = () => {},
    onAddressChange = () => {},
    onPhoneChange = () => {},
    onTableChange = () => {},
    onClearError = () => {}
  }: Props = $props();

  let cepLoading = $state(false);

  // Dine-in is always "na mesa", regardless of what the parent holds.
  let effectiveMode = $derived<CheckoutMode>(dineIn ? 'mesa' : mode);

  const MODES: { id: CheckoutMode; label: string; sub: string; icon: any }[] = [
    { id: 'mesa', label: 'Na mesa', sub: 'Receba no salão', icon: Armchair },
    { id: 'delivery', label: 'Delivery', sub: 'Entrega pelo app', icon: Bike },
    { id: 'whatsapp', label: 'WhatsApp', sub: 'Envie o pedido no Zap', icon: MessageCircle },
  ];

  const maskPhone = (v: string) => {
    const d = v.replace(/\D/g, '').slice(0, 11);
    if (d.length <= 2) return d;
    if (d.length <= 6) return `(${d.slice(0, 2)}) ${d.slice(2)}`;
    if (d.length <= 10) return `(${d.slice(0, 2)}) ${d.slice(2, 6)}-${d.slice(6)}`;
    return `(${d.slice(0, 2)}) ${d.slice(2, 7)}-${d.slice(7)}`;
  };

  const maskCep = (v: string) => {
    const d = v.replace(/\D/g, '').slice(0, 8);
    return d.length > 5 ? `${d.slice(0, 5)}-${d.slice(5)}` : d;
  };

  async function handleCepInput(e: Event) {
    const target = e.target as HTMLInputElement;
    const masked = maskCep(target.value);
    address.cep = masked;
    onClearError();
    onAddressChange(address);

    const digits = masked.replace(/\D/g, '');
    if (digits.length === 8) {
      cepLoading = true;
      if (fetchAddressFromCep) {
        const res = await fetchAddressFromCep(digits);
        cepLoading = false;
        if (res) {
          if (res.street) address.street = res.street;
          if (res.district) address.district = res.district;
          onAddressChange(address);
        }
      } else {
        try {
          const res = await fetch(`https://viacep.com.br/ws/${digits}/json/`);
          const data = await res.json();
          cepLoading = false;
          if (!data.erro) {
            address.street = data.logradouro || address.street;
            address.district = data.bairro || address.district;
            onAddressChange(address);
          }
        } catch {
          cepLoading = false;
        }
      }
    }
  }

  function handlePhoneInput(e: Event) {
    const target = e.target as HTMLInputElement;
    phone = maskPhone(target.value);
    onClearError();
    onPhoneChange(phone);
  }
</script>

<section class="overflow-hidden rounded-2xl bg-white shadow-sm ring-1 ring-zinc-200">
  <!-- Envelope candy-striped border top header -->
  <div
    class="h-1.5"
    style="background-image: repeating-linear-gradient(-45deg, #facc15 0 14px, transparent 14px 20px, #27272a 20px 34px, transparent 34px 40px);"
  ></div>

  <div class="p-4 sm:p-5">
    {#if dineIn}
      <!-- Dine-in: no mode options, only the table -->
      <h2 class="mb-3 flex items-center gap-2 font-bold text-zinc-900">
        <Armchair size={18} class="text-yellow-600" /> Qual é a sua mesa?
      </h2>
    {:else}
      <h2 class="mb-3 flex items-center gap-2 font-bold text-zinc-900">
        <MapPin size={18} class="text-yellow-600" /> Como você quer pedir?
      </h2>

      <!-- Mode Selector Cards -->
      <div class="grid grid-cols-3 gap-2" role="radiogroup" aria-label="Forma de pedido">
        {#each MODES as item (item.id)}
          {@const active = mode === item.id}
          <button
            type="button"
            role="radio"
            aria-checked={active}
            onclick={() => {
              mode = item.id;
              onClearError();
              onModeChange(mode);
            }}
            class={`relative flex flex-col items-center gap-1 rounded-xl border-2 px-2 py-3 text-center transition-colors ${
              active ? 'border-yellow-400 bg-yellow-50' : 'border-zinc-200 hover:border-zinc-300'
            }`}
          >
            {#if active}
              <span class="absolute right-1.5 top-1.5 grid h-4 w-4 place-items-center rounded-full bg-yellow-400">
                <Check size={10} strokeWidth={3} />
              </span>
            {/if}
            <svelte:component this={item.icon} size={22} class={active ? 'text-zinc-900' : 'text-zinc-500'} />
            <span class="text-sm font-bold text-zinc-900">{item.label}</span>
            <span class="hidden text-[11px] leading-tight text-zinc-500 sm:block">{item.sub}</span>
          </button>
        {/each}
      </div>
    {/if}

    <div class={dineIn ? '' : 'mt-4'}>
      {#if effectiveMode === 'mesa'}
        <label class="block">
          <span class="mb-1.5 block text-sm font-medium text-zinc-700">Número da mesa</span>
          <select
            bind:value={table}
            onchange={() => {
              onClearError();
              onTableChange(table);
            }}
            class="w-full rounded-xl border border-zinc-200 bg-white px-3.5 py-2.5 text-sm text-zinc-900 focus:border-yellow-400 focus:outline-none focus:ring-2 focus:ring-yellow-400/40"
          >
            <option value="">Sem mesa</option>
            {#each Array.from({ length: tableCount }, (_, i) => String(i + 1).padStart(2, '0')) as t}
              <option value={t}>Mesa {t}</option>
            {/each}
          </select>
        </label>
      {:else}
        <div class="space-y-3">
          {#if mode === 'whatsapp'}
            <p class="flex gap-2 rounded-xl bg-green-50 p-3 text-sm text-green-900 ring-1 ring-green-200">
              <MessageCircle size={18} class="mt-0.5 shrink-0" />
              Seu pedido vai pronto para o WhatsApp do restaurante. É só confirmar o envio e acompanhar pela conversa.
            </p>
          {/if}

          <!-- Address Fields -->
          <div class="grid grid-cols-6 gap-3">
            <div class="col-span-3">
              <label class="block">
                <span class="mb-1.5 block text-sm font-medium text-zinc-700">CEP</span>
                <div class="relative">
                  <input
                    type="text"
                    inputmode="numeric"
                    value={address.cep}
                    oninput={handleCepInput}
                    placeholder="00000-000"
                    class="w-full rounded-xl border border-zinc-200 bg-white px-3.5 py-2.5 text-sm text-zinc-900 focus:border-yellow-400 focus:outline-none"
                  />
                  {#if cepLoading}
                    <Loader2 size={16} class="absolute right-3 top-1/2 -translate-y-1/2 animate-spin text-zinc-400" />
                  {/if}
                </div>
              </label>
            </div>
            <div class="col-span-3">
              <label class="block">
                <span class="mb-1.5 block text-sm font-medium text-zinc-700">Número</span>
                <input
                  type="text"
                  bind:value={address.number}
                  oninput={() => {
                    onClearError();
                    onAddressChange(address);
                  }}
                  placeholder="123"
                  class="w-full rounded-xl border border-zinc-200 bg-white px-3.5 py-2.5 text-sm text-zinc-900 focus:border-yellow-400 focus:outline-none"
                />
              </label>
            </div>
            <div class="col-span-6">
              <label class="block">
                <span class="mb-1.5 block text-sm font-medium text-zinc-700">Rua</span>
                <input
                  type="text"
                  bind:value={address.street}
                  oninput={() => {
                    onClearError();
                    onAddressChange(address);
                  }}
                  placeholder="Rua das Flores"
                  class="w-full rounded-xl border border-zinc-200 bg-white px-3.5 py-2.5 text-sm text-zinc-900 focus:border-yellow-400 focus:outline-none"
                />
              </label>
            </div>
            <div class="col-span-6 sm:col-span-3">
              <label class="block">
                <span class="mb-1.5 block text-sm font-medium text-zinc-700">Bairro</span>
                <input
                  type="text"
                  bind:value={address.district}
                  oninput={() => {
                    onClearError();
                    onAddressChange(address);
                  }}
                  class="w-full rounded-xl border border-zinc-200 bg-white px-3.5 py-2.5 text-sm text-zinc-900 focus:border-yellow-400 focus:outline-none"
                />
              </label>
            </div>
            <div class="col-span-6 sm:col-span-3">
              <label class="block">
                <span class="mb-1.5 block text-sm font-medium text-zinc-700">Complemento</span>
                <input
                  type="text"
                  bind:value={address.complement}
                  oninput={() => onAddressChange(address)}
                  placeholder="Apto, bloco…"
                  class="w-full rounded-xl border border-zinc-200 bg-white px-3.5 py-2.5 text-sm text-zinc-900 focus:border-yellow-400 focus:outline-none"
                />
              </label>
            </div>
          </div>

          <label class="block">
            <span class="mb-1.5 block text-sm font-medium text-zinc-700">
              {mode === 'whatsapp' ? 'Seu WhatsApp' : 'Telefone para contato'}
            </span>
            <input
              type="tel"
              value={phone}
              oninput={handlePhoneInput}
              placeholder="(11) 99999-0000"
              class="w-full rounded-xl border border-zinc-200 bg-white px-3.5 py-2.5 text-sm text-zinc-900 focus:border-yellow-400 focus:outline-none"
            />
          </label>
        </div>
      {/if}
    </div>
  </div>
</section>
