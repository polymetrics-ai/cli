# Issue #3897: Connection-scoped flow warehouse reads — Discussion Log

> Audit trail only. `CONTEXT.md` is the implementation input.

**Date:** 2026-08-11
**Mode:** `scripts/gsd prompt discuss-phase issue-3897-flow-connection-scope-r1 --auto`

## Decisions supplied by the issue contract

| Area | Alternatives considered | Selected decision |
|---|---|---|
| Query selector | New ad-hoc flag, an existing manifest field, unscoped lookup | Existing query-step `connection` field scopes its warehouse read. |
| Action selector | Step-level sync connection, action-specific source field, no selector | `action_cfg.source_connection` scopes `source_table`. |
| Root tables | Treat as a normal connection, reject, canonical selector | Accept only `warehouse.UnattributedConnection` (`_unattributed`). |
| Ambiguity advice | Invent a CLI flag, generic message, manifest remedy | Preserve typed ambiguity and name only a manifest field the flow can accept. |
| Action scope | Dispatch a provider write, use test stub, defer dispatch | Preserve source identity at the existing action-runner boundary; defer dispatch to #3994. |

No additional product questions remained after applying the issue body, the
Sol topology report, and the shared implementation contract.
