---
phase: issue-3985-connector-canon-r1
plan: 01
subsystem: connector delivery canon and parity planning
tags: [connectors, documentation, archive, verification, postgres]
key-files:
  - docs/connector-canon/IMPLEMENTATION-PROCEDURE.md
  - scripts/tests/connector-canon.sh
  - data/cli-postgres-parity-issue-tree-r2/report.md
metrics:
  current_postgres_children: 12
  accepted_live_certifications: 0
coverage:
  - id: D1
    description: Canonical end-to-end connector procedure and recoverable archive
    verification:
      - kind: other
        ref: make connector-canon-check
        status: pass
    human_judgment: false
  - id: D2
    description: Declared implemented commands are checked through real runtime preflight
    verification:
      - kind: unit
        ref: make connector-runtime-preflight
        status: pass
    human_judgment: false
  - id: D3
    description: Documentation and generated artifacts state current honest connector status
    verification:
      - kind: other
        ref: make docs-check; pnpm --dir website typecheck; pnpm --dir website test:scripts
        status: pass
    human_judgment: false
  - id: D4
    description: PostgreSQL issue tree uses warehouse-only routes and all seven mode outcomes
    verification:
      - kind: other
        ref: gh-axi issue subissue list 3972 (12 of 12); data/cli-postgres-parity-issue-tree-r2/report.md
        status: pass
    human_judgment: false
---

# Summary — Issue 3985 connector canon

## Delivered

- One binding [connector implementation procedure](../../../docs/connector-canon/IMPLEMENTATION-PROCEDURE.md)
  covering discovery, definitions, derived surface, all four warehouse flows, schedules, layered
  tests, live proof, certification, and the Foundation Check.
- A self-checking delivery gate: `make connector-runtime-preflight` uses the actual commandrunner
  sweep, and `make connector-canon-check` validates required current/archived reports and public
  status language.
- Recoverable archive/index structure for stale planning, research, GSD material, wrong counts, and
  the former PostgreSQL r1 issue tree.
- A current PostgreSQL r2 issue-tree report. Live REST changes created #3987, moved final
  certification behind it, and made the warehouse-only flow/mode contract explicit.

## Commits

| Commit | Description |
| --- | --- |
| `168b97a0d` | Planning/TDD evidence for issue #3985. |
| `f2bac0ce8` | Canon procedure, source pins, archive, documentation, and executable guards; rebased onto #3970's merged GitHub parity surface. |
| `7b7cb4e8d` | PostgreSQL parity r2, archived r1, deterministic canon check, and #3972 REST audit evidence. |

## Deviations

- The formal GSD executor/reviewer roles could not be spawned under the repository's canonical
  single-worker contract; inline execution, verification, and review are recorded in this phase.
- The original GitHub documentation overlap was resolved in favor of merged source-pinned inventory
  counts. The void historical gap map remains archived and forbidden; current generated counts are
  not removed merely because the old report was wrong.
- No live provider certification was run or claimed. The current accepted count remains zero.

## Self-Check: PASSED

The current index points at the procedure and r2 tree, all current source hashes validate, archived
r1 cannot be mistaken for the active graph, the runtime preflight sweep passes, and REST read-back
confirmed #3972 has 12 children including the new pre-certification conformance gate.
