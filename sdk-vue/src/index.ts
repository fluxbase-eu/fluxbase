/**
 * Fluxbase Vue SDK
 *
 * Vue/Nuxt composables and SSR cookie storage for Fluxbase, built on the core
 * `@nimbleflux/fluxbase-sdk`.
 *
 * Provides the SSR-auth foundation (cookie storage, provide/inject
 * composable) plus AI composables (chatbots, conversations, knowledge-base
 * search, vector search) as thin reactive wrappers over the core SDK.
 */

export {
  createCookieStorage,
  type CookieStorageOptions,
  type NuxtCookieEvent,
} from "./cookie-storage";

export {
  FLUXBASE_CLIENT_KEY,
  useFluxbaseClient,
} from "./composables";

export {
  useFluxbaseAIChatbots,
  useFluxbaseAIConversations,
  useFluxbaseAIUsage,
  useFluxbaseKnowledgeBases,
  useFluxbaseKBSearch,
  useFluxbaseVectorSearch,
  useFluxbaseVectorCapabilities,
} from "./ai";

export type { FluxbaseClient, StorageAdapter } from "@nimbleflux/fluxbase-sdk";
