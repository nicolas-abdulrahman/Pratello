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
  role?: string;       // 'cliente' | 'funcionario' | 'gerente' | 'proprietario' | 'sysadmin'
  initials?: string;
}

export interface AccountState extends AccountData {
  isLoggedIn: boolean;
  token: string | null;
  error: string | null;
  loading: boolean;
}

export interface LoginParams {
  email: string;
  password: string;
  name?: string;
  role?: string;
}

export interface RegisterParams {
  email: string;
  password: string;
  phone?: string;
  name?: string;
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
  id: undefined,
  name: '',
  email: '',
  phone: '',
  role: '',
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
  if (typeof localStorage === 'undefined') return emptyState;
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return emptyState;
    const parsed = JSON.parse(raw);
    if (typeof parsed?.isLoggedIn === 'boolean') {
      return {
        ...emptyState,
        ...parsed,
        error: null,
        loading: false
      };
    }
    return emptyState;
  } catch {
    return emptyState;
  }
}

function saveToStorage(state: AccountState): void {
  if (typeof localStorage === 'undefined') return;
  // Never persist passwords or transient error/loading states
  const toSave = {
    isLoggedIn: state.isLoggedIn,
    token: state.token,
    id: state.id,
    name: state.name,
    email: state.email,
    phone: state.phone,
    role: state.role,
    initials: state.initials,
    created_at: state.created_at
  };
  localStorage.setItem(STORAGE_KEY, JSON.stringify(toSave));
}

function clearStorage(): void {
  if (typeof localStorage === 'undefined') return;
  localStorage.removeItem(STORAGE_KEY);
}

// ─────────────────────────────────────────────────────────────
// Account Store Definition
// ─────────────────────────────────────────────────────────────

