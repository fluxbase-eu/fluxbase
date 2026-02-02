import { describe, it, expect, beforeEach, vi } from "vitest";
import { FluxbaseAdminRealtime } from "./admin-realtime";
import { FluxbaseFetch } from "./fetch";
import type {
  EnableRealtimeResponse,
  RealtimeTableStatus,
  ListRealtimeTablesResponse,
} from "./types";

// Mock FluxbaseFetch
vi.mock("./fetch");

describe("FluxbaseAdminRealtime", () => {
  let realtime: FluxbaseAdminRealtime;
  let mockFetch: any;

  beforeEach(() => {
    vi.clearAllMocks();
    mockFetch = {
      get: vi.fn(),
      post: vi.fn(),
      put: vi.fn(),
      patch: vi.fn(),
      delete: vi.fn(),
    };
    realtime = new FluxbaseAdminRealtime(mockFetch as unknown as FluxbaseFetch);
  });

  describe("enableRealtime()", () => {
    it("should enable realtime on a table", async () => {
      const response: EnableRealtimeResponse = {
        success: true,
        message: "Realtime enabled for public.products",
        schema: "public",
        table: "products",
      };

      vi.mocked(mockFetch.post).mockResolvedValue(response);

      const result = await realtime.enableRealtime("products");

      expect(mockFetch.post).toHaveBeenCalledWith("/api/v1/admin/realtime/tables", {
        schema: "public",
        table: "products",
        events: undefined,
        exclude: undefined,
      });
      expect(result.success).toBe(true);
    });

    it("should enable realtime with custom schema", async () => {
      const response: EnableRealtimeResponse = {
        success: true,
        message: "Realtime enabled for sales.orders",
        schema: "sales",
        table: "orders",
      };

      vi.mocked(mockFetch.post).mockResolvedValue(response);

      const result = await realtime.enableRealtime("orders", {
        schema: "sales",
      });

      expect(mockFetch.post).toHaveBeenCalledWith("/api/v1/admin/realtime/tables", {
        schema: "sales",
        table: "orders",
        events: undefined,
        exclude: undefined,
      });
      expect(result.schema).toBe("sales");
    });

    it("should enable realtime with specific events", async () => {
      const response: EnableRealtimeResponse = {
        success: true,
        message: "Realtime enabled",
        schema: "public",
        table: "audit_log",
      };

      vi.mocked(mockFetch.post).mockResolvedValue(response);

      const result = await realtime.enableRealtime("audit_log", {
        events: ["INSERT"],
      });

      expect(mockFetch.post).toHaveBeenCalledWith("/api/v1/admin/realtime/tables", {
        schema: "public",
        table: "audit_log",
        events: ["INSERT"],
        exclude: undefined,
      });
      expect(result.success).toBe(true);
    });

    it("should enable realtime with excluded columns", async () => {
      const response: EnableRealtimeResponse = {
        success: true,
        message: "Realtime enabled",
        schema: "public",
        table: "posts",
      };

      vi.mocked(mockFetch.post).mockResolvedValue(response);

      const result = await realtime.enableRealtime("posts", {
        exclude: ["content", "raw_html"],
      });

      expect(mockFetch.post).toHaveBeenCalledWith("/api/v1/admin/realtime/tables", {
        schema: "public",
        table: "posts",
        events: undefined,
        exclude: ["content", "raw_html"],
      });
      expect(result.success).toBe(true);
    });

    it("should enable realtime with all options", async () => {
      const response: EnableRealtimeResponse = {
        success: true,
        message: "Realtime enabled",
        schema: "app",
        table: "documents",
      };

      vi.mocked(mockFetch.post).mockResolvedValue(response);

      const result = await realtime.enableRealtime("documents", {
        schema: "app",
        events: ["INSERT", "UPDATE"],
        exclude: ["blob_data"],
      });

      expect(mockFetch.post).toHaveBeenCalledWith("/api/v1/admin/realtime/tables", {
        schema: "app",
        table: "documents",
        events: ["INSERT", "UPDATE"],
        exclude: ["blob_data"],
      });
      expect(result.success).toBe(true);
    });
  });

  describe("disableRealtime()", () => {
    it("should disable realtime on a table", async () => {
      const response = {
        success: true,
        message: "Realtime disabled for public.products",
      };

      vi.mocked(mockFetch.delete).mockResolvedValue(response);

      const result = await realtime.disableRealtime("public", "products");

      expect(mockFetch.delete).toHaveBeenCalledWith(
        "/api/v1/admin/realtime/tables/public/products"
      );
      expect(result.success).toBe(true);
    });

    it("should handle special characters in names", async () => {
      const response = { success: true, message: "Disabled" };
      vi.mocked(mockFetch.delete).mockResolvedValue(response);

      await realtime.disableRealtime("my-schema", "my-table");

      expect(mockFetch.delete).toHaveBeenCalledWith(
        "/api/v1/admin/realtime/tables/my-schema/my-table"
      );
    });
  });

  describe("listTables()", () => {
    it("should list all realtime-enabled tables", async () => {
      const response: ListRealtimeTablesResponse = {
        tables: [
          {
            schema: "public",
            table: "products",
            realtime_enabled: true,
            events: ["INSERT", "UPDATE", "DELETE"],
          },
          {
            schema: "public",
            table: "orders",
            realtime_enabled: true,
            events: ["INSERT"],
          },
        ],
        count: 2,
      };

      vi.mocked(mockFetch.get).mockResolvedValue(response);

      const result = await realtime.listTables();

      expect(mockFetch.get).toHaveBeenCalledWith("/api/v1/admin/realtime/tables");
      expect(result.count).toBe(2);
      expect(result.tables).toHaveLength(2);
    });

    it("should list tables including disabled", async () => {
      const response: ListRealtimeTablesResponse = {
        tables: [],
        count: 0,
      };

      vi.mocked(mockFetch.get).mockResolvedValue(response);

      const result = await realtime.listTables({ includeDisabled: true });

      expect(mockFetch.get).toHaveBeenCalledWith(
        "/api/v1/admin/realtime/tables?enabled=false"
      );
    });

    it("should list tables without options", async () => {
      const response: ListRealtimeTablesResponse = {
        tables: [],
        count: 0,
      };

      vi.mocked(mockFetch.get).mockResolvedValue(response);

      const result = await realtime.listTables();

      expect(mockFetch.get).toHaveBeenCalledWith("/api/v1/admin/realtime/tables");
    });
  });

  describe("getStatus()", () => {
    it("should get realtime status for a table", async () => {
      const response: RealtimeTableStatus = {
        schema: "public",
        table: "products",
        realtime_enabled: true,
        events: ["INSERT", "UPDATE", "DELETE"],
        excluded_columns: ["internal_notes"],
      };

      vi.mocked(mockFetch.get).mockResolvedValue(response);

      const result = await realtime.getStatus("public", "products");

      expect(mockFetch.get).toHaveBeenCalledWith(
        "/api/v1/admin/realtime/tables/public/products"
      );
      expect(result.realtime_enabled).toBe(true);
      expect(result.events).toContain("INSERT");
    });

    it("should get status for disabled table", async () => {
      const response: RealtimeTableStatus = {
        schema: "public",
        table: "archived",
        realtime_enabled: false,
        events: [],
      };

      vi.mocked(mockFetch.get).mockResolvedValue(response);

      const result = await realtime.getStatus("public", "archived");

      expect(result.realtime_enabled).toBe(false);
    });

    it("should handle special characters in table name", async () => {
      const response: RealtimeTableStatus = {
        schema: "my_schema",
        table: "my_table",
        realtime_enabled: true,
        events: ["INSERT"],
      };

      vi.mocked(mockFetch.get).mockResolvedValue(response);

      await realtime.getStatus("my_schema", "my_table");

      expect(mockFetch.get).toHaveBeenCalledWith(
        "/api/v1/admin/realtime/tables/my_schema/my_table"
      );
    });
  });

  describe("updateConfig()", () => {
    it("should update realtime config", async () => {
      const response = {
        success: true,
        message: "Config updated",
      };

      vi.mocked(mockFetch.patch).mockResolvedValue(response);

      const result = await realtime.updateConfig("public", "products", {
        events: ["INSERT", "UPDATE"],
      });

      expect(mockFetch.patch).toHaveBeenCalledWith(
        "/api/v1/admin/realtime/tables/public/products",
        { events: ["INSERT", "UPDATE"] }
      );
      expect(result.success).toBe(true);
    });

    it("should update excluded columns", async () => {
      const response = { success: true, message: "Updated" };
      vi.mocked(mockFetch.patch).mockResolvedValue(response);

      const result = await realtime.updateConfig("public", "posts", {
        exclude: ["raw_content", "search_vector"],
      });

      expect(mockFetch.patch).toHaveBeenCalledWith(
        "/api/v1/admin/realtime/tables/public/posts",
        { exclude: ["raw_content", "search_vector"] }
      );
      expect(result.success).toBe(true);
    });

    it("should clear excluded columns", async () => {
      const response = { success: true, message: "Updated" };
      vi.mocked(mockFetch.patch).mockResolvedValue(response);

      const result = await realtime.updateConfig("public", "posts", {
        exclude: [],
      });

      expect(mockFetch.patch).toHaveBeenCalledWith(
        "/api/v1/admin/realtime/tables/public/posts",
        { exclude: [] }
      );
      expect(result.success).toBe(true);
    });

    it("should update both events and exclude", async () => {
      const response = { success: true, message: "Updated" };
      vi.mocked(mockFetch.patch).mockResolvedValue(response);

      const result = await realtime.updateConfig("app", "documents", {
        events: ["INSERT"],
        exclude: ["blob_data"],
      });

      expect(mockFetch.patch).toHaveBeenCalledWith(
        "/api/v1/admin/realtime/tables/app/documents",
        { events: ["INSERT"], exclude: ["blob_data"] }
      );
      expect(result.success).toBe(true);
    });
  });
});
