# Overview

Reads a single coin's current metadata/market-data snapshot and its exchange tickers from the
CoinGecko REST API (`GET /coins/{id}`, `GET /coins/{id}/tickers`), migrating
`internal/connectors/coingecko-coins` (the legacy hand-written connector, which stays registered
and unchanged until wave6's registry flip) at capability parity for its `coin` stream, plus a new
`tickers` stream added in this Pass B pass. This bundle is scoped deliberately to the single-coin
(coin_id-configured) resource family CoinGecko's API groups under `/coins/{id}/**` — not the full
CoinGecko API (search, exchanges, derivatives, NFTs, on-chain data, global stats, and more), which
is a distinct, much larger product surface with no per-connection coin_id scoping in common with
this bundle; see `api_surface.json` for the full per-endpoint review. The connector was previously
quarantined entirely (`ENGINE_GAP`, `docs/migration/quarantine.json`) for `market_chart`'s array-zip
requirement, then unblocked for `coin` alone in a prior pass. `market_chart`, `history`, and (newly
reviewed this pass) `ohlc`/`ohlc/range`/`market_chart/range` remain blocked; see Known limits. The
API is read-only; this bundle exposes no write actions.

## Auth setup

CoinGecko allows unauthenticated calls against its public base URL; an optional pro `api_key`
(secret) unlocks the pro base URL and higher rate limits, sent as the `x-cg-pro-api-key` header.
`streams.json`'s `base.auth` declares `api_key_header` gated by `when: {{ secrets.api_key }}` (sent
only when the secret is configured), falling back to `mode: none` (searxng's golden optional-secret
pattern, `conventions.md` §3). `base_url` is a **required** config value (see Known limits for why
this bundle cannot derive it the way legacy does) — set it to
`https://api.coingecko.com/api/v3` (public) or `https://pro-api.coingecko.com/api/v3` (pro, when
`api_key` is set).

## Streams notes

`coin` (`GET /coins/{coin_id}`, the path's `{{ config.coin_id }}` substitutes the connection's
configured coin, e.g. `bitcoin`) sends the same
`localization=false&tickers=false&community_data=false&developer_data=false` static query legacy
sends, and projects the identical field set legacy's `coinRecord` emits (`id`, `symbol`, `name`,
`market_cap_rank`, `hashing_algorithm`, `categories`, `market_data`, `last_updated`). The response
body IS the record (`records.path: "."`) — CoinGecko's single-coin endpoint has no envelope
wrapper, matching legacy's direct `r.DoJSON(..., &item)` decode. Not paginated (a single-object
response); not incremental (legacy publishes no `CursorFields` for `coin`).

