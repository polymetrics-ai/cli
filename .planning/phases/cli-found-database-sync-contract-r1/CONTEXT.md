# Issue 3810 — Shared database sync contract: context

**Gathered:** 2026-08-06\
**Status:** Ready for implementation\
**Source:** GitHub issue #3810, its captain attachment ruling, and the required database-sync research reports.

<domain>
## Phase boundary

Create the one product-wide contract that owns database sync semantics: the closed mode
vocabulary, checkpoint-state envelope, recovery outcomes, acknowledgement-before-commit rule,
delete/history semantics, native command admission, and reusable conformance fixtures. It is a
foundation for #3855 and #3862, not an implementation of a database engine.
</domain>

<decisions>
## Locked implementation decisions

### Ownership and compatibility

- #3810 is the sole owner of sync semantics. Consumer lanes import its types and fixtures; they
  do not restate mode names, state fields, recovery codes, or delete/history rules.
- Preserve the five existing legacy `internal/app` mode behaviors as compatibility adapters while
  introducing the seven closed contract names. A new contract mode is never executable merely
  because it parses: native executor identity and complete fixture evidence are both required.
- A persisted legacy scalar cursor is converted into a deliberately non-resumable legacy envelope.
  It returns a typed rebootstrap-required outcome; it is never cleared or silently used for a
  replacement full scan.

### State and recovery

- Persist opaque provider checkpoints as bytes. Validation may compare/copy bytes but never parse,
  normalize, reconstruct, or coerce them to a scalar string.
- State carries distinct source identity, snapshot barrier, primary/tie-breaker positions,
  per-partition records, schema/protocol versions, dedupe identity, observed time, and committed
  time.
- All invalid checkpoint conditions are explicit typed outcomes: invalid checkpoint, retention
  gap, invalidated slot, expired token, source-generation change, and source-identity mismatch.
- A checkpoint may be marked committed only through the downstream durable-ack path. A read can
  produce an observed candidate but cannot persist advancement.

### Deletes and native execution

- Tombstones are explicit envelopes with operation, stable event identity, delete image shape,
  record key, and ordering position. History mode always emits a validity-window close
  (`_valid_to`, `_is_current=false`), never a physical target delete.
- Native database commands use a fixed native contract (protocol, named executor, modes, fixture
  evidence). The contract deliberately contains no REST method/path/API-surface field and no
  caller-supplied SQL, HTTP, or shell text.

### Scope fence

- Do not implement PostgreSQL, DynamoDB, polling/watermark execution (#3855), any-to-any
  transport/credentials (#3862), #3747 conformance fixture migration, #3748 CLI/docs/website
  surfacing, #3749 connectorgen work, or #3762 bounded-query taxonomy.
- Do not edit `internal/connectors/commandrunner/runner.go`, connector bundles, or redaction
  behavior. No credentials or live-provider calls are used.
</decisions>

<canonical_refs>
## Canonical references

- `docs/migration/HANDOFF-CODEX.md` — parallel migration ownership and collision rules.
- `docs/migration/conventions.md` — Tier-3 native connector boundary; an empty REST API surface is
  correct for SQL/wire protocols.
- `docs/architecture/connector-architecture-v2-design.md` — existing descriptor and native
  connector architecture.
- `internal/app/sync_modes.go` — five legacy public names and scalar cursor comparison behavior.
- `internal/app/types.go` — scalar `StreamState.Cursor` persistence shape being replaced.
- `internal/app/app.go` — legacy read/write and final run-state persistence seam.
- `internal/app/local_warehouse.go` — local warehouse acknowledgement/swap seam and legacy physical
  dedupe behavior that this foundation must not retrofit.
- `internal/connectors/connectors.go` — #3861 changefeed descriptor/executor admission contract.
- `data/cli-database-engine-expansion-research-r1/report.md` §3 (workspace research source) —
  versioned state-envelope and native Tier-3 design.
- `data/cli-database-connector-parity-research-r1/report.md` (workspace research source) — measured
  legacy baseline and explicit gaps.
</canonical_refs>

<specifics>
## Baseline established before design

- `ParseSyncMode` currently accepts five legacy names: `full_refresh_append`,
  `full_refresh_overwrite`, `full_refresh_overwrite_deduped`, `incremental_append`, and
  `incremental_append_deduped`; numeric cursor comparison uses `float64` when timestamps do not
  parse, and state persistence stores one `Cursor string`.
- Existing ETL only assigns its stream state after final destination writes complete, but it has no
  versioned source identity, partition slots, typed invalidation, or independently recorded
  durable acknowledgement.
- #3861 is present at the branch base. Its changefeed descriptor remains the capability gate; this
  work supplies the shared sync contract it will be consumed through rather than duplicating it.
</specifics>

<deferred>
## Deferred work

- Per-engine executable contracts and their real conformance runs: #3856–#3859 and engine lanes.
- End-user mode migration/help/manual/website surfaces: #3748 and #3860.
- Native transport dispatch and cross-transport matrix: #3863–#3867.
</deferred>

---

*Issue: 3810 — shared database sync contract*
