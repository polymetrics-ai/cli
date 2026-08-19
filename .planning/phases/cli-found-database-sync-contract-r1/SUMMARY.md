---
coverage:
  - id: D1
    description: Closed database sync mode vocabulary and native admission
    verification:
      - kind: unit
        ref: internal/synccontract/contract_test.go:TestModeVocabularyIsClosed
        status: pass
      - kind: unit
        ref: internal/synccontract/contract_test.go:TestNativeContractNeedsMatchingExecutorAndFixtureEvidence
        status: pass
    human_judgment: false
  - id: D2
    description: Versioned opaque checkpoint envelope and explicit recovery
    verification:
      - kind: unit
        ref: internal/synccontract/contract_test.go:TestCheckpointEnvelopePreservesOpaqueTokensAndPartitionState
        status: pass
      - kind: unit
        ref: internal/synccontract/contract_test.go:TestResumeOutcomesRequireExplicitRebootstrap
        status: pass
    human_judgment: false
  - id: D3
    description: Downstream-acknowledged commitment and legacy app persistence
    verification:
      - kind: unit
        ref: internal/synccontract/contract_test.go:TestCommitAfterDownstreamAcknowledgement
        status: pass
      - kind: unit
        ref: internal/app/sync_state_test.go:TestIncrementalRunStoresCommittedStateEnvelopeAfterDownstreamSuccess
        status: pass
      - kind: unit
        ref: internal/app/sync_state_test.go:TestConnectorETLDoesNotCommitAfterPartialDestinationResult
        status: pass
    human_judgment: false
  - id: D4
    description: Tombstone and history-window close semantics
    verification:
      - kind: unit
        ref: internal/synccontract/contract_test.go:TestTombstoneClosesHistoryWindowInsteadOfPhysicalDelete
        status: pass
    human_judgment: false
  - id: D5
    description: No silent scalar-cursor migration or unsupported-mode source read
    verification:
      - kind: unit
        ref: internal/app/sync_state_test.go:TestLegacyScalarStreamStateRequiresRebootstrapBeforeRead
        status: pass
      - kind: unit
        ref: internal/app/sync_state_test.go:TestContractModeCannotReadWithoutNativeExecutor
        status: pass
    human_judgment: false
---

# Summary — Issue 3810: shared database sync contract

Implemented the top-level sync foundation in `internal/synccontract` and made `StreamState` persist
the versioned envelope instead of a scalar cursor. The contract provides the closed vocabulary,
opaque state shape with identity and replay-window semantics, typed rebootstrap outcomes,
acknowledgement-gated commits, tombstones/history, native command admission, and embedded
immutable fixture corpus.

The existing app paths remain explicit legacy compatibility adapters. Their persisted state now
uses a version-one envelope after successful downstream work; old scalar cursor JSON becomes a
version-zero state that blocks before source read and requires explicit rebootstrap. New contract
mode names do not execute through the legacy runner without a matching native executor plus full
fixture evidence.

No connector bundle, engine, command runner, CLI/docs/website surface, credentials, or live
provider path changed. #3855 and #3862 can consume `internal/synccontract` directly.

## Manual GSD fallback

The issue is not a numbered roadmap phase, and the task forbids spawning GSD role agents. The
generated `execute-phase`, `verify-work`, and `code-review` prompts were resolved and followed
inline: red-first tests, implementation, focused/local verification, and a manual deep review are
recorded in the sibling artifacts.
