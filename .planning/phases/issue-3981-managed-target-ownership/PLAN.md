# PLAN — Issue #3981 managed-target ownership and provisioning

## GSD execution record

The commands `discuss-phase`, `plan-phase --tdd`, `execute-phase`,
`verify-work`, and `code-review` were resolved with `scripts/gsd sources` and
their prompts generated through `scripts/gsd prompt`. `scripts/gsd doctor` and
`go run ./cmd/agentcontractgen check` passed before planning.

The project-local GSD adapter expects a numbered roadmap phase and role
execution. #3981 is an issue-local child, and
`.agents/agentic-delivery/canonical/delivery-contract.json` requires one inline
worker with delegation disabled. This is the explicit manual-GSD fallback, not a
workflow waiver: every required phase is recorded below.

Required skills loaded: `golang-how-to`, `golang-design-patterns`,
`golang-structs-interfaces`, `golang-naming`, `golang-error-handling`,
`golang-security`, `golang-safety`, `golang-testing`, `golang-context`,
`golang-concurrency`, `golang-database`, `golang-code-style`, `golang-lint`,
`github-issue-first-delivery`, `no-mistakes`, and the five GSD lifecycle skills.

## Scope and boundaries

Implement the shared driver-neutral managed-target identity and provisioning
contract in `internal/connectors/database`, reusing
`warehouse.ArtifactIdentity`. It derives opaque physical names from the source
workspace/connector/connection triple, holds a typed owner/ref/control record,
and pins expected native relation identity plus schema hash/version.

Excluded: all PostgreSQL-specific DDL or driver code, generic SQL, target
execution, direct write/query surfaces, CDC, transport, and capability changes.
No customer target is adopted or evolved.

## TDD execution sequence

1. **Plan checkpoint.** Create this artifact set before production edits and
   preserve the inline fallback reason.
2. **RED — closed truth table.** Add
   `TestManagedTargetProvisioningTruthTable`, table-driving absent namespace,
   first/repeat creation, correct/missing/foreign/unreadable owner, name
   collision, moved/native-ID replacement, schema drift, concurrent provisioning,
   typed-plan-only mutation, identity secrecy, and cancellation. Run its focused
   package command and retain the failing output at
   `traces/managed-target-provisioning-red.txt`.
3. **GREEN — typed values and port.** Add immutable owner/ref/schema/control
   values, a validated typed provisioning plan, typed failure classifications,
   and a narrow driver-fake-friendly observation/create port. Ensure physical
   names are deterministically derived from opaque owner identity, never a
   display or credential value.
4. **GREEN — fail-closed state machine.** Add target-scoped coordination that
   re-observes under lock; creates only on the no-target/no-control state; and
   asserts exact owner/ref/native/schema after every creation or repeat request.
   Rejection has no mutation side effect.
5. **REFACTOR/VERIFY.** Format and simplify without broadening scope. Run
   focused, race, cancellation, affected package, static/build, and individual
   repository gates. Complete inline `verify-work`; plan/execute gaps only if a
   real uncovered requirement appears; then complete code review and the
   Shepherd-compatible/no-mistakes gates.

## Acceptance checks

- The owner/ref/control record is typed and validates workspace, source
  connector, and source connection identity structurally.
- Source identity is the only owner derivation; returned names and error text do
  not expose display-name or credential inputs.
- Mutations require a valid typed plan and matching asserted owner.
- Only absent target plus absent control record permits create. Exact control and
  observed native/schema state is idempotently admitted.
- Missing, unreadable, foreign, colliding, moved/replaced, schema-drifted, and
  orphaned states refuse without adoption or evolution.
- Cancellation and concurrent requests do not allow a mutation without a
  reasserted exact record.
- The package remains driver-neutral; a fake is the sole test implementation.

## Commit and delivery checkpoints

1. Planning artifacts.
2. RED test and GREEN contract/state-machine implementation.
3. Verification/review/no-mistakes corrections, capped at five rounds. #4038
   is the correction child created before fixing the cross-provisioner locking
   gap found during required code review.
4. Push the child branch and open exactly one draft PR to
   `feat/3972-postgres-parity`, never `main`; do not merge.
