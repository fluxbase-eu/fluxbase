/**
 * Admin AI module for managing AI chatbots and providers
 * Provides administrative operations for chatbot lifecycle management
 */

import type { FluxbaseFetch } from "./fetch";
import type {
  AIMetrics,
  AdminConversationListResponse,
  AdminConversationQuery,
  AdminMessagesResponse,
  AIChatbot,
  AIChatbotSummary,
  AIProvider,
  AuditLogQuery,
  AuditLogResponse,
  ChatbotKnowledgeBaseLink,
  CreateSyncConfigRequest,
  CreateToolIntegrationRequest,
  DocumentPermissionGrant,
  ExportTableToKBRequest,
  ExportTableToKBResponse,
  ExportableTable,
  ExportableTablesResponse,
  FluxbaseError,
  GrantDocumentPermissionRequest,
  IntegrationType,
  LinkKnowledgeBaseRequest,
  SyncChatbotsOptions,
  SyncChatbotsResult,
  TableDetails,
  TableExportSyncConfig,
  TestToolIntegrationResult,
  ToolIntegration,
  UpdateChatbotKnowledgeBaseRequest,
  UpdateSyncConfigRequest,
  UpdateToolIntegrationRequest,
} from "./types";

/**
 * Admin AI manager for managing AI chatbots and providers
 * Provides create, update, delete, sync, and monitoring operations
 *
 * @category Admin
 */
export class FluxbaseAdminAI {
  private fetch: FluxbaseFetch;

  constructor(fetch: FluxbaseFetch) {
    this.fetch = fetch;
  }

  // ============================================================================
  // CHATBOT MANAGEMENT
  // ============================================================================

  /**
   * List all chatbots (admin view)
   *
   * @param namespace - Optional namespace filter
   * @returns Promise resolving to { data, error } tuple with array of chatbot summaries
   *
   * @example
   * ```typescript
   * const { data, error } = await client.admin.ai.listChatbots()
   * if (data) {
   *   console.log('Chatbots:', data.map(c => c.name))
   * }
   * ```
   */
  async listChatbots(
    namespace?: string,
  ): Promise<{ data: AIChatbotSummary[] | null; error: FluxbaseError | null }> {
    try {
      const params = namespace ? `?namespace=${namespace}` : "";
      const response = await this.fetch.get<{
        chatbots: AIChatbotSummary[];
        count: number;
      }>(`/api/v1/admin/ai/chatbots${params}`);
      return { data: response.chatbots || [], error: null };
    } catch (error) {
      return { data: null, error: error as FluxbaseError };
    }
  }

  /**
   * Get details of a specific chatbot
   *
   * @param id - Chatbot ID
   * @returns Promise resolving to { data, error } tuple with chatbot details
   *
   * @example
   * ```typescript
   * const { data, error } = await client.admin.ai.getChatbot('uuid')
   * if (data) {
   *   console.log('Chatbot:', data.name)
   * }
   * ```
   */
  async getChatbot(
    id: string,
  ): Promise<{ data: AIChatbot | null; error: FluxbaseError | null }> {
    try {
      const data = await this.fetch.get<AIChatbot>(
        `/api/v1/admin/ai/chatbots/${id}`,
      );
      return { data, error: null };
    } catch (error) {
      return { data: null, error: error as FluxbaseError };
    }
  }

  /**
   * Enable or disable a chatbot
   *
   * @param id - Chatbot ID
   * @param enabled - Whether to enable or disable
   * @returns Promise resolving to { data, error } tuple with updated chatbot
   *
   * @example
   * ```typescript
   * const { data, error } = await client.admin.ai.toggleChatbot('uuid', true)
   * ```
   */
  async toggleChatbot(
    id: string,
    enabled: boolean,
  ): Promise<{ data: AIChatbot | null; error: FluxbaseError | null }> {
    try {
      const data = await this.fetch.post<AIChatbot>(
        `/api/v1/admin/ai/chatbots/${id}/toggle`,
        { enabled },
      );
      return { data, error: null };
    } catch (error) {
      return { data: null, error: error as FluxbaseError };
    }
  }

