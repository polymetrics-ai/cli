# TDD Ledger — website lint script restoration

## Red

- Added focused Vitest coverage that asserts `website/package.json` uses the ESLint CLI (`eslint .`) instead of obsolete `next lint`.
- Added coverage that asserts a flat ESLint config exists and records generated/build artifact ignores.

Evidence:
- `corepack pnpm@11.7.0 exec vitest run tests/lint-tooling.test.ts` failed before implementation:
  - expected lint script `eslint .`, received `next lint`;
  - expected `eslint.config.mjs` to exist, received missing config.

## Green

- Added `eslint` and `eslint-config-next` dev dependencies locked by the repository's pnpm workflow.
- Replaced `website/package.json` lint script with `eslint .`.
- Added `website/eslint.config.mjs` using `eslint-config-next/core-web-vitals` and `eslint-config-next/typescript` flat configs.
- Added explicit ignores for Next/build outputs, Fumadocs/generated website artifacts, connector icons, and local report directories.
- Kept pre-existing React Hooks compiler findings as warnings only for the known existing files so the restored lint command is runnable without touching out-of-scope product code or weakening future files.

Evidence:
- `corepack pnpm@11.7.0 exec vitest run tests/lint-tooling.test.ts` passed.
- `corepack pnpm@11.7.0 run lint` passed with 18 warnings and 0 errors.

## Refactor

- Set `pnpm-workspace.yaml` `allowBuilds.unrs-resolver: false` because the new ESLint resolver dependency was auto-added as a build-script decision; clean frozen pnpm install and ESLint both pass without allowing that build script.
- Left the pre-existing npm lockfile untouched because website CI/Docker use pnpm and npm lockfile repair would pull in unrelated pre-existing website dependency drift.
- Scoped existing React Hooks lint warnings to the current debt files rather than globally weakening future files.

## Skills

`gsd-core`, `vercel-react-best-practices`, `context-mode`.
