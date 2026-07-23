# Issue #397 Wave 1 Verification Checklist

Status: local verification green; exact-head review and delivery pending
`verificationPassed`: false (final review/remote gates pending)

## Identity and scope

- [x] Isolated worktree path verified.
- [x] Branch created directly from current `origin/feat/cli-architecture-v2`.
- [x] Current main, parent, PR #438 head, #408 head, and merge base pinned.
- [x] Confirmed main was not already an ancestor of the parent and is now an ancestor of the task branch.
- [x] PR #438 remains draft/open, unchanged at `21d195aff0c7bd60b3bf54f14b1ce165cec9e03f`, and human-only.
- [x] #462 approval comment verified at exact PR #468 head.
- [x] #419 classified `deferred_by_human`.
- [x] #425–#436 waiver/review work classified pending and excluded.
- [x] #408 remains branch-only at `6c643f5c971d1fac4a83e4ffe653b83847c2fceb`; no implementation or PR action taken.

## GSD / skills

- [x] `scripts/gsd doctor`.
- [x] `scripts/gsd list` (69 commands).
- [x] `scripts/gsd sources` discovery and relevant source resolution.
- [x] `programming-loop` absence reconfirmed; manual lifecycle recorded.
- [x] Required CLI Architecture v2, GSD, and Go skills loaded.

## Merge and conflict correctness

- [x] Ordinary `git merge --no-ff origin/main` completed at `c545c3740c71b889fd2f1f64cec5491003f7b654`.
- [x] No rebase, reset, stash, force push, or history rewrite used.
- [x] `go.mod` / `go.sum` union verified.
- [x] `internal/cli/cli.go` preserves both Gong and Architecture v2 behavior.
- [x] `internal/connectors/connectors.go` preserves both runtime policy surfaces.
- [x] `internal/connectors/connsdk/http.go` preserves payload/retry and telemetry behavior.
- [x] Auto-merged related files inspected and tested.
- [x] No conflict markers remain.

## Focused behavior

- [x] Gong connector route(s) through dynamic dispatch.
- [x] Cobra/native namespace and config/event routing.
- [x] certify in-process re-entrancy.
- [x] golden stdout/stderr/JSON/exit contracts.
- [x] bare/help and invalid-action behavior.
- [x] connsdk payload/query/multipart/retry/telemetry/redaction behavior.
- [x] reverse plan → preview → approval → execute safety.
- [x] Focused race tests for Wave 1 CLI, connectors, connsdk, and app identities/reverse.

## Required commands

- [x] `gofmt -w cmd internal`
- [x] `git diff --exit-code -- cmd internal`
- [x] `git diff --check`
- [x] `go vet ./...`
- [x] `go test -timeout 20m ./...`
- [x] `go build ./cmd/pm`
- [x] `go mod verify`
- [x] `go mod tidy -diff`
- [x] `make verify`

## Representative CLI parity

- [x] `pm help gong --json`
- [x] bare `pm connectors`
- [x] `pm connectors --help`
- [x] native `pm version --json`
- [x] bare `pm gong --json`
- [x] invalid `pm connectors bogus --json` exits 2

## Review and delivery

- [ ] Fresh-context read-only exact-head Codex review passed.
- [ ] Every finding dispositioned; exact-head re-review after any changes.
- [ ] Trajectory/Shepherd validation passed against exact evidence.
- [x] Branch pushed normally through pre-evidence head.
- [ ] Draft stacked PR targets `feat/cli-architecture-v2` and uses `Refs #397`.
- [x] No Claude/Copilot requested; any automatic activity will be observed but not counted.
- [ ] Required PR checks green at exact head.
- [x] Parent PR #438 not edited, marked ready, or merged.

Runtime-backed, credentialed, external-mutation, deployment, and live connector checks: not applicable and forbidden for this Wave 1 slice. `make verify` exercised only its local temporary smoke, including reverse plan → preview → approval → execute.
