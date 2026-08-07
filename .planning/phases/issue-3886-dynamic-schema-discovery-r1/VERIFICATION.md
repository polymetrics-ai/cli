# Verification checklist — Issue 3886

## Focused gates

| Gate | Command | Result |
| --- | --- | --- |
| Discovery / catalog contract | `go test ./internal/connectors ./internal/connectors/discovery ./internal/connectors/engine -count=1` | passed |
| HubSpot fixture integration | `go test ./internal/connectors/native/hubspot ./internal/connectors/native/nativeset ./internal/connectors/bundleregistry -count=1` | passed |
| App + CLI catalog behavior | `go test ./internal/app ./internal/cli -count=1` | passed |
| Static conformance | `go test ./internal/connectors/conformance -count=1` | passed |
| Race-sensitive driver | `go test -race ./internal/connectors/discovery ./internal/connectors/native/hubspot -count=1` | passed |
| Format | `gofmt -l internal cmd` | passed (no output) |
| Vet | `go vet ./...` | passed |
| Build | `go build ./cmd/pm` | passed |
| Tidy | `make tidy-check` | passed; `go.mod` / `go.sum` unchanged |
| Lint | `make lint` | passed |
| Docs | `make docs-check` | passed |
| Smoke no build | `make smoke-no-build` | passed |
| Agent contract | `go run ./cmd/agentcontractgen check` | passed |
| Connector validation | `go run ./cmd/connectorgen validate internal/connectors/defs` | passed (550 connectors, 0 findings) |
| Connector surface sync | `go run ./cmd/connectorgen surface-sync --check` | passed (0 corrections) |
| Boundary | `make connector-boundary` | passed |
| Release workflow | `make release-workflow-check` | passed |

## Discovery safety assertions

| Assertion | Evidence | Result |
| --- | --- | --- |
| No secret/raw-credential/token-derived cache key | D8, D10 and source review | passed |
| No credential/error body output | D6, D10 and CLI JSON checks | passed |
| Work bounded and cancelable | D3, D4 | passed |
| Rate-limit backoff uses retry/jitter and context-aware sleep | D5, H4 | passed |
| Global failure and partial description are explicit | D7, H3 | passed |
| Stale data is visibly stale; refresh is explicit | D8, D9, A1–A3 | passed |
| Custom object unseen by source is callable | H1, H5 | passed |
| No new dependencies | `git diff -- go.mod go.sum` is empty | passed |

## CLI/help/docs/website parity

| Surface | Verification | Result |
| --- | --- | --- |
| Runtime help | `pm help catalog`, `pm catalog --help` | passed |
| Namespace behavior | `pm catalog`, `pm etl` exit successfully with contextual help | passed |
| Invalid action | `pm catalog invalid` retains usage failure | passed |
| JSON contract | catalog show/refresh tests assert `discovery.stale` | passed |
| CLI docs | `docs/cli/catalog.md` documents freshness, staleness and refresh | passed |
| Website | `website/content/docs/etl.mdx` mirrors refresh/staleness semantics | passed |
| Generated docs | generator/check updates only established generated outputs | passed; only HubSpot/catalog outputs retained |

## Live-provider evidence

**Not performed.** No HubSpot credential was supplied; fixture coverage is not
represented as a live run. The PR body will state this exact limitation.
