---
phase: issue-4043-postbudget-control-slot-reconciliation
plan: 4043-01
type: tdd
wave: 1
depends_on: []
files_modified:
  - internal/connectors/database/transaction_stage.go
  - internal/connectors/database/transaction_stage_fault_test.go
  - internal/connectors/database/transaction_stage_test.go
  - .planning/phases/issue-4043-postbudget-control-slot-reconciliation
autonomous: true
requirements:
  - ISSUE-4043-CONTROL-SLOT-TEMP-CLEANUP
  - ISSUE-4043-RECOVERED-RESERVATION-COVERAGE
---

<objective>
Enforce the finite one-control-slot-per-stage-instance invariant across
pre-rename cleanup, restart, recovery reconciliation, admission, and receiver
delivery.

Purpose: close the two confirmed #4043 acceptance gaps without changing the
public transaction-stage contract, receipt format, CLI surface, PostgreSQL
driver, source progress, target DML, generic SQL, provider scope, or
DuckDB/Parquet mediator.

Output: five durable behavioral regressions, two private reconciliation repairs,
and recorded Red/Green/repeat/race/restart/package/repository evidence.
</objective>

<context>
- .planning/phases/issue-4043-postbudget-control-slot-reconciliation/CONTEXT.md
- .planning/phases/issue-4043-postbudget-control-slot-reconciliation/RESEARCH.md
- /Users/karthiksivadas/karthik-agent-workspace/data/cli-pg-3975-postbudget-defects-audit-r1/report.md
- AGENTS.md
</context>

<threat_model>
## Threat Model

| Asset / boundary | Threat | Mitigation in this plan | Severity |
|---|---|---|---|
| Stage-root bounded namespace | A failed pre-rename temp cleanup releases a slot and permits unbounded owned controls. | Track post-create cleanup durably; retain Temporary and poison until exact remove plus discards sync reconcile. | high |
| Recovery delivery boundary | A retained recovered entry lacks a slot yet reaches receiver, receipt, and acknowledgement. | Require exact Reserved coverage before poison clears and at admit/commit/discard-retry gates. | high |
| Durable receipt evidence | A repair changes or weakens receipt-before-ack semantics. | Do not alter TransactionReceipt, persistReceipt, receipt validation, or Acknowledgement; test only pre-receiver rejection. | high |
| Local filesystem test paths | A fault test mutates broad or external paths. | Use t.TempDir and existing injected storage only; assert exact owned temp/final paths. | medium |

No credentials, network, provider, PostgreSQL service, warehouse mutation,
generic SQL, target DML, reverse ETL, external write tool, or dependency is in
scope.
</threat_model>

<feature>
  <name>Fail-closed control-slot reconciliation</name>
  <files>
    internal/connectors/database/transaction_stage.go
    internal/connectors/database/transaction_stage_fault_test.go
    internal/connectors/database/transaction_stage_test.go
  </files>
  <behavior>
    - Every retained receipt-less entry has exactly one matching Reserved
      control before cleanup poison can clear.
    - A created discard temporary retains its single slot until close, removal,
      and required parent directory sync are durably reconciled.
    - An unreserved recovered entry cannot admit, invoke the receiver, persist
      a receipt, or expose acknowledgement eligibility.
  </behavior>
  <implementation>
    Keep the repair private to the existing transaction-stage state machine,
    reuse its exact generation control key and existing faultable storage
    adapter, and leave receipt persistence and acknowledgement unchanged.
  </implementation>
</feature>

