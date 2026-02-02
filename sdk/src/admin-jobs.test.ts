import { describe, it, expect, beforeEach, vi } from "vitest";
import { FluxbaseAdminJobs } from "./admin-jobs";
import { FluxbaseFetch } from "./fetch";
import * as bundling from "./bundling";
import type {
  JobFunction,
  Job,
  JobStats,
  JobWorker,
  SyncJobsResult,
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

describe("FluxbaseAdminJobs", () => {
  let jobs: FluxbaseAdminJobs;
  let mockFetch: any;

  beforeEach(() => {
    vi.clearAllMocks();
    mockFetch = {
      get: vi.fn(),
      post: vi.fn(),
      put: vi.fn(),
      delete: vi.fn(),
    };
    jobs = new FluxbaseAdminJobs(mockFetch as unknown as FluxbaseFetch);
  });

  describe("create()", () => {
    it("should create a new job function", async () => {
      const response: JobFunction = {
        id: "job-1",
        name: "process-data",
        namespace: "default",
        version: 1,
        enabled: true,
        timeout_seconds: 300,
        created_at: "2024-01-26T10:00:00Z",
        updated_at: "2024-01-26T10:00:00Z",
      };

      vi.mocked(mockFetch.post).mockResolvedValue(response);

      const { data, error } = await jobs.create({
        name: "process-data",
        code: 'export async function handler(req) { return { success: true } }',
        enabled: true,
        timeout_seconds: 300,
      });

      expect(mockFetch.post).toHaveBeenCalledWith("/api/v1/admin/jobs/functions", {
        name: "process-data",
        code: 'export async function handler(req) { return { success: true } }',
        enabled: true,
        timeout_seconds: 300,
      });
      expect(error).toBeNull();
      expect(data).toBeDefined();
      expect(data!.name).toBe("process-data");
    });

    it("should handle create error", async () => {
      vi.mocked(mockFetch.post).mockRejectedValue(new Error("Job already exists"));

      const { data, error } = await jobs.create({
        name: "process-data",
        code: "export async function handler() {}",
      });

      expect(data).toBeNull();
      expect(error).toBeDefined();
    });
  });

  describe("listNamespaces()", () => {
    it("should list all namespaces", async () => {
      const response = { namespaces: ["default", "batch"] };
      vi.mocked(mockFetch.get).mockResolvedValue(response);

      const { data, error } = await jobs.listNamespaces();

      expect(mockFetch.get).toHaveBeenCalledWith("/api/v1/admin/jobs/namespaces");
      expect(error).toBeNull();
      expect(data).toEqual(["default", "batch"]);
    });

    it("should return default namespace when empty", async () => {
      const response = { namespaces: null };
      vi.mocked(mockFetch.get).mockResolvedValue(response);

      const { data, error } = await jobs.listNamespaces();

      expect(error).toBeNull();
      expect(data).toEqual(["default"]);
    });

    it("should handle error", async () => {
      vi.mocked(mockFetch.get).mockRejectedValue(new Error("Network error"));

      const { data, error } = await jobs.listNamespaces();

      expect(data).toBeNull();
      expect(error).toBeDefined();
    });
  });

  describe("list()", () => {
    it("should list all job functions", async () => {
      const response: JobFunction[] = [
        {
          id: "job-1",
          name: "job-a",
          namespace: "default",
          version: 1,
          enabled: true,
          created_at: "2024-01-26T10:00:00Z",
          updated_at: "2024-01-26T10:00:00Z",
        },
      ];

      vi.mocked(mockFetch.get).mockResolvedValue(response);

      const { data, error } = await jobs.list();

      expect(mockFetch.get).toHaveBeenCalledWith("/api/v1/admin/jobs/functions");
      expect(error).toBeNull();
      expect(data).toHaveLength(1);
    });

    it("should list jobs by namespace", async () => {
      vi.mocked(mockFetch.get).mockResolvedValue([]);

      const { data, error } = await jobs.list("batch");

      expect(mockFetch.get).toHaveBeenCalledWith(
        "/api/v1/admin/jobs/functions?namespace=batch"
      );
      expect(error).toBeNull();
    });

    it("should handle error", async () => {
      vi.mocked(mockFetch.get).mockRejectedValue(new Error("Unauthorized"));

      const { data, error } = await jobs.list();

      expect(data).toBeNull();
      expect(error).toBeDefined();
    });
  });

  describe("get()", () => {
    it("should get a specific job function", async () => {
      const response: JobFunction = {
        id: "job-1",
        name: "process-data",
        namespace: "default",
        version: 1,
        enabled: true,
        created_at: "2024-01-26T10:00:00Z",
        updated_at: "2024-01-26T10:00:00Z",
      };

      vi.mocked(mockFetch.get).mockResolvedValue(response);

      const { data, error } = await jobs.get("default", "process-data");

      expect(mockFetch.get).toHaveBeenCalledWith(
        "/api/v1/admin/jobs/functions/default/process-data"
      );
      expect(error).toBeNull();
      expect(data).toBeDefined();
    });

    it("should handle error", async () => {
      vi.mocked(mockFetch.get).mockRejectedValue(new Error("Not found"));

      const { data, error } = await jobs.get("default", "unknown");

      expect(data).toBeNull();
      expect(error).toBeDefined();
    });
  });

  describe("update()", () => {
    it("should update a job function", async () => {
      const response: JobFunction = {
        id: "job-1",
        name: "process-data",
        namespace: "default",
        version: 2,
        enabled: false,
        timeout_seconds: 600,
        created_at: "2024-01-26T10:00:00Z",
        updated_at: "2024-01-26T12:00:00Z",
      };

      vi.mocked(mockFetch.put).mockResolvedValue(response);

      const { data, error } = await jobs.update("default", "process-data", {
        enabled: false,
        timeout_seconds: 600,
      });

      expect(mockFetch.put).toHaveBeenCalledWith(
        "/api/v1/admin/jobs/functions/default/process-data",
        { enabled: false, timeout_seconds: 600 }
      );
      expect(error).toBeNull();
      expect(data!.version).toBe(2);
    });

    it("should handle error", async () => {
      vi.mocked(mockFetch.put).mockRejectedValue(new Error("Update failed"));

      const { data, error } = await jobs.update("default", "process-data", {
        enabled: true,
      });

      expect(data).toBeNull();
      expect(error).toBeDefined();
    });
  });

  describe("delete()", () => {
    it("should delete a job function", async () => {
      vi.mocked(mockFetch.delete).mockResolvedValue({});

      const { data, error } = await jobs.delete("default", "process-data");

      expect(mockFetch.delete).toHaveBeenCalledWith(
        "/api/v1/admin/jobs/functions/default/process-data"
      );
      expect(error).toBeNull();
      expect(data).toBeNull();
    });

    it("should handle error", async () => {
      vi.mocked(mockFetch.delete).mockRejectedValue(new Error("Delete failed"));

      const { data, error } = await jobs.delete("default", "process-data");

      expect(data).toBeNull();
      expect(error).toBeDefined();
    });
  });

  describe("listJobs()", () => {
    it("should list all jobs", async () => {
      const response: Job[] = [
        {
          id: "exec-1",
          job_name: "process-data",
          namespace: "default",
          status: "completed",
          created_at: "2024-01-26T10:00:00Z",
        },
      ];

      vi.mocked(mockFetch.get).mockResolvedValue(response);

      const { data, error } = await jobs.listJobs();

      expect(mockFetch.get).toHaveBeenCalledWith("/api/v1/admin/jobs/queue");
      expect(error).toBeNull();
      expect(data).toHaveLength(1);
    });

    it("should list jobs with filters", async () => {
      vi.mocked(mockFetch.get).mockResolvedValue([]);

      const { data, error } = await jobs.listJobs({
        status: "running",
        namespace: "default",
        limit: 50,
        offset: 10,
        includeResult: true,
      });

      expect(mockFetch.get).toHaveBeenCalledWith(
        "/api/v1/admin/jobs/queue?status=running&namespace=default&limit=50&offset=10&include_result=true"
      );
      expect(error).toBeNull();
    });

    it("should handle error", async () => {
      vi.mocked(mockFetch.get).mockRejectedValue(new Error("Access denied"));

      const { data, error } = await jobs.listJobs();

      expect(data).toBeNull();
      expect(error).toBeDefined();
    });
  });

  describe("getJob()", () => {
    it("should get a specific job", async () => {
      const response: Job = {
        id: "job-uuid",
        job_name: "process-data",
        namespace: "default",
        status: "completed",
        created_at: "2024-01-26T10:00:00Z",
      };

      vi.mocked(mockFetch.get).mockResolvedValue(response);

      const { data, error } = await jobs.getJob("job-uuid");

      expect(mockFetch.get).toHaveBeenCalledWith("/api/v1/admin/jobs/queue/job-uuid");
      expect(error).toBeNull();
      expect(data).toBeDefined();
    });

    it("should handle error", async () => {
      vi.mocked(mockFetch.get).mockRejectedValue(new Error("Not found"));

      const { data, error } = await jobs.getJob("unknown");

      expect(data).toBeNull();
      expect(error).toBeDefined();
    });
  });

  describe("cancel()", () => {
    it("should cancel a job", async () => {
      vi.mocked(mockFetch.post).mockResolvedValue({});

      const { data, error } = await jobs.cancel("job-uuid");

      expect(mockFetch.post).toHaveBeenCalledWith(
        "/api/v1/admin/jobs/queue/job-uuid/cancel",
        {}
      );
      expect(error).toBeNull();
      expect(data).toBeNull();
    });

    it("should handle error", async () => {
      vi.mocked(mockFetch.post).mockRejectedValue(new Error("Cannot cancel"));

      const { data, error } = await jobs.cancel("job-uuid");

      expect(data).toBeNull();
      expect(error).toBeDefined();
    });
  });

  describe("terminate()", () => {
    it("should terminate a job", async () => {
      vi.mocked(mockFetch.post).mockResolvedValue({});

      const { data, error } = await jobs.terminate("job-uuid");

      expect(mockFetch.post).toHaveBeenCalledWith(
        "/api/v1/admin/jobs/queue/job-uuid/terminate",
        {}
      );
      expect(error).toBeNull();
      expect(data).toBeNull();
    });

    it("should handle error", async () => {
      vi.mocked(mockFetch.post).mockRejectedValue(new Error("Cannot terminate"));

      const { data, error } = await jobs.terminate("job-uuid");

      expect(data).toBeNull();
      expect(error).toBeDefined();
    });
  });

  describe("retry()", () => {
    it("should retry a job", async () => {
      const response: Job = {
        id: "new-job-uuid",
        job_name: "process-data",
        namespace: "default",
        status: "pending",
        created_at: "2024-01-26T12:00:00Z",
      };

      vi.mocked(mockFetch.post).mockResolvedValue(response);

      const { data, error } = await jobs.retry("job-uuid");

      expect(mockFetch.post).toHaveBeenCalledWith(
        "/api/v1/admin/jobs/queue/job-uuid/retry",
        {}
      );
      expect(error).toBeNull();
      expect(data).toBeDefined();
    });

    it("should handle error", async () => {
      vi.mocked(mockFetch.post).mockRejectedValue(new Error("Cannot retry"));

      const { data, error } = await jobs.retry("job-uuid");

      expect(data).toBeNull();
      expect(error).toBeDefined();
    });
  });

  describe("getStats()", () => {
    it("should get job statistics", async () => {
      const response: JobStats = {
        pending: 5,
        running: 2,
        completed: 100,
        failed: 3,
        cancelled: 1,
      };

      vi.mocked(mockFetch.get).mockResolvedValue(response);

      const { data, error } = await jobs.getStats();

      expect(mockFetch.get).toHaveBeenCalledWith("/api/v1/admin/jobs/stats");
      expect(error).toBeNull();
      expect(data).toBeDefined();
      expect(data!.pending).toBe(5);
    });

    it("should get stats by namespace", async () => {
      vi.mocked(mockFetch.get).mockResolvedValue({});

      const { data, error } = await jobs.getStats("batch");

      expect(mockFetch.get).toHaveBeenCalledWith(
        "/api/v1/admin/jobs/stats?namespace=batch"
      );
      expect(error).toBeNull();
    });

    it("should handle error", async () => {
      vi.mocked(mockFetch.get).mockRejectedValue(new Error("Access denied"));

      const { data, error } = await jobs.getStats();

      expect(data).toBeNull();
      expect(error).toBeDefined();
    });
  });

  describe("listWorkers()", () => {
    it("should list active workers", async () => {
      const response: JobWorker[] = [
        {
          id: "worker-1",
          hostname: "host-1",
          current_jobs: 2,
          max_concurrent: 5,
          started_at: "2024-01-26T08:00:00Z",
        },
      ];

      vi.mocked(mockFetch.get).mockResolvedValue(response);

      const { data, error } = await jobs.listWorkers();

      expect(mockFetch.get).toHaveBeenCalledWith("/api/v1/admin/jobs/workers");
      expect(error).toBeNull();
      expect(data).toHaveLength(1);
    });

    it("should handle error", async () => {
      vi.mocked(mockFetch.get).mockRejectedValue(new Error("Access denied"));

      const { data, error } = await jobs.listWorkers();

      expect(data).toBeNull();
      expect(error).toBeDefined();
    });
  });

  describe("sync()", () => {
    it("should sync jobs with options object", async () => {
      const response: SyncJobsResult = {
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
          created: ["new-job"],
          updated: [],
          deleted: [],
          unchanged: [],
          errors: [],
        },
        dry_run: false,
      };

      vi.mocked(mockFetch.post).mockResolvedValue(response);

      const { data, error } = await jobs.sync({
        namespace: "default",
        functions: [{ name: "new-job", code: "export function handler() {}" }],
        options: { delete_missing: true, dry_run: false },
      });

      expect(mockFetch.post).toHaveBeenCalledWith("/api/v1/admin/jobs/sync", {
        namespace: "default",
        jobs: [{ name: "new-job", code: "export function handler() {}" }],
        options: { delete_missing: true, dry_run: false },
      });
      expect(error).toBeNull();
      expect(data).toBeDefined();
    });

    it("should sync with legacy string namespace", async () => {
      const response: SyncJobsResult = {
        message: "Sync completed",
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

      vi.mocked(mockFetch.post).mockResolvedValue(response);

      const { data, error } = await jobs.sync("default");

      expect(mockFetch.post).toHaveBeenCalledWith("/api/v1/admin/jobs/sync", {
        namespace: "default",
        jobs: undefined,
        options: { delete_missing: false, dry_run: false },
      });
      expect(error).toBeNull();
    });

    it("should handle error", async () => {
      vi.mocked(mockFetch.post).mockRejectedValue(new Error("Sync failed"));

      const { data, error } = await jobs.sync({ namespace: "default" });

      expect(data).toBeNull();
      expect(error).toBeDefined();
    });
  });

  describe("syncWithBundling()", () => {
    it("should sync with empty functions array", async () => {
      const syncResponse: SyncJobsResult = {
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

      const { data, error } = await jobs.syncWithBundling({
        namespace: "default",
        functions: [],
      });

      expect(error).toBeNull();
    });

    it("should sync without functions", async () => {
      const syncResponse: SyncJobsResult = {
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
          created: ["job-1"],
          updated: [],
          deleted: [],
          unchanged: [],
          errors: [],
        },
        dry_run: false,
      };

      vi.mocked(mockFetch.post).mockResolvedValue(syncResponse);

      const { data, error } = await jobs.syncWithBundling({
        namespace: "default",
      });

      expect(error).toBeNull();
    });

    it("should return error when esbuild is not available", async () => {
      vi.mocked(bundling.loadEsbuild).mockResolvedValue(false);

      const { data, error } = await jobs.syncWithBundling({
        namespace: "default",
        functions: [{ name: "job-1", code: 'import { foo } from "./bar"' }],
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

      const syncResponse: SyncJobsResult = {
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
          created: ["job-1"],
          updated: [],
          deleted: [],
          unchanged: [],
          errors: [],
        },
        dry_run: false,
      };

      vi.mocked(mockFetch.post).mockResolvedValue(syncResponse);

      const { data, error } = await jobs.syncWithBundling({
        namespace: "default",
        functions: [{ name: "job-1", code: 'import { foo } from "./bar"' }],
      });

      expect(bundling.bundleCode).toHaveBeenCalled();
      expect(error).toBeNull();
    });

    it("should skip bundling for pre-bundled functions", async () => {
      vi.mocked(bundling.loadEsbuild).mockResolvedValue(true);

      const syncResponse: SyncJobsResult = {
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
          created: ["job-1"],
          updated: [],
          deleted: [],
          unchanged: [],
          errors: [],
        },
        dry_run: false,
      };

      vi.mocked(mockFetch.post).mockResolvedValue(syncResponse);

      const { data, error } = await jobs.syncWithBundling({
        namespace: "default",
        functions: [
          { name: "job-1", code: "// already bundled", is_pre_bundled: true },
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

      const { data, error } = await jobs.syncWithBundling({
        namespace: "default",
        functions: [{ name: "job-1", code: "invalid code {{{" }],
      });

      expect(data).toBeNull();
      expect(error).toBeDefined();
    });

    it("should use per-function sourceDir and nodePaths", async () => {
      vi.mocked(bundling.loadEsbuild).mockResolvedValue(true);
      vi.mocked(bundling.bundleCode).mockResolvedValue({
        code: "// bundled",
        map: undefined,
      });

      const syncResponse: SyncJobsResult = {
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
          created: ["job-1"],
          updated: [],
          deleted: [],
          unchanged: [],
          errors: [],
        },
        dry_run: false,
      };

      vi.mocked(mockFetch.post).mockResolvedValue(syncResponse);

      await jobs.syncWithBundling(
        {
          namespace: "default",
          functions: [
            {
              name: "job-1",
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

      const result = await FluxbaseAdminJobs.bundleCode({
        code: 'export async function handler() { return "hello" }',
      });

      expect(bundling.bundleCode).toHaveBeenCalled();
      expect(result.code).toBe("// bundled");
    });
  });
});
