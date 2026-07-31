Part of #277 (S1). Write scope: `metadata.json`, `spec.json`, `api_surface.json` under `internal/connectors/defs/twenty/`. Deps: none.

- `metadata.json`: connector id `twenty`, tier-1, provider display.
- `spec.json`: bearer auth via `x-secret` API key from env/stdin (never prompt text); `check` behavior.
- `api_surface.json`: one row per endpoint (56 read + 112 write = 168), each with an `execution_model` from the closed vocabulary and a non-empty `source_url` (https://docs.twenty.com/...). No partial/planned rows; every write verb has a `covered_by` target.

Acceptance: `go run ./cmd/connectorgen validate` classification + source-link gates clean; secret scan clean.
Refs #277