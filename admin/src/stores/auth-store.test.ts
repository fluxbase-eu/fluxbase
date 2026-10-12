import { beforeEach, describe, expect, it, vi } from "vitest";

// The Fluxbase SDK wrapper is stubbed: the store only uses setAuthToken to
// mirror the access token into the SDK client. Keeps tests offline.
vi.mock("@/lib/fluxbase-client", () => ({ setAuthToken: vi.fn() }));

// The auth store reads browser storage ONCE at module init (cookie session
// restore + legacy refresh-token adoption), so every test re-imports the
// module with a clean storage state.
async function freshStore() {
  vi.resetModules();
  const mod = await import("@/stores/auth-store");
  return mod.useAuthStore;
}

async function freshSetAuthToken() {
  const mod = await import("@/lib/fluxbase-client");
  return vi.mocked(mod.setAuthToken);
}

const ACCESS_COOKIE = "fluxbase_admin_token";
const LEGACY_REFRESH_KEY = "fluxbase_admin_refresh_token";
const LEGACY_REFRESH_COOKIE = "fluxbase_admin_refresh_token";

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

beforeEach(() => {
  // Fresh browser storage for every test.
  sessionStorage.clear();
  localStorage.clear();
  removeCookie(ACCESS_COOKIE);
  removeCookie(LEGACY_REFRESH_COOKIE);
});

describe("auth store initial state", () => {
  it("starts signed out with no tokens or user on a cold browser", async () => {
    const store = await freshStore();
    const auth = store.getState().auth;

    expect(auth.user).toBeNull();
    expect(auth.accessToken).toBe("");
    expect(auth.refreshToken).toBe("");
  });

  it("restores the access token from the JS-readable session cookie", async () => {
    setCookie(ACCESS_COOKIE, JSON.stringify("cookie-access-token"));

    const store = await freshStore();
    const auth = store.getState().auth;

    expect(auth.accessToken).toBe("cookie-access-token");
    expect(auth.refreshToken).toBe("");
    expect(await freshSetAuthToken()).toHaveBeenCalledWith(
      "cookie-access-token",
    );
  });

  it("drops a corrupt access-token cookie instead of crashing", async () => {
    setCookie(ACCESS_COOKIE, "%not-json%");

    const store = await freshStore();

    expect(store.getState().auth.accessToken).toBe("");
    // The unreadable cookie is removed so it is not re-parsed next boot.
    expect(getCookie(ACCESS_COOKIE)).toBeUndefined();
  });
});

describe("legacy refresh token adoption (one transition release)", () => {
  it("adopts the sessionStorage refresh copy into memory and clears the key", async () => {
    sessionStorage.setItem(
      LEGACY_REFRESH_KEY,
      JSON.stringify("legacy-refresh"),
    );

    const store = await freshStore();

    expect(store.getState().auth.refreshToken).toBe("legacy-refresh");
    expect(sessionStorage.getItem(LEGACY_REFRESH_KEY)).toBeNull();
  });

  it("adopts the legacy JS-readable refresh cookie and removes it", async () => {
    setCookie(LEGACY_REFRESH_COOKIE, JSON.stringify("legacy-cookie-refresh"));

    const store = await freshStore();

    expect(store.getState().auth.refreshToken).toBe("legacy-cookie-refresh");
    expect(getCookie(LEGACY_REFRESH_COOKIE)).toBeUndefined();
  });

  it("prefers the sessionStorage copy when both legacy stores exist", async () => {
    sessionStorage.setItem(LEGACY_REFRESH_KEY, JSON.stringify("session-copy"));
    setCookie(LEGACY_REFRESH_COOKIE, JSON.stringify("cookie-copy"));

    const store = await freshStore();

    expect(store.getState().auth.refreshToken).toBe("session-copy");
    expect(getCookie(LEGACY_REFRESH_COOKIE)).toBeUndefined();
  });

  it("ignores a malformed sessionStorage value but still scrubs it", async () => {
    sessionStorage.setItem(LEGACY_REFRESH_KEY, "{not json");

    const store = await freshStore();

    expect(store.getState().auth.refreshToken).toBe("");
    expect(sessionStorage.getItem(LEGACY_REFRESH_KEY)).toBeNull();
  });

  it("starts with no refresh token when no legacy copy exists", async () => {
    const store = await freshStore();

    expect(store.getState().auth.refreshToken).toBe("");
    expect(sessionStorage.getItem(LEGACY_REFRESH_KEY)).toBeNull();
  });
});

