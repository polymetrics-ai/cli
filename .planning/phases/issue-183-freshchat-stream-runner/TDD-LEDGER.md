# TDD Ledger — Issue #183

## Red result

Added `TestFreshchatImplementedETLCommandsHaveReplayFixtures` in `internal/connectors/conformance`.

Initial command:

```bash
gofmt -w internal/connectors/conformance/freshchat_stream_test.go
go test ./internal/connectors/conformance -run TestFreshchatImplementedETLCommandsHaveReplayFixtures
```

Expected failure observed: missing replay fixture pages for Freshchat ETL streams declared in `cli_surface.json`:

- `agent_details`
- `agent_statuses`
- `business_hours_status`
- `conversation_detail`
- `conversation_fields`
- `conversation_messages`
- `historical_metrics`
- `instant_metrics`
- `outbound_messages`
- `report_status`
- `user_conversations`
- `user_details`

## Green result

Added replay fixtures for every implemented Freshchat ETL command stream and proved each emits at least one record through the real engine read path.

```bash
go test ./internal/connectors/conformance -run TestFreshchatImplementedETLCommandsHaveReplayFixtures
go test ./internal/connectors/conformance
go run ./cmd/connectorgen validate internal/connectors/defs
```

Results:

- Freshchat stream replay test: pass.
- `go test ./internal/connectors/conformance`: pass.
- `go run ./cmd/connectorgen validate internal/connectors/defs`: pass, `547 connector(s) checked, 0 findings`.

## Verification ledger

Full handoff gates:

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

Result: pass. `make verify` included docs validation, smoke, lint, and connectorgen validation; final standalone connectorgen validation reported `547 connector(s) checked, 0 findings`.
