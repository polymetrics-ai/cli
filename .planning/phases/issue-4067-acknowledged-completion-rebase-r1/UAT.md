---
status: complete
phase: issue-4067-acknowledged-completion-rebase-r1
source: SUMMARY.md
started: 2026-08-11T15:11:03Z
updated: 2026-08-11T15:11:03Z
mode: coverage-aware automated acceptance under documented inline/manual GSD fallback
---

## Current Test

[testing complete]

## Tests

### 1. Acknowledged completion survives an unrelated writer

expected: Each canonical sync mode returns a non-zero completed run that matches the reopened durable run after another App writes unrelated state after checkpoint acknowledgement.
result: pass
source: automated — `TestRunETLTransportAcknowledgedCompletionRebasesUnrelatedStateForAllModes`, repeated focused command, and `-race` run

### 2. Completion does not overwrite a changed target

expected: A changed, removed, or already terminal target stream/run is left untouched and reports a detectable ordinary revision conflict rather than applying a generic rebase.
result: pass
source: automated — `TestRunETLTransportAcknowledgedCompletionFailsClosedWhenTargetChanges`

### 3. Cancellation and persistence outcomes remain truthful

expected: Cancellation after checkpoint acknowledgement preserves the checkpoint and durably fails the run; definite non-commit returns zero while committed/indeterminate terminal writes return durable-consistent completed runs.
result: pass
source: automated — `TestRunETLTransportCancellationAfterAcknowledgedCheckpointForAllModes` and `TestRunETLTransportAcknowledgedCompletionReturnsTruthfulPersistenceOutcome`

### 4. Established stale-writer protections remain intact

expected: #4046's typed conflict finalization, R7/R8 source identity and per-stream CAS retain their prior behavior.
result: pass
source: automated — exact focused #4046/R7/R8 regression command

## Summary

total: 4
passed: 4
issues: 0
pending: 0
skipped: 0
blocked: 0

## Gaps

[none]

No human product-judgment checkpoint is required: every deliverable is deterministic and covered by package-local fake-backed tests. This UAT is not certification and does not attest to any live provider, credential, production warehouse, container, external service, CI result, or merge readiness.
