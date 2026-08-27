# Closed-schema composition foundation — context

## Task Delivery Header

- Issue: Closes #4366 — feat(connectorgen): add closed-schema composition foundation
- Base branch: main
- Merges into: main
- Delivery: Pull request open against `main`, with the direct-PR verification and review evidence recorded.
- Working branch: fm/cli-closed-schema-composition-foundation-r1
- Task: Add a shared, closed composition conversion for source-cited OpenAPI `oneOf`, `anyOf`, and `allOf`; retain every Batch 1 composition record's source identity and defer it unless an independently proven lane can execute it.
- Verification: Focused red/green importer, projection, engine-schema, deferred-preflight, operation-evidence, and reconciliation tests; source-import/check, validate, surface-sync, generated-artifact checks, an isolated built-binary credential-boundary proof for any newly runnable command, direct-PR gates, and review/re-review evidence.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Composition conversion preserves closed provider semantics | live | Source-import and engine tests assert the projected schema and validation result; the old converter returns a typed-gap error instead. |
| Invalid composition never reaches provider I/O | live | A request spy remains at zero calls after malformed, ambiguous, contradictory, cyclic, or external-reference input is rejected. |
| Composition records retain source provenance and stable targets | live | The reconciliation test reads the checked-in authoritative Batch 1 manifest and compares all 608 source method/path/location/citation records against their projected disposition. |
| No execution lane is fabricated | live | Declaration-admission and operation-evidence tests assert an unbound source operation remains `missing_foundation`, with no operations/writes/streams/binary binding. |
| Newly runnable commands are honestly credential-bound | live | For every nonzero promoted-command list, an isolated `pm` invocation reaches exactly `error: missing --credential` before provider I/O; no fake is used. |

## Decisions fixed by the brief

- Use the existing importer reference resolver only for reachable local references; external, unresolved, and cyclic references stay rejected or deferred before I/O.
- Keep source-native IDs, provider citation, location, method/path, and connector-relative target from `sourceOperationDescriptor`; conversion never derives them from a generated schema.
- The common engine schema dialect must model `oneOf`, `anyOf`, and `allOf` directly. It must not flatten a union, accept an open object, or infer fields.
- A schema conversion is not a lane admission. Existing declaration admission, operation evidence, and commandrunner preflight remain the authority for the six lanes.
- This foundation intentionally does not materialize an `operations.json`, `writes.json`, `streams.json`, or binary executor entry for a deferred source row.

## Required skills and workflow

- Loaded: `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, `golang-cli`, and `golang-documentation`.
- GSD lifecycle: `scripts/gsd prompt discuss-phase 4366`, `scripts/gsd prompt plan-phase 4366 --tdd`, then execute/verify/review prompts. The Pi adapter has no compatible isolated role runtime for this direct-PR worker, so the approved inline/manual fallback is used and recorded in these artifacts.
- CLI parity: this must not add a generic command. If an existing source-cited command gains a field or changes help, verify runtime help, bare namespace behavior when applicable, docs/website/generated artifacts, and record a not-applicable rationale otherwise.
