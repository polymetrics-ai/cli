# Issue #4043: Post-budget control-slot reconciliation - Context

**Gathered:** 2026-08-11
**Status:** Implementation committed; validation pending
**Mode:** Inline/manual GSD fallback

## Phase Boundary

Repair exactly two fail-closed transaction-stage state-machine gaps at pipeline
commit 6a82f3650ab4be0b511541f91721ce7cefe08762:

1. a pre-rename discard-control temporary may be removed unsafely or retained
   without holding its reserved slot; and
2. a recovered receipt-less stage without an exact control reservation may clear
   root poison and reach receiver, receipt, and acknowledgement.

The repair is private to internal/connectors/database transaction staging. It
uses deterministic local filesystem fakes and fake receivers only.

## Locked Decisions

### D-01: Temp cleanup is a durability transition

- A control slot remains held from Reserved through Temporary and Final until
  exact cleanup plus the relevant parent directory sync is durably reconciled.
- If a temp was created and pre-rename close, remove, or discards-directory sync
  fails or is indeterminate, retain Temporary state, retain the entry and slot,
  return typed cleanup-required, and poison the root.
- Do not create a second temp for a generation that remains Temporary.
- Reconciliation removes the exact owned temp, syncs discards even after an
  invisible prior unlink, then restores Reserved before retrying final intent.

### D-02: Reservation coverage is a root-wide invariant

- Every retained receipt-less entry needs one exact-generation Reserved control.
- Recovery preserves over-capacity sealed work as RecoveryHeld for inspection; it
  neither deletes nor auto-admits it.
- Reconciliation walks retained entries in deterministic key order, reserves
  missing controls only when capacity allows, and retains poison if coverage,
  state, or capacity is invalid.
- AdmitRecoveredTransaction, CommitTransaction before receiver invocation, and
  discard retry all require the exact-generation Reserved control.

### D-03: Receipt semantics remain unchanged

- No receiver call, durable receipt, or acknowledgement is allowed for an
  unreserved recovered generation.
- Valid within-capacity recovery still needs explicit admission.
- Existing immutable receipt-before-acknowledgement behavior, sealed recovery
  hold, source-agnostic identities, and DuckDB/Parquet mediation remain intact.

### D-04: TDD and delivery boundary

- Add the five named deterministic behavioral tests before production edits.
- Record literal Red and Green command output, artifact counts, sync calls,
  entry/control mappings, receiver calls, receipt presence, and acknowledgement
  eligibility. Exit status alone is not evidence.
- Keep production scope to transaction_stage.go. The only allowed test files are
  transaction_stage_fault_test.go and transaction_stage_test.go.

## Canonical References

- /Users/karthiksivadas/karthik-agent-workspace/data/cli-pg-3975-postbudget-defects-audit-r1/ship-brief.md
  - authoritative delivery route, branch, issue linkage, GSD, no-mistakes, and PR requirements.
- /Users/karthiksivadas/karthik-agent-workspace/data/cli-pg-3975-postbudget-defects-audit-r1/report.md
  - authoritative reproduction evidence, exact issue addendum, required repair
    design, five permanent tests, and verification matrix.
- AGENTS.md
  - lifecycle, safety, verification, and staged control-slot invariants.
- .agents/agentic-delivery/references/required-skills-routing.md
  - required Go skill routing and GSD path.
- .agents/agentic-delivery/references/gsd-pi-adapter.md
  - official adapter route and inline fallback requirement.
- .agents/agentic-delivery/contracts/issue-agent-contract.md
  - issue-first PR, review, and evidence contract.

## Existing Code Insights

### Reusable Assets

- internal/connectors/database/transaction_stage.go already has the private
  control-state map, exact generation keys, durable stage cleanup result,
  cleanup poison, recovery scanning, and receipt-before-acknowledgement path.
- internal/connectors/database/transaction_stage_fault_test.go has deterministic
  injected create, write, file sync, close, rename, remove, removeAll, and
  directory-sync faults.
- internal/connectors/database/transaction_stage_test.go contains valid
  capacity, recovery hold, receipt, and acknowledgement controls.

### Integration Points

- persistDiscardIntent and atomicWriteWithOutcome form the pre-rename temporary
  durability boundary.
- discardEntry and reconcileDiscardControls own slot release and root poison.
- addRecoveredEntry, AdmitRecoveredTransaction, and CommitTransaction own
  recovery reservation and receiver-admission boundaries.

## Inline Fallback Record

The current phase is intentionally named
issue-4043-postbudget-control-slot-reconciliation and has no numbered
ROADMAP.md entry. scripts/gsd prompt discuss-phase and plan-phase were resolved
and read, but the official numbered-phase init cannot own this issue directory.
The canonical single-worker fallback is therefore used inline. No planner,
researcher, verifier, reviewer, or other role is spawned. This document,
PLAN.md, TDD-LEDGER.md, VERIFICATION.md, RUN-STATE.json, and SUMMARY.md are the
durable equivalent evidence.

## Deferred Ideas

None. PostgreSQL driver, pgoutput, LSN/checkpoint acknowledgement, replication
slots, target DML, generic SQL, CLI, public documentation, connector capability,
provider, credential, network, runtime-service, and warehouse changes are
outside this phase.
