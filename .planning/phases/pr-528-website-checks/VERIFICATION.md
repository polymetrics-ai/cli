# Verification

Local gates run:

- `cd website && npx --yes pnpm@11.7.0 exec playwright test tests/e2e/blog-smoke.spec.ts -g "waits for profile settings"` — passed
- `cd website && npx --yes pnpm@11.7.0 exec playwright test tests/e2e/blog-annotations.spec.ts -g "profile previews gate details behind opt-in"` — skipped locally because `DATABASE_URL` is unset
- `cd website && npx --yes pnpm@11.7.0 run typecheck` — passed
- `PLAYWRIGHT_BASE_URL=http://127.0.0.1:3001 npx --yes pnpm@11.7.0 exec playwright test tests/e2e/blog-smoke.spec.ts` — passed with a manually started local Next dev server on port 3001

Notes:

- Local `DATABASE_URL` is unset, so the database-backed
  `blog-annotations.spec.ts` skips in this checkout. The mocked regression
  covers the profile-settings race that makes the database-backed author
  preview render private.
