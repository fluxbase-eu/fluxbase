/**
 * Tests for the KB sharing, debug-search, and chatbot-link methods
 * (the user knowledge-base surface added alongside the PATCH/DELETE routes).
 */

import { describe, it, expect, beforeEach, vi } from "vitest";
import { FluxbaseKnowledgeBase } from "./knowledge-base";
import { FluxbaseFetch } from "./fetch";

vi.mock("./fetch");

describe("FluxbaseKnowledgeBase — sharing, debug search, chatbot links", () => {
  let kb: FluxbaseKnowledgeBase;
  let mockFetch: any;

  beforeEach(() => {
    vi.clearAllMocks();
    mockFetch = {
      request: vi.fn(),
      get: vi.fn(),
      post: vi.fn(),
      patch: vi.fn(),
      delete: vi.fn(),
    };
    kb = new FluxbaseKnowledgeBase(mockFetch as unknown as FluxbaseFetch);
  });

  describe("share()", () => {
    it("should POST to the share route", async () => {
      const grant = { id: "perm-1", user_id: "u-2", permission: "viewer" };
      vi.mocked(mockFetch.request).mockResolvedValue(grant);

      const { data, error } = await kb.share("kb-1", { user_id: "u-2", permission: "viewer" });

      expect(mockFetch.request).toHaveBeenCalledWith("/api/v1/ai/knowledge-bases/kb-1/share", {
        method: "POST",
        body: { user_id: "u-2", permission: "viewer" },
      });
      expect(error).toBeNull();
      expect(data!.id).toBe("perm-1");
    });

    it("should surface errors", async () => {
      vi.mocked(mockFetch.request).mockRejectedValue(new Error("only owner can share"));

      const { data, error } = await kb.share("kb-1", { user_id: "u-2", permission: "viewer" });

      expect(data).toBeNull();
      expect(error).toBeDefined();
    });
  });

  describe("listPermissions()", () => {
    it("should GET the permissions route", async () => {
      vi.mocked(mockFetch.get).mockResolvedValue([{ id: "perm-1" }]);

      const { data, error } = await kb.listPermissions("kb-1");

      expect(mockFetch.get).toHaveBeenCalledWith("/api/v1/ai/knowledge-bases/kb-1/permissions");
      expect(error).toBeNull();
      expect(data).toHaveLength(1);
    });
  });

  describe("revokePermission()", () => {
    it("should DELETE the user's permission", async () => {
      vi.mocked(mockFetch.delete).mockResolvedValue(undefined);

      const { error } = await kb.revokePermission("kb-1", "u-2");

      expect(mockFetch.delete).toHaveBeenCalledWith("/api/v1/ai/knowledge-bases/kb-1/permissions/u-2");
      expect(error).toBeNull();
    });
  });

  describe("debugSearch()", () => {
    it("should POST the debug-search route", async () => {
      const response = {
        query: "hello",
        results: [],
        result_count: 0,
        search_options: { semantic_weight: 0.7 },
      };
      vi.mocked(mockFetch.request).mockResolvedValue(response);

      const { data, error } = await kb.debugSearch("kb-1", { query: "hello", limit: 5 });

      expect(mockFetch.request).toHaveBeenCalledWith("/api/v1/ai/knowledge-bases/kb-1/debug-search", {
        method: "POST",
        body: { query: "hello", limit: 5 },
      });
      expect(error).toBeNull();
      expect(data!.search_options).toBeDefined();
    });
  });

  describe("listChatbots()", () => {
    it("should GET the linked chatbots", async () => {
      vi.mocked(mockFetch.get).mockResolvedValue([{ id: "bot-1", name: "assistant" }]);

      const { data, error } = await kb.listChatbots("kb-1");

      expect(mockFetch.get).toHaveBeenCalledWith("/api/v1/ai/knowledge-bases/kb-1/chatbots");
      expect(error).toBeNull();
      expect(data![0].name).toBe("assistant");
    });
  });

  describe("search()", () => {
    it("should send the aligned request shape", async () => {
      vi.mocked(mockFetch.request).mockResolvedValue({ results: [], count: 0, query: "q" });

      await kb.search("kb-1", {
        query: "how do I reset",
        limit: 5,
        mode: "hybrid",
        semantic_weight: 0.7,
        threshold: 0.5,
        rerank: true,
      });

      expect(mockFetch.request).toHaveBeenCalledWith("/api/v1/ai/knowledge-bases/kb-1/search", {
        method: "POST",
        body: {
          query: "how do I reset",
          limit: 5,
          mode: "hybrid",
          semantic_weight: 0.7,
          threshold: 0.5,
          rerank: true,
        },
      });
    });
  });

  describe("update() and delete()", () => {
    it("should PATCH the user KB route", async () => {
      vi.mocked(mockFetch.request).mockResolvedValue({ id: "kb-1", name: "renamed" });

      const { data, error } = await kb.update("kb-1", {
        entity_extraction_strategy: "hybrid",
        rerank_enabled: false,
      });

      expect(mockFetch.request).toHaveBeenCalledWith("/api/v1/ai/knowledge-bases/kb-1", {
        method: "PATCH",
        body: { entity_extraction_strategy: "hybrid", rerank_enabled: false },
      });
      expect(error).toBeNull();
      expect(data!.name).toBe("renamed");
    });

    it("should DELETE the user KB route", async () => {
      vi.mocked(mockFetch.request).mockResolvedValue(undefined);

      const { error } = await kb.delete("kb-1");

      expect(mockFetch.request).toHaveBeenCalledWith("/api/v1/ai/knowledge-bases/kb-1", {
        method: "DELETE",
      });
      expect(error).toBeNull();
    });
  });
});
