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
- Parity-check red (2026-08-01 resume): independent parser over pinned Linear schema blob still matches `api_surface.json` at 164 Query + 371 Mutation + 80 Subscription = 615 unique rows with no missing/extra ledger paths, but command metadata had one stale duplicate planned `attachments list` direct command while `Query.attachments` is implemented as a stream, and 82 declared `writes.json` actions (54 destructive/delete-like) lacked canonical provider command entries. Fix by regenerating connector-owned `cli_surface.json`, removing the stale `operations.json` planned attachments row, and refreshing generated docs/goldens.
- Green slice 1: generated connector-local Linear inventory from pinned schema blob `3934265499c95f1d6b8e4d5c695ad0b6f1d52fec`; parser records 164 Query fields, 371 actual Mutation fields, and 80 Subscription fields (the earlier 376 mutation rough count included five description lines beginning with “Deletes”, not fields).
- Green slice 2: updated Linear bundle to 64 fixed GraphQL streams, 122 fixed GraphQL write actions, 93 destructive-confirmation actions, 615 operation-ledger endpoint rows, 429 blocked rows, 64 stream fixtures, and 122 write fixtures.
- Validation: `go run ./cmd/connectorgen validate internal/connectors/defs --json` passed with 0 findings.
- Validation: `go test ./internal/connectors/conformance -run 'TestConformance/linear' -count=1` passed.
- Parity validation: Linear CLI help/manual spot checks passed; generated CLI golden transcript was refreshed for the new Linear command surface; generated connector docs/catalog were refreshed for Linear only; `./pm docs validate --connectors-dir docs/connectors`, `go test ./...`, `go vet ./...`, `go build ./cmd/pm`, and `make verify` passed.
- Parity-check red (current-source follow-up): `gh-axi api` showed Linear current `packages/sdk/src/schema.graphql` blob `e92dc40c31e3b6e3962f93fa1d8cbe91f3e83034` at master commit `7ef4c5024f88667b2c85057ff4c905676c4a93c2`, superseding the older pinned blob by two root operations: `Query.partnerOfferWorkspaces` and `Mutation.partnerOfferRedeem`. Treat the two current-source omissions and mutable `master` source URLs as findings to fix, not waivers.
- Parity-check green (2026-08-01 resume): refreshed `cli_surface.json` to 188 commands (64 stream, 122 reverse-ETL write, 2 planned direct-read), removed the stale `linear.query.attachments` planned operation, and kept `Query.attachments` covered only by the implemented stream. Re-audit totals after current-source ledger refresh: current schema 165 Query + 372 Mutation + 80 Subscription = 617, `api_surface.json` 617 unique rows, 0 missing, 0 extra. Dispositions: 64 Query streams, 91 blocked Query direct reads, 10 blocked Query binary reads, 122 Mutation writes, 191 blocked admin reverse-ETL, 41 blocked sensitive reverse-ETL, 18 blocked destructive-action mutations, 80 blocked Subscriptions. Destructive/write command audit: 122/122 writes have canonical commands; 93/93 destructive actions require `confirm: "destructive"` and command approval text names typed destructive confirmation.
- Validation after parity-check fixes: `go run ./cmd/connectorgen validate internal/connectors/defs --json` (0 findings), `go test ./internal/connectors/conformance -run 'TestConformance/linear' -count=1`, `go run ./cmd/pm docs validate --connectors-dir docs/connectors`, `go test ./internal/cli -run TestGoldenTranscripts -count=1`, `go test ./...` (one initial concurrent run hit the default 10m certify package timeout; a focused `go test ./internal/connectors/certify -count=1 -timeout 20m` passed and the subsequent default `go test ./...` passed), `go vet ./...`, `go build ./cmd/pm`, and `make verify` all passed with isolated `GOCACHE=/tmp/linear-go-cache` where applicable.
