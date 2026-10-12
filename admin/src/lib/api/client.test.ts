import {
  afterEach,
  beforeEach,
  describe,
  expect,
  it,
  vi,
  type Mock,
} from "vitest";

// ---------------------------------------------------------------------------
// Module mocks (hoisted). axios is fully mocked: these tests exercise the
// interceptor/refresh-queue logic in client.ts, never the network.
// ---------------------------------------------------------------------------
const axiosMocks = vi.hoisted(() => {
  // The axios instance created by client.ts. It must be callable (client.ts
  // re-invokes `api(config)` to replay the original request after a refresh).
  const instance = Object.assign(
    vi.fn((_config?: unknown) =>
      Promise.resolve({ data: {}, status: 200, statusText: "OK" }),
    ),
    {
      interceptors: {
        request: { use: vi.fn() },
        response: { use: vi.fn() },
      },
    },
  );
  return { instance, post: vi.fn() };
});

vi.mock("axios", () => ({
  default: {
    create: vi.fn(() => axiosMocks.instance),
    post: axiosMocks.post,
  },
}));

// auth-store imports the Fluxbase SDK wrapper; stub it so no SDK/network code
// is pulled in. The real zustand stores are used.
vi.mock("@/lib/fluxbase-client", () => ({ setAuthToken: vi.fn() }));

import {
  ensureFreshAccessToken,
  getDashboardAuthHeaders,
} from "@/lib/api/client";
import { useAuthStore } from "@/stores/auth-store";
import { useBranchStore } from "@/stores/branch-store";
import { useImpersonationStore } from "@/stores/impersonation-store";
import { useTenantStore } from "@/stores/tenant-store";

// ---------------------------------------------------------------------------
// Interceptor handles: client.ts registers exactly one request and one
// response interceptor at module load.
// ---------------------------------------------------------------------------
const requestInterceptors = axiosMocks.instance.interceptors.request
  .use as unknown as Mock;
const responseInterceptors = axiosMocks.instance.interceptors.response
  .use as unknown as Mock;

const onRequestFulfilled: (config: unknown) => unknown =
  requestInterceptors.mock.calls[0]?.[0];
const onResponseFulfilled: (response: unknown) => unknown =
  responseInterceptors.mock.calls[0]?.[0];
const onResponseError: (error: unknown) => Promise<unknown> =
  responseInterceptors.mock.calls[0]?.[1];

interface FakeConfig {
  url?: string;
  headers: Record<string, string>;
  signal?: AbortSignal;
  _retry?: boolean;
}

function make401(config: FakeConfig) {
  return { config, response: { status: 401 } };
}

function refreshResponse(overrides: Record<string, unknown> = {}) {
  return {
    data: {
      access_token: "new-access-token",
      refresh_token: "new-refresh-token",
      user: {
        id: "user-1",
        email: "admin@example.com",
        role: "instance_admin",
      },
      expires_in: 900,
      ...overrides,
    },
  };
}

// Stub window.location: vitest's jsdom environment makes `window` the test
// global, so the stub is what `window.location.href = ...` in forceLogout
// writes to (jsdom's real Location is unforgeable and its navigation is a
// no-op outside a browser).
const locationStub: { href: string } = { href: "http://localhost:3000/admin" };

beforeEach(() => {
  // vitest 5 defaults to clearMocks: true (calls cleared before every test),
  // so per-test call assertions must only cover calls made DURING the test.
  // Interceptor handlers are captured once below, at module scope.
  axiosMocks.post.mockReset();
  axiosMocks.instance.mockClear();
  vi.stubGlobal("location", locationStub);
  locationStub.href = "http://localhost:3000/admin";

  useAuthStore.getState().auth.reset();
  useTenantStore.getState().clearTenant();
  useBranchStore.getState().setCurrentBranch(null);
  useImpersonationStore.getState().stopImpersonation();
});

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("api client request interceptor", () => {
  it("attaches the stored access token as a Bearer Authorization header", () => {
    useAuthStore.getState().auth.setTokens("access-123", "refresh-456");

    const config = onRequestFulfilled({ headers: {} }) as FakeConfig;

    expect(config.headers.Authorization).toBe("Bearer access-123");
  });

  it("does not overwrite an Authorization header the caller already set", () => {
    useAuthStore.getState().auth.setTokens("access-123", "refresh-456");

    const config = onRequestFulfilled({
      headers: { Authorization: "Bearer caller-token" },
    }) as FakeConfig;

    expect(config.headers.Authorization).toBe("Bearer caller-token");
  });

  it("omits Authorization when no access token is stored", () => {
    const config = onRequestFulfilled({ headers: {} }) as FakeConfig;

    expect(config.headers.Authorization).toBeUndefined();
  });

  it("propagates the tenant and non-main branch context headers", () => {
    useTenantStore.getState().setCurrentTenant({
      id: "tenant-1",
      slug: "acme",
      name: "Acme",
      is_default: false,
      metadata: {},
      created_at: "2026-01-01T00:00:00Z",
    });
    useBranchStore.getState().setCurrentBranch({
      id: "branch-1",
      name: "preview-1",
      slug: "preview-1",
      database_name: "branch_1",
      status: "ready",
      type: "preview",
      created_at: "2026-01-01T00:00:00Z",
      updated_at: "2026-01-01T00:00:00Z",
    });

    const config = onRequestFulfilled({ headers: {} }) as FakeConfig;

    expect(config.headers["X-FB-Tenant"]).toBe("tenant-1");
    expect(config.headers["X-Fluxbase-Branch"]).toBe("preview-1");
  });

  it("omits the branch header for the main branch", () => {
    useBranchStore.getState().setCurrentBranch({
      id: "branch-main",
      name: "main",
      slug: "main",
      database_name: "main",
      status: "ready",
      type: "main",
      created_at: "2026-01-01T00:00:00Z",
      updated_at: "2026-01-01T00:00:00Z",
    });

    const config = onRequestFulfilled({ headers: {} }) as FakeConfig;

    expect(config.headers["X-Fluxbase-Branch"]).toBeUndefined();
  });

  it("preserves the ApiRequestOptions signal on the outgoing config", () => {
    const controller = new AbortController();

    const config = onRequestFulfilled({
      headers: {},
      signal: controller.signal,
    }) as FakeConfig;

    expect(config.signal).toBe(controller.signal);
  });
});