function createAccountStore() {
  const store = writable<AccountState>(loadFromStorage());
  const { subscribe, set, update } = store;

  // Helper to get current bearer token
  function getToken(): string | null {
    return get(store).token;
  }

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
    if (res.status === 401) {
      return 'E-mail ou senha incorretos.';
    }
    if (res.status === 409) {
      return 'Este endereço de e-mail já está cadastrado no sistema.';
    }
    if (res.status === 400) {
      return 'Dados incompletos ou em formato inválido.';
    }
    if (res.status === 500) {
      return 'Erro interno no servidor do backend (HTTP 500).';
    }
    return `${fallback} (Status HTTP ${res.status})`;
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
    async login({ email, password, name, role }: LoginParams): Promise<boolean> {
      update((s) => ({ ...s, loading: true, error: null }));

      try {
        const res = await fetch(`${API_BASE}/auth/login`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Accept': 'application/json'
          },
          body: JSON.stringify({ email: email.trim(), password })
        });

        if (!res.ok) {
          const errMsg = await parseError(res, 'Falha ao autenticar.');
          update((s) => ({ ...s, loading: false, error: errMsg }));
          return false;
        }

        const data = await res.json();
        const accountInfo = data.account || {};
        const resolvedName = name || accountInfo.name || deriveNameFromEmail(accountInfo.email || email);
        const resolvedInitials = makeInitials(resolvedName, accountInfo.email || email);

        const nextState: AccountState = {
          isLoggedIn: true,
          token: data.token || null,
          id: accountInfo.id,
          email: accountInfo.email || email,
          phone: accountInfo.phone || '',
          name: resolvedName,
          role: role || (accountInfo.is_sysadmin ? 'sysadmin' : 'cliente'),
          initials: resolvedInitials,
          created_at: accountInfo.created_at,
          error: null,
          loading: false
        };

        saveToStorage(nextState);
        set(nextState);
        return true;
      } catch (err: any) {
        const isNetworkErr = err instanceof TypeError || String(err).includes('fetch') || String(err).includes('NetworkError');
        const message = isNetworkErr
          ? 'Não foi possível conectar ao servidor backend (http://localhost:8080). Verifique se o servidor está ativo.'
          : (err?.message || 'Erro inesperado na autenticação.');
        update((s) => ({ ...s, loading: false, error: message }));
        return false;
      }
    },

    /**
     * Register new global account via /api/auth/register and automatically authenticate
     */
    async register({ email, password, phone = '', name = '' }: RegisterParams): Promise<boolean> {
      update((s) => ({ ...s, loading: true, error: null }));

      try {
        const res = await fetch(`${API_BASE}/auth/register`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Accept': 'application/json'
          },
          body: JSON.stringify({ name: name.trim(), email: email.trim(), password, phone: phone.trim() })
        });

        if (!res.ok) {
          const errMsg = await parseError(res, 'Falha ao criar conta.');
          update((s) => ({ ...s, loading: false, error: errMsg }));
          return false;
        }

        // Successfully registered -> automatically log in
        return await this.login({ email, password, name });
      } catch (err: any) {
        const isNetworkErr = err instanceof TypeError || String(err).includes('fetch') || String(err).includes('NetworkError');
        const message = isNetworkErr
          ? 'Não foi possível conectar ao servidor backend (http://localhost:8080). Verifique se o servidor está ativo.'
          : (err?.message || 'Erro inesperado ao realizar cadastro.');
        update((s) => ({ ...s, loading: false, error: message }));
        return false;
      }
    },

    /**
     * Fetch authenticated profile via /api/auth/me
     */
    async fetchMe(): Promise<AccountData | null> {
      const token = getToken();
      if (!token) return null;

      try {
        const res = await fetch(`${API_BASE}/auth/me`, {
          headers: {
            'Authorization': `Bearer ${token}`,
            'Accept': 'application/json'
          }
        });

        if (!res.ok) {
          if (res.status === 401) {
            this.logout();
          }
          return null;
        }

        const data = await res.json();
        update((s) => {
          const resolvedName = s.name || deriveNameFromEmail(data.email);
          const updated: AccountState = {
            ...s,
            id: data.id,
            email: data.email,
            phone: data.phone,
            created_at: data.created_at,
            initials: makeInitials(resolvedName, data.email)
          };
          saveToStorage(updated);
          return updated;
        });

        return data;
      } catch {
        return null;
      }
    },

    /**
     * Terminate user session and clear stored tokens
     */
    logout(): void {
      clearStorage();
      set(emptyState);
    },

    /**
     * Create logical restaurant via POST /api/restaurants
     */
    async createRestaurant({ name, slug, description = '' }: CreateRestaurantParams): Promise<RestaurantResponse | null> {
      const token = getToken();
      update((s) => ({ ...s, loading: true, error: null }));

      try {
        const headers: Record<string, string> = {
          'Content-Type': 'application/json',
          'Accept': 'application/json'
        };
        if (token) {
          headers['Authorization'] = `Bearer ${token}`;
        }

        const res = await fetch(`${API_BASE}/restaurants`, {
          method: 'POST',
          headers,
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
      const token = getToken();
      update((s) => ({ ...s, loading: true, error: null }));

      try {
        const headers: Record<string, string> = {
          'Content-Type': 'application/json',
          'Accept': 'application/json'
        };
        if (token) {
          headers['Authorization'] = `Bearer ${token}`;
        }

        const res = await fetch(`${API_BASE}/restaurants/${encodeURIComponent(slug)}/managers`, {
          method: 'POST',
          headers,
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
      const token = getToken();
      update((s) => ({ ...s, loading: true, error: null }));

      try {
        const headers: Record<string, string> = {
          'Content-Type': 'application/json',
          'Accept': 'application/json'
        };
        if (token) {
          headers['Authorization'] = `Bearer ${token}`;
        }

        const res = await fetch(`${API_BASE}/restaurants/${encodeURIComponent(slug)}/owners`, {
          method: 'POST',
          headers,
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
