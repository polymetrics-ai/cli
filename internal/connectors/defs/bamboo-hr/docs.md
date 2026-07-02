# Overview

BambooHR is a read-only HR data source, migrated from the hand-written `internal/connectors/bamboo-hr`
package to this declarative bundle. It reads the employee directory and three HR metadata
endpoints (field definitions, list-field options, time off types) through the BambooHR REST API v1
(`https://<subdomain>.bamboohr.com/api/v1`).

## Auth setup

Provide your BambooHR account `subdomain` (the `<subdomain>` in `https://<subdomain>.bamboohr.com`)
and an `api_key` secret (Account settings > API Keys). The API key is sent as the HTTP Basic
username with a literal `x` password — BambooHR's documented API-key convention — and is never
logged. There is no `base_url` override in this bundle: `subdomain` is required and directly
templates the base URL (`https://{{ config.subdomain }}.bamboohr.com/api/v1`), matching legacy's
`bambooBaseURL` derivation.

## Streams notes

- `employees` reads `employees/directory`, records at `employees`, paginated with `page`/`limit`
  query params (`pagination.type: page_number`) and BambooHR's own short-page stop rule (a page
  returning fewer than `page_size` records is the last page) — the same rule legacy's `harvest`
  loop implements by hand. Primary key `id`.
- `meta_fields` reads `meta/fields` (a top-level JSON array, `records.path: ""`), single-page
  (`pagination.type: none`, matching legacy's non-paginated flat-endpoint branch). Primary key `id`.
- `meta_lists` reads `meta/lists` (also a top-level array), single-page. Primary key `field_id`,
  sourced from the raw API's `fieldId` via a `computed_fields` rename (schema projection is exact
  key match only).
- `time_off_types` reads `meta/time_off/types`, records at the nested `timeOffTypes` key,
  single-page. Primary key `id`.
- None of the 4 streams is incremental: legacy declares no `CursorFields` and BambooHR's own API
  offers no server-side updated-since filter for these endpoints, so every sync is a full read
  (matching legacy's `InitialState`, which always starts empty and is never advanced by a request
  param for this connector).
- **Id stringification parity**: legacy's four record mappers all route their primary-key field
  through a defensive `stringField` helper that coerces any JSON type (string OR number) to a Go
  string — `meta/fields`' real wire shape sends `id` as a bare JSON number (`"id": 1`), while
  `employees`/`time_off_types`/`meta_lists` send it as a string already, but legacy emits a STRING
  in every case. Each stream's `computed_fields` entry uses `{{ record.<id-field> | last_path_segment }}`
  rather than a bare `{{ record.<id-field> }}` reference: the engine's typed-extraction rule
  (bare-reference-only) would otherwise copy `meta_fields`' raw numeric `1` through as a JSON
  number, diverging from legacy's always-string output. `last_path_segment` is the documented
  identity filter for any value containing no `/` (every id here), so the string VALUE is
  unchanged while forcing the stringify path — deliberate reuse of a filter for its documented
  no-op-on-slash-free-input behavior, not its primary URI-segment purpose.

## Write actions & risks

None. BambooHR is read-only in this bundle (`capabilities.write: false`), matching legacy exactly
— `bamboohr.go`'s `Write` is an unconditional `ErrUnsupportedOperation` stub.

## Known limits

- Full BambooHR API surface (single-employee read/update, custom reports, time off requests,
  files, benefits, training, webhooks) is out of scope for this wave; see `api_surface.json`'s
  `excluded` entries. Only the 4 legacy-parity read streams are implemented.
- **`page_size`/`max_pages` config overrides are not modeled.** Legacy accepts optional `page_size`
  (1-1000, default 100) and `max_pages` (default unlimited) config keys read at request time
  (`bambooPageSize`/`bambooMaxPages`, `bamboohr.go:305-333`). The engine's `PaginationSpec.PageSize`/
  `MaxPages` fields are plain fixed JSON integers baked into `streams.json`'s `base.pagination`
  block — there is no templating/config-driven override mechanism for either. This bundle declares
  a fixed `page_size: 2` (chosen small so the required 2-page conformance fixture is realistic and
  exercises the short-page stop rule honestly; legacy's own default is 100) and no `max_pages` cap
  (unbounded, matching legacy's own default). Neither `page_size` nor `max_pages` is declared in
  `spec.json` (a declared-but-unwireable key is worse than an absent one).
- `employees/directory`'s `fields` envelope key (the account's configured custom-field list) is not
  modeled as a stream — it is directory metadata about the request, not a syncable object
  collection, matching legacy's own scope (legacy never surfaces it as a stream either).
