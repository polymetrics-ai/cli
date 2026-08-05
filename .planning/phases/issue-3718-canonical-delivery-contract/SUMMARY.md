---
phase: issue-3718-canonical-delivery-contract
status: complete
coverage:
  - id: D1
    description: One versioned canonical base flow and inheriting connector overlay
    verification:
      - kind: unit
        ref: internal/agentcontract/contract_test.go TestCanonicalContractRequiredInvariants
        status: pass
      - kind: unit
        ref: internal/agentcontract/contract_test.go TestRenderIsStableAndConnectorInheritsBase
        status: pass
    human_judgment: false
  - id: D2
    description: Registered projection divergence fails and bounded sync repairs it
    verification:
      - kind: integration
        ref: internal/agentcontract/check_test.go TestProjectionDriftCheckAndSync
        status: pass
      - kind: other
        ref: make agent-contract-check
        status: pass
    human_judgment: false
  - id: D3
    description: Every referenced GSD command resolves through the real adapter
    verification:
      - kind: integration
        ref: internal/agentcontract/check_test.go TestReferencedGSDCommandsResolve
        status: pass
    human_judgment: false
  - id: D4
    description: Single-worker parent ownership and authority gates are fail-closed
    verification:
      - kind: unit
        ref: internal/agentcontract/contract_test.go TestCanonicalContractRequiredInvariants
        status: pass
      - kind: other
        ref: policy and changed-path grep recorded in VERIFICATION.md
        status: pass
    human_judgment: false
---

# SUMMARY — issue #3718 canonical delivery contract

## Delivered

- Added the single versioned JSON source for `pm-delivery-worker` plus its inheriting connector
  overlay.
- Added deterministic generated blocks, exact drift checking, bounded marked-block sync, and six
  optional future harness targets without creating harness-native files.
- Added an executed contract test that resolves every named GSD command through `scripts/gsd
  sources` and removed stale mandatory `programming-loop` guidance from the changed canonical
  issue-first policy surfaces.
- Wired the check into serial and parallel `make verify` paths.
- Reframed parent state ownership around one canonical worker while preserving child checks/review
  coverage, parent-review fallback tracking, full parent gates, and explicit captain approval.

## TDD outcome

RED proved the absent command failed resolution and that the pre-enforcement checker accepted a
diverged projection. GREEN reads the real canonical source, resolves all commands, rejects drift,
and repairs only marked blocks. REFACTOR added stable render hashes, fail-closed source invariants,
checked output/error paths, and provider-neutral shared Go naming.

## Scope outcome

No `.claude/agents`, `.codex/agents`, `.pi/agents`, global home configuration, legacy role file,
project-local GSD adapter, or reserved connector-lane path changed.
