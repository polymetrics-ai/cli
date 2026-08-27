# Context — issue 4364 deferred visibility/admission bridge

## Task Delivery Header

- Issue: Closes #4364 — bridge deferred source-operation visibility admission.
- Base branch: main (`origin/main` at `cf29d302c`).
- Merges into: main.
- Delivery: A pull request from `feat/4364-deferred-visibility-admission` to `main`, with terminal green CI, fresh exact-head independent Codex audit, and the visibility/preflight proofs below. The PR is not self-merged.
- Working branch: feat/4364-deferred-visibility-admission.
- Task: Materialize a closed, source-cited deferred declaration for each of the 1,910 genuine Batch 1 foundation-pending operations in the 4,341-record mapping manifest. Preserve one stable command path, lane, exact target, and one concrete foundation, while keeping typed executor artifacts absent until their later foundation PRs.
- Verification: Focused red/green/refactor Go tests; manifest reconciliation; generated-artifact checks; offline public-command preflight probes that record zero transport setup; connectorgen validation, admission, evidence, surface, docs, build, boundary, and scoped CI-equivalent gates.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| The authoritative manifest reconciles to 4,341 rows with 1,910 exact deferred rows | live | A checked-in generator test loads the manifest and generated catalogs and asserts the exact total and rollups; a missing or swapped row changes the count or identity set. |
| Every deferred source identity has one citation, location, lane, path, target, and concrete foundation | live | A reconciliation test compares each generated record to the manifest and rejects duplicate/missing identities, citation/location, generic targets, lanes, and multiple foundations. |
| Deferred commands fail at missing foundation before transport or credentials | live | An offline public command test uses a request-counting transport and asserts the exact typed foundation error with zero requests and no credential lookup. |
| No fake typed executor is created | live | Generator tests assert deferred rows are absent from `operations.json`, `streams.json`, and `writes.json` unless an independently real typed binding already exists; the changed bundle inventory is checked against this rule. |
| Negative policy and identity cases fail closed | fake | Hermetic malformed manifest/bundle fixtures mutate one invariant at a time. Provider I/O cannot prove a repository-owned validation rule and is explicitly forbidden. |
| Delete, reverse-ETL, binary, ETL, unsafe, and mutation rows remain visible | live | The reconciliation test groups source-cited deferred records by lane/semantic and asserts each manifest record has a generated discoverable declaration. |

## Decisions from discussion

- `docs/architecture/batch1-source-operation-mapping-manifest.json` is the source of truth. The historical 1,908-row matrix is never read by the generator and remains clearly marked stale.
- The manifest is source data imported from the preserved Batch 1 branch while this delivery branch is based only on current `origin/main` at `cf29d302c`.
- Deferred visibility is a declaration/runtime-preflight capability only. It never creates a generic HTTP executor, a raw JSON/body route, a fabricated action/stream/operation, or a credential-bound usability claim.
- A manifest record must name one actual foundation from the closed component vocabulary and preserve its source citation, exact method/path, lane, semantic, and stable command identity.
- Existing runnable commands remain governed by the real implemented preflight. This work reports the runnable delta independently and expects it may be zero.
- Pi isolated workflow roles are unavailable in this Codex runtime, and the assigned-worker instruction forbids dispatching agents. The generated GSD discussion/planning/execution/verification/review prompts are therefore performed inline and recorded here.

## Required skills used

`golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, and `golang-documentation`.

