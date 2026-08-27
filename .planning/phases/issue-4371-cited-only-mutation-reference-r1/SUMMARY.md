---
coverage:
  - id: D1
    description: Cited-only mutation dispositions preserve the exact closed source-reference projection.
    requirement: "#4371 acceptance 1-2"
    verification:
      - kind: unit
        ref: cmd/connectorgen/sourceprojection_test.go:TestSourceProjectionCitedOnlyMutationDispositionsKeepReferenceClosed
        status: pass
      - kind: unit
        ref: cmd/connectorgen/sourceprojection_test.go:TestSourceProjectionWriteDisabledMutationArtifactsKeepCitedOnlyReferenceClosed
        status: pass
    human_judgment: false
  - id: D2
    description: Existing contract-complete mutation dispositions and fail-closed citation admission remain unchanged.
    requirement: "#4371 acceptance 3"
    verification:
      - kind: unit
        ref: cmd/connectorgen/sourceprojection_test.go:TestSourceProjectionMutationDispositionInputRemainsFailClosed
        status: pass
      - kind: unit
        ref: cmd/connectorgen/sourceprojection_test.go:TestSourceProjectionSourceCitedPartialMutationCoveragePreservesImplementedIncompleteAction
        status: pass
    human_judgment: false
  - id: D3
    description: The repository's generated connector surface remains executable and synchronized.
    requirement: "#4371 acceptance 4-5"
    verification:
      - kind: other
        ref: make connector-runtime-preflight; connectorgen validate; surface-sync --check; operation-evidence --check
        status: pass
    human_judgment: false
---

# Summary — issue 4371

The shared importer/projection now rejects explicit cited-only mutation
dispositions before they can mutate the descriptor, and skips the automatic
write-disabled artifact when the existing exact
`source_contract_unavailable` foundation already explains the operation. This
repairs source accounting and visible missing-foundation mapping only; it does
not make a provider mutation executable.

No provider source was fetched, no source lock was re-pinned, no credential was
used, no action/command/transport was created, and no existing runnable command
was downgraded. Usable CLI surface delta: **0**.

The canonical GSD lifecycle was executed inline because the project contract
forbids worker-role spawning: discussion and TDD plan, red/green/refactor
implementation, verification, and code review prompts are recorded in the
phase artifacts.