**`tickers`** (`GET /coins/{coin_id}/tickers`, new in this Pass B pass — legacy never implemented
this stream): records at `tickers` (the response envelope is `{name, tickers: [...]}`; `name` — the
coin's display name — is not itself emitted as a record field). CoinGecko documents 100 tickers per
page with no page-size query override and no pagination metadata in the response body itself, so
this is a `page_number` paginator with `size_param: ""` (matching searxng's "no page-size param is
ever sent" pattern) and `page_size: 100` as the client-side short-page stop threshold. Primary key
is the `(coin_id, target_coin_id, market_identifier)` triple — a single coin/target pair can be
traded on multiple exchanges, so the exchange identifier is part of identity;
`market_identifier` is a `computed_fields` derivation (`{{ record.market.identifier }}`) since the
raw wire field is nested inside the `market` sub-object, not a top-level field. Every other declared
field matches the raw wire field name exactly, so schema-mode projection copies them without any
further `computed_fields` entries.

## Write actions & risks

None. CoinGecko is `capabilities.write: false`; no `writes.json` is shipped, matching legacy's
`Write` always returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **Blocked: `market_chart` stream (`ENGINE_GAP`, unchanged from the original quarantine finding).**
  Legacy's `readMarketChart` requests `GET /coins/{id}/market_chart`, which returns three PARALLEL
  arrays of `[unix_ms, value]` pairs (`prices`/`market_caps`/`total_volumes`), and zips them
  index-by-timestamp (`indexByTimestamp`) into one record per timestamp. The dialect has no
  array-zip/parallel-array-join primitive: `records.path`'s extraction (`connsdk.RecordsAt`) only
  turns a JSON *object* array into records (it silently drops any array element that is not itself
  a `map[string]any`, so a bare `[ts, value]` pair array yields zero records outright); `join:<sep>`
  only joins ONE array into a delimited string, never zips multiple arrays by index; `computed_fields`
  templates can only walk `map[string]any` path segments (`resolveRecordPathValue`), so they cannot
  address an array element by numeric position either. Re-checked against every S3/S4 mini-wave
  addition (`fan_out`, `keyed_object`, typed `computed_fields` extraction, `const:`/`last_path_segment`
  filters, `incremental.lower_bound`) — none provide array-position addressing or multi-array
  zipping. A Tier-1 workaround would either silently drop `market_cap`/`total_volume` (an
  accepted-input record-data change, forbidden by the §5 meta-rule) or emit one record per array
  with no way to correlate them by timestamp. Would need either a new `records`-extraction primitive
  (e.g. a "zip these N array paths by index" spec) or a Tier-2 `StreamHook`; a `StreamHook` alone
  would still need a 2nd interface for nothing else in this connector, so — per this bundle's minimal
  scope — it is filed as `ENGINE_GAP` rather than escalated to a hooks package for one stream.
- **Blocked: `market_chart/range`, `ohlc`, and `ohlc/range` (`ENGINE_GAP`, newly reviewed this Pass
  B pass — same root cause as `market_chart`, not new distinct gaps).** `market_chart/range` is the
  identical 3-parallel-array-zip shape as `market_chart` (a client-supplied `from`/`to` Unix range
  has no bearing on the array-shape gap itself). `ohlc`/`ohlc/range` return a bare array of
  `[timestamp, open, high, low, close]` 5-element tuples — each array ELEMENT is itself a JSON
  array, not an object. `connsdk.RecordsAt`'s `[]any` branch only keeps elements that decode as
  `map[string]any`; a bare-tuple element fails that check and is silently dropped, yielding zero
  records for every page — confirmed by reading `RecordsAt`'s implementation directly in this pass
  rather than assumed by analogy. This is the exact same "no array-of-arrays/tuple-to-record
  primitive" limitation as `market_chart`'s array-zip gap, not a second distinct gap to track
  separately.
- **Blocked: `history` stream (`ENGINE_GAP`).** Legacy's
  `readHistory` walks one independent `GET /coins/{id}/history?date=DD-MM-YYYY` request per
  calendar day from `start_date` to `end_date` (inclusive) — the `date` query param IS the loop
  variable; CoinGecko's `history` endpoint has no range/from/to filter at all, only a single
  literal date per request. This does not fit any read shape the dialect expresses: it is not
  pagination (no cursor/offset/page-number state advances between requests; each day's request is
  wholly independent), it is not `incremental` (that mechanism issues exactly ONE request per read
  with a server-side lower-bound filter — see marketstack's `eod`/`param_format: date` for the
  correct use of that shape — never a client-side fan-out of N independent requests), and it is not
  `fan_out` (`ids_from` requires a discoverable id list via a config-held comma-separated value or a
  preliminary listing request; a client-computed contiguous calendar-date range between two config
  bounds is neither). Confirmed against the newly-available date-only lower-bound parsing
  (`parseLowerBoundTime`'s bare `YYYY-MM-DD` support) — that only widens what a SINGLE incremental
  request's lower-bound value can look like, it does not add a day-walking request-fan-out
  mechanism. Would need either a new pagination/fan-out type (a "date-range" driver: one request per
  day between two config-resolved bounds) or a Tier-2 `StreamHook`.
- `spec.json` requires `base_url` explicitly rather than defaulting it (legacy derives
  `https://pro-api.coingecko.com/api/v3` vs. `https://api.coingecko.com/api/v3` from whether
  `api_key` is set): `spec.json`'s `"default"` materialization fills in a fixed literal only, with
  no mechanism to make that literal conditional on another config/secret's presence. This never
  changes any emitted record's data — it only makes explicit a base-URL choice legacy inferred —
  but every connection MUST now set `base_url` itself. Parity-deviation ledger candidate,
  ACCEPTABLE under the meta-rule (no accepted-input record-data change; the same value legacy would
  have derived is simply now required input instead of an inferred default).
- Full CoinGecko API surface remains out of scope beyond `coin`/`tickers` above: this connector is
  deliberately scoped to the coin_id-scoped `/coins/{id}/**` resource family, not the much larger
  general CoinGecko surface (multi-coin listings, search, exchanges, derivatives, NFTs, on-chain
  data, categories, global stats, and more — none of which share this connection's per-coin_id
  scoping shape). See `api_surface.json` for the full per-endpoint review (every documented v3
  endpoint reviewed, not just the coin-scoped ones) with a specific real reason per exclusion —
  mostly `out_of_scope` (a distinct resource domain from coins) or `requires_elevated_scope`
  (documented as an Analyst-plan-or-above CoinGecko API feature, unavailable on the public/pro
  tiers this connector's `spec.json` models).
