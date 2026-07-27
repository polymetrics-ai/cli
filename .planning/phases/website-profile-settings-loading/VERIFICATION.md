# VERIFICATION — website profile settings loading

## Checklist

- [x] Focused red Playwright regression fails on current behavior.
- [x] Focused Playwright regression passes after implementation.
- [!] `pnpm run gen:website-data` was executed, but current `origin/main` generated website connector data is stale for Bahmni and the command dirties connector-derived generated files outside this PR's hard boundary.
- [x] `pnpm run typecheck` passes.
- [x] `pnpm run test:unit` passes.
- [x] `pnpm run test:e2e` passes in CI mode.
- [x] `pnpm run build` passes.
- [x] `git diff --check` passes.

## Results

- Install: `CI=1 corepack pnpm@11.7.0 install --frozen-lockfile --reporter=silent` passed. Local pnpm v9.15.4 cannot consume this lockfile with `--frozen-lockfile`; the website workflow pins pnpm 11.7.0.
- Red: `corepack pnpm@11.7.0 exec playwright test tests/e2e/blog-smoke.spec.ts --grep "waits for profile settings before enabling visibility edits"` exited 1 before production edits with `Expected: disabled`, `Received: enabled` for the profile visibility checkbox.
- Green focused: started the website manually on `127.0.0.1:3107` to avoid unrelated local services already bound to port 3000, then ran `PLAYWRIGHT_BASE_URL=http://127.0.0.1:3107 E2E_TEST_AUTH=1 BETTER_AUTH_URL=http://127.0.0.1:3107 BETTER_AUTH_SECRET=e2e-only-secret-at-least-32-characters corepack pnpm@11.7.0 exec playwright test tests/e2e/blog-smoke.spec.ts --grep "waits for profile settings before enabling visibility edits"`; exited 0, 1 passed.
- Generated data: `mise exec node@22.22.3 -- corepack pnpm@11.7.0 run gen:website-data` exited 0 but changed `website/data/connectors.generated.json` and `website/lib/connectors.catalog.data.generated.json` for Bahmni connector-derived data. Those generated files were reverted and not committed because this task forbids connector/generated-surface changes.
- Typecheck: `corepack pnpm@11.7.0 run typecheck` exited 0.
- Unit: `corepack pnpm@11.7.0 run test:unit` exited 0; 11 files / 70 tests passed.
- E2E: started the website manually on `127.0.0.1:3108` to avoid unrelated local services already bound to port 3000, then ran `PLAYWRIGHT_BASE_URL=http://127.0.0.1:3108 E2E_TEST_AUTH=1 BETTER_AUTH_URL=http://127.0.0.1:3108 BETTER_AUTH_SECRET=e2e-only-secret-at-least-32-characters CI=1 corepack pnpm@11.7.0 run test:e2e`; exited 0, 26 passed.
- Build: `corepack pnpm@11.7.0 run build` exited 0; Next build completed with pre-existing Better Auth default-secret warnings during static data collection.
- Diff check: `git diff --check` exited 0.
