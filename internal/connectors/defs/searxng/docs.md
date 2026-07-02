# Overview

SearXNG is the reference read-only, credential-free declarative-HTTP golden migration (wave0
`wave:F`). It reads web and Reddit search results from a self-hosted SearXNG metasearch instance's
JSON API (`GET {base_url}/search?q=<query>&format=json&pageno=<n>`). This bundle is
engine-vs-legacy parity-tested against `internal/connectors/searxng` (the hand-written connector it
migrates); the legacy package stays registered and unchanged until wave6's registry flip. It is a
pm-native connector (no `catalog_data.json` entry) and is live-registered via
`connectors.RegisterNativeLive("searxng")` in the legacy package's `init()`.

## Auth setup

No credentials are required by default: public SearXNG instances are open. Only `base_url` (your
instance's URL) must be provided; there is no default, so an instance must be named explicitly.

Legacy also supports an optional `api_key` secret sent as a Bearer token for instances behind an
auth proxy. This bundle's `spec.json` still declares `api_key` as an `x-secret` config property for
documentation purposes, but it is not wired into a conditional auth rule in `streams.json`: the
engine's declarative `when` truthiness check on a `secrets.*` reference errors (rather than
evaluating false) when the referenced secret key is entirely absent from the caller's
`RuntimeConfig.Secrets` map, which would break the default (99% common) no-credential case. Since
no read/pagination/manifest parity requirement in wave0 exercises the auth-proxy path, this bundle
intentionally omits an `auth` block entirely (an absent/empty `auth` list resolves to no
authentication on any request — see `engine/read.go`'s `newRuntime`), matching legacy's own default
behavior exactly. See "Known limits" below.

## Streams notes

Both streams (`search`, `reddit`) hit the same `GET /search` endpoint with `format=json`; they
differ only in how the `q` query parameter is scoped. `search` sends `config.query` verbatim.
`reddit` prefixes it with `site:reddit.com ` so any general SearXNG instance (without a dedicated
Reddit engine installed) returns Reddit results, matching legacy's `searxngQuery`
(`searxng.go:225-242`) base case (no `subreddit` configured). Records are extracted from the
top-level `results` array; primary key is `url`; `published_date` is declared as the cursor field
for manifest-surface parity, but neither legacy nor this bundle actually filters or advances reads
by it — SearXNG results aren't reliably orderable by a server-side cursor, so both connectors always
perform a full stream read.

Pagination is `pageno` (1-based `page_number` pagination), matching SearXNG's own pagination
scheme. No page-size query parameter is ever sent (SearXNG's per-page result count is
engine/instance-defined; `page_size` is used purely as the client-side short-page stop threshold,
exactly like legacy's `searxngDefaultPageSize`/`PageNumberPaginator.SizeParam: ""`). The bundle also
declares `max_pages: 1` (legacy's own default, `searxngDefaultMaxPages`) as a documented **known
limit**: see below.

## Write actions & risks

None. SearXNG is a read-only metasearch API with no mutation endpoints; `capabilities.write` is
`false` and this bundle ships no `writes.json`.

## Known limits

- **max_pages hard-stop is not enforced by the engine read path (ENGINE_GAP).** Legacy's
  `max_pages` config (default 1) is a hard request-count cap enforced independently of page
  fullness by `connsdk.Harvest`'s own `maxPages` parameter (`searxng.go:149`): even a source that
  always returns full pages stops after exactly `max_pages` requests. `internal/connectors/engine
  /read.go`'s `readDeclarative` — the only production path `engine.Connector.Read` dispatches to —
  never reads `PaginationSpec.MaxPages` at all (confirmed by grep: zero references outside
  `bundle.go`'s field declaration); its loop stops ONLY on a short/empty page. This bundle still
  declares `max_pages: 1` in `spec.json`/`streams.json` as the spec-correct, honest statement of
  legacy's real default — it simply has no effect on the engine read path today. This is reported as
  an ENGINE_GAP blocker (not silently worked around) in
  `.planning/phases/wave0-engine-harness/traces/waveF-b16-ledger.md`; every parity test in
  `parity_searxng_test.go` besides the dedicated `TestParitySearxng_MaxPagesStopEngineGap` uses a
  source whose last page is short, so the correctly-implemented short-page stop signal (both sides
  agree on this) terminates pagination identically regardless of the gap.
- **Optional Bearer-proxy auth is not modeled.** See "Auth setup" above: this bundle omits the
  optional `api_key`-present-then-Bearer conditional auth rule to avoid a `when`-on-absent-secret
  hard error in the default no-credential path. `api_key` remains declared in `spec.json` for
  documentation/future-work purposes only.
- **Subreddit-narrowing is not modeled.** Legacy's `reddit` stream additionally supports a
  `subreddit` config value that scopes the query to `site:reddit.com/r/<sub>` instead of the bare
  `site:reddit.com`. The engine's declarative query templating (`stream.Query`) has no
  conditional/default-value filter, so a subreddit-present-vs-absent branch cannot be expressed
  without also risking an unresolved-key error when `subreddit` is unset (the common case). This
  bundle models only the base case (no subreddit configured); the parity suite documents and tests
  against this base case only.
- **`stream` field is not modeled.** Legacy's `searxngResultRecord` stamps a derived `"stream"` key
  naming which stream a record came from (`streams.go:68`). That value is neither present on the
  raw SearXNG API response nor expressible via `computed_fields` (which resolves only against
  `record.*` — there is no static-literal/stream-name namespace). This bundle's schemas omit
  `stream`; it is not part of the PK/cursor contract (PK is `url`, cursor is `published_date`), so
  omitting it does not affect dedup or incremental semantics.
