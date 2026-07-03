# Overview

Reads a single coin's current metadata/market-data snapshot from the CoinGecko REST API
(`GET /coins/{id}`), migrating `internal/connectors/coingecko-coins` (the legacy hand-written
connector, which stays registered and unchanged until wave6's registry flip) at capability parity
for its `coin` stream only. This bundle is an **unblock re-review**, not a fresh migration: the
connector was previously quarantined entirely (`ENGINE_GAP`, `docs/migration/quarantine.json`) for
`market_chart`'s array-zip requirement. Re-reviewed against the new dialect additions (`fan_out`,
`keyed_object`, `start_page: 0`, oauth2 `extra_params`, date-only incremental bounds) — none of
those apply to this connector's remaining gaps. `market_chart` and `history` remain blocked; see
Known limits. The API is read-only; this bundle exposes no write actions.

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

Only `coin` is implemented. It reads `GET /coins/{coin_id}` (the path's `{{ config.coin_id }}`
substitutes the connection's configured coin, e.g. `bitcoin`) with the same
`localization=false&tickers=false&community_data=false&developer_data=false` static query legacy
sends, and projects the identical field set legacy's `coinRecord` emits (`id`, `symbol`, `name`,
`market_cap_rank`, `hashing_algorithm`, `categories`, `market_data`, `last_updated`). The response
body IS the record (`records.path: "."`) — CoinGecko's single-coin endpoint has no envelope
wrapper, matching legacy's direct `r.DoJSON(..., &item)` decode. Not paginated (a single-object
response); not incremental (legacy publishes no `CursorFields` for `coin`).

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
- **Blocked: `history` stream (`ENGINE_GAP`, newly identified in this unblock pass).** Legacy's
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
- Full CoinGecko API surface (search, exchanges, derivatives, NFTs, on-chain data, and the `ping`
  endpoint used only as the check probe) is out of scope until Pass B; see `api_surface.json`.
