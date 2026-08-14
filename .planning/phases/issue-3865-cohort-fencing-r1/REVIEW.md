# REVIEW — issue #3865 verified-auth cohort fencing

GSD code-review prompt: `scripts/gsd prompt code-review issue-3865-cohort-fencing-r1`.
The canonical single-worker contract forbids role spawning here, so the standard review was completed inline after targeted race, vet, lint, and build checks.

## Scope reviewed

- `internal/coordination/auth_cohort.go`
- `internal/coordination/auth_cohort_test.go`
- `.planning/phases/issue-3865-cohort-fencing-r1/`

## Review method

1. Read the final diff against `origin/integration/4015-mvp-flat-r1` and checked the changed-path scope.
2. Traced lock/store/cancellation ordering: health is stored while the coordinator lock excludes new members; cancellation callbacks execute only after the transition commits.
3. Checked stale epoch, released member, store failure, unknown outcome, and opaque-key error paths for fail-closed behavior and accidental disclosure.
4. Re-ran `go test -race -count=1 -timeout 20m ./internal/coordination/...`, `go vet ./internal/coordination/...`, `golangci-lint run ./internal/coordination/...`, and `go build ./cmd/pm`.

## Findings and dispositions

| Severity | Finding | Disposition |
| --- | --- | --- |
| Warning | Repair cleared `Fenced` without leaving a durable, secret-free record of the previous fence epoch, which weakened the issue's audit-evidence criterion. | Fixed before review completion: `AuthCohortHealth.LastFencedEpoch` is written with the verified-failure epoch, survives repair/restart, and is asserted by the repair test. |

No unresolved security, correctness, concurrency, scope, or code-quality findings remain. No out-of-scope file requires a `needs-decision` record.
