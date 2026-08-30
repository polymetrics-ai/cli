# Plan — #4410 Sentry source-to-seven-lane matrix

## Boundaries

- Target connector only: `internal/connectors/defs/sentry/` and `.planning/phases/issue-4410-sentry-source-lane-matrix-r1/`.
- Restore only absent retained files under `internal/connectors/defs/sentry/sources/` after proving their Batch R1 parent byte hashes and source-lock provider digest.
- Do not alter Sentry runtime declarations, generated surfaces, #4365 route override, shared mapping controls, engine, executor, certification, credentials, provider transport, or live state.

## Manual GSD / TDD execution

The inline/manual GSD fallback is authorized by the single-worker task and unavailable compatible Pi roles. It follows `discuss-phase` → `plan-phase --tdd` → `execute-phase` → `verify-work` → `code-review` through these artifacts and the generated prompt record.

1. Inspect frozen Batch R1 source files and current main; record per-file blob/SHA-256 identity before copying only absent retained source material.
2. Add a local failing Sentry matrix contract test before creating the matrix. It must require the exact source lock identity, all source IDs once, source citations/facts, seven cells, source-only states, required ETL/reverse-ETL dispositions, and artifact backlinks.
3. Materialize a connector-local matrix from the retained source lock. Capture source-backed pagination, scope/path, media, and event/cursor facts. Keep `mapped_unproven` distinct from execution.
4. Add red edge mutations for hidden/duplicate rows, invalid artifact backlink, missing pageable ETL/sync state, missing write reverse-ETL state, and count mismatch.
5. Run formatter, JSON checks, focused/race/full Sentry package tests, relevant source/map checks, review changed paths, commit/push, and post no-checkbox proof.

## Initial mapping policy

The retained source must decide every classification. A GET operation may be a source-only direct-read candidate; a provider mutation must retain direct-write and reverse-ETL cells, but none thereby becomes executable. Only provider-published media evidence can make a binary candidate. A source-backed ETL/sync candidate is either a documented pagination parameter or a GET with a JSON-array response; a direct-read cell is never a substitute. Missing source information remains a cited source-information disposition; it is not a runtime foundation gap.

The frozen facts are 223 rows: 120 GET, 103 mutations including 35 DELETE, 43 cursor facts, one `per_page`-only fact, 54 JSON-array GET response facts, and 61 unique pageable-or-extractable ETL/sync candidates. Two POST operations publish `multipart/form-data`; every published response media type is `application/json`; no operation publishes callbacks. The matrix binds 220 current `api_surface.json` records, three current stream records, and the existing Seer Models operation/command binding. Three lock rows have no exact current API-surface record and one current API-surface/stream route has no exact lock row, all recorded as source-information backlinks rather than omitted or synthesized identities.

Current-main `go run ./cmd/connectorgen validate internal/connectors/defs/sentry` rejects the restored v2 source lock at `source_operation` (`[source_projection] parse source lock: json: unknown field "source_operation"`). This is an out-of-scope shared source-projection compatibility gap; retain all 223 rows and do not modify the importer, generator, or runtime.

## Required skills

`golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, and `golang-testing` are required by the routing reference for this connector-local Go validator. No CLI, concurrency, database, or GraphQL runtime code is in scope.
