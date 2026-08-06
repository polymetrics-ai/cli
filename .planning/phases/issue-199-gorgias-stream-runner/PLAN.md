# Plan: Gorgias Stream Runner

Parent issue: #196  
Sub-issue: #199  
Parent PR: https://github.com/polymetrics-ai/cli/pull/229  
Parent branch: `feat/196-gorgias-cli-parity`  
Sub-issue branch: `feat/199-gorgias-stream-runner`  
Connector: `gorgias`

## GSD command path

- Planning prompt: `scripts/gsd prompt plan-phase "Issue #199 Gorgias stream runner: implement safe ETL stream coverage for list/read-sweep endpoints after #200 operation ledger; no credentials, no writes, no raw tools"`.
- Manual GSD universal runtime fallback remains active because `programming-loop` is not available in `scripts/gsd list --json`.

## Required skills loaded

- `gsd-core`
- `caveman`
- `golang-how-to`
- `golang-cli`
- `golang-testing`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-context`
- `golang-concurrency`
- `golang-documentation`
- `golang-lint`

## Objective

Expand Gorgias read stream coverage from the four legacy streams to the safe list/read-sweep endpoints exposed by the #200 operation ledger, without enabling direct-read detail commands, binary payload handling, or write/reverse-ETL actions.

## Stream scope

Keep existing streams:

- `tickets` -> `GET /tickets`
- `customers` -> `GET /customers`
- `messages` -> `GET /messages`
- `satisfaction_surveys` -> `GET /satisfaction-surveys`

Add top-level list/search streams with no required path parameter:

- `account_settings` -> `GET /account/settings`
- `custom_fields` -> `GET /custom-fields`
- `events` -> `GET /events`
- `integrations` -> `GET /integrations`
- `jobs` -> `GET /jobs`
- `macros` -> `GET /macros`
- `metric_cards` -> `GET /metric-cards`
- `rules` -> `GET /rules`
- `tags` -> `GET /tags`
- `teams` -> `GET /teams`
- `users` -> `GET /users`
- `views` -> `GET /views`
- `voice_calls` -> `GET /phone/voice-calls`
- `voice_call_events` -> `GET /phone/voice-call-events`
- `widgets` -> `GET /widgets`

Add single-level fan-out sub-resource streams where a parent list endpoint supplies IDs:

- `customer_custom_fields` -> `GET /customers/{customer_id}/custom-fields`
- `ticket_custom_fields` -> `GET /tickets/{ticket_id}/custom-fields`
- `ticket_tags` -> `GET /tickets/{ticket_id}/tags`
- `ticket_messages` -> `GET /tickets/{ticket_id}/messages`
- `view_items` -> `GET /views/{view_id}/items`

Out of stream scope for this issue:

- Detail/singleton direct reads such as `GET /account`, `GET /customers/{id}`, `GET /tickets/{id}`, `GET /metric-cards/{slug}` (#201).
- Binary/file/recording endpoints (`/upload`, `/{file_type}/download/...`, `/stats/{name}/download`, `/phone/voice-call-recordings...`) (#202).
- POST query/report endpoints (`/search`, `/stats/{name}`, `/reporting/stats`) (#202).
- All mutations/writes (#203).

## Allowed files for this slice

- `internal/connectors/defs/gorgias/streams.json`
- `internal/connectors/defs/gorgias/schemas/*.json`
- `internal/connectors/defs/gorgias/fixtures/streams/**`
- `internal/connectors/defs/gorgias/api_surface.json`
- `internal/connectors/defs/gorgias/cli_surface.json` only if implemented stream commands need availability/api mapping updates
- `internal/connectors/defs/gorgias/docs.md`
- `internal/connectors/defs/gorgias/metadata.json` only for read-risk wording
- `cmd/connectorgen/gorgias_*_test.go` for static stream/ledger regression coverage
- `.planning/phases/issue-199-gorgias-stream-runner/**`
- `.planning/phases/issue-196-gorgias-cli-parity/**` only for parent orchestration updates after PR/checks

## Implementation slices

### Slice 1 — red stream sweep regression

1. Add a focused static Go test that expects 24 Gorgias streams and matching API-surface `covered_by.stream` rows.
2. Run the focused test and record the expected failure against the current 4-stream parent baseline.

### Slice 2 — declarative stream expansion

1. Add stream definitions using existing cursor pagination (`cursor`, `meta.next_cursor`) and page-size query where appropriate.
2. Add minimal schemas with `x-primary-key` and practical cursor fields (`updated_datetime`, `created_datetime`, or `id`) for each stream.
3. Add fixture pages for each new stream; keep fixture payloads tiny and non-secret.
4. Use `fan_out` only for single-level parent-child reads where the parent IDs are available from `customers`, `tickets`, or `views`.
5. Update `api_surface.json` coverage from blocked `direct_read` rows to `covered_by.stream` rows for the implemented stream endpoints.

### Slice 3 — docs/metadata/help-surface alignment

1. Update connector docs/read-risk wording to enumerate stream coverage without claiming direct-read or write parity.
2. Mark corresponding `cli_surface.json` list commands implemented only if the CLI metadata already defines those paths and they map directly to streams.
3. Keep writes, detail reads, binary payloads, and advanced POST queries blocked/planned.

## TDD strategy

- Red: stream sweep test fails because the current Gorgias bundle exposes 4 streams instead of 24.
- Green: streams/schemas/fixtures/API surface validate; Gorgias conformance passes; ledger metrics are updated to reflect the new covered stream rows.
- Refactor: schema/fixture naming and docs wording only; no Go runtime changes unless validation exposes a real engine gap.

## Verification checklist

Focused:

- [ ] `go test ./cmd/connectorgen -run 'Gorgias(APISurfaceOperationLedger|StreamRunner)'`
- [ ] `jq empty internal/connectors/defs/gorgias/streams.json internal/connectors/defs/gorgias/api_surface.json internal/connectors/defs/gorgias/cli_surface.json internal/connectors/defs/gorgias/schemas/*.json`
- [ ] `go run ./cmd/connectorgen validate internal/connectors/defs`
- [ ] `go test ./internal/connectors/conformance -run 'TestConformance/gorgias'`
- [ ] `git diff --check`

Broader before sub-PR handoff:

- [ ] `gofmt -w cmd internal`
- [ ] `go vet ./...`
- [ ] `go test ./...`
- [ ] `go build ./cmd/pm`
- [ ] `make verify`

## CLI help/docs/website parity

This slice changes connector metadata/streams only. Runtime help/docs renderer work remains #198. Any CLI surface metadata changes must avoid claiming write/direct-read/binary execution.

## Safety gates

- No secrets requested, printed, summarized, or stored.
- No credentialed Gorgias checks.
- No reverse ETL execution.
- No destructive/admin external actions.
- No new dependencies.
- No generic shell, generic HTTP write, generic SQL write, direct_write, raw_api, or raw mutation escape hatches.
