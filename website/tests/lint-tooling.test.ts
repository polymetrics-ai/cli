import { existsSync, readFileSync } from 'node:fs';
import path from 'node:path';

import { describe, expect, it } from 'vitest';

const rootDir = process.cwd();
const repoDir = path.resolve(rootDir, '..');

function readWebsiteFile(relativePath: string): string {
  return readFileSync(path.join(rootDir, relativePath), 'utf8');
}

function readRepoFile(relativePath: string): string {
  return readFileSync(path.join(repoDir, relativePath), 'utf8');
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

  it('pins transitive brace expansion to the dependency-review-patched release', () => {
    const lockfile = readWebsiteFile('pnpm-lock.yaml');
    const braceExpansionEntries = [
      ...new Set([...lockfile.matchAll(/^  brace-expansion@(.+):$/gm)].map((match) => match[1])),
    ].sort();

    expect(braceExpansionEntries).toEqual(['5.0.8']);
    expect(lockfile).not.toContain('brace-expansion@1.1.16');
  });

  it('runs the restored lint command in website CI', () => {
    const githubWorkflow = readRepoFile('.github/workflows/website.yml');
    const gitlabWorkflow = readRepoFile('.gitlab-ci.yml');

    expect(githubWorkflow).toContain('run: pnpm run lint');
    expect(gitlabWorkflow).toContain('- pnpm run lint');
  });

  it('includes pnpm patches in the Docker dependency layer', () => {
    const workspaceConfig = readWebsiteFile('pnpm-workspace.yaml');
    const dockerfile = readWebsiteFile('Dockerfile');

    expect(workspaceConfig).toContain('patchedDependencies:');
    expect(dockerfile).toContain('COPY patches ./patches');
  });
});
