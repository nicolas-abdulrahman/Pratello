import { writable, derived, get } from 'svelte/store';

// ─────────────────────────────────────────────────────────────
// Types & Interfaces
// ─────────────────────────────────────────────────────────────

export interface AccountData {
  id?: number;
  email: string;
  name?: string;
  phone?: string;
  created_at?: string;
  initials?: string;
}

export interface AccountState extends AccountData {
  isLoggedIn: boolean;
  token: string | null;
  refreshToken: string | null;
  expiresAt: number;
  error: string | null;
  loading: boolean;
}

export interface LoginParams {
  email: string;
  password: string;
  rememberMe?: boolean;
}

export interface RegisterParams {
  email: string;
  password: string;
  name?: string;
  cpf?: string;
}

type AuthMode = 'login' | 'register';

interface TokenPair {
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

interface AccountProfile {
  id: number;
  email: string | null;
  phone: string | null;
  name: string | null;
  created_at?: string;
}

export interface CreateRestaurantParams {
  name: string;
  slug: string;
  description?: string;
}

export interface AssignManagerParams {
  slug: string;
  accountId: number;
}

export interface AssignOwnerParams {
  slug: string;
  accountId: number;
}

export interface RestaurantResponse {
  id?: number;
  name: string;
  slug: string;
  description?: string;
  created_at?: string;
}

export interface MembershipResponse {
  id: number;
  restaurant_slug: string;
  account_id: number;
}

// ─────────────────────────────────────────────────────────────
// Constants & Storage Helpers
// ─────────────────────────────────────────────────────────────

const STORAGE_KEY = 'pratello:account';
const API_BASE = '/api';

const emptyState: AccountState = {
  isLoggedIn: false,
  token: null,
  refreshToken: null,
  expiresAt: 0,
  id: undefined,
  name: '',
  email: '',
  phone: '',
  initials: '',
  created_at: undefined,
  error: null,
  loading: false
};

function makeInitials(name = '', email = ''): string {
  if (name && name.trim()) {
    const parts = name.trim().split(/\s+/).filter(Boolean);
    if (parts.length >= 2) {
      return (parts[0][0] + parts[1][0]).toUpperCase();
    }
    return parts[0].slice(0, 2).toUpperCase();
  }
  if (email) {
    const user = email.split('@')[0] || '';
    return user.slice(0, 2).toUpperCase();
  }
  return 'U';
}

function deriveNameFromEmail(email: string): string {
  if (!email) return 'Usuário';
  const prefix = email.split('@')[0] || 'Usuário';
  return prefix.charAt(0).toUpperCase() + prefix.slice(1);
}

function loadFromStorage(): AccountState {
  try {
    const raw = sessionStorage.getItem(STORAGE_KEY) ?? localStorage.getItem(STORAGE_KEY);
    if (!raw) return emptyState;
    const parsed = JSON.parse(raw);
    if (parsed?.isLoggedIn === true && typeof parsed.token === 'string' && parsed.token &&
        typeof parsed.refreshToken === 'string' && parsed.refreshToken && Number.isFinite(parsed.expiresAt)) {
      return {
        ...emptyState,
        isLoggedIn: true,
        token: parsed.token,
        refreshToken: parsed.refreshToken,
        expiresAt: parsed.expiresAt,
        id: parsed.id,
        name: parsed.name,
        email: parsed.email ?? '',
        phone: parsed.phone,
        initials: parsed.initials,
        created_at: parsed.created_at
      };
    }
    return emptyState;
  } catch {
    return emptyState;
  }
}

function saveToStorage(state: AccountState, rememberMe: boolean): void {
  // Never persist passwords or transient error/loading states
  const toSave = {
    isLoggedIn: state.isLoggedIn,
    token: state.token,
    refreshToken: state.refreshToken,
    expiresAt: state.expiresAt,
    id: state.id,
    name: state.name,
    email: state.email,
    phone: state.phone,
    initials: state.initials,
    created_at: state.created_at
  };
  try {
    (rememberMe ? localStorage : sessionStorage).setItem(STORAGE_KEY, JSON.stringify(toSave));
  } catch {
    // Storage may be unavailable; the session still works in memory.
  }
}

function clearStorage(): void {
  try { sessionStorage.removeItem(STORAGE_KEY); } catch { /* Storage unavailable */ }
  try { localStorage.removeItem(STORAGE_KEY); } catch { /* Storage unavailable */ }
}

function isRemembered(): boolean {
  try { return sessionStorage.getItem(STORAGE_KEY) === null; } catch { return true; }
}

function postJSON(path: string, body: object): Promise<Response> {
  return fetch(`${API_BASE}${path}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
    body: JSON.stringify(body)
  });
}

async function readTokens(res: Response) {
  const pair: TokenPair = await res.json();
  if (!pair?.access_token || !pair.refresh_token || !Number.isFinite(pair.expires_in) || pair.expires_in <= 0) {
    throw new Error('O servidor não retornou uma sessão válida.');
  }
  return {
    token: pair.access_token,
    refreshToken: pair.refresh_token,
    expiresAt: Date.now() + pair.expires_in * 1000
  };
}

function profileFields(profile: AccountProfile): AccountData {
  const email = profile.email ?? '';
  const name = profile.name || deriveNameFromEmail(email);

  return {
    id: profile.id,
    email,
    name,
    phone: profile.phone ?? '',
    created_at: profile.created_at,
    initials: makeInitials(name, email)
  };
}

// ─────────────────────────────────────────────────────────────
// Account Store Definition
// ─────────────────────────────────────────────────────────────

function createAccountStore() {
  const store = writable<AccountState>(loadFromStorage());
  const { subscribe, set, update } = store;
  let rememberMe = isRemembered();
  let sessionVersion = 0;
  let pendingRefresh: Promise<string | null> | null = null;

  // Helper to handle API response errors cleanly and descriptively
  async function parseError(res: Response, fallback: string): Promise<string> {
    try {
      const data = await res.json();
      if (data && typeof data.error === 'string' && data.error.trim()) {
        return data.error;
      }
      if (data && typeof data.message === 'string' && data.message.trim()) {
        return data.message;
      }
      if (data && typeof data.detail === 'string' && data.detail.trim()) {
        return data.detail;
      }
    } catch {
      // Body not JSON
    }
    if (res.status === 400) {
      return 'Dados incompletos ou em formato inválido.';
    }
    if (res.status === 500) {
      return 'Erro interno no servidor do backend (HTTP 500).';
    }
    return `${fallback} (Status HTTP ${res.status})`;
  }

  function setSession(state: AccountState) {
    saveToStorage(state, rememberMe);
    set(state);
  }

  function logout() {
    sessionVersion++;
    pendingRefresh = null;
    clearStorage();
    set(emptyState);
  }

  async function runAuth(action: (version: number) => Promise<void>, fallback: string): Promise<boolean> {
    if (get(store).loading) return false;
    const version = sessionVersion;
    update((s) => ({ ...s, loading: true, error: null }));
    try {
      await action(version);
      return version === sessionVersion;
    } catch (err) {
      if (version === sessionVersion) {
        const error = err instanceof TypeError
          ? 'Não foi possível conectar ao servidor. Tente novamente.'
          : err instanceof Error ? err.message : fallback;
        update((s) => ({ ...s, error }));
      }
      return false;
    } finally {
      if (version === sessionVersion && get(store).loading) {
        update((s) => ({ ...s, loading: false }));
      }
    }
  }

  function authenticate(path: string, body: object, fallback: string, persist: boolean) {
    return runAuth(async (version) => {
      const res = await postJSON(path, body);
      if (!res.ok) throw new Error(await parseError(res, fallback));
      const tokens = await readTokens(res);
      const me = await fetch(`${API_BASE}/auth/me`, {
        headers: { Authorization: `Bearer ${tokens.token}`, Accept: 'application/json' }
      });
      if (!me.ok) throw new Error(await parseError(me, 'Não foi possível carregar seu perfil.'));
      const profile: AccountProfile = await me.json();
      if (version !== sessionVersion) return;
      pendingRefresh = null;
      clearStorage();
      rememberMe = persist;
      setSession({ ...emptyState, ...profileFields(profile), ...tokens, isLoggedIn: true });
    }, fallback);
  }

  function refreshSession(): Promise<string | null> {
    if (pendingRefresh) return pendingRefresh;
    const { refreshToken } = get(store);
    if (!refreshToken) return Promise.resolve(null);
    const version = sessionVersion;
    const isCurrent = () => version === sessionVersion && get(store).refreshToken === refreshToken;
    const request = (async () => {
      const res = await postJSON('/auth/refresh', { refresh_token: refreshToken });
      if (!isCurrent()) return null;
      if (res.status === 401) {
        logout();
        return null;
      }
      if (!res.ok) throw new Error(await parseError(res, 'Não foi possível renovar a sessão.'));
      const tokens = await readTokens(res);
      if (!isCurrent()) return null;
      setSession({ ...get(store), ...tokens });
      return tokens.token;
    })().finally(() => {
      if (pendingRefresh === request) pendingRefresh = null;
    });
    pendingRefresh = request;
    return request;
  }

  // Use API-relative paths (e.g. '/auth/me') for shared authenticated requests.
  async function authenticatedFetch(path: string, init: RequestInit = {}): Promise<Response> {
    const { token, expiresAt } = get(store);
    const headers = new Headers(init.headers);
    headers.set('Accept', 'application/json');

    const send = (accessToken: string | null) => {
      if (accessToken) headers.set('Authorization', `Bearer ${accessToken}`);
      else headers.delete('Authorization');
      return fetch(`${API_BASE}${path}`, { ...init, headers });
    };

    if (!token) return send(null);
    if (expiresAt <= Date.now()) {
      const renewed = await refreshSession();
      return renewed ? send(renewed) : new Response(null, { status: 401 });
    }

    const version = sessionVersion;
    const res = await send(token);
    if (res.status !== 401 || version !== sessionVersion) return res;
    const currentToken = get(store).token;
    const renewed = currentToken !== token ? currentToken : await refreshSession();
    return renewed ? send(renewed) : res;
  }

  return {
    subscribe,

    /**
     * Clear active warning/error message
     */
    clearError() {
      update((s) => ({ ...s, error: null }));
    },

    /**
     * Set a custom error or warning message in the store
     */
    setError(message: string) {
      update((s) => ({ ...s, error: message }));
    },

    /**
     * Login via /api/auth/login
     */
    login({ email, password, rememberMe = true }: LoginParams): Promise<boolean> {
      return authenticate('/auth/login', { email: email.trim(), password }, 'Falha ao entrar com e-mail.', rememberMe);
    },

    /**
     * Register new global account via /api/auth/register and automatically authenticate
     */
    register({ email, password, name, cpf, code }: RegisterParams & { code: string }): Promise<boolean> {
      return authenticate('/auth/register', {
        email: email.trim(), password, name: name?.trim() || undefined, cpf: cpf?.trim() || undefined, code
      }, 'Falha ao criar conta.', true);
    },

    requestEmailCode(email: string): Promise<boolean> {
      const fallback = 'Não foi possível iniciar a confirmação do e-mail.';
      return runAuth(async () => {
        const res = await postJSON('/auth/register/email/request-code', { email: email.trim() });
        if (!res.ok) throw new Error(await parseError(res, fallback));
      }, fallback);
    },

    requestPhoneCode(mode: AuthMode, phone: string): Promise<boolean> {
      const fallback = 'Não foi possível iniciar a confirmação do telefone.';
      return runAuth(async () => {
        const res = await postJSON(`/auth/${mode}/phone/request-code`, { phone: phone.trim() });
        if (!res.ok) throw new Error(await parseError(res, fallback));
      }, fallback);
    },

    verifyPhone(mode: AuthMode, { phone, name, cpf, code, rememberMe = true }: {
      phone: string; name?: string; cpf?: string; code: string; rememberMe?: boolean;
    }): Promise<boolean> {
      const body = mode === 'register'
        ? { phone: phone.trim(), name: name?.trim() || undefined, cpf: cpf?.trim() || undefined, code }
        : { phone: phone.trim(), code };
      return authenticate(`/auth/${mode}/phone/verify`, body, 'Falha ao confirmar o telefone.', rememberMe);
    },

    /**
     * Fetch authenticated profile via /api/auth/me
     */
    async fetchMe(): Promise<AccountData | null> {
      if (!get(store).token) return null;
      const version = sessionVersion;

      try {
        const res = await authenticatedFetch('/auth/me');
        if (!res.ok) return null;
        const profile: AccountProfile = await res.json();
        if (version !== sessionVersion) return null;
        const data = profileFields(profile);
        setSession({ ...get(store), ...data });
        return data;
      } catch {
        return null;
      }
    },

    /**
     * Terminate user session and clear stored tokens
     */
    logout,
    authenticatedFetch,

    /**
     * Create logical restaurant via POST /api/restaurants
     */
    async createRestaurant({ name, slug, description = '' }: CreateRestaurantParams): Promise<RestaurantResponse | null> {
      update((s) => ({ ...s, loading: true, error: null }));

      try {
        const res = await authenticatedFetch('/restaurants', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ name, slug, description })
        });

        if (!res.ok) {
          const errMsg = await parseError(res, 'Erro ao criar restaurante.');
          update((s) => ({ ...s, loading: false, error: errMsg }));
          return null;
        }

        const created = await res.json();
        update((s) => ({ ...s, loading: false, error: null }));
        return created;
      } catch (err: any) {
        const message = err?.message || 'Erro de conexão ao criar restaurante.';
        update((s) => ({ ...s, loading: false, error: message }));
        return null;
      }
    },

    /**
     * Assign a manager to a restaurant via POST /api/restaurants/{slug}/managers
     */
    async assignManager(slug: string, accountId: number): Promise<MembershipResponse | null> {
      update((s) => ({ ...s, loading: true, error: null }));

      try {
        const res = await authenticatedFetch(`/restaurants/${encodeURIComponent(slug)}/managers`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ account_id: accountId })
        });

        if (!res.ok) {
          const errMsg = await parseError(res, 'Erro ao vincular gerente ao restaurante.');
          update((s) => ({ ...s, loading: false, error: errMsg }));
          return null;
        }

        const result = await res.json();
        update((s) => ({ ...s, loading: false, error: null }));
        return result;
      } catch (err: any) {
        const message = err?.message || 'Erro de conexão ao vincular gerente.';
        update((s) => ({ ...s, loading: false, error: message }));
        return null;
      }
    },

    /**
     * Assign / promote an owner to a restaurant via POST /api/restaurants/{slug}/owners
     */
    async assignOwner(slug: string, accountId: number): Promise<MembershipResponse | null> {
      update((s) => ({ ...s, loading: true, error: null }));

      try {
        const res = await authenticatedFetch(`/restaurants/${encodeURIComponent(slug)}/owners`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ account_id: accountId })
        });

        if (!res.ok) {
          const errMsg = await parseError(res, 'Erro ao vincular proprietário ao restaurante.');
          update((s) => ({ ...s, loading: false, error: errMsg }));
          return null;
        }

        const result = await res.json();
        update((s) => ({ ...s, loading: false, error: null }));
        return result;
      } catch (err: any) {
        const message = err?.message || 'Erro de conexão ao vincular proprietário.';
        update((s) => ({ ...s, loading: false, error: message }));
        return null;
      }
    }
  };
}

export const account = createAccountStore();
export const isLoggedIn = derived(account, ($a) => $a.isLoggedIn);
export const authToken = derived(account, ($a) => $a.token);
export const accountError = derived(account, ($a) => $a.error);
export const isAccountLoading = derived(account, ($a) => $a.loading);
