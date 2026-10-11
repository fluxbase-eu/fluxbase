import { create } from 'zustand'
import { getCookie, setCookie, removeCookie } from '@/lib/cookies'
import { setAuthToken as setFluxbaseAuthToken } from '@/lib/fluxbase-client'

const AUTH_COOKIE_NAME = 'fluxbase_admin_token'
// Token storage design: the ACCESS token lives in a JS-readable cookie so a
// page reload restores the session without extra requests. The REFRESH token
// is never persisted in JavaScript-readable storage: the server sets it as an
// HttpOnly cookie (`fluxbase_admin_refresh`, scoped to /api/v1/admin) and the
// client only keeps it in memory for the current tab. When the access token
// is gone but the HttpOnly refresh cookie still exists, the app performs a
// silent refresh (`attemptSessionRestore` in lib/auth.ts) to rebuild the
// session. Legacy JS-readable refresh copies (sessionStorage key and cookie
// below) are adopted into memory and cleared on startup for one transition
// release.
const REFRESH_SESSION_KEY = 'fluxbase_admin_refresh_token'
const LEGACY_REFRESH_COOKIE_NAME = 'fluxbase_admin_refresh_token'

interface AuthUser {
  accountNo: string
  email: string
  role: string[]
  exp: number
}

interface AuthState {
  auth: {
    user: AuthUser | null
    setUser: (user: AuthUser | null) => void
    accessToken: string
    refreshToken: string
    setAccessToken: (accessToken: string) => void
    setTokens: (accessToken: string, refreshToken: string) => void
    resetAccessToken: () => void
    reset: () => void
  }
}

function parseCookieToken(name: string): string {
  try {
    const raw = getCookie(name)
    return raw ? JSON.parse(raw) : ''
  } catch {
    removeCookie(name)
    return ''
  }
}

// One-transition-release migration: older builds persisted the refresh token
// in sessionStorage and a JS-readable cookie. If either is present, adopt it
// into memory (so the in-flight refresh flow keeps working) and clear both.
function adoptLegacySessionRefreshToken(): string {
  let token = ''

  if (typeof window !== 'undefined') {
    try {
      const raw = sessionStorage.getItem(REFRESH_SESSION_KEY)
      if (raw) token = JSON.parse(raw) || ''
    } catch {
      // ignore malformed values
    }
    try {
      sessionStorage.removeItem(REFRESH_SESSION_KEY)
    } catch {
      // sessionStorage may be unavailable (privacy mode)
    }
  }

  const legacyCookie = parseCookieToken(LEGACY_REFRESH_COOKIE_NAME)
  if (legacyCookie) {
    removeCookie(LEGACY_REFRESH_COOKIE_NAME)
    if (!token) token = legacyCookie
  }

  return token
}

function clearSessionRefreshToken(): void {
  if (typeof window === 'undefined') return
  try {
    sessionStorage.removeItem(REFRESH_SESSION_KEY)
  } catch {
    // ignore
  }
}

export const useAuthStore = create<AuthState>()((set) => {
  const initToken = parseCookieToken(AUTH_COOKIE_NAME)
  const initRefreshToken = adoptLegacySessionRefreshToken()

  if (initToken) {
    setFluxbaseAuthToken(initToken)
  }

  return {
    auth: {
      user: null,
      setUser: (user) =>
        set((state) => ({ ...state, auth: { ...state.auth, user } })),
      accessToken: initToken,
      refreshToken: initRefreshToken,
      setAccessToken: (accessToken) =>
        set((state) => {
          setCookie(AUTH_COOKIE_NAME, JSON.stringify(accessToken))
          setFluxbaseAuthToken(accessToken)
          return { ...state, auth: { ...state.auth, accessToken } }
        }),
      // The refresh token is kept in memory only; the server persists the
      // authoritative copy in the HttpOnly `fluxbase_admin_refresh` cookie.
      setTokens: (accessToken, refreshToken) =>
        set((state) => {
          setCookie(AUTH_COOKIE_NAME, JSON.stringify(accessToken))
          setFluxbaseAuthToken(accessToken)
          return {
            ...state,
            auth: { ...state.auth, accessToken, refreshToken },
          }
        }),
      resetAccessToken: () =>
        set((state) => {
          removeCookie(AUTH_COOKIE_NAME)
          setFluxbaseAuthToken(null)
          return { ...state, auth: { ...state.auth, accessToken: '' } }
        }),
      reset: () =>
        set((state) => {
          removeCookie(AUTH_COOKIE_NAME)
          // Scrub any legacy JS-readable refresh copies too.
          clearSessionRefreshToken()
          removeCookie(LEGACY_REFRESH_COOKIE_NAME)
          setFluxbaseAuthToken(null)
          return {
            ...state,
            auth: {
              ...state.auth,
              user: null,
              accessToken: '',
              refreshToken: '',
            },
          }
        }),
    },
  }
})
