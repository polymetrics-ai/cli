# Overview

WaiterAid is a restaurant reservation and table-management platform. This bundle reads and writes
restaurant reservations, meals, guests, and walk-in queue entries via the real WaiterAid API. It
originally migrated `internal/connectors/waiteraid` (the hand-written legacy connector, a single
unpaginated `GET /reservations` read), which stays registered and unchanged until wave6's registry
flip.

**Pass B full-surface expansion found that the real, currently-documented WaiterAid API
(`app.waiteraid.com/api-docs/index.html`) has NO REST-shaped resource surface at all.** It is
entirely a flat, POST-method, query-string-parameterized method-call API rooted at
`/wa-api/<methodName>` (e.g. `/wa-api/searchBooking`, `/wa-api/addBooking`) — the prior version's
`GET {base_url}/reservations` targeted a URL shape (and a host, `api.waiteraid.com`, distinct from
the real `app.waiteraid.com`) that the real API does not expose at all. This is a full rewrite of
the bundle's request shape, not an incremental extension; the connection config CONTRACT (which
secrets/config keys an operator provides) is preserved as closely as the real API allows — see Auth
setup.

## Auth setup

Two config values, matching legacy's own contract shape: `auth_hash` (`x-secret`) and `restid`
(reclassified from `x-secret` to a plain string in this revision — see below). **Pass B
correction**: the real API authenticates via an `auth_hash` QUERY STRING parameter on every
request (obtained once via the real API's own `POST /wa-api/auth?username=&password=&restaurant=`
exchange, or supplied directly if already known — this bundle expects the latter, an already-
obtained hash, exactly matching legacy's config contract of two opaque credential-shaped values
rather than raw username/password), NOT the `X-Auth-Hash`/`X-Restaurant-ID` HEADERS this bundle
(and legacy) previously sent — the real, documented API has no header-based auth scheme of any
kind. `streams.json`'s `base.auth` now declares `{"mode": "api_key_query", "param": "auth_hash",
"value": "{{ secrets.auth_hash }}"}`.

`restid` is now sent as the real API's own mandatory `restid` query parameter on literally every
documented endpoint ("Restaurant id is mandatory for api request" appears on every single method's
parameter list). It is reclassified from `x-secret: true` to a plain string property in this
revision: a restaurant id is a plain numeric identifier, not a credential (conventions.md's
`x-secret` rule is about a field's genuine credential-shaped NATURE, and a restaurant id has none —
unlike, say, an optional Bearer-proxy key that merely isn't currently wired but is still
credential-shaped). Since the engine dialect's write-action `path` templating has no base-level
"send this on every request" mechanism (only `base.headers`/`base.auth` provide that, and `restid`
is neither a header nor an auth candidate under the real API), every stream, the check request, AND
every write action individually templates `restid={{ config.restid }}` into its own `path`/`query`
— a necessary repetition, not an oversight.

`base_url` now defaults to `https://app.waiteraid.com` (was `https://api.waiteraid.com`) — the real
API's documented host; the prior host does not serve any of these endpoints.

## Streams notes

- **`reservations`** — `POST /wa-api/searchBooking` (real API: POST method, but every parameter —
  including `auth_hash`/`restid` — is a QUERY parameter, never a body field; `body_type: "none"` for
  the same reason writes below use it). Records are read from the response's `bookings` array
  (`records.path: "bookings"`, not the prior version's fictional `reservations` key). The real wire
  shape is richer than legacy, but emitted records stay at legacy parity:
  `{id, guest_name, date, status}`. `guest_name` is copied from a legacy-shaped top-level
  `guest_name` when present, or from the real API's nested `guest.name` field. `config.start_date` (preserved for
  config-shape backward compatibility) is now sent as the real API's own `date` query parameter
  (an exact-match date filter, not a lower-bound range) via the optional-query dialect
  (`omit_when_absent: true`) whenever configured.
- **`meals`** — `POST /wa-api/getMeals` (Pass B new stream): the list of bookable meal/service
  slots (`id`, `name`, `min_start`, `max_end`) for the configured restaurant. Useful as a lookup
  table for `add_booking`'s required `mealid` field.
- **`queue`** — `POST /wa-api/queue/list` (Pass B new stream): the restaurant's walk-in queue for a
  date (`queue_id`, `cust_id`, `amount`, `added_date`, `firstname`, `lastname`, `mobile`,
  `comment`), records read from the response's `queue` array. Sends the same `date`
  optional-query-dialect filter as `reservations` (`config.start_date`, `omit_when_absent: true`)
  since the real API's own `queue/list` accepts an identical exact-match `date` parameter,
  defaulting to the current date server-side when omitted. Completes this pass's coverage of the
  queue sub-resource family alongside the pre-existing `add_to_queue`/`delete_from_queue` writes.

