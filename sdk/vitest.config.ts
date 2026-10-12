import { defineConfig } from 'vitest/config';
import { rmSync } from 'fs';
import { join } from 'path';

// Clear Vitest cache when config loads
try {
  const cacheDir = join(process.cwd(), 'node_modules', '.vitest');
  rmSync(cacheDir, { recursive: true, force: true });
  console.log('Cleared Vitest cache directory');
} catch (err) {
  // Ignore if cache doesn't exist
}

export default defineConfig({
  test: {
    cache: false,
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html', 'lcov'],
      reportsDirectory: './coverage',
      include: ['src/**/*.ts'],
      // types-generated.ts is machine-generated from the served OpenAPI spec
      // (scripts/check-contract-drift.sh) — no runtime code, nothing imports
      // it, so it must not drag the coverage gate down.
      exclude: ['src/**/*.test.ts', 'src/**/*.d.ts', 'src/examples/**', 'src/types-generated.ts'],
      thresholds: {
        statements: 60,
        branches: 60,
        functions: 60,
        lines: 60
      }
    }
  }
});
