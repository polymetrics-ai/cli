# Overview

Cloudbeds reads guest, hotel, room, reservation, transaction, rate, inventory, and operational
(housekeeping/webhook/notes) data, and writes guest/reservation/payment/folio/housekeeping/
house-account/webhook/room-block actions, through the Cloudbeds v1.2 REST API
(`https://api.cloudbeds.com/api/v1.2/...`). This bundle originally migrated
`internal/connectors/cloudbeds` (the hand-written connector) to capability parity (5 read-only
streams), then (Pass B) expanded to the full documented v1.2 surface: 21 streams (up from 5) and
32 write actions (up from none), researched against Cloudbeds' official OpenAPI spec
(`github.com/cloudbeds/openapi-specs`, `src/pms-v1.2-openapi.yaml`, 115 endpoints) plus the
`developers.cloudbeds.com` reference pages for response field lists the spec itself leaves
undocumented. The legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Cloudbeds API access token via the `api_key` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_key>`) and is never logged, matching legacy's
`connsdk.Bearer(secret)` (`cloudbeds.go:238`). `base_url` defaults to
`https://api.cloudbeds.com/api/v1.2` and may be overridden for tests/proxies. Two new OPTIONAL
config keys support specific new streams only (never referenced by any other stream):
`property_id` (required by `groups`/`allotment_blocks` specifically — Cloudbeds' own API requires
`propertyID` for `GET /getGroups`/`GET /getAllotmentBlocks`) and `rate_plans_start_date`/
`rate_plans_end_date` (required by `rate_plans` specifically — Cloudbeds requires `startDate`/
`endDate` for `GET /getRatePlans`).

## Streams notes

The original 5 legacy-parity streams (`guests`, `hotels`, `rooms`, `reservations`, `transactions`)
are unchanged — see the parity notes below.

Pass B additions, all sharing the base `page_number` pagination (`pageNumber`/`pageSize`, 1-based,
page size 100) unless noted `pagination: none` (the endpoint itself has no `pageNumber`/`pageSize`
query params, confirmed against the OpenAPI spec):

- `rate_plans` (`GET /getRatePlans`, `pagination: none`) — requires `config.rate_plans_start_date`/
  `rate_plans_end_date` (Cloudbeds itself requires `startDate`/`endDate`); no stable single-field
  primary key exists upstream, so `ratePlanID` is used (matches the per-rate-plan record identity).
- `room_types` (`GET /getRoomTypes`) — inherits base pagination (the endpoint documents
  `pageNumber`/`pageSize`).
- `items` / `item_categories` / `taxes_and_fees` / `sources` / `house_accounts` /
  `custom_fields` / `payment_methods` / `webhooks` (all `pagination: none` — none of these
  endpoints document `pageNumber`/`pageSize` params).
- `groups` (`GET /getGroups?propertyID=...`) / `allotment_blocks`
  (`GET /getAllotmentBlocks?propertyID=...`) — both require `config.property_id` and inherit base
  pagination.
- `housekeepers` / `housekeeping_status` — inherit base pagination (both endpoints document
  `pageNumber`/`pageSize`).
- `guest_notes` (`GET /getGuestNotes?guestID=...`, `pagination: none`) — Cloudbeds has no
  "all guest notes" endpoint; notes are always scoped to one guest. Uses the `fan_out` dialect
  (conventions.md §3) over this bundle's own `guests` stream (`ids_from.request` re-reads
  `GET /getGuestList`, `records_path: data`, `id_field: guestID`); `into.query_param: guestID`,
  `stamp_field: guestID`.
