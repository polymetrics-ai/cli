# Issue #4421 Vercel semantic lane mapping repair

## Task Delivery Header

- Issue: Refs #4421 — Vercel source-to-seven-lane matrix
- Base branch: `feat/4421-vercel-track-a-matrix-r1` at immutable review target `e4b8a4da8795c30ea4e6fd948bd98cd116b8d043`
- Merges into: repair branch → `feat/4421-vercel-track-a-matrix-r1` → Batch R1 parent → `main` (all integration and `main` merge are outside this task and human-gated)
- Delivery: committed and pushed repair branch plus issue proof comment requesting fresh independent review; no PR, rebase, cherry-pick, integration, or merge in this task
- Working branch: `fix/4421-vercel-semantic-repair-r1`
- Task: Correct only Vercel’s source-lane matrix and its connector-local regression test so source semantics, not HTTP verb alone, classify bounded reads and binary media. Keep source lock, crosswalk, Atlas, shared mapping controls, runtime, transport, and certification untouched.
- Verification: targeted source-lane test; relevant connectorgen/connector checks; Go race and vet; JSON parse; agent-contract check; diff scope and no-source-lock-change checks.

## Immutable source ledger

| Item | Evidence | Rule |
| --- | --- | --- |
| Provider lock | `internal/connectors/defs/vercel/sources/vercel-operation-source-lock.json` | 400 rows; byte-identical to frozen target. It is read only in this repair. |
| Matrix | `internal/connectors/defs/vercel/sources/vercel-source-lane-matrix.json` | The only provider-facing mapping artifact changed. |
| Crosswalk | `internal/connectors/defs/vercel/sources/vercel-operation-crosswalk.json` | Read only; existing backlink destination. |
| Foundation Atlas | `docs/connector-canon/foundations/catalog.json` | Read-only reuse result: no runtime gap is created by this mapping correction. |

## Fixed decisions

1. A source operation’s HTTP method is evidence, not its complete semantic classification.
2. Existing `GET` direct-read coverage stays intact. A non-GET direct read needs retained source evidence of bounded read behavior: a successful response plus either a source-declared read/query summary without state-changing wording, or an explicit source statement that the request is equivalent to a GET read.
3. `artifactExists` is a bounded direct read because its retained description says the `HEAD` response is equivalent to a GET request without a body.
4. `artifactQuery` and `readSessionFile` are bounded direct reads because their retained operation summaries/descriptions say query/read and their successful responses are retained. They are not direct-write/reverse-ETL mutation candidates.
5. `readSessionFile` is also a binary download because the retained successful `200` response has `application/octet-stream` and a binary schema.
6. `writeSessionFiles` is a binary upload because its retained provider description expressly requires `Content-Type: application/gzip` for a gzipped tarball. The matrix preserves that exact source media fact even though the source lock’s structured `requestBody` is absent.
7. Normal POST mutations remain write/reverse-ETL candidates; textual read/query rules must not promote a create/update/delete/upload POST.
8. ETL and sync-transport statuses remain exactly source-evidenced by existing pagination/webhook facts. No runtime claims are added.

## Red → Green → Refactor plan

### Slice 1 — red semantic expectations

Add connector-local expectations covering source-cited `HEAD`/`POST` reads and binary media. The frozen matrix should fail because it presently rejects each non-GET direct read and both binary additions.

### Slice 2 — green source-semantic classification

Refactor the Vercel test helper from fixed IDs/method-only classification to retained source facts:

- semantic read title/description and successful response checks;
- explicit GET-equivalence for HEAD;
- documented structured and explicit `Content-Type` media extraction;
- binary schema/media checks; and
- mutation exclusion where an operation is source-declared as a read.

Update only the four affected matrix rows and summary counters. Preserve source-lock/crosswalk backlinks and all unrelated cells.

### Slice 3 — regression and scope proof

Add happy, bad, and edge cases for semantic POST/HEAD read, octet/gzip binary lanes, normal mutation rejection, missing success/media evidence, and contradictory mutation wording. Run the declared local checks and record exact results.

## Required skills and lifecycle

- Loaded: `connector-lane-build-order`, `go-engineering`, `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, and `golang-testing`.
- CodeGraph: absent from the frozen Vercel worktree; source was located with repository tooling only after recording the absence.
- GSD: `scripts/gsd doctor`, `scripts/gsd sources` and generated prompts for `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review` were run. The Pi runtime is unavailable in this isolated Codex worker and project policy forbids role spawning, so the documented inline/manual fallback is used and recorded in this phase directory.

## Non-goals

- No shared MIME admission, importer, runtime, transport, engine, certification, or Atlas modification.
- No source-lock, descriptor, or crosswalk rewrite.
- No direct endpoint implementation or connector execution claim.
- No PR, integration, or merge.
