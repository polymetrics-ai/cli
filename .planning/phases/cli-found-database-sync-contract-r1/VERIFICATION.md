# Verification checklist — Issue 3810: shared database sync contract

## Required behavioral checks

- [x] Seven and only seven contract modes validate.
- [x] A contract mode is not executable without matching native executor and complete reusable
      fixture evidence.
- [x] State-envelope JSON round-trips all required fields, including a dedupe window, and opaque
      non-UTF-8 tokens byte-for-byte.
- [x] Per-partition state remains an array of independent structured positions.
- [x] Every listed recovery condition is typed and requires rebootstrap without clearing state.
- [x] State committer is called only after an explicit durable downstream acknowledgement.
- [x] Tombstones require identity/key/image/order and history deletes close `_valid_to`/
      `_is_current` validity windows with an explicit `_valid_from` protocol column.
- [x] Native contract has no REST/API-surface/raw SQL/raw HTTP/shell field.
- [x] Legacy scalar `cursor` state blocks resumption explicitly rather than silently scanning.
- [x] Existing legacy sync tests still prove append/overwrite/incremental behavior through the
      envelope adapter.

## Local command checklist

- [x] `go test ./internal/synccontract -count=1`
- [x] `go test ./internal/app -count=1`
- [x] `go test ./internal/connectors -count=1`
- [x] `go test ./internal/connectors/engine -count=1`
- [x] `go test ./internal/cli -count=1`
- [x] `gofmt -w internal/synccontract internal/app`
- [x] `go vet ./...`
- [x] `go build ./cmd/pm`
- [x] `make tidy-check`
- [x] `make lint`
- [x] `make docs-check`
- [x] `make smoke-no-build`
- [x] `make agent-contract-check`
- [x] `make connectorgen-validate`
- [x] `make connectorgen-surface-sync`
- [x] `make connector-boundary`
- [x] `make release-workflow-check`
- [x] `scripts/gsd prompt verify-work cli-found-database-sync-contract-r1` fallback record / manual
      verification and `scripts/gsd prompt code-review cli-found-database-sync-contract-r1` review
      record.

## Deliberate not-applicable parity

No CLI command, flag, output, connector bundle surface, generated manual, docs, or website catalog
is changed here. #3748 and #3860 own user-facing surfacing; this foundation only creates importable
internal contracts and compatibility persistence. CLI/docs/website parity is therefore N/A for this
slice and must be addressed by those consumer lanes.

## Additional scoped lint note

`golangci-lint run ./internal/app/... ./internal/state/... ./internal/durability/... ./internal/synccontract/...`
was rerun during documentation and lint housekeeping. The scoped production close warnings in
`internal/app/local_warehouse.go` and `internal/app/util.go` were fixed. Two pre-existing
`errcheck` findings remain in `internal/app/query_engine_helpers_test.go` and
`internal/app/reverse_approval_test.go`; test changes are outside this phase.
`make lint` itself passes.

## CI remediation — rebased continuation

The original PR #3879 branch predated `2afd1a529`
(`fix(certify): bound certification harness cost`, #3878), so its hosted
`verify` run reached the old 20-minute certification deadline. This continuation
is rebased on current `origin/main`, which already contains #3878. It deliberately
does not vendor its test-only optimization: `git diff --name-only
origin/main...HEAD` contains no certification harness, certification CLI, workflow,
or timing-gate paths.

The upstream remediation retains real full-route and sample/outbox lifecycle
proofs while scripting duplicate stage coverage behind a protocol-validating
driver. Its recorded checks were:

- `go test -count=1 -timeout 5m ./internal/connectors/certify` — pass in
  22.069s.
- `go test -count=1 -timeout 5m -run '^TestCertifyCLI' ./internal/cli` —
  pass in 87.119s.

The `require-linked-issue` workflow reads the live GitHub PR body, not a source
change alone. [`PR-BODY.md`](PR-BODY.md) is the exact replacement input for
the outer PR phase; it preserves the issue-first intent and declares the
completed-work link `Closes #3810`.

- [x] `go run ./cmd/prissueguard --title 'feat(synccontract): add durable database sync contract' --body-file .planning/phases/cli-found-database-sync-contract-r1/PR-BODY.md`
      returns `issueguard: ok (1 linked issue)`.
- [x] PR #3879's live body was replaced with `PR-BODY.md`, including `## Intent`
      and `Closes #3810`; the completed no-mistakes run accepted the repaired
      linked-issue gate.
- [ ] The rebased successor PR must use the same issue-linked body and complete
      fresh CI against current `main` before human review.
Required remediation skills: `golang-how-to`, `golang-troubleshooting`,
`golang-performance`, `golang-benchmark`, `golang-testing`,
`golang-cli`, `golang-error-handling`, `golang-security`,
`golang-safety`, `golang-lint`, `golang-continuous-integration`, and
`github-issue-first-delivery`.
