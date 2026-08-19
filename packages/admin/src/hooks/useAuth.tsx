import { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { createPortal } from 'react-dom';

export type Role = 'associate' | 'seller' | 'guide' | 'manager';

export interface AuthUser {
  uid: string;
  email: string;
  roles: Role[];
  name: string;
}

export interface AuthState {
  user: AuthUser | null;
  token: string | null;
  loading: boolean;
}

const STORAGE_KEY = 'divoene_auth';

export function getStoredAuth(): AuthState {
  const stored = localStorage.getItem(STORAGE_KEY);
  if (stored) {
    try {
      const parsed = JSON.parse(stored);
      if (parsed.token && parsed.user) {
        // Backward compatibility: legacy sessions stored a single role.
        if (parsed.user.roles && Array.isArray(parsed.user.roles)) {
          return { user: parsed.user, token: parsed.token, loading: false };
        }
        if (parsed.user.role) {
          parsed.user.roles = [parsed.user.role];
          localStorage.setItem(STORAGE_KEY, JSON.stringify(parsed));
          return { user: parsed.user, token: parsed.token, loading: false };
        }
      }
    } catch { /* ignore */ }
  }
  return { user: null, token: null, loading: false };
}

export async function login(email: string, password: string): Promise<AuthState> {
  const res = await fetch('/api/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || 'Falha na autenticação');
  }
  const data = await res.json();
  // Decode JWT payload to get user info
  const payload = JSON.parse(atob(data.token.split('.')[1]));
  const roles: Role[] = Array.isArray(payload.roles) ? payload.roles : payload.role ? [payload.role] : [];
  const user: AuthUser = {
    uid: payload.uid,
    email: email,
    roles,
    name: payload.name,
  };
  const auth = { user, token: data.token, loading: false };
  localStorage.setItem(STORAGE_KEY, JSON.stringify(auth));
  return auth;
}

export function logout() {
  localStorage.removeItem(STORAGE_KEY);
  window.location.href = '/login';
}

export function getAuthToken(): string | null {
  const stored = localStorage.getItem(STORAGE_KEY);
  if (!stored) return null;
  try {
    return JSON.parse(stored).token;
  } catch { return null; }
}

export function fetchWithAuth(input: RequestInfo, init: RequestInit = {}): Promise<Response> {
  const token = getAuthToken();
  const headers = new Headers(init.headers);
  if (token) headers.set('Authorization', `Bearer ${token}`);
  const body = init.body;
  const hasBodyType = body instanceof FormData || body instanceof URLSearchParams || typeof body === 'string';
  const wantsJson = !headers.has('Content-Type') && hasBodyType && !(body instanceof FormData) && !(body instanceof URLSearchParams);
  if (wantsJson) headers.set('Content-Type', 'application/json');
  return fetch(input, { ...init, headers });
}

export const AuthContext = createContext<AuthState>({ user: null, token: null, loading: true });
export const useAuthContext = () => useContext(AuthContext);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [auth, setAuth] = useState<AuthState>({ user: null, token: null, loading: true });

  useEffect(() => {
    const stored = getStoredAuth();
    if (stored.user && stored.token) {
      setAuth(stored);
    } else {
      setAuth({ user: null, token: null, loading: false });
    }
  }, []);

  return <AuthContext.Provider value={auth}>{children}</AuthContext.Provider>;
}

export function ProtectedRoute({ children, roles }: { children: ReactNode; roles?: string[] }) {
  const { user, loading } = useAuthContext();

  if (loading) {
    return createPortal(
      <div className="min-h-screen flex items-center justify-center">
        <p className="text-olive-700">Carregando...</p>
      </div>,
      document.body
    );
  }

  if (!user) {
    window.location.href = '/login';
    return null;
  }

  if (roles && !roles.some(r => user.roles.includes(r as Role))) {
    window.location.href = '/pipeline';
    return null;
  }

  return <>{children}</>;
}

export function getHomePath(roles: string[]): string {
  if (roles.includes('manager') || roles.includes('seller')) return '/pipeline';
  if (roles.includes('guide')) return '/visits';
  if (roles.includes('associate')) return '/my-leads';
  return '/pipeline';
}
