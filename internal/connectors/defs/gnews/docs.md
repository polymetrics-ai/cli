# Overview

GNews is a wave2-fan-out declarative-HTTP migration. It reads news articles from the keyword
search and top-headlines endpoints of the GNews REST API (`https://gnews.io/api/v4`). This bundle
targets capability parity with `internal/connectors/gnews` (the hand-written connector it
migrates); the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a GNews API key via the `api_key` secret. It is sent as the `apikey` query parameter on
every request (`streams.json`'s `base.auth` declares `mode: api_key_query`, matching legacy's
`connsdk.APIKeyQuery("apikey", secret)`); it is never logged.

## Streams notes

Both streams (`search`, `top_headlines`) share the same envelope shape (`{totalArticles,
articles:[...]}`), records at `articles`, primary key `["id"]`, incremental cursor field
`published_at`. Pagination is `page_number` (`page_param: page`, `size_param: max`, `page_size:
10`), matching legacy's default page size and its "short page (fewer than `max` records) stops
pagination" behavior.

`search` sends `q` (defaulting to the literal `news` when unset, matching legacy's harmless-default
fallback for a required search endpoint) plus the shared optional filters `lang`/`country`/`in`/
`nullable`/`sortby`, each omitted entirely when unset via the opt-in `omit_when_absent` query
dialect. `top_headlines` sends an optional `topic` and its own `top_headlines_query`-derived `q`,
plus the same shared optional filters.

Both streams send `from` (the incremental lower bound — either the persisted sync cursor or the
RFC3339 `start_date` config value on a fresh sync) via `{{ incremental.lower_bound }}`, omitted
entirely on a full sync with no configured `start_date` (identical to legacy's
`incrementalLowerBound`, which returns `""` in that case and never sets the `from` query param).
`to` sends the `end_date` config value when set, matching legacy's `gnewsConfigDate(cfg,
"end_date")`. `param_format: rfc3339` is used because GNews's own date format
(`2006-01-02T15:04:05Z`) matches RFC3339 UTC without a numeric transform, exactly as legacy's
`normalizeGNewsDate` always renders it.

## Write actions & risks

None. Legacy's GNews connector returns `connectors.ErrUnsupportedOperation` from `Write`
unconditionally (`capabilities.write` is `false` and this bundle ships no `writes.json`) — GNews is
a read-only news search API with no reverse-ETL writes.

## Known limits

- **`top_headlines`'s `query` fallback is not modeled.** Legacy's `gnewsBaseQuery` sets `q` for the
  `top_headlines` stream from `firstNonEmpty(config["top_headlines_query"], config["query"])` — a
  two-key OR-fallback. The engine's `stream.Query` dialect (`QueryParam`) has no mechanism to
  reference a second config key when the first is absent (only a single template plus a *fixed
  literal* `default`, not another config reference); only `top_headlines_topic`'s query. This
  bundle's `top_headlines` therefore reads `top_headlines_query` only. When only the shared `query`
  config value is set (and `top_headlines_query` is unset), legacy would still populate `q` for
  `top_headlines`; this bundle omits it instead. Documented scope narrowing, not silently wrong —
  same shape as searxng's `subreddit` ledger item.
- **`page_size`/`max_pages` are not runtime-config-driven.** Legacy resolves `page_size` (default
  10, max 100) and `max_pages` (default 0/unlimited) from `RuntimeConfig.Config` at read time. The
  engine's `PaginationSpec.PageSize`/`MaxPages` are fixed values baked into `streams.json`'s
  `base.pagination` block with no per-read config override mechanism (same limitation the searxng
  golden documents for its own `page_size`/`max_pages`); `page_size: 10` matches legacy's own
  default. `page_size`/`max_pages` are still declared in `spec.json` for documentation/future-wiring
  purposes even though no template currently consumes them, since a future dialect increment may
  add config-driven pagination overrides.
- Full GNews API surface (this is the entirety of GNews's public catalog — search and
  top-headlines are the only two endpoints the API exposes) is fully covered; no `excluded` entries
  are needed in `api_surface.json`.
- Legacy's fixture-mode-only markers (`connector: "gnews"`, `fixture: true`, `previous_cursor`) are
  not modeled — these are credential-free conformance-harness affordances, not part of the live
  record shape; the engine's own conformance/fixture-replay harness provides the equivalent test
  affordance.