describe("api client response interceptor", () => {
  it("passes successful responses through untouched (no body sniffing)", () => {
    const response = { data: { message: "unauthorized in a log row" } };

    // The fulfilled handler must be a passthrough: payloads containing words
    // like "unauthorized" must not trigger token churn.
    expect(onResponseFulfilled(response)).toBe(response);
  });

  it("rejects non-401 errors without attempting a refresh", async () => {
    const error = { config: { headers: {} }, response: { status: 500 } };

    await expect(onResponseError(error)).rejects.toBe(error);
    expect(axiosMocks.post).not.toHaveBeenCalled();
  });

  it("rejects 401s that have no request config without a refresh", async () => {
    const error = { config: undefined, response: { status: 401 } };

    await expect(onResponseError(error)).rejects.toBe(error);
    expect(axiosMocks.post).not.toHaveBeenCalled();
  });
});

describe("401 refresh single-flight queue", () => {
  it("triggers exactly one refresh for concurrent 401s and replays every queued request with the new token", async () => {
    useAuthStore.getState().auth.setTokens("stale-access", "valid-refresh");
    axiosMocks.post.mockResolvedValueOnce(refreshResponse());

    // Three requests fail with 401 "concurrently" (no awaiting between them).
    const cfg1: FakeConfig = { url: "/api/v1/admin/users", headers: {} };
    const cfg2: FakeConfig = { url: "/api/v1/admin/tenants", headers: {} };
    const cfg3: FakeConfig = { url: "/api/v1/admin/jobs", headers: {} };
    const p1 = onResponseError(make401(cfg1));
    const p2 = onResponseError(make401(cfg2));
    const p3 = onResponseError(make401(cfg3));

    await Promise.all([p1, p2, p3]);

    // Exactly one POST /refresh (single-flight, refresh tokens rotate on use).
    expect(axiosMocks.post).toHaveBeenCalledTimes(1);
    expect(axiosMocks.post).toHaveBeenCalledWith(
      "/api/v1/admin/refresh",
      { refresh_token: "valid-refresh" },
      { withCredentials: true },
    );

    // All three original requests were replayed through the axios instance.
    expect(axiosMocks.instance).toHaveBeenCalledTimes(3);
    const replayedUrls = axiosMocks.instance.mock.calls.map(
      (c) => (c[0] as FakeConfig).url,
    );
    expect(replayedUrls).toEqual([
      "/api/v1/admin/users",
      "/api/v1/admin/tenants",
      "/api/v1/admin/jobs",
    ]);

    // Every replay (triggering + queued) carries the fresh token.
    for (const cfg of [cfg1, cfg2, cfg3]) {
      expect(cfg.headers.Authorization).toBe("Bearer new-access-token");
      expect(cfg._retry).toBe(true);
    }

    // The new tokens were adopted into the auth store.
    const auth = useAuthStore.getState().auth;
    expect(auth.accessToken).toBe("new-access-token");
    expect(auth.refreshToken).toBe("new-refresh-token");
  });

  it("updates the stored user from the refresh payload", async () => {
    axiosMocks.post.mockResolvedValueOnce(refreshResponse());

    await onResponseError(make401({ headers: {} }));

    const user = useAuthStore.getState().auth.user;
    expect(user).toMatchObject({
      accountNo: "user-1",
      email: "admin@example.com",
      role: ["instance_admin"],
    });
    expect(user?.exp).toBeGreaterThan(Date.now());
  });

  it("logs out and rejects queued requests when the refresh fails", async () => {
    useAuthStore.getState().auth.setTokens("stale-access", "dead-refresh");
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
    axiosMocks.post.mockRejectedValueOnce(new Error("refresh rejected"));

    const cfg1: FakeConfig = { headers: {} };
    const cfg2: FakeConfig = { headers: {} };
    const p1 = onResponseError(make401(cfg1));
    const p2 = onResponseError(make401(cfg2));

    // Queued requests reject with the refresh error...
    await expect(p2).rejects.toThrow("refresh rejected");

    // ...and forceLogout wiped auth + impersonation and redirected.
    const auth = useAuthStore.getState().auth;
    expect(auth.accessToken).toBe("");
    expect(auth.refreshToken).toBe("");
    expect(auth.user).toBeNull();
    expect(useImpersonationStore.getState().isImpersonating).toBe(false);
    expect(locationStub.href).toBe("/admin/login");

    // The triggering request never resolves (forceLogout returns a pending
    // promise to swallow the navigation) and no retry was attempted.
    expect(axiosMocks.instance).not.toHaveBeenCalled();
    let settled = false;
    void p1.then(() => {
      settled = true;
    });
    await new Promise((resolve) => setTimeout(resolve, 20));
    expect(settled).toBe(false);
  });

  it("logs out immediately on a 401 replay that 401s again (_retry guard)", async () => {
    axiosMocks.post.mockResolvedValueOnce(refreshResponse());
    // The replayed request 401s a second time.
    axiosMocks.instance.mockResolvedValueOnce({
      data: {},
      status: 401,
    } as never);

    const first = make401({ headers: {} });
    await onResponseError(first);

    expect(axiosMocks.post).toHaveBeenCalledTimes(1);

    // Second 401 on an already-retried config: straight to forceLogout.
    const cfg: FakeConfig = { headers: {}, _retry: true };
    const p = onResponseError(make401(cfg));
    expect(locationStub.href).toBe("/admin/login");
    let settled = false;
    void p.then(() => {
      settled = true;
    });
    await new Promise((resolve) => setTimeout(resolve, 20));
    expect(settled).toBe(false);
  });
});

