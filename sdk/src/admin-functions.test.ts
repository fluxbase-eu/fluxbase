import { describe, it, expect, beforeEach, vi } from "vitest";
import { FluxbaseAdminFunctions } from "./admin-functions";
import { FluxbaseFetch } from "./fetch";
import * as bundling from "./bundling";
import type {
  EdgeFunction,
  EdgeFunctionExecution,
  SyncFunctionsResult,
} from "./types";

// Mock FluxbaseFetch
vi.mock("./fetch");

// Mock bundling module
vi.mock("./bundling", async () => {
  const actual = await vi.importActual("./bundling");
  return {
    ...actual,
    bundleCode: vi.fn(),
    loadEsbuild: vi.fn(),
    loadImportMap: vi.fn(),
    denoExternalPlugin: vi.fn(),
  };
});

describe("FluxbaseAdminFunctions", () => {
  let functions: FluxbaseAdminFunctions;
  let mockFetch: any;

  beforeEach(() => {
    vi.clearAllMocks();
    mockFetch = {
      get: vi.fn(),
      post: vi.fn(),
      put: vi.fn(),
      delete: vi.fn(),
    };
    functions = new FluxbaseAdminFunctions(mockFetch as unknown as FluxbaseFetch);
  });

  describe("create()", () => {
    it("should create a new edge function", async () => {
      const response: EdgeFunction = {
        id: "func-1",
        name: "my-function",
        namespace: "default",
        version: 1,
        enabled: true,
        created_at: "2024-01-26T10:00:00Z",
        updated_at: "2024-01-26T10:00:00Z",
      };

      vi.mocked(mockFetch.post).mockResolvedValue(response);

      const { data, error } = await functions.create({
        name: "my-function",
        code: 'export default async function handler(req) { return { hello: "world" } }',
        enabled: true,
      });

      expect(mockFetch.post).toHaveBeenCalledWith("/api/v1/functions", {
        name: "my-function",
        code: 'export default async function handler(req) { return { hello: "world" } }',
        enabled: true,
      });
      expect(error).toBeNull();
      expect(data).toBeDefined();
      expect(data!.name).toBe("my-function");
    });

    it("should handle create error", async () => {
      const errorMessage = "Function already exists";
      vi.mocked(mockFetch.post).mockRejectedValue(new Error(errorMessage));

      const { data, error } = await functions.create({
        name: "my-function",
        code: "export default function handler() {}",
      });

      expect(data).toBeNull();
      expect(error).toBeDefined();
      expect(error!.message).toBe(errorMessage);
    });
  });

  describe("listNamespaces()", () => {
    it("should list all namespaces", async () => {
      const response = { namespaces: ["default", "custom-ns"] };
      vi.mocked(mockFetch.get).mockResolvedValue(response);

      const { data, error } = await functions.listNamespaces();

      expect(mockFetch.get).toHaveBeenCalledWith(
        "/api/v1/admin/functions/namespaces"
      );
      expect(error).toBeNull();
      expect(data).toEqual(["default", "custom-ns"]);
    });

    it("should return default namespace when empty", async () => {
      const response = { namespaces: null };
      vi.mocked(mockFetch.get).mockResolvedValue(response);

      const { data, error } = await functions.listNamespaces();

      expect(error).toBeNull();
      expect(data).toEqual(["default"]);
    });

    it("should handle listNamespaces error", async () => {
      vi.mocked(mockFetch.get).mockRejectedValue(new Error("Network error"));

      const { data, error } = await functions.listNamespaces();

      expect(data).toBeNull();
      expect(error).toBeDefined();
    });
  });

  describe("list()", () => {
    it("should list all functions", async () => {
      const response: EdgeFunction[] = [
        {
          id: "func-1",
          name: "func-a",
          namespace: "default",
          version: 1,
          enabled: true,
          created_at: "2024-01-26T10:00:00Z",
          updated_at: "2024-01-26T10:00:00Z",
        },
        {
          id: "func-2",
          name: "func-b",
          namespace: "default",
          version: 1,
          enabled: false,
          created_at: "2024-01-26T11:00:00Z",
          updated_at: "2024-01-26T11:00:00Z",
        },
      ];

      vi.mocked(mockFetch.get).mockResolvedValue(response);

      const { data, error } = await functions.list();

      expect(mockFetch.get).toHaveBeenCalledWith("/api/v1/functions");
      expect(error).toBeNull();
      expect(data).toHaveLength(2);
    });

    it("should list functions by namespace", async () => {
      const response: EdgeFunction[] = [];
      vi.mocked(mockFetch.get).mockResolvedValue(response);

      const { data, error } = await functions.list("my-namespace");

      expect(mockFetch.get).toHaveBeenCalledWith(
        "/api/v1/functions?namespace=my-namespace"
      );
      expect(error).toBeNull();
    });

    it("should handle list error", async () => {
      vi.mocked(mockFetch.get).mockRejectedValue(new Error("Unauthorized"));

      const { data, error } = await functions.list();

      expect(data).toBeNull();
      expect(error).toBeDefined();
    });
  });

  describe("get()", () => {
    it("should get a specific function", async () => {
      const response: EdgeFunction = {
        id: "func-1",
        name: "my-function",
        namespace: "default",
        version: 1,
        enabled: true,
        code: 'export default function handler() { return { ok: true } }',
        created_at: "2024-01-26T10:00:00Z",
        updated_at: "2024-01-26T10:00:00Z",
      };

      vi.mocked(mockFetch.get).mockResolvedValue(response);

      const { data, error } = await functions.get("my-function");

      expect(mockFetch.get).toHaveBeenCalledWith("/api/v1/functions/my-function");
      expect(error).toBeNull();
      expect(data).toBeDefined();
      expect(data!.name).toBe("my-function");
    });

    it("should handle get error for non-existent function", async () => {
      vi.mocked(mockFetch.get).mockRejectedValue(new Error("Function not found"));

      const { data, error } = await functions.get("non-existent");

      expect(data).toBeNull();
      expect(error).toBeDefined();
      expect(error!.message).toBe("Function not found");
    });
  });

  describe("update()", () => {
    it("should update a function", async () => {
      const response: EdgeFunction = {
        id: "func-1",
        name: "my-function",
        namespace: "default",
        version: 2,
        enabled: false,
        description: "Updated description",
        created_at: "2024-01-26T10:00:00Z",
        updated_at: "2024-01-26T12:00:00Z",
      };

      vi.mocked(mockFetch.put).mockResolvedValue(response);

      const { data, error } = await functions.update("my-function", {
        enabled: false,
        description: "Updated description",
      });

      expect(mockFetch.put).toHaveBeenCalledWith("/api/v1/functions/my-function", {
        enabled: false,
        description: "Updated description",
      });
      expect(error).toBeNull();
      expect(data).toBeDefined();
      expect(data!.version).toBe(2);
    });

    it("should handle update error", async () => {
      vi.mocked(mockFetch.put).mockRejectedValue(new Error("Update failed"));

      const { data, error } = await functions.update("my-function", {
        enabled: true,
      });

      expect(data).toBeNull();
      expect(error).toBeDefined();
    });
  });

  describe("delete()", () => {
    it("should delete a function", async () => {
      vi.mocked(mockFetch.delete).mockResolvedValue({});

      const { data, error } = await functions.delete("my-function");

      expect(mockFetch.delete).toHaveBeenCalledWith(
        "/api/v1/functions/my-function"
      );
      expect(error).toBeNull();
      expect(data).toBeNull();
    });

    it("should handle delete error", async () => {
      vi.mocked(mockFetch.delete).mockRejectedValue(new Error("Delete failed"));

      const { data, error } = await functions.delete("my-function");

      expect(data).toBeNull();
      expect(error).toBeDefined();
    });
  });

  describe("getExecutions()", () => {
    it("should get function executions", async () => {
      const response: EdgeFunctionExecution[] = [
        {
          id: "exec-1",
          function_name: "my-function",
          status: "success",
          duration_ms: 150,
          executed_at: "2024-01-26T10:00:00Z",
        },
        {
          id: "exec-2",
          function_name: "my-function",
          status: "error",
          duration_ms: 50,
          error: "Timeout",
          executed_at: "2024-01-26T11:00:00Z",
        },
      ];

      vi.mocked(mockFetch.get).mockResolvedValue(response);

      const { data, error } = await functions.getExecutions("my-function");

      expect(mockFetch.get).toHaveBeenCalledWith(
        "/api/v1/functions/my-function/executions"
      );
      expect(error).toBeNull();
      expect(data).toHaveLength(2);
    });

    it("should get function executions with limit", async () => {
      const response: EdgeFunctionExecution[] = [];
      vi.mocked(mockFetch.get).mockResolvedValue(response);

      const { data, error } = await functions.getExecutions("my-function", 10);

      expect(mockFetch.get).toHaveBeenCalledWith(
        "/api/v1/functions/my-function/executions?limit=10"
      );
      expect(error).toBeNull();
    });

    it("should handle getExecutions error", async () => {
      vi.mocked(mockFetch.get).mockRejectedValue(new Error("Access denied"));

      const { data, error } = await functions.getExecutions("my-function");

      expect(data).toBeNull();
      expect(error).toBeDefined();
    });
  });

  describe("sync()", () => {
    it("should sync functions", async () => {
      const response: SyncFunctionsResult = {
        message: "Sync completed",
        namespace: "default",
        summary: {
          created: 2,
          updated: 1,
          deleted: 0,
          unchanged: 3,
          errors: 0,
        },
        details: {
          created: ["new-func-1", "new-func-2"],
          updated: ["updated-func"],
          deleted: [],
          unchanged: ["func-a", "func-b", "func-c"],
          errors: [],
        },
        dry_run: false,
      };

      vi.mocked(mockFetch.post).mockResolvedValue(response);

      const { data, error } = await functions.sync({
        namespace: "default",
        functions: [
          { name: "new-func-1", code: "export default function() {}" },
          { name: "new-func-2", code: "export default function() {}" },
        ],
        options: { delete_missing: false, dry_run: false },
      });

      expect(mockFetch.post).toHaveBeenCalledWith("/api/v1/admin/functions/sync", {
        namespace: "default",
        functions: [
          { name: "new-func-1", code: "export default function() {}" },
          { name: "new-func-2", code: "export default function() {}" },
        ],
        options: { delete_missing: false, dry_run: false },
      });
      expect(error).toBeNull();
      expect(data).toBeDefined();
      expect(data!.summary.created).toBe(2);
    });

    it("should handle sync error", async () => {
      vi.mocked(mockFetch.post).mockRejectedValue(new Error("Sync failed"));

      const { data, error } = await functions.sync({
        namespace: "default",
        functions: [],
      });

      expect(data).toBeNull();
      expect(error).toBeDefined();
    });
  });

  describe("syncWithBundling()", () => {
    it("should sync with empty functions array", async () => {
      const syncResponse: SyncFunctionsResult = {
        message: "Nothing to sync",
        namespace: "default",
        summary: {
          created: 0,
          updated: 0,
          deleted: 0,
          unchanged: 0,
          errors: 0,
        },
        details: {
          created: [],
          updated: [],
          deleted: [],
          unchanged: [],
          errors: [],
        },
        dry_run: false,
      };

      vi.mocked(mockFetch.post).mockResolvedValue(syncResponse);

      const { data, error } = await functions.syncWithBundling({
        namespace: "default",
        functions: [],
      });

      expect(error).toBeNull();
      expect(data).toBeDefined();
    });

    it("should sync without functions (calls sync directly)", async () => {
      const syncResponse: SyncFunctionsResult = {
        message: "Synced from filesystem",
        namespace: "default",
        summary: {
          created: 1,
          updated: 0,
          deleted: 0,
          unchanged: 0,
          errors: 0,
        },
        details: {
          created: ["func-1"],
          updated: [],
          deleted: [],
          unchanged: [],
          errors: [],
        },
        dry_run: false,
      };

      vi.mocked(mockFetch.post).mockResolvedValue(syncResponse);

      const { data, error } = await functions.syncWithBundling({
        namespace: "default",
      });

      expect(error).toBeNull();
    });

    it("should return error when esbuild is not available", async () => {
      vi.mocked(bundling.loadEsbuild).mockResolvedValue(false);

      const { data, error } = await functions.syncWithBundling({
        namespace: "default",
        functions: [
          { name: "func-1", code: 'import { foo } from "./bar"' },
        ],
      });

      expect(data).toBeNull();
      expect(error).toBeDefined();
      expect(error!.message).toContain("esbuild is required");
    });

    it("should bundle functions before syncing", async () => {
      vi.mocked(bundling.loadEsbuild).mockResolvedValue(true);
      vi.mocked(bundling.bundleCode).mockResolvedValue({
        code: "// bundled code",
        map: undefined,
      });

      const syncResponse: SyncFunctionsResult = {
        message: "Sync completed",
        namespace: "default",
        summary: {
          created: 1,
          updated: 0,
          deleted: 0,
          unchanged: 0,
          errors: 0,
        },
        details: {
          created: ["func-1"],
          updated: [],
          deleted: [],
          unchanged: [],
          errors: [],
        },
        dry_run: false,
      };

      vi.mocked(mockFetch.post).mockResolvedValue(syncResponse);

      const { data, error } = await functions.syncWithBundling({
        namespace: "default",
        functions: [
          { name: "func-1", code: 'import { foo } from "./bar"' },
        ],
      });

      expect(bundling.bundleCode).toHaveBeenCalled();
      expect(error).toBeNull();
      expect(data).toBeDefined();
    });

    it("should skip bundling for pre-bundled functions", async () => {
      vi.mocked(bundling.loadEsbuild).mockResolvedValue(true);

      const syncResponse: SyncFunctionsResult = {
        message: "Sync completed",
        namespace: "default",
        summary: {
          created: 1,
          updated: 0,
          deleted: 0,
          unchanged: 0,
          errors: 0,
        },
        details: {
          created: ["func-1"],
          updated: [],
          deleted: [],
          unchanged: [],
          errors: [],
        },
        dry_run: false,
      };

      vi.mocked(mockFetch.post).mockResolvedValue(syncResponse);

      const { data, error } = await functions.syncWithBundling({
        namespace: "default",
        functions: [
          {
            name: "func-1",
            code: "// already bundled",
            is_pre_bundled: true,
          },
        ],
      });

      expect(bundling.bundleCode).not.toHaveBeenCalled();
      expect(error).toBeNull();
    });

    it("should handle bundling error", async () => {
      vi.mocked(bundling.loadEsbuild).mockResolvedValue(true);
      vi.mocked(bundling.bundleCode).mockRejectedValue(
        new Error("Bundle failed")
      );

      const { data, error } = await functions.syncWithBundling({
        namespace: "default",
        functions: [
          { name: "func-1", code: "invalid code syntax {{{" },
        ],
      });

      expect(data).toBeNull();
      expect(error).toBeDefined();
      expect(error!.message).toBe("Bundle failed");
    });

    it("should use per-function sourceDir and nodePaths", async () => {
      vi.mocked(bundling.loadEsbuild).mockResolvedValue(true);
      vi.mocked(bundling.bundleCode).mockResolvedValue({
        code: "// bundled",
        map: undefined,
      });

      const syncResponse: SyncFunctionsResult = {
        message: "Sync completed",
        namespace: "default",
        summary: {
          created: 1,
          updated: 0,
          deleted: 0,
          unchanged: 0,
          errors: 0,
        },
        details: {
          created: ["func-1"],
          updated: [],
          deleted: [],
          unchanged: [],
          errors: [],
        },
        dry_run: false,
      };

      vi.mocked(mockFetch.post).mockResolvedValue(syncResponse);

      await functions.syncWithBundling(
        {
          namespace: "default",
          functions: [
            {
              name: "func-1",
              code: 'import x from "./local"',
              sourceDir: "/custom/path",
              nodePaths: ["/extra/modules"],
            },
          ],
        },
        { minify: true }
      );

      expect(bundling.bundleCode).toHaveBeenCalledWith(
        expect.objectContaining({
          code: 'import x from "./local"',
          baseDir: "/custom/path",
          nodePaths: ["/extra/modules"],
          minify: true,
        })
      );
    });
  });

  describe("bundleCode() static method", () => {
    it("should call bundleCode utility", async () => {
      vi.mocked(bundling.bundleCode).mockResolvedValue({
        code: "// bundled",
        map: undefined,
      });

      const result = await FluxbaseAdminFunctions.bundleCode({
        code: 'export default function() { return "hello" }',
      });

      expect(bundling.bundleCode).toHaveBeenCalled();
      expect(result.code).toBe("// bundled");
    });
  });
});
