<script>
  import { onMount } from 'svelte';
  import { account } from '../lib/stores/account';

  export let onGoHome = () => {};

  // Mode: 'login' | 'register'
  let authMode = 'login';

  // Role toggle: 'customer' | 'employee'
  let loginType = 'customer';

  // Form fields
  let email = '';
  let password = '';
  let name = '';
  let phone = '';
  let rememberMe = true;
  let showPassword = false;

  // Scroll down slightly when the login page loads
  onMount(() => {
    window.scrollTo({
      top: 100,
      behavior: 'smooth'
    });
  });

  function toggleMode(mode) {
    authMode = mode;
    account.clearError();
  }

  async function handleSubmit(e) {
    e.preventDefault();
    account.clearError();

    if (authMode === 'login') {
      const ok = await account.login({
        email,
        password,
        name: name.trim() || undefined,
        role: loginType === 'employee' ? 'funcionario' : 'cliente'
      });
      if (ok) {
        onGoHome();
      }
    } else {
      const ok = await account.register({
        email,
        password,
        phone: phone.trim() || undefined,
        name: name.trim() || undefined
      });
      if (ok) {
        onGoHome();
      }
    }
  }
</script>

<div class="min-h-screen w-full flex bg-white font-sans selection:bg-[#f05a1f] selection:text-white">
  
  <!-- LEFT SIDE: Image & Quote -->
  <div class="hidden lg:flex w-1/2 relative bg-gray-900">
    <img 
      src="https://images.unsplash.com/photo-1579684947550-22e945225d9a?ixlib=rb-4.0.3&auto=format&fit=crop&w=1200&q=80" 
      alt="Pizza e pratos em um restaurante" 
      class="absolute inset-0 w-full h-full object-cover opacity-80"
    />
    <!-- Dark gradient overlay for text readability -->
    <div class="absolute inset-0 bg-gradient-to-t from-black/90 via-black/40 to-transparent"></div>
    
    <!-- Text Content -->
    <div class="relative z-10 flex flex-col justify-end p-16 text-white w-full max-w-2xl">
      <div class="mb-4">
        <span class="inline-block bg-white/20 backdrop-blur-md text-white text-xs font-bold px-3 py-1.5 rounded-md border border-white/30 tracking-wider uppercase">
          Sua Barriga Merece o Melhor
        </span>
      </div>
      <h1 class="text-5xl font-bold leading-[1.1] mb-4 tracking-tight">
        Comida boa. Companhia perfeita.
      </h1>
      <p class="text-gray-300 text-lg max-w-md">
        As melhores cozinhas do seu bairro aos momentos que importam.
      </p>
    </div>
  </div>

  <!-- RIGHT SIDE: Auth Form -->
  <div class="w-full lg:w-1/2 flex items-center justify-center p-8 sm:p-12 lg:p-24 relative">
    
    <!-- Back to Home Button -->
    <button 
      onclick={onGoHome}
      type="button"
      class="absolute top-8 left-8 lg:left-12 flex items-center gap-2 text-gray-500 hover:text-[#f05a1f] font-medium transition-colors cursor-pointer"
    >
      <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m15 18-6-6 6-6"/></svg>
      Voltar
    </button>

    <div class="w-full max-w-md space-y-6">
      
      <!-- Header -->
      <div>
        <span class="text-[#f05a1f] font-bold text-xs tracking-wider uppercase mb-2 block">Bem-vindo ao Pratello</span>
        <h2 class="text-3xl font-bold text-gray-900 mb-2">
          {authMode === 'login' ? 'Sua próxima refeição começa aqui.' : 'Crie sua conta global no Pratello.'}
        </h2>
        <p class="text-gray-500 text-sm">
          {authMode === 'login' 
            ? 'Faça login para acessar seus menus favoritos, finalizar pedidos e gerenciar estabelecimentos.'
            : 'Cadastre sua conta para fazer pedidos, acumular pontos ou gerenciar restaurantes.'}
        </p>
      </div>

      <!-- Mode Toggle (Entrar / Cadastrar) -->
      <div class="flex border-b border-gray-200">
        <button 
          type="button"
          onclick={() => toggleMode('login')}
          class="flex-1 pb-3 text-sm font-bold text-center border-b-2 transition-all cursor-pointer {authMode === 'login' ? 'border-[#f05a1f] text-[#f05a1f]' : 'border-transparent text-gray-400 hover:text-gray-600'}"
        >
          Entrar
        </button>
        <button 
          type="button"
          onclick={() => toggleMode('register')}
          class="flex-1 pb-3 text-sm font-bold text-center border-b-2 transition-all cursor-pointer {authMode === 'register' ? 'border-[#f05a1f] text-[#f05a1f]' : 'border-transparent text-gray-400 hover:text-gray-600'}"
        >
          Cadastrar
        </button>
      </div>

      <!-- Login Type Toggle (Customer vs Employee) -->
      {#if authMode === 'login'}
        <div class="flex bg-gray-100 p-1 rounded-lg">
          <button 
            type="button"
            onclick={() => loginType = 'customer'}
            class="flex-1 py-2 text-sm font-bold rounded-md transition-all cursor-pointer {loginType === 'customer' ? 'bg-white text-gray-900 shadow-sm' : 'text-gray-500 hover:text-gray-700'}"
          >
            Cliente
          </button>
          <button 
            type="button"
            onclick={() => loginType = 'employee'}
            class="flex-1 py-2 text-sm font-bold rounded-md transition-all cursor-pointer {loginType === 'employee' ? 'bg-white text-gray-900 shadow-sm' : 'text-gray-500 hover:text-gray-700'}"
          >
            Equipe / Gestor
          </button>
        </div>
      {/if}

      <!-- ERROR / WARNING BANNER -->
      {#if $account.error}
        <div 
          role="alert" 
          aria-live="assertive"
          class="flex items-start gap-3 p-4 rounded-lg bg-red-50 border border-red-200 text-red-800 text-sm animate-fade-in shadow-sm"
        >
          <svg class="w-5 h-5 text-red-600 flex-shrink-0 mt-0.5" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
          </svg>
          <div class="flex-1">
            <p class="font-semibold text-red-900">Atenção</p>
            <p class="mt-0.5">{$account.error}</p>
          </div>
          <button 
            type="button"
            onclick={() => account.clearError()}
            class="text-red-400 hover:text-red-700 cursor-pointer p-0.5 rounded transition-colors"
            title="Fechar aviso"
          >
            <svg class="w-4 h-4" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
            </svg>
          </button>
        </div>
      {/if}

      <!-- Form -->
      <form onsubmit={handleSubmit} class="space-y-4">
        
        {#if authMode === 'register'}
          <div>
            <label for="name" class="block text-sm font-medium text-gray-700 mb-1">Nome Completo</label>
            <input 
              type="text" 
              id="name" 
              bind:value={name}
              autocomplete="name"
              placeholder="Ex: Carlos Silva"
              class="w-full px-4 py-3 rounded-lg border border-gray-300 focus:ring-2 focus:ring-[#f05a1f]/20 focus:border-[#f05a1f] outline-none transition-all"
            />
          </div>
        {/if}

        <div>
          <label for="email" class="block text-sm font-medium text-gray-700 mb-1">Endereço de Email</label>
          <input 
            type="email" 
            id="email" 
            bind:value={email}
            autocomplete="email"
            placeholder="voce@exemplo.com"
            class="w-full px-4 py-3 rounded-lg border border-gray-300 focus:ring-2 focus:ring-[#f05a1f]/20 focus:border-[#f05a1f] outline-none transition-all"
            required
          />
        </div>

        {#if authMode === 'register'}
          <div>
            <label for="phone" class="block text-sm font-medium text-gray-700 mb-1">Telefone / WhatsApp</label>
            <input 
              type="tel" 
              id="phone" 
              bind:value={phone}
              autocomplete="tel"
              placeholder="(11) 99999-9999"
              class="w-full px-4 py-3 rounded-lg border border-gray-300 focus:ring-2 focus:ring-[#f05a1f]/20 focus:border-[#f05a1f] outline-none transition-all"
            />
          </div>
        {/if}

        <div>
          <label for="password" class="block text-sm font-medium text-gray-700 mb-1">Senha</label>
          <div class="relative">
            <input 
              type={showPassword ? 'text' : 'password'} 
              id="password" 
              bind:value={password}
              autocomplete={authMode === 'login' ? 'current-password' : 'new-password'}
              placeholder="Sua senha secreta"
              class="w-full px-4 py-3 rounded-lg border border-gray-300 focus:ring-2 focus:ring-[#f05a1f]/20 focus:border-[#f05a1f] outline-none transition-all pr-12"
              required
            />
            <button 
              type="button" 
              onclick={() => showPassword = !showPassword}
              class="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600 cursor-pointer p-1"
              title={showPassword ? 'Ocultar senha' : 'Exibir senha'}
            >
              {#if showPassword}
                <!-- Eye Off Icon -->
                <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9.88 9.88a3 3 0 1 0 4.24 4.24"/><path d="M10.73 5.08A10.43 10.43 0 0 1 12 5c7 0 10 7 10 7a13.16 13.16 0 0 1-1.67 2.68"/><path d="M6.61 6.61A13.526 13.526 0 0 0 2 12s3 7 10 7a9.74 9.74 0 0 0 5.39-1.61"/><line x1="2" x2="22" y1="2" y2="22"/></svg>
              {:else}
                <!-- Eye Icon -->
                <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M2 12s3-7 10-7 10 7 10 7-3 7-10 7-10-7-10-7Z"/><circle cx="12" cy="12" r="3"/></svg>
              {/if}
            </button>
          </div>
        </div>

        {#if authMode === 'login'}
          <div class="flex items-center justify-between">
            <div class="flex items-center">
              <input 
                id="remember-me" 
                type="checkbox" 
                bind:checked={rememberMe}
                class="h-4 w-4 text-[#f05a1f] focus:ring-[#f05a1f] border-gray-300 rounded cursor-pointer"
              />
              <label for="remember-me" class="ml-2 block text-sm text-gray-700 cursor-pointer">
                Lembrar de mim
              </label>
            </div>

            <div class="text-sm">
              <button 
                type="button" 
                onclick={() => account.setError('Para redefinir sua senha, solicite suporte ao administrador do sistema.')}
                class="font-bold text-[#f05a1f] hover:text-[#d94e17] cursor-pointer"
              >
                Esqueceu a senha?
              </button>
            </div>
          </div>
        {/if}

        <button 
          type="submit" 
          disabled={$account.loading}
          class="w-full bg-[#f05a1f] hover:bg-[#d94e17] disabled:opacity-50 disabled:cursor-not-allowed text-white font-bold py-3.5 rounded-lg transition-colors shadow-lg shadow-orange-500/20 cursor-pointer flex items-center justify-center gap-2"
        >
          {#if $account.loading}
            <svg class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            <span>{authMode === 'login' ? 'Autenticando...' : 'Criando conta...'}</span>
          {:else}
            <span>{authMode === 'login' ? 'Entrar' : 'Criar Conta'}</span>
          {/if}
        </button>
      </form>

      <!-- Toggle mode link -->
      {#if authMode === 'login'}
        <p class="text-center text-sm text-gray-600">
          Novo aqui? 
          <button 
            type="button" 
            onclick={() => toggleMode('register')} 
            class="font-bold text-[#f05a1f] hover:text-[#d94e17] cursor-pointer ml-1"
          >
            Crie uma conta
          </button>
        </p>
      {:else}
        <p class="text-center text-sm text-gray-600">
          Já possui uma conta? 
          <button 
            type="button" 
            onclick={() => toggleMode('login')} 
            class="font-bold text-[#f05a1f] hover:text-[#d94e17] cursor-pointer ml-1"
          >
            Faça login
          </button>
        </p>
      {/if}

      <p class="text-center text-xs text-gray-400">
        Ao continuar, você concorda com nossos <span class="underline hover:text-gray-600 cursor-pointer">Termos de Serviço</span> e <span class="underline hover:text-gray-600 cursor-pointer">Política de Privacidade</span>.
      </p>

    </div>
  </div>
</div>
