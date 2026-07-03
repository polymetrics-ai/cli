# Overview

Deputy is a workforce-management product (scheduling, timesheets, HR). This bundle reads
locations, employees, departments, timesheets, and tasks through the Deputy REST API. Read-only,
full refresh. This bundle migrates `internal/connectors/deputy` (the hand-written connector); the
legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Deputy bearer access token via the `api_key` secret; it is sent as the `Authorization`
header (`auth: [{"mode": "bearer", "token": "{{ secrets.api_key }}"}]`), matching legacy's
`connsdk.Bearer(secret)`. `base_url` is REQUIRED with no default — Deputy is install-specific
(`https://{installname}.{geo}.deputy.com`), matching legacy's own `deputyBaseURL`, which has no
in-code fallback.

## Streams notes

Deputy's raw resource objects use PascalCase field names (`Id`, `CompanyName`, `DisplayName`, ...);
every stream's `computed_fields` renames each field to this bundle's snake_case schema property
(`id`, `company_name`, `display_name`, ...) via a bare `{{ record.<PascalField> }}` reference —
schema projection alone only matches by exact key name, so without the rename every field would
silently drop. A bare single-reference `computed_fields` entry (no filter, no literal text)
performs typed extraction: the raw JSON value's native type (integer/boolean/string) is preserved,
not stringified, matching legacy's `deputy*Record` functions which likewise copy the raw
`map[string]any` values through unchanged.

- `locations` (`/api/v1/resource/Company`) and `departments` (`/api/v1/resource/OperationalUnit`)
  are Deputy's paginated `/resource/*` collection endpoints: `pagination.type: offset_limit` with
  `limit_param: max`, `offset_param: start` (Deputy's own query parameter names), matching legacy's
  `?start=N&max=N` convention; a page shorter than `page_size` is the last page.
- `employees` (`/api/v1/supervise/employee`), `timesheets` (`/api/v1/my/timesheets`), and `tasks`
  (`/api/v1/my/tasks`) are Deputy's curated, non-paginated endpoints; each stream overrides the
  base pagination with `pagination: {"type": "none"}`, matching legacy's `endpoint.paginated:
  false` branch (a single bounded request, no `start`/`max` query params sent at all).
- No stream declares an `incremental` block: Deputy is full-refresh only, matching legacy (no
  cursor fields declared for any Deputy stream).

## Write actions & risks

None. Deputy is read-only here (`capabilities.write: false`); legacy's `Write` unconditionally
returns `ErrUnsupportedOperation`.

## Known limits

- `base.pagination.page_size` is set to legacy's real production default/max
  (`deputyDefaultPageSize`/`deputyMaxPageSize`, both `500`) — this is the actual value a live
  deployment's paginator sends; it is not a fixture convenience. `page_size`/`max_pages` are still
  not declared in `spec.json` (dead config, F6, REVIEW.md): the engine's `offset_limit` paginator
  reads its page size only from `streams.json`'s statically-declared `pagination` block, with no
  config-driven override mechanism (the same limitation documented for searxng's `page_size`/
  `max_pages`, `docs/migration/conventions.md`'s Tier-1 read-only variant section). The `locations`
  stream declares a stream-level `pagination` override (`page_size: 5`) so its required 2-page
  conformance fixture (`fixtures/streams/locations/{page_1,page_2}.json`, §4 of
  `docs/migration/conventions.md`) can stay small and readable; since stream-level `pagination`
  replaces the base spec wholesale, this is an intentional, ledgered per-stream deviation from
  legacy's uniform 500-record page size — `locations` reads in smaller, more numerous pages than
  legacy would, everywhere else identical. `departments` (the other paginated stream) is
  unaffected and uses legacy's true 500-record page size end-to-end, matching its single-page
  fixture's `max=500` request/response. `metadata.json.batch.read_page_size` documents Deputy's
  real 500 default/max for operator awareness.
- Full Deputy API surface (rosters, leave, journals, resource create/update/delete) is out of
  scope for this wave; see `api_surface.json`'s `excluded` entries.
