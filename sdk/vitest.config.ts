import { defineConfig } from 'vitest/config';
import { fileURLToPath } from 'url';
import { dirname, resolve } from 'path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

export default defineConfig({
  test: {
    cache: false,
    globalSetup: resolve(__dirname, 'vitest.setup.ts'),
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html', 'lcov'],
      reportsDirectory: './coverage',
      include: ['src/**/*.ts'],
      exclude: ['src/**/*.test.ts', 'src/**/*.d.ts', 'src/examples/**'],
      thresholds: {
        statements: 80,
        branches: 80,
        functions: 60,
        lines: 80
      }
    }
  }
});