  /**
   * Update a chatbot's configuration (partial update).
   *
   * Only the provided fields are changed; omitted fields are left untouched.
   * For limit/budget fields, 0 means "unlimited". Changes take effect on the
   * next request (limits are read fresh from the database per message).
   *
   * @param id - Chatbot ID
   * @param updates - Fields to update (all optional)
   * @returns Promise resolving to { data, error } tuple with the updated chatbot
   *
   * @example
   * ```typescript
   * const { data, error } = await client.admin.ai.updateChatbot('uuid', {
   *   daily_request_limit: 500,   // 0 = unlimited
   *   daily_token_budget: 200000, // 0 = unlimited
   * })
   * ```
   */
  async updateChatbot(
    id: string,
    updates: {
      description?: string;
      enabled?: boolean;
      max_tokens?: number;
      temperature?: number;
      provider_id?: string | null;
      persist_conversations?: boolean;
      conversation_ttl_hours?: number;
      max_conversation_turns?: number;
      rate_limit_per_minute?: number;
      daily_request_limit?: number;
      daily_token_budget?: number;
      allow_unauthenticated?: boolean;
      is_public?: boolean;
    },
  ): Promise<{ data: AIChatbot | null; error: FluxbaseError | null }> {
    try {
      const data = await this.fetch.put<AIChatbot>(
        `/api/v1/admin/ai/chatbots/${id}`,
        updates,
      );
      return { data, error: null };
    } catch (error) {
      return { data: null, error: error as FluxbaseError };
    }
  }

  /**
   * Delete a chatbot
   *
   * @param id - Chatbot ID
   * @returns Promise resolving to { data, error } tuple
   *
   * @example
   * ```typescript
   * const { data, error } = await client.admin.ai.deleteChatbot('uuid')
   * ```
   */
  async deleteChatbot(
    id: string,
  ): Promise<{ data: null; error: FluxbaseError | null }> {
    try {
      await this.fetch.delete(`/api/v1/admin/ai/chatbots/${id}`);
      return { data: null, error: null };
    } catch (error) {
      return { data: null, error: error as FluxbaseError };
    }
  }

  /**
   * Sync chatbots from filesystem or API payload
   *
   * Can sync from:
   * 1. Filesystem (if no chatbots provided) - loads from configured chatbots directory
   * 2. API payload (if chatbots array provided) - syncs provided chatbot specifications
   *
   * Requires service_role or admin authentication.
   *
   * @param options - Sync options including namespace and optional chatbots array
   * @returns Promise resolving to { data, error } tuple with sync results
   *
   * @example
   * ```typescript
   * // Sync from filesystem
   * const { data, error } = await client.admin.ai.sync()
   *
   * // Sync with provided chatbot code
   * const { data, error } = await client.admin.ai.sync({
   *   namespace: 'default',
   *   chatbots: [{
   *     name: 'sql-assistant',
   *     code: myChatbotCode,
   *   }],
   *   options: {
   *     delete_missing: false, // Don't remove chatbots not in this sync
   *     dry_run: false,        // Preview changes without applying
   *   }
   * })
   *
   * if (data) {
   *   console.log(`Synced: ${data.summary.created} created, ${data.summary.updated} updated`)
   * }
   * ```
   */
  async sync(
    options?: SyncChatbotsOptions,
  ): Promise<{ data: SyncChatbotsResult | null; error: FluxbaseError | null }> {
    try {
      const data = await this.fetch.post<SyncChatbotsResult>(
        "/api/v1/admin/ai/chatbots/sync",
        {
          namespace: options?.namespace || "default",
          chatbots: options?.chatbots,
          options: {
            delete_missing: options?.options?.delete_missing ?? false,
            dry_run: options?.options?.dry_run ?? false,
          },
        },
      );
      return { data, error: null };
    } catch (error) {
      return { data: null, error: error as FluxbaseError };
    }
  }

  // ============================================================================
  // PROVIDER MANAGEMENT
  // ============================================================================

