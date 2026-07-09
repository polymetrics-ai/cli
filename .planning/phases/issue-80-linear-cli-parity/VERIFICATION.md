# Verification Checklist — Issue #80 Linear CLI parity parent

Date: 2026-07-09

## Required adapter checks

- [x] `scripts/gsd doctor` — pass.
- [x] `scripts/gsd verify-pi` — pass.
- [x] `scripts/gsd list --json` — pass; registry available.
- [x] `scripts/gsd prompt plan-phase issue-80-linear-cli-parity --skip-research` — generated prompt.
- [!] `scripts/gsd prompt programming-loop ...` — unavailable (`unknown GSD command: programming-loop`); manual-GSD fallback recorded.

## Focused Linear gates

```bash
go test ./internal/cli ./internal/connectors/engine -run 'TestLinear' -count=1
# pass

go test ./internal/cli ./internal/connectors/engine ./internal/connectors/conformance -run 'TestLinear|TestWriteGraphQL|TestConformance/linear' -count=1
# pass

go test ./internal/connectors/conformance -run 'TestConformance/linear' -count=1
# pass

go test ./cmd/connectorgen ./internal/connectors/commandrunner -count=1
# pass

go run ./cmd/connectorgen validate internal/connectors/defs --json
# pass: 0 findings, 547 connectors checked
```

## CLI help/docs/website parity checks

```bash
go run ./cmd/pm help docs
# pass; help reviewed before docs generation

go run ./cmd/pm docs generate --dir docs/cli --connectors-dir docs/connectors
# pass; retained scoped Linear connector generated docs

npm --prefix website run gen:website-data
# pass; Linear catalog/bundle data regenerated

./pm docs validate --connectors-dir docs/connectors
# pass

./pm connectors inspect linear
# pass; COMMAND SURFACE and write actions render

./pm connectors inspect linear --json
# pass; metadata-safe, no credentials read

./pm help linear
# pass; renders connector command-surface help

./pm linear --help
# pass; renders connector command-surface help
```

## Parent gates

```bash
gofmt -w cmd internal
# pass via make verify

go vet ./...
# pass

go test ./...
# pass

go build ./cmd/pm
# pass

make verify
# pass

git diff --check
# pass
```

## Website unit test note

`npm --prefix website run test:unit` remains a local-dependency blocker from the earlier slice (`vitest: command not found`); no dependency installation was performed.
