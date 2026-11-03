<script>
  import { onMount } from 'svelte';

  export let onGoHome = () => {};

  // State to toggle between Customer and Employee login
  let loginType = 'customer'; 
  
  // Form state
  let email = '';
  let password = '';
  let rememberMe = true;

  // Scroll down slightly when the login page loads
  onMount(() => {
    // 80px is roughly the height of your navbar. 
    // We scroll just enough to push it out of view.
    window.scrollTo({
      top: 100,
      behavior: 'smooth'
    });
  });

  function handleLogin(e) {
    e.preventDefault();
    console.log(`Logging in as ${loginType} with email: ${email}`);
    // Add your authentication logic here
  }
</script>

<div class="min-h-screen w-full flex bg-white font-sans selection:bg-[#f05a1f] selection:text-white">
  
  <!-- LEFT SIDE: Image & Quote -->
  <div class="hidden lg:flex w-1/2 relative bg-gray-900">
    <img 
      src="https://images.unsplash.com/photo-1579684947550-22e945225d9a?ixlib=rb-4.0.3&auto=format&fit=crop&w=1200&q=80" 
      alt="Pizza and Pasta on a table" 
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

  <!-- RIGHT SIDE: Login Form -->
  <div class="w-full lg:w-1/2 flex items-center justify-center p-8 sm:p-12 lg:p-24 relative">
    
    <!-- Back to Home Button -->
    <button 
      onclick={onGoHome}
      class="absolute top-8 left-8 lg:left-12 flex items-center gap-2 text-gray-500 hover:text-[#f05a1f] font-medium transition-colors"
    >
      <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m15 18-6-6 6-6"/></svg>
      Voltar
    </button>

    <div class="w-full max-w-md space-y-8">
      
      <!-- Header -->
      <div>
        <span class="text-[#f05a1f] font-bold text-xs tracking-wider uppercase mb-2 block">Bem-vindo ao Petrello</span>
        <h2 class="text-3xl font-bold text-gray-900 mb-2">Sua próxima refeição começa aqui.</h2>
        <p class="text-gray-500 text-sm">
          Faça login para acessar seus menus favoritos, finalizar a compra mais rápido e ganhar um mimo só para você.
        </p>
      </div>

      <!-- Login Type Toggle -->
      <div class="flex bg-gray-100 p-1 rounded-lg">
        <button 
          onclick={() => loginType = 'customer'}
          class="flex-1 py-2 text-sm font-bold rounded-md transition-all {loginType === 'customer' ? 'bg-white text-gray-900 shadow-sm' : 'text-gray-500 hover:text-gray-700'}"
        >
          Cliente
        </button>
        <button 
          onclick={() => loginType = 'employee'}
          class="flex-1 py-2 text-sm font-bold rounded-md transition-all {loginType === 'employee' ? 'bg-white text-gray-900 shadow-sm' : 'text-gray-500 hover:text-gray-700'}"
        >
          Funcionário
        </button>
      </div>

      <!-- Google Login Button -->
      <button class="w-full flex items-center justify-center gap-3 bg-white border border-gray-300 text-gray-700 font-medium py-3 rounded-lg hover:bg-gray-50 transition-colors shadow-sm">
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 48 48" width="20px" height="20px"><path fill="#FFC107" d="M43.611,20.083H42V20H24v8h11.303c-1.649,4.657-6.08,8-11.303,8c-6.627,0-12-5.373-12-12c0-6.627,5.373-12,12-12c3.059,0,5.842,1.154,7.961,3.039l5.657-5.657C34.046,6.053,29.268,4,24,4C12.955,4,4,12.955,4,24c0,11.045,8.955,20,20,20c11.045,0,20-8.955,20-20C44,22.659,43.862,21.35,43.611,20.083z"/><path fill="#FF3D00" d="M6.306,14.691l6.571,4.819C14.655,15.108,18.961,12,24,12c3.059,0,5.842,1.154,7.961,3.039l5.657-5.657C34.046,6.053,29.268,4,24,4C16.318,4,9.656,8.337,6.306,14.691z"/><path fill="#4CAF50" d="M24,44c5.166,0,9.86-1.977,13.409-5.192l-6.19-5.238C29.211,35.091,26.715,36,24,36c-5.202,0-9.619-3.317-11.283-7.946l-6.522,5.025C9.505,39.556,16.227,44,24,44z"/><path fill="#1976D2" d="M43.611,20.083H42V20H24v8h11.303c-0.792,2.237-2.231,4.166-4.087,5.571c0.001-0.001,0.002-0.001,0.003-0.002l6.19,5.238C36.971,39.205,44,34,44,24C44,22.659,43.862,21.35,43.611,20.083z"/></svg>
        Continuar com Google
      </button>

      <!-- Divider -->
      <div class="relative flex items-center">
        <div class="flex-grow border-t border-gray-200"></div>
        <span class="flex-shrink-0 mx-4 text-gray-400 text-xs uppercase">Ou cadastre-se com email</span>
        <div class="flex-grow border-t border-gray-200"></div>
      </div>

      <!-- Email/Password Form -->
      <form onsubmit={handleLogin} class="space-y-5">
        
        <div>
          <label for="email" class="block text-sm font-medium text-gray-700 mb-1">Endereço de Email</label>
          <input 
            type="email" 
            id="email" 
            bind:value={email}
            placeholder="voce@exemplo.com"
            class="w-full px-4 py-3 rounded-lg border border-gray-300 focus:ring-2 focus:ring-[#f05a1f]/20 focus:border-[#f05a1f] outline-none transition-all"
            required
          />
        </div>

        <div>
          <label for="password" class="block text-sm font-medium text-gray-700 mb-1">Senha</label>
          <div class="relative">
            <input 
              type="password" 
              id="password" 
              bind:value={password}
              placeholder="Sua senha"
              class="w-full px-4 py-3 rounded-lg border border-gray-300 focus:ring-2 focus:ring-[#f05a1f]/20 focus:border-[#f05a1f] outline-none transition-all pr-12"
              required
            />
            <button type="button" class="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600">
              <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M2 12s3-7 10-7 10 7 10 7-3 7-10 7-10-7-10-7Z"/><circle cx="12" cy="12" r="3"/></svg>
            </button>
          </div>
        </div>

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
            <a href="#forgot" class="font-bold text-[#f05a1f] hover:text-[#d94e17]">
              Esqueceu a senha?
            </a>
          </div>
        </div>

        <button 
          type="submit" 
          class="w-full bg-[#f05a1f] hover:bg-[#d94e17] text-white font-bold py-3.5 rounded-lg transition-colors shadow-lg shadow-orange-500/20"
        >
          Entrar
        </button>
      </form>

      <p class="text-center text-sm text-gray-600">
        Novo aqui? 
        <a href="#signup" class="font-bold text-[#f05a1f] hover:text-[#d94e17]">Crie uma conta</a>
      </p>

      <p class="text-center text-xs text-gray-400">
        Ao continuar, você concorda com nossos <a href="#terms" class="underline hover:text-gray-600">Termos de Serviço</a> e <a href="#privacy" class="underline hover:text-gray-600">Política de Privacidade</a>.
      </p>

    </div>
  </div>
</div>
