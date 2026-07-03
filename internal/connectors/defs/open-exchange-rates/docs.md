# Overview

Open Exchange Rates was quarantined during wave1 for an `ENGINE_GAP`: every stream (`latest`,
`historical`, `currencies`) flattens a raw JSON OBJECT into N records (one per map key —
currencies-in-rates for `latest`/`historical`, currency-code-to-name for `currencies`) and, at the
time, the engine's `records.path` extraction only ever yielded exactly ONE record for an object
path (the whole object passed through verbatim), never one record per key. The S4 engine mini-wave
since added `records.keyed_object`/`records.key_field` — but that mechanism explodes a map of
**object-valued** entries (`{"111": {...}, "222": {...}}`, e.g. appfigures' `products`), not a map
of **scalar-valued** entries. Open Exchange Rates' real wire shape is scalar-valued
(`"rates": {"EUR": 0.92, "GBP": 0.79, ...}`, `{"EUR": "Euro", ...}` for `currencies`) — every value
under the exploded key is a bare number or string, not a JSON object. `recordsAtKeyed`
(`internal/connectors/engine/read.go`) type-asserts each value to `map[string]any` and **silently
skips** any key whose value fails that assertion (`read.go`'s `valObj, ok :=
obj[k].(map[string]any); if !ok { continue }`) — declaring `keyed_object: true` against a
scalar-valued map would silently emit ZERO records for every stream, a silent data-loss bug, not a
working migration. This is therefore a genuinely DIFFERENT, still-open `ENGINE_GAP` from what the
S4 mini-wave closed; see Known limits for the per-stream detail. This bundle is a **partial**
unblock: only the `usage` stream (a single nested-object record, no map explosion needed) is
declarative-expressible and ported. The legacy `internal/connectors/open-exchange-rates` package
stays registered and authoritative for `latest`/`historical`/`currencies` until a future engine
increment adds scalar-valued keyed-object support (or a Tier-2/3 escalation is approved).

## Auth setup

Provide an Open Exchange Rates `app_id` secret; it is sent as the `app_id` query parameter
(`api_key_query` auth mode) on every request, matching legacy's `connsdk.APIKeyQuery("app_id",
secret)` exactly. Never logged.

## Streams notes

`usage` (`GET /usage.json`) is the only ported stream. The response is a single nested-object
payload (`{"status": 200, "data": {"app_id", "status", "plan": {"name"}, "usage": {...}}}`);
`records.path: "data"` + `single_object: true` selects the `data` object as the one emitted
record, and `computed_fields` reaches into the nested `plan.name` and every `usage.*` sub-field to
flatten them onto the top-level record — exactly matching legacy's `usageRecord` mapping
(`internal/connectors/open-exchange-rates/open-exchange-rates.go`'s `usageRecord` function).
Primary key is `app_id`. No `incremental` block is declared: legacy's `readUsage` sends no
cursor/filter parameter at all (a full-refresh-only snapshot stream), matching conventions.md §8
rule 2.

## Write actions & risks

None. Open Exchange Rates is read-only in both legacy and this bundle (`capabilities.write:
false`, no `writes.json`).

## Known limits

- **`latest` is not ported (blocked, `ENGINE_GAP`).** `GET /latest.json` returns
  `{"timestamp", "base", "rates": {"EUR": 0.92, "GBP": 0.79, ...}}`; legacy's `rateRecords` flattens
  `rates` into one `{base, currency, rate, timestamp}` record per currency code
  (`sortedKeys(rates)` for determinism). `records.keyed_object` cannot express this: every value
  under `rates` is a bare JSON number, and `recordsAtKeyed` requires each exploded value to itself
  decode as a `map[string]any` — a non-object value is silently skipped, which would emit zero
  records for this stream rather than one per currency. There is also no declarative way to stamp
  the SIBLING `timestamp`/`base` fields (which live one level up from `rates`, not on each
  currency's own "record") onto every exploded record once `records.path` points at `rates` — that
  would need `computed_fields` to reach outside the current record into the parent page body,
  which the dialect does not support (`computed_fields` templates resolve only against the current
  record + `config.*`, per conventions.md §3).
- **`historical` is not ported (blocked, `ENGINE_GAP`).** Same scalar-valued keyed-object gap as
  `latest` (identical `rates` shape at `historical/{date}.json`), compounded by a second gap:
  legacy's `readHistorical` issues one request per calendar date from the incremental lower bound
  through `end_date` (up to `oerMaxHistoricalDays` = 366 requests within a single `Read` call) —
  none of the engine's 6 pagination types can drive a per-call-varying PATH segment computed from a
  date-sequence loop counter (this is the same date-walk gap already documented as blocked for the
  sibling `exchange-rates` connector's `exchange_rates` stream, see
  `internal/connectors/defs/exchange-rates/docs.md`).
- **`currencies` is not ported (blocked, `ENGINE_GAP`).** `GET /currencies.json` returns a bare
  `{"EUR": "Euro", "GBP": "British Pound Sterling", ...}` map; legacy's `currencyRecords` explodes
  it into one `{currency, name}` record per key. Same scalar-valued (string, not object)
  keyed-object gap as `latest`/`historical` above — `name` is a bare JSON string, never an object,
  so `keyed_object` would again silently skip every key.
- Full Open Exchange Rates API surface is otherwise limited to these 4 endpoints; see
  `api_surface.json`'s `excluded` entries for the 3 blocked ones with fresh, specific reasons
  distinct from the original quarantine.json entry (which predates `keyed_object`'s
  object-valued-only implementation).
