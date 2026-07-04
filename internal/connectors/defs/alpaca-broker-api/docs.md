# Overview

Alpaca Broker API reads accounts, assets, market calendar, clock, country info, account activities,
journals, and per-account positions/watchlists/orders/document-metadata through the Alpaca Broker
REST API. This bundle originally migrated `internal/connectors/alpaca-broker-api` (the hand-written
connector, 5 legacy-parity streams) to a declarative defs bundle; this Pass B pass expands it to the
full documented Broker API surface (docs.alpaca.markets) — 11 read streams total, still read-only —
while keeping every original legacy-parity stream and behavior unchanged. The legacy package stays
registered and unchanged until wave6's registry flip.

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

Five streams matching legacy's `streamEndpoints` table, unchanged from the original migration:

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

None of the original 5 streams are incremental. Legacy declares a `CursorFields:
[]string{"created_at"}` hint on the `accounts` stream's catalog entry, but its `Read` never actually
performs cursor-based filtering (no `incremental`/state-driven query logic exists anywhere in
`alpaca_broker_api.go`) — the field is a dead catalog hint, not functioning incremental behavior.
This bundle therefore declares no `incremental` block and no `x-cursor-field` on `accounts`' schema,
matching legacy's REAL (full-refresh only) behavior rather than its unused catalog metadata.

**New in Pass B** (full-surface expansion — `api_surface.json` covers the complete Broker API
reference surface):

- `account_activities` — `GET /accounts/activities`, records at the top-level array, sends a static
  `page_size=100` query param. Paginated identically to `accounts` (`pagination.type: cursor`,
  `cursor_param: page_token`, `last_record_field: id`) — Alpaca's activities endpoint uses the exact
  same `page_token`-of-last-record convention. Every trade/non-trade activity across every account is
  returned (no `account_id` filter is set), matching the endpoint's own documented default scope.
- `journals` — `GET /journals`, records at the top-level array, unpaginated (Alpaca's own docs show a
  plain `limit` cap with no cursor token for this endpoint; the default response size is well within
  one page for a typical operator's journal volume).
- `positions`, `watchlists`, `orders`, `documents` — each a **fan_out stream** (the engine's
  `fan_out` dialect, `docs/migration/conventions.md` §3) resolving its account id list from
  `ids_from.request: {"path": "/accounts", "records_path": "", "id_field": "id"}` (a plain,
  unpaginated GET against the SAME `/accounts` endpoint the `accounts` stream itself reads, since
  none of these 4 fan_out streams declares its own `pagination` block — the id-listing sub-request
  therefore issues exactly one request, matching a `none`-type paginator), then repeating a full
  per-account request against:
  - `positions` — `GET /trading/accounts/{account_id}/positions`. Alpaca's response objects carry no
    `id` field of their own; `computed_fields` stamps `id` from the bare `{{ record.asset_id }}`
    reference (typed extraction, `docs/migration/conventions.md` §3) so the schema's
    `x-primary-key: ["id", "account_id"]` is satisfiable. `stamp_field: "account_id"` writes the
    resolved account id onto every emitted position record.
  - `watchlists` — `GET /trading/accounts/{account_id}/watchlists`. Each watchlist object already
    carries its own `account_id` field in the raw response; `stamp_field` writes the SAME value
    (the fan_out id and the raw field agree by construction, since the id being fanned out over IS
    that account), so no computed_fields override is needed.
  - `orders` — `GET /trading/accounts/{account_id}/orders?status=all` (the static `status=all` query
    param requests every order regardless of status, matching a full-history read rather than the
    endpoint's own `status=open` default).
  - `documents` — `GET /accounts/{account_id}/documents` (note: NOT under `/trading/accounts/...`
    like the other 3 — Alpaca's account-document-metadata endpoint lives directly under
    `/accounts/{account_id}`). Lists only document METADATA (id, type, sub_type, name, date); no
    document content is ever fetched (see Known limits — the binary download endpoint is excluded).

## Write actions & risks

None. This connector remains `capabilities.write: false`; no `writes.json` is shipped. Every mutation
endpoint in the Broker API's documented surface either moves real money (transfers, journals, ACH/
wire funding-instrument creation), places or cancels a live trade (orders, position liquidation),
or mutates compliance-sensitive account state (KYC/identity/PDT-status updates, account opening) —
none of these is a safe unattended reverse-ETL write target for a metadata/reporting connector.
See `api_surface.json`'s `destructive_admin`/`requires_elevated_scope` entries for the full,
endpoint-by-endpoint reasoning.

## Known limits

- Trading/order-placement, position-liquidation, funding (transfers/ACH relationships/recipient
  banks), journal-creation, account-opening/KYC-update, IPO offerings, OAuth flows, and IRA
  excess-contribution endpoints are out of scope; see `api_surface.json`'s `excluded` entries for the
  specific category (`destructive_admin`/`requires_elevated_scope`/`out_of_scope`) and reasoning per
  endpoint.
- Every Server-Sent-Events (SSE) subscription endpoint (`/v1/events/*`) is excluded as
  `non_data_endpoint`: SSE is a persistent-connection push protocol, not a declarative poll-style
  list/detail resource the engine's HTTP dialect can express.
- Document CONTENT is never fetched — only listing metadata (`documents` stream). The binary
  download endpoint (`GET .../documents/{document_id}/download`, a redirect to a pre-signed URL
  serving a PDF) is excluded as `binary_payload`.
- `GET /v1/accounts/{account_id}` (single-account detail), `GET /v1/assets/{symbol_or_asset_id}`
  (single-asset detail), and every single-order/single-position/single-watchlist detail-by-id
  endpoint are excluded as `duplicate_of` their corresponding list stream: Alpaca's list endpoints
  already return the complete object shape per item (confirmed against Alpaca's own docs for
  accounts — `GET /v1/accounts` returns the full `AccountExtended` shape, not a summary), so a
  detail-by-id endpoint adds no data a list stream doesn't already carry.
- `GET /v1/trading/accounts/{account_id}/account/pdt/status` and `GET /v1/accounts/positions` are
  excluded as `deprecated` per Alpaca's own documentation (the PDT-status endpoint sunsets
  2026-07-06; the bulk-positions endpoint is superseded by `/v1/reporting/eod/positions`).
  `GET /v1/corporate_actions/announcements` is likewise `deprecated`, and additionally requires a
  mandatory 90-day-max date range this connector's static config model has no natural place for.
- See the "Streams notes" section above for the `accounts` stream's pagination-stop parity deviation
  and the dropped (non-functioning) `created_at` cursor-field catalog hint — both predate this Pass B
  pass and are unchanged.
