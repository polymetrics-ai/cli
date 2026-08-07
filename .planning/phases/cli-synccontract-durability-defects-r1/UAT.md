# UAT — sync-contract durability defects

The generated `verify-work` prompt was executed inline under the single-worker fallback. Both
deliverables are automated and need no human judgment.

| ID | Deliverable | Evidence | Verdict |
| --- | --- | --- | --- |
| D1 | State reload preserves the legacy adapter through an unknown reverse-plan lookup. | `TestLegacyStateReloadRetainsSyncModeCompatibilityAfterReversePlanLookup` passed. | Pass |
| D2 | New warehouse parent entries are synced before the acknowledged checkpoint. | `TestRunWarehouseETLSyncsNewDirectoryParentChainBeforeAcknowledgement` passed. | Pass |

The D2 evidence observes every required sync invocation; it does not attempt to emulate a real
filesystem power loss.
