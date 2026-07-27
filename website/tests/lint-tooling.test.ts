import { existsSync, readFileSync } from 'node:fs';
import path from 'node:path';

import { describe, expect, it } from 'vitest';

const rootDir = process.cwd();

function readWebsiteFile(relativePath: string): string {
  return readFileSync(path.join(rootDir, relativePath), 'utf8');
}

describe('lint tooling', () => {
  it('uses the ESLint CLI instead of the removed Next.js lint command', () => {
    const packageJson = JSON.parse(readWebsiteFile('package.json')) as {
      scripts?: Record<string, string>;
    };

    expect(packageJson.scripts?.lint).toBe('eslint .');
    expect(packageJson.scripts?.lint).not.toContain('next lint');
  });

  it('keeps generated and build artifacts out of the lint scope', () => {
    const configPath = path.join(rootDir, 'eslint.config.mjs');

    expect(existsSync(configPath)).toBe(true);

    const config = readFileSync(configPath, 'utf8');

    expect(config).toContain('eslint-config-next/core-web-vitals');
    expect(config).toContain('eslint-config-next/typescript');

    for (const ignoredPattern of [
      '.next/**',
      'out/**',
      'build/**',
      'next-env.d.ts',
      '.source/**',
      'lib/**/*.generated.*',
      'data/**/*.generated.*',
      'public/connectors/icons/**',
      'playwright-report/**',
      'test-results/**',
      'coverage/**',
    ]) {
      expect(config).toContain(ignoredPattern);
    }
  });
});
