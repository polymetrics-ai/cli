# Plan — #4410 Sentry source-to-seven-lane matrix

## Boundaries

- Target connector only: `internal/connectors/defs/sentry/` and `.planning/phases/issue-4410-sentry-source-lane-matrix-r1/`.
- Restore only absent retained files under `internal/connectors/defs/sentry/sources/` after proving their Batch R1 parent byte hashes and source-lock provider digest.
- Do not alter Sentry runtime declarations, generated surfaces, #4365 route override, shared mapping controls, engine, executor, certification, credentials, provider transport, or live state.

## SE-R1-001 semantic repair

An independent review found that the initial matrix conflated collection extraction with continuation and mirrored ETL into `sync_transport`. This repair changes only the source matrix, its connector-local validator, and phase evidence. It must not edit the source lock, crosswalk, declaration disposition, importer, runtime, certification, or shared foundations.

The repair validator must derive lane truth from source-contract evidence rather than fixed HTTP-method, operation-ID, or schema-name allow-lists:

1. A direct read or mutation is identified from provider-published action language in the summary or operation ID plus a documented success response.
2. ETL requires a semantic read plus a query parameter whose provider description actually declares continuation/pagination; an array response, `per_page`, or a list-shaped name alone is insufficient.
3. Sync transport requires a mutation whose provider contract explicitly registers an event/webhook callback and requires both a callback URL and event-selector request field. A hook list or any pagination fact cannot satisfy it.
4. Binary lanes stay tied only to published request/response media.

The local tests must prove each positive and negative, retain every documented mutation/delete in `direct_write` plus `reverse_etl`, and preserve all source-row/backlink reconciliation.

## Manual GSD / TDD execution

The inline/manual GSD fallback is authorized by the single-worker task and unavailable compatible Pi roles. It follows `discuss-phase` → `plan-phase --tdd` → `execute-phase` → `verify-work` → `code-review` through these artifacts and the generated prompt record.

1. Inspect frozen Batch R1 source files and current main; record per-file blob/SHA-256 identity before copying only absent retained source material.
2. Add a local failing Sentry matrix contract test before creating the matrix. It must require the exact source lock identity, all source IDs once, source citations/facts, seven cells, source-only states, required ETL/reverse-ETL dispositions, and artifact backlinks.
3. Materialize a connector-local matrix from the retained source lock. Capture source-backed pagination, scope/path, media, and event/cursor facts. Keep `mapped_unproven` distinct from execution.
4. Add red edge mutations for hidden/duplicate rows, invalid artifact backlink, missing pageable ETL/sync state, missing write reverse-ETL state, and count mismatch.
5. Run formatter, JSON checks, focused/race/full Sentry package tests, relevant source/map checks, review changed paths, commit/push, and post no-checkbox proof.

## Initial mapping policy

The retained source must decide every classification. A bounded direct read is a provider action headline (summary, or operation ID when the source has no summary) with a documented success response; a mutation is the corresponding provider mutation action with a documented success response and must retain both direct-write and reverse-ETL cells. Neither selection is derived merely from an HTTP verb. Only provider-published media evidence can make a binary candidate. ETL requires both a semantic read and a source-documented continuation mechanism; an array response or page-size control alone is not continuation. Sync transport is independent of ETL: it requires a source-backed webhook/event-registration contract with a callback URL and event-selector request fact. A direct-read cell is never a substitute. Missing source information remains a cited source-information disposition; it is not a runtime foundation gap.

The frozen facts are 223 rows: 120 semantic reads, 103 semantic mutations including 35 DELETE actions, 45 source-described query continuations (43 cursor plus two SCIM `startIndex` facts), one `per_page`-only fact, 54 JSON-array GET response facts, and one source-backed webhook registration contract. Therefore 45 continuation reads are ETL candidates; the 17 JSON-array reads without continuation and the page-size-only read remain non-ETL. Two POST operations publish `multipart/form-data`; every published response media type is `application/json`; no operation publishes callbacks. The matrix binds 220 current `api_surface.json` records, three current stream records, and the existing Seer Models operation/command binding. Three lock rows have no exact current API-surface record and one current API-surface/stream route has no exact lock row, all recorded as source-information backlinks rather than omitted or synthesized identities.

Current-main `go run ./cmd/connectorgen validate internal/connectors/defs/sentry` rejects the restored v2 source lock at `source_operation` (`[source_projection] parse source lock: json: unknown field "source_operation"`). This is an out-of-scope shared source-projection compatibility gap; retain all 223 rows and do not modify the importer, generator, or runtime.

## Required skills

`golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, and `golang-testing` are required by the routing reference for this connector-local Go validator. No CLI, concurrency, database, or GraphQL runtime code is in scope.
