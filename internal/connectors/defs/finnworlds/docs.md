# Overview

Finnworlds is an unquarantine migration of `internal/connectors/finnworlds` (`finnworlds.go` +
`streams.go`, the read-only legacy connector this bundle migrates; the legacy package stays
registered and unchanged until wave6's registry flip). It reads global financial data (dividends,
stock splits, historical candlesticks, and commodity prices) from the Finnworlds REST API
(`https://api.finnworlds.com/api/v1`). Finnworlds has no page-token pagination: each request
returns the full dataset for one partition value (a ticker or a commodity name), so legacy fans out
across a configured list, one request per value — this is exactly the shape the engine's `fan_out`
dialect (`streams.json`'s `fan_out.ids_from.config_key`) was added to express (S4 engine mini-wave
item 2), which is why this connector was previously quarantined as an `ENGINE_GAP` and is now
buildable. Read-only: legacy's `Write` always returns `connectors.ErrUnsupportedOperation`
(`finnworlds.go:300-302`), and this bundle declares `capabilities.write: false` with no
`writes.json` to match.

## Auth setup

Provide one secret, `key` (Finnworlds API key), required. It is sent as the `key` query parameter
on every request (`base.auth`'s `api_key_query` mode), matching legacy's `connsdk.APIKeyQuery("key",
secret)` (`finnworlds.go:222`) exactly. `base_url` defaults to `https://api.finnworlds.com/api/v1`
and may be overridden for tests/proxies.

## Streams notes

All 4 streams share the identical fan-out shape: `fan_out.ids_from.config_key` names a comma-separated
config list (`tickers` for `dividends`/`stock_splits`/`historical_candlestick`, `commodities` for
`commodities`), `into.query_param` names the query parameter the resolved value is injected as
(`ticker`/`commodity_name`, matching legacy's `endpoint.partitionKey`, `streams.go:28`), and
`stamp_field` writes the current partition value onto every emitted record of that sub-sequence
(matching legacy's `stitch` logic, `finnworlds.go:130-135,152-156`) — one full read/pagination
sequence (here, unpaginated: `pagination.type: none`, matching legacy's own single-request-per-value
harvest with no page-token support at all, `finnworlds.go:119-123`) per configured value.
Each stream declares a bare `incremental.cursor_field` matching legacy's `CursorFields` catalog
metadata (`date` or `datetime`) without any `request_param`, because legacy never sends a
server-side cursor filter.

- **`dividends`** (`GET /dividends?ticker=<value>`): records at `result.output.dividends` (Finnworlds
  wraps every response as `{"result":{"output": ...}}`; this stream nests one level deeper than most
  — legacy's `recordsPath: "result.output.dividends"`, `streams.go:39`). Primary key
  `["ticker", "date"]`, cursor field `date`, matching legacy's catalog (`streams.go:76-79`).
- **`stock_splits`** (`GET /stocksplits?ticker=<value>`): records at `result.output.stocksplits`
  (same nested-one-level-deeper shape, `streams.go:47`). Primary key `["ticker", "date"]`, cursor
  field `date` (`streams.go:83-86`).
- **`historical_candlestick`** (`GET /historicalcandlestick?ticker=<value>`): records at
  `result.output` (the array lives directly at that path, no further nesting, `streams.go:55`).
  Primary key `["ticker", "date"]`, cursor field `date` (`streams.go:89-93`). `opentime`/`closetime`
  are copied as their real Finnworlds wire type (JSON numbers, per legacy's `int64` fields,
  `finnworlds.go:186-187`), matching schema types `["integer","null"]`; every other candlestick
  field (`open`/`high`/`low`/`close`/`adjusted_close`/`trade_volume`) is a string in the real API
  response, matching legacy's own `string`-typed `connectors.Field`s (`streams.go:121-134`).
- **`commodities`** (`GET /commodities?commodity_name=<value>`): records at `result.output`
  (`streams.go:63`). Primary key `["commodity_name", "datetime"]`, cursor field `datetime`
  (`streams.go:96-100`). Also doubles as the `check` request (unfiltered `GET /commodities`, no
  `commodity_name` query param — mirrors legacy's `Check`, which reads the commodities endpoint
  unconditionally to confirm the key and connectivity without depending on ticker configuration,
  `finnworlds.go:80-85`).

Every field name on every stream's raw Finnworlds API response already matches its schema property
name exactly — plain schema-mode projection copies every field by exact key match with zero
`computed_fields` needed, preserving legacy's field-built `connectors.Record{...}` mapping
(`streams.go:150-193`) field-for-field.

## Write actions & risks

None — Finnworlds is read-only. `capabilities.write: false`, no `writes.json` file, matching
legacy's `ErrUnsupportedOperation` (`finnworlds.go:300-302`): Finnworlds is a market-data source with
no reverse-ETL write surface.

## Known limits

- **`stamp_field`'s overwrite semantics are unconditional, unlike legacy's conditional stitch.**
  Legacy only writes the partition value onto `record[stitchField]` when that field is absent, nil,
  or empty (`finnworlds.go:152-156`: `if existing, ok := record[endpoint.stitchField]; !ok ||
  existing == nil || existing == "" { ... }`); the engine's `fan_out.stamp_field` (`read.go:355-357`)
  always overwrites `projected[fc.stampField] = fc.id` regardless of whether the raw record already
  carried a value there. In practice this never changes emitted data for any real Finnworlds
  response: every one of the 4 endpoints is queried scoped to exactly one ticker/commodity value at
  a time, and the API's own response for `dividends`/`stock_splits`/`historical_candlestick`/
  `commodities` always echoes that SAME ticker/commodity_name back on every record (legacy's own
  `mapRecord` functions copy `item["ticker"]`/`item["commodity_name"]` directly from the API
  response, `streams.go:150-192`) — so the unconditional overwrite and the conditional stitch write
  the identical value in every legacy-accepted case. Documented per conventions.md §5's meta-rule as
  ACCEPTABLE (never data-changing for any legacy-accepted input), not silently absorbed.
- **No config-driven `page_size`/`max_pages` override** — moot for this connector: Finnworlds has no
  page-token pagination at all (`pagination.type: none` on every stream, matching legacy's own
  single-request-per-partition-value harvest, `finnworlds.go:119-123`), so neither concept applies.
