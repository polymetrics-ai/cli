# Overview

Acuity Scheduling is a wave2 fan-out declarative-HTTP migration. It reads Acuity appointments,
clients, appointment types, calendars, and forms through the Acuity Scheduling REST API
(`GET https://acuityscheduling.com/api/v1/...`). This bundle targets capability parity with
`internal/connectors/acuity-scheduling` (package `acuityscheduling`, the hand-written connector it
migrates); the legacy package stays registered and unchanged until wave6's registry flip. Acuity is
read-only both in legacy and here.

## Auth setup

Provide the Acuity account User ID as the `username` config value and an Acuity API key as the
`password` secret; both are sent as HTTP Basic auth credentials
(`Authorization: Basic base64(username:password)`), matching legacy's `connsdk.Basic(username,
secret)` (`acuity_scheduling.go:249`). The password is never logged.

## Streams notes

Acuity's list endpoints return a JSON array at the response root (not wrapped in an envelope key),
so every stream declares `records.path: ""` (the empty path selects the raw body root, exactly
`connsdk.RecordsAt(resp.Body, "")`'s legacy call).

Only `appointments` paginates: it honors `max`/`page` query params (`pagination.type:
page_number`, `page_param: page`, `size_param: max`, `start_page: 1`, `page_size: 100` — legacy's
`acuityDefaultPageSize`), advancing until a page returns fewer than 100 records — exactly legacy's
`harvest` stop rule (`len(records) < pageSize || len(records) == 0`, `acuity_scheduling.go:170`).
The other 4 streams (`clients`, `appointment_types`, `calendars`, `forms`) override pagination to
`{"type": "none"}` at the stream level (streams.json's per-stream `pagination` replaces the base
wholesale): legacy's own `harvest` only attaches `max`/`page` when `endpoint.paginated` is true and
otherwise issues exactly one request per read (`acuity_scheduling.go:147`), so these 4 streams never
send a page/max param and always fetch the full list in a single call, matching legacy exactly.

`appointment_types` is this bundle's stream name (snake_case per convention) for legacy's
`appointment-types` resource; the underlying HTTP path (`/appointment-types`, per Acuity's own API)
is unchanged — this is a stream-identifier normalization only, not a wire-request or emitted-data
change.

Field renames: Acuity's raw JSON uses camelCase for several fields legacy flattens to snake_case
(`firstName`→`first_name`, `lastName`→`last_name`, `endTime`→`end_time`,
`appointmentTypeID`→`appointment_type_id`, `calendarID`→`calendar_id`, `amountPaid`→`amount_paid`,
`datetimeCreated`→`datetime_created`). Each is expressed as a bare single-reference
`computed_fields` entry (`{{ record.firstName }}`, etc.), which preserves the raw JSON value's
native type (the engine's typed-extraction rule for bare single references) — `id`,
`appointment_type_id`, and `calendar_id` stay JSON integers; `canceled` stays a JSON boolean;
matching Acuity's real wire shape exactly. `calendars`' `replyTo` field is NOT renamed (legacy's own
`acuityCalendarRecord` keeps the raw `replyTo` key verbatim), so no computed_fields entry is needed
for it — plain schema projection copies it by exact key match.

None of the 5 streams declare an `incremental` block: legacy's list endpoints expose no
server-side incremental cursor and `InitialState` never returns anything but an empty cursor — full
refresh only, matching legacy exactly. `appointments`' schema declares `datetime` as a soft
`x-cursor-field` per legacy's own catalog (`acuityStreams()`), for downstream informational use
only.

## Write actions & risks

None. Acuity Scheduling is read-only both in legacy and here: legacy's own package doc/`Write`
method returns `connectors.ErrUnsupportedOperation` unconditionally ("no approved reverse-ETL
actions"). `capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- Full Acuity API surface (appointment creation/reschedule/cancel, availability, blocks, orders,
  products) is out of scope for wave2; see `api_surface.json`'s `excluded` entries — appointment
  cancellation is additionally flagged `destructive_admin` since it mutates a live scheduled
  appointment.
- Legacy's fixture-mode-only fields (`acuity_scheduling.go`'s `readFixture`, e.g. `connector`,
  `stream`, `previous_cursor` markers) are not modeled — they only ever appeared in legacy's own
  credential-free fixture path, never in a live record; this bundle's schemas target the live
  record shape only, and the engine's own conformance/fixture-replay harness is the credential-free
  test affordance for this bundle.
- `clients` has no stable numeric/string id field in Acuity's own API; legacy and this bundle both
  use `email` as the primary key, which is only unique per-client in practice, not guaranteed
  unique by the API's own contract — matching legacy's `PrimaryKey: []string{"email"}` exactly
  (`streams.go:44-48`), not a new limitation introduced by this migration.
