# Overview

Invoiced is billing/invoicing software. This bundle reads Invoiced customers, invoices, payments,
subscriptions, and estimates through the Invoiced REST API (`https://api.invoiced.com`). It
migrates `internal/connectors/invoiced` (the legacy hand-written connector, kept registered and
unchanged until wave6's registry flip).

## Auth setup

Provide an Invoiced API key via the `api_key` secret; it is sent as the HTTP Basic username with a
blank password (`Authorization: Basic base64(api_key:)`), matching legacy's
`connsdk.Basic(secret, "")` exactly. Never logged.

## Streams notes

All 5 streams (`customers`, `invoices`, `payments`, `subscriptions`, `estimates`) share the same
shape: `GET` against the Invoiced list endpoint, records read directly off the response's top-level
JSON array (`records.path: ""`), primary key `["id"]`. Pagination is `page_number`
(`page`/`per_page` query params, 1-based `start_page`, `page_size: 100`, matching legacy's default
`pageSize`/`maxPageSize` of 100) — a page shorter than `per_page` is the last page.

Every stream's schema declares `x-cursor-field: updated_at` (every Invoiced object exposes a Unix
`updated_at` timestamp, matching legacy's `invoicedStreams()` `CursorFields: []string{"updated_at"}`
declaration) with `incremental.client_filtered: true` and no server-side request param — legacy
never sends any date-range filter to the Invoiced API (its `Read` always issues the same
page/per_page-only request regardless of any persisted cursor), so this bundle does not invent one
either; the engine performs the identical full page walk and then drops already-seen records
client-side by comparing `updated_at` against the incremental lower bound, an additive
behind-the-scenes optimization that changes no accepted-input behavior and produces the exact same
emitted records for a fresh full sync.

## Write actions & risks

None. Invoiced is a read-only source connector (legacy's `Write` always returns
`connectors.ErrUnsupportedOperation`); no `writes.json` is declared.

## Known limits

- Full Invoiced API surface (customer/invoice mutations, payment plans, estimate approval actions,
  etc.) is out of scope for this wave; see `api_surface.json`'s `excluded` entries.
- Legacy's fixture-mode `readFixture` stamps a `previous_cursor` field onto fixture-mode records
  when `req.State["cursor"]` is set — this is fixture/test-harness-only behavior with no live-API
  equivalent and is not modeled here (the engine's fixture replay is a distinct, declarative
  mechanism; this bundle's `fixtures/streams/**` are conformance fixtures, not a reproduction of
  legacy's in-code fixture generator).
