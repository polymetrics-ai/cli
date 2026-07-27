# Plan: website lint script restoration

## GSD command path

- `scripts/gsd doctor`: passed on this worktree.
- `scripts/gsd prompt plan-phase website-lint-script --skip-research --tdd`: generated `.planning/traces/gsd-plan-phase-website-lint-script-prompt.md`.
- `scripts/gsd prompt programming-loop init --phase website-lint-script --dry-run` and `scripts/gsd prompt gsd-programming-loop init --phase website-lint-script --dry-run`: unavailable (`unknown GSD command: programming-loop`); manual GSD/TDD fallback active.

## Required skills loaded

- `gsd-core`
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
- Run website lint, typecheck, unit, e2e, and build checks for changed website files.

## Non-goals / safety gates

- No broad dependency upgrades.
- No PM binary-release config/docs, release-triggered website behavior, profile settings product code, connectors, or issue guard changes.
- No credentialed connector checks, secrets, deploys, merges, or Herdr lifecycle commands.

## Implementation slices

1. Red: add a focused Vitest regression for the website lint command/config contract and verify it fails against the current `next lint` script.
2. Green: add the minimal Next.js 16 ESLint CLI setup (`eslint .`, `eslint.config.mjs`, required lint dev dependencies) with generated/build ignores.
3. Verification: run targeted lint regression plus website lint/typecheck/unit/e2e/build checks.

## Implementation result

- Replaced `next lint` with `eslint .`.
- Added the minimal Next.js 16 ESLint CLI dependencies and flat config (`eslint`, `eslint-config-next`, `eslint.config.mjs`).
- Included Next/TypeScript lint configs and explicit generated/build/report ignores.
- Scoped pre-existing React Hooks compiler findings to warnings for known existing files so the restored command is runnable without touching out-of-scope product code.
- Added focused Vitest regression coverage for the lint script/config contract.

## Reproduction evidence

- On `origin/main`/branch start (`902fd9403`), after `corepack pnpm@11.7.0 install --frozen-lockfile --reporter=silent`, `corepack pnpm@11.7.0 run lint` failed with `Invalid project directory provided, no such directory: .../website/lint` and exit 1.

## Verification result

- Website focused regression, lint, typecheck, unit, e2e, build, clean frozen install, npm lock dry-run, and `git diff --check` passed locally. See `VERIFICATION.md`.
