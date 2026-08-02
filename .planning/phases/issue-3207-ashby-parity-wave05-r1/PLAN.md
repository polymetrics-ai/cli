# GSD Plan — issue-3207 Ashby parity wave05-r1

## Scope

Parent #3207 and children #3208-#3214: refresh the Ashby official OpenAPI inventory, make `internal/connectors/defs/ashby/**` and Ashby-owned native/hooks expose every documented operation exactly once, keep unsupported/future rows evidence-backed, regenerate Ashby docs/surfaces, run local gates, update issues once, and stop after a clean local commit.

## GSD path and fallback

- `scripts/gsd doctor`: ok.
- `scripts/gsd prompt plan-phase issue-3207-ashby-parity-wave05-r1 --skip-research`: generated `traces/gsd-plan-phase-prompt.md`.
- `scripts/gsd prompt programming-loop init --phase issue-3207-ashby-parity --dry-run`: unavailable (`unknown GSD command: programming-loop`). Manual GSD/TDD fallback used and recorded here per adapter policy.

## Required skills loaded

`gsd-core`, `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, `golang-context`, `golang-cli`, `golang-documentation`; refs read: required-skill routing, GSD Pi adapter, CLI help/docs/website parity, migration handoff/conventions/design.

## Inventory plan

1. Fetch the public Ashby developer reference page without credentials and extract the embedded OpenAPI 3.1 schema plus `webhooks` object.
2. Generate a deterministic parity inventory: all REST path operations plus all webhook operations, each with method/path/operationId/summary/source/category.
3. Reconcile current official counts against parent issue r2 counts; use current official source as truth and record any count delta in docs/issues.

## Implementation plan

1. Replace the Ashby quarantine 4-row ledger with a complete `operation_ledger_version: 1` `api_surface.json`.
2. Expand `streams.json`, schemas, fixtures, and Ashby native read routing for every list/changefeed operation that the current native connector can safely stream with fixed POST body pagination (`limit`/`cursor`) and fixture mode.
3. Add `operations.json` and `cli_surface.json` for bounded direct/search/binary metadata that can execute through the existing operation direct-read contract; keep blocked rows for unsupported binary uploads/webhooks when no safe executor exists.
4. Add `writes.json` for non-binary reverse-ETL REST mutations with closed request schemas, risk text, destructive confirmation on deletes/history-destructive operations, and no generic request escape hatches.
5. Update Ashby-owned native connector to source metadata/catalog/command surface from the bundle and delegate typed direct reads/writes to the engine while keeping Ashby POST-cursor reads native.
6. Update docs/certification/generated Ashby surfaces, with fixture-only certification truth.

## Safety constraints

- No provider credentials, live provider calls, or provider writes.
- No generic HTTP method/path/body/query, shell, file, SQL, or passthrough surfaces.
- Destructive actions remain reverse-ETL only and require plan -> preview -> approval -> execute plus typed confirmation where declared.
- Stay in Ashby defs/native/hooks/tests/fixtures/docs/generated Ashby surfaces and planning artifacts. The only registry touch is the Ashby factory entry so the native Ashby connector can expose its own generated Definition/Manifest/CommandSurface without changing other promoted connectors.

## Exit criteria

- Current official operation counts are documented.
- `go run ./cmd/connectorgen validate internal/connectors/defs` passes (the tool treats the argument as the defs root; see VERIFICATION.md for the documented child-path caveat).
- Ashby conformance and focused CLI/golden/docs checks pass.
- Build/boundary/diff/make verify pass or any blocker is documented with exact evidence.
- Parent and children receive one final gh-axi comment with truthful counts and verification.
- Clean commit on `fm/cli-ashby-parity-wave05-r1`; no push/PR/no-mistakes final pipeline.

## Review fix round 1

1. Withdraw inferred timestamp incremental metadata and reject `syncToken` input until `ashby-sync-token-checkpoint-foundation` provides an Ashby-owned persisted opaque-state seam.
2. Route native reads, checks, engine-backed direct reads, and engine-backed writes through one credential-safe Ashby success-envelope validator.
3. Close every modeled nested request object generated for writes while preserving explicitly documented map-valued schemas.
4. Reject repeated Ashby page cursors before another request can reuse the token.
5. Remove executable repeatable array flags from Ashby stream commands and record `connector-stream-repeatable-array-foundation` as the blocked capability; direct-read and write arrays retain their existing supported runner paths.

This review phase uses the existing manual GSD/TDD fallback because `scripts/gsd prompt programming-loop ...` is unavailable. Production changes remain confined to Ashby-owned native code, bundle generation inputs/outputs, tests, and phase evidence.