  /**
   * List all AI providers
   *
   * @param options - Optional request options, including an AbortSignal so
   *   callers (e.g. TanStack Query) can cancel the request
   * @returns Promise resolving to { data, error } tuple with array of providers
   *
   * @example
   * ```typescript
   * const { data, error } = await client.admin.ai.listProviders()
   * if (data) {
   *   console.log('Providers:', data.map(p => p.name))
   * }
   * ```
   */
  async listProviders(options: {
    signal?: AbortSignal | null;
  } = {}): Promise<{
    data: AIProvider[] | null;
    error: FluxbaseError | null;
  }> {
    try {
      const response = await this.fetch.get<{
        providers: AIProvider[];
        count: number;
      }>("/api/v1/admin/ai/providers", { signal: options.signal });
      return { data: response.providers || [], error: null };
    } catch (error) {
      return { data: null, error: error as FluxbaseError };
    }
  }

  /**
   * Get details of a specific AI provider
   *
   * @param id - Provider ID
   * @returns Promise resolving to { data, error } tuple with provider details
   */
  async getProvider(
    id: string,
  ): Promise<{ data: AIProvider | null; error: FluxbaseError | null }> {
    try {
      const data = await this.fetch.get<AIProvider>(
        `/api/v1/admin/ai/providers/${id}`,
      );
      return { data, error: null };
    } catch (error) {
      return { data: null, error: error as FluxbaseError };
    }
  }

  /**
   * Create a new AI provider
   *
   * @param params - Provider configuration including name, provider_type, and optional config
   * @returns Promise resolving to { data, error } tuple with created provider
   */
  async createProvider(params: {
    name: string;
    display_name?: string;
    provider_type: string;
    is_default?: boolean;
    config?: Record<string, unknown>;
  }): Promise<{ data: AIProvider | null; error: FluxbaseError | null }> {
    try {
      const body: Record<string, unknown> = {
        name: params.name,
        provider_type: params.provider_type,
      };
      if (params.display_name !== undefined)
        body.display_name = params.display_name;
      if (params.is_default !== undefined) body.is_default = params.is_default;
      if (params.config) {
        body.config = normalizeConfig(params.config);
      }
      const data = await this.fetch.post<AIProvider>(
        "/api/v1/admin/ai/providers",
        body,
      );
      return { data, error: null };
    } catch (error) {
      return { data: null, error: error as FluxbaseError };
    }
  }

  /**
   * Update an existing AI provider
   *
   * @param id - Provider ID
   * @param updates - Fields to update
   * @returns Promise resolving to { data, error } tuple with updated provider
   */
  async updateProvider(
    id: string,
    updates: {
      display_name?: string;
      enabled?: boolean;
      config?: Record<string, unknown>;
      embedding_model?: string | null;
    },
  ): Promise<{ data: AIProvider | null; error: FluxbaseError | null }> {
    try {
      const body: Record<string, unknown> = { ...updates };
      if (updates.config) {
        body.config = normalizeConfig(updates.config);
      }
      const data = await this.fetch.put<AIProvider>(
        `/api/v1/admin/ai/providers/${id}`,
        body,
      );
      return { data, error: null };
    } catch (error) {
      return { data: null, error: error as FluxbaseError };
    }
  }

  /**
   * Set a provider as the default provider
   *
   * @param id - Provider ID
   * @returns Promise resolving to { data, error } tuple with updated provider
   */
  async setDefaultProvider(
    id: string,
  ): Promise<{ data: AIProvider | null; error: FluxbaseError | null }> {
    try {
      const data = await this.fetch.put<AIProvider>(
        `/api/v1/admin/ai/providers/${id}/default`,
        {},
      );
      return { data, error: null };
    } catch (error) {
      return { data: null, error: error as FluxbaseError };
    }
  }

  /**
   * Delete an AI provider
   *
   * @param id - Provider ID
   * @returns Promise resolving to { data, error } tuple
   */
  async deleteProvider(
    id: string,
  ): Promise<{ data: null; error: FluxbaseError | null }> {
    try {
      await this.fetch.delete(`/api/v1/admin/ai/providers/${id}`);
      return { data: null, error: null };
    } catch (error) {
      return { data: null, error: error as FluxbaseError };
    }
  }

