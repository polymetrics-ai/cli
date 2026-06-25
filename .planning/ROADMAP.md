# Roadmap

## Phase: go-cli-mvp

Build a working Go CLI vertical slice for local ETL and reverse ETL using the architecture in `POLYMETRICS_GO_CLI_MONOLITH_PRD_ARCHITECTURE.md`.

Acceptance:

- `poly init` creates a usable project directory.
- `poly help` and `poly man` expose detailed docs.
- Credentials can be added from environment values and stored encrypted.
- A connection can sync sample data into a local JSONL warehouse.
- A reverse ETL plan can preview warehouse data and write approved mapped records to an outbox.
- Commands support JSON output for agent callers.
- `go test ./...` and `go build ./cmd/poly` pass.