describe("ensureFreshAccessToken", () => {
  it("returns true after a successful refresh and adopts the new tokens", async () => {
    axiosMocks.post.mockResolvedValueOnce(
      refreshResponse({ refresh_token: "rotated-refresh" }),
    );

    await expect(ensureFreshAccessToken()).resolves.toBe(true);
    expect(useAuthStore.getState().auth.accessToken).toBe("new-access-token");
    expect(useAuthStore.getState().auth.refreshToken).toBe("rotated-refresh");
  });

  it("returns false on refresh failure without logging out", async () => {
    axiosMocks.post.mockRejectedValueOnce(new Error("network down"));

    await expect(ensureFreshAccessToken()).resolves.toBe(false);
    expect(locationStub.href).toBe("http://localhost:3000/admin");
  });

  it("shares one in-flight refresh across concurrent callers", async () => {
    axiosMocks.post.mockResolvedValueOnce(refreshResponse());

    const [a, b, c] = await Promise.all([
      ensureFreshAccessToken(),
      ensureFreshAccessToken(),
      ensureFreshAccessToken(),
    ]);

    expect([a, b, c]).toEqual([true, true, true]);
    expect(axiosMocks.post).toHaveBeenCalledTimes(1);
  });
});

describe("forceLogout side effects (via refresh failure)", () => {
  it("clears the dashboard auth headers", async () => {
    useAuthStore.getState().auth.setTokens("stale-access", "dead-refresh");
    expect(getDashboardAuthHeaders()).toEqual({
      Authorization: "Bearer stale-access",
    });
    axiosMocks.post.mockRejectedValueOnce(new Error("refresh rejected"));

    void onResponseError(make401({ headers: {} }));
    await new Promise((resolve) => setTimeout(resolve, 20));

    expect(getDashboardAuthHeaders()).toEqual({});
  });
});

describe("getDashboardAuthHeaders", () => {
  it("returns an empty object when no access token is set", () => {
    expect(getDashboardAuthHeaders()).toEqual({});
  });

  it("returns the Bearer header when a token is set", () => {
    useAuthStore.getState().auth.setTokens("tok-1", "r-1");
    expect(getDashboardAuthHeaders()).toEqual({
      Authorization: "Bearer tok-1",
    });
  });
});

describe("signal propagation through the retry path", () => {
  it("replays the original request with its AbortSignal intact", async () => {
    axiosMocks.post.mockResolvedValueOnce(refreshResponse());
    const controller = new AbortController();
    const cfg: FakeConfig = { headers: {}, signal: controller.signal };

    await onResponseError(make401(cfg));

    expect(axiosMocks.instance).toHaveBeenCalledTimes(1);
    const replayed = axiosMocks.instance.mock.calls[0]?.[0] as FakeConfig;
    expect(replayed.signal).toBe(controller.signal);
    expect(replayed._retry).toBe(true);
    expect(replayed.headers.Authorization).toBe("Bearer new-access-token");
  });
});