  /**
   * Set a provider as the embedding provider
   *
   * @param id - Provider ID
   * @returns Promise resolving to { data, error } tuple with updated provider
   */
  async setEmbeddingProvider(
    id: string,
  ): Promise<{
    data: { use_for_embeddings: boolean } | null;
    error: FluxbaseError | null;
  }> {
    try {
      const data = await this.fetch.put<{ use_for_embeddings: boolean }>(
        `/api/v1/admin/ai/providers/${id}/embedding`,
        {},
      );
      return { data, error: null };
    } catch (error) {
      return { data: null, error: error as FluxbaseError };
    }
  }

  /**
   * Clear the embedding provider assignment for a provider
   *
   * @param id - Provider ID
   * @returns Promise resolving to { data, error } tuple
   */
  async clearEmbeddingProvider(
    id: string,
  ): Promise<{
    data: { use_for_embeddings: boolean } | null;
    error: FluxbaseError | null;
  }> {
    try {
      const data = await this.fetch.delete<{ use_for_embeddings: boolean }>(
        `/api/v1/admin/ai/providers/${id}/embedding`,
      );
      return { data, error: null };
    } catch (error) {
      return { data: null, error: error as FluxbaseError };
    }
  }

  // ============================================================================
  // CHATBOT KNOWLEDGE BASE LINKING
  // ============================================================================

  /**
   * List knowledge bases linked to a chatbot
   *
   * @param chatbotId - Chatbot ID
   * @returns Promise resolving to { data, error } tuple with linked knowledge bases
   *
   * @example
   * ```typescript
   * const { data, error } = await client.admin.ai.listChatbotKnowledgeBases('chatbot-uuid')
   * if (data) {
   *   console.log('Linked KBs:', data.map(l => l.knowledge_base_id))
   * }
   * ```
   */
  async listChatbotKnowledgeBases(
    chatbotId: string,
  ): Promise<{ data: ChatbotKnowledgeBaseLink[] | null; error: FluxbaseError | null }> {
    try {
      const response = await this.fetch.get<{
        knowledge_bases: ChatbotKnowledgeBaseLink[];
        count: number;
      }>(`/api/v1/admin/ai/chatbots/${chatbotId}/knowledge-bases`);
      return { data: response.knowledge_bases || [], error: null };
    } catch (error) {
      return { data: null, error: error as FluxbaseError };
    }
  }

  /**
   * Link a knowledge base to a chatbot
   *
   * @param chatbotId - Chatbot ID
   * @param request - Link configuration
   * @returns Promise resolving to { data, error } tuple with link details
   *
   * @example
   * ```typescript
   * const { data, error } = await client.admin.ai.linkKnowledgeBase('chatbot-uuid', {
   *   knowledge_base_id: 'kb-uuid',
   *   priority: 1,
   *   max_chunks: 5,
   *   similarity_threshold: 0.7,
   * })
   * ```
   */
  async linkKnowledgeBase(
    chatbotId: string,
    request: LinkKnowledgeBaseRequest,
  ): Promise<{ data: ChatbotKnowledgeBaseLink | null; error: FluxbaseError | null }> {
    try {
      const data = await this.fetch.post<ChatbotKnowledgeBaseLink>(
        `/api/v1/admin/ai/chatbots/${chatbotId}/knowledge-bases`,
        request,
      );
      return { data, error: null };
    } catch (error) {
      return { data: null, error: error as FluxbaseError };
    }
  }

