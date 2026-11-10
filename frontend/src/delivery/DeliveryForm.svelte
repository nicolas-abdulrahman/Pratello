<script>
  export let onSubmit = () => {};
  
  let cep = "";
  let phone = "";
  let street = "";
  let number = "";
  let complement = "";
  let neighborhood = "";
  let city = "";

  function handleSubmit(e) {
    e.preventDefault();
    // Pass the form data back up
    onSubmit({ cep, phone, street, number, complement, neighborhood, city });
  }
</script>

<div class="bg-white border border-gray-200 rounded-2xl p-6 shadow-sm mt-6 animate-slide-in">
  <div class="flex items-center gap-2 mb-6">
    <div class="w-8 h-8 bg-orange-50 rounded-full flex items-center justify-center text-primary">
      <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 10c0 6-8 12-8 12s-8-6-8-12a8 8 0 0 1 16 0Z"/><circle cx="12" cy="10" r="3"/></svg>
    </div>
    <h2 class="text-xl font-bold text-gray-900">Endereço de Entrega</h2>
  </div>

  <form onsubmit={handleSubmit} class="space-y-4">
    
    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <div>
        <label class="block text-xs font-semibold text-gray-700 mb-1">CEP</label>
        <input type="text" bind:value={cep} placeholder="00000-000" class="w-full text-sm text-gray-900 px-3 py-2.5 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:border-primary transition" required />
      </div>
      <div>
        <label class="block text-xs font-semibold text-gray-700 mb-1">Telefone / WhatsApp</label>
        <input type="tel" bind:value={phone} placeholder="(81) 99999-9999" class="w-full text-sm text-gray-900 px-3 py-2.5 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:border-primary transition" required />
      </div>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
      <div class="md:col-span-2">
        <label class="block text-xs font-semibold text-gray-700 mb-1">Rua / Avenida</label>
        <input type="text" bind:value={street} placeholder="Ex: Rua das Flores" class="w-full text-sm text-gray-900 px-3 py-2.5 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:border-primary transition" required />
      </div>
      <div>
        <label class="block text-xs font-semibold text-gray-700 mb-1">Número</label>
        <input type="text" bind:value={number} placeholder="123" class="w-full text-sm text-gray-900 px-3 py-2.5 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:border-primary transition" required />
      </div>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <div>
        <label class="block text-xs font-semibold text-gray-700 mb-1">Complemento (Opcional)</label>
        <input type="text" bind:value={complement} placeholder="Apto 101, Bloco B" class="w-full text-sm text-gray-900 px-3 py-2.5 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:border-primary transition" />
      </div>
      <div>
        <label class="block text-xs font-semibold text-gray-700 mb-1">Bairro</label>
        <input type="text" bind:value={neighborhood} placeholder="Ex: Boa Viagem" class="w-full text-sm text-gray-900 px-3 py-2.5 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:border-primary transition" required />
      </div>
    </div>

    <div class="pt-2">
      <label class="block text-xs font-semibold text-gray-700 mb-1">Cidade</label>
      <input type="text" bind:value={city} placeholder="Recife" class="w-full text-sm text-gray-900 px-3 py-2.5 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:border-primary transition" required />
    </div>

    <button type="submit" class="w-full mt-4 py-3 bg-primary hover:bg-primary-hover text-white font-semibold text-sm rounded-xl shadow-md transition">
      Confirmar Endereço e Pagar
    </button>
  </form>
</div>

<style>
  .animate-slide-in {
    animation: slideIn 0.4s cubic-bezier(0.25, 1, 0.5, 1) forwards;
  }
  @keyframes slideIn {
    0% { opacity: 0; transform: translateY(-10px); }
    100% { opacity: 1; transform: translateY(0); }
  }
</style>
