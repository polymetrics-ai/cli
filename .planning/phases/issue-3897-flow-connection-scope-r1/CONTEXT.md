# Issue #3897: Connection-scoped flow warehouse reads — Context

**Gathered:** 2026-08-11  
**Status:** Ready for TDD planning  
**GSD mode:** Manual issue-phase fallback

## Phase Boundary

Allow a flow manifest to select the owning warehouse connection for `query`
and `action` source reads. The selector must reach the real Parquet/DuckDB
read path, accept `_unattributed`, preserve fail-closed ambiguity for omitted
selectors, and survive the action-dispatch boundary. This issue does not add
connector action dispatch, approval-token plumbing, scheduling, rate policies,
transport, generic writes, or provider calls.

## Manual-GSD Fallback

`scripts/gsd doctor` and all required command resolutions passed. The normal
GSD `phase-op` registry cannot execute for this issue because the active
`.planning/ROADMAP.md` intentionally contains no numbered phases and directs
current work to issue-specific artifacts under `.planning/phases/<issue-or-phase>/`.
The required lifecycle is therefore executed inline in this issue directory,
with the resolved `scripts/gsd prompt` traces recorded in `PLAN.md` and each
RED/GREEN/REFACTOR/verification result kept in the ledger. The canonical
delivery contract forbids GSD-role delegation, so no role agent is spawned.

## Locked Decisions

### Manifest shape

- **D-01:** A `query` step uses its existing optional `connection` field as
  its warehouse selector. It remains optional so legacy manifests deserialize
  unchanged and an omitted selector can still be refused only when ambiguous.
- **D-02:** An `action` step records its source selector as
  `action_cfg.source_connection`, separate from any step-level sync
  connection. It is optional for the same backwards-compatibility and
  fail-closed reasons.
- **D-03:** `_unattributed` is accepted exactly as
  `warehouse.UnattributedConnection`; it selects only root-owned tables and
  never resolves through the project connection registry.

### Read and ambiguity behaviour

- **D-04:** Flow actions use the existing connection-aware table read request
  for `source_table`. Flow queries pass their optional selector into the
  DuckDB-backed SQL read path, where selected views expose only that owner.
- **D-05:** An omitted selector against a duplicated source remains a typed
  `*warehouse.AmbiguousTableError`. The generic SQL CLI path keeps its current
  no-invented-flag wording; flow-only error decoration may name the manifest
  field (`connection` or `action_cfg.source_connection`) because that is a
  remedy the caller can actually supply.
- **D-06:** No user-controlled selector is interpolated into SQL. Warehouse
  layout resolution owns connection validation and DuckDB view registration.

### Action boundary and scope fence

- **D-07:** `ActionConfig.SourceConnection` is preserved by JSON
  parse/serialize and passed with the entire `FlowStep` to the existing action
  runner. The runner and future #3994 approval path therefore receive the
  identity of the table that was read; this issue does not introduce an action
  preview, approval token, provider mutation, or generic HTTP write.
- **D-08:** Tests must exercise both a real two-owner Parquet warehouse and
  the flow engine/adapter boundary. They assert returned rows, typed errors,
  and selected identity — never exit status alone.

## Existing Code Insights

- `internal/flow/manifest.go` owns the JSON manifest model and validation.
- `internal/flow/engine.go` invokes `AppAdapter.QuerySQL` for query steps and
  currently constructs a SQL string for action source reads.
- `internal/cli/flow_cli.go` adapts the flow engine to `app.App`.
- `internal/app/types.go` and `internal/app/app.go` already expose
  `QueryTableRequest.Connection`; `_unattributed` is intentionally accepted.
- `internal/app/query_engine_duckdb.go` builds Parquet-backed DuckDB views,
  but duplicate names are only registered as qualified views. It needs
  connection-aware view selection and a typed ambiguity preflight for the flow
  source path.
- `internal/warehouse/layout.go` owns `FindTable`,
  `AmbiguousTableError`, `WithAmbiguityRemedy`, and the `_unattributed`
  semantics. Do not restate those rules elsewhere.

## Canonical References

Downstream implementation and review must read:

- `AGENTS.md` — delivery lifecycle, warehouse ownership, direct-read, and
  safety contracts.
- `.agents/agentic-delivery/canonical/delivery-contract.json` — single-worker
  lifecycle, no-mistakes vector, and prohibition on `--yes`.
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md` —
  public flow syntax parity checklist.
- `docs/architecture/connector-architecture-v2-design.md` — only as the
  boundary for deferred #3994 connector action dispatch; do not implement it.
- `internal/warehouse/layout.go` — authoritative selected/unattributed and
  ambiguity semantics.
- `internal/app/query_engine_duckdb.go` — real Parquet/DuckDB query path.
- `internal/app/warehouse_connection_isolation_test.go` — existing two-owner
  fixtures and reverse-plan pinning precedent.

## Deferred Ideas

- #3994 owns connector-backed action dispatch and approval/preview integration.
- #3992 owns durable scheduling.
- #3990 owns rate policy behaviour.
- #3864/#3862 own transport dispatch; no generic HTTP or SQL write is added.
