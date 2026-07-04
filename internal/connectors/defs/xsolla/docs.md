# Overview

Xsolla is a declarative bundle migrated from `internal/connectors/xsolla` (the hand-written legacy
connector, which stays registered and unchanged until wave6's registry flip). Pass B full-surface
expansion added the documented Xsolla Pay Station reporting surface (transaction search,
transaction registry, payouts, payout currency breakdown, financial reports) and full/partial
transaction refund writes. The original legacy streams `projects`, `orders`, and `transactions` are
also preserved with their original paths and raw `items` envelope for fidelity.

## Auth setup

Provide the Xsolla merchant ID and merchant API key as the `merchant_id`/`api_key` config/secret.
Both are sent as HTTP Basic auth (`merchant_id` as username, `api_key` as password) on every
request, matching Xsolla's own documented Basic-auth convention for merchant-scoped Pay Station
API calls. `api_key` is never logged.

## Streams notes

The 3 legacy streams use the paths from `internal/connectors/xsolla`: `projects` reads `GET
/projects`, while `orders` and `transactions` read `GET /projects/{{ config.project_id }}/orders`
and `GET /projects/{{ config.project_id }}/transactions`. They use legacy `page`/`limit`
pagination with `max_pages: 1`, records at `items`, `projection: "passthrough"`, and
`x-cursor-field: updated_at`.

The 5 added reporting streams are scoped under `/merchants/{merchant_id}/...` (urlencoded by
default per `InterpolatePath`); `merchant_id` is required.

- `transactions_search` (`GET .../reports/transactions/search.json`): the general transaction
  search/list endpoint. Paginated (`offset_limit`, `offset`/`limit` params, `page_size: 100`) —
  the only stream of the 5 that documents pagination at all. Optional `datetime_from`/
  `datetime_to` (config, `YYYY-MM-DD`) are sent via the opt-in optional-query dialect
  (`omit_when_absent: true`). Records are objects with `transaction`/`user`/`payment_details`/
  `purchase`/`payment_system` nested sub-objects (Xsolla's own documented response shape); a
  `computed_fields` pair flattens `transaction.id`→`transaction_id` and
  `transaction.create_date`→`transaction_create_date` as the schema's primary key / cursor field,
  since neither is a top-level response field on its own.
- `transactions_registry` (`GET .../reports/transactions/registry.json`): a richer per-transaction
  registry view (adds `user_balance`, drops `payment_details`/`payment_system`), always sent with
  the literal `in_transfer_currency=0` query param (Xsolla requires the source-currency, not
  transfer-currency, amounts for this bundle's purposes). Not documented as paginated by Xsolla;
  `pagination: none`. Same `transaction.id`/`transaction.transfer_date` flattening pattern as
  `transactions_search`.
- `payouts` (`GET .../reports/transfers`): payout batch records (`payout`/`transfer` nested
  sub-objects, `rate`, `canceled`). `payout.id`/`payout.date` flattened to `payout_id`/
  `payout_date` for the primary key/cursor.
- `payout_currency_breakdown` (`GET .../reports/transactions/summary/transfer`): per-ISO-currency
  payout aggregate rows (`IsoCurrency` is already a flat top-level field — Xsolla's own PascalCase
  naming for this endpoint's response, reproduced verbatim per `passthrough` projection; no
  computed_fields renaming since it would silently diverge from the API's real wire field name).
  No natural cursor field (an aggregate summary row, not a timestamped event) — no
  `x-cursor-field` declared.
- `financial_reports` (`GET .../reports`): per-month/currency financial report metadata
  (`report_id`, `month`, `year`, `currency`, `agreement_document_id`) — already flat, no
  computed_fields needed. No natural cursor field (an already-closed monthly report, not an
  incrementally-updated stream).

All 5 streams declare `"projection": "passthrough"`: Xsolla's documented response objects are
deeply nested and vary in shape across the 5 endpoints, and this bundle's schemas model only the
primary/cursor-key-bearing fields plus each response's well-known top-level nested objects as a
documentation surface — passthrough guarantees every real API field survives to the emitted
record regardless of schema completeness (matching this bundle's pre-existing passthrough
precedent for the old 3-stream shape, and the general rule that an externally-owned, deeply-nested
JSON API response should never be schema-narrowed without a field-for-field verified mapping).

## Write actions & risks

2 write actions, both requiring approval (`capabilities.write: true`):

- `request_refund` (`PUT .../reports/transactions/{transaction_id}/refund`): issues a full refund
  to the user for the given transaction. Irreversible.
- `request_partial_refund` (`PUT .../reports/transactions/{transaction_id}/partial_refund`): issues
  a partial refund for a specific `refund_amount`. Irreversible.

Both require `description` (Xsolla's own required refund-reason field); `request_refund` also
accepts an optional `email` override, `request_partial_refund` requires `refund_amount`.

## Known limits

- `financial_reports` requires its `datetime_from`/`datetime_to` window to be 92 days or less per
  Xsolla's documented constraint on the reports endpoint; not enforced client-side (a request
  outside that window is rejected by the live API itself, surfaced as an ordinary HTTP error
  through `error_map`).
- `transactions_search`'s real API also supports a `simple_search` fast-lookup variant scoped to a
  single known `transaction_id`/`external_id`, and a `/{transaction_id}/details` single-transaction
  detail endpoint — both excluded as `duplicate_of` (`api_surface.json`): the same transaction
  record shape is already reachable via this stream's list output, and `transaction_id` is already
  a field on every emitted record.
- Token/tokenization endpoints (payment-UI session generation, session expiry, saved
  payment-account listing/charge/delete) are excluded as `out_of_scope`/`requires_elevated_scope`:
  they operate on live checkout sessions and per-user saved payment methods, not syncable business
  records or dialect-expressible data mutations — see `api_surface.json`.
- The sandbox-only chargeback-simulation endpoint is excluded as `destructive_admin`: test-only,
  requires an elevated Publisher Account role, not a production data mutation.
- No `max_pages` config is modeled on `transactions_search`; pagination runs to exhaustion (the
  short-page stop signal) with a fixed `page_size: 100`.
