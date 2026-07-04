# Overview

Acuity Scheduling reads Acuity appointments, clients, appointment types, calendars, forms,
products, orders, and labels, and writes appointment/block/certificate mutations, through the
Acuity Scheduling REST API (`https://acuityscheduling.com/api/v1/...`). This bundle originally
targeted capability parity with `internal/connectors/acuity-scheduling` (package
`acuityscheduling`, the hand-written connector it migrates; the legacy package stays registered and
unchanged until wave6's registry flip; Acuity was read-only there) and was expanded in Pass B to
cover the full documented REST surface (`api_surface.json`).

## Auth setup

Provide the Acuity account User ID as the `username` config value and an Acuity API key as the
`password` secret; both are sent as HTTP Basic auth credentials
(`Authorization: Basic base64(username:password)`), matching legacy's `connsdk.Basic(username,
secret)` (`acuity_scheduling.go:249`). The password is never logged.

## Streams notes

Acuity's list endpoints return a JSON array at the response root (not wrapped in an envelope key),
so every stream declares `records.path: ""` (the empty path selects the raw body root), including
all 3 new Pass B streams (`products`, `orders`, `labels`) — confirmed from each resource's own
reference page.

Only `appointments` paginates with Acuity's own page/max convention: it honors `max`/`page` query
params (`pagination.type: page_number`, `page_param: page`, `size_param: max`, `start_page: 1`,
`page_size: 100` — legacy's own `acuityDefaultPageSize`), advancing until a page returns fewer than
100 records. `clients`/`appointment_types`/`calendars`/`forms`/`products`/`labels` override
pagination to `{"type": "none"}` at the stream level and issue exactly one request per read,
matching each resource's own documented lack of a `page` parameter. `orders` is a distinct third
shape: Acuity's own `/orders` reference page documents a single `max` query parameter (default
100, described as "maximum number of results") with **no** corresponding `page` parameter at all —
there is no way to page past the first `max` results — so `orders` also declares
`pagination: {"type": "none"}` but sends a static `query: {"max": "100"}` to request the maximum
single-page size Acuity's own default already uses.

`appointment_types` is this bundle's stream name (snake_case per convention) for legacy's
`appointment-types` resource; the underlying HTTP path (`/appointment-types`) is unchanged.

Field renames: several of Acuity's raw JSON payloads use camelCase fields this bundle normalizes to
snake_case via bare single-reference `computed_fields` entries (preserving each field's native JSON
type per the engine's typed-extraction rule): `appointments` renames `firstName`/`lastName`/
`endTime`/`appointmentTypeID`/`calendarID`/`amountPaid`/`datetimeCreated`; `clients` renames
`firstName`/`lastName`; the new `orders` stream renames `firstName`/`lastName` the same way (Acuity's
own `/orders` response uses the identical camelCase convention as `/clients`). `calendars`'
`replyTo` and `products`'/`labels`' fields are NOT renamed (each is already the schema's target
field name verbatim), so no computed_fields entry is needed for them.

None of the 8 streams declare an `incremental` block: none of Acuity's documented list endpoints
(old or newly added) expose a server-side incremental/updated-since filter parameter — full refresh
only, across the whole bundle. `appointments`' schema declares `datetime` as a soft
`x-cursor-field` per legacy's own catalog, for downstream informational use only; the 3 new streams
declare no cursor field since none of their documented record shapes carry a timestamp field at all
(`products`/`labels` are static catalog/configuration lists; `orders`' `time` field is a
human-formatted display string, not the resource's own creation timestamp, so it is not declared as
a cursor).

## Write actions & risks

Pass B adds 5 write actions, all newly modeled (legacy shipped none — "no approved reverse-ETL
actions", `acuity_scheduling.go`'s package doc):

- `create_appointment` (`POST /appointments`) — books a live appointment; depending on account
  settings, triggers a client confirmation email/SMS. Approval required.
- `update_appointment` (`PUT /appointments/{id}`) — updates a live appointment from Acuity's own
  white-list of updatable client-facing attributes. Approval required.
- `cancel_appointment` (`PUT /appointments/{id}/cancel`) — **note the real HTTP method is `PUT`, not
  `DELETE`** (the wave2 `api_surface.json` entry incorrectly recorded `DELETE`; corrected here per
  Acuity's own `put-appointments-id-cancel` reference page). Permanently cancels a live scheduled
  appointment — Acuity's own docs state it is not possible to un-cancel — and by default sends the
  client a cancellation notification. Approval required.
- `create_block` (`POST /blocks`) — blocks off a calendar time range, preventing client bookings in
  it. Approval required.
- `create_certificate` (`POST /certificates`) — issues a live, redeemable package or coupon
  certificate code. Approval required.

`capabilities.write` is now `true`.

## Known limits

- `reschedule` (`PUT /appointments/{id}/reschedule`) is a distinct lifecycle action from
  `update_appointment` (date/time-only, its own validation path) and is deferred out of scope
  pending real demand beyond the covered create/update/cancel set.
- `blocks` has no documented `GET /blocks` list endpoint (only `GET /blocks/{id}` by id and
  `POST /blocks` to create) — there is no catalog list to build a read stream from without a
  separate id source, so only `create_block` is modeled; no `blocks` read stream exists.
  `DELETE /blocks/{id}` is excluded as `destructive_admin` (re-opens a blocked time range to
  bookings).
- `availability/times`, `availability/dates`, and `availability/classes` are excluded as
  `out_of_scope`: each is a computed scheduling-UI query (available slots for a given
  appointment-type/date), not a syncable catalog object with a stable id.
- Per-appointment payment-transaction history (`GET /appointments/{id}/payments`) is out of scope:
  a per-parent sub-resource with no top-level list endpoint, which would require per-appointment
  `fan_out` this pass does not scope.
- `POST /calendars` (creating a new staff calendar/seat) is excluded as `requires_elevated_scope`: a
  billable-seat administrative action beyond ordinary API-key scope.
- Legacy's fixture-mode-only fields (`acuity_scheduling.go`'s `readFixture`, e.g. `connector`,
  `stream`, `previous_cursor` markers) are not modeled — they only ever appeared in legacy's own
  credential-free fixture path, never in a live record; this bundle's schemas target the live
  record shape only, and the engine's own conformance/fixture-replay harness is the credential-free
  test affordance for this bundle.
- `clients` has no stable numeric/string id field in Acuity's own API; legacy and this bundle both
  use `email` as the primary key, which is only unique per-client in practice, not guaranteed
  unique by the API's own contract — matching legacy's `PrimaryKey: []string{"email"}` exactly, not
  a new limitation introduced by this migration. `orders` uses its own numeric `id` field, which
  Acuity's `/orders` response does document as a stable per-order identifier.
