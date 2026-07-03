# Overview

Finnhub exposes free real-time stock, forex, and crypto market data. This bundle reads 4 of 5
legacy streams — supported stock symbols for an exchange, global market news, and the two
symbol-partitioned `company_profile`/`stock_recommendations` streams — at capability parity with
`internal/connectors/finnhub` (the hand-written connector it migrates). The legacy package stays
registered and unchanged until wave6's registry flip. `company_profile`'s and
`stock_recommendations`' symbol partitioning, previously blocked (`ENGINE_GAP`, no fan-out
mechanism), is now expressed via `streams.json`'s `fan_out.ids_from.config_key` dialect (S4 engine
mini-wave item 2); see Streams notes. The 5th legacy stream, `company_news`, is **not** ported
here — its symbol partitioning is now expressible, but its `to` date-window upper bound is legacy's
`time.Now()` at read time, a dynamic value the engine has no mechanism to derive; see Known limits
and this migration's reported blocker.

## Auth setup

Provide a Finnhub API key via the `api_key` secret; it is sent only as the `X-Finnhub-Token`
header (`auth: [{"mode": "api_key_header", "header": "X-Finnhub-Token", ...}]`) and is never
logged. No other credential shape exists for this API.

## Streams notes

`stock_symbols` (`GET /stock/symbol`) is scoped by the `exchange` config value (default `US`,
matching legacy's `defaultExchange`); `market_news` (`GET /news`) is scoped by
`market_news_category` (default `general`, matching legacy's `defaultNewsCategory`). Both are
single-request, unpaginated GETs — Finnhub's array endpoints return their full result in one body
(no list pagination), matching legacy's `scopeExchange`/`scopeOnce` fan-out branches;
`records.path: ""` selects the root array.

`market_news`'s raw items carry a `related` field (comma-separated related tickers, often a single
symbol) rather than a plain `symbol` field; a `computed_fields` rename
(`"symbol": "{{ record.related }}"`) reproduces legacy's `newsRecord`'s fallback
(`rec["symbol"] = item["related"]` when the request was not itself symbol-scoped — always true for
this non-partitioned stream). The stream's `x-cursor-field: datetime` is declared for
downstream dedup/sort parity with legacy's published `CursorFields`, but — matching legacy exactly
— no `incremental` request-time filter is wired: legacy's `market_news`/`scopeOnce` branch never
applies a date-window query param (only the symbol-partitioned `company_news` stream does, and
that stream is not migrated here).

`company_profile` (`GET /stock/profile2`) and `stock_recommendations` (`GET /stock/recommendation`)
are symbol-partitioned: legacy's `harvest`'s `scopeSymbol` branch issues one GET per value in the
runtime `symbols` config list, stamping the requested symbol as a fallback onto each emitted record
(`internal/connectors/finnhub/finnhub.go`'s `harvest`, `streams.go`'s `companyProfileRecord`/
`recommendationRecord`). This bundle reproduces the fan-out with `stream.fan_out`:
`ids_from.config_key: symbols` splits/trims/upper-cases-nothing (see the case-normalization
deviation below) the comma-, whitespace-, or semicolon-separated `symbols` config value into an id
list; `into.query_param: symbol` adds the resolved symbol as a `?symbol=` query parameter on every
request of that symbol's sub-sequence (matching legacy's `url.Values{"symbol": []string{symbol}}`);
`stamp_field` (`ticker` for `company_profile`, `symbol` for `stock_recommendations`) writes the
current symbol onto every emitted record after projection, exactly matching legacy's fallback
stamp — `company_profile`'s `companyProfileRecord` and `stock_recommendations`'
`recommendationRecord` both only override `ticker`/`symbol` when the raw API response omits or
empties it (`if ticker == nil || ticker == ""`), while the engine's `fan_out.stamp_field` always
overwrites; this never changes emitted data for either endpoint's real response, since Finnhub's
`/stock/profile2` always echoes the requested ticker in its own `ticker` field and
`/stock/recommendation` always echoes the requested symbol in each item's `symbol` field — the
unconditional overwrite and the conditional fallback write the identical value for every
legacy-accepted input (documented parity deviation, ACCEPTABLE per conventions.md §5's meta-rule,
the same shape as finnworlds' `stamp_field` deviation). `company_profile`'s single-object response
and `stock_recommendations`' array-of-periods response are both handled by `records.path: ""`
(a bare object decodes as one whole-object record; an array decodes as N records — the same
`connsdk.RecordsAt` behavior every non-fan-out stream in this bundle already relies on).
Pagination, incremental state, and rate-limiting are independent per symbol (both streams declare
no pagination — Finnhub's array/object endpoints return a full result in one body — and no
`incremental` block, matching legacy's own unfiltered per-symbol harvest for both streams).

## Write actions & risks

None. Finnhub is a read-only market-data API in this bundle; `capabilities.write` is `false` and
no `writes.json` is declared, matching legacy's `Write` stub
(`connectors.ErrUnsupportedOperation`).

## Known limits

- **`company_news` is not migrated (ENGINE_GAP, blocked)**: like `company_profile`/
  `stock_recommendations`, `company_news` is symbol-partitioned via legacy's `harvest`'s
  `scopeSymbol` branch — the fan-out half of that gap is now closed by `stream.fan_out` (see Streams
  notes) and is no longer the blocker. What remains unexpressible is `company_news`'s `from`/`to`
  date-window: legacy's `dateWindow` (`internal/connectors/finnhub/finnhub.go:275-295`) computes
  `to` as `time.Now().UTC()` — literally "the moment this sync runs" — with no config override at
  all, and `from` as either the incremental cursor, the `start_date` config value, or (absent both)
  30 days before `to`. The declarative dialect's template namespaces (`config.*`, `secrets.*`,
  `record.*`, `cursor`, `incremental.lower_bound`, `fanout.id`) have no "current wall-clock time"
  pseudo-reference to derive `to` from — every existing bundle's date-range upper bound
  (`stockdata`'s `date_to`, `sparkpost`'s `to`, etc.) is a static operator-supplied config value,
  never a dynamically-computed "now". A fixed `default` value would silently go stale (it would stop
  matching "today" the moment real time moves past it), and requiring the operator to supply an
  explicit end date on every sync is an accepted-input-behavior change legacy never demands
  (§5's meta-rule: never data-changing for any legacy-accepted input — an operator who configures
  nothing at all is a legacy-accepted input, and it would newly fail here). Expressing this
  correctly needs either an engine "now" reference or a `StreamHook` (Tier 2), neither of which this
  wave's fan-out pass adds. The `start_date` config key legacy reads for `company_news` is therefore
  not declared in this bundle's `spec.json` (a declared-but-unwireable key is worse than an absent
  one) and `symbols` is documented as not driving `company_news` in this bundle (see `spec.json`'s
  description). Legacy's `finnhub` package remains the authoritative implementation for
  `company_news` until a follow-up wave adds an engine "now" primitive or a hook.
