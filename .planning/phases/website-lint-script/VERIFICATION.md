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
