# TDD ledger — Issue #4293 retained source-evidence receipt cohort R1

## Red

Before production implementation, the focused command was:

```text
GOCACHE=/private/tmp/gocache-4293-retained-source-receipt-r1 \
  go test -timeout 20m ./cmd/connectorgen \
  -run '^(TestBatch1SourceOperationMappingCohortRetentionReceipts|TestRetainedSourceMapping)' \
  -count=1
```

It exited `1` in 2.520s with the intended missing-control compile failure:

```text
cmd/connectorgen/sourceoperationmapping_test.go:224:19: undefined: sourceOperationMappingCohortRetentionReceiptCheck
cmd/connectorgen/sourceoperationmapping_test.go:304:21: undefined: sourceOperationMappingCohortRetentionReceiptCheck
```

The new tests therefore specify a cohort receipt check and its flag before any
production implementation exists. They also specify missing/byte-drifted
fixture receipts, CircleCI's alternate matrix form, and zero executable
declarations.

## Green target

The same focused suite must prove:

- all eight eligible v2 sidecars are deterministic retention-only receipts for
  2,340 source IDs;
- every receipt has zero executable declarations and no runtime fields;
- a missing or byte-drifted fixture receipt fails with its connector identity;
- the existing CircleCI matrix alias/form survives cohort verification;
- `--check-retention-receipts` requires the existing check-only command form;
- canonical-evidence source-import admission remains a distinct, unchanged
  boundary.

## Green

After the minimal receipt/cohort implementation, the focused command passed:

```text
GOCACHE=/private/tmp/gocache-4293-retained-source-receipt-r1 \
  go test -timeout 20m ./cmd/connectorgen \
  -run '^(TestBatch1SourceOperationMappingCohortRetentionReceipts|TestSourceOperationMappingCohortRetentionReceiptOptionsAndHelp|TestRetainedSourceMapping)' \
  -count=1
```

Result: exit `0` in 88.236s. The first green attempt exposed only an overly
specific test assertion prefix for the intentionally connector-scoped byte
drift diagnostic; the test now checks both the `stripe` scope and the stable
`bytes drifted` reason. No production behavior was weakened.

## Refactor guard

The validator may share source-lock/matrix parsing with the existing retained
mapping command, but it must not call source import, source materialization,
source projection, runtime bundle loading, or any persistence function.  It
must neither introduce a descriptor nor mutate a sidecar.

## Baseline separation

The scoped package's six unrelated failures were rerun as the exact subset
against a clean detached checkout of frozen base
`ceaae873aef0dd19aa23c036b9cb598f9b3eacc8`:

```text
GOCACHE=/private/tmp/gocache-4293-baseline-ceaae \
  go test -timeout 6m ./cmd/connectorgen \
  -run '^(TestEnabledConnectorContractsKeepExecutableLanesImplementedWhenSourceMappingIsPartial|TestOperationEvidenceGitLabSourceLockBridge|TestRetainedAsanaSourceImportRejectsReadProjectionDrift|TestRetainedAsanaMutationDispositionsCoverEveryDeferredSourceOperation|TestSourceProjectionGapCreatesCommandFromExistingClosedActionVariant|TestSourceProjectionSourceCitedMutationDispositionLeavesExistingProjectionByteIdentical)$' \
  -count=1
```

It exited `1` in 13.834s with the same Asana/GitLab/GitHub expectation and
projection-drift failures. Those tests are defined in
`enabledcontract_final_test.go`, `operationevidence_test.go`, and
`sourceprojection_test.go`; none is changed by this issue. This establishes
the full-package result as a frozen-base failure, not a #4293 regression.