- **Documented parity deviation (ACCEPTABLE)**: `company_profile`'s and `stock_recommendations`'
  `stamp_field` overwrite is unconditional, unlike legacy's conditional fallback — see Streams notes
  for the full explanation; never data-changing for any real Finnhub response.
- **Documented parity deviation (ACCEPTABLE)**: legacy normalizes the `exchange` config value to
  upper-case (`finnhubExchange`'s `strings.ToUpper`) and `market_news_category` to lower-case
  (`finnhubNewsCategory`'s `strings.ToLower`) before sending either as a query parameter. The
  engine's template dialect has no case-transform filter (only `urlencode`/`unix_seconds`/
  `base64`/`join:<sep>`/`last_path_segment`/`const:<value>` exist), so this bundle sends whatever
  case the operator configures, verbatim. This never diverges from legacy for the common case
  (operators configuring the documented-case values `US`/`general`, or omitting the config
  entirely and getting the same default), and only differs for an operator who deliberately
  supplies a differently-cased override — a narrow, non-data-shape input-hygiene difference, not a
  change to any stream's record shape. `symbols` (new in this pass) is passed through to
  `?symbol=` verbatim per configured entry; legacy's `finnhubSymbols` upper-cases each symbol before
  use, so an operator who deliberately supplies lower-case symbols would see a case difference on
  the wire versus legacy — the same class of narrow, non-data-shape deviation as the `exchange`/
  `market_news_category` case above, for the same reason (no case-transform filter exists).
- 4 of 5 legacy streams are now implemented. Finnhub's much larger documented surface (real-time
  quotes, forex, crypto, financials) is out of scope until Pass B; see `api_surface.json`'s
  `excluded: {category: out_of_scope, reason: "Pass B capability expansion"}` entries.
- `metadata.json` declares no `rate_limit` block: legacy enforces no client-side rate limiting, so
  this bundle adds none either (matching legacy's real, absent throttling behavior).
