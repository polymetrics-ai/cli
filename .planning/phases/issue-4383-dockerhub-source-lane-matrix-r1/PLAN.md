# Plan — #4383 Docker Hub source-to-seven-lane matrix

## Boundaries

- Target connector only: `internal/connectors/defs/dockerhub/` and `.planning/phases/issue-4383-dockerhub-source-lane-matrix-r1/`.
- Restore only absent source sidecars after proving their Batch R1 parent byte hashes and source-lock provider digest.
- Add a local Go reconciliation test and a connector-local `sources/dockerhub-source-lane-matrix.json`.
- Do not alter Docker Hub runtime declarations, shared mapping controls, engine, executor, generator, certification, credentials, provider transport, or live state.

## Manual GSD / TDD execution

The inline/manual GSD fallback is authorized by the single-worker task and unavailable compatible Pi roles. It follows `discuss-phase` → `plan-phase --tdd` → `execute-phase` → `verify-work` → `code-review` through these artifacts and the generated prompt record.

1. Verify current-main absence and exact Batch R1 source-sidecar SHA-256 bytes; restore only these immutable connector-local files.
2. Add a failing local Docker Hub source-lane contract test before creating the matrix. It must require the exact source-lock identity, all 54 source IDs once, source citations/facts, seven cells, source-only states, required ETL/sync and reverse-ETL dispositions, and artifact backlinks.
3. Materialize the smallest source-lock-bound matrix from retained source facts. Retain pagination, scope/path variables, media, and callback/cursor absence; distinguish source candidacy from execution.
4. Add red edge mutations for hidden rows, duplicate IDs, invalid or source-less backlinks, missing source facts, pageable/extractable ETL/sync cells, missing mutation direct-write/reverse-ETL cells, and count mismatch.
5. Run formatter, JSON checks, focused/race/full Docker Hub package tests, source/map validation checks, agent-contract validation, changed-path review, commit/push, and no-checkbox issue proof.

## Initial mapping policy

The 54 locked rows, not importer/certification/runtime admission, are the denominator. A source GET is a direct-read candidate only; a source mutation, including every DELETE, retains direct-write and reverse-ETL candidate cells without becoming executable. Source-published pagination or extractable JSON-family collection evidence independently requires ETL and sync candidate cells. Binary cells require source-published binary media; `application/scim+json` is a JSON-family source fact rather than a binary assertion. Callbacks, events, cursors, stable keys, provider idempotency, and runtime execution remain absent unless the locked source says otherwise.

No `implemented` or `missing_foundation` state is allowed in this Track A matrix. `mapped_unproven` means source-backed mapping candidacy only; `not_applicable` has the exact source citation and stable source-information reason.

## Required skills

`golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, and `golang-testing` were selected by repository routing for this connector-local Go validator. No CLI, concurrency, database, GraphQL, runtime, or shared schema code is in scope.
