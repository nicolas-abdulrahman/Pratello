<script>
  // State for payment method
  let paymentMethod = $state('card'); // 'card' or 'pix'

  // Form State
  let cardNumber = $state('');
  let cardName = $state('');
  let expiry = $state('');
  let cvv = $state('');
  let isProcessing = $state(false);

  // Fake cart summary data
  const cartSummary = [
    { name: "Combo Casal", qty: 1, price: 55.90 },
    { name: "Brownie com Sorvete", qty: 1, price: 18.50 }
  ];

  // Derived total
  let subtotal = $derived(cartSummary.reduce((acc, item) => acc + (item.price * item.qty), 0));
  let deliveryFee = 8.50;
  let total = $derived(subtotal + deliveryFee);

  // Formatting functions
  function formatCardNumber(e) {
    let value = e.target.value.replace(/\D/g, '');
    value = value.replace(/(\d{4})(?=\d)/g, '$1 ');
    cardNumber = value;
  }

  function formatExpiry(e) {
    let value = e.target.value.replace(/\D/g, '');
    if (value.length > 2) {
      value = value.replace(/^(\d{2})(\d)/, '$1/$2');
    }
    expiry = value;
  }

  function handleSubmit(e) {
    e.preventDefault();
    isProcessing = true;
    
    // Simulate API call
    setTimeout(() => {
      isProcessing = false;
      alert(`Pagamento via ${paymentMethod === 'card' ? 'Cartão de Crédito' : 'PIX'} aprovado!`);
    }, 2000);
  }
</script>

