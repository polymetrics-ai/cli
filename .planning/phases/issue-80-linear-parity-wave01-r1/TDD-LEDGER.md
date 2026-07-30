# TDD ledger — issue #80 Linear parity wave01-r1

## Red / characterization before edits

- Current `internal/connectors/defs/linear` exposes 4 read-only streams, no writes, no operation ledger mode, no `operations.json`, no `cli_surface.json`, no `certification.json`.
- Existing docs classify all mutations as out-of-scope; captain policy now says destructive/delete operations are in scope when typed with destructive confirmation.
- Expected initial validation gap: `api_surface.json` lacks exact operation rows for the documented Linear GraphQL surface.

## Planned green checks

- `go run ./cmd/connectorgen validate internal/connectors/defs --json`
- `go test ./internal/connectors/conformance -run 'TestConformance/linear' -count=1`

## Evidence log

- Red/characterization: pre-existing bundle had only 4 read-only streams and no write/operation-ledger files.
- Green slice 1: generated connector-local Linear inventory from pinned schema blob `3934265499c95f1d6b8e4d5c695ad0b6f1d52fec`; parser records 164 Query fields, 371 actual Mutation fields, and 80 Subscription fields (the earlier 376 mutation rough count included five description lines beginning with “Deletes”, not fields).
- Green slice 2: updated Linear bundle to 64 fixed GraphQL streams, 122 fixed GraphQL write actions, 93 destructive-confirmation actions, 615 operation-ledger endpoint rows, 429 blocked rows, 64 stream fixtures, and 122 write fixtures.
- Validation: `go run ./cmd/connectorgen validate internal/connectors/defs --json` passed with 0 findings.
- Validation: `go test ./internal/connectors/conformance -run 'TestConformance/linear' -count=1` passed.
- Parity validation: Linear CLI help/manual spot checks passed; generated CLI golden transcript was refreshed for the new Linear command surface; generated connector docs/catalog were refreshed for Linear only; `./pm docs validate --connectors-dir docs/connectors`, `go test ./...`, `go vet ./...`, `go build ./cmd/pm`, and `make verify` passed.
