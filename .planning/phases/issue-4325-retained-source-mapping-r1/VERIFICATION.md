# Verification — retained-source mapping-only bridge

## Planned scoped checks

```text
go test -timeout 20m ./cmd/connectorgen -run '^TestRetainedSourceMapping' -count=1
go test -race -timeout 20m ./cmd/connectorgen -run '^TestRetainedSourceMapping' -count=1
go vet ./cmd/connectorgen
go run ./cmd/connectorgen source-operation-mapping-cohort data/connector-canon/batch1-source-operation-mapping-cohort.json --check
go run ./cmd/connectorgen retained-source-mapping bitbucket --check
go run ./cmd/agentcontractgen check
jq empty data/connector-canon/batch1-source-operation-mapping-cohort.json
git diff --check
```

## Observed scoped checks

| Check | Result |
| --- | --- |
| Red command registration | Expected failure: unknown `retained-source-mapping` subcommand before implementation. |
| Focused suite | `go test ./cmd/connectorgen -run '^TestRetainedSourceMapping' -count=1 -timeout 20m` — PASS (50.082s). |
| Focused race suite | `go test -race ./cmd/connectorgen -run '^TestRetainedSourceMapping' -count=1 -timeout 20m` — PASS (381.721s). |
| Vet | `go vet ./cmd/connectorgen` — PASS. |
| Cohort | `go run ./cmd/connectorgen source-operation-mapping-cohort data/connector-canon/batch1-source-operation-mapping-cohort.json --check` — PASS: 10 connectors, 4,341 source operations, 0 findings. |
| New command | `go run ./cmd/connectorgen retained-source-mapping bitbucket --check` — PASS: mapping-only, 297 source operations/297 verified, 0 executable declarations, 7 lanes. |
| Help | Root help and `retained-source-mapping --help` expose the developer command — PASS. |
| Agent contract | `go run ./cmd/agentcontractgen check` — PASS. |
| JSON / diff | `jq empty` for the cohort manifest and run state, plus `git diff --check` — PASS. |

The package-wide `go test ./cmd/connectorgen` was intentionally interrupted before completion after the parent narrowed this candidate to lightweight checks. It is not reported as a pass or a failure.

## Runtime boundary assertion

The new command's tests exercise source-lock parsing, structural in-memory source-contract/operation identity reconstruction, matrix decoding, and `EnabledConnectorContract` validation/reconciliation only. It does not invoke `runSourceImport`, schema/descriptor materialization, source projection, runtime bundle load, connector executor, credential, transport, or certification package.

Structural proof deliberately stops before schema resolution: Docker Hub's valid retained operation inventory demonstrates this boundary. Descriptor schema materialization may independently need a later runtime foundation, but that does not make a source row unmappable.

## CLI parity assertion

`go run ./cmd/connectorgen --help` and `go run ./cmd/connectorgen retained-source-mapping --help` must expose the developer command. No `pm` help/manual/website surface applies.
