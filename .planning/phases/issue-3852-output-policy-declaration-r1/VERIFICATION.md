# VERIFICATION — issue #3852 output-policy declaration

Status: planned.

## Checklist

- [ ] RED test recorded: schema rejects runtime-supported direct-write `json` before the fix.
- [ ] `none` and `json` are declarable without changing direct-write behavior.
- [ ] Every runtime-supported direct-read and direct-write policy is declarable.
- [ ] No schema-only direct-read/write output policy survives the regression comparison.
- [ ] Existing `repository_contents_*`, `json_redacted`, `clinical_json_redacted`, and
  `binary_file_bounded` declarations remain valid.
- [ ] No existing connector bundle was rewritten.
- [ ] Authoring guidance chooses `json` or `none` for non-redacting write results.
- [ ] No #3771-owned functions or redaction behavior changed.
- [ ] Focused tests, formatting/vet, schema/connectorgen, docs, and boundary gates pass.
- [ ] GSD `verify-work` and `code-review` prompts executed inline with findings documented.

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
