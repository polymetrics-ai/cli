# ADR 0001 — Connectors as data (JSON defs bundles + declarative engine)

- Status: Accepted (2026-07-02)
- Deciders: user (approved plan `~/.claude/plans/please-check-all-the-serialized-storm.md`)
- Context docs: `docs/architecture/connector-architecture-v2-design.md` (full design),
  `docs/architecture/connector-certification-design.md`,
  `docs/plans/universal-programming-loop-prd.md`

## Context

The `pm` CLI carries ~556 hand-written Go connector packages (~309k lines) whose manifests,
schemas, and write actions live in Go structs (`internal/connectors/manifest.go`, per-connector
`streams.go`). The catalog derives from Airbyte residue (`catalog_data.json`, `source-`/
`destination-` slugs via `slug.go`), only 2 connectors implement reverse-ETL writes, and
conformance runs against synthetic fixture records that bypass real request/pagination/cursor
logic.

## Decision

1. **Connector definitions become split JSON bundles** under `internal/connectors/defs/<name>/`
   (`metadata.json`, `spec.json` with `x-secret`, `streams.json`, `writes.json`,
   `api_surface.json`, `schemas/*.json` with `x-primary-key`/`x-cursor-field`, `fixtures/`,
   `docs.md`), embedded via a single `go:embed`, executed by a well-tested declarative engine
   (`internal/connectors/engine/`) built on the existing `connsdk` toolkit.
2. **Three strict tiers**: declarative-only (target ≥90%), bundle + named hooks (≤300 lines Go),
   full native with mandated component split (~10 connectors) — escape hatches are additive,
   never replacements; conformance rejects Go where JSON suffices.
3. **Sync modes are derived, never declared** (from `x-primary-key` / `incremental` presence).
4. **Unified bare names, clean break** (no slug aliases) and **the catalog is a view over loaded
   definitions** — executed at convergence (wave6), not incrementally.
5. **Minimal internal draft-07 validator and tiny interpolator** — no new Go module dependencies;
   only the keywords/filters the bundles actually use.
6. **Migration is staged**: wave0 builds engine+harness alongside legacy with three golden
   parity-tested migrations (stripe, searxng, postgres); registry flip and deletion happen only
   at wave6 behind a human gate.

## Alternatives considered

- Keep Go connectors, generate manifests to JSON: rejected — doesn't close the capability gap or
  enable ~105-agent parallel authoring; drift remains possible.
- One `manifest.json` per connector: rejected — poor diff hygiene and agent readability
  (design §A decision).
- Adopt an existing JSON-Schema library / template engine: rejected — dependency-free rule; the
  needed subset is small and a full engine invites bundle complexity beyond the three-tier policy.
- Big-bang rewrite with immediate registry flip: rejected — parity risk across 556 connectors;
  staged waves with parity gates chosen instead.

## Consequences

- (+) A connector like aircall drops from ~723 lines of Go to zero; agents author/diff JSON.
- (+) Conformance v2 exercises the real engine against recorded pages; certification becomes
  possible per credential.
- (−) The engine is a single point of failure → mitigated by ≥85% coverage gate, golden parity
  tests, and the `ENGINE_GAP` blocker protocol (≥3 same-type gaps → engine extension wave).
- (−) Two representations coexist until wave6 → mitigated by test-only construction of engine
  connectors and a byte-identical `registryset` invariant during waves 0–5.
