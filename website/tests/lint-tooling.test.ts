import { spawnSync } from 'node:child_process';
import { existsSync, readFileSync } from 'node:fs';
import path from 'node:path';

import { ESLint } from 'eslint';
import { describe, expect, it } from 'vitest';

const rootDir = process.cwd();
const repoDir = path.resolve(rootDir, '..');

function readWebsiteFile(relativePath: string): string {
  return readFileSync(path.join(rootDir, relativePath), 'utf8');
}

function readRepoFile(relativePath: string): string {
  return readFileSync(path.join(repoDir, relativePath), 'utf8');
}

type PackageManifest = {
  dependencies?: Record<string, string>;
  devDependencies?: Record<string, string>;
  scripts?: Record<string, string>;
};

type NpmLockfile = {
  lockfileVersion?: number;
  packages?: Record<
    string,
    {
      dependencies?: Record<string, string>;
      devDependencies?: Record<string, string>;
    }
  >;
};

function readPackageManifest(): PackageManifest {
  return JSON.parse(readWebsiteFile('package.json')) as PackageManifest;
}

function readNpmLockfile(): NpmLockfile {
  return JSON.parse(readWebsiteFile('package-lock.json')) as NpmLockfile;
}

describe('lint tooling', () => {
  it('uses the ESLint CLI instead of the removed Next.js lint command', () => {
    const packageJson = readPackageManifest();

    expect(packageJson.scripts?.lint).toBe('eslint .');
    expect(packageJson.scripts?.lint).not.toContain('next lint');
  });

  it('keeps npm ci direct dependency records synchronized with the manifest', () => {
    const packageJson = readPackageManifest();
    const packageLock = readNpmLockfile();
    const lockRoot = packageLock.packages?.[''];
    const directDependencies = {
      ...packageJson.dependencies,
      ...packageJson.devDependencies,
    };

    expect(packageLock.lockfileVersion).toBe(3);
    expect(lockRoot?.dependencies).toEqual(packageJson.dependencies);
    expect(lockRoot?.devDependencies).toEqual(packageJson.devDependencies);

    for (const dependencyName of Object.keys(directDependencies)) {
      expect(packageLock.packages?.[`node_modules/${dependencyName}`]).toBeDefined();
    }
  });

  it('rejects an error-level TypeScript violation through the configured ESLint rules', async () => {
    const eslint = new ESLint({ cwd: rootDir });
    const [result] = await eslint.lintText('const lintProof: any = "must fail";\n', {
      filePath: 'lint-failure-proof.tsx',
    });

    expect(result.errorCount).toBeGreaterThan(0);
    expect(result.messages).toEqual(
      expect.arrayContaining([
        expect.objectContaining({
          ruleId: '@typescript-eslint/no-explicit-any',
          severity: 2,
        }),
      ]),
    );
  });

  it('returns a failure status from the ESLint CLI for an error', () => {
    const result = spawnSync(
      process.execPath,
      [
        path.join(rootDir, 'node_modules', 'eslint', 'bin', 'eslint.js'),
        '--no-config-lookup',
        '--rule',
        'no-undef:error',
        '--stdin',
        '--stdin-filename',
        'lint-failure-proof.js',
      ],
      {
        cwd: rootDir,
        encoding: 'utf8',
        input: 'lintProof;\n',
      },
    );

    expect(result.error).toBeUndefined();
    expect(result.status).toBe(1);
    expect(result.stdout).toContain('no-undef');
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