describe("setTokens / setAccessToken / setUser", () => {
  it("stores both tokens in memory and persists the access token cookie", async () => {
    const store = await freshStore();

    store.getState().auth.setTokens("access-1", "refresh-1");

    const auth = store.getState().auth;
    expect(auth.accessToken).toBe("access-1");
    expect(auth.refreshToken).toBe("refresh-1");
    expect(getCookie(ACCESS_COOKIE)).toBe(JSON.stringify("access-1"));
    expect(await freshSetAuthToken()).toHaveBeenCalledWith("access-1");
  });

  it("never persists the refresh token to JS-readable storage", async () => {
    const store = await freshStore();

    store.getState().auth.setTokens("access-1", "refresh-secret");

    expect(sessionStorage.getItem(LEGACY_REFRESH_KEY)).toBeNull();
    expect(getCookie(LEGACY_REFRESH_COOKIE)).toBeUndefined();
    expect(document.cookie).not.toContain("refresh-secret");
  });

  it("setAccessToken updates only the access token and syncs the SDK", async () => {
    const store = await freshStore();
    store.getState().auth.setTokens("access-1", "refresh-1");

    store.getState().auth.setAccessToken("access-2");

    const auth = store.getState().auth;
    expect(auth.accessToken).toBe("access-2");
    expect(auth.refreshToken).toBe("refresh-1");
    expect(await freshSetAuthToken()).toHaveBeenCalledWith("access-2");
  });

  it("setUser stores the dashboard user object", async () => {
    const store = await freshStore();
    const user = {
      accountNo: "u-1",
      email: "admin@example.com",
      role: ["instance_admin"],
      exp: 1893456000000,
    };

    store.getState().auth.setUser(user);

    expect(store.getState().auth.user).toEqual(user);
  });
});

describe("resetAccessToken / reset", () => {
  it("resetAccessToken clears only the access token and the cookie", async () => {
    const store = await freshStore();
    store.getState().auth.setTokens("access-1", "refresh-1");
    store.getState().auth.setUser({
      accountNo: "u-1",
      email: "a@b.c",
      role: ["instance_admin"],
      exp: 1893456000000,
    });

    store.getState().auth.resetAccessToken();

    const auth = store.getState().auth;
    expect(auth.accessToken).toBe("");
    expect(auth.refreshToken).toBe("refresh-1");
    expect(auth.user).not.toBeNull();
    expect(getCookie(ACCESS_COOKIE)).toBeUndefined();
    expect(await freshSetAuthToken()).toHaveBeenCalledWith(null);
  });

  it("reset clears tokens, user, the cookie and every legacy refresh copy", async () => {
    sessionStorage.setItem(LEGACY_REFRESH_KEY, JSON.stringify("legacy"));
    setCookie(LEGACY_REFRESH_COOKIE, JSON.stringify("legacy-cookie"));
    const store = await freshStore();
    store.getState().auth.setTokens("access-1", "refresh-1");

    store.getState().auth.reset();

    const auth = store.getState().auth;
    expect(auth.accessToken).toBe("");
    expect(auth.refreshToken).toBe("");
    expect(auth.user).toBeNull();
    expect(getCookie(ACCESS_COOKIE)).toBeUndefined();
    expect(sessionStorage.getItem(LEGACY_REFRESH_KEY)).toBeNull();
    expect(getCookie(LEGACY_REFRESH_COOKIE)).toBeUndefined();
    expect(await freshSetAuthToken()).toHaveBeenCalledWith(null);
  });
});
