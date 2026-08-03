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

## Downstream-impact evidence

`cmd/prissueguard` was built from base `5d61794f7` and from this branch head, then both binaries were
run against every open PR body fetched live from GitHub via `gh api repos/polymetrics-ai/cli/pulls`.

| Check | Result | Notes |
|---|---|---|
| Open-PR before/after sweep | 36 changed, 63 unchanged | All 36 verdict changes are `stage unvalidated parity checkpoint` bodies: #3540, #3543, and the contiguous #3544–#3577 wave range. Zero non-checkpoint PRs changed verdict, so the general require-linked-issue rule is not loosened. |
| #3540 airtable, #3543 aws-cloudtrail | `blocked` -> `ok (8 linked issues)` | The two named PRs this change unblocks. |
| #3541 amazon-sqs, #3542 ashby | `ok` -> `ok` | Not blocked for this reason at base; their bodies carry explicit `Refs #3134` / `Refs #3207`. Verdict unchanged by this patch. |
| #3578 marketo | `blocked` -> `blocked` | Deliberately unrecognized wording variant; see the open item below. |
| #3676 nativeset foundation | `blocked` -> `blocked` | Not fixed by this change; its body has no issue-reference keyword at all. |

## CLI/help/docs/website parity evidence

| Check | Result | Notes |
|---|---|---|
| `pm` help, manual, `docs/cli/**`, `website/**` | Not applicable | `internal/coordination/issueguard` backs the `pr-issue-guard` GitHub Actions check through `cmd/prissueguard`; no `pm` command, flag, output, help topic, or connector surface changes in this slice. |
| Accepted PR-body routes | Known out of sync, follow-up | `.agents/agentic-delivery/contracts/issue-agent-contract.md` owns the list of bodies that satisfy the guard, and this slice deliberately does not touch it. Its required-workflow step 14 still enumerates only the `Closes #N` / `Refs #N` route and the no-mistakes delivery-record route, so it does not yet describe the checkpoint route added here. Updating it is a deliberate follow-up left to firstmate to file as a separate issue; precedent is commit `235c7b22f`, which updated the same step 14 for the prior delivery-record route addition. |

## Repository gate evidence

| Check | Result | Notes |
|---|---|---|
| `scripts/verify-gsd-workflow` requirement | Addressed | `internal/**` changed, so the gate requires changed `.planning/**` evidence; this phase directory supplies `PLAN.md`, `TDD-LEDGER.md`, `VERIFICATION.md`, and `RUN-STATE.json` with explicit manual-GSD fallback and `Red:`/`Green:` evidence. |

## Open items for the human merge gate

- `completedTaskPattern` still requires the trailing noun `task`, so the marketo checkpoint body
  (PR #3578), whose sentence ends in `implementation`, stays unrecognized. Widening the trailing
  noun is deliberately deferred to firstmate.
- PR #3676 (nativeset foundation) is **not** unblocked by this change. Its body carries no
  issue-reference keyword at all, so it is blocked at both base and head and needs its own PR-body
  fix, tracked separately from this slice.
- Required-workflow step 14 of `.agents/agentic-delivery/contracts/issue-agent-contract.md` is left
  out of sync on purpose; see the parity table above for the detail and the follow-up to file.
- PR stays draft; merge is human-gated.

## Safety notes

- No secrets requested, printed, summarized, or stored.
- No credentialed connector checks or live provider calls.
- No branch protection or repository settings mutation.
- No dependencies added.
- No reverse ETL execution.
