# Verification — issue 4293 whole-cohort validation R2

## Completed

- Red test recorded in `TDD-LEDGER.md`.
- Focused real-cohort and negative evidence suite passed with the isolated
  cache recorded in the ledger.
- Affected existing cohort/receipt regression suite passed in 102.316s:

  ```text
  GOCACHE=/private/tmp/gocache-4293-green.o8TgvF \
    go test ./cmd/connectorgen \
    -run '^Test(Batch1SourceOperationMappingCohort(CheckAcceptsTrackedDenominator|CheckValidatesMatricesAndDeclaredArtifactLinks|RetentionReceipts|RetentionReceiptsRejectMissingAndDriftedSidecars|CheckRejectsDigestCountAndMembershipDefects)|SourceOperationMappingCohortRetentionReceiptOptionsAndHelp)$' \
    -count=1 -v
  ```

- `GOCACHE=/private/tmp/gocache-4293-green.o8TgvF go vet ./cmd/connectorgen`:
  passed.
- Direct public proof passed:

  ```text
  GOCACHE=/private/tmp/gocache-4293-green.o8TgvF \
    go run ./cmd/connectorgen source-operation-mapping-cohort \
    data/connector-canon/batch1-source-operation-mapping-cohort.json --check
  ```

  It reported `10 / 4341` primary denominator, `4343 / 30401` matrix
  accounting, `917 / 930` explicit link accounting, `5886` typed deferred
  projection deficits, and `0` executable declarations.
- `GOCACHE=/private/tmp/gocache-4293-green.o8TgvF go run
  ./cmd/agentcontractgen check`: passed.
- `gofmt` was applied to all changed Go files.
- `git diff --check` passed after the green implementation.
- `jq empty` passed for the Batch R1 cohort, all ten source-lane matrices, and
  this phase's run state.

## Remaining before candidate commit

- Inspect the exact staged file list and diff, then commit/push the candidate
  only after confirming the scope has no connector/runtime/Atlas change.

## Explicitly not attempted

No broad or race Go suite, parent integration, runtime test, connector
artifact generation, source lock/matrix rewrite, Atlas update, or credential/
provider-I/O path is part of this slice.
