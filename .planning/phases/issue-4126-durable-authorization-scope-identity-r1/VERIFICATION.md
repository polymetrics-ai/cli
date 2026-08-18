# Verification — Issue 4126

## Automated checks

- [x] Red command recorded before production authorization code.
- [x] Targeted authorization tests pass.
- [x] `go test -timeout 20m ./internal/app/... ./internal/flow/... ./internal/schedule/...` passes.
- [x] Changed-package vet, `go build ./cmd/pm`, `gofmt`, and `git diff --check` pass.
- [x] Required non-suite `make verify` gates pass individually.
- [ ] Protected-branch CI completes on the integration-based PR; no-mistakes delivery is intentionally not run under the captain's recovery decision.
- [x] CI replay-classification regression: `go test -timeout 20m ./internal/cli -run '^TestReverseApprovalReplayRejectsBeforeOpeningLegacyProject$' -count=1` and `go test -timeout 20m ./internal/cli` pass with `validation` / `validation_error` and exit 3.

## Requirement checks

- [x] Scope stays stable despite payload records, counts, and timestamps.
- [x] Every bound property changes the scope identity.
- [x] Identical scope can dispatch unattended with no token.
- [x] Changed scope, revocation, expiry, and replay have typed refusals and zero sends.
- [x] The persisted record and output contain no secret, credential material, or approval token.
- [x] `reversePlanHash` is unchanged (implementation was not modified).

## Manual review focus

- [x] No flow runner, schedule firing, or GitHub destination implementation leaked into this foundation.
- [x] Scope serialization holds only safe references and derived digests, never raw destination configuration.
- [x] Repeat authorization checks run before provider dispatch.
