import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    cache: false,
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
