# Plan: website lint script restoration

## GSD command path

- `scripts/gsd doctor`: passed on this worktree.
- `scripts/gsd prompt plan-phase website-lint-script --skip-research --tdd`: generated `.planning/traces/gsd-plan-phase-website-lint-script-prompt.md`.
- `scripts/gsd prompt programming-loop init --phase website-lint-script --dry-run` and `scripts/gsd prompt gsd-programming-loop init --phase website-lint-script --dry-run`: unavailable (`unknown GSD command: programming-loop`); manual GSD/TDD fallback active.

## Required skills loaded

- `gsd-core`
- `gsd-programming-loop`
- `no-mistakes`
- `javascript-testing-patterns`
- `vercel-react-best-practices`
- `context-mode`
- `frontend-design` and `web-design-guidelines` were referenced by routing for website/UI work but are not installed in this Pi skill registry; no UI/visual code is in scope.

## Objective

Restore a deterministic runnable website lint command after Next.js 16 removed `next lint`.

## Scope

- Reproduce current `pnpm run lint` failure from `origin/main` with installed website dependencies.
- Replace `website/package.json` lint script with the smallest supported ESLint CLI invocation.
- Add only lint dependencies/config required by Next.js 16 and ESLint flat config.
- Preserve lint scope for authored website source while ignoring generated/build artifacts.
- Add focused regression coverage for the lint script/config contract.
- Keep added lint-tooling transitive dependencies clear of Dependency Review high-severity advisories.
- Run website lint, typecheck, unit, e2e, and build checks for changed website files.

## Non-goals / safety gates

- No broad dependency upgrades.
- No PM binary-release config/docs, release-triggered website behavior, profile settings product code, connectors, or issue guard changes.
- No credentialed connector checks, secrets, deploys, merges, or Herdr lifecycle commands.

## Implementation slices

1. Red: add a focused Vitest regression for the website lint command/config contract and verify it fails against the current `next lint` script.
2. Green: add the minimal Next.js 16 ESLint CLI setup (`eslint .`, `eslint.config.mjs`, required lint dev dependencies) with generated/build ignores.
3. Red: add a focused lockfile regression for the `brace-expansion` Dependency Review finding and verify it fails while `brace-expansion@1.1.16` is present.
4. Green: override `brace-expansion` to the patched `5.0.8` release and patch `minimatch@3.1.5` to consume the patched CommonJS `{ expand }` export.
5. Red: add a Dockerfile regression proving pnpm patch files are copied before the deps-stage frozen install.
6. Green: copy `website/patches/` into the Docker deps layer before `pnpm install`.
7. Verification: run targeted lint regression plus website lint/typecheck/unit/e2e/build checks.

## Implementation result

- Replaced `next lint` with `eslint .`.
- Added the minimal Next.js 16 ESLint CLI dependencies and flat config (`eslint`, `eslint-config-next`, `eslint.config.mjs`).
- Included Next/TypeScript lint configs and explicit generated/build/report ignores.
- Scoped pre-existing React Hooks compiler findings to warnings for known existing files so the restored command is runnable without touching out-of-scope product code.
- Added focused Vitest regression coverage for the lint script/config contract.
- Added a pnpm override for `brace-expansion@5.0.8`, removed the vulnerable `brace-expansion@1.1.16` lockfile entry, and added a minimal pnpm patch so `minimatch@3.1.5` remains compatible with the patched export shape.
- Updated `website/Dockerfile` so the image deps layer copies pnpm patch files before the frozen install.
- Kept the issue guard source unchanged; the live PR body needs an incremental issue link (`Refs #67`) instead.

## Reproduction evidence

- On `origin/main`/branch start (`902fd9403`), after `corepack pnpm@11.7.0 install --frozen-lockfile --reporter=silent`, `corepack pnpm@11.7.0 run lint` failed with `Invalid project directory provided, no such directory: .../website/lint` and exit 1.

## Verification result

- Website focused regression, lint, typecheck, unit, e2e, build, clean frozen install, npm lock dry-run, and `git diff --check` passed locally. See `VERIFICATION.md`.
- Follow-up CI-fix verification passed locally for the focused regression, frozen pnpm install, and full website lint after the `brace-expansion` override and `minimatch` patch. See `VERIFICATION.md`.
- Follow-up image verification passed locally for the Docker deps stage after copying `patches/`. See `VERIFICATION.md`.

## Continuation — 2026-07-30 current-main refresh

- Revalidated isolation and checked out the canonical PR #542 branch `fix/website-lint-script`; merged current `origin/main@c85740b6f` with a non-rewriting merge to preserve all prior PR commits and pipeline commits.
- Rechecked official guidance: Next.js 16 removes `next lint` and says to use ESLint directly or the `next-lint-to-eslint-cli` codemod (https://nextjs.org/docs/app/guides/upgrading/version-16#next-lint-command); current Next.js ESLint setup installs `eslint` and `eslint-config-next`, uses `eslint.config.mjs`, and runs `pnpm exec eslint .` (https://nextjs.org/docs/app/api-reference/config/eslint). ESLint current docs use native flat config files (`eslint.config.js|mjs|cjs`) with `defineConfig`/`globalIgnores` for global ignore patterns (https://eslint.org/docs/latest/use/configure/configuration-files and https://eslint.org/docs/latest/use/configure/ignore).
- Version rationale: website CI pins Node 22 and pnpm 11.7.0, `website/package.json` currently uses Next.js 16.2.12, React 19, and TypeScript 6; the migration pins `eslint-config-next` to the same 16.2.12 Next.js line and lets pnpm lock ESLint 9.39.5, matching ESLint's current flat-config documentation line.
- Rechecked package-manager/CI conventions: website CI uses Node 22 and pnpm 11.7.0 with `website/pnpm-lock.yaml`; Docker also uses pnpm. The tracked npm lockfile remains pre-existing and intentionally not repaired by this pnpm-scoped migration.
- Current focused delta remains confined to website lint tooling, website docs/CI support, and planning evidence; connector behavior changes are only from merged `origin/main`, not this PR diff.
- Continuation slice: add website lint execution to GitHub and GitLab website CI so the restored supported lint command is actually exercised in CI, refresh generated website catalog data after the current-main merge so existing CI generation checks stay green, then validate clean install, lint, typecheck, focused regression, unit/e2e/build where supported, and no-mistakes.
