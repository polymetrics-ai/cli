# Discussion log — Issue #3973 mapping contract completion

The issue body, parent #3972, the task brief, and #3982's recorded foundation
gap settle every material decision. No product choice is reopened:

1. #4139 already delivered plan sealing, preview/approval, pinned sessions,
   commit outcomes, and ledger-backed acknowledgement eligibility.
2. The remaining shared vocabulary is a typed mapping, explicit tombstones,
   and a named durable receipt; PostgreSQL implementation is deferred to #3982.
3. `LogicalType` and `TypePlan` already implement the lossless-or-refused
   policy in `internal/connectors/database`, so putting a second logical type
   system in a driver or transport would be drift.
4. The only delete authority is an explicit validated `synccontract.Tombstone`.
   Physical absence from records cannot be represented as a deletion request.
5. The canonical single-worker contract forbids role spawning. The generated
   GSD prompts are executed inline/manual and recorded in this phase directory.

`scripts/gsd prompt discuss-phase 3973` was generated and executed inline.

