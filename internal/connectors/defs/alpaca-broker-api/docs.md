# Overview

Alpaca Broker API reads accounts, assets, market calendar, clock, and country info through the
Alpaca Broker REST API. This bundle migrates `internal/connectors/alpaca-broker-api` (the
hand-written connector) to a declarative defs bundle at capability parity; the legacy package stays
registered and unchanged until wave6's registry flip. It is read-only — the upstream catalog declares
full_refresh only, matching legacy.

## Auth setup

Provide `username` (the Alpaca API Key ID, a plain config value) and `password` (the Alpaca API
Secret Key, a secret) — sent as HTTP Basic auth (`Authorization: Basic base64(username:password)`).
The secret is never logged.

`base_url` defaults to the sandbox host (`https://broker-api.sandbox.alpaca.markets/v1`), matching
legacy's default when its `environment` config is unset. Legacy also accepts an `environment` enum
(`api`/`paper-api`/`broker-api.sandbox`) that derives one of two fixed URLs at runtime; this bundle
requires the resolved `base_url` directly instead, since the engine's `spec.json` `"default"`
mechanism only materializes a single fixed literal, not an enum-derived choice between two URLs
(documented scope narrowing). Set `base_url` to `https://broker-api.alpaca.markets/v1` for
production.

## Streams notes

Five streams, matching legacy's `streamEndpoints` table:

- `accounts` — `GET /accounts`, records at the top-level array (`records.path: "."`), sends `limit`
  (default 20, matching legacy's `defaultLimit`). Paginated with `pagination.type: cursor`,
  `cursor_param: page_token`, `last_record_field: id` — the next page's `page_token` is the raw `id`
  of the last record on the current page, matching legacy's `harvest` loop exactly. Legacy stops on
  "page shorter than limit, or no last-record id"; the engine's `last_record_field` paginator stops on
  "page has zero records, or no last-record id" — in the edge case where a true final page happens to
  return exactly `limit` records, this bundle issues one additional (correctly empty) request before
  stopping where legacy would have already stopped on the short-page check. This never changes any
  emitted record's data (the extra request's page is itself empty) — see
  `docs/migration/conventions.md`'s parity-deviation ledger meta-rule.
- `assets`, `calendar`, `country_info` — unpaginated `GET` endpoints returning a full top-level array
  in one response (`records.path: "."`), each sending `limit` exactly like legacy's `readList`.
- `clock` — `GET /clock`, a singleton object response (`records.path: "."`, which `connsdk.RecordsAt`
  treats as a one-element record set for a top-level JSON object), never sends `limit`, matching
  legacy's `readSingleton`.

None of these streams are incremental. Legacy declares a `CursorFields: []string{"created_at"}` hint
on the `accounts` stream's catalog entry, but its `Read` never actually performs cursor-based
filtering (no `incremental`/state-driven query logic exists anywhere in `alpaca_broker_api.go`) — the
field is a dead catalog hint, not functioning incremental behavior. This bundle therefore declares no
`incremental` block and no `x-cursor-field` on `accounts`' schema, matching legacy's REAL (full-refresh
only) behavior rather than its unused catalog metadata.

## Write actions & risks

None. This connector is `capabilities.write: false`; no `writes.json` is shipped, matching legacy's
`Write` always returning `connectors.ErrUnsupportedOperation`.

## Known limits

- Trading/order/position, funding/transfer, journal, and account-status-event endpoints are out of
  scope for this migration; see `api_surface.json`'s `excluded: {category: out_of_scope, reason:
  "Pass B capability expansion"}` entries. Only the 5 legacy-parity read streams are implemented.
  `POST /v1/journals` is separately excluded as `destructive_admin` (moves funds between accounts).
- See the "Streams notes" section above for the `accounts` stream's pagination-stop parity deviation
  and the dropped (non-functioning) `created_at` cursor-field catalog hint.
