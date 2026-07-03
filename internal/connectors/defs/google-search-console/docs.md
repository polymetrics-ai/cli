# Overview

Google Search Console reads site properties, submitted sitemaps, and Search Analytics performance
reports (aggregated by date, query, page, country, and device) through the Search Console v3 REST
API. This is a full legacy-parity migration of the hand-written connector
(`internal/connectors/google-search-console`), which stays registered and unchanged until wave6's
registry flip. Read-only: legacy sets `Capabilities.Write = false` and `Write` always returns
`ErrUnsupportedOperation`; this bundle matches with `capabilities.write: false` and no
`writes.json`.

`sites` and `sitemaps` are fully declarative (Tier 1). The 5 `search_analytics_by_*` streams are a
**Tier-2 StreamHook** (`internal/connectors/hooks/google-search-console/hooks.go`): the Search
Console `searchAnalytics.query` endpoint is a `POST` whose JSON request body carries
`startDate`/`endDate`/`dimensions`/`type`/`dataState`/`rowLimit`/`startRow`, and whose pagination
state (`startRow`, advanced by the number of rows returned each page) lives INSIDE that body —
`internal/connectors/engine/bundle.go`'s `StreamSpec.Body` field exists but
`internal/connectors/engine/read.go`'s declarative read path never sends a body at all (confirmed
at the `rt.Requester.Do(ctx, methodOrDefault(stream.Method), reqPath, query, nil)` call — the final
argument is a literal `nil`), so a POST-body read with in-body pagination cannot be expressed in
`streams.json` alone. This is a documented, sanctioned Tier-2 trigger (`docs/migration/
conventions.md` §1's "multipart/XML bodies... compound writes" list; monday's GraphQL migration is
the precedent for the identical `StreamSpec.Body`-is-unwired gap).

## Auth setup

Provide a Google OAuth 2.0 access token with Search Console read scope via the `access_token`
secret; it is used only for Bearer auth (`Authorization: Bearer <access_token>`) and is never
logged. `base_url` defaults to `https://www.googleapis.com/webmasters/v3` and may be overridden for
tests/proxies (validated as an absolute http/https URL with a host by the engine's own base-URL
resolution). The 3-legged OAuth consent/acquisition and refresh-token-exchange dance is out of
scope (the access token arrives pre-issued; the credentials layer already owns
acquisition/refresh) — matching legacy, which also only ever consumed a bare access token (`
authorization.access_token`, with a bare `access_token` fallback) and never itself performed a
refresh-token exchange.

## Streams notes

`sites` (`GET /sites`, `records.path: "siteEntry"`) lists every site property accessible to the
account; primary key `site_url`. `computed_fields` rename the raw API's camelCase `siteUrl`/
`permissionLevel` to the schema's `site_url`/`permission_level` (plain schema projection matches by
exact key name only).

