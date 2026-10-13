/**
 * Tests for the admin AI hooks (metrics, conversations/audit, table export,
 * sync configs, document permissions) and the AI user-surface hooks added
 * alongside them.
 */

import { describe, it, expect, vi } from "vitest";
import { renderHook, waitFor, act } from "@testing-library/react";
import {
  useAdminAIMetrics,
  useAdminAIConversations,
  useAdminAIAuditLog,
  useAdminSyncConfigs,
  useAdminCreateSyncConfig,
  useAdminDocumentPermissions,
  useAdminGrantDocumentPermission,
  useAdminAIProviders,
} from "./use-admin-ai";
import { useUsage, useUpdateConversation } from "./use-ai";
import {
  useShareKB,
  useRevokeKBPermission,
  useKBLinkedChatbots,
  useUpdateDocument,
  useDeleteDocumentsByFilter,
} from "./use-knowledge-base";
import { useVectorCapabilities } from "./use-vector";
import { createMockAIClient, createWrapper } from "./test-utils";

describe("useAdminAIMetrics", () => {
  it("should fetch aggregated metrics", async () => {
    const client = createMockAIClient();
    const metrics = { total_requests: 7, error_rate: 0 };
    (client.admin.ai.getMetrics as ReturnType<typeof vi.fn>).mockResolvedValue({
      data: metrics,
      error: null,
    });

    const { result } = renderHook(() => useAdminAIMetrics(), {
      wrapper: createWrapper(client),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(client.admin.ai.getMetrics).toHaveBeenCalled();
    expect(result.current.data).toEqual(metrics);
  });
});

describe("useAdminAIConversations", () => {
  it("should pass the query through", async () => {
    const client = createMockAIClient();
    const { result } = renderHook(
      () => useAdminAIConversations({ chatbot_id: "bot-1" }),
      { wrapper: createWrapper(client) },
    );

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(client.admin.ai.listConversations).toHaveBeenCalledWith({ chatbot_id: "bot-1" });
  });
});

describe("useAdminAIAuditLog", () => {
  it("should fetch audit entries", async () => {
    const client = createMockAIClient();
    const { result } = renderHook(() => useAdminAIAuditLog({ limit: 5 }), {
      wrapper: createWrapper(client),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(client.admin.ai.getAuditLog).toHaveBeenCalledWith({ limit: 5 });
  });
});

describe("useAdminSyncConfigs", () => {
  it("should not fetch without a KB id", () => {
    const client = createMockAIClient();
    const { result } = renderHook(() => useAdminSyncConfigs(null), {
      wrapper: createWrapper(client),
    });

    expect(result.current.fetchStatus).toBe("idle");
    expect(client.admin.ai.listSyncConfigs).not.toHaveBeenCalled();
  });

  it("should fetch sync configs for a KB", async () => {
    const client = createMockAIClient();
    const { result } = renderHook(() => useAdminSyncConfigs("kb-1"), {
      wrapper: createWrapper(client),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(client.admin.ai.listSyncConfigs).toHaveBeenCalledWith("kb-1");
  });
});

describe("useAdminCreateSyncConfig", () => {
  it("should create and invalidate the sync-config list", async () => {
    const client = createMockAIClient();
    (client.admin.ai.createSyncConfig as ReturnType<typeof vi.fn>).mockResolvedValue({
      data: { id: "sync-1" },
      error: null,
    });

    const { result } = renderHook(() => useAdminCreateSyncConfig(), {
      wrapper: createWrapper(client),
    });

    await act(async () => {
      await result.current.mutateAsync({
        kbId: "kb-1",
        request: { schema_name: "public", table_name: "trips" },
      });
    });

    expect(client.admin.ai.createSyncConfig).toHaveBeenCalledWith("kb-1", {
      schema_name: "public",
      table_name: "trips",
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
  });
});

describe("useAdminDocumentPermissions + grant", () => {
  it("should list permissions for a document", async () => {
    const client = createMockAIClient();
    (client.admin.ai.listDocumentPermissions as ReturnType<typeof vi.fn>).mockResolvedValue({
      data: [{ id: "perm-1" }],
      error: null,
    });

    const { result } = renderHook(
      () => useAdminDocumentPermissions("kb-1", "doc-1"),
      { wrapper: createWrapper(client) },
    );

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(client.admin.ai.listDocumentPermissions).toHaveBeenCalledWith("kb-1", "doc-1");
  });

  it("should grant a permission", async () => {
    const client = createMockAIClient();
    const { result } = renderHook(() => useAdminGrantDocumentPermission(), {
      wrapper: createWrapper(client),
    });

    await act(async () => {
      await result.current.mutateAsync({
        kbId: "kb-1",
        documentId: "doc-1",
        request: { user_id: "u-1", permission: "viewer" },
      });
    });

    expect(client.admin.ai.grantDocumentPermission).toHaveBeenCalledWith("kb-1", "doc-1", {
      user_id: "u-1",
      permission: "viewer",
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
  });
});

describe("useAdminAIProviders", () => {
  it("should list providers", async () => {
    const client = createMockAIClient();
    const { result } = renderHook(() => useAdminAIProviders(), {
      wrapper: createWrapper(client),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(client.admin.ai.listProviders).toHaveBeenCalled();
  });
});

describe("useUsage", () => {
  it("should not fetch without a chatbot id", () => {
    const client = createMockAIClient();
    const { result } = renderHook(() => useUsage(null), {
      wrapper: createWrapper(client),
    });

    expect(result.current.fetchStatus).toBe("idle");
    expect(client.ai.getUsage).not.toHaveBeenCalled();
  });

  it("should fetch usage for a chatbot", async () => {
    const client = createMockAIClient();
    const { result } = renderHook(() => useUsage("bot-1"), {
      wrapper: createWrapper(client),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(client.ai.getUsage).toHaveBeenCalledWith("bot-1");
  });
});

describe("useUpdateConversation", () => {
  it("should patch the conversation", async () => {
    const client = createMockAIClient();
    const { result } = renderHook(() => useUpdateConversation(), {
      wrapper: createWrapper(client),
    });

    await act(async () => {
      await result.current.mutateAsync({
        conversationId: "conv-1",
        updates: { title: "renamed" },
      });
    });

    expect(client.ai.updateConversation).toHaveBeenCalledWith("conv-1", { title: "renamed" });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
  });
});

describe("KB sharing hooks", () => {
  it("useShareKB posts the grant", async () => {
    const client = createMockAIClient();
    const { result } = renderHook(() => useShareKB(), {
      wrapper: createWrapper(client),
    });

    await act(async () => {
      await result.current.mutateAsync({
        kbId: "kb-1",
        userId: "u-2",
        permission: "viewer",
      });
    });

    expect(client.knowledgeBase.share).toHaveBeenCalledWith("kb-1", {
      user_id: "u-2",
      permission: "viewer",
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
  });

  it("useRevokeKBPermission deletes the grant", async () => {
    const client = createMockAIClient();
    const { result } = renderHook(() => useRevokeKBPermission(), {
      wrapper: createWrapper(client),
    });

    await act(async () => {
      await result.current.mutateAsync({ kbId: "kb-1", userId: "u-2" });
    });

    expect(client.knowledgeBase.revokePermission).toHaveBeenCalledWith("kb-1", "u-2");
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
  });

  it("useKBLinkedChatbots lists linked chatbots", async () => {
    const client = createMockAIClient();
    const { result } = renderHook(() => useKBLinkedChatbots("kb-1"), {
      wrapper: createWrapper(client),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(client.knowledgeBase.listChatbots).toHaveBeenCalledWith("kb-1");
  });
});

describe("document mutation hooks", () => {
  it("useUpdateDocument patches a document", async () => {
    const client = createMockAIClient();
    const { result } = renderHook(() => useUpdateDocument(), {
      wrapper: createWrapper(client),
    });

    await act(async () => {
      await result.current.mutateAsync({
        kbId: "kb-1",
        documentId: "doc-1",
        updates: { title: "new title" },
      });
    });

    expect(client.knowledgeBase.updateDocument).toHaveBeenCalledWith("kb-1", "doc-1", {
      title: "new title",
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
  });

  it("useDeleteDocumentsByFilter deletes by filter", async () => {
    const client = createMockAIClient();
    const { result } = renderHook(() => useDeleteDocumentsByFilter(), {
      wrapper: createWrapper(client),
    });

    await act(async () => {
      await result.current.mutateAsync({
        kbId: "kb-1",
        filter: { tags: ["stale"] },
      });
    });

    expect(client.knowledgeBase.deleteDocumentsByFilter).toHaveBeenCalledWith("kb-1", {
      tags: ["stale"],
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
  });
});

describe("useVectorCapabilities", () => {
  it("should fetch capabilities", async () => {
    const client = createMockAIClient();
    (client.vector.getCapabilities as ReturnType<typeof vi.fn>).mockResolvedValue({
      data: { enabled: true, embedding_model: "text-embedding-3-small" },
      error: null,
    });

    const { result } = renderHook(() => useVectorCapabilities(), {
      wrapper: createWrapper(client),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(client.vector.getCapabilities).toHaveBeenCalled();
    expect(result.current.data!.enabled).toBe(true);
  });
});
