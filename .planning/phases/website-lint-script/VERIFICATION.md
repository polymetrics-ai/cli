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
