# PLAN — issue #180 Freshchat parity wave02 r2 (fixture evidence gap)

Branch: `fm/cli-freshchat-parity-wave02-r2`
Base: `9c977b208` → target `05a5a39c9` (PR #3679)

## Context

The Freshchat connector was already built and merged at 31/34 operations executable
(PR #3536, wave02-r1). The remaining r2 gap is **evidence, not implementation**:
12 of 18 streams had no `fixtures/streams/<name>/page_1.json`, so the conformance
harness's `read_fixture_nonempty` check silently `Skip`ped them instead of exercising
the real engine against recorded data.

## Scope (this change only)

- Add 12 replay fixtures under `internal/connectors/defs/freshchat/fixtures/streams/`:
  `agent_details`, `agent_statuses`, `business_hours_status`, `conversation_detail`,
  `conversation_fields`, `conversation_messages`, `historical_metrics`, `instant_metrics`,
  `outbound_messages`, `report_status`, `user_conversations`, `user_details`.
- No Go code, no shared runtime, no other connector touched.

## Why a GSD evidence gate fired

`scripts/verify-gsd-workflow` classifies any `internal/` path change (including these
fixtures) as an implementation change and requires at least one GSD/TDD planning-evidence
file to change alongside. This phase directory records that evidence.

## Verification plan

- `go run ./cmd/connectorgen validate internal/connectors/defs/freshchat` — expect 0 findings.
- `go test ./internal/connectors/conformance -run 'TestConformance/freshchat' -count=1` — expect pass.
- `gofmt`, `go vet`, `go build ./cmd/pm`, `make verify` as available.
