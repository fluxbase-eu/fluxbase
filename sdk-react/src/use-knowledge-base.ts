/**
 * Knowledge Base hooks for Fluxbase React SDK
 * Provides hooks for RAG document management, semantic search, and knowledge graph
 */

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useFluxbaseClient } from "./context";
import type {
  CreateKnowledgeBaseRequest,
  UpdateKnowledgeBaseRequest,
  AddDocumentRequest,
  SearchKnowledgeBaseRequest,
  UpdateDocumentRequest,
  DeleteDocumentsByFilterRequest,
} from "@nimbleflux/fluxbase-sdk";

/**
 * Hook to list all knowledge bases the user has access to
 */
export function useKnowledgeBases() {
  const client = useFluxbaseClient();

  return useQuery({
    queryKey: ["fluxbase", "knowledge-base", "list"],
    queryFn: async () => {
      const { data, error } = await client.knowledgeBase.list();
      if (error) throw error;
      return data || [];
    },
  });
}

/**
 * Hook to get a single knowledge base by ID
 */
export function useKnowledgeBase(kbId: string | null) {
  const client = useFluxbaseClient();

  return useQuery({
    queryKey: ["fluxbase", "knowledge-base", kbId],
    queryFn: async () => {
      if (!kbId) return null;
      const { data, error } = await client.knowledgeBase.get(kbId);
      if (error) throw error;
      return data;
    },
    enabled: !!kbId,
  });
}

/**
 * Hook to create a new knowledge base
 */
export function useCreateKnowledgeBase() {
  const client = useFluxbaseClient();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (request: CreateKnowledgeBaseRequest) => {
      const { data, error } = await client.knowledgeBase.create(request);
      if (error) throw error;
      return data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["fluxbase", "knowledge-base", "list"],
      });
    },
  });
}

/**
 * Hook to update a knowledge base
 */
export function useUpdateKnowledgeBase() {
  const client = useFluxbaseClient();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      kbId,
      updates,
    }: {
      kbId: string;
      updates: UpdateKnowledgeBaseRequest;
    }) => {
      const { data, error } = await client.knowledgeBase.update(kbId, updates);
      if (error) throw error;
      return data;
    },
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({
        queryKey: ["fluxbase", "knowledge-base", "list"],
      });
      queryClient.invalidateQueries({
        queryKey: ["fluxbase", "knowledge-base", variables.kbId],
      });
    },
  });
}

/**
 * Hook to delete a knowledge base
 */
export function useDeleteKnowledgeBase() {
  const client = useFluxbaseClient();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (kbId: string) => {
      const { error } = await client.knowledgeBase.delete(kbId);
      if (error) throw error;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["fluxbase", "knowledge-base", "list"],
      });
    },
  });
}

/**
 * Hook to list documents in a knowledge base
 */
export function useKBDocuments(kbId: string | null) {
  const client = useFluxbaseClient();

  return useQuery({
    queryKey: ["fluxbase", "knowledge-base", kbId, "documents"],
    queryFn: async () => {
      if (!kbId) return [];
      const { data, error } = await client.knowledgeBase.listDocuments(kbId);
      if (error) throw error;
      return data || [];
    },
    enabled: !!kbId,
  });
}

/**
 * Hook to add a document to a knowledge base
 */
export function useAddDocument() {
  const client = useFluxbaseClient();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      kbId,
      request,
    }: {
      kbId: string;
      request: AddDocumentRequest;
    }) => {
      const { data, error } = await client.knowledgeBase.addDocument(
        kbId,
        request,
      );
      if (error) throw error;
      return data;
    },
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({
        queryKey: ["fluxbase", "knowledge-base", variables.kbId, "documents"],
      });
    },
  });
}

/**
 * Hook to upload a file document to a knowledge base
 */
export function useUploadDocument() {
  const client = useFluxbaseClient();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      kbId,
      file,
      filename,
      metadata,
    }: {
      kbId: string;
      file: File | Blob | ArrayBuffer;
      filename: string;
      metadata?: Record<string, string>;
    }) => {
      const { data, error } = await client.knowledgeBase.uploadDocument(
        kbId,
        file,
        filename,
        metadata,
      );
      if (error) throw error;
      return data;
    },
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({
        queryKey: ["fluxbase", "knowledge-base", variables.kbId, "documents"],
      });
    },
  });
}

/**
 * Hook to delete a document
 */
export function useDeleteDocument() {
  const client = useFluxbaseClient();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      kbId,
      docId,
    }: {
      kbId: string;
      docId: string;
    }) => {
      const { error } = await client.knowledgeBase.deleteDocument(kbId, docId);
      if (error) throw error;
    },
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({
        queryKey: ["fluxbase", "knowledge-base", variables.kbId, "documents"],
      });
    },
  });
}

/**
 * Hook to search a knowledge base with semantic similarity
 *
 * @example
 * ```tsx
 * const { data, isPending } = useKBSearch(kbId, {
 *   query: 'How to configure auth?',
 *   max_chunks: 5,
 * })
 * ```
 */
export function useKBSearch(
  kbId: string | null,
  request: SearchKnowledgeBaseRequest | null,
) {
  const client = useFluxbaseClient();

  return useQuery({
    queryKey: ["fluxbase", "knowledge-base", kbId, "search", request],
    queryFn: async () => {
      if (!kbId || !request) return null;
      const { data, error } = await client.knowledgeBase.search(kbId, request);
      if (error) throw error;
      return data;
    },
    enabled: !!kbId && !!request,
  });
}

