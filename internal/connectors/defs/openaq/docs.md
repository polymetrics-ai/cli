# Overview

OpenAQ is a public air-quality reference API (v3). This bundle reads five reference-data
streams (`countries`, `parameters`, `locations`, `instruments`, `manufacturers`) from
`https://api.openaq.org/v3`. It is read-only, migrated from `internal/connectors/openaq` (the
hand-written connector this bundle replaces at parity); the legacy package stays registered and
unchanged until wave6's registry flip.

## Auth setup

Provide an OpenAQ v3 API key via the `api_key` secret; it is sent as the `X-API-Key` header on
every request and is never logged. OpenAQ v3 requires an API key for all requests (unlike openFDA,
there is no anonymous tier), matching legacy's `openaq connector requires secret api_key` Check
failure when unset.

## Streams notes

All 5 streams share OpenAQ v3's common list envelope: `{"meta":{"page","limit","found"},
"results":[...]}` with 1-indexed page/limit pagination (`pagination.type: page_number`,
`page_param: page`, `size_param: limit`). `streams.json`'s `base.pagination.page_size` is set to
`2` purely so the required 2-page conformance fixture (`fixtures/streams/countries/{page_1,
page_2}.json`) can prove real pagination termination without an oversized fixture; legacy's actual
runtime default (`openaqDefaultPageSize`, 100) is not itself expressible as a spec-overridable
value (see the `page_size`/`max_pages` note below) — `2` is a fixture-authoring convenience,
matching the identical pattern in `internal/connectors/defs/aviationstack`'s golden. Pagination
stops on a short page (fewer than `page_size` records) —
legacy's additional `meta.found`-based early stop is a defensive optimization only reachable at
the exact page-size boundary; the engine's short-page stop alone terminates correctly for every
input legacy itself would accept (a final page that exactly fills `page_size` returns zero records
on the next request, which is also a short page). Every stream is full-refresh (OpenAQ reference
data has no updated-at cursor); `countries`/`locations` accept an optional `countries_id` filter
(comma-separated OpenAQ country ids), applied via the `omit_when_absent` optional-query dialect —
absent entirely when unset. Every field in every stream's schema is copied verbatim from the raw
OpenAQ result object (1:1 passthrough, matching legacy's `mapRecord` functions field-for-field);
no renames or computed fields are needed.

## Write actions & risks

None. OpenAQ is a read-only public reference API; `capabilities.write` is `false` and no
`writes.json` is shipped, matching legacy's `ErrUnsupportedOperation` `Write` stub.

## Known limits

- Legacy accepted a `country_ids` alias alongside the canonical `countries_id` query key
  (`openaqCountryFilter` checked both). This bundle declares only `countries_id` (OpenAQ v3's own
  parameter name) — a documented config-surface narrowing, not a data change, since any caller
  using the canonical key sees identical behavior.
- Measurement/sensor time-series endpoints (`/locations/{id}/sensors`, `/sensors/{id}/measurements`)
  and account-scoped endpoints (`/owners`, `/providers`) are out of scope for this wave; see
  `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass B capability expansion"}`
  entries. Only the 5 legacy-parity reference streams are implemented.
- Legacy's runtime-configurable `page_size`/`max_pages` config keys are not declared here: the
  engine's `page_number` paginator reads its page size from `streams.json`'s fixed
  `base.pagination.page_size` and has no per-request config override mechanism, and `MaxPages`
  (the engine's own hard request-count cap) is left unbounded (0) since legacy's own default was
  "all" (unlimited) with no meaningful ceiling. A declared-but-unwireable `page_size`/`max_pages`
  spec property would be dead config (F6), so neither is declared — a scope narrowing on the
  operator-tuning surface, not a data change for any default-configured legacy caller.
