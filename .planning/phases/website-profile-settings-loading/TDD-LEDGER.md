# TDD LEDGER — website profile settings loading

## Red

- Added `website/tests/e2e/blog-smoke.spec.ts` regression `waits for profile settings before enabling visibility edits`.
- Evidence: `corepack pnpm@11.7.0 exec playwright test tests/e2e/blog-smoke.spec.ts --grep "waits for profile settings before enabling visibility edits"` exited 1 before production edits.
- Failure: `expect(locator).toBeDisabled()` expected the dialog checkbox to be disabled while `GET /api/profile` was stalled; received enabled.

## Green

- Implemented loading state in `ProfileSettingsDialog` by resetting `settings` on open, deriving `loading`, cancelling stale profile responses, and disabling visibility/url/Save controls while loading or saving.
- Evidence: manual Next server on `127.0.0.1:3107` plus `PLAYWRIGHT_BASE_URL=http://127.0.0.1:3107 E2E_TEST_AUTH=1 BETTER_AUTH_URL=http://127.0.0.1:3107 BETTER_AUTH_SECRET=e2e-only-secret-at-least-32-characters corepack pnpm@11.7.0 exec playwright test tests/e2e/blog-smoke.spec.ts --grep "waits for profile settings before enabling visibility edits"` exited 0; 1 test passed.

## Refactor

- No release coupling, docs, connector generation, dependency, or unrelated UI refactor retained.
