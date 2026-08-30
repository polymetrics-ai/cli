# Context — issue 4293 source-operation multi-lane manifest

## Task Delivery Header

- Issue: Refs #4293 — Batch R1 source-operation multi-lane manifest and validator.
- Base branch: fm/cli-top100-declaration-batch-r1
- Merges into: fm/cli-top100-declaration-batch-r1 → main.
- Delivery: Committed and pushed scoped branch, ready for the Batch R1 parent to integrate; no pull request is opened by this task.
- Working branch: fix/4293-source-operation-multilane-manifest-r1.
- Task: Add an authoring-only, source-lock-bound mapping manifest and deterministic checker. It must retain each lock operation as one cited source row, retain all seven lane-cell dispositions and classification facts, and reject membership or artifact-link regressions without changing runtime or certification behavior.
- Verification: Focused red/green tests in `./cmd/connectorgen`, `gofmt`, JSON/schema validation, `connectorgen` mapping and source/declaration checks, and `git diff --check`.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Locked source operations retain one unique cited row | live | A synthetic locked source set passes when every identity appears once; duplicate IDs and a removed row fail with named diagnostics. |
| All classification facts and seven lanes are representable | live | The fixture records pagination, scope/path, media, event/cursor evidence and all requested lanes with the four allowed states. |
| Pageable reads have an explicit ETL disposition | live | Removing the ETL cell from a cursor-paginated source fails even when other cells remain. |
| Artifact links cannot create identities | live | An artifact link to a nonexistent source/lane cell fails; the link is accepted only after the actual source row and cell exist. |
| Supplemental source lineage does not inflate a canonical denominator | live | Two locks for one connector produce three cited source rows but two canonical operations only when the supplemental row explicitly points to a same-route self-canonical source row. |
| Existing runtime and certification behavior stays separate | live | The checker reads only fixture source locks and mapping data; it performs no bundle runtime preflight, provider I/O, credential lookup, or certification loading. |

## Discussion record

- The authoritative denominator is each selected connector-owned source lock. Existing importer, declaration-admission, operation-evidence, and certification records are overlays and cannot suppress a row.
- The implementation reuses the Atlas `source.projection-admission.v1` owner seam: `parseDeclarationAdmissionSourceLock`, the existing `cmd/connectorgen` authoring command pattern, and engine JSON meta-schema validation.
- Multiple source locks may belong to one connector. `canonical_operation_id` is an explicit source-row representative, never route-based inference; the checker requires the referenced source row to be self-canonical and to have the same locked protocol/method/path (and GraphQL root field).
- No Atlas runtime gap exists. This is a mapping/admission control only; no executor, transport, credential, ETL, reverse-ETL, or sync implementation changes are allowed.
- Inline/manual GSD execution is used because this runner cannot use compatible isolated GSD roles and repository policy forbids role spawning.

## Skills and lifecycle

Loaded: `connector-lane-build-order`, `go-engineering` (fundamentals and agentic ETL guidance), `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, and the repository CLI-help/docs/website parity reference.

Resolved command path: `scripts/gsd doctor`, `scripts/gsd sources` for `discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, and `code-review`, plus generated inline prompts. The issue contract supplies the necessary decisions; the lifecycle is recorded here and in the TDD ledger.
