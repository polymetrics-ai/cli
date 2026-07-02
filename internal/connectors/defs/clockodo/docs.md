# Overview

Clockodo is a wave2 fan-out declarative-HTTP migration. It reads Clockodo customers, projects,
services, and users through the Clockodo REST API (`GET https://my.clockodo.com/api/v2/...`). This
bundle targets capability parity with `internal/connectors/clockodo` (the hand-written connector
it migrates); the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Clockodo authenticates with three custom request headers rather than a bearer token or a single
API-key header, matching legacy's `requester` function (`clockodo.go:223-254`) exactly:

- `X-ClockodoApiUser`: the account email (`config.email_address`, required)
- `X-ClockodoApiKey`: the API key (`secrets.api_key`, required, `x-secret`, never logged)
- `X-Clockodo-External-Application`: an `application;contact` identifier Clockodo requires of every
  API client (`config.external_application`, required)

An optional `Accept-Language` header is sent when `config.language` is set (declared as an
optional — not `required[]` — spec property, so `streams.json`'s conditional-header omission rule
(conventions.md §3, the Stripe-Account/`account_id` pattern) leaves it off the request entirely
when unset, matching legacy's `if lang := ...; lang != ""` guard, `clockodo.go:245-247`). `auth` is
declared as `[{"mode": "none"}]` since none of Clockodo's three custom headers are one of the
engine's built-in auth modes (bearer/basic/api_key_header carries exactly one header) — all three
credential-bearing values are wired through `streams.json`'s `base.headers` instead, which the
`secrets.*` always-hard-error-on-absence rule (conventions.md §3) covers identically to a
dedicated auth mode would.

## Streams notes

`customers` and `projects` are page-based paginated endpoints: Clockodo returns
`{paging:{current_page,count_pages,...}, <recordsKey>:[...]}` and legacy requests
`page=<n+1>` until `current_page >= count_pages` (`clockodo.go:133-171`). The engine's declarative
dialect has no field capable of reading `current_page`/`count_pages` out of the response body to
drive the next page number (the `cursor` pagination type's `token_path`/`last_record_field`
sources don't fit a plain incrementing integer with no token semantics) — this bundle instead uses
`pagination.type: page_number` with `size_param: ""` (Clockodo's real wire behavior: legacy never
sends a page-size query param at all; the server applies its own fixed `items_per_page`, 50 by
default) and `page_size: 50` purely as the engine's client-side short-page stop threshold. Because
Clockodo's real per-page record count is a stable, server-controlled 50 across every page except
the last, "stop when a page returns fewer than 50 records" and legacy's own "stop when
`current_page >= count_pages`" terminate at the exact same page in every observed case (both
degrade to the identical "an empty/short final page ends the stream" rule when the last page
happens to be exactly full — see Known limits).

`services` and `users` are non-paginated in legacy (`paginated: false`, `clockodo.go:159-161`
returns after the first request) — this bundle overrides `pagination` to `{"type": "none"}` at the
stream level for both, matching legacy exactly (`streams.json`'s per-stream `pagination` replaces
the base spec wholesale per conventions.md §3).

None of Clockodo's four list endpoints expose an incremental cursor field (legacy's own
`clockodoStreams` comment: "Clockodo administrative resources are keyed by an integer `id` and have
no natural incremental cursor, so all are full-refresh") — this bundle declares no `incremental`
block for any stream, matching legacy exactly.

## Write actions & risks

None. Clockodo is read-only (legacy's `Write` always returns
`connectors.ErrUnsupportedOperation` wrapped with a read-only message, `clockodo.go:120-125`);
`capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- **The engine's `page_number` paginator stops on a short PAGE (record count < page_size), not on
  legacy's own `current_page >= count_pages` signal.** These two stop rules are DATA-equivalent for
  every real Clockodo response shape observed in legacy's own test fixtures and this bundle's
  parity/conformance fixtures (a full, non-final page always has exactly `items_per_page` records;
  a final page is always short or empty) — see Streams notes above. The one theoretical edge case
  where they could diverge (the true last page happens to contain EXACTLY `items_per_page`
  records) still terminates correctly on both sides: legacy would issue one further request that
  returns zero records and stop via its own `len(records)==0` check, while the engine's paginator
  would issue that same extra request and stop on the same short/empty final page — an extra
  request, never an extra or missing RECORD, so emitted data is unaffected either way.
- **`max_pages` is not runtime-configurable.** Legacy exposes `max_pages` as a config-driven
  override (`clockodoMaxPages`, `clockodo.go:284-297`) applied as a hard request-count cap. The
  engine's `PaginationSpec.MaxPages` field is a plain JSON value in `streams.json`, not templated
  against `config.*` — there is no mechanism in this dialect to wire a runtime config value into
  it. This bundle leaves `max_pages` unset (unbounded), matching legacy's own default
  (`max_pages=0` meaning unlimited) but removing the operator's ability to lower it at read time.
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only reached
  when `config.mode == "fixture"`, a credential-free conformance-harness affordance,
  `clockodo.go:187-217`) stamps a broader, cross-stream synthetic record shape (every fixture
  record carries `email`, `role`, `budget_money`, `teams_id`, etc. regardless of stream) that does
  not match any single stream's real live-API record shape. This bundle's schemas and fixtures
  target the live per-stream record shape only; the engine's own conformance/fixture-replay harness
  provides the credential-free test affordance this bundle needs, so no fixture-mode equivalent is
  needed here.