- `reservation_notes` (`GET /getReservationNotes?reservationID=...`, `pagination: none` for the
  per-reservation request; the fan-out id-listing request over this bundle's own `reservations`
  stream ALSO uses `reservation_notes`'s own `pagination: none` override, since the fan_out
  dialect's `ids_from.request` reuses the SURROUNDING stream's effective pagination spec, not the
  source stream's — see conventions.md §3) — `into.query_param: reservationID`,
  `stamp_field: reservationID`.

`taxes_and_fees`'s primary key is a composite `["type", "feeID", "taxID"]`: Cloudbeds' own schema
documents `feeID` as populated only when `type == "fee"` and `taxID` only when `type == "tax"` —
there is no single field always populated regardless of type. `custom_fields`'s primary key is
`["propertyID", "shortcode"]` (no numeric id is ever returned).

### Original 5 legacy-parity streams (unchanged)

All five streams (`guests` → `/getGuestList`, `hotels` → `/getHotels`, `rooms` →
`/getRoomBlocks`, `reservations` → `/getReservations`, `transactions` → `/getTransactions`) are
simple list endpoints; records live at the top-level `data` key alongside `success`/`count`
fields legacy never reads. Pagination is Cloudbeds' page-increment convention
(`pagination.type: page_number`, `page_param: pageNumber`, `size_param: pageSize`, `page_size: 100`
matching legacy's `cloudbedsDefaultPageSize`) — Cloudbeds sends BOTH `pageNumber` AND `pageSize` on
every request (legacy's `harvest`, `cloudbeds.go:144-146`). A page returning fewer than `page_size`
records is the last page, matching `cloudbeds.go:164-169`'s exact stop rule.

Cloudbeds is full-refresh only upstream — legacy's own `cloudbedsStreams` comment: "Cloudbeds is
full-refresh only upstream, so no cursor fields are published" — no stream in this bundle (legacy
or Pass B) declares an `incremental` block. Primary keys follow each entity's own Cloudbeds id
field rather than a uniform `id`, matching legacy's per-stream `PrimaryKey` declarations exactly
(`streams.go:30-63`) for the original 5; new streams follow the identical convention.

## Write actions & risks

Pass B capability expansion — legacy shipped no writes at all (`capabilities.write` flips `false`
→ `true` in this bundle). All 32 write actions use `body_type: form` (Cloudbeds' own v1.2 API is
`application/x-www-form-urlencoded` for every mutation, confirmed against the OpenAPI spec) except
the 4 query-parameterized DELETE actions (`delete_guest_note`, `delete_reservation_note`,
`delete_adjustment`, `delete_webhook`), which template their identifying fields directly into the
path's query string (`body_type: none`, no body at all — matching Cloudbeds' own `DELETE` shape,
which takes no request body for these 4 endpoints).

- **Guest**: `create_guest`/`update_guest`, `create_guest_note`/`update_guest_note`/
  `delete_guest_note`.
- **Reservation**: `create_reservation`/`update_reservation`, `create_reservation_note`/
  `update_reservation_note`/`delete_reservation_note`, `check_in_room`/`check_out_room`/
  `assign_room`.
- **Financial** (approval required — real folio/payment transactions): `post_payment`/
  `void_payment`, `post_adjustment`/`delete_adjustment`, `post_item`/`void_item`.
- **Catalog/ops** (no approval required): `create_item_category`, `create_group_note`,
  `update_housekeeping_status`, `create_housekeeper`/`update_housekeeper`,
  `assign_housekeeping`, `create_house_account`/`update_house_account_status`.
- **Webhooks**: `create_webhook` (no approval), `delete_webhook` (approval required).
- **Room blocks**: `create_room_block`/`update_room_block` (no approval),
  `delete_room_block` (approval required — releases held rooms, irreversible for any already
  picked up).

## Known limits

- **`getTransactions` may no longer be a live Cloudbeds v1.2 endpoint.** The `transactions`
  stream is unchanged from the original legacy-parity migration (`GET /getTransactions`), but this
  Pass B research fetched Cloudbeds' current official OpenAPI spec
  (`github.com/cloudbeds/openapi-specs/src/pms-v1.2-openapi.yaml`) and confirmed `/getTransactions`
  does **not** appear anywhere in the current 115-endpoint spec — every other legacy-parity stream's
  path (`getGuestList`/`getHotels`/`getRoomBlocks`/`getReservations`) DOES appear. This is flagged
  as a real, confirmed API-surface discrepancy worth investigating (possible upstream removal/
  rename since legacy was written, or a different API version) rather than silently fixed or
  ignored — repointing a legacy-parity stream to a different real endpoint is a behavior change
  outside this Pass B expansion's scope (touch-only-what's-needed), so `api_surface.json` still
  declares `GET /getTransactions` as `covered_by: {stream: transactions}` (matching what the bundle
  actually requests) rather than silently dropping it.
- **`getUsers` is NOT implemented** (`api_surface.json`'s `out_of_scope` entry): its response's
  `data` is an object keyed by property ID whose VALUES are themselves arrays of user objects (a
  "keyed array of arrays" shape) — neither `records.path` (one dotted path to a single array) nor
  `records.keyed_object` (explodes a keyed object's OBJECT-valued entries, not array-valued ones)
  can express this. The endpoint's own OpenAPI schema also does not document individual user
  object fields at all (no `userID`/`name`/`email`/`role` properties defined), so even a
  hypothetical dialect extension would need undocumented-shape guessing.
- **`getPackages`/`getPackageNames` are NOT implemented**: both return `data.packageNames`, a bare
  array of package NAME strings, not record-shaped objects — `connsdk.RecordsAt` silently drops
  non-object array elements, yielding zero records always. Not a record catalog in the shape this
  dialect (or any list-endpoint dialect) expects.
- **Bulk rate management (`putRate`/`patchRate`/`getRateJobs`) is out of scope**: Cloudbeds' rate
  writes are an async, large-JSON-rate-grid bulk-update flow (submit a grid, poll `getRateJobs` for
  completion), not a single-record declarative mutation; `rate_plans` remains read-only.
- **Allotment-block writes (`create`/`update`/`delete`/`*Notes`) are out of scope**: the
  `allotment_blocks` stream is read-only; the block-management write surface is a cohesive group
  left for a future capability-expansion pass rather than partially implemented.
- **File/photo/document upload endpoints are excluded as `binary_payload`**
  (`postFile`/`postGuestPhoto`/`postGuestDocument`/`postReservationDocument`), and their
  corresponding list endpoint (`getFiles`) is excluded alongside them since it has no useful
  write-side counterpart in this bundle.
- **Payment-adjacent endpoints requiring PCI-scoped or redirect-driven flows are excluded**:
  `postCreditCard` (stores a raw PAN — out of scope for a declarative connector to ever transmit),
  `postCharge` (a hosted-payment-page redirect flow, not a single request/response mutation).
- **Group-profile CRUD beyond notes is out of scope**: `create_group_note` covers the common
  note-adding case; `putGroup`/`patchGroup` (full group-profile update) were deprioritized in this
  pass.
- **Integration-partner/app-settings endpoints are excluded as `requires_elevated_scope`**: these
  (`get/post/put/deleteAppPropertySettings`, `getAppSettings`, `get/postAppState`, `postAppError`)
  are scoped to the calling app's own third-party-integration configuration, not property/guest/
  reservation data, and several are documented as unavailable to ordinary property-level API
  clients at all.
