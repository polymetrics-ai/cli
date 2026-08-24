# Issue #4347 — retained source artifact mirror plan

## GSD and skills

- Resolved commands: `scripts/gsd sources discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, and `code-review`; generated `discuss-phase 4347` and `plan-phase 4347 --tdd` prompts were reviewed inline.
- Manual execution is required because the task contract prohibits spawning GSD roles.
- Required skills loaded: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, and `golang-documentation`.
- CLI parity: `connectorgen source-import` is a developer command, not the `pm` user CLI. Update its own help and `docs/migration/conventions.md`; `pm` manual and website pages are not applicable.

## Plan

1. **Red — retained reader:** add tests before production code for a lock whose tracked artifact is read successfully despite an unreachable provider URL; missing file; incorrect bytes/digest; and a zip/gzip-shaped opaque bundle. Add Elasticsearch-drift and Zoom-404 regression fixtures that prove only an independently retained correct copy permits import.
2. **Green — foundation hook:** add strict path derivation and bounded, symlink-safe retained-artifact reads under the connector's `sources/artifacts/` directory. Verify each raw payload against the *preexisting lock* byte count and SHA-256; remove the cache/network fallback from normal command execution.
3. **Green — provenance contract:** introduce a checked-in per-connector artifact manifest that records retained file, source URL, retrieval timestamp, and license/terms field. It is provenance-only: it cannot alter or replace lock identity.
4. **Refactor and docs:** simplify source-import help and migration documentation to describe hermetic retained-source verification and the migration/recovery workflow. Keep old cache helper tests only where they cover isolated legacy functions; command behavior must never reach them.
5. **Artifact migration:** copy a lock before editing its `sources/` directory. Preserve every still-valid pin. Under **Firstmate's** narrow 2026-08-24 authorization, explicitly re-pin only a fresh response that is classified as a real provider document; record connector/artifact, old/new byte and SHA-256 identities, and retrieval date in the visible report. Never pin an error, redirect, login wall, or other non-document response; record it as irrecoverable.
6. **Verification/review:** run focused package tests followed by applicable formatting, generators, validation, surface-sync, boundary, docs, and build checks. Execute inline verify-work/code-review, disposition findings, commit/push, open the direct PR, and verify its API-reported base is `main`.

## Test matrix

| Behaviour | Happy path | Bad input | Edge/regression |
| --- | --- | --- | --- |
| Retained machine-readable spec | exact file imports with an uncontactable source URL | byte/digest mismatch rejected | absent retained file rejected without fallback |
| Rendered citation | exact opaque retained citation validates | malformed manifest/path rejected | historic provider URL may be gone without affecting retained verification |
| Zip/gzip bundle | exact archive bytes are returned and validated opaquely | archive bytes differ from lock | bundle path remains confined to connector sources |
| Elasticsearch drift | retained locked bytes import while simulated live body differs | mismatched live bytes never become a replacement | source-lock digest remains unchanged |
| Zoom 404 | retained locked accounts bytes import while simulated provider returns 404 | missing copy names the retained-file recovery requirement | no historical raw copy is falsely claimed recovered |

## Scope guard

The shared `cmd/connectorgen` change is a genuine foundation hook. Retained artifacts and manifests stay at the connector boundary. Firstmate inbox 004 explicitly keeps unmerged lane locks and artifacts out of this foundation branch: it retains current-main GitHub artifacts only. Full coverage is reached when Batch 8–10, Elasticsearch, and Zoom retain their own connector-boundary artifacts in their respective lanes.
