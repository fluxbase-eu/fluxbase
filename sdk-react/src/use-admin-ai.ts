/**
 * Admin AI hooks for Fluxbase React SDK
 *
 * Covers the AI operations surfaces: metrics, conversations/audit,
 * table export + sync configs, document permissions, and list queries
 * for chatbots/providers/integrations. CRUD mutations for providers and
 * integrations are intentionally not wrapped — drive them through the
 * core SDK's client.admin.ai directly until a consumer needs hooks.
 */

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useFluxbaseClient } from "./context";
import type {
  AdminConversationQuery,
  AuditLogQuery,
  CreateSyncConfigRequest,
  ExportTableToKBRequest,
  GrantDocumentPermissionRequest,
  UpdateSyncConfigRequest,
} from "@nimbleflux/fluxbase-sdk";

/**
 * Hook to get aggregated AI metrics
 */
export function useAdminAIMetrics() {
  const client = useFluxbaseClient();

  return useQuery({
    queryKey: ["fluxbase", "admin", "ai", "metrics"],
    queryFn: async () => {
      const { data, error } = await client.admin.ai.getMetrics();
      if (error) throw error;
      return data;
    },
  });
}

/**
 * Hook to list AI conversations across chatbots (admin)
 */
export function useAdminAIConversations(query: AdminConversationQuery = {}) {
  const client = useFluxbaseClient();

  return useQuery({
    queryKey: ["fluxbase", "admin", "ai", "conversations", query],
    queryFn: async () => {
      const { data, error } = await client.admin.ai.listConversations(query);
      if (error) throw error;
      return data;
    },
  });
}

/**
 * Hook to get all messages of a conversation (admin)
 */
export function useAdminAIConversationMessages(
  conversationId: string | null,
  query: { limit?: number; offset?: number } = {},
) {
  const client = useFluxbaseClient();

  return useQuery({
    queryKey: ["fluxbase", "admin", "ai", "conversation-messages", conversationId, query],
    queryFn: async () => {
      if (!conversationId) return null;
      const { data, error } = await client.admin.ai.getConversationMessages(conversationId, query);
      if (error) throw error;
      return data;
    },
    enabled: !!conversationId,
  });
}

/**
 * Hook to get the AI SQL audit log (admin)
 */
export function useAdminAIAuditLog(query: AuditLogQuery = {}) {
  const client = useFluxbaseClient();

  return useQuery({
    queryKey: ["fluxbase", "admin", "ai", "audit", query],
    queryFn: async () => {
      const { data, error } = await client.admin.ai.getAuditLog(query);
      if (error) throw error;
      return data;
    },
  });
}

/**
 * Hook to list tables available for knowledge base export (admin)
 */
export function useAdminExportableTables() {
  const client = useFluxbaseClient();

  return useQuery({
    queryKey: ["fluxbase", "admin", "ai", "tables"],
    queryFn: async () => {
      const { data, error } = await client.admin.ai.listTables();
      if (error) throw error;
      return data || [];
    },
  });
}

/**
 * Hook to export a table into a knowledge base (admin)
 */
export function useAdminExportTableToKB() {
  const client = useFluxbaseClient();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      schema,
      table,
      request,
    }: {
      schema: string;
      table: string;
      request: ExportTableToKBRequest;
    }) => {
      const { data, error } = await client.admin.ai.exportTableToKB(schema, table, request);
      if (error) throw error;
      return data;
    },
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({
        queryKey: ["fluxbase", "admin", "ai", "kb-documents", variables.request.knowledge_base_id],
      });
    },
  });
}

/**
 * Hook to list table export sync configs for a knowledge base (admin)
 */
export function useAdminSyncConfigs(kbId: string | null) {
  const client = useFluxbaseClient();

  return useQuery({
    queryKey: ["fluxbase", "admin", "ai", "sync-configs", kbId],
    queryFn: async () => {
      if (!kbId) return [];
      const { data, error } = await client.admin.ai.listSyncConfigs(kbId);
      if (error) throw error;
      return data || [];
    },
    enabled: !!kbId,
  });
}

/**
 * Hook to create a table export sync config (admin)
 */
export function useAdminCreateSyncConfig() {
  const client = useFluxbaseClient();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      kbId,
      request,
    }: {
      kbId: string;
      request: CreateSyncConfigRequest;
    }) => {
      const { data, error } = await client.admin.ai.createSyncConfig(kbId, request);
      if (error) throw error;
      return data;
    },
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({
        queryKey: ["fluxbase", "admin", "ai", "sync-configs", variables.kbId],
      });
    },
  });
}

