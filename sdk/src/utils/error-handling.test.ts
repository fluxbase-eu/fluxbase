import { describe, it, expect } from "vitest";
import { wrapAsync, wrapAsyncVoid, wrapSync } from "./error-handling";

describe("error-handling utilities", () => {
  describe("wrapAsync()", () => {
    it("should return data on success", async () => {
      const result = await wrapAsync(async () => {
        return { message: "success" };
      });

      expect(result.data).toEqual({ message: "success" });
      expect(result.error).toBeNull();
    });

    it("should return primitive data types", async () => {
      const numberResult = await wrapAsync(async () => 42);
      expect(numberResult.data).toBe(42);
      expect(numberResult.error).toBeNull();

      const stringResult = await wrapAsync(async () => "hello");
      expect(stringResult.data).toBe("hello");
      expect(stringResult.error).toBeNull();

      const boolResult = await wrapAsync(async () => true);
      expect(boolResult.data).toBe(true);
      expect(boolResult.error).toBeNull();
    });

    it("should return null data on Error", async () => {
      const error = new Error("Something went wrong");
      const result = await wrapAsync(async () => {
        throw error;
      });

      expect(result.data).toBeNull();
      expect(result.error).toBe(error);
    });

    it("should convert non-Error to Error", async () => {
      const result = await wrapAsync(async () => {
        throw "string error";
      });

      expect(result.data).toBeNull();
      expect(result.error).toBeInstanceOf(Error);
      expect(result.error!.message).toBe("string error");
    });

    it("should handle numbers thrown as errors", async () => {
      const result = await wrapAsync(async () => {
        throw 404;
      });

      expect(result.data).toBeNull();
      expect(result.error).toBeInstanceOf(Error);
      expect(result.error!.message).toBe("404");
    });

    it("should handle objects thrown as errors", async () => {
      const result = await wrapAsync(async () => {
        throw { code: "ERR_001", message: "Custom error" };
      });

      expect(result.data).toBeNull();
      expect(result.error).toBeInstanceOf(Error);
    });

    it("should handle async operations with delay", async () => {
      const result = await wrapAsync(async () => {
        await new Promise((resolve) => setTimeout(resolve, 10));
        return "delayed result";
      });

      expect(result.data).toBe("delayed result");
      expect(result.error).toBeNull();
    });

    it("should handle arrays", async () => {
      const result = await wrapAsync(async () => {
        return [1, 2, 3];
      });

      expect(result.data).toEqual([1, 2, 3]);
      expect(result.error).toBeNull();
    });

    it("should handle null return value", async () => {
      const result = await wrapAsync(async () => {
        return null;
      });

      expect(result.data).toBeNull();
      expect(result.error).toBeNull();
    });

    it("should handle undefined return value", async () => {
      const result = await wrapAsync(async () => {
        return undefined;
      });

      expect(result.data).toBeUndefined();
      expect(result.error).toBeNull();
    });
  });

  describe("wrapAsyncVoid()", () => {
    it("should return null error on success", async () => {
      let sideEffect = false;
      const result = await wrapAsyncVoid(async () => {
        sideEffect = true;
      });

      expect(result.error).toBeNull();
      expect(sideEffect).toBe(true);
    });

    it("should return error on failure", async () => {
      const error = new Error("Void operation failed");
      const result = await wrapAsyncVoid(async () => {
        throw error;
      });

      expect(result.error).toBe(error);
    });

    it("should convert non-Error to Error", async () => {
      const result = await wrapAsyncVoid(async () => {
        throw "string error in void";
      });

      expect(result.error).toBeInstanceOf(Error);
      expect(result.error!.message).toBe("string error in void");
    });

    it("should handle async operations with delay", async () => {
      let completed = false;
      const result = await wrapAsyncVoid(async () => {
        await new Promise((resolve) => setTimeout(resolve, 10));
        completed = true;
      });

      expect(result.error).toBeNull();
      expect(completed).toBe(true);
    });

    it("should handle delayed error", async () => {
      const result = await wrapAsyncVoid(async () => {
        await new Promise((resolve) => setTimeout(resolve, 10));
        throw new Error("Delayed error");
      });

      expect(result.error).toBeDefined();
      expect(result.error!.message).toBe("Delayed error");
    });
  });

  describe("wrapSync()", () => {
    it("should return data on success", () => {
      const result = wrapSync(() => {
        return { value: 42 };
      });

      expect(result.data).toEqual({ value: 42 });
      expect(result.error).toBeNull();
    });

    it("should return primitive data types", () => {
      const numberResult = wrapSync(() => 100);
      expect(numberResult.data).toBe(100);
      expect(numberResult.error).toBeNull();

      const stringResult = wrapSync(() => "sync hello");
      expect(stringResult.data).toBe("sync hello");
      expect(stringResult.error).toBeNull();

      const boolResult = wrapSync(() => false);
      expect(boolResult.data).toBe(false);
      expect(boolResult.error).toBeNull();
    });

    it("should return null data on Error", () => {
      const error = new Error("Sync error");
      const result = wrapSync(() => {
        throw error;
      });

      expect(result.data).toBeNull();
      expect(result.error).toBe(error);
    });

    it("should convert non-Error to Error", () => {
      const result = wrapSync(() => {
        throw "sync string error";
      });

      expect(result.data).toBeNull();
      expect(result.error).toBeInstanceOf(Error);
      expect(result.error!.message).toBe("sync string error");
    });

    it("should handle numbers thrown as errors", () => {
      const result = wrapSync(() => {
        throw 500;
      });

      expect(result.data).toBeNull();
      expect(result.error).toBeInstanceOf(Error);
      expect(result.error!.message).toBe("500");
    });

    it("should handle arrays", () => {
      const result = wrapSync(() => ["a", "b", "c"]);

      expect(result.data).toEqual(["a", "b", "c"]);
      expect(result.error).toBeNull();
    });

    it("should handle null return value", () => {
      const result = wrapSync(() => null);

      expect(result.data).toBeNull();
      expect(result.error).toBeNull();
    });

    it("should handle undefined return value", () => {
      const result = wrapSync(() => undefined);

      expect(result.data).toBeUndefined();
      expect(result.error).toBeNull();
    });

    it("should handle complex objects", () => {
      const complexObject = {
        nested: {
          deeply: {
            value: "found",
          },
        },
        array: [1, 2, 3],
        date: new Date("2024-01-26"),
      };

      const result = wrapSync(() => complexObject);

      expect(result.data).toEqual(complexObject);
      expect(result.error).toBeNull();
    });
  });

  describe("type safety", () => {
    it("should preserve types correctly for wrapAsync", async () => {
      interface User {
        id: string;
        name: string;
      }

      const result = await wrapAsync<User>(async () => {
        return { id: "1", name: "Test User" };
      });

      // Type should be inferred correctly
      if (result.data) {
        expect(result.data.id).toBe("1");
        expect(result.data.name).toBe("Test User");
      }
    });

    it("should preserve types correctly for wrapSync", () => {
      interface Config {
        setting: boolean;
        value: number;
      }

      const result = wrapSync<Config>(() => {
        return { setting: true, value: 42 };
      });

      if (result.data) {
        expect(result.data.setting).toBe(true);
        expect(result.data.value).toBe(42);
      }
    });
  });
});
