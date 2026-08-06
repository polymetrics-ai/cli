---
phase: cli-found-polling-watermark-executor-r1
issue: 3855
status: implemented
coverage:
  - id: D1
    description: An implemented test-only polling-watermark bundle gains public CDC only through the existing matching-executor gate.
    verification:
      - kind: unit
        ref: internal/connectors/engine/polling_watermark_test.go:TestPollingWatermarkTestBundlePromotesCDCOnlyWithRegisteredExecutor
        status: pass
    human_judgment: false
  - id: D2
    description: Polling watermark declarations require every checkpoint field and describe a closed, bounded mechanism.
    verification:
      - kind: unit
        ref: internal/connectors/engine/polling_watermark_test.go:TestPollingWatermarkBundleRejectsEveryMissingCheckpointField
        status: pass
      - kind: unit
        ref: internal/connectors/engine/polling_watermark_test.go:TestPollingWatermarkConnectorRejectsUndeclaredStream
        status: pass
    human_judgment: false
  - id: D3
    description: Inclusive tuple boundaries replay ties, declared timestamp lag overlaps late arrivals, and no initial clock boundary silently skips history.
    verification:
      - kind: unit
        ref: internal/connectors/engine/polling_watermark_test.go:TestPollingWatermarkExecutorReplaysTieAtPageBoundary
        status: pass
      - kind: unit
        ref: internal/connectors/engine/polling_watermark_test.go:TestPollingWatermarkExecutorAppliesDeclaredSafetyLag
        status: pass
      - kind: unit
        ref: internal/connectors/engine/polling_watermark_test.go:TestPollingWatermarkExecutorLeavesInitialSnapshotBoundaryToSource
        status: pass
    human_judgment: false
  - id: D4
    description: Delete observability is explicit: only declared soft-delete fields or fixed deletion endpoints emit tombstones; hard deletes remain unavailable.
    verification:
      - kind: unit
        ref: internal/connectors/engine/polling_watermark_test.go:TestPollingWatermarkExecutorOnlyEmitsDeclaredSoftDeletes
        status: pass
      - kind: unit
        ref: internal/connectors/engine/polling_watermark_test.go:TestPollingWatermarkExecutorUsesDeclaredDeletionEndpoint
        status: pass
      - kind: unit
        ref: internal/connectors/engine/polling_watermark_test.go:TestPollingWatermarkExecutorRejectsUndeclaredDeletionRecords
        status: pass
      - kind: unit
        ref: internal/connectors/engine/polling_watermark_test.go:TestPollingWatermarkDeclarationRejectsHardDeleteTombstoneClaim
        status: pass
    human_judgment: false
  - id: D5
    description: The executor commits only after destination acknowledgement, replays after checkpoint persistence failure, and honours all declared work/cancellation bounds.
    verification:
      - kind: unit
        ref: internal/connectors/engine/polling_watermark_test.go:TestPollingWatermarkExecutorReplaysAfterCheckpointPersistenceCrash
        status: pass
      - kind: unit
        ref: internal/connectors/engine/polling_watermark_test.go:TestPollingWatermarkExecutorDoesNotCommitWhenDestinationRejectsPage
        status: pass
      - kind: unit
        ref: internal/connectors/engine/polling_watermark_test.go:TestPollingWatermarkExecutorHonorsRequestBudgetAndCancellation
        status: pass
    human_judgment: false
---

# SUMMARY — polling-watermark changefeed executor

Implemented the shared, declaration-driven polling-watermark executor with a
test-only bundle. It preserves the existing fail-closed CDC capability gate,
uses inclusive tuple replay for timestamp ties, supports declared overlap lag,
and refuses any undeclared delete visibility. Destination acknowledgement
precedes checkpoint persistence, so every interruption replays instead of
skipping.

No production connector declaration, provider call, credential, runtime
service, webhook, or event-stream mechanism was touched. Checkpoint persistence
remains a small consumer adapter pending #3810's versioned durable envelope.