  /**
   * Update a chatbot-knowledge base link
   *
   * @param chatbotId - Chatbot ID
   * @param knowledgeBaseId - Knowledge base ID
   * @param updates - Fields to update
   * @returns Promise resolving to { data, error } tuple with updated link
   *
   * @example
   * ```typescript
   * const { data, error } = await client.admin.ai.updateChatbotKnowledgeBase(
   *   'chatbot-uuid',
   *   'kb-uuid',
   *   { max_chunks: 10, enabled: true }
   * )
   * ```
   */
  async updateChatbotKnowledgeBase(
    chatbotId: string,
    knowledgeBaseId: string,
    updates: UpdateChatbotKnowledgeBaseRequest,
  ): Promise<{ data: ChatbotKnowledgeBaseLink | null; error: FluxbaseError | null }> {
    try {
      const data = await this.fetch.put<ChatbotKnowledgeBaseLink>(
        `/api/v1/admin/ai/chatbots/${chatbotId}/knowledge-bases/${knowledgeBaseId}`,
        updates,
      );
      return { data, error: null };
    } catch (error) {
      return { data: null, error: error as FluxbaseError };
    }
  }

  /**
   * Unlink a knowledge base from a chatbot
   *
   * @param chatbotId - Chatbot ID
   * @param knowledgeBaseId - Knowledge base ID
   * @returns Promise resolving to { data, error } tuple
   *
   * @example
   * ```typescript
   * const { data, error } = await client.admin.ai.unlinkKnowledgeBase('chatbot-uuid', 'kb-uuid')
   * ```
   */
  async unlinkKnowledgeBase(
    chatbotId: string,
    knowledgeBaseId: string,
  ): Promise<{ data: null; error: FluxbaseError | null }> {
    try {
      await this.fetch.delete(
        `/api/v1/admin/ai/chatbots/${chatbotId}/knowledge-bases/${knowledgeBaseId}`,
      );
      return { data: null, error: null };
    } catch (error) {
      return { data: null, error: error as FluxbaseError };
    }
  }

  // ============================================================================
  // TABLE DETAILS
  // ============================================================================

  /**
   * Get detailed table information including columns
   *
   * Use this to discover available columns before exporting.
   *
   * @param schema - Schema name (e.g., 'public')
   * @param table - Table name
   * @returns Promise resolving to { data, error } tuple with table details
   *
   * @example
   * ```typescript
   * const { data, error } = await client.admin.ai.getTableDetails('public', 'users')
   * if (data) {
   *   console.log('Columns:', data.columns.map(c => c.name))
   *   console.log('Primary key:', data.primary_key)
   * }
   * ```
   */
  async getTableDetails(
    schema: string,
    table: string,
  ): Promise<{ data: TableDetails | null; error: FluxbaseError | null }> {
    try {
      const data = await this.fetch.get<TableDetails>(
        `/api/v1/admin/ai/tables/${schema}/${table}`,
      );
      return { data, error: null };
    } catch (error) {
      return { data: null, error: error as FluxbaseError };
    }
  }

  // ==========================================================================
  // Tool Integrations (Web Search via Tavily, future Brave/Jina, URL fetch)
  // ==========================================================================
  // Mirrors the provider CRUD shape plus a /test endpoint that runs a real
  // "hello world" call against the integration's provider and stores the
  // result. Used by the admin UI's Test button.

  /**
   * List tool integrations, optionally filtered by integration_type.
   *
   * @example
   * const { data, error } = await client.admin.ai.listIntegrations({ integration_type: "web_search" })
   */
  async listIntegrations(
    params?: { integration_type?: IntegrationType; signal?: AbortSignal | null },
  ): Promise<{
    data: ToolIntegration[] | null;
    error: FluxbaseError | null;
  }> {
    try {
      const query = params?.integration_type
        ? `?type=${params.integration_type}`
        : "";
      const result = await this.fetch.get<{
        integrations: ToolIntegration[];
        count: number;
      }>(`/api/v1/admin/ai/integrations${query}`, {
        signal: params?.signal,
      });
      return { data: result.integrations || [], error: null };
    } catch (error) {
      return { data: null, error: error as FluxbaseError };
    }
  }

  /**
   * Get a single tool integration by ID.
   */
  async getIntegration(
    id: string,
  ): Promise<{ data: ToolIntegration | null; error: FluxbaseError | null }> {
    try {
      const data = await this.fetch.get<ToolIntegration>(
        `/api/v1/admin/ai/integrations/${id}`,
      );
      return { data, error: null };
    } catch (error) {
      return { data: null, error: error as FluxbaseError };
    }
  }

