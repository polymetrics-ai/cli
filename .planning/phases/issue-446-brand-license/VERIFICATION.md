# Verification: Issue 446

## Focused Gates

- [x] Logo/license contract test records an expected red failure.
- [x] Logo/license contract test passes after implementation (4 tests).
- [x] Website unit suite passes (11 files, 70 tests).
- [x] Website typecheck passes after generating the ignored Fumadocs `.source` module.
- [x] Website production build passes (Next.js 16.2.10, 1,121 static pages).

## Repository Gates

- [x] `go test ./...`
- [x] `go vet ./...`
- [x] `go build ./cmd/pm`
- [x] `make verify` (547 connector definitions, 0 findings; lint and smoke pass).
- [x] No stale Elastic License declarations remain in maintained copy.
- [x] Root `LICENSE` SHA-256 matches the official GNU AGPL v3 text.
- [x] Nested definitions license matches the selected MIT text and copyright.

## Visual Gates

- [x] Desktop navbar/sidebar/footer render the same PM mark.
- [x] Mobile navbar renders the mark without clipping or layout shift.
- [x] `P` computed opacity remains `1`; every sampled `M`/`_` opacity pair is
  complementary (`1/0` or `0/1`).
- [x] Reduced-motion mode reports no animation and shows a stable `PM`.
- [x] All three marks alternate complementary `M`/`_` states in the same slot.

## Review Gates

- [ ] PR uses `Closes #446` and a Conventional Commit title.
- [ ] Automated review covers the final commit range.
- [ ] All actionable findings are dispositioned.
- [ ] Repository-owner/legal approval is requested before merge.
- [ ] Production deployment and `main` merge remain human-gated.

## Environment Notes

- Current `origin/main` has an auth dependency/lockfile mismatch, so `npm ci`
  fails before installation. Verification used `npm install --package-lock=false`
  without changing tracked dependencies or the lockfile.
- The production build completes but logs the existing Better Auth default-secret
  warning while collecting static page data. Local visual verification ran with
  a non-secret, development-only value on port 3101.
- The verified local site is available at `http://localhost:3101` while this
  worktree remains active.
