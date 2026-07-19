# PRD — Connector Architecture v2 (unified, JSON-schema-declarative, fully-capable connectors)

Program of record for milestone `connector-architecture-v2`. Full approved plan:
`~/.claude/plans/please-check-all-the-serialized-storm.md` (2026-07-02). This PRD is the repo-local
source of truth consumed by the GSD Universal Programming Loop's PRD-coverage gate.

## Problem

The `pm` CLI has ~556 native Go connectors, but the architecture has drifted from the product goals:

1. **Airbyte residue** — `catalog_data.json` (646 entries) uses `source-X`/`destination-X` slugs
   with a `type: source|destination` split; `internal/connectors/slug.go` exists purely as a
   compatibility shim. The product promise (README/PRD) is unified connectors — one name, both
   directions.
2. **Manifests are Go code, not JSON** — stream schemas, config fields, and write actions live in
   Go structs (`internal/connectors/manifest.go`, per-connector `streams.go`), unlike the Ruby
   reference architecture (`polymetrics-ai/polymetrics`, latest connector `snowflake_connector`)
   where `metadata.json`, `connection_specification.json`, `schemas/<stream>.json` are declarative
   JSON-Schema files.
3. **Capability gap** — only 2 of 556 connectors implement reverse-ETL writes (github: 13 actions,
   stripe: 2). ~554 are read-only; stream coverage is partial.
4. **Catalog divergence** — 556 live connectors vs 2 catalog-enabled; catalog derives from upstream
   Airbyte instead of from the connectors themselves.
5. **No live certification** — conformance is fixture-mode only; the documented enablement policy
   has no automated harness.

## Goals

- **G1 Connectors as data**: per-connector JSON bundles under `internal/connectors/defs/<name>/`
  (`metadata.json`, `spec.json` with `x-secret`, `streams.json`, `writes.json`, `api_surface.json`,
  `schemas/*.json` with `x-primary-key`/`x-cursor-field`, `fixtures/`, `docs.md`) executed by a
  declarative engine (`internal/connectors/engine/`) built on the existing `connsdk`.
  Three strict tiers: declarative-only (≥90%), bundle+hooks (~8%, ≤300 lines Go), full native
  (~10 connectors, mandated component split: connector/connection/reader/cataloger/writer/cdc).
- **G2 Unified naming, clean break**: bare connector names everywhere; delete slug/alias/catalog
  legacy machinery; the catalog is a view over loaded definitions. No backward compatibility.
- **G3 Full capability on first run**: every documented GET endpoint is an ETL stream; every
  POST/PUT/PATCH/DELETE is a reverse-ETL write action — or is consciously excluded in
  `api_surface.json` with a closed-vocabulary category. Conformance fails silent gaps.
- **G4 Certification harness**: `pm connectors certify <name>` exercises the full command surface
  per connector (check, catalog, ETL in all 5 sync modes — 2 live reads + 3 capture replays,
  reverse-ETL create-then-cleanup with write-ahead leak ledger + `--sweep`, flow roundtrip,
  schedule roundtrip with zero crontab residue, secret-redaction scan, JSON-contract assertions).
  Tier 0 fixture / Tier 1 recorded replay / Tier 2 live (credential-gated; no credential =
  `uncertified`, never `failed`). Exit 3 = leaked resources.
- **G5 Migration at scale**: waves orchestrated by the GSD programming loop across Claude, Codex,
  OpenCode, or future runtimes; implementation agents own disjoint scopes; active parent
  orchestrators spawn or assign all ready workers until human-ready or blocked; compact handoffs
  reduce coordination tokens without changing TDD, verification, or safety gates.

## Non-goals

- Backward-compatible aliases for old slugs (explicit clean break).
- Runtime plugin loading or reflection-heavy magic (declarative JSON interpreted by a well-tested
  engine; `go:embed` only).
- Rewriting `internal/app` pipeline/sync-mode semantics (unchanged; sync modes become derived).

## Users

- **Agents** (primary): stable `--json` + exit codes; bundles and certification reports are
  machine-readable; guides render from `docs.md`.
- **Connector authors**: add/modify a connector by editing JSON (+ optional hook), scaffolded by
  `connectorgen new`; conformance and certification gate merge.
- **Data engineers**: unified connector names; certified capability matrix they can trust.

## Phases & acceptance

See `.planning/ROADMAP.md` (milestone `connector-architecture-v2`, phases `wave0-engine-harness`
… `wave6-convergence`) for per-phase acceptance criteria. Program-level acceptance:

- All non-quarantined connectors load from defs bundles; `go build ./... && go test ./... &&
  golangci-lint run` green; full fixture conformance passes.
- `pm connectors certify sample --json` green in CI without secrets; batch live certification remains credential-file gated.
- Live certification passes for every connector with available credentials, including
  create-then-cleanup writes with zero leaked resources.
- Legacy machinery (slug.go, catalog_data.json, registryset, native_port/conformance v1,
  manifest.go structs) deleted; catalog generated from manifests.
- Quarantine ≤5% with typed, documented reasons (`docs/migration/quarantine.json`).

## Key design decisions (from the approved plan)

- Split JSON files per connector (not one manifest.json) — agent-readable diffs, parallel authoring.
- Sync modes derived from schema metadata (`x-primary-key` → dedup modes; `incremental` block →
  incremental modes) — never declared, cannot drift.
- `Write` leaves the core `Connector` interface → optional `Writer` interface.
- Certification: all 5 sync modes covered per connector but only 2 live API reads (full +
  incremental/resume); destination-axis modes replay spooled capture through the real pipeline.
- Parallel-mutation safety: agents own disjoint `defs/<name>/` dirs; shared/generated files are
  orchestrator-only, regenerated once per wave (`registrygen`/`connectorgen gen`).
- Blocked connectors quarantine with typed reasons; ≥3 same-type `ENGINE_GAP` blockers → extend the
  engine, then un-quarantine wave.

## Risks

- **Template defect replicated at scale** → pilot wave + Fable line-by-line review before fan-out;
  >30% review-failure rate halts a wave.
- **Live write tests leak data** → write-ahead ledger, mandatory cleanup verification, orphan
  sweeper, sandbox-gated writes, exit 3 dominance.
- **Engine expressiveness gaps** → three-tier escape hatches; ENGINE_GAP blocker protocol.
- **Token/cost overrun on Pass B** → decision deferred until pilot produces real numbers.