`sitemaps` is **site-scoped**: legacy's own read path (`readMeta`, `google_search_console.go:
144-190`) loops over every configured site URL and issues one `GET /sites/{site}/sitemaps` request
per site, stamping `site_url` onto every emitted record. This bundle reproduces that exact pattern
with the engine's `stream.fan_out` dialect: `ids_from.config_key: site_urls` splits the configured
comma/newline-separated site list; `into.path_var` makes the resolved site URL referenceable in the
stream's own `path` as `{{ fanout.id }}` (`sites/{{ fanout.id }}/sitemaps`, urlencoded by
`InterpolatePath`'s default per-segment behavior — matching legacy's own `url.PathEscape(site)`);
`stamp_field: site_url` writes the current site URL onto every emitted record after
projection/computed_fields, exactly matching legacy's manual `rec["site_url"] = site` stamp.
`warnings`/`errors` are the Search Console API's own `string (int64 format)` wire fields (Google's
JSON convention for int64 values) — schema projection copies them by exact key match verbatim,
matching legacy's `stringify()` helper's normal-case behavior (a string value passes through
unchanged; `stringify`'s `fmt.Sprintf("%v", ...)` branch only guards a non-string edge case the API
never actually produces for these two fields).

The 5 `search_analytics_by_*` streams (`by_date`/`by_country`/`by_device`/`by_page`/`by_query`) are
also site-scoped, fanning out over the identical `site_urls` config the same way `sitemaps` does.
`hooks/google-search-console/hooks.go`'s `StreamHook.ReadStream` ports legacy's `readAnalytics`
(`google_search_console.go:207-263`) verbatim: for each configured site, it POSTs
`sites/{site}/searchAnalytics/query` with a JSON body built from `start_date`/`end_date` (or the
incremental cursor, a previously-synced date, as the lower bound), the stream's fixed one-dimension
`dimensions` array (`date`/`country`/`device`/`page`/`query`), `search_type`, `data_state`, and
`rowLimit: page_size`/`startRow`; the next page advances `startRow` by the number of rows just
received, stopping on a short (or empty) page — exactly legacy's own loop. `max_pages` (parsed
identically to legacy's `gscMaxPages`: absent/`"all"`/`"unlimited"` mean no cap; a non-negative
integer caps the page count per site) bounds this independently of the short-page stop signal, also
matching legacy exactly. Every row's positionally-aligned `keys[]` array is mapped back onto its
declared dimension name (`analyticsRecord`, ported verbatim), plus the fixed `clicks`/
`impressions`/`ctr`/`position` metric fields.

All 5 analytics streams publish `date` as their incremental cursor field (`x-cursor-field`), and
the hook's `analyticsDateRange` genuinely honors it as a server-side lower bound on `startDate` —
this is a true `request_param`-shaped incremental filter, just expressed inside a hook-managed
POST body instead of a declarative `stream.Query` entry, since the entire request is
hook-dispatched (§8 rule 2's truth table conceptually applies here even though the wiring lives in
Go, not JSON, because the transport itself is Tier-2).

## Write actions & risks

None. `capabilities.write` is `false`; no `writes.json` is shipped. Legacy itself implements no
write path for Search Console (`Write` always returns `ErrUnsupportedOperation`).

## Known limits

- **`StreamSpec.Body` is unwired (`ENGINE_GAP`, documented, non-blocking) — occurrence #2 after
  monday.** The engine's declarative read path never sends a request body, so the 5
  `search_analytics_by_*` streams' `POST` + in-body-pagination read cannot be expressed in
  `streams.json` alone. `hooks/google-search-console/hooks.go`'s `StreamHook` implements the real
  POST + in-body pagination entirely within the sanctioned Tier-2 hook seam, reusing `rt.Requester`
  (the engine's already-built HTTP client/auth/base-URL plumbing) exactly as the declarative path
  itself would. Per `docs/migration/conventions.md` §6's recurrence rule, a 3rd occurrence should
  promote this to a real engine feature (e.g. a `stream.body` templating dialect with pagination
  writing into named body fields) rather than a 3rd per-connector hook.
- **The 5 `search_analytics_by_*` streams' declarative `streams.json` path is never live-dispatched**
  (mirrors monday's identical note): each carries a `conformance.skip_dynamic` marker naming
  `hooks/google-search-console/hooks_test.go` as the authoritative substitute proving the real POST
  body construction, in-body pagination, and record mapping — conformance's dynamic (fixture-replay)
  checks Skip these streams outright rather than exercising a declarative GET-shaped fixture that
  could never match a POST-body read.
- **`access_token` scope narrowing**: legacy accepted the secret under either a dotted
  `authorization.access_token` key or a bare `access_token` fallback. This bundle declares a single
  `access_token` secret key (matching every other bearer-token bundle's convention, e.g.
  google-tasks' `api_key`) — a config-surface simplification, never a behavior change for any
  caller migrating off the dotted key naming (the credentials layer maps to whichever key name each
  connector's `spec.json` declares).
- **`records_limit`/`page_size`'s legacy-enforced bounds (1-25000) are not statically validated by
  the engine dialect** — `spec.json` declares `page_size` a plain `integer` with a default; an
  out-of-range value is passed straight to the live API rather than rejected client-side the way
  legacy's `gscPageSize` helper does. Not a data-parity issue (the API itself still rejects an
  invalid value), just a shifted validation boundary (same shape as google-tasks' identical, already
  ledgered deviation).
