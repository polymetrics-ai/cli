# Overview

Finage is a real-time market-data API. This bundle reads the 4 non-partitioned US
market-information streams — most active stocks, top gainers, top losers, sector performance, and
delisted companies — at capability parity with `internal/connectors/finage` (the hand-written
connector it migrates). The legacy package stays registered and unchanged until wave6's registry
flip. The legacy connector's 6th stream, `market_news`, is **not** ported here — see Known limits
and this migration's reported blocker.

## Auth setup

Provide a Finage API key via the `api_key` secret; it is sent only as the `apikey` query parameter
(`auth: [{"mode": "api_key_query", "param": "apikey", ...}]`) and is never logged. No other
credential shape exists for this API.

## Streams notes

All 5 streams are single-request, unpaginated GETs — Finage's market-information list endpoints
return a full top-level JSON array per request, matching legacy's `fetchPage` (no pagination
declared; `records.path: ""` selects the root array). `most_active_us_stocks`, `most_gainers`, and
`most_losers` share an identical record shape (`symbol`, `company_name`, `change`,
`change_percentage`, `price`); `sector_performance` emits `sector`/`change_percentage`;
`delisted_companies` emits `symbol`/`company_name`/`exchange`/`ipo_date`/`delisted_date` and sends
the static query params `limit=1000`/`period=annual` legacy always sent for this endpoint. Every
raw field name is already the exact snake_case legacy emits, so no `computed_fields` renames are
needed for these 5 streams.

## Write actions & risks

None. Finage is a read-only market-data source; `capabilities.write` is `false` and no
`writes.json` is declared, matching legacy's `Write` stub (`connectors.ErrUnsupportedOperation`).

## Known limits

- **`market_news` is not migrated (ENGINE_GAP, blocked)**: legacy's `market_news` stream is
  symbol-partitioned — it issues one HTTP GET per value in the runtime `symbols` config list
  (`/news/market/{symbol}` fetched once per configured symbol, per
  `internal/connectors/finage/finage.go`'s `Read`/`configSymbols` fan-out). The declarative
  dialect's `stream.path`/`query` resolves to exactly one request path per pagination page; there
  is no mechanism (pagination type, computed_fields, or otherwise) to fan a single stream out
  across an arbitrary runtime-configured list of independent request values. Expressing this
  correctly requires a `StreamHook` (Tier 2), which this wave's fan-out pass does not create
  (Tier-2/3 hooks are a follow-up wave per `docs/migration/conventions.md` §1's hard rule). The
  `symbols` config key is therefore not declared in this bundle's `spec.json` (a
  declared-but-unwireable key is worse than an absent one). Legacy's `finage` package remains the
  authoritative implementation for `market_news` until a follow-up wave adds the hook.
- Only the 4 legacy-parity non-partitioned streams plus `delisted_companies` are implemented (5 of
  6 legacy streams). Finage's much larger documented surface (real-time quotes, forex, crypto,
  per-symbol fundamentals) is out of scope until Pass B; see `api_surface.json`'s
  `excluded: {category: out_of_scope, reason: "Pass B capability expansion"}` entries.
- `metadata.json` declares no `rate_limit` block: legacy enforces no client-side rate limiting, so
  this bundle adds none either (matching legacy's real, absent throttling behavior).
