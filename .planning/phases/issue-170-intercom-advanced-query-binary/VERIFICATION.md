# Verification Checklist: Intercom Complete CLI Implementation (#166-#171)

## Targeted gates

- [ ] `go test ./cmd/connectorgen -run TestIntercomAPISurfaceFullCoverage -count=1`
- [ ] `go test ./internal/connectors/commandrunner -run 'TestIntercomDirectRead|TestIntercomWriteCommand' -count=1`
- [ ] `go test ./internal/cli -run 'TestConnectorCommandHelp|TestIntercomConnectorCommandHelp|TestRunMaybeConnectorCommand' -count=1`
- [ ] `go test ./internal/connectors/engine -run 'DirectRead|Write|Intercom|CLISurface' -count=1`
- [ ] `go test ./internal/connectors/conformance -run 'TestConformance/intercom' -count=1`
- [ ] `go run ./cmd/connectorgen validate internal/connectors/defs/intercom`
- [ ] `go run ./cmd/connectorgen validate internal/connectors/defs`

## CLI help/manual/website parity

- [ ] `go build ./cmd/pm`
- [ ] `./pm help connectors`
- [ ] `./pm intercom`
- [ ] `./pm intercom contact list --help`
- [ ] `./pm intercom contact view --help`
- [ ] `./pm intercom contact create --preview --json --config base_url=https://example.invalid --email test@example.com` does not execute live write and creates a plan/preview only when no credentialed check is needed.
- [ ] `rg -n "intercom|connector command|reverse ETL" docs/cli website/content/docs internal/connectors/defs/intercom/docs.md`

## Broad gates before handoff

- [ ] `gofmt -w cmd internal`
- [ ] `go vet ./...`
- [ ] `go test ./... -timeout=20m`
- [ ] `go build ./cmd/pm`
- [ ] `make verify`

## Safety verification

- [ ] No secrets printed/stored.
- [ ] No new dependencies.
- [ ] No credentialed Intercom checks.
- [ ] No generic raw HTTP/SQL/shell write surface.
- [ ] Reverse ETL writes stay plan -> preview -> approval -> execute.
- [ ] Destructive/admin writes declare risk and confirmation policy.


## Local verification: 2026-07-10

- [x] Targeted package tests for Intercom command surface/engine/commandrunner/connectorgen/conformance passed.
- [x] `go vet ./...` passed.
- [x] `go test ./... -timeout=20m` passed.
- [x] `go build ./cmd/pm` passed.
- [x] `make verify` passed.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs` passed with 547 connectors and 0 findings.
