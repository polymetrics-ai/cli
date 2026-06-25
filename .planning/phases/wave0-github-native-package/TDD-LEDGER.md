# TDD Ledger

Phase: wave0-github-native-package

Record failing test evidence before production code for every behavior-adding task.

---

## b-registry-factory
- Test: `internal/connectors/registry_factory_test.go` → `TestRegisterFactoryIsResolvedByRegistry`
- Command: `go test ./internal/connectors/ -run TestRegisterFactoryIsResolvedByRegistry`
- Red output (before code):
  ```
  internal/connectors/registry_factory_test.go:31:2: undefined: RegisterFactory
  internal/connectors/registry_factory_test.go:32:21: undefined: unregisterFactory
  FAIL  polymetrics/internal/connectors [build failed]
  ```
- Status: red-confirmed

## b-github-package
- Test: `internal/connectors/github/github_test.go` → `TestNewContract`, `TestCatalogStreams`
- Command: `go test ./internal/connectors/github/`
- Red output (before code):
  ```
  polymetrics/internal/connectors/github: no non-test Go files in .../internal/connectors/github
  FAIL  polymetrics/internal/connectors/github [build failed]
  ```
- Status: red-confirmed

## b-registry-wiring
- Test: `internal/connectors/github/github_test.go` → `TestRegisteredInRegistry`
- Command: `go test ./internal/connectors/github/`
- Red output (before code): same build failure as above — package github does not yet exist, so
  self-registration (init → RegisterFactory) cannot resolve `github` / `source-github` in the registry.
- Status: red-confirmed
