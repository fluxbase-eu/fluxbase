import { describe, it, expect, beforeEach, vi } from "vitest";
import { FluxbaseAdminMigrations } from "./admin-migrations";
import { FluxbaseFetch } from "./fetch";
import type {
  Migration,
  MigrationExecution,
  SyncMigrationsResult,
} from "./types";

// Mock FluxbaseFetch
vi.mock("./fetch");

describe("FluxbaseAdminMigrations", () => {
  let migrations: FluxbaseAdminMigrations;
  let mockFetch: any;

  beforeEach(() => {
    vi.clearAllMocks();
    mockFetch = {
      get: vi.fn(),
      post: vi.fn(),
      put: vi.fn(),
      delete: vi.fn(),
    };
    migrations = new FluxbaseAdminMigrations(mockFetch as unknown as FluxbaseFetch);
  });

  describe("register()", () => {
    it("should register a migration", () => {
      const { error } = migrations.register({
        name: "001_create_users",
        namespace: "myapp",
        up_sql: "CREATE TABLE users (id UUID PRIMARY KEY)",
        down_sql: "DROP TABLE users",
        description: "Create users table",
      });

      expect(error).toBeNull();
    });

    it("should register migration without namespace (default)", () => {
      const { error } = migrations.register({
        name: "001_create_users",
        up_sql: "CREATE TABLE users (id UUID PRIMARY KEY)",
      });

      expect(error).toBeNull();
    });

    it("should return error when name is missing", () => {
      const { error } = migrations.register({
        name: "",
        up_sql: "CREATE TABLE users (id UUID PRIMARY KEY)",
      });

      expect(error).toBeDefined();
      expect(error!.message).toContain("name and up_sql are required");
    });

    it("should return error when up_sql is missing", () => {
      const { error } = migrations.register({
        name: "001_create_users",
        up_sql: "",
      });

      expect(error).toBeDefined();
      expect(error!.message).toContain("name and up_sql are required");
    });
  });

  describe("sync()", () => {
    it("should sync registered migrations", async () => {
      // Register some migrations
      migrations.register({
        name: "001_create_users",
        namespace: "myapp",
        up_sql: "CREATE TABLE users (id UUID PRIMARY KEY)",
        down_sql: "DROP TABLE users",
      });

      const response: SyncMigrationsResult = {
        message: "Sync completed",
        namespace: "myapp",
        summary: {
          created: 1,
          updated: 0,
          unchanged: 0,
          skipped: 0,
          applied: 0,
          errors: 0,
        },
        details: {
          created: ["001_create_users"],
          updated: [],
          unchanged: [],
          skipped: [],
          applied: [],
          errors: [],
        },
        dry_run: false,
      };

      vi.mocked(mockFetch.post).mockResolvedValue(response);

      const { data, error } = await migrations.sync();

      expect(mockFetch.post).toHaveBeenCalledWith(
        "/api/v1/admin/migrations/sync",
        expect.objectContaining({
          namespace: "myapp",
          migrations: expect.arrayContaining([
            expect.objectContaining({
              name: "001_create_users",
            }),
          ]),
        })
      );
      expect(error).toBeNull();
      expect(data).toBeDefined();
      expect(data!.summary.created).toBe(1);
    });

    it("should sync with auto_apply option", async () => {
      migrations.register({
        name: "001_create_users",
        up_sql: "CREATE TABLE users (id UUID PRIMARY KEY)",
      });

      const response: SyncMigrationsResult = {
        message: "Sync completed",
        namespace: "default",
        summary: {
          created: 1,
          updated: 0,
          unchanged: 0,
          skipped: 0,
          applied: 1,
          errors: 0,
        },
        details: {
          created: ["001_create_users"],
          updated: [],
          unchanged: [],
          skipped: [],
          applied: ["001_create_users"],
          errors: [],
        },
        dry_run: false,
      };

      vi.mocked(mockFetch.post).mockResolvedValue(response);
      vi.mocked(mockFetch.post).mockResolvedValueOnce(response);
      // Mock schema refresh
      vi.mocked(mockFetch.post).mockResolvedValue({ message: "Refreshed", tables: 1, views: 0 });

      const { data, error } = await migrations.sync({ auto_apply: true });

      expect(error).toBeNull();
    });

    it("should handle dry_run option", async () => {
      migrations.register({
        name: "001_create_users",
        up_sql: "CREATE TABLE users (id UUID PRIMARY KEY)",
      });

      const response: SyncMigrationsResult = {
        message: "Dry run completed",
        namespace: "default",
        summary: {
          created: 1,
          updated: 0,
          unchanged: 0,
          skipped: 0,
          applied: 0,
          errors: 0,
        },
        details: {
          created: ["001_create_users"],
          updated: [],
          unchanged: [],
          skipped: [],
          applied: [],
          errors: [],
        },
        dry_run: true,
      };

      vi.mocked(mockFetch.post).mockResolvedValue(response);

      const { data, error } = await migrations.sync({ dry_run: true });

      expect(mockFetch.post).toHaveBeenCalledWith(
        "/api/v1/admin/migrations/sync",
        expect.objectContaining({
          options: expect.objectContaining({
            dry_run: true,
          }),
        })
      );
      expect(error).toBeNull();
      expect(data!.dry_run).toBe(true);
    });

    it("should handle sync errors with 422 status", async () => {
      migrations.register({
        name: "001_invalid",
        namespace: "test",
        up_sql: "INVALID SQL",
      });

      const syncError = new Error("Sync failed") as any;
      syncError.status = 422;
      syncError.details = {
        message: "Sync failed with errors",
        namespace: "test",
        summary: {
          created: 0,
          updated: 0,
          unchanged: 0,
          skipped: 0,
          applied: 0,
          errors: 1,
        },
        details: {
          created: [],
          updated: [],
          unchanged: [],
          skipped: [],
          applied: [],
          errors: [{ name: "001_invalid", error: "Invalid SQL" }],
        },
        dry_run: false,
      };

      vi.mocked(mockFetch.post).mockRejectedValue(syncError);

      const { data, error } = await migrations.sync();

      expect(data).toBeDefined();
      expect(error).toBeDefined();
    });

    it("should handle non-422 errors", async () => {
      migrations.register({
        name: "001_test",
        up_sql: "CREATE TABLE test (id INT)",
      });

      vi.mocked(mockFetch.post).mockRejectedValue(new Error("Network error"));

      const { data, error } = await migrations.sync();

      expect(data).toBeNull();
      expect(error).toBeDefined();
      expect(error!.message).toBe("Network error");
    });

    it("should combine results from multiple namespaces", async () => {
      migrations.register({
        name: "001_ns1",
        namespace: "ns1",
        up_sql: "CREATE TABLE t1 (id INT)",
      });
      migrations.register({
        name: "001_ns2",
        namespace: "ns2",
        up_sql: "CREATE TABLE t2 (id INT)",
      });

      const response1: SyncMigrationsResult = {
        message: "Sync ns1",
        namespace: "ns1",
        summary: {
          created: 1,
          updated: 0,
          unchanged: 0,
          skipped: 0,
          applied: 0,
          errors: 0,
        },
        details: {
          created: ["001_ns1"],
          updated: [],
          unchanged: [],
          skipped: [],
          applied: [],
          errors: [],
        },
        dry_run: false,
      };

      const response2: SyncMigrationsResult = {
        message: "Sync ns2",
        namespace: "ns2",
        summary: {
          created: 1,
          updated: 0,
          unchanged: 0,
          skipped: 0,
          applied: 0,
          errors: 0,
        },
        details: {
          created: ["001_ns2"],
          updated: [],
          unchanged: [],
          skipped: [],
          applied: [],
          errors: [],
        },
        dry_run: false,
      };

      vi.mocked(mockFetch.post)
        .mockResolvedValueOnce(response1)
        .mockResolvedValueOnce(response2);

      const { data, error } = await migrations.sync();

      expect(error).toBeNull();
      expect(data!.summary.created).toBe(2);
      expect(data!.details.created).toContain("001_ns1");
      expect(data!.details.created).toContain("001_ns2");
    });
  });

  describe("create()", () => {
    it("should create a migration", async () => {
      const response: Migration = {
        id: "mig-1",
        name: "001_create_users",
        namespace: "myapp",
        status: "pending",
        up_sql: "CREATE TABLE users (id UUID PRIMARY KEY)",
        down_sql: "DROP TABLE users",
        created_at: "2024-01-26T10:00:00Z",
      };

      vi.mocked(mockFetch.post).mockResolvedValue(response);

      const { data, error } = await migrations.create({
        namespace: "myapp",
        name: "001_create_users",
        up_sql: "CREATE TABLE users (id UUID PRIMARY KEY)",
        down_sql: "DROP TABLE users",
        description: "Create users table",
      });

      expect(mockFetch.post).toHaveBeenCalledWith("/api/v1/admin/migrations", {
        namespace: "myapp",
        name: "001_create_users",
        up_sql: "CREATE TABLE users (id UUID PRIMARY KEY)",
        down_sql: "DROP TABLE users",
        description: "Create users table",
      });
      expect(error).toBeNull();
      expect(data).toBeDefined();
    });

    it("should handle create error", async () => {
      vi.mocked(mockFetch.post).mockRejectedValue(new Error("Already exists"));

      const { data, error } = await migrations.create({
        name: "001_create_users",
        up_sql: "CREATE TABLE users (id UUID PRIMARY KEY)",
      });

      expect(data).toBeNull();
      expect(error).toBeDefined();
    });
  });

  describe("list()", () => {
    it("should list migrations", async () => {
      const response: Migration[] = [
        {
          id: "mig-1",
          name: "001_create_users",
          namespace: "default",
          status: "applied",
          created_at: "2024-01-26T10:00:00Z",
        },
        {
          id: "mig-2",
          name: "002_add_posts",
          namespace: "default",
          status: "pending",
          created_at: "2024-01-26T11:00:00Z",
        },
      ];

      vi.mocked(mockFetch.get).mockResolvedValue(response);

      const { data, error } = await migrations.list();

      expect(mockFetch.get).toHaveBeenCalledWith(
        "/api/v1/admin/migrations?namespace=default"
      );
      expect(error).toBeNull();
      expect(data).toHaveLength(2);
    });

    it("should list migrations by namespace and status", async () => {
      vi.mocked(mockFetch.get).mockResolvedValue([]);

      const { data, error } = await migrations.list("myapp", "pending");

      expect(mockFetch.get).toHaveBeenCalledWith(
        "/api/v1/admin/migrations?namespace=myapp&status=pending"
      );
      expect(error).toBeNull();
    });

    it("should handle error", async () => {
      vi.mocked(mockFetch.get).mockRejectedValue(new Error("Access denied"));

      const { data, error } = await migrations.list();

      expect(data).toBeNull();
      expect(error).toBeDefined();
    });
  });

  describe("get()", () => {
    it("should get a specific migration", async () => {
      const response: Migration = {
        id: "mig-1",
        name: "001_create_users",
        namespace: "default",
        status: "applied",
        up_sql: "CREATE TABLE users (id UUID PRIMARY KEY)",
        created_at: "2024-01-26T10:00:00Z",
      };

      vi.mocked(mockFetch.get).mockResolvedValue(response);

      const { data, error } = await migrations.get("001_create_users");

      expect(mockFetch.get).toHaveBeenCalledWith(
        "/api/v1/admin/migrations/001_create_users?namespace=default"
      );
      expect(error).toBeNull();
      expect(data).toBeDefined();
    });

    it("should get migration with namespace", async () => {
      vi.mocked(mockFetch.get).mockResolvedValue({});

      await migrations.get("001_create_users", "myapp");

      expect(mockFetch.get).toHaveBeenCalledWith(
        "/api/v1/admin/migrations/001_create_users?namespace=myapp"
      );
    });

    it("should handle error", async () => {
      vi.mocked(mockFetch.get).mockRejectedValue(new Error("Not found"));

      const { data, error } = await migrations.get("unknown");

      expect(data).toBeNull();
      expect(error).toBeDefined();
    });
  });

  describe("update()", () => {
    it("should update a migration", async () => {
      const response: Migration = {
        id: "mig-1",
        name: "001_create_users",
        namespace: "default",
        status: "pending",
        description: "Updated description",
        created_at: "2024-01-26T10:00:00Z",
      };

      vi.mocked(mockFetch.put).mockResolvedValue(response);

      const { data, error } = await migrations.update(
        "001_create_users",
        { description: "Updated description" }
      );

      expect(mockFetch.put).toHaveBeenCalledWith(
        "/api/v1/admin/migrations/001_create_users?namespace=default",
        { description: "Updated description" }
      );
      expect(error).toBeNull();
      expect(data).toBeDefined();
    });

    it("should update with namespace", async () => {
      vi.mocked(mockFetch.put).mockResolvedValue({});

      await migrations.update(
        "001_create_users",
        { description: "Updated" },
        "myapp"
      );

      expect(mockFetch.put).toHaveBeenCalledWith(
        "/api/v1/admin/migrations/001_create_users?namespace=myapp",
        { description: "Updated" }
      );
    });

    it("should handle error", async () => {
      vi.mocked(mockFetch.put).mockRejectedValue(new Error("Update failed"));

      const { data, error } = await migrations.update("001_create_users", {});

      expect(data).toBeNull();
      expect(error).toBeDefined();
    });
  });

  describe("delete()", () => {
    it("should delete a migration", async () => {
      vi.mocked(mockFetch.delete).mockResolvedValue({});

      const { data, error } = await migrations.delete("001_create_users");

      expect(mockFetch.delete).toHaveBeenCalledWith(
        "/api/v1/admin/migrations/001_create_users?namespace=default"
      );
      expect(error).toBeNull();
      expect(data).toBeNull();
    });

    it("should delete with namespace", async () => {
      vi.mocked(mockFetch.delete).mockResolvedValue({});

      await migrations.delete("001_create_users", "myapp");

      expect(mockFetch.delete).toHaveBeenCalledWith(
        "/api/v1/admin/migrations/001_create_users?namespace=myapp"
      );
    });

    it("should handle error", async () => {
      vi.mocked(mockFetch.delete).mockRejectedValue(new Error("Cannot delete"));

      const { data, error } = await migrations.delete("001_create_users");

      expect(data).toBeNull();
      expect(error).toBeDefined();
    });
  });

  describe("apply()", () => {
    it("should apply a migration", async () => {
      const response = { message: "Migration applied successfully" };
      vi.mocked(mockFetch.post).mockResolvedValue(response);

      const { data, error } = await migrations.apply("001_create_users");

      expect(mockFetch.post).toHaveBeenCalledWith(
        "/api/v1/admin/migrations/001_create_users/apply",
        { namespace: "default" }
      );
      expect(error).toBeNull();
      expect(data!.message).toBe("Migration applied successfully");
    });

    it("should apply with namespace", async () => {
      vi.mocked(mockFetch.post).mockResolvedValue({ message: "Applied" });

      await migrations.apply("001_create_users", "myapp");

      expect(mockFetch.post).toHaveBeenCalledWith(
        "/api/v1/admin/migrations/001_create_users/apply",
        { namespace: "myapp" }
      );
    });

    it("should handle error", async () => {
      vi.mocked(mockFetch.post).mockRejectedValue(new Error("Apply failed"));

      const { data, error } = await migrations.apply("001_create_users");

      expect(data).toBeNull();
      expect(error).toBeDefined();
    });
  });

  describe("rollback()", () => {
    it("should rollback a migration", async () => {
      const response = { message: "Migration rolled back successfully" };
      vi.mocked(mockFetch.post).mockResolvedValue(response);

      const { data, error } = await migrations.rollback("001_create_users");

      expect(mockFetch.post).toHaveBeenCalledWith(
        "/api/v1/admin/migrations/001_create_users/rollback",
        { namespace: "default" }
      );
      expect(error).toBeNull();
      expect(data!.message).toBe("Migration rolled back successfully");
    });

    it("should rollback with namespace", async () => {
      vi.mocked(mockFetch.post).mockResolvedValue({ message: "Rolled back" });

      await migrations.rollback("001_create_users", "myapp");

      expect(mockFetch.post).toHaveBeenCalledWith(
        "/api/v1/admin/migrations/001_create_users/rollback",
        { namespace: "myapp" }
      );
    });

    it("should handle error", async () => {
      vi.mocked(mockFetch.post).mockRejectedValue(new Error("Rollback failed"));

      const { data, error } = await migrations.rollback("001_create_users");

      expect(data).toBeNull();
      expect(error).toBeDefined();
    });
  });

  describe("applyPending()", () => {
    it("should apply all pending migrations", async () => {
      const response = {
        message: "Applied 2 migrations",
        applied: ["001_create_users", "002_add_posts"],
        failed: [],
      };
      vi.mocked(mockFetch.post).mockResolvedValue(response);

      const { data, error } = await migrations.applyPending();

      expect(mockFetch.post).toHaveBeenCalledWith(
        "/api/v1/admin/migrations/apply-pending",
        { namespace: "default" }
      );
      expect(error).toBeNull();
      expect(data!.applied).toHaveLength(2);
    });

    it("should apply pending with namespace", async () => {
      vi.mocked(mockFetch.post).mockResolvedValue({
        message: "Applied",
        applied: [],
        failed: [],
      });

      await migrations.applyPending("myapp");

      expect(mockFetch.post).toHaveBeenCalledWith(
        "/api/v1/admin/migrations/apply-pending",
        { namespace: "myapp" }
      );
    });

    it("should handle error", async () => {
      vi.mocked(mockFetch.post).mockRejectedValue(new Error("Apply failed"));

      const { data, error } = await migrations.applyPending();

      expect(data).toBeNull();
      expect(error).toBeDefined();
    });
  });

  describe("getExecutions()", () => {
    it("should get execution history", async () => {
      const response: MigrationExecution[] = [
        {
          id: "exec-1",
          migration_name: "001_create_users",
          action: "apply",
          status: "success",
          executed_at: "2024-01-26T10:00:00Z",
          duration_ms: 150,
        },
      ];

      vi.mocked(mockFetch.get).mockResolvedValue(response);

      const { data, error } = await migrations.getExecutions("001_create_users");

      expect(mockFetch.get).toHaveBeenCalledWith(
        "/api/v1/admin/migrations/001_create_users/executions?namespace=default&limit=50"
      );
      expect(error).toBeNull();
      expect(data).toHaveLength(1);
    });

    it("should get executions with custom namespace and limit", async () => {
      vi.mocked(mockFetch.get).mockResolvedValue([]);

      await migrations.getExecutions("001_create_users", "myapp", 10);

      expect(mockFetch.get).toHaveBeenCalledWith(
        "/api/v1/admin/migrations/001_create_users/executions?namespace=myapp&limit=10"
      );
    });

    it("should handle error", async () => {
      vi.mocked(mockFetch.get).mockRejectedValue(new Error("Access denied"));

      const { data, error } = await migrations.getExecutions("001_create_users");

      expect(data).toBeNull();
      expect(error).toBeDefined();
    });
  });
});
