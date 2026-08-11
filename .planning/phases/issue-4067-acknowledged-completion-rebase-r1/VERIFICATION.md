---
phase: issue-4067-acknowledged-completion-rebase-r1
issue: 4067
status: r2_local_goal_verified
mode: inline_manual_gsd_fallback
verification_date: 2026-08-11
---

# #4067 R2 goal-backward verification

**Verifier mode:** The official `verify-work` prompt was resolved, but `gsd-sdk` reports `phase_found: false` for this named issue phase and the repository custody contract forbids lifecycle-role spawning. This is the documented inline/manual fallback, not a lifecycle waiver. It records local deterministic evidence only; it is not certification, CI, live-provider verification, or merge authority.

## Goal and result

The r2 correction must terminalize only a post-acknowledgement transport error after an unrelated revision, preserve the exact acknowledged/winner and unrelated state, retain the initiating error chain, and make a missing target in an acknowledged completion rebase a typed conflict. It must not broaden #4046's typed-conflict exception or replay any checkpoint/source/destination work.

| Goal-backward truth | Result | Direct evidence |
|---|---|---|
| F1 cancellation after acknowledgement plus unrelated persisted state returns a matching durable failed run in all seven modes. | Verified | `TestRunETLTransportAcknowledgedFailureAfterUnrelatedRevisionForAllModes` — exact selector exits `0`; each witness asserts one apply, `context.Canceled`, exact acknowledged/unrelated preservation, and returned/reopened failed-run identity. |
| F1 representative source error keeps its original error identity and has the same narrow finalization. | Verified | `TestRunETLTransportAcknowledgedFailurePreservesSourceError` — exact selector exits `0`; source sentinel remains detectable with one apply and matching failed run. |
| F1 refuses any changed/missing acknowledged stream or terminal/missing target run without mutating latest state. | Verified | `TestRunETLTransportAcknowledgedFailureFailsClosedWhenTargetChanges` — exit `0`; four guards return `Run{}` and preserve the latest state while retaining the source error plus `errStateRevisionConflict`. |
| F1 persistence outcomes are truthful. | Verified | `TestRunETLTransportAcknowledgedFailureReturnsTruthfulPersistenceOutcome` — exit `0`; definite non-commit returns `Run{}`, while committed and indeterminate outcomes return the matching durable failed run and retain `CommitOutcomeError`. |
| F2 missing exact target run in an acknowledged completion rebase is a typed conflict, with zero run and no mutation, in all seven modes. | Verified | `TestRunETLTransportAcknowledgedCompletionMissingRunIsTypedConflictForAllModes` — exact selector exits `0`; all cases retain `errors.Is(err, errStateRevisionConflict)`, one apply, and unchanged reopened state. |
| #4046 typed-conflict-only failure finalization and R7/R8 source identity/per-stream CAS are unchanged. | Verified | Fresh #4046 repeated/reopen/cancellation and all-seven selectors, plus the R7/R8 selector, each exit `0`; the focused r2 race selector also includes stale-writer finalization. |

## Validation record

- RED was committed before production mutation in `e522b7bfd`: C8 and C9 real JSON-store selectors exited non-zero because the rejected candidate returned `Run{}` and left the target run durably `running`, or returned an untyped missing-run error.
- The smallest production implementation was committed in `17d6d2aaa`: it preserves only a confirmed post-acknowledgement witness, invokes a separate exact-stream/still-running failure finalizer, and wraps only an acknowledged-rebase missing target as `errStateRevisionConflict`.
- Review-driven guard coverage was committed in `e8a541a7`: `TestRunETLTransportAcknowledgedFailureFailsClosedWhenTargetChanges` directly covers changed/removed stream and terminal/removed run behavior.
- `go test -timeout 20m ./internal/app`, `go test -timeout 20m ./internal/synctransport`, `go test -timeout 20m ./internal/connectors/...`, focused `internal/cli`, `go vet ./...`, and `go build ./cmd/pm` each exit `0`.
- Individual repository gates `tidy-check`, `lint`, `docs-check`, `agent-contract-check`, connector validation/surface/certification/boundary/canon/runtime-preflight gates, release-workflow check, and GitHub-parity-artifacts check each exit `0`.
- The focused r2 `-race` selector exits `0` with no Go race report. The macOS linker emits its known malformed `LC_DYSYMTAB` warning only.

## Verification limits and remaining delivery gates

- No provider, credential, network, warehouse, container, or external service was used. `make smoke-no-build` was not run because it writes a local warehouse, which the controlling handoff forbids.
- Website generation was clean in an isolated disposable worktree. Its dependencies were unavailable there, so the active worktree's existing dependencies ran read-only `pnpm run typecheck` and `pnpm run test:scripts`; both passed. The disposable worktree was removed.
- The fresh no-mistakes correction budget remains 2/5 consumed. Loop 3/5 and any update of the existing draft #4059 are still pending; no PR/CI/merge result is claimed here.

**Local verification result:** all r2 behavioral and local quality truths above are verified. External delivery and independent Sol audit remain intentionally pending.
