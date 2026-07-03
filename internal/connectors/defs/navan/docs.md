# Overview

Navan is a corporate travel-management platform exposing a REST travel API. This bundle reads
flight, hotel, car, and rail bookings from `https://api.navan.com/v1/bookings` (bookingType-
partitioned into four streams) using OAuth2 client-credentials authentication. It is read-only,
migrated from `internal/connectors/navan` (the hand-written connector this bundle replaces at
capability parity); the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide `client_id` and `client_secret` secrets; the engine's declarative
`oauth2_client_credentials` auth mode exchanges them for a Bearer token via the OAuth2
client-credentials grant against `{{ config.base_url }}/ta-auth/oauth/token` (matching legacy's
`navanTokenPath` derivation off the same host as the API), refreshing automatically before expiry —
matching legacy's `connsdk.OAuth2ClientCredentials` usage exactly (both connectors send a standard
`grant_type=client_credentials` form request with no custom body shape or extra params). Both
secrets flow only into the token exchange and are never logged.

Set `base_url` to override the default (`https://api.navan.com`) for a test proxy; the token
endpoint always derives from the same host, matching legacy.

## Streams notes

All four streams share one underlying endpoint, `GET /v1/bookings`, partitioned by a fixed
`bookingType` query value per stream — `bookings` (`FLIGHT`), `hotel_bookings` (`HOTEL`),
`car_bookings` (`CAR`), `rail_bookings` (`RAIL`) — matching legacy's `navanStreamEndpoints` routing
table, which points all four at the identical `v1/bookings` resource with only the `bookingType`
fixed param differing.

Pagination is genuinely 0-indexed page-increment (`pagination.type: page_number`, `page_param:
page`, `size_param: size`, `start_page: 0`, `page_size: 50` — legacy's `navanDefaultPageSize`),
matching legacy's `harvest` loop exactly (`for page := 0; ...`, sending `page=0` on the very first
request). Pagination stops on a short page (fewer than 50 records returned), matching legacy's own
`len(records) < pageSize` stop rule; the engine's `page_number` paginator implements the identical
check.

Every stream declares `incremental.cursor_field: last_modified`, `request_param: createdFrom`,
`start_config_key: start_date` — matching legacy's `incrementalLowerBound` exactly: the stored state
cursor (tracking `last_modified`) is preferred, falling back to the `start_date` config value, and
sent (whichever resolves) as the `createdFrom` query filter; when neither resolves (a fresh full
sync with no `start_date` set), `createdFrom` is omitted entirely, matching legacy's `if
createdFrom != "" { base.Set("createdFrom", createdFrom) }`. No `param_format` is declared since
legacy sends both the RFC3339 `start_date` config value and the raw persisted cursor string
verbatim with no reformatting — the engine's default (`rfc3339`, send-as-resolved) matches this
exactly.

Every camelCase API field is renamed to its legacy snake_case column name via `computed_fields`
(`bookingId` -> `booking_id`, `lastModified` -> `last_modified`, etc. — matching legacy's
`navanBookingRecord` mapper field-for-field); `uuid`/`currency`/`destination`/`created`/`domestic`/
`expensed` need no rename since legacy passes them through under the same name, so plain schema
projection copies them verbatim with no `computed_fields` entry required. Primary key is `["uuid"]`
and the incremental cursor field is `last_modified` for every stream, matching legacy's uniform
catalog.

## Write actions & risks

None. Navan is read-only in both legacy and this bundle (`capabilities.write: false`); no
`writes.json` file is shipped. Legacy's own comment states Navan is "a read-only source connector."

## Known limits

- **Dynamic (fixture-replay) conformance checks are marked `skip_dynamic` at the bundle level**
  (`metadata.json`'s `conformance` block). `oauth2_client_credentials` auth's `token_url` is
  derived from `{{ config.base_url }}/ta-auth/oauth/token`; conformance's `withReplayURL` only
  overrides `b.HTTP.URL` (the base request URL used for stream/check paths), never
  `RuntimeConfig.Config["base_url"]` itself, so the `token_url` template still resolves to the
  synthetic non-secret value (`"synthetic-conformance-value/ta-auth/oauth/token"`), an unreachable
  non-URL — the OAuth token exchange fails before any declarative stream/check request is ever
  issued, so every auth-resolving dynamic check would otherwise fail identically and
  uninformatively. Static checks (spec/schema validity, `interpolations_resolve`, docs/fixtures
  presence, secret redaction) still run and pass. This bundle has no Tier-2 `AuthHook` (auth is
  fully declarative `oauth2_client_credentials`), so there is no `paritytest/navan` package for this
  wave; the read/pagination/schema-projection shape is proven by structural review against legacy
  `internal/connectors/navan` instead. Matches `dwolla`'s, `clazar`'s, and `sendpulse`'s identical
  documented precedent.
- Legacy accepted runtime-configurable `page_size` (1-200, default 50) and `max_pages` (0/all/
  unlimited or a positive integer cap) config values. The engine's `PaginationSpec.PageSize`/
  `MaxPages` fields are fixed literals declared once in `streams.json`'s pagination block — there is
  no dialect mechanism to make either dynamically config-driven per request (matching the same
  documented limitation `docs/migration/conventions.md` and this bundle's sibling `beamer`/
  `sendpulse` record for their own `page_size` config surface). This bundle bakes in legacy's own
  default page size (`size=50`, legacy's `navanDefaultPageSize`) and leaves `max_pages` unset
  (unbounded, matching legacy's own default when `max_pages` is unset/`all`/`unlimited`). Neither
  `page_size` nor `max_pages` is declared in `spec.json` — a declared-but-unwireable config key is
  worse than an absent one (F6, `docs/migration/conventions.md`). This narrows accepted
  CONFIGURATION surface only, never emitted record data for any input legacy would accept with its
  own default page size.
- `metadata.json` declares no `rate_limit` block: legacy enforces no client-side rate limiting for
  Navan, so none is added here either (matches legacy's real, lack-of, throttling behavior).
