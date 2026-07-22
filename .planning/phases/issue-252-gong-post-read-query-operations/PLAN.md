# Plan: Gong typed POST read-query operation execution

Parent: #133
Issue: #252
Branch: `feat/133-gong-cli-parity`
Parent PR: https://github.com/polymetrics-ai/cli/pull/232 (draft)

## GSD command path

- `scripts/gsd doctor`: passed.
- `scripts/gsd prompt plan-phase 252 --skip-research --tdd`: generated `/tmp/gsd-plan-phase-252.prompt.md`.
- `scripts/gsd prompt programming-loop init --phase issue-133-gong-engine-shapes --dry-run`: unavailable (`unknown GSD command: programming-loop`); manual-GSD fallback active per AGENTS.md/GSD adapter.

## Required skills loaded

- `gsd-core`
- `golang-how-to`
- `golang-cli`
- `golang-testing`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-context`
- `golang-documentation`

## Objective

Implement a narrow, schema-gated operation direct-read path for POST read-query operations without exposing raw request bodies or generic HTTP commands.

## Scope

- Add connector/engine operation direct-read types or interfaces for `rest_read` operations.
- Permit executable POST only with `content_type: application/json`, `body_schema`, response `max_bytes`, and a supported output policy.
- Materialize body values from connector-authored fixed body defaults and typed CLI `body.*` flag mappings.
- Validate the request body against the operation schema before network send.
- Reuse existing runtime auth/base URL, error maps, response max bytes, and output redaction.
- Update Gong commands only where typed flags/defaults make the endpoint safe.

## Non-goals

- No `--body`, `--json`, arbitrary method/path, URL, or generic raw API command.
- No credentialed Gong checks.
- No unbounded arbitrary filter object execution.

## Implementation slices

1. Red tests for engine operation direct-read body construction/schema validation/redaction.
2. Red tests for commandrunner operation dispatch and `body.*` flag validation.
3. Red tests for `connectorgen validate` safety rules.
4. Implement small operation direct-reader and validator support.
5. Flip safe Gong POST read-query commands from `planned` to `implemented` only when typed inputs exist.


## Implementation result

- Added typed operation direct-read support for schema-gated `rest_read` GET/POST operations.
- Added commandrunner `body.*` flag mapping while keeping unknown/raw body flags blocked.
- Flipped only safe Gong POST read-query commands (`meetings integration-status`, `flows steps`, `flows prospects`) to implemented; broad arbitrary-filter POST reads remain planned.

## Verification

- `go test ./internal/connectors/engine -run 'OperationDirectRead|DirectRead' -count=1`
- `go test ./internal/connectors/commandrunner -run OperationDirectRead -count=1`
- `go test ./cmd/connectorgen -run 'Operation|Gong' -count=1`
- `go run ./cmd/connectorgen validate internal/connectors/defs`
- `go test ./internal/connectors/conformance -run 'TestConformance/gong|Static' -count=1`

## Safety

No secrets, no credentialed checks, no generic raw body/API surface, bounded responses, and CLI/docs parity for command availability changes.
