# TDD Ledger — website lint script restoration

## Red

- Added focused Vitest coverage that asserts `website/package.json` uses the ESLint CLI (`eslint .`) instead of obsolete `next lint`.
- Added coverage that asserts a flat ESLint config exists and records generated/build artifact ignores.
- Added focused lockfile coverage that asserts `brace-expansion` is pinned to the Dependency Review patched release and the vulnerable `1.1.16` entry is absent.
- Added focused Dockerfile coverage that asserts pnpm patch files are copied into the Docker dependency layer when `patchedDependencies` is configured.

Evidence:
- `corepack pnpm@11.7.0 exec vitest run tests/lint-tooling.test.ts` failed before implementation:
  - expected lint script `eslint .`, received `next lint`;
  - expected `eslint.config.mjs` to exist, received missing config.
- `CI=true COREPACK_HOME="$PWD/.corepack" corepack pnpm@11.7.0 --dir website exec vitest run tests/lint-tooling.test.ts --reporter=verbose` failed before the Dependency Review fix:
  - expected `["5.0.8"]`, received `["1.1.16", "5.0.8"]`.
- `CI=true COREPACK_HOME="$PWD/.corepack" corepack pnpm@11.7.0 --dir website exec vitest run tests/lint-tooling.test.ts --reporter=verbose` failed before the Dockerfile fix:
  - expected `Dockerfile` to contain `COPY patches ./patches`.

## Green

- Added `eslint` and `eslint-config-next` dev dependencies locked by the repository's pnpm workflow.
- Replaced `website/package.json` lint script with `eslint .`.
- Added `website/eslint.config.mjs` using `eslint-config-next/core-web-vitals` and `eslint-config-next/typescript` flat configs.
- Added explicit ignores for Next/build outputs, Fumadocs/generated website artifacts, connector icons, and local report directories.
- Kept pre-existing React Hooks compiler findings as warnings only for the known existing files so the restored lint command is runnable without touching out-of-scope product code or weakening future files.
- Added `pnpm-workspace.yaml` override `brace-expansion: 5.0.8` because GitHub advisory GHSA-mh99-v99m-4gvg marks `<=5.0.7` affected and `5.0.8` patched.
- Added `website/patches/minimatch@3.1.5.patch` so legacy CommonJS `minimatch@3.1.5` can call the patched `brace-expansion@5.0.8` `{ expand }` export.
- Updated `website/Dockerfile` to copy `patches/` into the deps stage before `pnpm install --frozen-lockfile`.

Evidence:
- `corepack pnpm@11.7.0 exec vitest run tests/lint-tooling.test.ts` passed.
- `corepack pnpm@11.7.0 run lint` passed with 13 warnings and 0 errors.
- `CI=true COREPACK_HOME="$PWD/.corepack" corepack pnpm@11.7.0 --dir website install --frozen-lockfile --ignore-scripts --store-dir ../.pnpm-store --reporter=append-only` passed.
- `CI=true COREPACK_HOME="$PWD/.corepack" corepack pnpm@11.7.0 --dir website exec vitest run tests/lint-tooling.test.ts --reporter=verbose` passed.
- `CI=true COREPACK_HOME="$PWD/.corepack" corepack pnpm@11.7.0 --dir website run lint` passed with 13 warnings and 0 errors.
- `docker build --target deps -f website/Dockerfile -t polymetrics-website-deps:ci-fix website` passed; temporary local image tag was removed.

## Refactor

- Set `pnpm-workspace.yaml` `allowBuilds.unrs-resolver: false` because the new ESLint resolver dependency was auto-added as a build-script decision; clean frozen pnpm install and ESLint both pass without allowing that build script.
- Left the pre-existing npm lockfile untouched because website CI/Docker use pnpm and npm lockfile repair would pull in unrelated pre-existing website dependency drift.
- Scoped existing React Hooks lint warnings to the current debt files rather than globally weakening future files.
- Removed safe pre-existing non-product lint warnings from website generator/config files.
- Kept `.github/workflows/pr-issue-guard.yml` and `internal/coordination/issueguard` unchanged; the fix for `require-linked-issue` is to update the live PR body with `Refs #67`.

## Skills

`gsd-core`, `gsd-programming-loop`, `no-mistakes`, `javascript-testing-patterns`, `vercel-react-best-practices`, `context-mode`.
