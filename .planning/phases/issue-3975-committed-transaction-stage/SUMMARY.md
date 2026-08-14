---
coverage:
  - id: D1
    description: Private chunks become visible only as one ordered committed transaction.
    verification:
      - kind: unit
        ref: internal/connectors/database/transaction_stage_test.go:TestCommittedTransactionStagePublishesOnlyAfterDurableReceipt
        status: pass
    human_judgment: false
  - id: D2
    description: Durable receipt is the sole acknowledgement-eligibility fact.
    verification:
      - kind: unit
        ref: internal/connectors/database/transaction_stage_test.go:TestTransactionReceiptCannotForgeAcknowledgementEligibility
        status: pass
      - kind: unit
        ref: internal/connectors/database/transaction_stage_fault_test.go:TestCommittedTransactionStageReceiverAndReceiptFaultsRemainReceiptless
        status: pass
    human_judgment: false
  - id: D3
    description: Failures and restart leave only safe cleanup or sealed retry states.
    verification:
      - kind: unit
        ref: internal/connectors/database/transaction_stage_fault_test.go:TestCommittedTransactionStageBeginAppendAndSealFaultBoundariesCleanIncompleteState
        status: pass
      - kind: unit
        ref: internal/connectors/database/transaction_stage_test.go:TestCommittedTransactionStageRestartRecoveryRetainsOnlySealedReceiptlessWork
        status: pass
    human_judgment: false
---

# Summary — Issue #3975 committed-transaction staging and durable receipts

## Delivered

`internal/connectors/database/transaction_stage.go` now supplies the private,
source-agnostic committed-transaction boundary: begin, streamed append, seal,
whole-transaction receiver delivery, abort, startup recovery, quota refusal,
and immutable durable receipt loading. It deliberately does not know anything
about PostgreSQL, an LSN, source acknowledgement, a target, or a generic write.

Chunks use a deterministic hashed key below the configured root and are stored
as private complete files. A valid receiver must consume all ordered chunks;
only its durable downstream declaration can lead to a receipt. The receipt is
persisted with file and parent-directory durability before it exposes the
existing `synccontract` acknowledgement type. Incomplete state is discarded;
sealed/no-receipt state is retryable; receipt-backed residue is cleanup-only.

## TDD and correction result

The immutable remote audit checkpoint `2afa128e` contains the exact named
behavioural RED test and its retained output. Local GREEN implementation adds
normal, cancellation, restart, race, and injected crash-boundary suites. One
of five correction rounds was consumed by overflow-safe quota arithmetic found
during code review; no newly discovered repository-gate defect occurred.

## Remaining delivery gates

All local GSD verify/review, focused/race/static, and individual repository
gates are green. The next local gate is canonical no-mistakes. Per the user
directive, there is no subsequent push or draft child PR yet; the immutable
audit checkpoint is preserved and the eventual child PR must still target
`feat/3972-postgres-parity` only.
