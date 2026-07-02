# Overview

Twelve Data is a wave2 fan-out declarative-HTTP migration. It reads Twelve Data OHLCV time series,
latest quotes, and stock/forex-pair reference data through the Twelve Data REST API
(`GET https://api.twelvedata.com/<resource>`). This bundle is capability-parity migrated from
`internal/connectors/twelve-data` (the hand-written connector it migrates); the legacy package
stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Twelve Data API key via the `api_key` secret; it is sent as the `apikey` query parameter
on every request (an `api_key_query` auth candidate), matching legacy's
`connsdk.APIKeyQuery("apikey", key)` (`twelve_data.go:140`) and is never logged.

## Streams notes

- `time_series` — `GET /time_series?symbol=<symbol>&interval=<interval>&outputsize=<n>`, records
  at `values`, primary key `["symbol","datetime"]`, cursor field `datetime` (informational only —
  see Known limits). `symbol`/`interval`/`output_size` are all `spec.json` config values with
  legacy-matching defaults (`AAPL`/`1day`/`100`); `symbol` is additionally stamped onto every
  emitted record via `computed_fields` (`"symbol": "{{ config.symbol }}"`), matching legacy's
  `spec.includeSymbol` post-mapRecord override (`twelve_data.go:98-100`) — the configured symbol
  always wins over whatever the raw record itself carries (the raw `values[]` items carry no
  `symbol` field at all; Twelve Data returns it once at the sibling `meta.symbol` level).
- `quote` — `GET /quote?symbol=<symbol>`, a single-object response (`records.path: "."`), primary
  key `["symbol"]`. Same config-wins `computed_fields` symbol stamp as `time_series`, matching
  legacy's identical `spec.includeSymbol` override for this stream.
- `stocks` — `GET /stocks`, records at `data`, primary key `["symbol"]`. No `includeSymbol`
  override (legacy's `referenceRecord` copies `item["symbol"]` verbatim, since the raw API record
  already carries its own `symbol` field) — matched here by NOT declaring a `computed_fields`
  entry, letting plain schema projection copy the raw `symbol` through.
- `forex_pairs` — `GET /forex_pairs`, records at `data`, identical shape to `stocks`.

None of the 4 streams paginates: legacy issues exactly one request per `Read` call for every
stream (`twelve_data.go:85`), so no `pagination` block is declared anywhere in `streams.json`.

## Write actions & risks

None. Legacy `Write` always returns `connectors.ErrUnsupportedOperation`; `metadata.json` declares
`capabilities.write: false` and no `writes.json` file exists, matching legacy exactly.

## Known limits

- **`time_series`/`quote` have no incremental/server-side filter.** Legacy itself only supports
  full refresh for every stream (no `request_param`/incremental logic anywhere in
  `twelve_data.go`); this bundle declares no `incremental` block on any stream, matching legacy
  exactly. `datetime` is still declared as `x-cursor-field` on `time_series` (as legacy declares
  `CursorFields: []string{"datetime"}`) so downstream `*_deduped` sync modes remain available, but
  no request-time filtering happens on either side — every read returns Twelve Data's own
  `output_size`-bounded window of the most recent bars.
- **`output_size` bounds (1-5000) are not separately re-validated by this bundle.** Legacy clamps
  `output_size` to `[1, maxOutputSize=5000]` before sending it (`boundedInt`,
  `twelve_data.go:188-201`); this bundle sends `config.output_size` verbatim as the `outputsize`
  query parameter. Twelve Data's own API independently validates and rejects an out-of-range
  `outputsize` value with an error response, so no accepted-input behavior differs for any
  legacy-valid value — only the point at which an out-of-range value is rejected moves from
  client-side to server-side.
