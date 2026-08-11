# #4069 — Context

**Gathered:** 2026-08-12
**Status:** Ready for strict TDD execution
**Mode:** Inline/manual GSD fallback in an issue-numbered phase

## Locked decisions

- The audited #4060 head is immutable: implementation begins only from
  `659efd8a0d69f26b55fcbd3c02150e995c159519` on the dedicated #4069 branch.
- #4066 remains the contract owner. This child starts its own 0 / 5 ledger;
  it does not reopen #4066 or consume a sixth correction there.
- The regression uses real local Parquet/DuckDB data: one owner has `records`,
  one has `RECORDS`, and each exact spelling is unique in the resolver.
- Explicit flow/app reads continue to use the manifest connection selector.
  Omitted flow reads must preserve `errors.As` to
  `*warehouse.AmbiguousTableError` so the flow engine can add its truthful
  `connection` remedy.
- Generic SQL remains deliberately available. The required control is an
  unrelated `SELECT 1`; do not remove generic aliases or parse/rewrite caller
  SQL to solve the collision.
- The solution is derived from the query's existing immutable
  `warehouse.TableResolver` snapshot. No second warehouse scan, hand-authored
  SQL filter, or provider-side behavior is allowed.

## Implementation direction

`newQueryViewPolicy` currently recognizes only an exact name for which
`resolver.Find(name, "")` is already ambiguous. The correction will represent
canonical DuckDB-name collisions from the same snapshot, suppressing duplicate
bare view registrations while retaining generic owner alias behavior. For an
unscoped flow, the canonical collision resolves to a fresh typed ambiguity
whose logical table spelling is deterministic and whose connections come from
the captured owners. Connection-scoped requests keep the zero policy and
remain isolated.

## Required skills

`golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`,
`golang-security`, `golang-safety`, `golang-database`,
`golang-design-patterns`, and `golang-structs-interfaces` were loaded. The
later review pass additionally loads the review-required lint guidance.

## CLI/docs parity disposition

No command, flag, output schema, manifest field, help text, manual page, or
website contract changes. This is an internal fail-closed query-binding
correction. Runtime/docs/website generation remains not applicable to the
implementation diff; later verification will run the repository docs/help
checks that are affected by the final diff and record that result.

## Manual-GSD fallback

The required adapter and all lifecycle commands resolve, but this issue is not
an active numbered ROADMAP phase and this execution environment cannot run the
Pi workflow directly. The canonical single-worker contract also forbids
spawning a GSD role. The generated prompts are executed inline: decisions are
recorded here, the plan and ledger precede production edits, and RED/GREEN,
verify-work, and code-review evidence will be captured in this phase.
