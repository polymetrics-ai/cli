# Native Go All Connectors Test Plan

## Unit Tests

- Catalog counts remain 647 total, 591 sources, 56 destinations.
- All catalog entries are enabled and have baseline native capabilities.
- Registry binds every catalog slug.
- Native conformance reports pass for every catalog entry.
- Fixture source read, destination write, query, and CDC paths work.

## CLI Tests

- `pm connectors list --all --json` includes enabled runtime metadata.
- `pm connectors inspect destination-postgres` shows enabled write/query capability.
- `pm etl check/catalog/read --connector source-stripe` works in fixture mode.
- `pm reverse plan/run` works for `destination-postgres` after approval and writes a receipt.

## Verification

- `go test ./...`
- `go build ./cmd/pm`
- `./pm docs validate --connectors-dir docs/connectors`
- `make verify`
