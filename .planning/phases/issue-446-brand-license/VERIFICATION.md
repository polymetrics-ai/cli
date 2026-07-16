# Verification: Issue 446

## Focused Gates

- [ ] Logo/license contract test records an expected red failure.
- [ ] Logo/license contract test passes after implementation.
- [ ] Website unit suite passes.
- [ ] Website typecheck passes.
- [ ] Website production build passes.

## Repository Gates

- [ ] `go test ./...`
- [ ] `go vet ./...`
- [ ] `go build ./cmd/pm`
- [ ] `make verify`
- [ ] No stale Elastic License declarations remain in maintained copy.
- [ ] License files match the selected SPDX identifiers.

## Visual Gates

- [ ] Desktop navbar/sidebar/footer render the same PM mark.
- [ ] Mobile navbar renders the mark without clipping or layout shift.
- [ ] `P` remains visible while `M` blinks.
- [ ] Reduced-motion mode shows a stable `PM`.
- [ ] No underscore cursor is rendered.

## Review Gates

- [ ] PR uses `Closes #446` and a Conventional Commit title.
- [ ] Automated review covers the final commit range.
- [ ] All actionable findings are dispositioned.
- [ ] Repository-owner/legal approval is requested before merge.
- [ ] Production deployment and `main` merge remain human-gated.
