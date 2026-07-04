# Overview

When I Work is a Pass B full-surface declarative-HTTP connector for the When I Work REST API v2
(`https://api.wheniwork.com`). It originated as a wave2 fan-out migration of
`internal/connectors/when-i-work` (the hand-written connector it migrates, which reads only 4
streams and writes nothing); the legacy package stays registered and unchanged until wave6's
registry flip. This bundle now reads 14 streams and writes 26 create/update/delete actions across
the practical, non-workflow portion of the real, documented API surface (verified against the live
OpenAPI 3.1 document, 70 paths, at `https://apidocs.wheniwork.com/external/monolith/docs-master.json`).

## Auth setup

Provide a When I Work account `email` and `password` (both secrets); they are sent as HTTP Basic
auth credentials on every request (`auth: [{"mode": "basic", "username": "{{ secrets.email }}",
"password": "{{ secrets.password }}"}]`) and are never logged. `base_url` defaults to
`https://api.wheniwork.com` and may be overridden for tests or proxies.

**This is a known, carried-forward auth-scheme mismatch, not a new deviation introduced this
pass.** The real When I Work API's documented `securitySchemes` (`W-Token`, per the live OpenAPI
document) is a token-EXCHANGE scheme: a developer key (`W-Key` header) plus the account's
email/password are POSTed to a *separate host* (`https://api.login.wheniwork.com/login`,
documented at `https://apidocs.wheniwork.com/external/index.html?repo=login`), which returns a
short-lived JWT; that JWT is then sent as `Authorization: Bearer <token>` (or the `W-Token` header)
on every subsequent `api.wheniwork.com` request. Plain HTTP Basic auth against `api.wheniwork.com`
was verified LIVE to be rejected (`401 {"error":"User login required for this resource.","code":
1000}`) — this bundle's declared Basic-auth candidate does not actually authenticate against the
real production API today, exactly as legacy's own Basic-auth implementation never did either. This
is an `AUTH_COMPLEX` gap (`docs/migration/conventions.md` §6): the real scheme is a genuine
token-exchange flow requiring a preliminary cross-host POST and short-lived-token management,
which is a legitimate `AuthHook` shape (`internal/connectors/engine/hooks.go`'s `AuthHook`
interface) — but this task's instructions forbid creating a new hook package for a connector that
does not already have one (`internal/connectors/hooks/when-i-work/` does not exist), so this pass
carries the existing Basic-auth declaration forward unchanged rather than fabricating a
partially-correct workaround. A caller with a genuinely valid, already-issued JWT can still use this
bundle today by configuring `email`/`password` as a placeholder pair and instead setting `base_url`
to a proxy that injects the real `Authorization: Bearer <token>` header — this is a documented
operational workaround, not a bundle capability.

## Streams notes

14 streams, all `GET`, none paginated (every endpoint's real response is a single unpaginated JSON
object/array — verified per-endpoint against the live OpenAPI document, not just the 4 legacy
streams) and all `"projection": "passthrough"` (§8 rule 1: legacy's `Read` emits every decoded
record verbatim with no field-building, and this bundle's schemas remain a non-exhaustive
documentation surface rather than a filter):

- `users` (`GET /2/users`, records at `users`) — the 4 legacy-parity streams below this line are
  unchanged from the prior migration.
- `locations` (`GET /2/locations`, records at `locations`)
- `positions` (`GET /2/positions`, records at `positions`)
- `shifts` (`GET /2/shifts`, records at `shifts`)
- `sites` (`GET /2/sites`, records at `sites`) — new. Physical sub-locations within a `location`.
- `blocks` (`GET /2/blocks`, records at `blocks`) — new. Reusable shift templates.
- `annotations` (`GET /2/annotations`, records at `annotations`) — new. Schedule-wide notices/
  closures.
- `availabilityevents` (`GET /2/availabilityevents`, records at `availabilityevents`) — new. Per-
  user availability/unavailability windows.
- `requesttypes` (`GET /2/requesttypes`, records at `request-types` — note the real API's hyphenated
  response key) — new. The account's configured time-off request type taxonomy.
- `times` (`GET /2/times`, records at `times`) — new. Worked-time entries.
- `timezones` (`GET /2/timezones`, records at `timezone` — note the real API's SINGULAR response
  key even though it returns an array) — new. Primary key is `timezone_id`, not `id` (the real
  `Timezones` schema has no `id` field at all).
- `payrolls` (`GET /2/payrolls`, records at `payrolls`) — new. Payroll period metadata.
- `openshiftapprovalrequests` (`GET /2/openshiftapprovalrequests`, records at
  `openshiftapprovalrequests`) — new.
