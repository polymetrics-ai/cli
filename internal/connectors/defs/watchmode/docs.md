# Overview

Watchmode is a read-only declarative-HTTP connector migrated from `internal/connectors/watchmode`
(legacy wave2 fan-out). It reads title search results and streaming source metadata from the
Watchmode REST API. This bundle is capability-parity with the legacy hand-written connector; the
legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Watchmode API key via the `api_key` secret; it is sent as the `apiKey` query parameter on
every request (`auth: [{"mode": "api_key_query", "param": "apiKey", ...}]`) and is never logged.
`base_url` defaults to `https://api.watchmode.com` and may be overridden for tests or proxies.

## Streams notes

2 streams: `search` (`GET /v1/search/`, records at `title_results`) and `sources` (`GET
/v1/sources/`, records at the response body root, matching legacy's `recordsPath: "."` — the
endpoint returns a bare JSON array of source objects). Primary key is `["id"]` for both; neither
stream is incremental (legacy has no cursor/date-filtered read path for either endpoint beyond the
optional passthrough below).

`search` always sends `search_field=name` and a `search_value` query parameter derived from the
`search_val` config value, defaulting to the literal `Terminator` when unset — matching legacy's
in-code fallback (`queryParams`'s `if value == "" { value = "Terminator" }`), expressed here via
`spec.json`'s query-param `default` dialect.

Both streams additionally forward an optional `start_date` query parameter when the `start_date`
config value is set (`omit_when_absent: true`), matching legacy's unconditional (both streams)
`if start := ...; start != "" { q.Set("start_date", start) }` check that sits outside the
stream-specific branch in `queryParams`.

Both streams use `projection: "passthrough"`: legacy's `Read` emits each decoded record verbatim
via `emit(connectors.Record(item))` with no field-building step (`watchmode.go:110`), so
schema-mode projection would silently drop any Watchmode response field not in the declared list.
The schemas document legacy's own known field surface (legacy's `streamEndpoints` field
declarations, `watchmode.go:39-40`) but do not constrain what is actually emitted.

## Write actions & risks

None. Watchmode is a read-only media metadata API with no mutation endpoints in legacy;
`capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- Only the 2 legacy-parity read streams are implemented; other Watchmode endpoints (title details,
  per-title sources, bulk title listing, network changes) are out of scope for this migration wave
  — see `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass B capability
  expansion"}` entries.
- Neither stream declares pagination: legacy's `Read` issues exactly one request per stream with no
  paging loop, and this bundle mirrors that (no `pagination` block in `streams.json`).
