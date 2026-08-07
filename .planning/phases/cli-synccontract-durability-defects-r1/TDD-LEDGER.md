# TDD ledger — sync-contract durability defects

| ID | Guarantee | Required red assertion | Green proof |
| --- | --- | --- | --- |
| D1 | Every state reload retains normalization | Persist current-credential, legacy `incremental_append` state; run `Open → RunReverseETL` with an unknown plan, then `RunETL`; the legacy stream must remain admitted. | Every `a.store.Load()` assignment audit entry uses the shared normalizer, and the exact sequence passes. |
| D2 | Newly created warehouse path is durable as a chain | Start with no warehouse parent directories; perform the acknowledged write/run and observe that every new directory's own parent has been synced through the first pre-existing ancestor. | `syncLocalWarehouseDirectoryChain` syncs the raw directory through the filesystem root before acknowledgement, covering each new `MkdirAll` entry without pre-checking existence. |

## Red evidence

Both tests failed before their respective production fixes; output is retained in
[`traces/red-run.txt`](traces/red-run.txt). The durability test uses a behavior-preserving
observation seam so it can see the existing sync calls; it still fails against the unfixed
leaf-only behavior.

## Green evidence

Both focused tests passed together after the repair:

```sh
go test ./internal/app -run '^(TestLegacyStateReloadRetainsSyncModeCompatibilityAfterReversePlanLookup|TestRunWarehouseETLSyncsNewDirectoryParentChainBeforeAcknowledgement)$' -count=1
```

## Limitation statement

The directory test proves ordering and coverage of sync invocations for the full newly-created
parent chain. It cannot emulate a real kernel/filesystem power-loss window; the production
guarantee relies on the durability package's directory sync implementation for that primitive.
