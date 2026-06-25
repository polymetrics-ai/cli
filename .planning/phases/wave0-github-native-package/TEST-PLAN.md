# TEST-PLAN — GitHub Native Package + Data-Driven Registry

## Gate command
`make verify` (gofmt + go vet + `go test ./...` + build + end-to-end smoke). Deterministic, bash.

## Red-first tests (must fail before code)
1. **Registry factory** (`internal/connectors/registry_gen_test.go` or `registry_test.go`):
   register a throwaway factory via `RegisterFactory("zz_test_conn", ...)`, build `NewRegistry()`,
   assert `Get("zz_test_conn")` resolves. Red until `RegisterFactory` exists.
2. **GitHub package** (`internal/connectors/github/github_test.go`):
   - `github.New().Name() == "github"`.
   - `Metadata().Capabilities` has Read && Write (reverse-ETL writable).
   - `Catalog` returns the known streams (issues, pull_requests, repository, …).
   - `Check`/`Read` against an `httptest` server (token auth) returns expected records.
   - `ValidateWrite` accepts a known write action (e.g. create_issue) and rejects an unknown one.
   Red until the package exists.

## Parity / regression tests (stay green)
- All existing `internal/connectors`, `internal/app`, `internal/cli` tests.
- `pm connectors inspect github --json` → kind "Connector".
- Registry resolves both `github` and `source-github`.
- `NativeConformanceReports` length == catalog length (unchanged).

## Evidence
Red runs captured in TDD-LEDGER.md (Status: red-confirmed) before implementation; green runs after.
