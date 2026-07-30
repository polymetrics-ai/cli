# Summary: Issue 2985 Provider Search/Query Foundation

Status: implementation verified locally; awaiting commit/rebase/no-mistakes.

## What changed

- Added additive `provider_search` and `provider_query` capability metadata distinct from existing warehouse `query`.
- Added typed `provider_search` / `provider_query` operation contracts under `operations.json` with request schema, response schema, positive bounds, pagination, output policy, and fixture seam requirements.
- Added loader/connectorgen/conformance validation to reject metadata-only provider capability enablement, mismatched CLI intents, missing typed provider contracts, unsafe raw SQL/GraphQL/HTTP/payload request fields, and unbounded provider operations.
- Exposed provider operation descriptors through connector `Definition` and `Manifest` synthesis.
- Kept runtime fail-closed: provider operation commands are blocked until a dedicated executor lands.
- Updated help/docs/website copy to keep `pm query` local warehouse-only and provider search/query separate.

## Local verification

- `go test ./cmd/connectorgen ./internal/connectors/engine ./internal/connectors/commandrunner ./internal/connectors ./internal/cli` — pass.
- `go run ./cmd/connectorgen validate internal/connectors/defs --json` — pass, 548 checked, zero findings/warnings.
- `go build ./cmd/pm` — pass.
- `make connector-boundary` — pass, zero findings.

## Notes

Website typecheck could not run because this worktree lacks installed website dependencies (`tsc: command not found`).