  /**
   * Create a new tool integration. api_key in config is encrypted at the
   * storage layer; never appears in plaintext in API responses.
   */
  async createIntegration(
    request: CreateToolIntegrationRequest,
  ): Promise<{ data: ToolIntegration | null; error: FluxbaseError | null }> {
    try {
      const data = await this.fetch.post<ToolIntegration>(
        "/api/v1/admin/ai/integrations",
        request,
      );
      return { data, error: null };
    } catch (error) {
      return { data: null, error: error as FluxbaseError };
    }
  }

  /**
   * Update an existing tool integration. Passing config.api_key =
   * "***masked***" preserves the existing encrypted value (useful when
   * the admin changes only the name and not the API key).
   */
  async updateIntegration(
    id: string,
    request: UpdateToolIntegrationRequest,
  ): Promise<{ data: ToolIntegration | null; error: FluxbaseError | null }> {
    try {
      const data = await this.fetch.put<ToolIntegration>(
        `/api/v1/admin/ai/integrations/${id}`,
        request,
      );
      return { data, error: null };
    } catch (error) {
      return { data: null, error: error as FluxbaseError };
    }
  }

  /**
   * Delete a tool integration. Refuses to delete read-only integrations
   * configured via env/YAML.
   */
  async deleteIntegration(
    id: string,
  ): Promise<{ error: FluxbaseError | null }> {
    try {
      await this.fetch.delete(`/api/v1/admin/ai/integrations/${id}`);
      return { error: null };
    } catch (error) {
      return { error: error as FluxbaseError };
    }
  }

  /**
   * Mark an integration as the default for its integration_type within
   * the current tenant. The server clears any prior default in the same
   * transaction.
   */
  async setDefaultIntegration(
    id: string,
  ): Promise<{ error: FluxbaseError | null }> {
    try {
      await this.fetch.put(
        `/api/v1/admin/ai/integrations/${id}/default`,
        {},
      );
      return { error: null };
    } catch (error) {
      return { error: error as FluxbaseError };
    }
  }

  /**
   * Test an integration by running a real "hello world" call against
   * its provider. Stores last_tested_at + last_test_status + error
   * on the integration row so the admin UI can display health.
   *
   * @example
   * const { data, error } = await client.admin.ai.testIntegration(integrationId)
   * if (data?.status === "ok") { console.log("Tavily credentials verified") }
   */
  async testIntegration(
    id: string,
  ): Promise<{
    data: TestToolIntegrationResult | null;
    error: FluxbaseError | null;
  }> {
    try {
      const data = await this.fetch.post<TestToolIntegrationResult>(
        `/api/v1/admin/ai/integrations/${id}/test`,
        {},
      );
      return { data, error: null };
    } catch (error) {
      return { data: null, error: error as FluxbaseError };
    }
  }

  // ===========================================================================
  // METRICS
  // ===========================================================================

  /**
   * Get aggregated AI metrics (requests, tokens, per-chatbot/provider stats)
   *
   * @example
   * ```typescript
   * const { data } = await client.admin.ai.getMetrics()
   * ```
   */
  async getMetrics(): Promise<{ data: AIMetrics | null; error: FluxbaseError | null }> {
    try {
      const data = await this.fetch.get<AIMetrics>("/api/v1/admin/ai/metrics");
      return { data, error: null };
    } catch (error) {
      return { data: null, error: error as FluxbaseError };
    }
  }

  // ===========================================================================
  // ADMIN CONVERSATIONS & AUDIT
  // ===========================================================================

