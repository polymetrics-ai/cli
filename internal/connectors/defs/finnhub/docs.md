# Overview

Finnhub exposes free real-time stock, forex, and crypto market data. This bundle reads the 2
non-partitioned streams — supported stock symbols for an exchange, and global market news — at
capability parity with `internal/connectors/finnhub` (the hand-written connector it migrates). The
legacy package stays registered and unchanged until wave6's registry flip. Legacy's 3
symbol-partitioned streams (`company_news`, `company_profile`, `stock_recommendations`) are **not**
ported here — see Known limits and this migration's reported blocker.

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

## Write actions & risks

None. Finnhub is a read-only market-data API in this bundle; `capabilities.write` is `false` and
no `writes.json` is declared, matching legacy's `Write` stub
(`connectors.ErrUnsupportedOperation`).

## Known limits

- **`company_news`, `company_profile`, `stock_recommendations` are not migrated (ENGINE_GAP,
  blocked)**: all three are symbol-partitioned in legacy — `harvest`'s `scopeSymbol` branch issues
  one HTTP GET per value in the runtime `symbols` config list
  (`internal/connectors/finnhub/finnhub.go`). The declarative dialect's `stream.path`/`query`
  resolves to exactly one request path per pagination page; there is no mechanism to fan a single
  stream out across an arbitrary runtime-configured list of independent request values (the exact
  same gap class as finage's `market_news` and finnworlds' partitioned streams). Additionally,
  `company_profile`'s and `stock_recommendations`' record mappers apply a conditional
  field-fallback (`ticker`/`symbol` defaults to the request-scoped symbol only when the raw API
  response omits it — `companyProfileRecord`/`recommendationRecord`'s `if ticker == nil || ticker
  == ""`) that `computed_fields` cannot express (no conditional "use A if present, else B"
  primitive; only rename/join/const/static-literal/bare-typed-copy). Both limitations require a
  `StreamHook` (Tier 2), which this wave's fan-out pass does not create (Tier-2/3 hooks are a
  follow-up wave per `docs/migration/conventions.md` §1's hard rule). The `symbols`/`start_date`
  config keys legacy reads for these 3 streams are therefore not declared in this bundle's
  `spec.json`. Legacy's `finnhub` package remains the authoritative implementation for these
  streams until a follow-up wave adds the hook.
- **Documented parity deviation (ACCEPTABLE)**: legacy normalizes the `exchange` config value to
  upper-case (`finnhubExchange`'s `strings.ToUpper`) and `market_news_category` to lower-case
  (`finnhubNewsCategory`'s `strings.ToLower`) before sending either as a query parameter. The
  engine's template dialect has no case-transform filter (only `urlencode`/`unix_seconds`/
  `base64`/`join:<sep>`/`last_path_segment`/`const:<value>` exist), so this bundle sends whatever
  case the operator configures, verbatim. This never diverges from legacy for the common case
  (operators configuring the documented-case values `US`/`general`, or omitting the config
  entirely and getting the same default), and only differs for an operator who deliberately
  supplies a differently-cased override — a narrow, non-data-shape input-hygiene difference, not a
  change to any stream's record shape.
- Only the 2 legacy-parity non-partitioned streams are implemented (2 of 5 legacy streams).
  Finnhub's much larger documented surface (real-time quotes, forex, crypto, financials) is out of
  scope until Pass B; see `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass B
  capability expansion"}` entries.
- `metadata.json` declares no `rate_limit` block: legacy enforces no client-side rate limiting, so
  this bundle adds none either (matching legacy's real, absent throttling behavior).
