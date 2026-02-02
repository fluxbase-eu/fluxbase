import { describe, it, expect, beforeEach, vi } from "vitest";
import { FluxbaseAdminRPC } from "./admin-rpc";
import { FluxbaseFetch } from "./fetch";
import type {
  RPCProcedure,
  RPCProcedureSummary,
  RPCExecution,
  RPCExecutionLog,
  SyncRPCResult,
} from "./types";

// Mock FluxbaseFetch
vi.mock("./fetch");

describe("FluxbaseAdminRPC", () => {
  let rpc: FluxbaseAdminRPC;
  let mockFetch: any;

  beforeEach(() => {
    vi.clearAllMocks();
    mockFetch = {
      get: vi.fn(),
      post: vi.fn(),
      put: vi.fn(),
      delete: vi.fn(),
    };
    rpc = new FluxbaseAdminRPC(mockFetch as unknown as FluxbaseFetch);
  });

  describe("sync()", () => {
    it("should sync RPC procedures without options", async () => {
      const response: SyncRPCResult = {
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
          created: ["get-user-orders"],
          updated: [],
          deleted: [],
          unchanged: [],
          errors: [],
        },
        dry_run: false,
      };

      vi.mocked(mockFetch.post).mockResolvedValue(response);

      const { data, error } = await rpc.sync();

      expect(mockFetch.post).toHaveBeenCalledWith("/api/v1/admin/rpc/sync", {
        namespace: "default",
        procedures: undefined,
        options: { delete_missing: false, dry_run: false },
      });
      expect(error).toBeNull();
      expect(data).toBeDefined();
    });

    it("should sync with provided procedures", async () => {
      const response: SyncRPCResult = {
        message: "Sync completed",
        namespace: "custom",
        summary: {
          created: 1,
          updated: 0,
          deleted: 0,
          unchanged: 0,
          errors: 0,
        },
        details: {
          created: ["my-procedure"],
          updated: [],
          deleted: [],
          unchanged: [],
          errors: [],
        },
        dry_run: false,
      };

      vi.mocked(mockFetch.post).mockResolvedValue(response);

      const { data, error } = await rpc.sync({
        namespace: "custom",
        procedures: [{ name: "my-procedure", code: "SELECT * FROM users" }],
        options: { delete_missing: true, dry_run: false },
      });

      expect(mockFetch.post).toHaveBeenCalledWith("/api/v1/admin/rpc/sync", {
        namespace: "custom",
        procedures: [{ name: "my-procedure", code: "SELECT * FROM users" }],
        options: { delete_missing: true, dry_run: false },
      });
      expect(error).toBeNull();
    });

    it("should handle sync error", async () => {
      vi.mocked(mockFetch.post).mockRejectedValue(new Error("Sync failed"));

      const { data, error } = await rpc.sync();

      expect(data).toBeNull();
      expect(error).toBeDefined();
    });
  });

  describe("list()", () => {
    it("should list all procedures", async () => {
      const response = {
        procedures: [
          {
            id: "proc-1",
            name: "get-user-orders",
            namespace: "default",
            enabled: true,
          },
          {
            id: "proc-2",
            name: "calculate-total",
            namespace: "default",
            enabled: true,
          },
        ] as RPCProcedureSummary[],
        count: 2,
      };

      vi.mocked(mockFetch.get).mockResolvedValue(response);

      const { data, error } = await rpc.list();

      expect(mockFetch.get).toHaveBeenCalledWith("/api/v1/admin/rpc/procedures");
      expect(error).toBeNull();
      expect(data).toHaveLength(2);
    });

    it("should list procedures by namespace", async () => {
      const response = { procedures: [], count: 0 };
      vi.mocked(mockFetch.get).mockResolvedValue(response);

      const { data, error } = await rpc.list("custom");

      expect(mockFetch.get).toHaveBeenCalledWith(
        "/api/v1/admin/rpc/procedures?namespace=custom"
      );
      expect(error).toBeNull();
    });

    it("should handle empty response", async () => {
      const response = { procedures: null, count: 0 };
      vi.mocked(mockFetch.get).mockResolvedValue(response);

      const { data, error } = await rpc.list();

      expect(error).toBeNull();
      expect(data).toEqual([]);
    });

    it("should handle error", async () => {
      vi.mocked(mockFetch.get).mockRejectedValue(new Error("Access denied"));

      const { data, error } = await rpc.list();

      expect(data).toBeNull();
      expect(error).toBeDefined();
    });
  });

  describe("listNamespaces()", () => {
    it("should list all namespaces", async () => {
      const response = { namespaces: ["default", "custom", "admin"] };
      vi.mocked(mockFetch.get).mockResolvedValue(response);

      const { data, error } = await rpc.listNamespaces();

      expect(mockFetch.get).toHaveBeenCalledWith("/api/v1/admin/rpc/namespaces");
      expect(error).toBeNull();
      expect(data).toEqual(["default", "custom", "admin"]);
    });

    it("should handle empty namespaces", async () => {
      const response = { namespaces: null };
      vi.mocked(mockFetch.get).mockResolvedValue(response);

      const { data, error } = await rpc.listNamespaces();

      expect(error).toBeNull();
      expect(data).toEqual([]);
    });

    it("should handle error", async () => {
      vi.mocked(mockFetch.get).mockRejectedValue(new Error("Network error"));

      const { data, error } = await rpc.listNamespaces();

      expect(data).toBeNull();
      expect(error).toBeDefined();
    });
  });

  describe("get()", () => {
    it("should get a specific procedure", async () => {
      const response: RPCProcedure = {
        id: "proc-1",
        name: "get-user-orders",
        namespace: "default",
        sql_query: "SELECT * FROM orders WHERE user_id = $1",
        enabled: true,
        created_at: "2024-01-26T10:00:00Z",
        updated_at: "2024-01-26T10:00:00Z",
      };

      vi.mocked(mockFetch.get).mockResolvedValue(response);

      const { data, error } = await rpc.get("default", "get-user-orders");

      expect(mockFetch.get).toHaveBeenCalledWith(
        "/api/v1/admin/rpc/procedures/default/get-user-orders"
      );
      expect(error).toBeNull();
      expect(data).toBeDefined();
      expect(data!.name).toBe("get-user-orders");
    });

    it("should handle special characters in names", async () => {
      vi.mocked(mockFetch.get).mockResolvedValue({});

      await rpc.get("my-namespace", "my-procedure");

      expect(mockFetch.get).toHaveBeenCalledWith(
        "/api/v1/admin/rpc/procedures/my-namespace/my-procedure"
      );
    });

    it("should handle error", async () => {
      vi.mocked(mockFetch.get).mockRejectedValue(new Error("Not found"));

      const { data, error } = await rpc.get("default", "unknown");

      expect(data).toBeNull();
      expect(error).toBeDefined();
    });
  });

  describe("update()", () => {
    it("should update a procedure", async () => {
      const response: RPCProcedure = {
        id: "proc-1",
        name: "get-user-orders",
        namespace: "default",
        enabled: false,
        max_execution_time_seconds: 60,
        created_at: "2024-01-26T10:00:00Z",
        updated_at: "2024-01-26T12:00:00Z",
      };

      vi.mocked(mockFetch.put).mockResolvedValue(response);

      const { data, error } = await rpc.update("default", "get-user-orders", {
        enabled: false,
        max_execution_time_seconds: 60,
      });

      expect(mockFetch.put).toHaveBeenCalledWith(
        "/api/v1/admin/rpc/procedures/default/get-user-orders",
        { enabled: false, max_execution_time_seconds: 60 }
      );
      expect(error).toBeNull();
      expect(data!.enabled).toBe(false);
    });

    it("should handle error", async () => {
      vi.mocked(mockFetch.put).mockRejectedValue(new Error("Update failed"));

      const { data, error } = await rpc.update("default", "proc", { enabled: true });

      expect(data).toBeNull();
      expect(error).toBeDefined();
    });
  });

  describe("toggle()", () => {
    it("should enable a procedure", async () => {
      const response: RPCProcedure = {
        id: "proc-1",
        name: "get-user-orders",
        namespace: "default",
        enabled: true,
        created_at: "2024-01-26T10:00:00Z",
        updated_at: "2024-01-26T12:00:00Z",
      };

      vi.mocked(mockFetch.put).mockResolvedValue(response);

      const { data, error } = await rpc.toggle("default", "get-user-orders", true);

      expect(mockFetch.put).toHaveBeenCalledWith(
        "/api/v1/admin/rpc/procedures/default/get-user-orders",
        { enabled: true }
      );
      expect(error).toBeNull();
      expect(data!.enabled).toBe(true);
    });

    it("should disable a procedure", async () => {
      const response: RPCProcedure = {
        id: "proc-1",
        name: "get-user-orders",
        namespace: "default",
        enabled: false,
        created_at: "2024-01-26T10:00:00Z",
        updated_at: "2024-01-26T12:00:00Z",
      };

      vi.mocked(mockFetch.put).mockResolvedValue(response);

      const { data, error } = await rpc.toggle("default", "get-user-orders", false);

      expect(mockFetch.put).toHaveBeenCalledWith(
        "/api/v1/admin/rpc/procedures/default/get-user-orders",
        { enabled: false }
      );
      expect(error).toBeNull();
      expect(data!.enabled).toBe(false);
    });
  });

  describe("delete()", () => {
    it("should delete a procedure", async () => {
      vi.mocked(mockFetch.delete).mockResolvedValue({});

      const { data, error } = await rpc.delete("default", "get-user-orders");

      expect(mockFetch.delete).toHaveBeenCalledWith(
        "/api/v1/admin/rpc/procedures/default/get-user-orders"
      );
      expect(error).toBeNull();
      expect(data).toBeNull();
    });

    it("should handle error", async () => {
      vi.mocked(mockFetch.delete).mockRejectedValue(new Error("Delete failed"));

      const { data, error } = await rpc.delete("default", "proc");

      expect(data).toBeNull();
      expect(error).toBeDefined();
    });
  });

  describe("listExecutions()", () => {
    it("should list all executions", async () => {
      const response = {
        executions: [
          {
            id: "exec-1",
            procedure_name: "get-user-orders",
            namespace: "default",
            status: "success",
            duration_ms: 50,
            executed_at: "2024-01-26T10:00:00Z",
          },
        ] as RPCExecution[],
        count: 1,
      };

      vi.mocked(mockFetch.get).mockResolvedValue(response);

      const { data, error } = await rpc.listExecutions();

      expect(mockFetch.get).toHaveBeenCalledWith("/api/v1/admin/rpc/executions");
      expect(error).toBeNull();
      expect(data).toHaveLength(1);
    });

    it("should list executions with filters", async () => {
      const response = { executions: [], count: 0 };
      vi.mocked(mockFetch.get).mockResolvedValue(response);

      const { data, error } = await rpc.listExecutions({
        namespace: "default",
        procedure: "get-user-orders",
        status: "failed",
        user_id: "user-123",
        limit: 10,
        offset: 5,
      });

      expect(mockFetch.get).toHaveBeenCalledWith(
        "/api/v1/admin/rpc/executions?namespace=default&procedure_name=get-user-orders&status=failed&user_id=user-123&limit=10&offset=5"
      );
      expect(error).toBeNull();
    });

    it("should handle empty executions", async () => {
      const response = { executions: null, count: 0 };
      vi.mocked(mockFetch.get).mockResolvedValue(response);

      const { data, error } = await rpc.listExecutions();

      expect(error).toBeNull();
      expect(data).toEqual([]);
    });

    it("should handle error", async () => {
      vi.mocked(mockFetch.get).mockRejectedValue(new Error("Access denied"));

      const { data, error } = await rpc.listExecutions();

      expect(data).toBeNull();
      expect(error).toBeDefined();
    });
  });

  describe("getExecution()", () => {
    it("should get a specific execution", async () => {
      const response: RPCExecution = {
        id: "exec-1",
        procedure_name: "get-user-orders",
        namespace: "default",
        status: "success",
        duration_ms: 50,
        result: { orders: [] },
        executed_at: "2024-01-26T10:00:00Z",
      };

      vi.mocked(mockFetch.get).mockResolvedValue(response);

      const { data, error } = await rpc.getExecution("exec-1");

      expect(mockFetch.get).toHaveBeenCalledWith(
        "/api/v1/admin/rpc/executions/exec-1"
      );
      expect(error).toBeNull();
      expect(data!.status).toBe("success");
    });

    it("should handle error", async () => {
      vi.mocked(mockFetch.get).mockRejectedValue(new Error("Not found"));

      const { data, error } = await rpc.getExecution("unknown");

      expect(data).toBeNull();
      expect(error).toBeDefined();
    });
  });

  describe("getExecutionLogs()", () => {
    it("should get execution logs", async () => {
      const response = {
        logs: [
          { line: 1, level: "info", message: "Starting execution", timestamp: "2024-01-26T10:00:00Z" },
          { line: 2, level: "info", message: "Query completed", timestamp: "2024-01-26T10:00:01Z" },
        ] as RPCExecutionLog[],
        count: 2,
      };

      vi.mocked(mockFetch.get).mockResolvedValue(response);

      const { data, error } = await rpc.getExecutionLogs("exec-1");

      expect(mockFetch.get).toHaveBeenCalledWith(
        "/api/v1/admin/rpc/executions/exec-1/logs"
      );
      expect(error).toBeNull();
      expect(data).toHaveLength(2);
    });

    it("should get logs after specific line", async () => {
      const response = { logs: [], count: 0 };
      vi.mocked(mockFetch.get).mockResolvedValue(response);

      const { data, error } = await rpc.getExecutionLogs("exec-1", 10);

      expect(mockFetch.get).toHaveBeenCalledWith(
        "/api/v1/admin/rpc/executions/exec-1/logs?after=10"
      );
      expect(error).toBeNull();
    });

    it("should handle empty logs", async () => {
      const response = { logs: null, count: 0 };
      vi.mocked(mockFetch.get).mockResolvedValue(response);

      const { data, error } = await rpc.getExecutionLogs("exec-1");

      expect(error).toBeNull();
      expect(data).toEqual([]);
    });

    it("should handle error", async () => {
      vi.mocked(mockFetch.get).mockRejectedValue(new Error("Access denied"));

      const { data, error } = await rpc.getExecutionLogs("exec-1");

      expect(data).toBeNull();
      expect(error).toBeDefined();
    });
  });

  describe("cancelExecution()", () => {
    it("should cancel an execution", async () => {
      const response: RPCExecution = {
        id: "exec-1",
        procedure_name: "get-user-orders",
        namespace: "default",
        status: "cancelled",
        executed_at: "2024-01-26T10:00:00Z",
      };

      vi.mocked(mockFetch.post).mockResolvedValue(response);

      const { data, error } = await rpc.cancelExecution("exec-1");

      expect(mockFetch.post).toHaveBeenCalledWith(
        "/api/v1/admin/rpc/executions/exec-1/cancel",
        {}
      );
      expect(error).toBeNull();
      expect(data!.status).toBe("cancelled");
    });

    it("should handle error", async () => {
      vi.mocked(mockFetch.post).mockRejectedValue(new Error("Cannot cancel"));

      const { data, error } = await rpc.cancelExecution("exec-1");

      expect(data).toBeNull();
      expect(error).toBeDefined();
    });
  });
});
