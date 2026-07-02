# Overview

Care Quality Commission (CQC) is a wave2 fan-out migration. This bundle reads the public CQC
Syndication API's core top-level streams — registered locations, registered providers, and
inspection areas — migrating `internal/connectors/care-quality-commission` (the legacy
hand-written connector, which stays registered and unchanged until wave6's registry flip) at
capability parity. The API is read-only; this bundle exposes no write actions.

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

## Write actions & risks

None. This connector is read-only, matching legacy's `Write` stub (`connectors.ErrUnsupportedOperation`).

## Known limits

- `page_size`/`max_pages` runtime overrides are not exposed (see Streams notes above) — every
  read uses the fixed `page_size: 1000`/unbounded-pages shape baked into `streams.json`. Legacy
  accepted `config.page_size` (1-1000) and `config.max_pages` (0/all/unlimited/N) overrides; a
  caller who previously tuned these now gets legacy's own default values unconditionally. This
  never changes any single emitted record's DATA, only how many requests a sync issues and at
  what page size — parity-deviation ledger candidate, ACCEPTABLE under the meta-rule (no
  accepted-input record-data change).
- Per-parent detail/fan-out endpoints (`/locations/{id}`, `/providers/{id}`,
  `/providers/{id}/locations`) and the `/changes/*` delta endpoints are out of scope for wave2;
  see `api_surface.json`'s `excluded: {category: out_of_scope}` entries. Legacy itself never
  implemented these (its own package doc says the substream/detail endpoints are "intentionally
  out of scope for this read-only core connector"), so this is a like-for-like parity boundary,
  not a new narrowing introduced by migration.
- Full CQC API surface (locations_detailed/providers_detailed substreams) is out of scope until
  Pass B.
