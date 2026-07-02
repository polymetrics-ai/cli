# Overview

Cloudbeds is a wave2 fan-out declarative-HTTP migration. It reads Cloudbeds guests, hotels, rooms,
reservations, and transactions through the Cloudbeds v1.2 REST API
(`GET https://api.cloudbeds.com/api/v1.2/...`). This bundle targets capability parity with
`internal/connectors/cloudbeds` (the hand-written connector it migrates); the legacy package stays
registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Cloudbeds API access token via the `api_key` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_key>`) and is never logged, matching legacy's
`connsdk.Bearer(secret)` (`cloudbeds.go:238`). `base_url` defaults to
`https://api.cloudbeds.com/api/v1.2` and may be overridden for tests/proxies (legacy's own
`cloudbedsBaseURL` validates scheme+host the same way; the engine's base-URL resolution has no
equivalent runtime validation, but every conformance fixture only ever points at an httptest
server, so this is not exercised differently on either side).

## Streams notes

All five streams (`guests` → `/getGuestList`, `hotels` → `/getHotels`, `rooms` →
`/getRoomBlocks`, `reservations` → `/getReservations`, `transactions` → `/getTransactions`) are
simple list endpoints; records live at the top-level `data` key alongside `success`/`count`
fields legacy never reads. Pagination is Cloudbeds' page-increment convention
(`pagination.type: page_number`, `page_param: pageNumber`, `size_param: pageSize`, `page_size: 100`
matching legacy's `cloudbedsDefaultPageSize`) — unlike Clockify, Cloudbeds sends BOTH `pageNumber`
AND `pageSize` on every request (legacy's `harvest`, `cloudbeds.go:144-146`), so `size_param` is
populated (not left empty) and the engine re-sends `pageSize` on every page, matching legacy's own
per-page query construction exactly. A page returning fewer than `page_size` records is the last
page, matching `cloudbeds.go:164-169`'s exact stop rule.

Cloudbeds is full-refresh only upstream — legacy's own `cloudbedsStreams` comment: "Cloudbeds is
full-refresh only upstream, so no cursor fields are published" — this bundle declares no
`incremental` block for any stream, matching legacy exactly. Primary keys follow each entity's own
Cloudbeds id field (`guestID`, `propertyID`, `reservationID`, `transactionID`) rather than a
uniform `id`, matching legacy's per-stream `PrimaryKey` declarations exactly (`streams.go:30-63`).

## Write actions & risks

None. Cloudbeds is read-only in this connector (legacy's `Write` always returns
`connectors.ErrUnsupportedOperation`, `cloudbeds.go:309-311`); `capabilities.write` is `false` and
this bundle ships no `writes.json`.

## Known limits

- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes both as config-driven
  overrides (`cloudbedsPageSize`/`cloudbedsMaxPages`, `cloudbeds.go:271-299`). The engine's
  `page_number` paginator's `PageSize`/`MaxPages` fields are plain JSON values in `streams.json`,
  not templated against `config.*` — there is no mechanism in this dialect to wire a runtime config
  value into either field. This bundle ships legacy's own default (`page_size: 100`, `max_pages`
  unbounded) as a static value.
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only reached
  when `config.mode == "fixture"`, a credential-free conformance-harness affordance,
  `cloudbeds.go:175-222`) stamps a broader, cross-stream synthetic record shape (every fixture
  record carries `guestID`, `reservationID`, `transactionID`, `amount`, `status`, etc. regardless of
  which stream it represents) that does not match any single stream's real live-API record shape.
  This bundle's schemas and fixtures target the live per-stream record shape only; the engine's own
  conformance/fixture-replay harness provides the credential-free test affordance this bundle
  needs, so no fixture-mode equivalent is needed here.
- **`success`/`count` envelope fields are not modeled.** Every Cloudbeds list response wraps
  `data` alongside `success: true` and a `count` field; legacy never reads either (only `data` via
  `connsdk.RecordsAt(resp.Body, "data")`, `cloudbeds.go:152`), so this bundle likewise ignores them
  — they carry no record-level data and are outside `stream.projection`'s scope by construction.
