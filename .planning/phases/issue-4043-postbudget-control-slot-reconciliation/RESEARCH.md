# Research - #4043 post-budget control-slot reconciliation

## Source

This phase uses the completed exact-head audit at
/Users/karthiksivadas/karthik-agent-workspace/data/cli-pg-3975-postbudget-defects-audit-r1/report.md.
No external package, service, provider, database, or network research is needed.

## Confirmed root causes

### Pre-rename discard temp

atomicWriteWithOutcome creates an owned temporary, but its deferred error path
discarded close/remove errors and did not sync the temporary parent after a
successful unlink. persistDiscardIntent interpreted not-applied as safe and
reset Temporary to Reserved. discardEntry then removed the transaction stage,
released the control, and left no root poison. Repeated Begin/Abort can therefore
leave more owned temporary controls than MaxStagedTransactions.

### Recovered over-capacity stage

addRecoveredEntry inserts a retained entry even when reserveControlLocked fails.
Recovery initially poisons the root, but reconcileDiscardControls examines only
discard-failed entries and controls that exist. It does not establish that every
retained entry has an exact Reserved control. Poison then clears, allowing
AdmitRecoveredTransaction and CommitTransaction to invoke the receiver and
produce otherwise-valid durable receipt and acknowledgement evidence.

## Proven valid behavior to preserve

1. Reserved control before durable Begin.
2. Final-control path: durable final intent, durable transaction removal,
   durable final retirement, then control release.
3. Within-capacity sealed recovery remains RecoveryHeld until explicit admission.
4. Receipt durability remains the only acknowledgement eligibility boundary.
5. Restart removes recognized owned regular discard temps only and preserves
   unknown/corrupt artifacts.

## Repair constraints

- No new public API or dependency.
- Use errors.Join for independent cleanup errors where existing error semantics
  require preserving both causes.
- Keep all shared map mutation under the existing stage mutex.
- Do not hold the mutex across filesystem or receiver I/O.
- Never make a root clean merely because an artifact is currently invisible;
  required parent directory synchronization is part of durable reconciliation.

## Test strategy

The five named behavioral tests are permanent regressions. They use t.TempDir,
the existing in-package fault storage adapter, exact owned artifact paths, and
fake receivers. They must prove the full observable boundary rather than an
exit status:

1. discard temporary removal failure retains one slot and blocks admission;
2. successful unlink requires a recorded discards directory sync;
3. over-capacity recovery cannot clear poison or reach receiver/receipt/ack;
4. reservation restart matrix covers cap transitions and ordering; and
5. temporary crash matrix covers every durable/indeterminate boundary.
