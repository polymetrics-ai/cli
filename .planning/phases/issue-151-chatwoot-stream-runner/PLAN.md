# Issue #151 — Chatwoot stream runner (CLI parity)

Refs #148 / #151.

## Required skills loaded

- `gsd-core` (repo-local workflow adapter; `scripts/gsd prompt programming-loop` unavailable, manual fallback recorded)
- `golang-how-to`
- `golang-cli`
- `golang-testing`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-structs-interfaces`
- `golang-context`
- `golang-documentation`
- CLI/help/docs/website parity reference (no command-path changes planned; generated connector docs/data may change if stream metadata changes)

## Scope

Implement and lock the Chatwoot ETL stream runner slice:

1. Preserve the seven implemented Chatwoot streams from #149/#150: conversations, contacts, inboxes, agents, teams, labels, and messages.
2. Correct message-stream pagination to the official Chatwoot `after` message-id cursor while keeping the fan-out parent conversation sweep page-number paginated.
3. Add a narrow declarative engine extension so `fan_out.ids_from.request` can declare its own pagination spec when the parent id-list endpoint and child stream endpoint use different pagination models.
4. Ensure PK/cursor metadata is explicit and validated for the Chatwoot streams.
5. Add full fixture-backed read sweep coverage proving every Chatwoot stream emits records through the real engine/conformance harness.
6. Regenerate connector docs/website data if generated stream metadata changes.

## Out of scope

- No new direct-read commands (#153 owns bounded point/query reads).
- No binary/file download/upload runner (#154 owns binary policy/engine work).
- No sensitive/admin/destructive write enablement (#155 owns policy gating).
- No credentialed/live Chatwoot calls; fixtures only.

## TDD plan

1. Red: add an engine fan-out test where parent id listing uses page-number pagination and child reads use cursor `after` pagination.
2. Red: add Chatwoot stream sweep tests asserting all seven stream fixtures pass `read_fixture_nonempty`, messages use `id` as cursor/request cursor, and the message fan-out parent request keeps page-number pagination.
3. Green: extend `FanOutIDsRequest` with optional `pagination` support and schema validation.
4. Green: update Chatwoot `streams.json`, message schema cursor metadata, and message fixtures.
5. Regenerate Chatwoot docs/website connector data if output changes.

## Verification checklist

- `go test ./internal/connectors/engine -run FanOut -count=1`
- `go test ./internal/connectors/conformance -run 'TestChatwoot' -count=1`
- `go test ./internal/connectors/conformance -run 'TestConformance/chatwoot' -count=1`
- `go test ./cmd/connectorgen -run Chatwoot -count=1`
- `go run ./cmd/connectorgen validate internal/connectors/defs`
- `npm test -- --runInBand` and/or targeted website connector tests if generated website data changes
- `git diff --check`
- Handoff gates before PR: `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, `make verify`

## Manual GSD fallback

`2026-07-10`: `scripts/gsd prompt programming-loop ...` and `scripts/gsd prompt quick-validate ...` both returned `scripts/gsd: unknown GSD command`, so this issue uses the repo planning/TDD/verification artifacts as the manual GSD fallback. `scripts/gsd list --json` and `scripts/gsd doctor` outputs were captured under `traces/`.
