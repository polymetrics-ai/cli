# Plan: Gong top-level JSON array request bodies

Parent: #133
Issue: #253
Branch: `feat/133-gong-cli-parity`
Parent PR: https://github.com/polymetrics-ai/cli/pull/232 (draft)

## GSD command path

- `scripts/gsd doctor`: passed.
- `scripts/gsd prompt plan-phase 253 --skip-research --tdd`: generated `/tmp/gsd-plan-phase-253.prompt.md`.
- `scripts/gsd prompt programming-loop init --phase issue-133-gong-engine-shapes --dry-run`: unavailable (`unknown GSD command: programming-loop`); manual-GSD fallback active.

## Required skills loaded

gsd-core, golang-how-to, golang-cli, golang-testing, golang-design-patterns, golang-structs-interfaces, golang-error-handling, golang-security, golang-safety, golang-context, golang-documentation.

## Objective

Support top-level JSON array request bodies for typed reverse-ETL write actions without raw JSON writes.

## Scope

- Add a schema-gated `json_array` write body mode or equivalent.
- Marshal a selected record field as the root JSON array.
- Validate the root array against the declared schema before network send.
- Keep reverse ETL plan -> preview -> approval -> execute.
- Redact content/path-like fields in previews and persisted command samples.

## Non-goals

- No generic `--json` or raw body command.
- No execution for broad `additionalProperties: true` array schemas without tighter connector-owned schema/input constraints.
- No credentialed Gong execution.

## Implementation slices

1. Red tests for root-array wire shape and schema failures.
2. Extend schema/loader/write action model for safe array body selection.
3. Update dry-run/preview redaction as needed.
4. Add validator rules for implemented array commands.
5. Keep Gong array operation blocked unless its schema/input contract is safe enough.


## Implementation result

- Added `body_type: json_array` write support with `body_field` and `body_schema` validation before network send.
- Added Gong CRM entity-schema write metadata for schema-gated top-level arrays without a raw JSON CLI body flag.

## Verification

- `go test ./internal/connectors/engine -run 'WriteJSONArray|Write' -count=1`
- `go test ./internal/connectors/commandrunner -run JSONArray -count=1`
- `go test ./cmd/connectorgen -run 'Operation|Gong' -count=1`
- `go run ./cmd/connectorgen validate internal/connectors/defs`
- `go test ./internal/connectors/conformance -run 'TestConformance/gong|Static' -count=1`
