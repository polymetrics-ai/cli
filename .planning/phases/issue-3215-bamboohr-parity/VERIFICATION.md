# Verification Checklist — issue #3215 BambooHR parity

## Focused required gates

```bash
# pass
go test ./cmd/connectorgen -run TestBambooHRSurfaceTracksCurrentOfficialInventory -count=1

# pass: connectorgen validate: 550 connector(s) checked, 0 findings
go run ./cmd/connectorgen validate

# pass
go test ./internal/connectors/conformance -run 'TestConformance/bamboo-hr' -count=1

# pass
go test ./cmd/connectorgen -run 'BambooHR|CLISurface|APISurface|Connector|Gong|GitHub' -count=1

# pass
go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1 -timeout=8m

# pass
go vet ./internal/connectors/... ./internal/cli/...

# pass
go build ./cmd/pm

# pass: outcome=clean, 0 findings; existing non-BambooHR exceptions only
make connector-boundary

# pass
git diff --check
```

## CLI/help/docs parity probes

```bash
# pass: 117-line connector manual
./pm help connectors

# pass: bare namespace renders contextual help and exits successfully
./pm connectors

# pass: JSON manifest includes BambooHR connector manual/manifest without secrets
./pm connectors inspect bamboo-hr --json

# pass: provider-style BambooHR command surface renders command groups and safety topics
./pm bamboo-hr --help

# pass: direct-read leaf help renders fixed operation/output policy, no generic passthrough
./pm bamboo-hr custom report request --help
./pm bamboo-hr data from dataset get-v1 --help

# pass: BambooHR docs/defs/website generated data include updated terms/counts
rg -n "BambooHR|bamboo-hr|direct_read|binary_read|operation-ledger" docs/connectors/bamboo-hr internal/connectors/defs/bamboo-hr website/data/connectors.generated.json website/lib/connectors.catalog.data.generated.json

# pass
./pm docs validate --connectors-dir docs/connectors
```

## Official inventory check

```text
Official BambooHR OpenAPI operations: 316 (310 path operations + 6 top-level webhooks)
Local api_surface rows: 316
Missing official rows: 0
Stale local extra rows: 0
Duplicate method/path rows: 0
```

Final local coverage:

| total | streams | direct reads | writes | blocked/planned | excluded |
| ---: | ---: | ---: | ---: | ---: | ---: |
| 316 | 138 | 9 | 149 | 20 | 0 |

Blocked models:

| model | count | rationale |
| --- | ---: | --- |
| `binary_read` | 6 | binary/download/export responses need bounded binary output policy |
| `admin_reverse_etl` | 7 | multipart/file/form-login mutations need connector-owned payload/approval support |
| `disallowed` | 1 | login/form operation is not exposed through connector commands |
| `local_workflow` | 6 | top-level webhook deliveries need inbound receiver/CDC/durable state workflow |

## Broad gate

```bash
# pass on final pipefail run after focused cold-cache timeout triage
make verify
```

Notes:
- An earlier cold-cache `make verify` reached the Makefile `go test -timeout 20m ./...` timeout while `internal/cli` consumed 978s and `internal/connectors/certify` was still running. Focused reruns of `internal/connectors/certify` passed, and the final pipefail `make verify` run passed.
- `make verify` includes `gofmt -w cmd internal`, `go mod tidy`/tidy diff, `go vet ./...`, `go test -timeout 20m ./...`, `go build ./cmd/pm`, docs validate, smoke, golangci-lint, connectorgen validate, connector-boundary, and release workflow check.

## no-mistakes follow-up gates

Pending after review-finding fix commit:

```bash
no-mistakes axi run --intent "Complete documented BambooHR connector parity ..."
```
