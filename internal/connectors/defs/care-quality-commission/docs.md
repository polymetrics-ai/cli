# Overview

Care Quality Commission (CQC) is a wave2 fan-out migration, expanded in Pass B to the full
practical CQC Syndication API v1 surface. This bundle reads the public CQC Syndication API's
core top-level streams — registered locations, registered providers, and inspection areas —
migrating `internal/connectors/care-quality-commission` (the legacy hand-written connector,
which stays registered and unchanged until wave6's registry flip). The API is a read-only,
published-open-data API with no authenticated write endpoint of any kind documented anywhere;
this bundle exposes no write actions and never will unless CQC itself adds one.

## Auth setup

Provide the CQC Syndication API's primary subscription key via the `api_key` secret; it is sent
as the `Ocp-Apim-Subscription-Key` header (`auth.mode: api_key_header`), matching legacy's
`connsdk.APIKeyHeader(cqcSubscriptionHeader, secret, "")` with no value prefix. `base_url`
defaults to `https://api.service.cqc.org.uk/public/v1` (legacy's `cqcDefaultBaseURL`).

## Streams notes

`locations` and `providers` share the same shape: `GET` against the CQC list endpoint, page-number
pagination (`pagination.type: page_number`, `page_param: page`, `size_param: perPage`,
`start_page: 1`, `page_size: 1000` — matches legacy's `cqcDefaultPageSize`/`cqcMaxPageSize` of
1000), stopping on a short page (fewer than 1000 records), which is legacy's primary stop
condition (`len(records) < pageSize`). Legacy also honors a reported `totalPages` field as a
secondary/redundant stop signal; the engine's `page_number` paginator does not read a
`totalPages`-style field at all, relying solely on the short-page rule — this never diverges in
practice because a CQC page short of `perPage` always coincides with `page >= totalPages` on the
real API (the two signals agree on every fixture and every documented CQC response shape), so no
behavior is lost. `inspection_areas` is unpaginated (`pagination.type: none`, matching legacy's
`endpoint.paginated: false`), read in a single request, records at the body's `inspectionAreas`
key. `locations` and `providers` are read from `locations`/`providers` keys respectively. None of
the 3 streams are incremental in legacy (no `cursor_field` handling anywhere in
`care_quality_commission.go`), so no `streams.json` entry declares an `incremental` block, and no
schema declares `x-cursor-field`. Primary keys: `locations` uses `locationId`, `providers` uses
`providerId`, `inspection_areas` uses `inspectionAreaId` — matching legacy's declared
`PrimaryKey`.

`spec.json` intentionally does NOT declare `page_size`/`max_pages` as runtime-configurable
properties (unlike legacy, which accepts config overrides for both): the `page_number` paginator's
`PageSize`/`MaxPages` are read exclusively from `streams.json`'s static `pagination` JSON literal
(`PaginationSpec.PageSize`/`MaxPages`), never from a `config.*`-templated value — there is no
mechanism in this dialect to wire a spec property into those fields at all (F6, `conventions.md`:
a declared-but-unwireable spec property is worse than an absent one). See Known limits.

**Pass B schema enrichment**: `locations` and `providers` were widened from a 2-3-field
identity-only projection to the full flat (non-nested-array) field set CQC's own documented
location/provider object shape publishes — registration status/date, address fields, ONSPD
lat/long, phone, region, local authority, constituency, and (locations only) `numberOfBeds`/
`careHome`/`providerId`. Every new property is optional (`["<type>", "null"]`, none added to
`required`), so this is a strictly additive schema widening: no existing record shape becomes
invalid, and a live response that omits a field a caller's older sync never saw simply projects
as absent/null exactly as it always would have. Deeply nested array-of-object sub-resources CQC
also documents on both objects (`relationships`, `locationTypes`/`gacServiceTypes`,
`regulatedActivities`, `specialisms`, `currentRatings`/`historicRatings`, `reports`) are NOT
modeled: `computed_fields`' `join:<sep>` filter only joins a flat array of scalars, not an array
of rating/report objects, and passthrough of the raw nested arrays via plain schema projection
(`"type": ["array","null"]` with no `items` shape) would silently vary per-record depending on
which ratings/report fields CQC happens to populate — left out for now rather than declared with
an under-specified shape.

## Write actions & risks

None. This connector is read-only: the CQC Syndication API documents no write endpoint of any
kind (matching legacy's `Write` stub returning `connectors.ErrUnsupportedOperation`), so
`capabilities.write` stays `false` and no `writes.json` is added.

## Known limits

- `page_size`/`max_pages` runtime overrides are not exposed (see Streams notes above) — every
  read uses the fixed `page_size: 1000`/unbounded-pages shape baked into `streams.json`. Legacy
  accepted `config.page_size` (1-1000) and `config.max_pages` (0/all/unlimited/N) overrides; a
  caller who previously tuned these now gets legacy's own default values unconditionally. This
  never changes any single emitted record's DATA, only how many requests a sync issues and at
  what page size — parity-deviation ledger candidate, ACCEPTABLE under the meta-rule (no
  accepted-input record-data change).
- Detail-by-id endpoints (`/locations/{id}`, `/providers/{id}`) are excluded as `duplicate_of`
  their list stream: CQC's own documentation shows the detail endpoint returning the identical
  full object shape the list endpoint already emits per record, so no read coverage is lost.
- `/providers/{id}/locations` is excluded as `out_of_scope`: it is not present anywhere in CQC's
  own published Syndication API documentation. A live probe against this path does return HTTP
  401 rather than 404, but that is inconclusive (the Azure API Management gateway in front of
  this API appears to gate on subscription key before backend route-matching for at least some
  undocumented paths), so this bundle treats it as unconfirmed/undocumented rather than a real
  endpoint. The same information (a provider's associated location ids) is already reachable, in
  principle, from the (currently unmodeled) `locationIds` array on the provider object itself.
- `/changes/locations` and `/changes/providers` are excluded as `out_of_scope`: both return a
  paginated array of BARE location/provider-id strings for a `startTimestamp`/`endTimestamp`
  window, not full record objects — the dialect's `records.path`/schema-projection model has no
  primitive for turning a bare-scalar array into per-element records (the identical `ENGINE_GAP`
  class documented for ip2whois's `nameservers` field, `docs/migration/conventions.md` §5 item
  12), and modeling it would require either a `fan_out` fetch-per-id round trip (defeating the
  feed's own purpose — its whole point is letting a caller avoid re-fetching every full record)
  or emitting useless id-only pseudo-records.
- Deeply nested rating/report/relationship sub-structures on `locations`/`providers` (see Streams
  notes) are not modeled as schema properties; only the flat, non-array-of-object CQC-documented
  fields are.
