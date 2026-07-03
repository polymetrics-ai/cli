# Overview

Hubplanner is a read-only resource-scheduling source connector. It reads resources, projects,
clients, events, holidays, bookings, and billing rates through the Hubplanner REST API
(`https://api.hubplanner.com/v1`). This bundle migrates `internal/connectors/hubplanner` (the
hand-written legacy connector); the legacy package stays registered and unchanged until wave6's
registry flip.

## Auth setup

Provide a Hubplanner API key via the `api_key` secret. It is sent verbatim (no `Bearer` prefix) as
the `Authorization` header — `{"mode": "api_key_header", "header": "Authorization", "value": "{{
secrets.api_key }}", "prefix": ""}` — matching legacy's `connsdk.APIKeyHeader("Authorization",
secret, "")` wiring exactly. It is never logged.

## Streams notes

All 7 streams (`resources`, `projects`, `clients`, `events`, `holidays`, `bookings`,
`billing_rates`) share the same shape: `GET` against a Hubplanner singular-resource endpoint that
returns a bare JSON array at the response root (`records.path: ""`), primary key `["_id"]`
(every Hubplanner object exposes a string `_id`). Pagination is genuinely 0-indexed
(`pagination.type: page_number`, `start_page: 0`, `page_param: page`, `size_param: limit`,
`page_size: 200`) — the first request sends `page=0`, matching legacy's `hubplanner_test.go:34`
assertion and `harvest`'s `for page := 0; ...` loop exactly; pagination stops on a short/empty page
(fewer than `page_size` records), identical to legacy's `len(records) < pageSize` check. None of
the 7 streams declare an `incremental` block — the Hubplanner API only supports full-refresh pulls
of scheduling data (legacy's own `hubplannerStreams()` declares no `CursorFields` on any stream).

`check` issues a single bounded `GET /resource?page=0&limit=1`, mirroring legacy's `Check`
implementation exactly (a 1-record probe of the `resource` endpoint confirms auth and connectivity
without mutating anything).

## Write actions & risks

None. Hubplanner is read-only (`capabilities.write: false`); legacy's own `Write` always returns
`connectors.ErrUnsupportedOperation` and there is no reverse-ETL write target for this API.

## Known limits

- Only the 7 legacy-parity read streams are implemented; see `api_surface.json`. Hubplanner's
  documented API surface (timesheets, custom fields, teams, webhooks) is out of scope until Pass B.
- Every stream is full-refresh only; the upstream API exposes no incremental filter, matching
  legacy exactly.
