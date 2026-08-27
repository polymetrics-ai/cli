# TDD ledger — issue 4371

## Planned slices

| Slice | Red assertion before production edit | Green assertion | Refactor guard |
| --- | --- | --- | --- |
| Cited-only non-executable mutation | Existing application adds a second runtime gap and then conflicts with strict closed-reference validation. | Incompatible disposition is rejected before output with the exact cited source ID/location; descriptor bytes cannot change. | Normal contract-complete non-executable dispositions stay byte-stable. |
| Cited-only partial mutation coverage | Partial disposition likewise attempts to add metadata/gap to an exact cited-only descriptor. | It is rejected before output and does not make the validator accept extra gaps. | Existing partial-coverage match/duplicate/mutation checks stay fail-closed. |
| Cohort and runtime truth | Salesloft/Copper source-reference evidence cannot reconcile after the contradiction. | Each retained operation is still visible exactly once with the sole source-contract-unavailable foundation state; no generated command claims implementation. | `missing_foundation` command behavior, if an existing unavailable command is present, precedes credential/provider I/O. |

## Red

2026-08-27 — executed before production edits:

```text
go test -timeout 20m ./cmd/connectorgen -run 'TestSourceProjection(CitedOnlyMutationDispositionsKeepReferenceClosed|MutationDispositionInputRemainsFailClosed|SourceCitedNonExecutableMutationDispositionRejectsCompleteAction|SourceCitedPartialMutationCoveragePreservesImplementedIncompleteAction|SourceCitedMutationDispositionLeavesExistingProjectionByteIdentical)$' -count=1 -v
```

Failed as intended only in the new cited-only cases:

- `non executable`: `sourceProjectionApplyNonExecutableMutationDispositions`
  returned `nil` after layering its runtime disposition/gap onto
  `salesloft.rest.people.post`.
- `partial coverage`: `sourceProjectionApplyPartialMutationCoverageDispositions`
  likewise returned `nil` after adding partial-coverage metadata/gap.

Both failures prove the contradiction before descriptor output: the test
expects an actionable closed-reference refusal and byte-for-byte unchanged
descriptor. Existing contract-complete controls passed: complete-action and
implemented-action refusal, valid partial coverage with an existing typed
foundation, existing GitHub projection byte stability, and the new absent,
duplicate, unknown, mismatched, and non-mutating citation matrix.

## Green

Pending implementation.

## Refactor

Pending green evidence.
