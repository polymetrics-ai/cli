# Verification: Issue #3866

Status: passed locally on `fm/cli-3866-any-to-any-conformance-r1` at base
`integration/4015-mvp-flat-r1` / `a7a4c98e41af0ea32aa33db802a2a2afe0f6b8d0`.

This phase has only simulated fake/fixture assertions. It does not provide live
provider/database certification and must not be cited as such.

## Acceptance evidence

- [x] `TestTransportFamilyHalfPathConformance` executes the four named family half-paths over every member of `synccontract.AllModes()`. It asserts exact staged/applied record IDs, descriptor-selected strategy, durable acknowledgement, and committed checkpoint values.
- [x] `TestTransportFamilyHalfPathConformanceRefusesUnboundSourceBeforeIO` asserts the typed `*DestinationSourceIneligibleError` and zero source/stage/plan/apply/commit calls.
- [x] `TestTransportFamilyHalfPathConformanceAcknowledgementAndCancellation` asserts `synccontract.ErrDownstreamAcknowledgementRequired` prevents checkpoint commit and `context.Canceled` prevents destination apply/commit after staging.
- [x] `TestSharedTransportCoordinationConformance` asserts verified-auth cohort fencing, rate parking admission, durable `ErrRateParkingConflict`, exact checkpoint restart/resume without replay, and cancellation-before-resume.
- [x] The scratch sensitivity proof is recorded in `TDD-LEDGER.md`: schema validation passed; a schema-valid test-fixture binding substitution made only `api_source_to_warehouse/full_append` red; restoring it made the exact named test pass.
- [x] No production connector registration, connector definition, certification matrix/profile/capability flag, provider/database call, generic protocol executor, or CLI surface was touched.

## Commands and results

All commands below exited zero unless a deliberately documented RED command says otherwise.

```text
go test -count=1 -timeout 20m ./internal/synctransport ./internal/coordination ./internal/synccontract
go test -count=1 -timeout 20m ./internal/app
go test -count=1 -timeout 20m ./internal/cli
go test -count=1 -timeout 20m ./cmd/connectorgen
go test -race -count=1 -timeout 20m ./internal/synctransport ./internal/coordination
go vet ./internal/synctransport ./internal/coordination ./internal/synccontract ./internal/app ./internal/cli ./cmd/connectorgen
go build ./cmd/pm
go run ./cmd/agentcontractgen check
go run ./cmd/connectorgen validate internal/connectors/defs
make fmt
make tidy-check
make lint
make docs-check-no-build
make smoke-no-build
make agent-contract-check
make connectorgen-validate
make connectorgen-surface-sync
make github-parity-artifacts-check
make connectorgen-certification-matrix
make connector-boundary
make connector-canon-check
make release-workflow-check
pnpm --dir website run gen:docs
git diff --exit-code -- website/lib/docs.generated.ts
pnpm --dir website run gen:docs
git diff --exit-code -- website/lib/docs.generated.ts
git diff --check
```

The first focused RED command intentionally failed with the missing private
matrix helper. The final sensitivity RED intentionally failed with
`source family = "database", want "api"`; neither scratch change was
committed. The aggregate `go test -timeout 20m ./...` and `make verify` were
not run as single commands because the repository policy identifies their
duration under this command harness as ambiguous; changed packages, consumers,
`internal/cli`, `cmd/connectorgen`, and every non-test verify gate were run
individually.