  /**
   * List AI conversations across all chatbots (admin)
   */
  async listConversations(
    query: AdminConversationQuery = {},
  ): Promise<{ data: AdminConversationListResponse | null; error: FluxbaseError | null }> {
    try {
      const params = new URLSearchParams();
      if (query.chatbot_id) params.set("chatbot_id", query.chatbot_id);
      if (query.user_id) params.set("user_id", query.user_id);
      if (query.status) params.set("status", query.status);
      if (query.limit !== undefined) params.set("limit", String(query.limit));
      if (query.offset !== undefined) params.set("offset", String(query.offset));
      const qs = params.toString();
      const data = await this.fetch.get<AdminConversationListResponse>(
        `/api/v1/admin/ai/conversations${qs ? `?${qs}` : ""}`,
      );
      return { data, error: null };
    } catch (error) {
      return { data: null, error: error as FluxbaseError };
    }
  }

  /**
   * Get all messages of a conversation (admin)
   */
  async getConversationMessages(
    conversationId: string,
    query: { limit?: number; offset?: number } = {},
  ): Promise<{ data: AdminMessagesResponse | null; error: FluxbaseError | null }> {
    try {
      const params = new URLSearchParams();
      if (query.limit !== undefined) params.set("limit", String(query.limit));
      if (query.offset !== undefined) params.set("offset", String(query.offset));
      const qs = params.toString();
      const data = await this.fetch.get<AdminMessagesResponse>(
        `/api/v1/admin/ai/conversations/${conversationId}/messages${qs ? `?${qs}` : ""}`,
      );
      return { data, error: null };
    } catch (error) {
      return { data: null, error: error as FluxbaseError };
    }
  }

  /**
   * Get the AI SQL audit log (admin)
   */
  async getAuditLog(
    query: AuditLogQuery = {},
  ): Promise<{ data: AuditLogResponse | null; error: FluxbaseError | null }> {
    try {
      const params = new URLSearchParams();
      if (query.chatbot_id) params.set("chatbot_id", query.chatbot_id);
      if (query.user_id) params.set("user_id", query.user_id);
      if (query.success !== undefined) params.set("success", String(query.success));
      if (query.limit !== undefined) params.set("limit", String(query.limit));
      if (query.offset !== undefined) params.set("offset", String(query.offset));
      const qs = params.toString();
      const data = await this.fetch.get<AuditLogResponse>(
        `/api/v1/admin/ai/audit${qs ? `?${qs}` : ""}`,
      );
      return { data, error: null };
    } catch (error) {
      return { data: null, error: error as FluxbaseError };
    }
  }

  // ===========================================================================
  // TABLE EXPORT
  // ===========================================================================

  /**
   * List tables available for knowledge base export
   */
  async listTables(): Promise<{ data: ExportableTable[] | null; error: FluxbaseError | null }> {
    try {
      const data = await this.fetch.get<ExportableTablesResponse>("/api/v1/admin/ai/tables");
      return { data: data.tables, error: null };
    } catch (error) {
      return { data: null, error: error as FluxbaseError };
    }
  }

  /**
   * Export a database table into a knowledge base as a document
   */
  async exportTableToKB(
    schema: string,
    table: string,
    request: ExportTableToKBRequest,
  ): Promise<{ data: ExportTableToKBResponse | null; error: FluxbaseError | null }> {
    try {
      const data = await this.fetch.post<ExportTableToKBResponse>(
        `/api/v1/admin/ai/tables/${schema}/${table}/export`,
        request,
      );
      return { data, error: null };
    } catch (error) {
      return { data: null, error: error as FluxbaseError };
    }
  }

  // ===========================================================================
  // TABLE EXPORT SYNC CONFIGS
  // ===========================================================================

  /**
   * Create a table export sync configuration for a knowledge base
   */
  async createSyncConfig(
    kbId: string,
    request: CreateSyncConfigRequest,
  ): Promise<{ data: TableExportSyncConfig | null; error: FluxbaseError | null }> {
    try {
      const data = await this.fetch.post<TableExportSyncConfig>(
        `/api/v1/admin/ai/knowledge-bases/${kbId}/sync-configs`,
        request,
      );
      return { data, error: null };
    } catch (error) {
      return { data: null, error: error as FluxbaseError };
    }
  }

