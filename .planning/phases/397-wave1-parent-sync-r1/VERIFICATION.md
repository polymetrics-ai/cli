# Issue #397 Wave 1 Verification Checklist

Status: pending
`verificationPassed`: false

## Identity and scope

- [x] Isolated worktree path verified.
- [x] Branch created directly from current `origin/feat/cli-architecture-v2`.
- [x] Current main, parent, PR #438 head, and merge base pinned.
- [x] Confirmed main is not already an ancestor.
- [x] PR #438 remains draft/open and human-only.
- [x] #462 approval comment verified at exact PR #468 head.
- [x] #419 classified `deferred_by_human`.
- [x] #425–#436 waiver/review work classified pending and excluded.
- [x] #408 and downstream implementation excluded.

## GSD / skills

- [x] `scripts/gsd doctor`.
- [x] `scripts/gsd list` (69 commands).
- [x] `scripts/gsd sources` discovery and relevant source resolution.
- [x] `programming-loop` absence reconfirmed; manual lifecycle recorded.
- [x] Required CLI Architecture v2, GSD, and Go skills loaded.

## Merge and conflict correctness

- [ ] Ordinary `git merge --no-ff origin/main` completed.
- [ ] No rebase, reset, stash, force push, or history rewrite used.
- [ ] `go.mod` / `go.sum` union verified.
- [ ] `internal/cli/cli.go` preserves both Gong and Architecture v2 behavior.
- [ ] `internal/connectors/connectors.go` preserves both sides.
- [ ] `internal/connectors/connsdk/http.go` preserves both sides.
- [ ] Auto-merged related files inspected and tested.
- [ ] No conflict markers remain.

## Focused behavior

- [ ] Gong connector route(s) through dynamic dispatch.
- [ ] Cobra/native namespace and config/event routing.
- [ ] certify in-process re-entrancy.
- [ ] golden stdout/stderr/JSON/exit contracts.
- [ ] bare/help and invalid-action behavior.
- [ ] connsdk payload/query/multipart/retry/telemetry/redaction behavior.
- [ ] reverse plan → preview → approval → execute safety.

## Required commands

- [ ] `gofmt -w cmd internal`
- [ ] `git diff --exit-code -- cmd internal`
- [ ] `git diff --check`
- [ ] `go vet ./...`
- [ ] `go test -timeout 20m ./...`
- [ ] `go build ./cmd/pm`
- [ ] `go mod verify`
- [ ] `go mod tidy -diff`
- [ ] `make verify`

## Review and delivery

- [ ] Fresh-context read-only exact-head Codex review passed.
- [ ] Every finding dispositioned; exact-head re-review after any changes.
- [ ] Trajectory/Shepherd validation passed against exact evidence.
- [ ] Branch pushed normally.
- [ ] Draft stacked PR targets `feat/cli-architecture-v2` and uses `Refs #397`.
- [ ] No Claude/Copilot requested; any automatic activity observed but not counted.
- [ ] Required PR checks green at exact head.
- [ ] Parent PR #438 not edited, marked ready, or merged.

Runtime-backed, credentialed, external-mutation, deployment, and live connector checks: not applicable and forbidden for this Wave 1 slice.
