<script>
  import { onMount } from 'svelte';
  import { account } from '../lib/stores/account';

  export let onGoHome = () => {};

  // Mode: 'login' | 'register'
  let authMode = 'login';

  let authMethod = 'email';
  let phoneToConfirm = '';
  let emailToConfirm = '';
  let emailCodeResent = false;
  let code = '';

  // Form fields
  let email = '';
  let password = '';
  let name = '';
  let phone = '';
  let rememberMe = true;
  let showPassword = false;
  $: verifyingEmail = authMethod === 'email' && authMode === 'register' && Boolean(emailToConfirm);
  $: awaitingCode = authMethod === 'phone' ? Boolean(phoneToConfirm) : verifyingEmail;

  // Scroll down slightly when the login page loads
  onMount(() => {
    window.scrollTo({
      top: 100,
      behavior: 'smooth'
    });
  });

  function toggleMode(mode) {
    if (authMode === mode) return;
    authMode = mode;
    resetStep();
  }

  function toggleMethod(method) {
    if (authMethod === method) return;
    authMethod = method;
    resetStep();
  }

  function resetStep() {
    phoneToConfirm = '';
    emailToConfirm = '';
    emailCodeResent = false;
    code = '';
    password = '';
    showPassword = false;
    account.clearError();
  }

  function formatPhone(digits) {
    if (!digits) return '';
    if (digits.length <= 2) return `(${digits}`;

    const local = digits.slice(2);
    const prefixLength = local.length >= 9 ? 5 : 4;
    const suffix = local.slice(prefixLength);
    return `(${digits.slice(0, 2)}) ${local.slice(0, prefixLength)}${suffix ? `-${suffix}` : ''}`;
  }

  function handlePhoneInput(event) {
    const input = event.currentTarget;
    const digits = input.value.replace(/\D/g, '');
    // Accept full Brazilian numbers from paste or autofill, too.
    const hasCountryCode = event.inputType !== 'insertText' && /^55\d{10,11}$/.test(digits);
    const local = (hasCountryCode ? digits.slice(2) : digits).slice(0, 11);
    phone = formatPhone(local);
    if (input.value !== phone) input.value = phone;
  }

  function backToEmailDetails() {
    emailToConfirm = '';
    emailCodeResent = false;
    code = '';
    account.clearError();
  }

  async function resendEmailCode() {
    emailCodeResent = false;
    if (await account.requestEmailCode(emailToConfirm)) {
      code = '';
      emailCodeResent = true;
    }
  }

  async function handleSubmit(e) {
    e.preventDefault();
    if ($account.loading) return;

    const isRegistering = authMode === 'register';
    if (authMethod === 'phone') {
      if (!phoneToConfirm) {
        const digits = phone.replace(/\D/g, '');
        if (digits.length !== 10 && digits.length !== 11) {
          account.setError('Informe um telefone com DDD e 8 ou 9 dígitos.');
        } else if (await account.requestPhoneCode(authMode, `+55${digits}`)) {
          phoneToConfirm = digits;
        }
        return;
      }
      const ok = await account.verifyPhone(authMode, { phone: `+55${phoneToConfirm}`, name, code, rememberMe: isRegistering || rememberMe });
      if (ok) onGoHome();
      return;
    }

    if (isRegistering && !emailToConfirm) {
      const address = email.trim();
      if (await account.requestEmailCode(address)) { emailToConfirm = address; code = ''; }
      return;
    }
    const ok = isRegistering
      ? await account.register({ email: emailToConfirm, password, name, code })
      : await account.login({ email, password, rememberMe });
    if (ok) onGoHome();
  }
</script>

