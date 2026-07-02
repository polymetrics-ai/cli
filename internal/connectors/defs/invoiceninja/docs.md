# Overview

Invoice Ninja is invoicing/billing software. This bundle reads Invoice Ninja clients, invoices,
products, payments, and quotes through the Invoice Ninja v5 REST API
(`https://invoicing.co/api/v1`). It migrates `internal/connectors/invoiceninja` (the legacy
hand-written connector, kept registered and unchanged until wave6's registry flip).

## Auth setup

Provide an Invoice Ninja API token via the `api_key` secret; it is sent as the `X-API-TOKEN`
request header, matching legacy's `connsdk.APIKeyHeader("X-API-TOKEN", secret, "")` exactly. Never
logged. `base_url` defaults to the hosted `invoicing.co` instance but may be overridden for
self-hosted Invoice Ninja installs.

## Streams notes

All 5 streams (`clients`, `invoices`, `products`, `payments`, `quotes`) share the same shape: `GET`
against the Invoice Ninja list endpoint, records at `data` (the API wraps lists in
`{"data":[...],"meta":{...}}`), primary key `["id"]` (a string id, not numeric — matches legacy's
`Type: "string"` field declarations). Pagination is `page_number` (`page`/`per_page` query params,
1-based `start_page`, `page_size: 100`, matching legacy's default `pageSize` of 100) — a page
shorter than `per_page` is the last page. No stream declares an incremental cursor field: legacy's
`invoiceNinjaStreams()` sets no `CursorFields` for any stream (genuinely full-refresh only, unlike
sibling connector `invoiced` which at least catalogs an unused `updated_at` cursor), so this bundle
matches that exactly.

## Write actions & risks

None. Invoice Ninja is exposed as a read-only source (legacy's `Write` always returns
`connectors.ErrUnsupportedOperation` with every record counted failed); no `writes.json` is
declared.

## Known limits

- Full Invoice Ninja API surface (invoice emailing/reminders, recurring invoices, tasks, expenses,
  vendor management, etc.) is out of scope for this wave; see `api_surface.json`'s `excluded`
  entries.
- `page_size`'s legacy upper bound is 1000 (`invoiceNinjaMaxPageSize`), higher than every sibling
  connector in this wave (100); `spec.json`'s `page_size` description documents the 1-1000 range
  matching legacy's own validation bound.
