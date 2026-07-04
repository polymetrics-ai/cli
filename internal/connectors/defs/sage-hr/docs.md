# Overview

Sage HR is a Pass B full-surface-expansion declarative-HTTP migration. It reads employees, teams,
time off (requests, policies, KIT days, out-of-office, individual allowance reports), terminated
employees, termination reasons, positions, recruitment positions and per-position applicants, and
onboarding/offboarding/document task categories, and writes employee/leave/task lifecycle
mutations, through the Sage HR API. This bundle targets capability parity with
`internal/connectors/sage-hr` (the hand-written connector it migrates, package `sagehr`) as a
**superset**: legacy's original 3 read streams are preserved and corrected (see "Envelope shape
correction" below), and 11 new read streams plus 10 write actions are added against the full
documented Sage HR v1.0 OpenAPI surface. The legacy package stays registered and unchanged until
wave6's registry flip.

Real API docs (`developers.sage.hr`, `apidoc.sage.hr`, `sagehr.docs.apiary.io`) all return a
Cloudflare bot-challenge 403/502 to automated fetches (both `curl` and the migration harness's
WebFetch tool). The full Sage HR v1.0 OpenAPI spec (47 documented paths) was instead obtained from
its GitHub mirror (`https://raw.githubusercontent.com/api-evangelist/sage-hr/main/openapi/sage-hr-openapi.yml`,
fetched 2026-07-03) — every stream/write/schema/field below is derived from that spec's real
request/response examples, not guessed.

## Auth setup

Provide a Sage HR API key via the `api_key` secret; it is sent as the `X-Auth-Token` request header
(`api_key_header` auth mode), never logged, matching legacy's
`connsdk.APIKeyHeader("X-Auth-Token", token, "")` (`sage_hr.go:143`). `base_url` defaults to
`https://api.sage.hr/v1`, matching legacy's own in-code default; the real OpenAPI spec's server
block addresses tenants at `https://<subdomain>.sage.hr/api` (a per-tenant subdomain, not a fixed
host) — `base_url` is the mechanism for pointing at a real tenant, tests, or a proxy.

## Streams notes

**Envelope shape correction (from legacy's untested assumption to the real, documented wire
shape):** legacy's `employees`/`teams`/`timeoff_requests` streams assumed a bare top-level JSON
array with `records.path: ""`, based only on `sage_hr_test.go`'s own fixture (which never actually
exercised the real API). The real Sage HR OpenAPI spec's response examples for every list endpoint
(`/employees`, `/teams`, `/leave-management/requests`, and every other list stream below) show a
consistent `{"data": [...], "meta": {"current_page", "next_page", "previous_page", "total_pages",
"per_page", "total_entries"}}` envelope with page-number pagination (`?page=N`) — this bundle
corrects `records.path` to `"data"` and declares `page_number` pagination (`page_param: "page"`,
`size_param: ""` since the real API's list endpoints do not accept a page-size override param) for
every list stream, superseding the untested `""`/no-pagination assumption. `timeoff_requests`'s
real path is also corrected from legacy's invented `/timeoff/requests` (which does not exist in the
documented API) to the real `/leave-management/requests`.

**All streams declare `"projection": "passthrough"`** (conventions.md §8 rule 1): legacy's `Read`
performs zero field mapping (`recordsAtAny` + a direct `connectors.Record(rec)` cast, no
renaming/filtering), so schema-mode projection would silently drop real API fields; every stream
here follows the same passthrough precedent for parity and honesty about actual emitted shape.

New streams added this pass (all `GET`, `page_number` pagination where the endpoint supports it,
`records.path: "data"`):

- `terminated_employees` (`/terminated-employees`), `positions` (`/positions`),
  `termination_reasons` (`/termination-reasons`) — simple paginated lists.
- `leave_policies` (`/leave-management/policies`), `out_of_office_today`
  (`/leave-management/out-of-office-today`) — unpaginated lists (the real API's response examples
  for these two endpoints carry no `meta` pagination block at all).
- `individual_allowances` (`/leave-management/reports/individual-allowances`) — paginated,
  `size_param: "per_page"` (this endpoint's real query parameters include `per_page` alongside
  `page`, unlike the other list endpoints).