<div class="min-h-screen w-full flex bg-background text-text font-sans selection:bg-primary selection:text-white">

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
    {#if !verifyingEmail}
    <button
      onclick={onGoHome}
      type="button"
      class="absolute top-8 left-8 lg:left-12 flex items-center gap-2 text-text-muted hover:text-primary font-medium transition-colors cursor-pointer"
    >
      <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m15 18-6-6 6-6"/></svg>
      Voltar
    </button>
    {/if}

    <div class="w-full max-w-md space-y-6">
      {#if !verifyingEmail}
      <!-- Header -->
      <div>
        <span class="text-primary font-bold text-xs tracking-wider uppercase mb-2 block">Bem-vindo ao Pratello</span>
        <h2 class="text-3xl font-bold text-text mb-2">
          {authMode === 'login' ? 'Entre na sua conta.' : 'Crie sua conta no Pratello.'}
        </h2>
        <p class="text-text-muted text-sm">
          {authMode === 'login'
            ? 'Faça login para acessar seus menus favoritos, finalizar pedidos e gerenciar estabelecimentos.'
            : 'Cadastre sua conta para fazer pedidos, acumular pontos ou gerenciar restaurantes.'}
        </p>
      </div>

      <!-- Mode Toggle (Entrar / Cadastrar) -->
      <div class="flex border-b border-border">
        <button
          type="button"
          onclick={() => toggleMode('login')}
          disabled={$account.loading}
          aria-pressed={authMode === 'login'}
          class="flex-1 pb-3 text-sm font-bold text-center border-b-2 transition-all cursor-pointer {authMode === 'login' ? 'border-primary text-primary' : 'border-transparent text-text-muted hover:text-text'}"
        >
          Entrar
        </button>
        <button
          type="button"
          onclick={() => toggleMode('register')}
          disabled={$account.loading}
          aria-pressed={authMode === 'register'}
          class="flex-1 pb-3 text-sm font-bold text-center border-b-2 transition-all cursor-pointer {authMode === 'register' ? 'border-primary text-primary' : 'border-transparent text-text-muted hover:text-text'}"
        >
          Cadastrar
        </button>
      </div>

      <div class="flex bg-surface-hover p-1 rounded-lg border border-border" role="group" aria-label="Método de autenticação">
          <button
            type="button"
            onclick={() => toggleMethod('email')}
            disabled={$account.loading}
            aria-pressed={authMethod === 'email'}
            class="flex-1 py-2 text-sm font-bold rounded-md transition-all cursor-pointer {authMethod === 'email' ? 'bg-surface text-text shadow-sm' : 'text-text-muted hover:text-text'}"
          >
            E-mail
          </button>
          <button
            type="button"
            onclick={() => toggleMethod('phone')}
            disabled={$account.loading}
            aria-pressed={authMethod === 'phone'}
            class="flex-1 py-2 text-sm font-bold rounded-md transition-all cursor-pointer {authMethod === 'phone' ? 'bg-surface text-text shadow-sm' : 'text-text-muted hover:text-text'}"
          >
            Telefone
          </button>
      </div>

      {/if}
      <!-- ERROR / WARNING BANNER -->
      {#if $account.error}
        <div
          role="alert"
          aria-live="assertive"
          class="flex items-start gap-3 p-4 rounded-lg bg-error/10 border border-error/30 text-error text-sm animate-fade-in shadow-sm"
        >
          <svg class="w-5 h-5 text-error flex-shrink-0 mt-0.5" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
          </svg>
          <div class="flex-1">
            <p class="font-semibold text-error">Atenção</p>
            <p class="mt-0.5">{$account.error}</p>
          </div>
          <button
            type="button"
            onclick={() => account.clearError()}
            class="text-error/70 hover:text-error cursor-pointer p-0.5 rounded transition-colors"
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

        {#if authMode === 'register' && !awaitingCode}
          <div>
            <label for="name" class="block text-sm font-medium text-text mb-1">Nome completo (opcional)</label>
            <input
              type="text"
              id="name"
              bind:value={name}
              autocomplete="name"
              placeholder="Ex: Carlos Silva"
              class="w-full px-4 py-3 rounded-lg bg-surface text-text placeholder:text-text-muted/70 border border-border focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none transition-all"
            />
          </div>
        {/if}

        {#if authMethod === 'email' && !verifyingEmail}
        <div>
          <label for="email" class="block text-sm font-medium text-text mb-1">Endereço de Email</label>
          <input
            type="email"
            id="email"
            bind:value={email}
            readonly={$account.loading}
            autocomplete="email"
            placeholder="voce@exemplo.com"
            class="w-full px-4 py-3 rounded-lg bg-surface text-text placeholder:text-text-muted/70 border border-border focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none transition-all"
            required
          />
        </div>

        <div>
          <label for="password" class="block text-sm font-medium text-text mb-1">Senha</label>
          <div class="relative">
            <input
              type={showPassword ? 'text' : 'password'}
              id="password"
              bind:value={password}
              autocomplete={authMode === 'login' ? 'current-password' : 'new-password'}
              minlength={authMode === 'register' ? 8 : undefined}
              placeholder="Sua senha secreta"
              class="w-full px-4 py-3 rounded-lg bg-surface text-text placeholder:text-text-muted/70 border border-border focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none transition-all pr-12"
              required
            />
            <button
              type="button"
              onclick={() => showPassword = !showPassword}
              class="absolute right-3 top-1/2 -translate-y-1/2 text-text-muted hover:text-text cursor-pointer p-1"
              title={showPassword ? 'Ocultar senha' : 'Exibir senha'}
            >
              {#if showPassword}
                <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9.88 9.88a3 3 0 1 0 4.24 4.24"/><path d="M10.73 5.08A10.43 10.43 0 0 1 12 5c7 0 10 7 10 7a13.16 13.16 0 0 1-1.67 2.68"/><path d="M6.61 6.61A13.526 13.526 0 0 0 2 12s3 7 10 7a9.74 9.74 0 0 0 5.39-1.61"/><line x1="2" x2="22" y1="2" y2="22"/></svg>
              {:else}
                <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M2 12s3-7 10-7 10 7 10 7-3 7-10 7-10-7-10-7Z"/><circle cx="12" cy="12" r="3"/></svg>
              {/if}
            </button>
          </div>
          {#if authMode === 'register'}
            <p class="mt-1 text-xs text-text-muted">Use pelo menos 8 caracteres.</p>
          {/if}
        </div>
        {:else if authMethod === 'phone' && phoneToConfirm}
          <div class="space-y-2 rounded-lg bg-surface-hover border border-border p-4 text-sm text-text">
            <p role="status">Confirme o telefone <strong>{formatPhone(phoneToConfirm)}</strong> com o código de seis dígitos.</p>
            <button type="button" onclick={resetStep} disabled={$account.loading} class="font-bold text-primary hover:text-primary-hover cursor-pointer">
              Alterar telefone
            </button>
          </div>
        {:else if authMethod === 'phone'}
          <div>
            <label for="phone" class="block text-sm font-medium text-text mb-1">Telefone</label>
            <input
              type="tel"
              id="phone"
              value={phone}
              oninput={handlePhoneInput}
              autocomplete="tel-national"
              placeholder="(81) 90000-0000"
              readonly={$account.loading}
              class="w-full px-4 py-3 rounded-lg bg-surface text-text placeholder:text-text-muted/70 border border-border focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none transition-all"
              required
            />
            <p class="mt-1 text-xs text-text-muted">Inclua o DDD.{authMode === 'login' ? ' Use um telefone já verificado na sua conta.' : ''}</p>
          </div>
        {/if}

        {#if awaitingCode}
          <label for="verification-code" class="block text-sm font-medium text-text">Código de verificação</label>
          <input id="verification-code" type="text" inputmode="numeric" autocomplete="one-time-code" pattern={'[0-9]{6}'} maxlength="6" bind:value={code} required class="w-full px-4 py-3 rounded-lg bg-surface text-text border border-border focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none transition-all" />
          {#if verifyingEmail}
            <div class="flex items-center justify-between gap-4">
              <button type="button" onclick={backToEmailDetails} disabled={$account.loading} class="text-sm font-bold text-text-muted hover:text-text">Voltar</button>
              <button type="button" onclick={resendEmailCode} disabled={$account.loading} class="text-sm font-bold text-primary hover:text-primary-hover">Enviar código novamente</button>
            </div>
            {#if emailCodeResent}
              <p role="status" class="text-sm text-text-muted">Novo código solicitado.</p>
            {/if}
          {:else}
            <button type="button" onclick={() => { phoneToConfirm = ''; code = ''; }} disabled={$account.loading} class="text-sm font-bold text-primary hover:text-primary-hover">Solicitar novo código</button>
          {/if}
        {/if}

        {#if authMode === 'login'}
          <div class="flex items-center justify-between">
            <div class="flex items-center">
              <input
                id="remember-me"
                type="checkbox"
                bind:checked={rememberMe}
                class="h-4 w-4 text-primary focus:ring-primary border-border rounded cursor-pointer accent-primary"
              />
              <label for="remember-me" class="ml-2 block text-sm text-text cursor-pointer">
                Lembrar de mim
              </label>
            </div>

            {#if authMethod === 'email'}
            <div class="text-sm">
              <button
                type="button"
                onclick={() => account.setError('Para redefinir sua senha, solicite suporte ao administrador do sistema.')}
                class="font-bold text-primary hover:text-primary-hover cursor-pointer"
              >
                Esqueceu a senha?
              </button>
            </div>
            {/if}
          </div>
        {/if}

        <button
          type="submit"
          disabled={$account.loading}
          class="w-full bg-primary hover:bg-primary-hover disabled:opacity-50 disabled:cursor-not-allowed text-white font-bold py-3.5 rounded-lg transition-colors shadow-lg shadow-primary/20 cursor-pointer flex items-center justify-center gap-2"
        >
          {#if $account.loading}
            <svg class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            <span>Aguarde...</span>
          {:else}
            <span>{authMethod === 'phone' ? (phoneToConfirm ? 'Confirmar telefone' : 'Continuar') : (authMode === 'login' ? 'Entrar' : (verifyingEmail ? 'Confirmar código' : 'Continuar'))}</span>
          {/if}
        </button>
      </form>

      <!-- Toggle mode link -->
      {#if !verifyingEmail}
      {#if authMode === 'login'}
        <p class="text-center text-sm text-text-muted">
          Novo aqui?
          <button
            type="button"
            onclick={() => toggleMode('register')}
            disabled={$account.loading}
            class="font-bold text-primary hover:text-primary-hover cursor-pointer ml-1"
          >
            Crie uma conta
          </button>
        </p>
      {:else}
        <p class="text-center text-sm text-text-muted">
          Já possui uma conta?
          <button
            type="button"
            onclick={() => toggleMode('login')}
            disabled={$account.loading}
            class="font-bold text-primary hover:text-primary-hover cursor-pointer ml-1"
          >
            Faça login
          </button>
        </p>
      {/if}

      <p class="text-center text-xs text-text-muted/70">
        Ao continuar, você concorda com nossos <span class="underline hover:text-text cursor-pointer">Termos de Serviço</span> e <span class="underline hover:text-text cursor-pointer">Política de Privacidade</span>.
      </p>
      {/if}

    </div>
  </div>
</div>
