# Overview

Watchmode is a read-only declarative-HTTP connector for the Watchmode Streaming Availability API
(v1). It reads title search results, streaming source/network/region/genre reference data,
paginated title listings, recent/upcoming releases, per-title details/streaming-sources/
seasons/episodes/cast-crew (fanned out over a configured list of title IDs), and per-person
details (fanned out over a configured list of person IDs). This bundle was originally migrated
from `internal/connectors/watchmode` (legacy wave2 fan-out: `search`/`sources` only) and has since
been expanded to the full practical Watchmode v1 surface (Pass B). The legacy package stays
registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Watchmode API key via the `api_key` secret; it is sent as the `apiKey` query parameter on
every request (`auth: [{"mode": "api_key_query", "param": "apiKey", ...}]`) and is never logged.
Watchmode's newer docs also accept an `X-API-Key` header or `Authorization: Bearer` — this bundle
keeps the original query-parameter shape for parity with the legacy connector's accepted-input
behavior (a deviation to header auth would be a strictly-different accepted-credential-shape
change, not a superset). `base_url` defaults to `https://api.watchmode.com` and may be overridden
for tests or proxies.

## Streams notes

16 streams total:

- `search` (`GET /v1/search/`, records at `title_results`) — unchanged from the legacy-parity
  migration; sends `search_field=name` and a `search_value` derived from `search_val`
  (default `Terminator`, matching legacy's in-code fallback), plus an optional `start_date`.
- `sources` (`GET /v1/sources/`, records at the response root) — now also forwards an optional
  `regions` filter.
- `regions` (`GET /v1/regions/`, records at the response root) — the ~54 supported country codes;
  primary key `["country"]`.
- `networks` (`GET /v1/networks/`, records at the response root) — TV network reference data.
- `genres` (`GET /v1/genres/`, records at the response root) — movie/TV genre reference data.
- `titles` (`GET /v1/list-titles/`, records at `titles`) — bulk title listing with optional
  `types`/`regions` filters; paginated (`page_number`, `page_param: page`, no size query param
  sent — Watchmode's list-titles endpoint takes only an optional `limit`, wired here via the
  `page_size` config override — default page size 250, Watchmode's own documented per-page
  maximum). A full 250-record page 1 + a short page 2 fixture proves two-page termination.
- `releases` (`GET /v1/releases/`, records at `releases`) — recent/upcoming streaming releases,
  optionally filtered by `regions`.
- `title_details` / `title_sources` / `title_seasons` / `title_episodes` / `title_cast_crew` — all
  5 fan out over the `title_ids` config value (comma/whitespace/semicolon-separated Watchmode
  title IDs; `fan_out.ids_from.config_key`), each issuing one request per configured id against
  `/v1/title/{{ fanout.id }}/{details,sources,seasons,episodes,cast-crew}/` and stamping the
  driving id onto every emitted record as `watchmode_title_id`. `title_details` records at the
  response root (a single object); the other 4 are bare top-level arrays (`records.path: "."`).
  `title_sources`'s primary key is a composite (`watchmode_title_id`, `source_id`, `region`,
  `type`) since one title can appear on the same source under different regions/availability
  types; `title_cast_crew`'s primary key is similarly composite
  (`watchmode_title_id`, `person_id`, `type`, `role`) since one person can appear as both cast and
  crew, or in multiple crew roles.
- `person_details` — fans out over the `person_ids` config value the same way, one request per id
  against `/v1/person/{{ fanout.id }}/`, stamping `watchmode_person_id`.

None of the streams declares an `incremental` block: Watchmode's list/detail endpoints have no
documented server-side "updated since" filter parameter, and the legacy connector's own catalog
never set `CursorFields` on any stream (§8 incremental truth table: no `CursorFields` in legacy's
catalog → no `incremental` block).

## Write actions & risks

None. Watchmode is a read-only media metadata API — `capabilities.write` is `false` and this
bundle ships no `writes.json`. The only endpoint resembling a mutation
(`GET /v1/title/{title_id}/incorrect-data/`) is itself a `GET` request (its "write" is an
out-of-band email Watchmode's team receives) and is excluded as `non_data_endpoint` — see
`api_surface.json`.

## Known limits

- `/v1/changes/new_titles`, `/v1/changes/new_people`, `/v1/changes/titles_sources_changed`,
  `/v1/changes/titles_details_changed`, and `/v1/changes/titles_episodes_changed` are **not**
  implemented: every one of these endpoints' documented response body is a bare JSON array of
  scalar integer IDs (e.g. `{"titles": [123, 456], "page": 1, ...}`), not an array of objects.
  Neither `connsdk.RecordsAt` (which requires each array element to decode as a JSON object, or the
  whole node to itself be a single object) nor `records.keyed_object` (which explodes a JSON
  object's OBJECT-valued entries) can represent "one record per scalar array element" — this is the
  same class of gap documented for `ip2whois`'s `nameservers` field (§5 ledger item 12). Filed as
  `ENGINE_GAP` in `api_surface.json` rather than silently shipping a stream that would emit zero
  records against every real response. These 5 endpoints are also gated to paid Watchmode plans
  (401 on a free-tier key), which is a secondary reason but not the blocking one.
- `/v1/title-release-dates/` (advanced release dates) is excluded as `requires_elevated_scope`:
  Watchmode's docs gate it to paid plans; this connector is provisioned against the free/base API
  tier.
- `/v1/autocomplete-search/` and `/v1/status/` are excluded as `non_data_endpoint` (typeahead helper
  and account-quota snapshot respectively, neither a syncable object stream).
- `title_details`/`title_sources`/`title_seasons`/`title_episodes`/`title_cast_crew` require the
  `title_ids` config value to be set (comma/whitespace/semicolon-separated Watchmode title IDs);
  `person_details` requires `person_ids` similarly. Neither the legacy connector nor Watchmode
  itself offers a way to discover "all title IDs" or "all person IDs" without an unbounded
  brute-force ID scan (Watchmode's `list-titles` returns titles, but chaining its output into
  another stream's fan-out input is a cross-stream dependency the `fan_out` dialect does not
  support — `ids_from.request` reads exactly one preliminary path, not another declared stream) —
  operators seed these ID lists from prior `titles`/`search` sync output or their own known-title
  catalog.
- `titles`/`changes-feed`-shaped streams issue exactly one request per configured page (the
  `page_number` paginator's short-page stop applies); no client-side incremental filtering is
  applied to any stream (see Streams notes).
