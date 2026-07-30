# TDD ledger — connector engine/direct-read policy migration

## Red / characterization before production edits

- Baseline boundary report captured: 24 applied exceptions, 0 findings, 0 warnings; 12 in-scope provider-policy exceptions targeted `github_date_range`, `github_contents_*`, `directReadPolicyGitHub*`, and `redactGitHubContentsObject`.
- Existing tests characterized current behavior:
  - `internal/connectors/engine/direct_read_test.go` covered file metadata redaction, directory shape rejection, sensitive repository path blocking, absolute URL/path traversal rejection, and generic JSON redaction.
  - `internal/connectors/engine/read_test.go` covered current date-range normalization for numeric/date-only inputs.
  - `internal/connectors/conformance/dynamic_test.go` covered conformance cursor assertion normalization.
  - `cmd/connectorgen/main_test.go` and `internal/connectors/commandrunner/runner_test.go` covered direct-read output policy validation/dispatch.

## Red evidence captured

- Updated regression tests first to use generic `rfc3339_utc` + `operator_prefix` and generic `repository_contents_*` output policies.
- Command: `go test ./internal/connectors/engine -run 'TestReadIncrementalParamFormats|TestDirectReadRepositoryContentsPolicyIsConnectorNeutral'`
- Expected failure before production edits: `unknown field OperatorPrefix in struct literal of type IncrementalSpec`.

## Green evidence

- `go test ./internal/connectors/engine -run 'TestReadIncrementalParamFormats|TestFormatParamRFC3339UTC|TestDirectRead'` — pass.
- `go test ./internal/connectors/engine ./internal/connectors/conformance ./internal/connectors/commandrunner ./cmd/connectorgen` — pass.
- `go run ./cmd/connectorgen validate internal/connectors/defs --json` — `connectors_checked=548`, `findings=0`, `warnings=0`.
- `make connector-boundary` — pass; `findings=[]`, `warnings=[]`, applied exceptions=12.
- `go vet ./...` — pass.
- `go test ./...` — pass.
- `go build ./cmd/pm` — pass.
- `make verify` — pass, including docs validation, smoke, golangci-lint, connectorgen validate, and boundary.
- `git diff --check` — pass.

## Refactor evidence

- Removed provider-specific policy names from shared production Go for the targeted date-range and repository contents contracts.
- Drained 12 boundary exceptions rather than weakening scanner rules or broadening the ledger.