<tasks>
<task type="tdd" id="1">
  <name>RED: capture durable temp and recovered-capacity failures</name>
  <read_first>
    - internal/connectors/database/transaction_stage_fault_test.go
    - internal/connectors/database/transaction_stage_test.go
    - internal/connectors/database/transaction_stage.go
    - .planning/phases/issue-4043-postbudget-control-slot-reconciliation/TDD-LEDGER.md
    - /Users/karthiksivadas/karthik-agent-workspace/data/cli-pg-3975-postbudget-defects-audit-r1/report.md
  </read_first>
  <action>
    Add these exact behavioral tests before production edits:
    TestCommittedTransactionStageDiscardControlTempCleanupRetainsSlotUntilReconciled,
    TestCommittedTransactionStageDiscardControlTempRemovalRequiresDirectorySync,
    TestCommittedTransactionStageRecoveredOverCapacityCannotClearPoisonOrDeliver,
    TestCommittedTransactionStageRecoveredControlReservationRestartMatrix, and
    TestCommittedTransactionStageDiscardControlTemporaryCrashMatrix.

    Reuse transactionStageFault and installTransactionStageFaults. Assert exact
    owned temporary/final paths, slot and entry/control mapping, typed
    cleanup-required root poison, recorded discards directory sync attempts,
    zero receiver calls, absent receipt, and unavailable acknowledgement for
    ineligible work. Use real t.TempDir reopen paths; do not use sleeps,
    containers, credentials, or network.
  </action>
  <acceptance_criteria>
    - The focused five-test command fails on the unmodified production code for
      durable behavioral assertions, not a compile failure.
    - The ledger records literal Red output including the current no-poison
      temporary or receiver/receipt/acknowledgement symptom.
    - Existing valid-path tests remain separately runnable.
  </acceptance_criteria>
  <verify>
    go test -timeout 20m -count=1 -v ./internal/connectors/database -run '^(TestCommittedTransactionStageDiscardControlTempCleanupRetainsSlotUntilReconciled|TestCommittedTransactionStageDiscardControlTempRemovalRequiresDirectorySync|TestCommittedTransactionStageRecoveredOverCapacityCannotClearPoisonOrDeliver|TestCommittedTransactionStageRecoveredControlReservationRestartMatrix|TestCommittedTransactionStageDiscardControlTemporaryCrashMatrix)$'
  </verify>
</task>

<task type="tdd" id="2">
  <name>GREEN slice A: make pre-rename temporary cleanup durable</name>
  <read_first>
    - internal/connectors/database/transaction_stage.go
    - internal/connectors/database/transaction_stage_fault_test.go
    - .planning/phases/issue-4043-postbudget-control-slot-reconciliation/RESEARCH.md
    - .planning/phases/issue-4043-postbudget-control-slot-reconciliation/TDD-LEDGER.md
  </read_first>
  <action>
    Refactor atomicWriteWithOutcome or introduce a private equivalent result so
    a created temporary records close, exact temporary remove, and temporary
    parent directory sync outcomes. Preserve not-applied only when no temp was
    created or its absence is durably reconciled. In persistDiscardIntent,
    retain Temporary for unresolved cleanup; return indeterminate typed cleanup
    information; never reset the control to Reserved merely because the write
    itself was not applied.

    Update discardEntry and reconcileDiscardControls so unresolved Temporary
    retains the entry and slot, poisons the root, blocks new work, reaps exact
    visible owned temps, syncs discards even after an invisible previous unlink,
    and only then restores Reserved for a retry. Never create another temp for
    the same exact generation while Temporary remains unresolved.
  </action>
  <acceptance_criteria>
    - A failed remove or discards sync produces ErrTransactionStageCleanupRequired,
      leaves one control for cap one, and blocks Begin, Append, Commit, and Admit.
    - A successful temporary unlink records a successful discards directory sync
      before slot release.
    - Reconciliation restores safe admission only after exact durable cleanup;
      it cannot manufacture a second temporary for one generation.
  </acceptance_criteria>
  <verify>
    Run the first, second, and fifth focused tests, then the valid final-control
    and owned-temp recovery controls from TDD-LEDGER.md.
  </verify>
</task>