- `swaps` (`GET /2/swaps`, records at `swaps`) — new. Shift-swap/drop requests.

Every stream's primary key is `["id"]` except `timezones` (`["timezone_id"]`). No stream declares
`x-cursor-field`/`incremental`: the live OpenAPI document publishes no server-side date-range filter
parameter as a REQUIRED, connector-safe default for any of the 14 (the 2 endpoints that DO expose
`start`/`end` filtering — `requests` and `payrolls`'s sibling `requests` resource — require a
caller-supplied date range with no sensible connector-level default, and are excluded; see Known
limits and `api_surface.json`).

## Write actions & risks

26 write actions across 9 resources — every action is `external mutation; approval required`, and
every `delete_*` action is additionally irreversible (`confirm: "destructive"`):

| Resource | Create | Update | Delete |
|---|---|---|---|
| users | `create_user` | `update_user` | `delete_user` |
| locations | `create_location` | `update_location` | `delete_location` |
| positions | `create_position` | `update_position` | `delete_position` |
| sites | `create_site` | `update_site` | `delete_site` |
| blocks (shift templates) | `create_block` | `update_block` | `delete_block` |
| annotations | `create_annotation` | `update_annotation` | `delete_annotation` |
| availability events | `create_availability_event` | `update_availability_event` | `delete_availability_event` |
| time entries | `create_time` | `update_time` | `delete_time` |
| shifts | `create_shift` | — (see Known limits) | `delete_shift` |

Every action uses `body_type: "json"` (the real API's documented `content-type:
application/json`), `PUT` for update (matching the live OpenAPI document's method, not `PATCH`),
and idempotent-delete semantics (`delete.missing_ok_status: [404]`) since the real API's
`DELETE /2/<resource>/{id}` endpoints are path-scoped single-record deletes. `create_availability_event`
requires `start_time` and `type` (matching the live schema's own `required` list — `type` is an
enum-coded integer, `1` for "unavailable" or `2` for "preferred"). `create_shift`/`create_block`
require `start_time`/`end_time`/`location_id` (matching the live `Shift`/`ShiftTemplate` schema's
own `required` list).

## Known limits

- **Auth-scheme mismatch (AUTH_COMPLEX).** See Auth setup above — this bundle's Basic auth does not
  authenticate against the real production API; carried forward from legacy, not newly introduced.
  Closing this properly requires a token-exchange `AuthHook` (JWT-issuing POST to a separate host,
  matching GitHub's App-JWT hook shape), which this pass does not add since
  `internal/connectors/hooks/when-i-work/` does not already exist.
- **`shift` update (`PUT /2/shifts/{id}`) is not covered.** Excluded pending a dedicated review of
  repeating-shift/shiftchain update propagation semantics (the same `chain` query-parameter concern
  documented on shift delete) — `create_shift`/`delete_shift` are covered; see `api_surface.json`.
- **`requests` (time-off requests) and its full CRUD surface are out of scope.** The real
  `GET /2/requests` endpoint requires caller-supplied `start`/`end` query parameters with no
  connector-level default that would not silently narrow every sync to an arbitrary date window;
  `create`/`update`/`delete` on time-off requests also carry an approval-workflow state machine
  (approve/deny), not plain data-field mutation. See `api_surface.json`.
- **Bulk, notify, publish/unassign/unpublish, punch-clock, and shift-swap-workflow endpoints are
  out of scope.** These are stateful workflow transitions, bulk multi-record operations, or
  notification side effects, not single-record CRUD this dialect's `writes.json` shape expresses
  cleanly. See `api_surface.json` for the full per-endpoint disposition.
- **`requesttypes` create/update (`POST /2/requesttypes`) is not covered.** The real endpoint's body
  is a bulk array wrapper (`{"request-types": [...]}`), not a single-record body shape; the
  `requesttypes` stream remains read-only.
- No stream declares pagination or incremental sync: every one of the 14 endpoints returns its full
  result set in one unpaginated response, verified per-endpoint against the live OpenAPI document.
- `shifts.start_time`/`shifts.end_time` are declared as `string` in the schema (legacy typed them
  `timestamp`, a legacy-only Field type with no draft-07 JSON Schema equivalent; the API returns
  them as RFC3339-shaped strings) — unchanged from the prior migration.
- All 14 streams declare `"projection": "passthrough"` (§8 rule 1) — schemas remain a non-exhaustive
  documentation surface; any additional live-response field still reaches the emitted record instead
  of being silently projected away.
