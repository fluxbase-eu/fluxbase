/**
 * AI composables for the Fluxbase Vue SDK.
 *
 * Thin, dependency-free wrappers around the core SDK's AI surface
 * (chatbots, conversations, knowledge bases, vector search). Each composable
 * returns reactive refs plus a `run`/`refresh` function — no external query
 * library required.
 *
 * @example
 * ```vue
 * <script setup lang="ts">
 * import { useFluxbaseAIChatbots } from '@nimbleflux/fluxbase-sdk-vue'
 * const { data, refresh } = useFluxbaseAIChatbots()
 * </script>
 * ```
 */

import { ref, type Ref } from "vue";
import type {
  AIChatbotSummary,
  AIDailyQuotaSnapshot,
  ListConversationsResult,
  FluxbaseError,
  KnowledgeBaseSummary,
  SearchKnowledgeBaseRequest,
  SearchKnowledgeBaseResponse,
  VectorCapabilities,
  VectorSearchOptions,
  VectorSearchResult,
} from "@nimbleflux/fluxbase-sdk";
import { useFluxbaseClient } from "./composables";

/** Reactive state shared by the AI composables. */
interface AsyncState<T> {
  data: Ref<T | null>;
  error: Ref<FluxbaseError | null>;
  loading: Ref<boolean>;
}

/** Runs an async fetcher into reactive state. */
async function runInto<T>(
  state: AsyncState<T>,
  fetcher: () => Promise<{ data: T | null; error: FluxbaseError | null }>,
): Promise<T | null> {
  state.loading.value = true;
  state.error.value = null;
  try {
    const { data, error } = await fetcher();
    state.error.value = error;
    state.data.value = data;
    return data;
  } finally {
    state.loading.value = false;
  }
}

/**
 * List public/enabled AI chatbots.
 */
export function useFluxbaseAIChatbots(): AsyncState<AIChatbotSummary[]> & {
  refresh: () => Promise<AIChatbotSummary[] | null>;
} {
  const client = useFluxbaseClient();
  const state: AsyncState<AIChatbotSummary[]> = {
    data: ref(null),
    error: ref(null),
    loading: ref(false),
  };
  const refresh = () => runInto(state, () => client.ai.listChatbots());
  void refresh();
  return { ...state, refresh };
}

/**
 * List the current user's AI conversations.
 */
export function useFluxbaseAIConversations(): AsyncState<ListConversationsResult> & {
  refresh: () => Promise<ListConversationsResult | null>;
} {
  const client = useFluxbaseClient();
  const state: AsyncState<ListConversationsResult> = {
    data: ref(null),
    error: ref(null),
    loading: ref(false),
  };
  const refresh = () => runInto(state, () => client.ai.listConversations());
  void refresh();
  return { ...state, refresh };
}

/**
 * Get the current user's usage for a chatbot (requests, tokens, limits).
 * Does not auto-run; call `fetch(chatbotId)`.
 */
export function useFluxbaseAIUsage(): AsyncState<AIDailyQuotaSnapshot> & {
  fetch: (chatbotId: string) => Promise<AIDailyQuotaSnapshot | null>;
} {
  const client = useFluxbaseClient();
  const state: AsyncState<AIDailyQuotaSnapshot> = {
    data: ref(null),
    error: ref(null),
    loading: ref(false),
  };
  const fetch = (chatbotId: string) =>
    runInto(state, () => client.ai.getUsage(chatbotId));
  return { ...state, fetch };
}

/**
 * List knowledge bases accessible to the current user.
 */
export function useFluxbaseKnowledgeBases(): AsyncState<
  KnowledgeBaseSummary[]
> & { refresh: () => Promise<KnowledgeBaseSummary[] | null> } {
  const client = useFluxbaseClient();
  const state: AsyncState<KnowledgeBaseSummary[]> = {
    data: ref(null),
    error: ref(null),
    loading: ref(false),
  };
  const refresh = () => runInto(state, () => client.knowledgeBase.list());
  void refresh();
  return { ...state, refresh };
}

/**
 * Search a knowledge base. Does not auto-run; call `run(kbId, request)`.
 */
export function useFluxbaseKBSearch(): AsyncState<SearchKnowledgeBaseResponse> & {
  run: (
    kbId: string,
    request: SearchKnowledgeBaseRequest,
  ) => Promise<SearchKnowledgeBaseResponse | null>;
} {
  const client = useFluxbaseClient();
  const state: AsyncState<SearchKnowledgeBaseResponse> = {
    data: ref(null),
    error: ref(null),
    loading: ref(false),
  };
  const run = (kbId: string, request: SearchKnowledgeBaseRequest) =>
    runInto(state, () => client.knowledgeBase.search(kbId, request));
  return { ...state, run };
}

/**
 * Vector similarity search with automatic query embedding.
 * Does not auto-run; call `run(options)`.
 */
export function useFluxbaseVectorSearch(): AsyncState<
  VectorSearchResult<Record<string, unknown>>
> & {
  run: (
    options: VectorSearchOptions,
  ) => Promise<VectorSearchResult<Record<string, unknown>> | null>;
} {
  const client = useFluxbaseClient();
  const state: AsyncState<VectorSearchResult<Record<string, unknown>>> = {
    data: ref(null),
    error: ref(null),
    loading: ref(false),
  };
  const run = (options: VectorSearchOptions) =>
    runInto(state, () => client.vector.search(options));
  return { ...state, run };
}

/**
 * Vector search capabilities (pgvector availability, embedding model).
 */
export function useFluxbaseVectorCapabilities(): AsyncState<VectorCapabilities> & {
  refresh: () => Promise<VectorCapabilities | null>;
} {
  const client = useFluxbaseClient();
  const state: AsyncState<VectorCapabilities> = {
    data: ref(null),
    error: ref(null),
    loading: ref(false),
  };
  const refresh = () => runInto(state, () => client.vector.getCapabilities());
  void refresh();
  return { ...state, refresh };
}
