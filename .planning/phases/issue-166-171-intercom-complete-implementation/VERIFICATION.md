# Verification Checklist: Intercom Complete CLI Implementation (#166-#171)

## Targeted gates

- [x] `go test ./cmd/connectorgen -run TestIntercomAPISurfaceFullCoverage -count=1`
- [x] `go test ./internal/connectors/commandrunner -run 'TestRunImplementedIntercomJSONDirectReadCommand|TestBuildWriteCommand' -count=1`
- [x] `go test ./internal/cli -run 'TestIntercom(CommandSurface|ConnectorCommand)' -count=1`
- [x] `go test ./internal/cli -run 'TestIntercom|TestHelp' -count=1`
- [x] `go test ./internal/connectors/engine -run 'TestDirectRead(JSONResponse|TextResponse|BinaryMetadata)|TestWrite' -count=1`
- [x] `go test ./internal/connectors/conformance -run 'TestConformance/intercom' -count=1`
- [x] `go test ./cmd/connectorgen ./internal/connectors/engine ./internal/connectors/commandrunner ./internal/cli ./internal/connectors/conformance -count=1`
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs`
- [x] Connector-scoped validation covered by `go test ./internal/connectors/conformance -run 'TestConformance/intercom' -count=1`; `connectorgen validate` expects the defs root, not an individual connector subdir.

## CLI help/manual/website parity

- [x] `go build ./cmd/pm`
- [x] `./pm help intercom`
- [x] `./pm intercom`
- [x] `./pm intercom contact list --help`
- [x] `./pm intercom contact view --help`
- [x] `./pm intercom contact create --help`
- [x] `./pm intercom contact create --root "$SMOKE_DIR" --credential intercom-local --email test@example.com --preview --json` returned `ConnectorCommandWritePlan`; setup used a temp `pm init` root and a synthetic env-sourced Intercom credential with `base_url=https://example.invalid`, and no external Intercom call executed.
- [x] `rg -n "intercom|connector command|reverse ETL" docs/cli website/content/docs internal/connectors/defs/intercom/docs.md`
- [x] Added `docs/cli/intercom.md`.
- [x] Added `website/content/docs/intercom-cli-surface.mdx` and `website/content/docs/meta.json` entry.

## Broad gates before handoff

- [x] `gofmt -w cmd internal`
- [x] `go vet ./...`
- [x] `go test ./... -timeout=20m`
- [x] `go build ./cmd/pm`
- [x] `make verify`

## Safety verification

- [x] No secrets printed/stored; examples use env/stdin placeholders, and conformance `secret_redaction` passes.
- [x] No new dependencies.
- [x] No credentialed/live Intercom checks.
- [x] No generic raw HTTP/SQL/shell write surface.
- [x] Reverse ETL writes stay plan -> preview -> approval -> execute.
- [x] Destructive/admin writes declare risk and confirmation policy.
