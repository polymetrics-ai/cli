# Overview

CallRail is a wave2 fan-out declarative-HTTP migration. It reads CallRail calls, companies, users,
and text messages through the CallRail v3 REST API (`GET https://api.callrail.com/v3/...`). This
bundle targets capability parity with `internal/connectors/callrail` (the hand-written connector it
migrates); the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a CallRail API key via the `api_key` secret; it is sent as `Authorization: Token
token="<api_key>"` (with the value itself quoted inside the header, matching legacy's
`fmt.Sprintf("Token token=%q", apiKey)` at `callrail.go:254`) via `auth.mode: api_key_header` with
`prefix: "Token token="` and a `value` template (`"\"{{ secrets.api_key }}\""`) that supplies the
surrounding literal quotes; the secret itself is never logged. The required `account_id` config
value is substituted into every request path (`/a/{{ config.account_id }}/...`, urlencoded by
`InterpolatePath`'s per-segment default, matching legacy's own `url.PathEscape(account)` in
`accountPath`). `base_url` defaults to `https://api.callrail.com/v3` and may be overridden for
tests/proxies.

## Streams notes

All 4 streams (`calls`, `companies`, `users`, `text_messages`) share the same pagination shape:
CallRail's page-number convention (`pagination.type: page_number`, `page_param: page`,
`size_param: per_page`) — a page shorter than `page_size` stops pagination, which is functionally
equivalent to legacy's own primary stop signal (`total_pages` reached) for every real CallRail
response (the last page is never longer than `per_page`); the one edge case where a result set is
an exact multiple of `per_page` costs one extra, empty-page request on the engine side that legacy's
`total <= page` check would have avoided, never a data difference (both sides stop after emitting
the same records). `page_size` is `100`, matching legacy's own default (see Known limits for why it
is not runtime-configurable).

Each stream sends `start_date` (`param_format: date`, converting a resolved RFC3339 lower bound to
`YYYY-MM-DD` for the wire) computed either from the sync's persisted cursor or, on a fresh sync, from
the `start_date` config value — matching legacy's `startDateParam`/`dateOnly` exactly for the
RFC3339-input case (see Known limits for the accepted-input narrowing this requires). Per-stream
cursor fields match legacy's own `CursorFields` declarations: `calls` -> `start_time`,
`companies`/`users` -> `created_at`, `text_messages` -> `last_message_at`.

## Write actions & risks

None. CallRail is exposed read-only, matching legacy's `Write` returning
`connectors.ErrUnsupportedOperation`; `capabilities.write` is `false` and this bundle ships no
`writes.json`.

## Known limits

- **`start_date` config input is narrowed to RFC3339 (or bare Unix-seconds), no longer bare
  `YYYY-MM-DD`.** Legacy's `startDateParam` (`callrail.go:266-282`) accepts EITHER a bare
  `YYYY-MM-DD` value or full RFC3339 for the `start_date` config value, narrowing either to a date
  string itself before sending it as the `start_date` query param. The engine's `param_format: date`
  conversion (`formatParam`/`parseLowerBoundTime`) only accepts an all-digits (Unix-seconds) value or
  a full RFC3339 timestamp — a bare `"2026-01-01"` fails to parse as RFC3339 and hard-errors. This
  bundle's `spec.json` therefore declares `start_date` as `format: date-time` (RFC3339 only), a
  documented config-surface narrowing versus legacy's more permissive YYYY-MM-DD-or-RFC3339
  acceptance; any RFC3339 `start_date` value (e.g. `"2026-01-01T00:00:00Z"`) still produces the exact
  same `YYYY-MM-DD` wire value legacy would send for the equivalent date.
- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes both as config overrides
  (`callrailPageSize`/`callrailMaxPages`, `callrail.go:348-376`, `page_size` defaulting to 100,
  capped at 250). The engine's `page_number` paginator's `PageSize` is a static bundle-authored int
  (not templated), and there is no `MaxPages`-equivalent config-driven knob either; `max_pages` is
  unbounded (matching legacy's own `max_pages=0`/`all`/`unlimited` default). `page_size` is fixed at
  `100` to match legacy's own default exactly; the conformance fixture for `calls` is a single page
  of 3 records (all `total_records`) — a short page relative to `page_size: 100` — so
  `pagination_terminates` observes exactly one request, matching the real one-request-in-production
  behavior for any result set under 100 records; `companies`/`users`/`text_messages` are likewise
  single fixture pages.
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only reached
  when `config.mode == "fixture"`) stamps a `previous_cursor` field onto every fixture-mode record
  when a prior cursor happens to be set (`callrail.go:206-239`); this is not part of the live record
  shape. This bundle's schemas and fixtures target the live path only.
- Full CallRail API surface (trackers, form submissions, tags, call tagging writes) is out of scope;
  see `api_surface.json`'s `excluded` entries.
