# Issue #4063: Flow-authoring discovery metadata - Discussion Log

> Audit trail only. CONTEXT.md is the implementation input.

**Date:** 2026-08-11
**Mode:** scripts/gsd prompt discuss-phase issue-4063-flow-matrix-metadata-r1 --auto
**Fallback:** phase-op cannot register non-numbered issue work in the archived
roadmap; decisions were recorded inline under the repository manual-GSD rule.

## Generated source coordinate

| Option | Result |
|---|---|
| Hand-edit the checked-in JSON | Rejected: generated artifact ownership forbids it. |
| Run the canonical generator and verify a one-scalar diff | Selected: the only allowed implementation path. |
| Change discovery behavior or certification facts | Rejected: outside #4063. |

**Auto-selected decision:** regenerate only the stale flow_authoring source
coordinate, then require both canonical drift checking and a semantic diff
boundary.

## Correction lineage

| Option | Result |
|---|---|
| Infer a new correction count | Rejected: correction budget must be proven. |
| Use #3897 TDD ledger and RUN-STATE at exact head | Selected: both record 3 / 5. |
| Continue beyond five corrections | Rejected: forbidden by the issue and delivery contract. |

**Auto-selected decision:** reserve #4063 as correction 4 / 5 before mutation.
