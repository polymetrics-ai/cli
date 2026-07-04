# Overview

Deputy is a workforce-management product (scheduling, timesheets, HR). This bundle reads
locations, employees, departments, timesheets, tasks, leave requests, rosters, webhooks, and teams,
and writes department/leave/roster/webhook/team create/update/delete mutations, through the Deputy
REST API. Full refresh. This bundle migrates `internal/connectors/deputy` (the hand-written
connector); the legacy package stays registered and unchanged until wave6's registry flip. Pass B
(2026-07-04) expanded this bundle against Deputy's own current documentation
(`developer.deputy.com`, superseding the legacy `www.deputy.com/api-doc/...` docs_url, which now
301-redirects there) — see `api_surface.json` for the full endpoint-by-endpoint accounting of
Deputy's ~1700-page documented surface, the overwhelming majority of which is a single uniform
reflection-generated CRUD matrix repeated across 68 ORM resource types.

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
- `leave` (`/api/v1/resource/Leave`), `rosters` (`/api/v1/resource/Roster`), `webhooks`
  (`/api/v1/resource/Webhook`), and `teams` (`/api/v1/resource/Team`) are Pass B additions on the
  same generic `/resource/{Resource}` collection router `locations`/`departments` already use —
  each inherits the base `offset_limit` pagination (`max`/`start`, `page_size: 500`) unchanged, and
  each stream's `computed_fields` renames the raw PascalCase API fields to this bundle's snake_case
  schema properties exactly like every pre-existing Deputy stream.
- No stream declares an `incremental` block: Deputy is full-refresh only, matching legacy (no
  cursor fields declared for any Deputy stream) — this includes all 4 Pass B streams, none of which
  expose a documented updated-since filter parameter on their generic-resource list endpoint.

## Write actions & risks

Pass B (2026-07-04) added write actions for the 5 generic-resource types with concretely-typed
OpenAPI create/update/delete schemas and no elevated-scope requirement (`capabilities.write:
true`): `create_department`/`update_department`/`delete_department` (`OperationalUnit`),
`create_leave`/`update_leave`/`delete_leave` (`Leave` — `update_leave` is how a leave request's
`Status` approval field is changed), `create_roster`/`update_roster`/`delete_roster` (`Roster`, a
scheduled shift — creating or updating one may trigger a real notification to the assigned
employee), `create_webhook`/`update_webhook`/`delete_webhook` (`Webhook`, a live event-subscription
registration), and `create_team`/`update_team`/`delete_team` (`Team`). All 15 actions use Deputy's
generic-resource router shape: `POST /api/v1/resource/{Resource}` to create, `POST
/api/v1/resource/{Resource}/{{ record.Id }}` to update (Deputy's own API uses POST, not PUT/PATCH,
for updates against this router), `DELETE /api/v1/resource/{Resource}/{{ record.Id }}` to delete
(`delete.missing_ok_status: [404]`, idempotent). Every `record_schema` uses Deputy's own PascalCase
wire field names directly (no `computed_fields`-style rename layer exists on the write path — only
reads project/rename via `computed_fields`), matching every write action's `record_schema` in this
codebase using the API's real wire field names as-is (see stripe's `email`/`name` vs its
`computed_fields`-free writes.json). `Company`(locations)/`Employee`/`Timesheet`/`Task` mutations
are deliberately NOT covered — see Known limits and `api_surface.json`'s per-endpoint reasons.

## Known limits

- `base.pagination.page_size` is set to legacy's real production default/max
  (`deputyDefaultPageSize`/`deputyMaxPageSize`, both `500`) — this is the actual value a live
  deployment's paginator sends; it is not a fixture convenience. `page_size`/`max_pages` are still
  not declared in `spec.json` (dead config, F6, REVIEW.md): the engine's `offset_limit` paginator
  reads its page size only from `streams.json`'s statically-declared `pagination` block, with no
  config-driven override mechanism (the same limitation documented for searxng's `page_size`/
  `max_pages`, `docs/migration/conventions.md`'s Tier-1 read-only variant section).
  `fixtures/streams/locations/{page_1,page_2}.json` uses a full 500-record first page and a short
  second page so conformance exercises the same `max=500` live request shape legacy uses.
- **Company/Employee/Timesheet/Task mutations are excluded, not silently dropped**: `Company`
  (location) create/update/delete only exist via the separately-gated, untyped-body
  `/supervise/company*` endpoints (no generic-resource `/resource/Company` mutation path exists at
  all); `Employee`/`Timesheet` mutations touch HR/payroll-sensitive fields (`Role`,
  `TimeApproved`/`PayRuleApproved`/`Exported`, termination fields) that this pass judged to need
  elevated scope beyond a general-purpose data-sync connector; `Task` creation/update requires a
  pre-existing Task Sheet template (`TaskSetupId`/`GroupId`) this connector's config surface has no
  way to discover. See `api_surface.json`'s per-endpoint `excluded` reasons for the full accounting.
- **The remaining 59 generic-resource types** (`Address`, `Contact`, `CustomField`,
  `EmployeeAgreement`, `PayRules`, `TrainingRecord`, and so on — the full list is in
  `api_surface.json`) share the identical CRUD-matrix shape as the 9 resource types this bundle
  does implement, but were not selected as streams/writes this pass — each is a secondary/internal/
  reference table, a payroll-detail export table, or a plausible-but-deferred future addition; see
  `api_surface.json`'s single consolidated exclusion entry for the specific per-group reasoning
  (not a blanket "Pass B" bucket).
- `docs_url` in `metadata.json` was updated from the legacy `www.deputy.com/api-doc/API/
  Getting_Started` (which now 301-redirects) to the canonical current `developer.deputy.com`.
