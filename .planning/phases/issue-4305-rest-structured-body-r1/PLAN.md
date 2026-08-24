# Issue #4305: Declaration-bound REST structured body

## Task Delivery Header

- Issue: Refs #4305 — feat(engine): materialize declaration-bound structured REST bodies
- Base branch: main
- Merges into: main
- Delivery: Committed task branch with all local gates green; Firstmate owns the later no-mistakes, push, and one PR stage.
- Working branch: fm/cli-rest-structured-body-r1
- Task: Add a closed source-declaration-owned structured JSON request-body path shared by generated REST CLI commands and typed reverse-ETL actions. It must support declared nested object and array fields, preserve path/query/header separation and approval gates, fail closed before I/O, preserve existing body strategies, and document mechanical downstream composition without editing production connector definitions. Add the approved shared write-query slice: a missing `record.*` query value may be omitted only when that exact source-locked `QueryParam` declares `omit_when_absent`; required, undeclared, wrong-source, malformed, and explicit invalid values remain pre-I/O failures.
- Verification: Red/green engine, commandrunner, and installed-CLI tests; go vet; pm binary build; generator validation and surface sync; generated help/manual checks; completion-tracked connector boundary; make verify; and configured code review.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Nested object and array values become the exact provider request. | live | An httptest provider records the JSON body and declared metadata, proving the request differs from a no-op or flattened input. |
| Route/query and body inputs cannot cross-bind. | live | A synthetic operation with separate path/query/body schemas rejects body field placement in route/query and leaves the transport counter at zero. |
| Actions cannot reuse fields from another action. | live | Two declared actions with disjoint body schemas reject the other action's field before their provider double sees a request. |
| Bad structured input fails before I/O. | live | Malformed JSON, unknown fields, absent required fields, wrong values, over-depth, over-item, over-byte, and request-metadata override cases retain a zero provider counter. |
| Approval is bound to the exact structured payload. | live | A previewed structured request fails confirmation after one nested value changes and the provider counter remains zero. |
| Generated command surfaces expose typed fields only. | live | Generated help/schema tests include declared field flags and omit raw-body/raw-json/template bypasses. |
| Existing body strategies keep their behavior. | live | Existing scalar, form, SCIM, binary, and specialized GitHub focused tests remain green without changed expected outcomes. |
| Downstream lanes can compose actions mechanically. | live | A connector authoring document names the declared-field mapping and shared executor contract; docs-check reads it. |

## Foundation check

| Need | Existing evidence | Plan |
| --- | --- | --- |
| Request shape ownership | REST operations already own routes, metadata, and content types. | Extend the operation schema and runtime with a declaration-only nested input projection. |
| Typed action reuse | Typed-action execution already has operation resolution and write gates. | Invoke the same materializer after action/command resolution rather than creating a reverse-ETL request builder. |
| Safe input boundary | Command runner validates flat flags before I/O. | Decode structured field values once, validate against the declaration recursively, then construct the request. |
| Limits | Existing direct-write/binary paths use bounded inputs. | Define explicit depth, field, collection, and byte caps before marshal/transport. |
| Generated discoverability | Command surfaces are synchronized from operations. | Synchronize typed structured fields and prove no opaque raw-body flag appears. |

## TDD slices

1. Red: install a synthetic nested-body command fixture and tests that demonstrate the current refusal, exact-body expectation, cross-binding, and zero-I/O failures. Green: admit only declared structured body specifications through a canonical materializer.
2. Red: add recursive schema tests for requiredness, type, object/array shape, unknown field, bounds, and metadata override. Green: enforce source/schema agreement and bounded recursive validation before I/O.
3. Red: add typed-action approval mutation and action-isolation tests. Green: share the canonical materializer and payload identity across CLI and typed-action paths.
4. Red: add generated help/schema and downstream-contract expectations. Green: synchronize generators/docs while keeping opaque raw-body inputs absent.
5. Refactor: preserve scalar/form/SCIM/binary/GitHub behavior with their existing focused regression suites; remove duplication only after all tests pass.
6. Red: a synthetic `WriteAction.Query` with `{{ record.optional }}` and `omit_when_absent:true` still rejects its missing record value. Green: add a write-only resolver boundary that omits that exact absence while retaining config, secret, incremental, required-record, wrong-source, malformed, and explicit-value behavior.
7. Red: a declared GraphQL direct-write error whose provider receipt echoes a configured base-URL value persists that value into plaintext `state.json`. Green: confine the persisted error-path projection so that plaintext state excludes the configured secret while the returned provider error's status, headers, and body remain verbatim. Do not add a general response-redaction layer or change CLI output.

## CLI help/manual/website parity

- [x] Existing connector-command help resolves and documents declared structured fields.
- [x] Bare connector namespace behavior remains unchanged; no namespace command is introduced.
- [x] Generated manual/schema artifacts are synchronized.
- [x] Documentation records JSON/approval/confirmation safety and no raw-body escape hatch.
- [x] Website documentation records the same declaration-bound write boundary.

## Lifecycle and skills

- Resolved GSD commands: discuss-phase, plan-phase --tdd, execute-phase, verify-work, code-review.
- Inline/manual fallback: issue #4305 is not in the stale roadmap and the task forbids role spawning.
- Loaded skills: golang-how-to, golang-cli, golang-testing, golang-error-handling, golang-security, golang-safety, golang-design-patterns, golang-structs-interfaces, and golang-documentation.
- CI remediation additionally used no-mistakes (active validation-step boundary only), golang-troubleshooting, golang-lint, and golang-continuous-integration; pipeline control remained with the outer executor.
- 2026-08-24 persistence remediation: Firstmate confirmed that `state.json` is plaintext JSON while the credential vault alone is AES-GCM encrypted. The GSD adapter prompts were executed inline because this issue is outside the active roadmap and the canonical contract forbids role spawning. The bounded follow-up uses golang-how-to, golang-testing, golang-error-handling, golang-security, golang-safety, golang-design-patterns, and golang-structs-interfaces.

## Commit checkpoints

1. Planning artifact and recorded red test evidence.
2. Closed materializer and validation green slice.
3. Typed-action/approval and generated-surface green slice.
4. Docs plus full local verification and review fixes.
