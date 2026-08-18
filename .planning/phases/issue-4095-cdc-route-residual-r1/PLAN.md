# Plan — Issue 4095 residual: live PostgreSQL CDC route and refusal matrix

## Task Delivery Header

- Issue: Refs #4095 — PostgreSQL CDC delete source-to-target route residual.
- Base branch: `integration/4015-mvp-flat-r1` (verified at `ff6a8710199c10f209d9d47cce87e5c8f7c429e6` before production edits).
- Merges into: `integration/4015-mvp-flat-r1 → main`.
- Delivery: Pull request open against `integration/4015-mvp-flat-r1`, backed by local live database proof and the repository verification gates.
- Working branch: `fm/cli-4095-cdc-route-residual-r1`.
- Task: Add the single real pgoutput PostgreSQL-to-PostgreSQL route proof and the complete declared pre-I/O refusal matrix; record R3 as non-executable without introducing an API destination.
- Verification: Targeted and tagged live PostgreSQL tests; application refusal tests; consumer `./cmd/connectorgen` tests; generation, lint, build, and repository no-build gates.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Real R4 pgoutput route covers insert/update/delete | live | A real logical replication stream produces source events; independent target SQL read-back finds the final keyed/history state and a replay/restart remains idempotent. |
| A durable receipt precedes the replication LSN acknowledgement | live | The test reads receipt/checkpoint state and the replication slot acknowledgement ordering after the source transaction. |
| R1/R2 and every declared destination `change_capture` cell refuse before I/O | fake | Deterministic application fixtures count source and target calls while asserting the concrete typed pre-I/O error for each declared cell; real provider I/O cannot be safely triggered for a refusal. |
| R3 is never represented as executable | fake | Declaration inspection records GitHub's absent CDC source binding and unavailable delete capability; this is a static capability fact, not a fake provider route. |

## GSD and required skills

Inline/manual GSD execution is required because this direct issue phase has a non-numeric identifier and the canonical single-worker contract forbids role spawning. Resolved adapter command path: `discuss-phase` → `plan-phase --tdd` → `execute-phase` → `verify-work` → `code-review`, each generated with `scripts/gsd prompt`.

Loaded skills: `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, `golang-context`, `golang-concurrency`, and `golang-database`.

CLI help/manual/website parity is not applicable: this changes no command, flag, help text, generated CLI artifact, or public capability declaration. The R3 note is evidence-only and must not imply a new surface.

## Scope and guardrails

- Exactly one target connector: PostgreSQL. This uses existing generic application and database seams only; no shared-runtime, schema, generated-index, or unrelated-connector change is admissible.
- R3 is deferred by a captain decision: GitHub's destination has no CDC source binding and marks deletes unavailable. Do not create a generic writer or an API destination action.
- The real source event must originate from the actual PostgreSQL pgoutput stream, never a hand-constructed tombstone at the target helper.
- The live test requires PostgreSQL 14+ via the supplied shared Docker endpoint. It must not restart, reconfigure, or remove shared Docker/Colima resources.
- Refusal tests must assert a typed error and zero source/target I/O for every enumerated declared destination cell.

## TDD execution plan

1. **Red — full R4 route test.** Add the tagged test around a real PostgreSQL replication transaction and an independent live target read-back. It must initially demonstrate the missing source-to-target wiring or lack of end-to-end evidence.
2. **Green — minimal route seam.** Reuse durable staging, receipt/workset, mapping, and the PostgreSQL keyed apply/history target. Change only the smallest test seam or production wiring required to traverse it from a real pgoutput event.
3. **Red — enumerated refusal matrix.** Add one named table entry for R1, R2, and every destination-side `change_capture` declaration; each must name its expected typed error and count I/O before existing broader guard coverage is relied upon.
4. **Green — declared matrix proof.** Keep production behavior unchanged if the guard already satisfies it; make tests enumerate the current definitions and verify the specific type/pre-I/O boundary.
5. **Verify R3 declaration.** Record a non-`pass` `non_executable` result with its exact missing source binding/delete action cause. Search for and report any capability surface that contradicts it.
6. **Verify and review.** Run the live proof and refusal tests, consumers, independent no-build gates, `verify-work`, and a focused inline code review. Record exact result/status rather than treating skips as live success.

## Checkpoints

1. Commit the planning evidence and the useful red-test failure.
2. Commit each green coherent slice with its focused test output.
3. Commit verification/review evidence, push only this branch, open the explicit-base PR, and read the base back through GitHub's API.
