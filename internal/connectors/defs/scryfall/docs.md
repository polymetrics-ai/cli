# Overview

Scryfall is a wave2-fan-out declarative-HTTP, read-only, no-auth migration (the same shape as
`searxng`). It reads Magic: The Gathering cards and sets from the public Scryfall API
(`GET https://api.scryfall.com/...`). This bundle targets capability parity with
`internal/connectors/scryfall` (the hand-written connector it migrates); the legacy package stays
registered and unchanged until wave6's registry flip.

## Auth setup

None. Scryfall's public API requires no credentials, matching legacy's `requester` (no `Auth`
field set on `connsdk.Requester`) and its own test asserting no `Authorization` header is ever
sent (`scryfall_test.go:23-25`). `spec.json` declares no secret fields; `base_url` defaults to
`https://api.scryfall.com` and may be overridden for tests/proxies.

## Streams notes

`cards` (`GET /cards/search`) sends the `q` config value as the search query, defaulting to `*`
(match-all) when unset — identical to legacy's own fallback (`scryfall.go:108-112`). `sets`
(`GET /sets`) sends no query parameters at all, also matching legacy (legacy only special-cases
the `q` param when `spec.path == "cards/search"`). Both streams read records at `data` and use
`pagination.type: next_url` with `next_url_path: next_page`, matching Scryfall's own real
`next_page` absolute-URL wire convention (confirmed by legacy's own test fixture,
`scryfall_test.go:36`, which serves `next_page` as a fully-qualified URL). Neither legacy stream
declares an incremental cursor field (legacy's `Catalog` never sets `CursorFields`), so no
`incremental` block is declared here either.

`metadata.json` declares no `rate_limit` — legacy enforces no client-side rate limiting for
Scryfall (no throttling logic anywhere in `scryfall.go`), so this bundle adds none, matching
legacy's actual behavior.

## Write actions & risks

None. Legacy's Scryfall connector returns `connectors.ErrUnsupportedOperation` from `Write`
unconditionally; `capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- **The declarative `check` request cannot replicate legacy's `q=*` check query.** Legacy's
  `Check()` sends `GET /cards/search?q=*&page=1` (`scryfall.go:57-58`); the engine's declarative
  `HTTP.Check` (`bundle.go`'s `RequestSpec`) is method+path only and never attaches query
  parameters (`read.go`'s `Check` calls `rt.Requester.Do(..., nil, nil)`). This bundle's
  `streams.json` therefore declares `check: {method: GET, path: /cards/search}` with no query —
  a real (non-fixture) Scryfall API would 400 this exact request (`q` is a required search
  parameter), a genuine request-shape narrowing versus legacy's own check behavior. This is an
  `ENGINE_GAP`-adjacent limitation of the `check` dialect (no query-templating support), not a
  fixable per-bundle workaround; documented here rather than silently worked around. The
  conformance/fixture-replay harness's check-fixture server does not match on request query at
  all (`conformance/replay.go`'s `newCheckReplayServer`), so this gap is invisible to fixture-mode
  testing and only manifests against a real, unmocked Scryfall endpoint.
- **`max_pages` is not modeled.** Legacy exposes a config-driven `max_pages` override
  (`scryfall.go:182-195`, default `100`) as a hard request-count cap. The engine's `next_url`
  paginator has no `MaxPages`-equivalent knob wired to a config value; pagination here is bounded
  only by the short/empty-`next_page` stop signal, matching Scryfall's own real termination
  behavior. `max_pages` is not declared in `spec.json` (a declared-but-unwireable config key is
  worse than an absent one, per F6/REVIEW.md precedent).
- **Legacy's fixture-mode-only marker field is not modeled.** Legacy's `readFixture` path stamps a
  synthetic `fixture: true` marker onto every record (`scryfall.go:149`); this is a
  credential-free conformance-harness affordance, not part of the live record shape, and is
  intentionally not modeled — the engine's own conformance/fixture-replay harness provides the
  equivalent test affordance.
- **Every stream declares `projection: "passthrough"`** (conventions.md §8 rule 1): legacy performs
  zero record shaping on either stream — `scryfall.go:126-129`'s `emit(connectors.Record(rec))`
  passes the raw decoded `data[]` entries straight through with no field renaming, computation, or
  filtering. Schema-mode projection (the dialect default) would silently drop every real Scryfall
  card/set field beyond `id`/`name`/`set`, a parity-breaking behavior change versus legacy;
  `passthrough` keeps every raw field, matching legacy's actual emitted-record shape exactly.
  Schema stays intentionally minimal (`id`/`name`/`set`) as a **documentation surface only** now
  that `passthrough` — not schema shape — governs what survives projection; full field-level schema
  expansion is Pass B (wave5) scope.
- Full Scryfall API surface (single-card lookup, symbology, bulk data) is out of scope for this
  wave; see `api_surface.json`'s `excluded` entries.
