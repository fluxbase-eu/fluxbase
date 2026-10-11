import { useAuthStore } from "@/stores/auth-store";
import { useImpersonationStore } from "@/stores/impersonation-store";
import { ensureFreshAccessToken } from "@/lib/api/client";
import { removeCookie } from "@/lib/cookies";

const USER_KEY = "fluxbase_admin_user";
// Legacy JS-readable refresh copies written by older builds. They are read
// once (adopted into memory by the auth store on startup) and scrubbed here
// and on logout; the server-side HttpOnly `fluxbase_admin_refresh` cookie is
// the authoritative refresh storage now.
const LEGACY_REFRESH_SESSION_KEY = "fluxbase_admin_refresh_token";
const LEGACY_REFRESH_COOKIE_NAME = "fluxbase_admin_refresh_token";

export interface AdminUser {
  id: string;
  email: string;
  role: string;
  email_verified: boolean;
  metadata?: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

export interface DashboardUser {
  id: string;
  email: string;
  email_verified: boolean;
  full_name: string | null;
  avatar_url: string | null;
  totp_enabled: boolean;
  is_active: boolean;
  is_locked: boolean;
  last_login_at: string | null;
  created_at: string;
  updated_at: string;
  role?: string;
}

export interface TokenPair {
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

export function getAccessToken(): string | null {
  if (typeof window === "undefined") return null;
  const token = useAuthStore.getState().auth.accessToken;
  return token || null;
}

export function getRefreshToken(): string | null {
  if (typeof window === "undefined") return null;
  const token = useAuthStore.getState().auth.refreshToken;
  return token || null;
}

export function getStoredUser(): AdminUser | DashboardUser | null {
  if (typeof window === "undefined") return null;
  const userJson =
    localStorage.getItem(USER_KEY) || localStorage.getItem("user");
  if (!userJson) return null;
  try {
    return JSON.parse(userJson);
  } catch {
    return null;
  }
}

export function setTokens(tokens: TokenPair, user: AdminUser): void {
  if (typeof window === "undefined") return;
  const store = useAuthStore.getState();
  store.auth.setTokens(tokens.access_token, tokens.refresh_token);
  store.auth.setUser({
    accountNo: user.id,
    email: user.email,
    role: [user.role],
    exp: Date.now() + tokens.expires_in * 1000,
  });
  localStorage.setItem(USER_KEY, JSON.stringify(user));
}

export function clearTokens(): void {
  if (typeof window === "undefined") return;
  // End any active impersonation session when the admin logs out
  useImpersonationStore.getState().stopImpersonation();
  useAuthStore.getState().auth.reset();
  localStorage.removeItem(USER_KEY);
  localStorage.removeItem("user");
  localStorage.removeItem("fluxbase_admin_access_token");
  localStorage.removeItem("fluxbase_admin_refresh_token");
  localStorage.removeItem("access_token");
  localStorage.removeItem("refresh_token");
  // Legacy JS-readable refresh copies (sessionStorage key + cookie).
  try {
    sessionStorage.removeItem(LEGACY_REFRESH_SESSION_KEY);
  } catch {
    // sessionStorage may be unavailable (privacy mode)
  }
  removeCookie(LEGACY_REFRESH_COOKIE_NAME);
}

export function isAuthenticated(): boolean {
  return !!getAccessToken();
}

/**
 * Silent session restore for cold page loads.
 *
 * The access token cookie restores the session instantly; when it is missing
 * (or already expired) but the HttpOnly `fluxbase_admin_refresh` cookie is
 * still present, this performs one silent token refresh so the user does not
 * have to sign in again. Safe to call on every startup: it is a no-op when a
 * valid access token exists and simply resolves to logged-out when no
 * refresh cookie is present.
 */
export async function attemptSessionRestore(): Promise<void> {
  if (typeof window === "undefined") return;
  if (useAuthStore.getState().auth.accessToken) return;
  await ensureFreshAccessToken();
}

export function logout(): void {
  clearTokens();
  window.location.href = "/admin/login";
}
