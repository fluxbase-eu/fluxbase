import { fileURLToPath } from "node:url";
import { defineConfig } from "vitest/config";

// Unit-test config for the admin dashboard. Deliberately separate from
// vite.config.ts: the app config wires the TanStack Router plugin (route
// scanning) and SWC React plugin, none of which the unit tests need — these
// tests cover plain TS modules (stores, lib helpers) with no component
// rendering. E2E browser tests remain on Playwright (playwright.config.ts).
//
// Coverage is scoped to the files that HAVE unit tests so the 40% thresholds
// are meaningful from day one (issue #408: admin had zero unit tests); widen
// the include list as more tests land instead of lowering the bar.
export default defineConfig({
  resolve: {
    alias: {
      "@": fileURLToPath(new URL("./src", import.meta.url)),
      // Same source-level aliasing the app tsconfig uses: unit tests must not
      // depend on the SDK packages being built to dist first.
      "@nimbleflux/fluxbase-sdk": fileURLToPath(
        new URL("../sdk/src/index.ts", import.meta.url),
      ),
      "@nimbleflux/fluxbase-sdk-react": fileURLToPath(
        new URL("../sdk-react/src/index.ts", import.meta.url),
      ),
    },
  },
  test: {
    environment: "jsdom",
    setupFiles: ["./src/test/setup.ts"],
    // Unit tests live next to the code they cover. Playwright browser tests
    // (admin/tests/*.spec.ts) have their own runner and must not be picked up.
    include: ["src/**/*.test.{ts,tsx}"],
    coverage: {
      provider: "v8",
      reporter: ["text", "html"],
      reportsDirectory: "./coverage",
      include: [
        "src/lib/api/client.ts",
        "src/lib/auth.ts",
        "src/stores/auth-store.ts",
      ],
      thresholds: {
        statements: 40,
        branches: 40,
        functions: 40,
        lines: 40,
      },
    },
  },
});
