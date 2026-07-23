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

- [x] Fresh-context read-only Codex review passed at `f3df1b169625891b60dce15e332c7b535dd6ff21`; no Wave 1 code findings.
- [x] Initial evidence findings dispositioned: stale #462 state corrected and base-to-head whitespace errors removed; broad security observations confirmed pre-existing/not worsened and escalated for captain follow-up.
- [x] Fresh exact-head review passed at original synchronization head `3fd63fbe0f526873fa3adb8a75fa5f20342d52a6`.
- [x] Original synchronization trajectory/Shepherd validation returned `PROCEED` (geomean 4.87) at `3fd63fbe0f526873fa3adb8a75fa5f20342d52a6`.
- [x] Branch pushed normally without force.
- [x] Draft stacked PR #495 targets `feat/cli-architecture-v2` and uses `Refs #397`.
- [x] No Claude/Copilot requested; any automatic activity is observed but not counted.
- [x] Original synchronization required branch-specific workflows passed.
- [x] Parent PR #438 not edited, marked ready, or merged.

## Captain-approved PM-orchestrator extension

- [x] Extension PLAN/TDD/VERIFY artifacts created before canonical guidance edits.
- [x] Focused RED and GREEN contract evidence recorded.
- [x] PR #493 skill/routing/Makefile paths remain untouched.
- [x] Full credential-free gates passed at implementation head `d72a93018933541d390884f96b285856e269a1ab` and corrected evidence head `0665ad7aad1ec083f4bb0572a88ac1a38f417a35`.
- [x] Captain-authorized Gong follow-up created at https://github.com/polymetrics-ai/cli/issues/497 without product changes in PR #495.
- [x] Correction round 2 closes authoritative PR #493 queue-gate, transitive PM-template, and stable correction-lineage findings without touching PR #493-owned paths.
- [ ] Final evidence-head full verification and local Codex review pass.
- [ ] Final evidence-head independent Shepherd validation passes.
- [ ] PR #495 required checks green at the final extension head.

Runtime-backed, credentialed, external-mutation, deployment, and live connector checks: not applicable and forbidden for this Wave 1 slice. `make verify` exercised only its local temporary smoke, including reverse plan → preview → approval → execute.
