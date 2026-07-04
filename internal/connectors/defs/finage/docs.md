# Overview

Finage is a real-time market-data API. This bundle reads all 6 legacy streams — the 4
non-partitioned US market-information streams (most active stocks, top gainers, top losers, sector
performance), delisted companies, and the symbol-partitioned `market_news` stream — at capability
parity with `internal/connectors/finage` (the hand-written connector it migrates), PLUS two Pass B
full-surface-expansion additions within the same Fundamentals product family: `earnings_calendar`
and `ipo_calendar`. The legacy package stays registered and unchanged until wave6's registry flip.
`market_news`'s symbol partitioning, previously blocked (`ENGINE_GAP`, no fan-out mechanism), is now
expressed via `streams.json`'s `fan_out.ids_from.config_key` dialect (S4 engine mini-wave item 2);
see Streams notes.

## Auth setup

Provide a Finage API key via the `api_key` secret; it is sent only as the `apikey` query parameter
(`auth: [{"mode": "api_key_query", "param": "apikey", ...}]`) and is never logged. No other
credential shape exists for this API.

## Streams notes

The 4 non-partitioned market-information streams plus `delisted_companies` are single-request,
unpaginated GETs — Finage's market-information list endpoints return a full top-level JSON array
per request, matching legacy's `fetchPage` (no pagination declared; `records.path: ""` selects the
root array). `most_active_us_stocks`, `most_gainers`, and `most_losers` share an identical record
shape (`symbol`, `company_name`, `change`, `change_percentage`, `price`); `sector_performance` emits
`sector`/`change_percentage`; `delisted_companies` emits
`symbol`/`company_name`/`exchange`/`ipo_date`/`delisted_date` and sends the static query params
`limit=1000`/`period=annual` legacy always sent for this endpoint. Every raw field name is already
the exact snake_case legacy emits, so no `computed_fields` renames are needed for these 5 streams.

`market_news` (`GET /news/market/{symbol}`) is symbol-partitioned: legacy issues one GET per value
in the runtime `symbols` config list, stamping the requested symbol onto each emitted record
(`internal/connectors/finage/finage.go`'s `Read`/`configSymbols` fan-out,
`streams.go`'s `finageNewsRecord`). This bundle reproduces that exactly with `stream.fan_out`:
`ids_from.config_key: symbols` splits/trims the comma- or whitespace-separated `symbols` config
value into an id list (matching legacy's `configSymbols` parsing rules); `into.path_var` makes the
resolved symbol referenceable in the stream's own `path` as `{{ fanout.id }}`
(`/news/market/{{ fanout.id }}`); `stamp_field: symbol` writes the current symbol onto every emitted
record after projection, exactly matching legacy's `item["symbol"] = symbol` fallback stamp (legacy
only stamps when the raw record has no `symbol` field at all — Finage's `/news/market/{symbol}`
response never includes one, so the fan_out's unconditional overwrite and legacy's conditional
stamp write the identical value for every real response). The static query param `limit=30` matches
legacy's `extraParams`; `records.path: "news"` selects Finage's `{"symbol":...,"news":[...]}`
envelope for this endpoint (unlike the root-array shape the other 5 streams use). Pagination,
incremental state, and rate-limiting are independent per symbol, mirroring legacy's own
per-symbol `fetchPage` call — this stream declares no pagination (Finage's news endpoint returns a
full array per symbol, no pagination) and no `incremental` block (legacy's `market_news` reads no
date-window filter either).

`earnings_calendar` (`GET /fnd/earning-calendar`, Pass B addition) and `ipo_calendar`
(`GET /fnd/ipo-calendar`, Pass B addition) are both unpaginated, top-level-array GETs in the same
`/fnd/*` Fundamentals family as `delisted_companies`, requiring a `from`/`to` (`YYYY-MM-DD`)
date-range window — plain (non-optional) `{{ config.calendar_from }}`/`{{ config.calendar_to }}`
query templates, so omitting either hard-errors exactly like every other required-but-unset
`config.*` reference in this dialect (no `omit_when_absent`/`default` — there is no sane default
date range to fall back to, and Finage's own docs never document one). Field names for both streams
are derived from a documented real-world usage example of these exact endpoints (not guessed —
`earnings_calendar`: `symbol`/`date`/`time`/`eps`/`estimated_eps`/`revenue`/`estimated_revenue`,
with `eps`/`revenue` nullable per that sample; `ipo_calendar`:
`symbol`/`date`/`company`/`exchange`/`status`/`shares`/`price_range`/`market_cap`), since Finage's
own documentation site renders its endpoint reference pages client-side (every URL fetched returned
identical generic template content, not the actual per-endpoint reference) and no OpenAPI/Postman
spec is publicly published. Neither stream has a legacy counterpart to derive fields from — legacy
never modeled either endpoint at all.

## Write actions & risks

None. Finage is a read-only market-data source; `capabilities.write` is `false` and no
`writes.json` is declared, matching legacy's `Write` stub (`connectors.ErrUnsupportedOperation`).

## Known limits

- **`stamp_field`'s overwrite semantics are unconditional, unlike legacy's conditional stamp**
  (documented parity deviation, ACCEPTABLE per conventions.md §5's meta-rule): legacy only writes
  the requested symbol onto `item["symbol"]` when that key is absent (`finage.go`'s `fetchPage`:
  `if _, ok := item["symbol"]; !ok { item["symbol"] = symbol }`), while the engine's
  `fan_out.stamp_field` always overwrites the projected record's `symbol` field regardless. This
  never changes emitted data for any real Finage response: `/news/market/{symbol}` never includes a
  `symbol` field on its news items, so the unconditional overwrite and the conditional stamp write
  the identical value for every legacy-accepted input.
- All 6 legacy streams are now implemented. Finage's much larger documented surface (real-time
  quotes, forex, crypto, per-symbol fundamentals) is out of scope until Pass B; see
  `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass B capability expansion"}`
  entries.
- `metadata.json` declares no `rate_limit` block: legacy enforces no client-side rate limiting, so
  this bundle adds none either (matching legacy's real, absent throttling behavior).
