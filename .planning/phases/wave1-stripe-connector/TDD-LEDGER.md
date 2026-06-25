# TDD Ledger

Phase: wave1-stripe-connector

Record failing test evidence before production code for every behavior-adding task.

---

## b-stripe-read
- Test: `internal/connectors/stripe/stripe_test.go` → `TestReadPaginatesAndAuthenticates`
- Command: `go test ./internal/connectors/stripe/`
- Red output (before code):
  ```
  polymetrics/internal/connectors/stripe: no non-test Go files in .../internal/connectors/stripe
  FAIL  polymetrics/internal/connectors/stripe [build failed]
  ```
- Status: red-confirmed

## b-stripe-write
- Test: `internal/connectors/stripe/stripe_test.go` → `TestWriteValidateAllowList`
- Command: `go test ./internal/connectors/stripe/`
- Red output: same build failure — package stripe does not yet exist, so WriteValidator/allow-list
  cannot be exercised.
- Status: red-confirmed

## b-stripe-catalog
- Test: `internal/connectors/stripe/stripe_test.go` → `TestRegisteredWithWriteCapability` +
  `pm connectors inspect stripe`
- Red output: same build failure — stripe is not registered and the catalog entry is still
  planned_native_port, so the registry/inspect path cannot resolve a live stripe connector.
- Status: red-confirmed
