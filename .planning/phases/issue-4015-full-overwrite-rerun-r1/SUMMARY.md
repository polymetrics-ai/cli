---
phase: issue-4015-full-overwrite-rerun-r1
status: ready_for_pr
coverage:
  - id: D1
    description: Full-refresh source modes ignore a saved checkpoint on every run.
    requirement: Refs #4015
    verification:
      - kind: unit
        ref: internal/synctransport/transport_test.go TestOrchestratorSourceCheckpointFollowsRefreshSemantics
        status: pass
    human_judgment: false
  - id: D2
    description: A changed-source full overwrite replaces the target contents on rerun.
    requirement: Refs #4015
    verification:
      - kind: e2e
        ref: internal/cli/postgres_transport_binary_integration_test.go TestPMBinaryPostgresFullOverwriteRetainsEverySourcePage
        status: pass
    human_judgment: false
  - id: D3
    description: Incremental modes retain checkpoint resumption and unchanged-source skipping.
    requirement: Refs #4015
    verification:
      - kind: unit
        ref: internal/synctransport/transport_test.go TestOrchestratorSourceCheckpointFollowsRefreshSemantics
        status: pass
      - kind: e2e
        ref: internal/cli/postgres_transport_binary_integration_test.go TestPMBinaryPostgresIncrementalUpsertStillSkipsUnchangedSource
        status: pass
    human_judgment: false
---

# Full-overwrite rerun correctness — Summary

## Outcome

The source checkpoint decision now follows the canonical source-refresh contract. `full_append` and `full_overwrite` begin every run without a previous source position; `incremental_append`, `incremental_dedupe`, `incremental_dedupe_history`, `incremental_upsert`, and other resumable modes preserve their saved position. Resume identity and generation remain unchanged.

The rule is shared by generic transport, run-scoped full overwrite, serial Arrow, and pipelined Arrow extraction. No destination publication, checkpoint commit ordering, connector surface, CLI output, or dependency changed.

## Live result

Before the fix, run one transferred `3/3` rows. After the source changed from IDs `[1 2 3]` to `[2 3 4]`, run two reported completed `0/0`; on the PostgreSQL control route the independent target query showed zero rows. This confirms the checkpoint-resumption hypothesis, while refuting the narrower claim that every destination route necessarily retains stale rows.

After the fix, the same run two transferred `3/3`, the target contained exactly IDs `[2 3 4]`, ID 2 changed from `page-one-b` to `replacement-two`, and deleted ID 1 was absent. A separate live `incremental_upsert` replay stayed `0/0` and preserved all three target rows plus sample `id=2 label="event-two"`.

## Delivery

Ready for a direct PR from `fm/cli-full-overwrite-fix-r1` to `integration/4015-mvp-flat-r1` with `Refs #4015`. Full-suite CI remains the repository-owned monolithic gate; local verification followed the required scoped-package and individual-gate strategy.