/**
 * Hook to update a table export sync config (admin)
 */
export function useAdminUpdateSyncConfig() {
  const client = useFluxbaseClient();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      kbId,
      syncId,
      request,
    }: {
      kbId: string;
      syncId: string;
      request: UpdateSyncConfigRequest;
    }) => {
      const { data, error } = await client.admin.ai.updateSyncConfig(kbId, syncId, request);
      if (error) throw error;
      return data;
    },
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({
        queryKey: ["fluxbase", "admin", "ai", "sync-configs", variables.kbId],
      });
    },
  });
}

/**
 * Hook to delete a table export sync config (admin)
 */
export function useAdminDeleteSyncConfig() {
  const client = useFluxbaseClient();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ kbId, syncId }: { kbId: string; syncId: string }) => {
      const { error } = await client.admin.ai.deleteSyncConfig(kbId, syncId);
      if (error) throw error;
    },
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({
        queryKey: ["fluxbase", "admin", "ai", "sync-configs", variables.kbId],
      });
    },
  });
}

/**
 * Hook to trigger a table export sync immediately (admin)
 */
export function useAdminTriggerSync() {
  const client = useFluxbaseClient();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ kbId, syncId }: { kbId: string; syncId: string }) => {
      const { data, error } = await client.admin.ai.triggerSync(kbId, syncId);
      if (error) throw error;
      return data;
    },
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({
        queryKey: ["fluxbase", "admin", "ai", "sync-configs", variables.kbId],
      });
    },
  });
}

/**
 * Hook to grant a user permission on a document (admin)
 */
export function useAdminGrantDocumentPermission() {
  const client = useFluxbaseClient();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      kbId,
      documentId,
      request,
    }: {
      kbId: string;
      documentId: string;
      request: GrantDocumentPermissionRequest;
    }) => {
      const { data, error } = await client.admin.ai.grantDocumentPermission(kbId, documentId, request);
      if (error) throw error;
      return data;
    },
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({
        queryKey: ["fluxbase", "admin", "ai", "document-permissions", variables.kbId, variables.documentId],
      });
    },
  });
}

/**
 * Hook to list document permissions (admin)
 */
export function useAdminDocumentPermissions(kbId: string | null, documentId: string | null) {
  const client = useFluxbaseClient();

  return useQuery({
    queryKey: ["fluxbase", "admin", "ai", "document-permissions", kbId, documentId],
    queryFn: async () => {
      if (!kbId || !documentId) return [];
      const { data, error } = await client.admin.ai.listDocumentPermissions(kbId, documentId);
      if (error) throw error;
      return data || [];
    },
    enabled: !!kbId && !!documentId,
  });
}

/**
 * Hook to revoke a document permission (admin)
 */
export function useAdminRevokeDocumentPermission() {
  const client = useFluxbaseClient();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      kbId,
      documentId,
      userId,
    }: {
      kbId: string;
      documentId: string;
      userId: string;
    }) => {
      const { error } = await client.admin.ai.revokeDocumentPermission(kbId, documentId, userId);
      if (error) throw error;
    },
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({
        queryKey: ["fluxbase", "admin", "ai", "document-permissions", variables.kbId, variables.documentId],
      });
    },
  });
}

/**
 * Hook to list AI chatbots as seen by admins (includes disabled)
 */
export function useAdminAIChatbots() {
  const client = useFluxbaseClient();

  return useQuery({
    queryKey: ["fluxbase", "admin", "ai", "chatbots"],
    queryFn: async () => {
      const { data, error } = await client.admin.ai.listChatbots();
      if (error) throw error;
      return data || [];
    },
  });
}

/**
 * Hook to list AI providers (admin)
 */
export function useAdminAIProviders() {
  const client = useFluxbaseClient();

  return useQuery({
    queryKey: ["fluxbase", "admin", "ai", "providers"],
    queryFn: async () => {
      const { data, error } = await client.admin.ai.listProviders();
      if (error) throw error;
      return data || [];
    },
  });
}

/**
 * Hook to list tool integrations (admin)
 */
export function useAdminAIIntegrations() {
  const client = useFluxbaseClient();

  return useQuery({
    queryKey: ["fluxbase", "admin", "ai", "integrations"],
    queryFn: async () => {
      const { data, error } = await client.admin.ai.listIntegrations();
      if (error) throw error;
      return data || [];
    },
  });
}
