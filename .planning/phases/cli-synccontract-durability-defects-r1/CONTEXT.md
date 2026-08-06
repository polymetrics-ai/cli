# Sync-contract durability defects — context

- **Gathered:** 2026-08-06
- **Status:** Ready for TDD execution
- **Source:** Error-severity review findings against shipped PR #3882.

## Locked decisions

- Fix only the two reported defects and any identical `a.store.Load()` assignment bypass found in
  the complete `internal/app` audit. Do not alter the surrounding sync contract.
- All reloads that replace `a.state` must use the one established state-normalization boundary;
  no call site may selectively retain raw persisted state.
- Warehouse durability must sync every directory entry newly created by `MkdirAll`, from the leaf
  through its parent chain to the first directory that pre-existed the operation. Reuse
  `internal/durability.SyncDirectory`; introduce no dependency.
- Tests must first fail against the unmodified code. The state regression must drive
  `Open → RunReverseETL(unknown plan) → RunETL` from persisted legacy-shaped state. The durability
  regression must observe all required directory sync calls, because a portable unit test cannot
  faithfully simulate a power loss.
- No credentials, connection strings, or warehouse contents may be printed or stored in artifacts.

## Scope

- `internal/app/app.go` and focused app tests for the state reload audit and legacy-mode admission.
- `internal/app/local_warehouse.go`, its narrow test seam if needed, and focused tests for
  parent-chain directory syncing.
- This phase's plan, TDD ledger, red-run trace, and verification checklist.

## Explicit exclusions

- No new dependency, connector, command, documentation, or external service behavior.
- No durability-check weakening and no broad sync-contract refactor.

## GSD fallback

`scripts/gsd prompt discuss-phase 3882` and `scripts/gsd prompt plan-phase 3882 --tdd` were
resolved and executed inline. The adapter's phase workflow requires a numbered roadmap phase, but
this is a bounded post-merge remediation for PR #3882 rather than a roadmap phase. Inline/manual
execution is therefore recorded here under the project contract's single-worker fallback.
