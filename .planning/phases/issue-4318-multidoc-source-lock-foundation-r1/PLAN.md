# Issue 4318 — additive v3 multi-document source locks

## Task Delivery Header

- Issue: Refs #4318 — feat(connectorgen): add additive v3 multi-document source locks
- Base branch: main (`6410fe59c7ed9017dbe3f830f4361d4d015cd8e9` is an ancestor of `origin/main`, verified before implementation)
- Merges into: main
- Delivery: A direct pull request from `fm/cli-multidoc-source-lock-foundation-r1` to `main`, committed, pushed, locally verified, and API-verified for its exact base.
- Working branch: fm/cli-multidoc-source-lock-foundation-r1
- Task: Add a strictly versioned v3 source-lock/import/projection path for many REST documents without changing the v1/v2 wire contract or GitHub artifacts; no Zoom capture/adoption work belongs in this slice.
- Verification: TDD behavioral tests through the real importer; frozen GitHub byte/hash assertions; focused package tests; `go vet ./...`; `go build ./cmd/pm`; every repository `make verify` gate and full `make verify`; GSD verify/code-review prompts; PR-base API read-back.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| V3 documents import with document-owned provenance | live | The production importer parses hermetic HTTPS fixture artifacts and emits each operation with its locked `source_id`, document ID, artifact identity, and published identity; those fields would be absent/wrong without the v3 path. |
| V2 stays compatible | live | The checked-in GitHub v2 lock imports through the production path and the immutable lock, descriptor, and ledger bytes hash to the frozen values. |
| Duplicate provider operation IDs remain traceable | live | Two fixture documents use the same provider `operationId`; the generated descriptors retain it but have distinct locked source IDs. |
| Missing/mismatched source documents fail closed | live | The real document fetch/verification path reports missing input and a digest mismatch before descriptor output exists. |
| Query-bearing published citation is bounded safely | live | A v3 fixture accepts a valid citation query without fetching it; a credential-like or malformed query fails strict validation before retrieval. |
| Projection consumers understand v3 provenance | live | Source-projection and `surface-sync` accept a schema-3 descriptor paired with its v3 lock, and detect a changed document/published identity as drift. |
| OpenAPI numeric YAML keys retain their JSON meaning | live | The production source-lock importer accepts an unquoted `200:` response key as the same canonical member as `'200'`, rejects the pair as a duplicate, and still rejects compound keys. |

## Discussion decisions

- Keep v1/v2 decoding and import emission on their existing single-document path. GitHub remains schema-v2 and its checked-in source lock, descriptor, combined operation ledger, JavaScript generators, `rate_limits.json`, and certification subject are out of scope.
- V3 models `rest.source_documents`; each document owns a stable queryless imported artifact and published-source provenance. Aggregate v3 `openapi` inventory is a set/array only; each document records its own exact OpenAPI/Swagger version.
- A query is permitted only in `published_source.source_url`, which is a non-fetched citation. It must be HTTPS with no userinfo, fragment, controls, credential-like names, repeated keys, or oversized key/value/query data. Imported artifact and capture URLs must remain public, absolute, HTTPS, and queryless under the existing fetch guard.
- Use a bounded v3-only per-document path with deterministic document ordering and duplicate-digest synchronization. The v2 single-fetch sequence must not be routed through it.
- The locked v3 operation `id` is global descriptor `source_id`; its `operation_id` remains the provider exact value and may duplicate across documents. Every v3 raw method/path route remains unique across the corpus.
- Inline/manual GSD fallback: generated Pi prompts were resolved and enacted inline because no compatible isolated lifecycle-worker runtime is available and the canonical repository contract prohibits role spawning.
- The narrow Docker Hub-adjacent correction is inside the same YAML-to-JSON importer boundary: standard JSON scalar mapping keys (`!!int`, `!!float`, `!!bool`, `!!null`) become canonical JSON member names; YAML-only/custom tags and compound keys remain invalid. This leaves duplicate detection after normalization, so `200:` and `'200':` collide.

## TDD slices

1. **Red — v3 wire and inventory contract.** Add hermetic multi-document lock fixtures and behavioral importer tests for valid document-owned operation provenance, duplicate provider operation IDs, missing documents, digest mismatch, and safe/unsafe published URL query handling. Capture their failing output in `TDD-LEDGER.md` before implementation.
2. **Green — versioned parser/import core.** Add exact version dispatch, a normalized document context, strict v3 document/inventory validation, per-document cache/digest verification, and module-qualified source identity. Preserve the v2 route and re-run the red tests.
3. **Red — v3 downstream consumer contract.** Add source-projection and surface-sync tests requiring descriptor schema 3, all document/published provenance comparison, and v2/schema mismatch refusal.
4. **Green — descriptor/projection/surface synchronization.** Emit v3 provenance from import, rebuild expected provenance across documents, and accept only matching lock/descriptor versions while retaining schema-2 output unchanged.
5. **Refactor and freeze checks.** Add duplicate-digest synchronization/bounds only to the v3 path; update generator help and migration conventions. Freeze GitHub lock, descriptor, and ledger hashes in behavioral regression tests.
6. **Green — YAML scalar mapping keys.** Normalize standard JSON scalar mapping keys inside the pre-existing strict YAML adapter. Prove numeric/quoted equivalence through source-lock import, duplicate rejection after normalization, and compound-key rejection.
7. **Verify/review.** Run the complete local repository gates, execute generated verify-work/code-review prompts inline, resolve gaps if any, and document review dispositions.

## Planned files and boundaries

- Production: `cmd/connectorgen/sourceimport.go`, `cmd/connectorgen/sourceprojection.go`, `cmd/connectorgen/surfacesync.go`, and the source-import help text in `cmd/connectorgen/main.go` only as needed for plural-document accuracy.
- Tests: existing `cmd/connectorgen` source-import/projection/surface-sync tests and hermetic checked-in fixtures only.
- Documentation: `docs/migration/conventions.md` for the versioned multi-document/citation contract.
- Evidence: this phase directory. No change under `internal/connectors/defs/github/**`, especially `rate_limits.json`; no Zoom lock/descriptor/capture script; no certification subject.

## CLI help/manual/website parity

- `connectorgen source-import --help`: update/verify singular wording only if production help changes; its closed flag set remains unchanged.
- `pm` runtime help, bare namespaces, `docs/cli/**`, website docs, generated `pm` manual, and completions: not applicable because this modifies the developer-only `connectorgen` source-import contract, not the `pm` command surface.
- `docs/migration/conventions.md`: update with v1/v2 versus v3 source-lock rules and citation-only query policy.

## Skills loaded

`golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, `golang-documentation`, `golang-naming`, and `golang-lint`.

## Commit checkpoints

1. Planning/header and TDD ledger (`Refs #4318`).
2. Red behavioral tests and fixtures (`Refs #4318`).
3. Green v3 importer/projection/surface-sync/docs slice (`Refs #4318`).
4. Full verification, review evidence, and any review fixes (`Refs #4318`).
