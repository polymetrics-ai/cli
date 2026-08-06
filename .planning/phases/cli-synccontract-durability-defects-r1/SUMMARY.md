---
coverage:
  - id: D1
    description: Raw state reloads retain legacy sync-mode compatibility after a reverse-plan lookup.
    verification:
      - kind: unit
        ref: internal/app/sync_state_test.go:TestLegacyStateReloadRetainsSyncModeCompatibilityAfterReversePlanLookup
        status: pass
    human_judgment: false
  - id: D2
    description: A newly created warehouse directory chain is synced through its first pre-existing ancestor before acknowledgement.
    verification:
      - kind: unit
        ref: internal/app/sync_state_test.go:TestRunWarehouseETLSyncsNewDirectoryParentChainBeforeAcknowledgement
        status: pass
    human_judgment: false
---

# Summary — sync-contract durability defects

The two error-severity defects shipped in PR #3882 are repaired without changing the surrounding
sync contract. `normalizeLoadedState` now owns all `a.store.Load()` assignments, and local
warehouse writes synchronize the raw-directory ancestor chain before building the durable
downstream acknowledgement.

The manual GSD fallback executed the generated `discuss-phase`, `plan-phase --tdd`,
`execute-phase`, `verify-work`, and `code-review` guidance inline because this is a bounded
post-merge remediation rather than a numbered roadmap phase. No GSD role was spawned.
