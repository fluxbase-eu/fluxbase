import { beforeEach, describe, expect, it, vi } from "vitest";

// client.ts is mocked: lib/auth.ts only consumes ensureFreshAccessToken from
// it (attemptSessionRestore), and mocking keeps the axios instance + SDK out
// of this suite entirely.
const ensureFreshAccessToken = vi.hoisted(() => vi.fn());

vi.mock("@/lib/api/client", () => ({ ensureFreshAccessToken }));

import {
  attemptSessionRestore,
  clearTokens,
  getAccessToken,
  getRefreshToken,
  getStoredUser,
  isAuthenticated,
  logout,
  setTokens,
  type AdminUser,
} from "@/lib/auth";
import { useAuthStore } from "@/stores/auth-store";
import { useImpersonationStore } from "@/stores/impersonation-store";

const LEGACY_REFRESH_SESSION_KEY = "fluxbase_admin_refresh_token";
const LEGACY_REFRESH_COOKIE_NAME = "fluxbase_admin_refresh_token";

function setCookie(name: string, value: string) {
  document.cookie = `${name}=${value}; path=/`;
}

function getCookie(name: string): string | undefined {
  const match = document.cookie.match(new RegExp(`(?:^|; )${name}=([^;]*)`));
  return match?.[1];
}

function removeCookie(name: string) {
  document.cookie = `${name}=; path=/; max-age=0`;
}

const sampleAdminUser: AdminUser = {
  id: "user-1",
  email: "admin@example.com",
  role: "instance_admin",
  email_verified: true,
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-01-01T00:00:00Z",
};

// window.location stub (vitest jsdom: window === test global, so this is
// what logout()'s redirect writes to).
const locationStub: { href: string } = { href: "http://localhost:3000/admin" };

beforeEach(() => {
  sessionStorage.clear();
  localStorage.clear();
  removeCookie(LEGACY_REFRESH_COOKIE_NAME);
  useImpersonationStore.getState().stopImpersonation();
  const store = useAuthStore.getState().auth;
  store.reset();
  ensureFreshAccessToken.mockReset();
  locationStub.href = "http://localhost:3000/admin";
  vi.stubGlobal("location", locationStub);
});

describe("token accessors", () => {
  it("return null when the store has no tokens", () => {
    expect(getAccessToken()).toBeNull();
    expect(getRefreshToken()).toBeNull();
    expect(isAuthenticated()).toBe(false);
  });

  it("return the store tokens when set", () => {
    useAuthStore.getState().auth.setTokens("access-1", "refresh-1");

    expect(getAccessToken()).toBe("access-1");
    expect(getRefreshToken()).toBe("refresh-1");
    expect(isAuthenticated()).toBe(true);
  });
});

describe("setTokens", () => {
  it("persists tokens in the store, maps the user, and stores the profile", () => {
    setTokens(
      {
        access_token: "access-1",
        refresh_token: "refresh-1",
        expires_in: 900,
      },
      sampleAdminUser,
    );

    const auth = useAuthStore.getState().auth;
    expect(auth.accessToken).toBe("access-1");
    expect(auth.refreshToken).toBe("refresh-1");
    expect(auth.user).toEqual({
      accountNo: "user-1",
      email: "admin@example.com",
      role: ["instance_admin"],
      exp: expect.any(Number),
    });
    expect(auth.user?.exp).toBeGreaterThan(Date.now());
    expect(
      JSON.parse(localStorage.getItem("fluxbase_admin_user") ?? ""),
    ).toEqual(sampleAdminUser);
  });
});

describe("getStoredUser", () => {
  it("returns null when nothing is stored", () => {
    expect(getStoredUser()).toBeNull();
  });

  it("parses the stored dashboard user", () => {
    localStorage.setItem(
      "fluxbase_admin_user",
      JSON.stringify(sampleAdminUser),
    );
    expect(getStoredUser()).toEqual(sampleAdminUser);
  });

  it('falls back to the ancient "user" key and tolerates corrupt JSON', () => {
    localStorage.setItem("user", JSON.stringify(sampleAdminUser));
    expect(getStoredUser()).toEqual(sampleAdminUser);

    localStorage.setItem("fluxbase_admin_user", "{broken");
    localStorage.removeItem("user");
    expect(getStoredUser()).toBeNull();
  });
});

describe("clearTokens (logout scrubbing)", () => {
  it("wipes the store, every legacy localStorage copy, and impersonation", () => {
    // Seed everything a long-lived browser could have accumulated.
    setTokens(
      {
        access_token: "access-1",
        refresh_token: "refresh-1",
        expires_in: 900,
      },
      sampleAdminUser,
    );
    localStorage.setItem("fluxbase_admin_access_token", "legacy-a");
    localStorage.setItem("fluxbase_admin_refresh_token", "legacy-r");
    localStorage.setItem("access_token", "older-a");
    localStorage.setItem("refresh_token", "older-r");
    sessionStorage.setItem(LEGACY_REFRESH_SESSION_KEY, JSON.stringify("sess"));
    setCookie(LEGACY_REFRESH_COOKIE_NAME, JSON.stringify("cookie"));
    useImpersonationStore.getState().startImpersonation(
      "imp-token",
      "imp-refresh",
      { id: "target-1", email: "target@example.com" },
      {
        id: "session-1",
        admin_user_id: "admin-1",
        impersonation_type: "user",
        reason: "support",
        started_at: "2026-01-01T00:00:00Z",
        is_active: true,
      },
      "user",
    );

    clearTokens();

    const auth = useAuthStore.getState().auth;
    expect(auth.accessToken).toBe("");
    expect(auth.refreshToken).toBe("");
    expect(auth.user).toBeNull();
    for (const key of [
      "fluxbase_admin_user",
      "user",
      "fluxbase_admin_access_token",
      "fluxbase_admin_refresh_token",
      "access_token",
      "refresh_token",
    ]) {
      expect(localStorage.getItem(key)).toBeNull();
    }
    expect(sessionStorage.getItem(LEGACY_REFRESH_SESSION_KEY)).toBeNull();
    expect(getCookie(LEGACY_REFRESH_COOKIE_NAME)).toBeUndefined();
    expect(useImpersonationStore.getState().isImpersonating).toBe(false);
  });
});

describe("attemptSessionRestore", () => {
  it("is a no-op when a valid access token is already present", async () => {
    useAuthStore.getState().auth.setTokens("access-1", "refresh-1");

    await attemptSessionRestore();

    expect(ensureFreshAccessToken).not.toHaveBeenCalled();
  });

  it("performs one silent refresh when no access token exists", async () => {
    ensureFreshAccessToken.mockResolvedValue(true);

    await attemptSessionRestore();

    expect(ensureFreshAccessToken).toHaveBeenCalledTimes(1);
  });

  it("resolves even when the silent refresh fails", async () => {
    ensureFreshAccessToken.mockResolvedValue(false);

    await expect(attemptSessionRestore()).resolves.toBeUndefined();
  });
});

describe("logout", () => {
  it("clears tokens and redirects to the login page", () => {
    useAuthStore.getState().auth.setTokens("access-1", "refresh-1");

    logout();

    expect(useAuthStore.getState().auth.accessToken).toBe("");
    expect(locationStub.href).toBe("/admin/login");
  });
});
