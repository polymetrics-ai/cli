# Plan — YouTube Analytics documented-operation parity resume

## Delivery record

- Phase: `youtube-analytics-parity-resume-r1`
- Scope owner: `internal/connectors/defs/youtube-analytics/`, generated artifacts derived solely from that bundle, and the focused `internal/connectors/engine/read.go` query-resolution repair with its regression tests
- GSD path: `scripts/gsd prompt programming-loop` is unavailable despite a healthy adapter; use the local `gsd-programming-loop` helper and this committed manual-GSD trace.
- Required skills: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-documentation`, `golang-design-patterns`, `golang-structs-interfaces`, and `no-mistakes`.
- Agent decision: `local_critical_path`; this task owns one tightly coupled connector bundle and the active runtime policy disallows proactive delegation.

## Tasks

1. Recover the preserved `fc727ac88` delta onto current main, inspect every changed path, and remove stale assertions about runtime foundations that have landed.
2. Build an official-provider operation and request-field evidence matrix for YouTube Analytics and Reporting APIs. Reconcile it with `operations.json`, `api_surface.json`, and `cli_surface.json`; use the landing shared citation format rather than modifying shared schemas or conventions.
3. Red-test the newly executable `media.download` command and its binary-download metadata; implement only the minimum bundle changes and fixtures to make it executable, then record green evidence.
4. Keep `reports.query` planned solely for issue #2985's typed provider-query foundation; verify it is not exposed as `provider_search` or `rest_write`.
5. Synchronize generated surface/manual/website output late, only for YouTube Analytics, then run the focused connector, commandrunner, CLI, vet, build, help, and website-data gates.
6. Update the phase summary, verification record, and explicit operation/citation accounting; commit the coherent slice and drive `$no-mistakes` from the committed branch.
7. Repair the shared `buildInitialQuery` boundary so typed `ReadRequest.Query` values resolve required stream query templates while missing values fail closed; reconcile the phase contract with all seven approval-gated `writes.json` mutations.

## Ownership guard

Do not change `internal/connectors/engine/write.go`, `internal/connectors/engine/schema/writes.schema.json`, `internal/connectors/engine/schema/operations.schema.json`, connector-definition validator code, or `docs/migration/conventions.md`. A requirement to touch one is a separately owned foundation task.