None of the 3 streams declare `pagination` or `incremental`: the real API documents no cursor/page
parameter for any of them (a `limit`/`page` param exists for `searchBooking` but paginates by
RESULT COUNT truncation, not a cursor — legacy never implemented pagination either, and this bundle
does not introduce new pagination behavior legacy never had).

## Write actions & risks

6 write actions added in this Pass B expansion (legacy shipped none — `Write` always returned
`ErrUnsupportedOperation`). All are `body_type: "none"` (every parameter, including the dynamic
per-record ones, is templated directly into the `path`'s query string — the write dialect has no
`query` field of its own, unlike streams/check, so a literal `?key={{ }}&key2={{ }}` query string is
embedded straight into `path`, which the engine's `InterpolatePath`+`url.Parse` correctly treats as
the request's query component). External mutations visible to restaurant staff and/or guests;
**approval required** for every one:

- `add_booking` — `POST /wa-api/addBooking`. Requires `start_time`, `amount`, `date`, `mealid`.
- `set_booking_status` — `POST /wa-api/setBookingStatus`. Sets a booking's status to one of
  `arrived`/`guest_left`/`paid`/`all_seated`/`deleted`/`undelete` — this is how a reservation is
  cancelled (`deleted`) via this API, there being no distinct delete endpoint.
- `edit_booking` — `POST /wa-api/editBooking`. Only `start_time` is modeled as a required,
  always-sent field; see Known limits for why the real API's other optional edit keys
  (`sign`/`end_time`/`mealid`/`comment`) are not modeled.
- `add_guest` — `POST /wa-api/addGuest`. Only `firstname`/`lastname` are modeled as required,
  always-sent fields; see Known limits for the same reason.
- `add_to_queue` — `POST /wa-api/queue/add`. Requires `name`, `amount`.
- `delete_from_queue` — `POST /wa-api/queue/delete`. Requires `queue_id`.

## Known limits

- **Optional real-API fields are not modeled on `edit_booking`/`add_guest`/`add_booking`.** The
  write-action dialect's `path` templating (unlike `stream.Query`'s opt-in object form) has NO
  absent-key-falsy tolerance at all — every `{{ record.x }}` reference in a write action's `path`
  is a hard error if `x` is absent from the record. The real API documents several genuinely
  OPTIONAL edit/create keys (`edit_booking`'s `sign`/`end_time`/`mealid`/`comment`;
  `add_booking`'s `end_time`/`children_amount`/`comment`/`guestId`; `add_guest`'s
  `email`/`phone`/`address`/`city`/`postalcode`/`guestId`) that this bundle does not expose, since
  doing so would force every write to always supply them (turning an optional field into a
  required one — a real accepted-input narrowing, not a cosmetic gap). If this exact shape (a
  write action needing optional path/query-string parameters) recurs elsewhere, it is an
  `ENGINE_GAP` candidate for a `write.query`-style dialect addition mirroring `stream.Query`'s
  opt-in object form (conventions.md §6's recurrence rule) — not a per-connector workaround.
- **`searchGuest` is not modeled as a stream** (`SCHEMA_AMBIGUOUS`). Its documented example response
  is a top-level object with numeric-string keys sitting alongside a sibling `status` key
  (`{"status":"OK", 0: {...}, 1: {...}}`) — neither a named array (`connsdk.RecordsAt`) nor the
  `records.keyed_object` shape (whose exploded VALUES must all be objects at one dotted path;
  `status` is a string, which `keyed_object` would still try to explode). See `api_surface.json`.
- **Messaging (`getMessageThreads`/`getMessage`/`writeMessage`) and scheduling-configuration
  endpoints (`getAvailibilityStatus`/`toggleMealAvailibility`/`toggleMealStatus`/
  `getSeatingLength`/`controlSeatingLength`) are out of scope** — distinct product surfaces from
  reservations/guests/queue; see `api_surface.json` for the full per-endpoint reasoning.
- **No live-API verification was possible.** All endpoint shapes come from the real API's own
  published documentation (`app.waiteraid.com/api-docs/index.html`), not a live test account. Given
  how substantially this revision changes the connector's entire request shape (host, method, auth
  scheme, and every path), a real-account smoke test is strongly recommended before first
  production use.
