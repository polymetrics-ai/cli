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

- `pagination.page_size` is declared as `5` in this bundle rather than legacy's real default/max
  (`deputyDefaultPageSize`/`deputyMaxPageSize`, both 500): the engine's `offset_limit` paginator
  reads its page size only from `streams.json`'s statically-declared `pagination` block, with no
  config-driven override mechanism (the same limitation documented for searxng's `page_size`/
  `max_pages`, `docs/migration/conventions.md`'s Tier-1 read-only variant section) — `page_size`/
  `max_pages` are consequently not declared in `spec.json` at all (dead config, F6, REVIEW.md). A
  smaller page size only changes how many HTTP round-trips a full sync makes (more, smaller pages
  for the same total dataset against a real Deputy account with more than 5 locations/departments),
  never which records are emitted — `metadata.json.batch.read_page_size` still documents Deputy's
  real 500 default/max for operator awareness, matching the same informational-vs-enforced split
  used for `metadata.json.rate_limit` elsewhere in this codebase.
- Full Deputy API surface (rosters, leave, journals, resource create/update/delete) is out of
  scope for this wave; see `api_surface.json`'s `excluded` entries.
