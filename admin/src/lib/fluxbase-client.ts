import { createClient } from "@nimbleflux/fluxbase-sdk";
import { useAuthStore } from "@/stores/auth-store";
import { useTenantStore } from "@/stores/tenant-store";
import { useImpersonationStore } from "@/stores/impersonation-store";
import { ensureFreshAccessToken } from "@/lib/api/client";

// Declare the runtime config type injected by the server
declare global {
  interface Window {
    __FLUXBASE_CONFIG__?: {
      publicBaseURL?: string;
    };
  }
}

// Base URL for the API - priority order:
// 1. Runtime config injected by server (FLUXBASE_PUBLIC_BASE_URL)
// 2. Build-time environment variable (VITE_API_URL)
// 3. Current origin (works when dashboard is served from the same domain)
const API_BASE_URL =
  window.__FLUXBASE_CONFIG__?.publicBaseURL ||
  import.meta.env.VITE_API_URL ||
  window.location.origin;
const API_KEY = import.meta.env.VITE_API_KEY || "anonymous";

const getImpersonationToken = (): string | null => {
  return useImpersonationStore.getState().impersonationToken;
};

const getActiveToken = (): string | null => {
  const accessToken = useAuthStore.getState().auth.accessToken;
  return getImpersonationToken() || accessToken || null;
};

// Create the Fluxbase client
export const fluxbaseClient = createClient(API_BASE_URL, API_KEY, {
  auth: {
    autoRefresh: false,
    persist: false,
  },
  timeout: 30000,
});

fluxbaseClient.setBeforeRequestCallback((headers) => {
  const currentTenant = useTenantStore.getState().currentTenant;
  if (currentTenant?.id) {
    headers["X-FB-Tenant"] = currentTenant.id;
  } else {
    delete headers["X-FB-Tenant"];
  }
});

// Helper to set the auth token
export function setAuthToken(token: string | null) {
  if (token) {
    fluxbaseClient.setAuthToken(token);
  } else {
    fluxbaseClient.setAuthToken(null);
  }
}

// On a 401, let the SDK retry through the SAME refresh flow the axios client
// uses (single in-flight refresh in lib/api/client.ts). A successful refresh
// updates the auth store, which pushes the new access token into this client
// via setAuthToken, so the SDK's automatic retry picks it up.
//
// FluxbaseClient does not expose the underlying FluxbaseFetch (its `fetch`
// field is private) and FluxbaseAuth unconditionally registers its own
// refresh callback on it during createClient(). The admin app never uses
// client.auth sessions (persist: false, autoRefresh: false), so replacing
// that callback is safe: the SDK's own refreshSession() path would be a
// no-op here anyway.
const sdkFetch = (
  fluxbaseClient as unknown as {
    fetch: {
      setRefreshTokenCallback: (cb: () => Promise<boolean>) => void;
    };
  }
).fetch;
sdkFetch.setRefreshTokenCallback(() => ensureFreshAccessToken());

// Helper to get the auth token
export function getAuthToken(): string | null {
  return fluxbaseClient.getAuthToken();
}

// Helper to sync SDK token with current active token (impersonation or admin)
export function syncAuthToken() {
  const activeToken = getActiveToken();
  setAuthToken(activeToken);
}

export default fluxbaseClient;
