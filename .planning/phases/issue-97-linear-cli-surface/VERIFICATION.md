# Verification Checklist — Issue #97 Linear CLI surface metadata

Date: 2026-07-09

## Focused TDD gates

```bash
go test ./internal/connectors/engine -run TestLinearCLISurfaceMapsImplementedStreams -count=1
go run ./cmd/connectorgen validate internal/connectors/defs/linear --json
go test ./internal/connectors/engine ./internal/connectors/commandrunner -count=1
```

## Broader gates before handoff when feasible

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

## CLI parity checks

Metadata-only slice; issue #98 owns renderer/docs expansion.

- [x] `./pm help linear` — ran and returned `help topic "linear" not found`; expected for metadata-only slice because no runtime connector help topic/router is added here.
- [x] `./pm connectors --help` — ran; namespace help still exits successfully.
- [x] `./pm connectors inspect linear` — ran; existing connector manual renders Linear streams.
- [x] `./pm connectors inspect linear --json` — ran; metadata-only, no credentials read.
- [x] `./pm linear` — not applicable; no provider-style command router is implemented in #97.
- [x] `./pm linear --help` — not applicable; no provider-style command router is implemented in #97.
- [x] `docs/cli/**` — grep found no Linear command docs; deferred to #98 help/docs renderer.
- [x] `website/**` — regenerated generated connector data with Linear `cli_surface` metadata.
- [x] Generated help/manual artifacts — no runtime help/manual generator consumes this metadata in #97; website connector data regenerated.

## Current result

Focused #97 checks completed:

```bash
go test ./internal/connectors/engine -run TestLinearCLISurfaceMapsImplementedStreams -count=1
# pass

go run ./cmd/connectorgen validate internal/connectors/defs --json
# pass: 0 findings, 547 connectors checked

go test ./internal/connectors/engine ./internal/connectors/commandrunner -count=1
# pass

go test ./internal/connectors/conformance -run 'TestConformance/linear' -count=1
# pass

npm --prefix website run gen:website-data
# pass; regenerated connector website data

npm --prefix website run test:unit
# blocked: local website dependencies are not installed (`vitest: command not found`); no npm install run because new/dependency installation is human-gated for this slice
```

Note: `go run ./cmd/connectorgen validate internal/connectors/defs/linear --json` is not valid for this validator implementation because it expects a defs root and treats child directories as connector bundles.
