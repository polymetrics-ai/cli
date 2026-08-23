## Intent

Refs #4318

Deliver the shared additive v3 multi-document source-lock foundation. This PR does not add the
Zoom capture adapter or adopt/migrate a connector source lock.

## What changed

- Added exact v3 lock decoding for `rest.retrieval`, aggregate `rest.openapi`, and sorted
  document-owned `rest.source_documents`, while retaining v1/v2's existing single-document wire
  contract and schema-2 descriptor output.
- Fetches and verifies every v3 artifact independently, bounds the corpus, synchronizes duplicate
  content digests, and preserves per-document artifact plus published-source provenance.
- Uses locked document-qualified source IDs so repeated provider `operationId` values remain
  independently traceable; source projection and surface sync understand schema 3.
- Allows only bounded, non-secret query strings in non-fetched v3 citation URLs. Artifact and
  capture URLs remain queryless HTTPS retrieval URLs.
- Corrected the directly adjacent YAML bridge to normalize JSON-scalar mapping keys before strict
  duplicate detection: unquoted OpenAPI `200:` works, `200:` plus `'200':` fails closed.

## Red / Green / refactor evidence

- Red: v3 fixture lock failed strict legacy decoding at aggregate `rest.openapi`; YAML `200:` was
  rejected as non-string. Both commands and failure output are in
  `.planning/phases/issue-4318-multidoc-source-lock-foundation-r1/TDD-LEDGER.md`.
- Green: real hermetic source-lock/import tests cover multi-document provenance, duplicate provider
  IDs, missing/drifted documents, citation query policy, duplicate-digest synchronization, and
  YAML scalar key collision behavior.
- Refactor/freeze: v1/v2 retain their original route; frozen GitHub lock, descriptor, and combined
  ledger bytes/SHA-256 values are asserted in `TestSourceImportPreservesFrozenGitHubArtifacts`.

## Verification

- `make verify` — pass in full (tests, docs, smoke, lint: 0 issues, source/ledger generation,
  certification, boundary, release/install checks).
- `go test -race -count=1 -timeout 20m ./cmd/connectorgen -run '^TestSourceImportVersion3SynchronizesDuplicateArtifactDigests$'` — pass.
- `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestSourceImportPreservesFrozenGitHubArtifacts$'` — pass.
- `go run ./cmd/connectorgen source-import github --check` — pass (1,525 operations).
- `go run ./cmd/connectorgen surface-sync internal/connectors/defs --check` — pass (552 connectors, zero drift).
- `go run ./cmd/connectorgen certification-subject --check` — pass.
- `node scripts/github-combined-operation-ledger.mjs --check` — pass.

The frozen GitHub artifacts are byte-identical: source lock `3420025` /
`281b1cfcc67eb63e19ef83daf06197bf3d3b23db0b6bc9b73e02fc18ee278fb6`, descriptor `43354021` /
`d1978c0c6fd0eb66e9fcd4d78d637864a6e486f558aaad1e51550bc43758b899`, and ledger `2553169` /
`c4a904a919f30065fcc8453c6689e1a3dcc7be5ac8e11a7154d310b334972de3`.

## GSD lifecycle and skills

Resolved and executed inline: `discuss-phase`, `plan-phase --tdd`, `execute-phase`,
`verify-work`, and `code-review` through `scripts/gsd prompt ...`. Inline/manual fallback is
recorded because this repo's canonical single-worker contract forbids lifecycle role spawning in
this environment. Evidence: phase `PLAN.md`, `TDD-LEDGER.md`, `VERIFICATION.md`, `UAT.md`, and
`REVIEW.md`.

Skills used: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`,
`golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`,
`golang-context`, `golang-concurrency`, `golang-documentation`, `golang-naming`, and `golang-lint`.

## CLI/docs parity and safety

Updated developer-only `connectorgen source-import` help and
`docs/migration/conventions.md`. `pm` help/manual/website/completions are not applicable: this
does not alter the `pm` command surface or connector command schema.

No credentials, live provider writes, Zoom capture, Docker Hub artifact, GitHub definition
artifact, or `internal/connectors/defs/github/rate_limits.json` changed. Docker Hub's source
directory is not present on this base, so the YAML fix is exercised through hermetic real-import
fixtures rather than a connector capture.

## Checkpoints and automated review

One coherent direct-PR checkpoint contains the completed plan/TDD evidence, implementation, tests,
documentation, full verification, and manual standard-depth review. The base is `main`.

Automated review route: `claude_auto` did not create a workflow run after PR open, despite the
GitHub API reporting the author association as `MEMBER`. Per the repository fallback policy, one
`claude_manual` request has been made; its result is pending. No Copilot backup has been requested.
Any automated finding will be dispositioned before merge; final human approval remains required.