  /**
   * List table export sync configurations for a knowledge base
   */
  async listSyncConfigs(
    kbId: string,
  ): Promise<{ data: TableExportSyncConfig[] | null; error: FluxbaseError | null }> {
    try {
      const data = await this.fetch.get<TableExportSyncConfig[]>(
        `/api/v1/admin/ai/knowledge-bases/${kbId}/sync-configs`,
      );
      return { data, error: null };
    } catch (error) {
      return { data: null, error: error as FluxbaseError };
    }
  }

  /**
   * Update a table export sync configuration
   */
  async updateSyncConfig(
    kbId: string,
    syncId: string,
    request: UpdateSyncConfigRequest,
  ): Promise<{ data: TableExportSyncConfig | null; error: FluxbaseError | null }> {
    try {
      const data = await this.fetch.patch<TableExportSyncConfig>(
        `/api/v1/admin/ai/knowledge-bases/${kbId}/sync-configs/${syncId}`,
        request,
      );
      return { data, error: null };
    } catch (error) {
      return { data: null, error: error as FluxbaseError };
    }
  }

  /**
   * Delete a table export sync configuration
   */
  async deleteSyncConfig(
    kbId: string,
    syncId: string,
  ): Promise<{ data: null; error: FluxbaseError | null }> {
    try {
      await this.fetch.delete(`/api/v1/admin/ai/knowledge-bases/${kbId}/sync-configs/${syncId}`);
      return { data: null, error: null };
    } catch (error) {
      return { data: null, error: error as FluxbaseError };
    }
  }

  /**
   * Trigger a table export sync immediately
   */
  async triggerSync(
    kbId: string,
    syncId: string,
  ): Promise<{ data: ExportTableToKBResponse | null; error: FluxbaseError | null }> {
    try {
      const data = await this.fetch.post<ExportTableToKBResponse>(
        `/api/v1/admin/ai/knowledge-bases/${kbId}/sync-configs/${syncId}/trigger`,
        {},
      );
      return { data, error: null };
    } catch (error) {
      return { data: null, error: error as FluxbaseError };
    }
  }

  // ===========================================================================
  // DOCUMENT PERMISSIONS
  // ===========================================================================

  /**
   * Grant a user permission on a document (admin)
   */
  async grantDocumentPermission(
    kbId: string,
    documentId: string,
    request: GrantDocumentPermissionRequest,
  ): Promise<{ data: DocumentPermissionGrant | null; error: FluxbaseError | null }> {
    try {
      const data = await this.fetch.post<DocumentPermissionGrant>(
        `/api/v1/admin/ai/knowledge-bases/${kbId}/documents/${documentId}/permissions`,
        request,
      );
      return { data, error: null };
    } catch (error) {
      return { data: null, error: error as FluxbaseError };
    }
  }

  /**
   * List permissions granted on a document (admin)
   */
  async listDocumentPermissions(
    kbId: string,
    documentId: string,
  ): Promise<{ data: DocumentPermissionGrant[] | null; error: FluxbaseError | null }> {
    try {
      const data = await this.fetch.get<DocumentPermissionGrant[]>(
        `/api/v1/admin/ai/knowledge-bases/${kbId}/documents/${documentId}/permissions`,
      );
      return { data, error: null };
    } catch (error) {
      return { data: null, error: error as FluxbaseError };
    }
  }

  /**
   * Revoke a document permission (admin)
   */
  async revokeDocumentPermission(
    kbId: string,
    documentId: string,
    userId: string,
  ): Promise<{ data: null; error: FluxbaseError | null }> {
    try {
      await this.fetch.delete(
        `/api/v1/admin/ai/knowledge-bases/${kbId}/documents/${documentId}/permissions/${userId}`,
      );
      return { data: null, error: null };
    } catch (error) {
      return { data: null, error: error as FluxbaseError };
    }
  }
}

/**
 * Normalize config values: convert non-strings to strings, drop null/undefined entries.
 */
function normalizeConfig(
  config: Record<string, unknown>,
): Record<string, string> {
  const result: Record<string, string> = {};
  for (const [key, value] of Object.entries(config)) {
    if (value === undefined || value === null) continue;
    result[key] = String(value);
  }
  return result;
}
