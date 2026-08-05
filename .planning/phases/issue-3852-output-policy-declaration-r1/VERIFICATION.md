# VERIFICATION — issue #3852 output-policy declaration

Status: passed (issue-appropriate local gates; full monolithic suite remains CI-owned by repository policy).

## Checklist

- [x] RED test recorded: schema rejects runtime-supported direct-write `json` before the fix.
- [x] `none` and `json` are declarable without changing direct-write behavior.
- [x] Every runtime-supported direct-read and direct-write policy is declarable.
- [x] No schema-only direct-read/write output policy survives the regression comparison.
- [x] Existing `repository_contents_*`, `json_redacted`, `clinical_json_redacted`, and
  `binary_file_bounded` declarations remain valid.
- [x] No existing connector bundle was rewritten.
- [x] Authoring guidance chooses `json` or `none` for non-redacting write results.
- [x] No #3771-owned functions or redaction behavior changed.
- [x] Focused tests, formatting/vet, schema/connectorgen, docs, and boundary gates pass.
- [x] GSD `verify-work` and `code-review` prompts executed inline with findings documented.

## Planned commands

- `go test ./internal/connectors/engine -run 'TestBundleLoadAcceptsRuntimeSupportedNonRedactingDirectWriteJSONPolicy'`
- `go test ./internal/connectors/engine ./internal/connectors/commandrunner`
- `gofmt -w internal/connectors/commandrunner/runner.go internal/connectors/commandrunner/runner_test.go internal/connectors/engine/bundle_test.go`
- `go vet ./internal/connectors/engine ./internal/connectors/commandrunner`
- `go run ./cmd/connectorgen surface-sync --check`
- `go run ./cmd/connectorgen validate`
- `make docs-check-no-build`
- `make connector-boundary`
- `git diff --check`

The full `go test ./...` and `make verify` monolith remain CI-owned under the repository timeout
policy; their applicable component gates will run separately.

## Executed so far

- RED: `go test ./internal/connectors/engine -run '^TestBundleLoadAcceptsRuntimeSupportedNonRedactingDirectWriteJSONPolicy$' -count=1` failed as expected before the schema edit; the exact error is in `TDD-LEDGER.md`.
- GREEN focused: the same engine test passed after the schema edit.
- GREEN drift guard: `go test ./internal/connectors/commandrunner -run '^TestCLISurfaceOutputPolicyEnumMatchesRuntimePolicySets$' -count=1` passed.
- Regression packages: `go test ./internal/connectors/engine ./internal/connectors/commandrunner` passed.
- Bundle surface metadata: `go run ./cmd/connectorgen surface-sync --check` scanned 550 connectors with 0 fields filled/corrected.
- Bundle validation: `go run ./cmd/connectorgen validate` checked 550 connectors with 0 findings.
- Formatting/check: `gofmt -w internal/connectors/commandrunner/runner.go internal/connectors/commandrunner/runner_test.go internal/connectors/engine/bundle_test.go` and `git diff --check` passed.
- Package tests: `go test ./internal/connectors/engine ./internal/connectors/commandrunner` passed; this includes command-runner preflight coverage for implemented commands. `go test ./internal/cli` passed.
- Static/build: `go vet ./internal/connectors/engine ./internal/connectors/commandrunner` and `go build ./cmd/pm` passed.
- Individual verification gates passed: `make tidy-check`, `make docs-check-no-build`, `make smoke-no-build`, `make lint`, `make connector-boundary`, `make agent-contract-check`, and `make release-workflow-check`.
- GSD evidence: the inline automated UAT passed three of three deliverables (`UAT.md`); the inline standard source review found no Critical, Warning, or Info findings (`REVIEW.md`). The contract forbids spawned review roles, so both artifacts explicitly record the fallback.

## CLI/docs/website parity result

`rg -n -i 'output[_ ]policy' docs/cli website --glob '*.md'` returned no public command-documentation
mentions. This is metadata/schema work rather than a CLI command, flag, help topic, manual, or website
page change; `docs/migration/conventions.md` is the applicable authoring surface and passed its docs gate.
