# Overview

Concord is a contract lifecycle management platform. This bundle reads Concord agreements, user
organizations, folders, reports, and tags through the Concord REST API. It migrates
`internal/connectors/concord` (the hand-written legacy connector), which stays registered and
unchanged until wave6's registry flip. Read-only: Concord's API is full-refresh only upstream,
matching legacy's `Capabilities.Write = false`.

This bundle was UNBLOCKED from `docs/migration/quarantine.json` once the engine's `page_number`
paginator gained an explicit 0-indexed `start_page: 0` (S4 engine mini-wave item 1) — legacy's
`harvest` deliberately starts at page 0 (`for page := 0; ...; page++`), which the pre-increment
paginator could not express (a zero `start_page` was indistinguishable from an unset one).

## Auth setup

Provide a Concord API key via the `api_key` secret; it is sent as the `X-API-KEY` header
(`api_key_header` auth mode) and is never logged, matching legacy's `connsdk.APIKeyHeader`
wiring exactly. `base_url` defaults to the production host
(`https://api.concordnow.com/api/rest/1`); override it for the UAT environment
(`https://uat.concordnow.com/api/rest/1`) or a test/proxy URL.

## Streams notes

All 5 streams share Concord's page-increment pagination (`pagination.type: page_number`,
`page_param: page`, `start_page: 0`, no `size_param` — legacy sends the page size as `limit`,
which this bundle sends via each stream's own templated `query.limit` entry instead of the
paginator's size-param mechanism): the first request sends `page=0`, matching legacy's
`harvest` loop and `concord_test.go`'s pagination assertions exactly; pagination stops on a
short/empty page (fewer than the effective page size), identical to legacy's
`len(records) < pageSize` check.

`limit` is sent as `{{ config.page_size }}` (default `"100"`, matching legacy's
`concordDefaultPage`) via an opt-in optional-query entry with a `default`, since
`PaginationSpec.PageSize`/`SizeParam` are static (non-templated) fields and cannot carry a
config-driven page size directly — the paginator's own static `page_size: 100` governs only the
short-page stop-detection threshold (mirrors legacy's compile-time `concordDefaultPage` for that
purpose), while the actual `limit` query value sent on the wire is fully config-driven, exactly
matching legacy's runtime-overridable page size end to end.

Three streams are **organization-scoped** (`agreements`, `folders`, `reports`): their `path`
templates `{{ config.organization_id }}` directly. `organization_id` is intentionally NOT in
`spec.json`'s `required[]` (it is only required for these 3 of 5 streams, matching legacy's own
conditional `concordOrgID` check that never runs for `user_organizations`/`tags`) — reading an
org-scoped stream without `organization_id` set hard-errors at path-resolution time with an
unresolved-key error naming `organization_id`, the same practical outcome as legacy's dedicated
`errors.New("concord config organization_id is required...")` check, just raised at a different
point in the call stack.

`user_organizations` reads `user/me/organizations` (records at `organizations`); `tags` reads the
flat `tags` endpoint (records at `tags`); `reports` reads `organizations/{id}/reports` (records at
`reports`); `agreements` and `folders` return a bare top-level array (`records.path: ""`) —
matching each stream's `recordsPath` in legacy's `concordStreamEndpoints` routing table exactly.

None of the 5 streams exposes a legacy-recognized incremental cursor field — Concord only supports
full-refresh sync upstream (legacy's own catalog publishes no `CursorFields` for any stream); all
5 are full-refresh only.

`check` issues a single bounded `GET user/me/organizations`, mirroring legacy's `Check`
implementation exactly (listing the authenticated user's organizations confirms auth and
connectivity without mutating anything and without needing an org id).

## Write actions & risks

None. `capabilities.write` is `false`; no `writes.json` is shipped. Legacy itself implements no
write path for Concord (`Write` is a stub returning `ErrUnsupportedOperation`).

## Known limits

- Only the 5 legacy-parity read streams are implemented; see `api_surface.json`. Concord's
  broader documented API surface is out of scope until Pass B.
- Legacy's config-overridable `max_pages` (a per-read hard page-count cap, defaulting to
  unlimited) has no engine-dialect equivalent: `PaginationSpec.MaxPages` is a static integer
  declared in `streams.json`, not a templated/config-driven value (the same limitation documented
  for `page_size`'s pagination-block threshold, above — only the per-request `limit` query value is
  config-driven, not the paginator's own construction-time fields). This bundle declares no
  `max_pages` in `streams.json` (absent = unbounded), matching legacy's own default (unlimited)
  behavior for every caller that never overrides it; `max_pages` is not declared in `spec.json`
  either, since a declared-but-unwireable key is worse than an absent one (F6, matching
  cisco-meraki's identical precedent).
- Legacy's `env` config knob (`"uat"` or `"api"`, selecting the host prefix) has no dialect
  equivalent as a derived default (`base_url`'s `spec.json` default is a fixed literal, not a
  function of another config value — see `docs/migration/conventions.md`'s note on derived
  defaults). This bundle instead declares `base_url` with a fixed default of the production host
  and documents overriding it to the full UAT URL directly; `env` itself is not declared in
  `spec.json` (a declared-but-unwireable key is worse than an absent one, per the F6 dead-config
  rule).
- `fixtures/streams/agreements/page_1.json` ships a FULL 100-record page (matching the static
  `page_size: 100` pagination threshold) plus a 1-record `page_2.json`, satisfying
  `conformance`'s `pagination_terminates` 2-page requirement; the other 4 streams ship a
  single-page fixture each (pagination is already proven by `agreements`).
