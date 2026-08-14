# PLAN — Issue #3973 mapping, tombstone, and delivery receipt contract

## GSD and skills record

- `scripts/gsd doctor`, all five lifecycle `sources` resolutions, and `go run ./cmd/agentcontractgen check` passed before planning.
- `scripts/gsd prompt discuss-phase 3973` and `scripts/gsd prompt plan-phase issue-3973-mapping-contract-r2 --tdd` were generated and executed inline. This issue is not a numbered roadmap phase and the canonical contract forbids GSD-role delegation.
- Required skills used: `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, and `golang-database`.
- CLI help/manual/website parity is not applicable: this PR has no command, flag, generated surface, or documentation behavior.

## Production scope

- `internal/connectors/database/mapping_contract.go` — shared V1 mapping, value projection, explicit tombstone envelope/input, and named receipt.
- `internal/connectors/database/database_write_session.go` — seal the mapping/tombstone counts into plans; pass bounded typed input through one session; return `DeliveryReceiptV1`.
- Focused `internal/connectors/database` tests and this issue's GSD evidence.

## TDD slices

1. **Red — typed mapping seal:** add external contract tests requiring a versioned source-to-target column map to be reachable from a write plan and approval equality. Capture the failing targeted run.
2. **Green — lossless mapping:** use the existing `TypePlan` policy to construct immutable V1 columns and to project values. Prove a widening integer round trip and prove narrowing/unrepresentable values cannot produce a target projection.
3. **Red/green — explicit delete envelope:** add a bounded input/envelope test with a stateful fake target. Its observable state must retain a row absent from records, and remove it only when the explicit tombstone appears. Thread the envelope/count through plan validation, batches, and the pinned session.
4. **Red/green — receipt naming and composition:** introduce `DeliveryReceiptV1` as the driver/session return type while retaining compatibility for the existing receipt constructor as an alias. Verify result acknowledgement remains unavailable until the distinct managed-target ledger records the receipt identity.
5. **Refactor/verification:** format, run package tests and required scoped gates, execute `verify-work` and `code-review` prompts inline, and use the GSD gap loop only for an identified coverage gap.

## Guardrails

- A mapping has no SQL, DDL, connection details, arbitrary destination relation, or driver-specific type. It references only existing closed logical types and identifier components.
- Records are not inferred as deletes; tombstone validation runs before a session opens and plans bind their exact collection counts.
- Returned slices/maps/JSON/binary values are defensively copied where this contract crosses an ownership boundary.
- No new dependency, live credential, database connection, capability flip, source workset derivation, or target implementation is added.

## Checkpoints

1. Commit GSD plan/TDD evidence.
2. Commit preserved Red test output.
3. Commit green shared-contract implementation and focused tests.
4. Commit only verification/review or gap fixes with their green proof.

