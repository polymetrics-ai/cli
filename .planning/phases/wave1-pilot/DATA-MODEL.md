# DATA-MODEL — wave1-pilot

No new data model in this phase. The authoritative shapes are unchanged from wave0:

- Bundle file contracts: `internal/connectors/engine/schema/*.schema.json` +
  `internal/connectors/engine/bundle.go` types (Metadata, Spec, HTTPBase/StreamSpec/WriteAction,
  PaginationSpec, AuthSpec) — documented for authors in `docs/migration/conventions.md` §1–§3 and
  `.planning/phases/wave0-engine-harness/DATA-MODEL.md`.
- Per-stream record schemas: draft-07 subset, `x-primary-key`/`x-cursor-field`, derived sync
  modes (conventions §2) — each pilot adds `defs/<name>/schemas/*.json` INSTANCES of this model,
  field-for-field from its legacy mapRecord (no model change).
- Persisted state/cursor shape: `internal/app/sync_modes.go` `recordCursor` →
  `toComparableString` digit-string persistence (wave0 B1) — pilots must round-trip it in parity
  tests (TEST-PLAN §1), not alter it.
- Agent I/O: `docs/migration/result.schema.json` (executor), `docs/migration/review.schema.json`
  (reviewer) — unchanged.
- New FILES this phase are instances, not models: `docs/migration/pilot-costs.json` (shape
  defined in PLAN.md P-13 / OBSERVABILITY.md; consumed once for the Pass B decision, not a
  runtime artifact).
