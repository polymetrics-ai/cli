---
phase: issue-4067-acknowledged-completion-rebase-r1
status: local_goal_verified
---

# #4067 verification checklist

**Manual GSD verifier:** completed inline on 2026-08-11 after `scripts/gsd prompt verify-work issue-4067-acknowledged-completion-rebase-r1` resolved the official workflow. The custom issue phase is not a numeric roadmap phase and role spawning is forbidden in this custody lane; `UAT.md` records coverage-aware automated acceptance. This is a local deterministic verification result, not a certification, CI result, live-provider check, or merge authorization.

## Goal-backward result

| Truth | Status | Evidence |
|---|---|---|
| A post-checkpoint unrelated write no longer leaves an acknowledged run durably `running` or returns zero. | Verified | All-seven real JSON-store witness; repeated focused command. |
| Completion rebases only the exact still-running acknowledged target; changed/missing/terminal targets fail closed. | Verified | `TestRunETLTransportAcknowledgedCompletionFailsClosedWhenTargetChanges`. |
| Target checkpoint and unrelated current state are retained while only final run/stream metadata changes. | Verified | all-seven witness plus truthful-outcome fixture. |
| Return/error outcome matches persistence truth and cancellation remains durable. | Verified | outcome table plus cancellation all-seven test. |
| #4046/R7/R8 behavior remains preserved. | Verified | exact focused regression command and race run. |

**Score:** 5/5 local observable truths verified.

## Pre-production gates

- [x] Controlling Sol correction directive read completely.
- [x] #3862 child tree, #3864, #4046, #4059 checks/comments/reviews, and the three Sol/R9 reports read before edits.
- [x] New child #4067 created, linked under #3864, and read back before production/generated changes.
- [x] Existing branch/PR custody preserved; rejected `883a86cf0040d559edcd4777413d1c2de20cd94a` is an immutable baseline.
- [x] CodeGraph absence recorded; required skills and GSD/agent-contract checks completed.
- [x] Named inline/manual GSD problem, context, discussion, TDD plan, execution record, verifier evidence, review, summary, and run-state artifacts exist before code.
- [x] Behavioral RED exits non-zero for the durable completion leak before production mutation: all seven canonical modes retained acknowledged/unrelated state first, then observed zero returned run plus durable reopened `running` run.

## Required focused matrix

- [x] Exact post-checkpoint/pre-completion two-App interleaving in all seven modes.
- [x] Returned/durable identity and reopened terminal truth for the all-seven-mode main witness.
- [x] Target running/exact-checkpoint eligibility and fail-closed changed/missing/terminal targets.
- [x] Winner/acknowledged checkpoint and unrelated state preservation.
- [x] Cancellation after acknowledgement in all seven modes.
- [x] All seven canonical modes.
- [x] Focused race detector (`-count=3`; no Go race report).
- [x] #4046 typed-conflict and R7/R8 regression suite.

## Subsequent gates

- [x] Canonical website generator/check refreshes `website/lib/docs.generated.ts` only; the output contains the candidate-owned transport-eligibility section.
- [x] Canonical certification generator/check refreshes `internal/connectors/certifications/flow-matrix.json` only; its two discovery-source lines match current `internal/cli/cli.go`.
- [x] Affected tests, lint, vet/build, and individual repository gates pass after the heavy-validation window notification.
- [x] Manual `verify-work` record and coverage-aware automated UAT contain local evidence.
- [x] Manual `code-review` record dispositions every finding.
- [ ] Fresh #4067 no-mistakes run starts at 0/5 without `--yes`; old run is not queried for control or modified.
- [ ] Existing draft #4059 is updated normally, stays unmerged, and exact-head CI is green before requesting an independent Sol audit.

## Heavy local matrix completed

- `go test -count=1 -timeout 20m ./internal/app` — exit `0`.
- `go vet ./internal/app/...` and `go build ./cmd/pm` — exit `0`.
- `make tidy-check`, `make lint`, `make docs-check`, `make smoke-no-build`, `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync`, `make github-parity-artifacts-check`, `make connectorgen-certification-matrix`, `make connector-boundary`, `make connector-canon-check`, and `make release-workflow-check` — each exit `0` individually.
- `cd website && pnpm typecheck` — exit `0`; `pnpm lint` exits `0` with only existing warnings in unrelated website components.
- `scripts/verify-gsd-workflow` — exit `0`, recognizing this phase's planning/TDD evidence against `origin/main`.
