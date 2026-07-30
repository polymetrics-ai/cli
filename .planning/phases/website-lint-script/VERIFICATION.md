# Verification — website lint script restoration

- [x] Reproduced pre-existing `pnpm run lint` failure on `origin/main`/branch start with installed dependencies: `next lint` treated `lint` as a project directory and exited 1.
- [x] Focused red regression: `corepack pnpm@11.7.0 exec vitest run tests/lint-tooling.test.ts` failed before implementation.
- [x] Focused green regression: `corepack pnpm@11.7.0 exec vitest run tests/lint-tooling.test.ts` passed.
- [x] `corepack pnpm@11.7.0 run lint` — passed with 13 warnings, 0 errors.
- [x] `corepack pnpm@11.7.0 run typecheck`
- [x] `corepack pnpm@11.7.0 run test:unit` — 12 files / 72 tests passed.
- [x] `corepack pnpm@11.7.0 run test:e2e` — 18 passed / 7 skipped.
- [x] `BETTER_AUTH_SECRET=<build-only placeholder> corepack pnpm@11.7.0 run build`
- [x] Clean dependency install: `rm -rf website/node_modules && corepack pnpm@11.7.0 install --frozen-lockfile --reporter=append-only`.
- [x] Post-clean-install lint error check: `corepack pnpm@11.7.0 exec eslint . --quiet`.
- [x] Tracked npm lockfile intentionally left unchanged; website CI/Docker use `pnpm-lock.yaml`, and npm lock repair would include unrelated pre-existing dependency drift.
- [x] `git diff --check`

## CI check repair — 2026-07-27

- [x] Dependency Review log inspected: high-severity `brace-expansion@1.1.16` from `website/pnpm-lock.yaml` (GHSA-mh99-v99m-4gvg).
- [x] Focused red regression: `CI=true COREPACK_HOME="$PWD/.corepack" corepack pnpm@11.7.0 --dir website exec vitest run tests/lint-tooling.test.ts --reporter=verbose` failed while lockfile contained `brace-expansion@1.1.16`.
- [x] Lockfile repair: `CI=true COREPACK_HOME="$PWD/.corepack" corepack pnpm@11.7.0 --dir website install --lockfile-only --ignore-scripts --store-dir ../.pnpm-store --reporter=append-only`.
- [x] Clean frozen install: `CI=true COREPACK_HOME="$PWD/.corepack" corepack pnpm@11.7.0 --dir website install --frozen-lockfile --ignore-scripts --store-dir ../.pnpm-store --reporter=append-only`.
- [x] Focused green regression: `CI=true COREPACK_HOME="$PWD/.corepack" corepack pnpm@11.7.0 --dir website exec vitest run tests/lint-tooling.test.ts --reporter=verbose` — 3 tests passed.
- [x] Website lint compatibility after pnpm patch: `CI=true COREPACK_HOME="$PWD/.corepack" corepack pnpm@11.7.0 --dir website run lint` — passed with 13 warnings, 0 errors.
- [x] Issue guard source left unchanged; local body-shape validation uses `Refs #67` for the live PR body update.
- [x] Website image log inspected: Docker deps stage failed because `patches/minimatch@3.1.5.patch` was not copied before `pnpm install --frozen-lockfile`.
- [x] Focused Dockerfile red regression: `CI=true COREPACK_HOME="$PWD/.corepack" corepack pnpm@11.7.0 --dir website exec vitest run tests/lint-tooling.test.ts --reporter=verbose` failed until `Dockerfile` included `COPY patches ./patches`.
- [x] Focused Dockerfile green regression: `CI=true COREPACK_HOME="$PWD/.corepack" corepack pnpm@11.7.0 --dir website exec vitest run tests/lint-tooling.test.ts --reporter=verbose` — 4 tests passed.
- [x] Docker deps-stage verification: `docker build --target deps -f website/Dockerfile -t polymetrics-website-deps:ci-fix website`; removed the temporary local image tag with `docker image rm polymetrics-website-deps:ci-fix`.

## Current-main continuation — 2026-07-30

- [x] Isolated worktree path verified: `/Users/karthiksivadas/.treehouse/cli-83d592/5/cli`.
- [x] Revalidated PR #542 base/head/ownership: open PR to `main`, head `fix/website-lint-script`, author association `MEMBER`, merge state `behind` before local non-rewriting merge.
- [x] Fetched current `origin/main@c85740b6f` and merged it into `fix/website-lint-script` without force-push, reset, or dropping prior commits.
- [x] Official Next.js 16 guidance checked: `next lint` removed; use ESLint directly and the migration codemod can migrate `next lint` to the ESLint CLI.
- [x] Official Next.js ESLint setup checked: install `eslint`/`eslint-config-next`, create `eslint.config.mjs`, include `core-web-vitals` and `typescript`, run `pnpm exec eslint .`.
- [x] Official ESLint flat-config/ignore guidance checked: use native `eslint.config.*` flat config plus `defineConfig`/`globalIgnores` for global ignores.
- [x] Clean pnpm install from no `website/node_modules`: `CI=true COREPACK_HOME="$PWD/.corepack" corepack pnpm@11.7.0 --dir website install --frozen-lockfile --store-dir ../.pnpm-store --reporter=append-only`.
- [x] Actual website lint after clean install: `CI=true COREPACK_HOME="$PWD/.corepack" corepack pnpm@11.7.0 --dir website run lint` — passed with 13 warnings, 0 errors.
- [x] Generated website data refresh after current-main merge: `CI=true COREPACK_HOME="$PWD/.corepack" corepack pnpm@11.7.0 --dir website run gen:website-data`; generated GitHub catalog output policies now match the current generic repository read policies from `origin/main` (`github_contents_*` -> `repository_contents_*`), with no connector source changes in this PR.
- [x] Focused lint-tooling regression after CI lint coverage update: `CI=true COREPACK_HOME="$PWD/.corepack" corepack pnpm@11.7.0 --dir website exec vitest run tests/lint-tooling.test.ts --reporter=verbose` — 5 tests passed.
- [x] Typecheck after current-main merge: `CI=true COREPACK_HOME="$PWD/.corepack" corepack pnpm@11.7.0 --dir website run typecheck`.
- [x] Unit tests after current-main merge with a temporary local `postgres:17-alpine` container: `CI=true COREPACK_HOME="$PWD/.corepack" corepack pnpm@11.7.0 --dir website run test:unit` — 12 files / 75 tests passed.
- [x] End-to-end tests after current-main merge with a temporary local `postgres:17-alpine` container and an isolated Next dev port because 127.0.0.1:3000 was already in use: `PLAYWRIGHT_BASE_URL=http://127.0.0.1:3107 corepack pnpm@11.7.0 --dir website run test:e2e` — 26 passed.
- [x] Build after current-main merge with a build-only placeholder auth secret and temporary local Postgres: `corepack pnpm@11.7.0 --dir website run build`.
- [x] `git diff --check`.
- [ ] no-mistakes run and GitHub checks green for PR #542.