<task type="tdd" id="3">
  <name>GREEN slice B: require exact reservation coverage before delivery</name>
  <read_first>
    - internal/connectors/database/transaction_stage.go
    - internal/connectors/database/transaction_stage_test.go
    - internal/connectors/database/transaction_stage_fault_test.go
    - .planning/phases/issue-4043-postbudget-control-slot-reconciliation/RESEARCH.md
    - .planning/phases/issue-4043-postbudget-control-slot-reconciliation/TDD-LEDGER.md
  </read_first>
  <action>
    Add one private mutex-held predicate that validates exact
    transaction-key/instance control ownership and Reserved state. Use it in
    reconciliation, AdmitRecoveredTransaction, CommitTransaction before
    receiver invocation, and discard retry. At reconciliation completion,
    traverse retained receipt-less entries in deterministic transaction-key
    order, reserve missing controls only if capacity allows, validate no
    entry/control mismatch and len(controls) within MaxStagedTransactions, and
    retain typed cleanup poison otherwise.

    Preserve over-capacity sealed entries as RecoveryHeld. A later recovery with
    sufficient capacity may reserve all retained stages, but each remains held
    until its own explicit admission. Do not modify receipt persistence or
    acknowledgement code.
  </action>
  <acceptance_criteria>
    - Reconcile under cap one for two sealed recovered stages returns
      cleanup-required repeatedly and does not clear poison.
    - The unreserved stage's Admit and Commit fail before receiver invocation;
      receiver calls remain zero, receipt is absent/unavailable, and
      acknowledgement fails.
    - Reopen with cap two reserves both exact generations yet retains recovery
      hold until explicit admission; existing within-capacity behavior stays green.
  </acceptance_criteria>
  <verify>
    Run the third and fourth focused tests, the within-capacity recovery test,
    and receipt-before-ack valid-path regressions from TDD-LEDGER.md.
  </verify>
</task>

<task type="execute" id="4">
  <name>Refactor, verification, GSD closeout, and delivery readiness</name>
  <read_first>
    - .planning/phases/issue-4043-postbudget-control-slot-reconciliation/TDD-LEDGER.md
    - .planning/phases/issue-4043-postbudget-control-slot-reconciliation/VERIFICATION.md
    - AGENTS.md
    - .agents/agentic-delivery/contracts/issue-agent-contract.md
  </read_first>
  <action>
    Run gofmt on only the three allowed Go files. Record Green output from the
    focused, count-20, race count-10, valid-path, full package,
    synccontract/app, vet, build, diff, agent-contract, and named repository
    gates. Execute GSD execute-phase, verify-work, and code-review prompts
    inline using the same named-phase fallback; disposition every finding.

    Commit coherent green checkpoints: plan/TDD, Red tests, Green repair, and
    any review fix. Only after committed local green evidence inspect
    no-mistakes AXI home on this branch and start a new 0/5 run without --yes.
    Never attach to, respond to, synchronize, abort, rerun, or recover the
    historical 01KZQ5Q1WFN264HPDS2HN8V1BD run.
  </action>
  <acceptance_criteria>
    - All commands marked required in VERIFICATION.md pass or have a recorded
      external blocker; no runtime service is started.
    - GSD verification and code review have recorded disposition.
    - New no-mistakes budget remains separate from the historical exhausted run.
    - CLI help/manual/website parity is explicitly Not Applicable because no
      CLI-visible contract changes.
  </acceptance_criteria>
  <verify>
    Execute the ordered verification matrix in VERIFICATION.md.
  </verify>
</task>
</tasks>

<verification>
The authoritative exact commands, expected Red symptoms, Green assertions,
repeat/race/restart matrices, valid-path controls, package gates, and repository
gates are recorded in TDD-LEDGER.md and VERIFICATION.md.
</verification>

<must_haves>
<truths>
- No owned control artifact count exceeds MaxStagedTransactions across
  pre-rename failure, repeat, reconcile, or restart.
- Root poison cannot clear while an entry lacks an exact Reserved control or any
  Temporary/Final cleanup remains unresolved.
- No unreserved generation invokes the receiver, creates a receipt, or exposes
  acknowledgement eligibility.
- Valid final retirement, bare sealed recovery hold, explicit recovery admission,
  and immutable receipt-before-acknowledgement remain unchanged.
</truths>
</must_haves>

<success_criteria>
- Five named behavioral tests exist and demonstrate genuine Red then Green.
- Only allowed production/test/planning paths change.
- The branch starts at 6a82f3650ab4be0b511541f91721ce7cefe08762 and uses a
  fresh no-mistakes 0/5 budget.
- A stacked PR can truthfully state Refs #4043, #3975, #3972, and #4015 with
  base feat/3972-postgres-parity and no closing keyword.
</success_criteria>
