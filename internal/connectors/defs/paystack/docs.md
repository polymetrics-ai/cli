# Overview

Paystack is a read-only declarative-HTTP connector for the Paystack REST API. It reads customers,
transactions, subscriptions, invoices (payment requests), and disputes. This bundle migrates
`internal/connectors/paystack` (the hand-written connector); the legacy package stays registered
and unchanged until wave6's registry flip.

## Auth setup

Provide a Paystack secret key (`sk_...`) via the `secret_key` secret; it is used only for Bearer
auth (`Authorization: Bearer <secret_key>`) and is never logged.

## Streams notes

All 5 streams (`customers`, `transactions`, `subscriptions`, `invoices`, `disputes`) share the same
shape: `GET` against the Paystack list endpoint (`/customer`, `/transaction`, `/subscription`,
`/paymentrequest`, `/dispute`), records at `data`, primary key `["id"]` (a bare JSON integer,
matching Paystack's real wire shape — `id` and `createdAt`-derived cursors keep their real types,
not stringified), incremental cursor field `createdAt`. Pagination follows Paystack's
`meta.next`-page-number convention (`pagination.type: cursor` with `cursor_param: page` and
`token_path: meta.next`): the next request's `page` query param is read verbatim from the response
body's `meta.next` field (an integer page number, or `null` on the last page — `null` resolves to
an empty token via the engine's `StringAt` extraction, which stops pagination exactly like
legacy's `parseNextPage("null") -> false`). Every request sends the configured `perPage`
(`config.page_size`, default `100`, matching legacy's `paystackDefaultPageSize`/
`paystackMaxPageSize`, both 100). Incremental reads send a `from` query param carrying the RFC3339
lower bound (persisted cursor, or the `start_date` config on a fresh sync) via the engine's
`stream.Query` optional-query dialect (`omit_when_absent: true`) — present only when a lower bound
resolves, matching legacy's `incrementalLowerBound` exactly (legacy sends `from` verbatim as
RFC3339, no unit conversion, so no `param_format` override is declared).

## Write actions & risks

Not applicable — this connector is read-only (`capabilities.write: false`), matching legacy
exactly (legacy's own doc comment: "the API has no obviously-safe reverse-ETL write actions for the
core streams").

## Known limits

- Full Paystack API surface (transfers, refunds, plans, products, settlements) is out of scope for
  this wave; see `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass B
  capability expansion"}` entries.
- Documented parity deviation: legacy's `harvest` loop falls back to a short-page stop check
  (`len(records) < pageSize`) ONLY when `meta.next` is absent/unparseable — a defensive path for a
  malformed or undocumented response shape. This bundle's `cursor`+`token_path` paginator relies
  solely on `meta.next` (present on every real Paystack list response per its documented API
  contract, and on every fixture page in this bundle) and does not reproduce the short-page
  fallback, since the declarative pagination dialect's `token_path` variant has no secondary
  short-page stop condition. This never changes emitted data for any input Paystack's documented,
  well-formed API actually returns (the two stop conditions coincide for every conforming
  response); it would only diverge if Paystack's real API ever omitted `meta.next` on a full page,
  which its documented contract does not allow. See `docs/migration/conventions.md`'s
  parity-deviation ledger.