/**
 * Hook to list entities in a knowledge base
 */
export function useKBEntities(kbId: string | null, type?: string) {
  const client = useFluxbaseClient();

  return useQuery({
    queryKey: ["fluxbase", "knowledge-base", kbId, "entities", type],
    queryFn: async () => {
      if (!kbId) return [];
      const { data, error } = await client.knowledgeBase.listEntities(
        kbId,
        type as any,
      );
      if (error) throw error;
      return data || [];
    },
    enabled: !!kbId,
  });
}

/**
 * Hook to get the knowledge graph (entities + relationships)
 */
export function useKnowledgeGraph(kbId: string | null) {
  const client = useFluxbaseClient();

  return useQuery({
    queryKey: ["fluxbase", "knowledge-base", kbId, "graph"],
    queryFn: async () => {
      if (!kbId) return null;
      const { data, error } = await client.knowledgeBase.getKnowledgeGraph(kbId);
      if (error) throw error;
      return data;
    },
    enabled: !!kbId,
  });
}

/**
 * Hook to search entities in a knowledge base by name
 */
export function useSearchEntities(kbId: string | null, query: string | null) {
  const client = useFluxbaseClient();

  return useQuery({
    queryKey: ["fluxbase", "knowledge-base", kbId, "entities-search", query],
    queryFn: async () => {
      if (!kbId || !query) return [];
      const { data, error } = await client.knowledgeBase.searchEntities(kbId, query);
      if (error) throw error;
      return data || [];
    },
    enabled: !!kbId && !!query,
  });
}

/**
 * Hook to get the relationships of an entity
 */
export function useEntityRelationships(kbId: string | null, entityId: string | null) {
  const client = useFluxbaseClient();

  return useQuery({
    queryKey: ["fluxbase", "knowledge-base", kbId, "entity-relationships", entityId],
    queryFn: async () => {
      if (!kbId || !entityId) return [];
      const { data, error } = await client.knowledgeBase.getEntityRelationships(kbId, entityId);
      if (error) throw error;
      return data || [];
    },
    enabled: !!kbId && !!entityId,
  });
}

/**
 * Hook to update a document (title, tags, metadata, content)
 */
export function useUpdateDocument() {
  const client = useFluxbaseClient();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      kbId,
      documentId,
      updates,
    }: {
      kbId: string;
      documentId: string;
      updates: UpdateDocumentRequest;
    }) => {
      const { data, error } = await client.knowledgeBase.updateDocument(kbId, documentId, updates);
      if (error) throw error;
      return data;
    },
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({
        queryKey: ["fluxbase", "knowledge-base", variables.kbId, "documents"],
      });
    },
  });
}

/**
 * Hook to delete documents matching a filter
 */
export function useDeleteDocumentsByFilter() {
  const client = useFluxbaseClient();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      kbId,
      filter,
    }: {
      kbId: string;
      filter: DeleteDocumentsByFilterRequest;
    }) => {
      const { data, error } = await client.knowledgeBase.deleteDocumentsByFilter(kbId, filter);
      if (error) throw error;
      return data;
    },
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({
        queryKey: ["fluxbase", "knowledge-base", variables.kbId, "documents"],
      });
    },
  });
}

/**
 * Hook to run a debug search (results plus effective search options)
 */
export function useKBDebugSearch(
  kbId: string | null,
  request: SearchKnowledgeBaseRequest | null,
) {
  const client = useFluxbaseClient();

  return useQuery({
    queryKey: ["fluxbase", "knowledge-base", kbId, "debug-search", request],
    queryFn: async () => {
      if (!kbId || !request) return null;
      const { data, error } = await client.knowledgeBase.debugSearch(kbId, request);
      if (error) throw error;
      return data;
    },
    enabled: !!kbId && !!request,
  });
}

/**
 * Hook to share a knowledge base with another user (owner only)
 */
export function useShareKB() {
  const client = useFluxbaseClient();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      kbId,
      userId,
      permission,
    }: {
      kbId: string;
      userId: string;
      permission: "viewer" | "editor" | "owner";
    }) => {
      const { data, error } = await client.knowledgeBase.share(kbId, {
        user_id: userId,
        permission,
      });
      if (error) throw error;
      return data;
    },
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({
        queryKey: ["fluxbase", "knowledge-base", variables.kbId, "permissions"],
      });
    },
  });
}

/**
 * Hook to revoke a user's knowledge base permission (owner only)
 */
export function useRevokeKBPermission() {
  const client = useFluxbaseClient();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ kbId, userId }: { kbId: string; userId: string }) => {
      const { error } = await client.knowledgeBase.revokePermission(kbId, userId);
      if (error) throw error;
    },
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({
        queryKey: ["fluxbase", "knowledge-base", variables.kbId, "permissions"],
      });
    },
  });
}

/**
 * Hook to list chatbots linked to a knowledge base
 */
export function useKBLinkedChatbots(kbId: string | null) {
  const client = useFluxbaseClient();

  return useQuery({
    queryKey: ["fluxbase", "knowledge-base", kbId, "chatbots"],
    queryFn: async () => {
      if (!kbId) return [];
      const { data, error } = await client.knowledgeBase.listChatbots(kbId);
      if (error) throw error;
      return data || [];
    },
    enabled: !!kbId,
  });
}