<div class="min-h-screen bg-[#fff9f5] py-12 px-4 md:px-8 font-sans">
  <div class="max-w-6xl mx-auto">
    
    <!-- Header -->
    <div class="mb-10 text-center md:text-left">
      <h1 class="text-3xl font-black text-gray-900 mb-2">Finalizar Pagamento</h1>
      <p class="text-gray-500">Escolha como deseja pagar o seu pedido.</p>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
      
      <!-- LEFT SIDE: Payment Methods & Form -->
      <div class="lg:col-span-2 space-y-6">
        
        <!-- Payment Method Selector -->
        <div class="bg-white rounded-2xl p-2 shadow-sm border border-gray-200 flex gap-2">
          <button 
            onclick={() => paymentMethod = 'card'}
            class="flex-1 py-3 px-4 rounded-xl font-bold text-sm flex items-center justify-center gap-2 transition-all
            {paymentMethod === 'card' ? 'bg-orange-50 text-primary border border-orange-200 shadow-sm' : 'text-gray-500 hover:bg-gray-50'}"
          >
            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="20" height="14" x="2" y="5" rx="2"/><line x1="2" x2="22" y1="10" y2="10"/></svg>
            Cartão de Crédito
          </button>
          
          <button 
            onclick={() => paymentMethod = 'pix'}
            class="flex-1 py-3 px-4 rounded-xl font-bold text-sm flex items-center justify-center gap-2 transition-all
            {paymentMethod === 'pix' ? 'bg-orange-50 text-primary border border-orange-200 shadow-sm' : 'text-gray-500 hover:bg-gray-50'}"
          >
            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 2v20"/><path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"/></svg>
            PIX
          </button>
        </div>

        <!-- Main Payment Box -->
        <div class="bg-white rounded-2xl p-6 md:p-8 shadow-sm border border-gray-200">
          
          {#if paymentMethod === 'card'}
            <!-- CREDIT CARD FORM -->
            <form onsubmit={handleSubmit} class="space-y-5 animate-slide-in">
              <h2 class="text-xl font-bold text-gray-900 mb-4">Dados do Cartão</h2>
              
              <div>
                <label class="block text-xs font-semibold text-gray-700 mb-1.5">Número do Cartão</label>
                <div class="relative">
                  <input 
                    type="text" 
                    bind:value={cardNumber}
                    oninput={formatCardNumber}
                    placeholder="0000 0000 0000 0000" 
                    maxlength="19"
                    class="w-full text-sm text-gray-900 px-4 py-3.5 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:border-primary transition pl-10" 
                    required 
                  />
                  <div class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400">
                    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="20" height="14" x="2" y="5" rx="2"/><line x1="2" x2="22" y1="10" y2="10"/></svg>
                  </div>
                </div>
              </div>

              <div>
                <label class="block text-xs font-semibold text-gray-700 mb-1.5">Nome impresso no Cartão</label>
                <input 
                  type="text" 
                  bind:value={cardName}
                  placeholder="Como está no cartão" 
                  class="w-full text-sm text-gray-900 px-4 py-3.5 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:border-primary transition" 
                  required 
                />
              </div>

              <div class="grid grid-cols-2 gap-4">
                <div>
                  <label class="block text-xs font-semibold text-gray-700 mb-1.5">Validade</label>
                  <input 
                    type="text" 
                    bind:value={expiry}
                    oninput={formatExpiry}
                    placeholder="MM/AA" 
                    maxlength="5"
                    class="w-full text-sm text-gray-900 px-4 py-3.5 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:border-primary transition" 
                    required 
                  />
                </div>
                <div>
                  <label class="block text-xs font-semibold text-gray-700 mb-1.5">CVV</label>
                  <input 
                    type="text" 
                    bind:value={cvv}
                    placeholder="123" 
                    maxlength="4"
                    class="w-full text-sm text-gray-900 px-4 py-3.5 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:border-primary transition" 
                    required 
                  />
                </div>
              </div>

              <div class="pt-4">
                <button 
                  type="submit" 
                  disabled={isProcessing}
                  class="w-full py-4 bg-primary hover:bg-primary-hover text-white font-bold text-base rounded-xl shadow-lg shadow-orange-500/30 transition-all flex items-center justify-center gap-2 disabled:opacity-70"
                >
                  {#if isProcessing}
                    <svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path></svg>
                    Processando...
                  {:else}
                    <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>
                    Pagar R$ {total.toFixed(2).replace('.', ',')}
                  {/if}
                </button>
              </div>
            </form>

          {:else}
            <!-- PIX PAYMENT -->
            <div class="flex flex-col items-center text-center animate-slide-in py-4">
              <h2 class="text-xl font-bold text-gray-900 mb-2">Pague com PIX</h2>
              <p class="text-sm text-gray-500 mb-6 max-w-sm">Escaneie o código QR abaixo com o aplicativo do seu banco para finalizar o pagamento.</p>
              
              <!-- Fake QR Code Box -->
              <div class="p-4 bg-white border-2 border-dashed border-primary/40 rounded-3xl mb-6 shadow-sm">
                <div class="w-48 h-48 bg-gray-100 rounded-xl flex items-center justify-center relative overflow-hidden">
                  <!-- Fake QR pattern -->
                  <svg class="w-full h-full text-gray-800 opacity-80" viewBox="0 0 100 100" fill="currentColor">
                    <rect x="10" y="10" width="20" height="20" />
                    <rect x="15" y="15" width="10" height="10" fill="white" />
                    <rect x="70" y="10" width="20" height="20" />
                    <rect x="75" y="15" width="10" height="10" fill="white" />
                    <rect x="10" y="70" width="20" height="20" />
                    <rect x="15" y="75" width="10" height="10" fill="white" />
                    <rect x="40" y="10" width="10" height="10" />
                    <rect x="50" y="20" width="10" height="10" />
                    <rect x="40" y="40" width="20" height="20" />
                    <rect x="10" y="40" width="10" height="20" />
                    <rect x="70" y="40" width="20" height="10" />
                    <rect x="40" y="70" width="10" height="10" />
                    <rect x="60" y="60" width="20" height="20" />
                    <rect x="80" y="80" width="10" height="10" />
                  </svg>
                  <!-- Center Logo -->
                  <div class="absolute inset-0 flex items-center justify-center">
                    <div class="w-10 h-10 bg-white rounded-lg flex items-center justify-center shadow-md">
                      <span class="text-primary font-black text-lg">P</span>
                    </div>
                  </div>
                </div>
              </div>

              <div class="w-full max-w-xs mb-6">
                <label class="block text-xs font-semibold text-gray-700 mb-1.5 text-left">Ou copie e cole o código:</label>
                <div class="flex gap-2">
                  <input type="text" readonly value="00020126580014br.gov.bcb.pix..." class="w-full text-xs text-gray-500 px-3 py-2.5 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none" />
                  <button class="px-4 bg-gray-100 hover:bg-gray-200 text-gray-700 font-bold text-xs rounded-xl transition">Copiar</button>
                </div>
              </div>

              <div class="w-full max-w-xs">
                <button class="w-full py-4 bg-primary hover:bg-primary-hover text-white font-bold text-base rounded-xl shadow-lg shadow-orange-500/30 transition-all">
                  Já realizei o pagamento
                </button>
              </div>
            </div>
          {/if}

          <!-- Security Badge -->
          <div class="mt-8 pt-6 border-t border-gray-100 flex items-center justify-center gap-2 text-xs text-gray-400">
            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>
            Pagamento 100% seguro e criptografado
          </div>
        </div>
      </div>

      <!-- RIGHT SIDE: Order Summary -->
      <div class="lg:col-span-1">
        <div class="bg-white rounded-2xl p-6 shadow-sm border border-gray-200 sticky top-8">
          <h3 class="text-lg font-bold text-gray-900 mb-4">Resumo do Pedido</h3>
          
          <div class="space-y-3 mb-6">
            {#each cartSummary as item}
              <div class="flex justify-between items-start text-sm">
                <div class="flex items-start gap-2 text-gray-600">
                  <span class="font-bold text-primary">{item.qty}x</span>
                  <span>{item.name}</span>
                </div>
                <span class="font-semibold text-gray-900">R$ {item.price.toFixed(2).replace('.', ',')}</span>
              </div>
            {/each}
          </div>

          <div class="space-y-2 pt-4 border-t border-gray-100 text-sm">
            <div class="flex justify-between text-gray-500">
              <span>Subtotal</span>
              <span>R$ {subtotal.toFixed(2).replace('.', ',')}</span>
            </div>
            <div class="flex justify-between text-gray-500">
              <span>Taxa de entrega</span>
              <span>R$ {deliveryFee.toFixed(2).replace('.', ',')}</span>
            </div>
          </div>

          <div class="flex justify-between items-center pt-4 mt-4 border-t border-gray-100">
            <span class="text-base font-bold text-gray-900">Total</span>
            <span class="text-2xl font-black text-primary">R$ {total.toFixed(2).replace('.', ',')}</span>
          </div>
        </div>
      </div>

    </div>
  </div>
</div>

<style>
  .animate-slide-in {
    animation: slideIn 0.3s ease-out forwards;
  }
  @keyframes slideIn {
    0% { opacity: 0; transform: translateY(10px); }
    100% { opacity: 1; transform: translateY(0); }
  }
</style>
