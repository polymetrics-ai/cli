# Overview

Perk is a wave2 fan-out declarative-HTTP migration. It reads Perk/TravelPerk trips and invoices
through read-only REST list endpoints (`GET https://api.travelperk.com/...`). This bundle is
engine-vs-legacy parity-tested against `internal/connectors/perk` (the hand-written connector it
migrates); the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Perk/TravelPerk API key via the `api_key` secret; it is sent as an `Authorization: ApiKey
<api_key>` header (`auth.mode: api_key_header`, `header: Authorization`, `prefix: "ApiKey "`) and is
never logged, matching legacy's `connsdk.APIKeyHeader("Authorization", key, "ApiKey ")`
(`perk.go:128`). Every request also carries a static `Api-Version: 1` header, matching legacy's
`DefaultHeaders`. `base_url` defaults to `https://api.travelperk.com` and may be overridden for
tests/proxies.

## Streams notes

Both streams (`trips`, `invoices`) are `GET` list endpoints returning records at a top-level key
matching the stream name (`trips`/`invoices`). Pagination is offset+limit
(`pagination.type: offset_limit`, `limit_param: limit`, `offset_param: offset`, `page_size: 50`),
stopping on a short page — identical to legacy's `connsdk.OffsetPaginator{LimitParam: "limit",
OffsetParam: "offset", PageSize: size}`. `trips`' incremental cursor field is `modified`, sent as
`modified_gte`; `invoices`' incremental cursor field is `issuing_date`, sent as
`issuing_date_gte` — both computed from the sync's persisted cursor or, on a fresh sync, the
RFC3339 `start_date` config value, matching legacy's `firstNonEmpty(req.State["cursor"],
req.Config.Config["start_date"])` exactly (`param_format: rfc3339`, i.e. forwarded verbatim, same
as legacy's `base.Set(endpoint.startParam, start)` with no reformatting).

`invoices`' primary key is `serial_number` (not `id`) — legacy's own `streams()` declares
`PrimaryKey: []string{"serial_number"}` for this stream; this bundle's schema matches that exactly.

Both streams declare `"projection": "passthrough"` (post-wave2 review §8 rule 1): legacy's `Read`
emits `emit(connectors.Record(rec))` for both `trips` and `invoices` — a verbatim type-cast of the
raw harvested record, with no `mapRecord`-style field-building — so schema-mode projection would
silently drop any raw field this bundle's schema omits. Each schema remains a documentation surface
only; it does not gate what is emitted.

## Write actions & risks

None. Legacy `perk.Write` always returns `connectors.ErrUnsupportedOperation`; `capabilities.write`
is `false` and this bundle ships no `writes.json`.

## Known limits

- **`docs_url` points at the API host, not a docs page.** The bundle assignment's documentation
  reference for this connector is "manual intervention needed" (the public Perk/TravelPerk API
  docs span two separate hosts/brands and are not machine-reachable in a single canonical URL).
  Per `docs/migration/conventions.md` ("legacy is ground truth over any doc"), this bundle is
  authored entirely from the legacy Go package (`internal/connectors/perk/perk.go`), which is
  itself a conservative, docs-limited implementation — its own package comment notes it
  "conservatively reads documented TravelPerk list endpoints using Authorization: ApiKey" for
  exactly this reason. No behavior in this bundle depends on unreachable documentation; legacy code
  is the complete and sufficient source of truth here.
- Full Perk/TravelPerk API surface (trip creation, expenses, users) is out of scope; see
  `api_surface.json`'s `excluded: {category: out_of_scope}` entries — legacy itself never
  implemented these.
- `fixtures/streams/**` and `fixtures/check.json` use synthetic values only; real Perk/TravelPerk
  responses were not available (docs unreachable, no live credential), so fixture shapes are
  derived from legacy's own `readFixture` record shape and `perk_test.go`'s live-server fixture
  (`{"trips":[...],"limit":...,"offset":...,"total":...}`), which is itself an accurate rendering
  of the real wire envelope legacy's `Harvest`/`OffsetPaginator` consumes.
