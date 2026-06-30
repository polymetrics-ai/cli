import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { defineConfig } from 'vitest/config';

const rootDir = fileURLToPath(new URL('.', import.meta.url));

export default defineConfig({
  resolve: {
    alias: {
      '@': rootDir,
    },
  },
  test: {
    environment: 'node',
    exclude: ['node_modules/**', '.next/**', 'tests/e2e/**'],
    include: ['tests/**/*.test.ts'],
    root: path.resolve(rootDir),
  },
});
