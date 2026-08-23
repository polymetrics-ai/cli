# Refs #4303 — Action-binding canonical repair R2 context

## Scope and authority

- **Only behavioral authority:** `/Users/karthiksivadas/karthik-agent-workspace/data/cli-reverse-etl-action-binding-final-review-r1/report.md`.
- **Reviewed identity:** `origin/fm/cli-reverse-etl-action-binding-foundation-r1` at `846ff9e5d9f56cef3f9835d35b57b1b4d468b379`.
- **Working branch:** `fm/cli-reverse-etl-action-binding-foundation-r1`, started exactly from that SHA; no rebase, cherry-pick, or Foundation-wave change is permitted.
- **Execution mode:** `scripts/gsd` registered all lifecycle commands, but `gsd-sdk query init.phase-op 4303` reports no roadmap phase. This issue-local phase is the documented inline/manual fallback. Compatible isolated GSD workers are unavailable and the canonical single-worker contract forbids role spawning.

## Locked repair decisions

1. Plan, preview, authorization, apply, and both human/JSON outputs must carry one complete, deterministic, digest-bound physical action list. A tombstone delete is visible as its own destructive action before approval; no later expansion may introduce executable work.
2. Read-back is selected by the persisted action and validated against that action's identity, fields, and limits before any provider request. Ordinary and tombstone policies remain distinct.
3. An idempotent missing-ok delete may count as unchanged only when all input rows are accounted for and no failures occurred. It still proves independent absence before checkpointing and never replays a mutation.
4. Every provider write unit is sealed against the narrowest action/read-back/receipt/acknowledgement constraint and the actual escaped/composite receipt representation before I/O. A local receipt, composition, output, or acknowledgement failure after a provider result preserves sanitized ordered results and the error chain.
5. Provider idempotency derives from the durable workset occurrence plus action and record identity. Retrying the same occurrence is stable; equal payload/index values from another durable workset are different.
6. Schema and declaration closure are strict: optional batches preserve absence, record/tombstone mappings are exclusive, every eligible action has exactly one executable strategy, and every binding belongs to the reachable strategy and write action set.
7. Help, manual, website, and generated surfaces must accurately state effective batch clamping and independently sealed tombstone delete/read-back semantics. No connector definition, credential, provider mutation, generic write primitive, or Foundation-wave integration is in scope.

## Canonical references

- `/Users/karthiksivadas/karthik-agent-workspace/data/cli-reverse-etl-action-binding-final-review-r1/report.md` — immutable repair ledger; final `BLOCKED` verdict.
- `docs/sync-transport-definition.md` — declaration and operator contract.
- `docs/cli/etl.md`, `website/content/docs/etl.mdx`, `internal/cli/docs.go` — W05 parity surfaces.
- `internal/connectors/sync_transport.go`, `internal/connectors/engine/schema/sync_transport.schema.json` — transport model/schema closure.
- `internal/app/issue_label_warehouse_transport.go`, `internal/synctransport/orchestrator.go`, `internal/connectors/engine/write.go`, `internal/synccontract/commit.go` — execution, result retention, and identity boundaries.

## Delivery constraints

- Use production-shaped synthetic tests; no credentials or provider mutation.
- Add a failing behavioral test before each production repair, then record the exact red and green commands.
- Commit and non-force push every coherent green group to the named remote branch immediately.
- No PR, merge, tag, release, broad generated rewrite, or modifications sourced from the parallel Foundation repair.