- `recruitment_positions` (`/recruitment/positions`) — paginated, `size_param: "per_page"`.
- `recruitment_applicants` (`/recruitment/positions/{id}/applicants`) — a **`fan_out` stream**
  (conventions.md §3): there is no top-level endpoint to list applicants across all positions, so
  this stream fans out over every id returned by `recruitment_positions`'s own list request
  (`fan_out.ids_from.request`), issuing one paginated applicants sub-request per position and
  stamping `position_id` onto every emitted applicant record.
- `onboarding_categories`, `offboarding_categories`, `document_categories` — small unpaginated
  category lists.

None of the 14 streams declare an `incremental` block, matching legacy's `Catalog` (no
`CursorFields`) and the real API's documented query parameters (no updated-since filter is
published for any of these list endpoints; `leave-management/requests`'s `from`/`to` params are a
date-range window, not an incremental cursor, and are not wired here to avoid silently narrowing a
default full-history sync).

## Write actions & risks

`capabilities.write` is now `true` (10 actions added; legacy shipped none):

- `create_employee` (`POST /employees`, form body) — creates a new employee; may trigger a welcome
  email (`send_email`). Approval required.
- `update_employee` (`PUT /employees/{id}`, JSON body) — updates org placement (team/position/
  location), reporting line (`leader_id`), and leave-type eligibility. Approval required.
- `update_employee_custom_field` (`PUT /employees/{id}/custom-fields/{custom_field_id}`, form body)
  — updates one custom-field value for one employee. Approval required.
- `terminate_employee` (`POST /employees/{id}/terminations`, form body) — **destructive/
  irreversible**: ends an employee's active record in Sage HR. Approval required.
- `create_timeoff_request` (`POST /leave-management/requests`, form body) — creates a new time off
  request against an employee's leave balance. Approval required.
- `create_kit_day` (`POST /leave-management/kit-days`, form body) — creates a Keeping-In-Touch day
  entry. Approval required.
- `update_kit_day_status` (`PATCH /leave-management/kit-days/{id}`, form body) — approves, declines,
  or cancels a KIT day. Approval required.
- `update_leave_policy_kit_days` (`PATCH /leave-management/policies/{id}`, form body) — changes a
  company-wide leave policy's KIT-day configuration. Approval required.
- `create_onboarding_task` / `create_offboarding_task` (`POST /onboarding/tasks` /
  `POST /offboarding/tasks`, form body) — creates a task template for the employee
  onboarding/offboarding lifecycle. Approval required.

## Known limits

- **Multipart/file-upload endpoints are excluded (dialect limitation, not an oversight).**
  `POST /documents` (employee/company document upload) and
  `POST /recruitment/positions/{id}/applicants` (applicant creation with an optional resume
  attachment) both require a `multipart/form-data` body; the engine's write dialect's `body_type`
  (`json`/`form`/`none` over a fixed field set) has no multipart/binary support. See
  `api_surface.json`'s `binary_payload` entries.
- **`POST /timesheets/clock-in` is excluded.** Its request body is a dynamic date-string-keyed
  nested object (`clocked_time: {"YYYY/MM/DD": {employee_id: [{clock_in, clock_out}, ...]}}`) — an
  arbitrary caller-supplied key structure the dialect's fixed-field body construction cannot
  express.
- **The entire Vikarina payroll-bridge integration surface (10 `POST /vikarina/*` endpoints) is
  excluded as out of scope.** These transfer Sage HR data INTO a third-party payroll product
  (Vikarina); they are not Sage HR data resources in their own right.
- **`GET /leave-management/kit-days` is excluded.** It requires BOTH a `policy_id` and an
  `employee_id` as required query filters, with no discovery/listing endpoint for either dimension
  independent of the other — not a syncable top-level collection without externally-supplied ids
  already in hand.
- **Several per-employee/per-applicant sub-resources are excluded as Pass B breadth-vs-cost
  triage**: `GET /employees/{id}/compensations`, `GET /employees/{id}/custom-fields`,
  `GET /employees/{id}/leave-management/balances`, and `GET /recruitment/applicants/{id}/actions`
  would each require a fan_out read issuing one request per employee/applicant id for a
  low-cardinality, rarely-changing field set; not implemented this pass.
- **Performance goal-progress endpoints (4 `/performance/goals/quarterly-progress/*` paths) are
  excluded as `non_data_endpoint`.** Each returns a single org-wide aggregate/rollup snapshot with
  no per-record identity, not a syncable object stream.
- **No incremental filtering is modeled for any stream**, matching legacy's `Catalog` (no declared
  `CursorFields`) and the real API's lack of a documented updated-since filter parameter on any of
  these list endpoints.
