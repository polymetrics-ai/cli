# Verification checklist — connector engine/direct-read policy migration

## Targeted gates

- [x] `gofmt -w cmd internal`
- [x] `go test ./internal/connectors/engine -run 'TestReadIncrementalParamFormats|TestFormatParamRFC3339UTC|TestDirectRead'`
- [x] `go test ./internal/connectors/engine ./internal/connectors/conformance ./internal/connectors/commandrunner ./cmd/connectorgen`
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs --json` — `connectors_checked=548`, `findings=0`, `warnings=0`.
- [x] `go run ./cmd/connectorgen boundary . --json` / `make connector-boundary` — `outcome=clean`, `findings=0`, `warnings=0`, `exceptions=12`.

Note: the current `connectorgen validate` command expects a root containing connector bundle directories; `internal/connectors/defs/github` is not used as a single-bundle validation root because it treats `schemas/` and `fixtures/` as candidate bundle directories. Full defs validation covers GitHub and all other bundles.

## Repository gates before commit

- [x] `go vet ./...`
- [x] `go test ./...`
- [x] `go build ./cmd/pm`
- [x] `make verify`
- [x] `git diff --check`

## CLI/help/docs/website parity

This slice changes internal connector definition policy metadata and validator/runtime behavior, not CLI commands, flags, output text, generated manuals, or website docs. Runtime GitHub direct-read behavior remains compatible under generic internal policy names in `internal/connectors/defs/github/cli_surface.json`.

## Boundary inventory

Before: 24 applied exceptions, 12 in-scope targeted exceptions for date-range/repository-contents policies.

After: 12 applied exceptions, 0 findings, 0 warnings. Removed in-scope exception IDs:

- `github-connectorgen-param-format`
- `github-connectorgen-output-policy-file`
- `github-connectorgen-output-policy-directory`
- `github-commandrunner-output-policy-file`
- `github-commandrunner-output-policy-directory`
- `github-conformance-date-range-format`
- `github-engine-date-range-format`
- `github-engine-direct-read-policy-const-file`
- `github-engine-direct-read-policy-const-directory`
- `github-engine-direct-read-policy-value-file`
- `github-engine-direct-read-policy-value-directory`
- `github-engine-direct-read-redaction-helper`

## Safety checks

- [x] No secrets, credentials, or live connector checks.
- [x] No new dependencies.
- [x] No generic raw write tools.
- [x] Reverse ETL flow untouched except existing `make verify` smoke flow.
- [x] Sensitive repository path blocking and content/download URL redaction retained.
