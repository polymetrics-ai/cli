# Verification — Issue 3674 Issueguard Unvalidated-Cloud-Checkpoint Issue Links

## Required commands

| Command | Result | Notes |
|---|---|---|
| `scripts/gsd doctor` | Pass | Adapter healthy; node v24.13.1, 69 commands registered. |
| `scripts/gsd prompt programming-loop init --phase issue-3674-issueguard-checkpoint-links --dry-run` | Fallback | Registry returned `unknown GSD command: programming-loop`; explicit manual-GSD fallback recorded in `PLAN.md`. |
| `scripts/gsd prompt plan-phase issue-3674-issueguard-checkpoint-links --skip-research` | Pass | 142-line planning prompt generated and applied inline. |
| `gofmt -l internal/coordination/issueguard` | Pass | No output; both changed files are formatted. |
| `go vet ./internal/coordination/issueguard/...` | Pass | Exit 0. |
| `go test ./internal/coordination/issueguard/... -count=1` | Pass | `ok polymetrics.ai/internal/coordination/issueguard 0.420s`. |
| `go build ./cmd/pm` | Pass | Exit 0. |

## Gate-pinning evidence

| Check | Result | Notes |
|---|---|---|
| Base-commit red run (`guard.go` from `5d61794f7` + current tests) | Red as expected | All four `TestValidatePRAcceptsUnvalidatedCheckpointCanonicalIssueLinks` subtests fail with `PR body must reference an issue ...`. |
| Heading-gate mutant (`unvalidatedCheckpointHeadingPattern` lookup removed) | Red as expected | `TestValidatePRRejectsCanonicalIssueSectionWithoutCheckpointHeading` fails: `ValidatePR() OK = true, want false`. |
| Wording-gate mutant (`hasCompletedTaskWording` condition removed) | Red as expected | `TestValidatePRRejectsCanonicalIssueSectionWithoutCompletedTaskWording` fails: `ValidatePR() OK = true, want false`. |

Mutation runs were executed in throwaway copies of the package outside the worktree; the branch
itself never carried a weakened gate.

## CLI/help/docs/website parity evidence

| Check | Result | Notes |
|---|---|---|
| `pm` help, manual, `docs/cli/**`, `website/**` | Not applicable | `internal/coordination/issueguard` backs the `pr-issue-guard` GitHub Actions check through `cmd/prissueguard`; no `pm` command, flag, output, help topic, or connector surface changes in this slice. |

## Repository gate evidence

| Check | Result | Notes |
|---|---|---|
| `scripts/verify-gsd-workflow` requirement | Addressed | `internal/**` changed, so the gate requires changed `.planning/**` evidence; this phase directory supplies `PLAN.md`, `TDD-LEDGER.md`, `VERIFICATION.md`, and `RUN-STATE.json` with explicit manual-GSD fallback and `Red:`/`Green:` evidence. |

## Open items for the human merge gate

- `completedTaskPattern` still requires the trailing noun `task`, so the marketo checkpoint body
  (PR #3578), whose sentence ends in `implementation`, stays unrecognized. Widening the trailing
  noun is deliberately deferred to firstmate.
- PR stays draft; merge is human-gated.

## Safety notes

- No secrets requested, printed, summarized, or stored.
- No credentialed connector checks or live provider calls.
- No branch protection or repository settings mutation.
- No dependencies added.
- No reverse ETL execution.
