import { rmSync } from 'fs';
import { join } from 'path';

export default async function globalSetup() {
  // Clear Vitest cache before running tests
  const cacheDir = join(process.cwd(), 'node_modules', '.vitest');
  try {
    rmSync(cacheDir, { recursive: true, force: true });
    console.log('Cleared Vitest cache directory');
  } catch (err) {
    // Ignore errors if directory doesn't exist
  }
}
