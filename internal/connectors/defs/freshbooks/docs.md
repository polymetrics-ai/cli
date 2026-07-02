# Overview

FreshBooks is an accounting platform. This bundle reads FreshBooks clients, invoices, expenses,
payments, and items through the FreshBooks accounting REST API (`https://api.freshbooks.com`). It
is read-only, matching legacy `internal/connectors/freshbooks` exactly (no reverse-ETL writes).

## Auth setup

Provide a FreshBooks OAuth2 access token via the `oauth_access_token` secret; it is sent as
`Authorization: Bearer <oauth_access_token>` and is never logged. The refresh token / client id /
client secret used for the broader OAuth dance are not consumed directly by this connector, exactly
as in legacy. A FreshBooks `account_id` config value is required — every accounting list endpoint
is scoped under `/accounting/account/{account_id}/`.

## Streams notes

All 5 streams (`clients`, `invoices`, `expenses`, `payments`, `items`) share the same shape: `GET`
against `/accounting/account/{account_id}/<resource>`, records at
`response.result.<array_key>` (e.g. `response.result.clients`), primary key `["id"]`, and the
`x-cursor-field` set to `updated` (matching legacy's `CursorFields: ["updated"]` on every stream) —
no `incremental` request param is declared because legacy itself applies none server-side for any
FreshBooks stream (full refresh only; `x-cursor-field` here exists solely for
`incremental_append_deduped` sync-mode eligibility, exactly mirroring legacy's own stream catalog
declaration without any actual server-side filter).

Pagination is `page_number` (`page`/`per_page`, `start_page: 1`, `page_size: 100`), stopping on a
short page. Legacy's `harvest` loop primarily stops using the response's own authoritative
`response.result.pages` count (`page >= total`), falling back to short-page detection only when
that count is unparseable. The engine's `page_number` paginator stops purely via short-page
detection (`recordCount < page_size`). **Documented parity deviation**: for the extremely unusual
case of a final page that is exactly full-sized (`total` records is an exact multiple of
`page_size`) AND `page >= total` pages, legacy would stop one request earlier via the authoritative
`pages` count while this bundle would issue one additional request that returns 0 records before
stopping — no records are ever duplicated, dropped, or reordered by this difference; it is a
request-count-only divergence, never a data divergence, so it is scoped ACCEPTABLE per this file's
meta-rule. See `docs/migration/conventions.md` §5.

## Write actions & risks

None. FreshBooks is exposed read-only, matching legacy's `Capabilities{Write: false}` and its
`Write` method, which always returns `connectors.ErrUnsupportedOperation`.

## Known limits

- Full FreshBooks API surface (estimates, time entries, projects, staff, bills, journal entries,
  webhooks, etc.) is out of scope for this wave; see `api_surface.json`'s
  `excluded: {category: out_of_scope, reason: "Pass B capability expansion"}` entries. Only the 5
  legacy-parity read streams are implemented.
- The page-count stop-signal deviation described above under Streams notes (request-count only,
  never a data divergence).
- `account_id` is required config (not optional), matching legacy's own hard requirement
  (`freshbooksAccountID` errors when unset).
