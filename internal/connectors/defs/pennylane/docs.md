# Overview

Pennylane is a read-only declarative migration of `internal/connectors/pennylane` (legacy Go
connector). It reads customers, customer invoices, suppliers, products, and categories from
Pennylane's External API v2. This bundle is capability-parity with legacy; legacy stays registered
and unchanged until wave6's registry flip.

## Auth setup

Provide a Pennylane API key via the `api_key` secret; it is used only for Bearer auth
(`Authorization: Bearer <api_key>`) and is never logged. Legacy hard-errors when `api_key` is
unset (`pennylane connector requires secret api_key`), matching this bundle's `required:
["api_key"]`.

## Streams notes

All 5 streams (`customers`, `customer_invoices`, `suppliers`, `products`, `categories`) share the
identical shape: `GET` against the Pennylane v2 list endpoint, records at `items`, primary key
`["id"]`, `x-cursor-field: updated_at` (catalog metadata only — see Known limits). Pagination
follows legacy's `connsdk.CursorPaginator{CursorParam: "cursor", TokenPath: "next_cursor"}` shape
exactly: `pagination.type: cursor` with `cursor_param: cursor` and `token_path: next_cursor`, no
`stop_path` (legacy stops purely on an absent/empty `next_cursor`, with no separate boolean stop
flag). `limit` is sent on every request from `page_size` (default `50`, matching legacy's
`defaultPageSize`).

All 5 streams declare `"projection": "passthrough"` (post-wave2 review §8 rule 1): legacy's `Read`
emits `emit(connectors.Record(rec))` — a verbatim type-cast of the raw harvested record, with no
`mapRecord`-style field-building — so schema-mode projection would silently drop any raw field this
bundle's schema omits. The schema (`id`, `name`, `created_at`, `updated_at`) remains a documentation
surface only; it does not gate what is emitted.

Legacy also forwards two optional, verbatim passthrough config values as query params whenever
set: `filter` and `sort` (`if filter := ...; filter != "" { base.Set("filter", filter) }` /
same for `sort`). This bundle wires both through the engine's opt-in optional-query dialect
(`query.filter`/`query.sort` object-form entries with `omit_when_absent: true`) — present only
when the corresponding config value is set, absent entirely otherwise, matching legacy's
conditional `url.Values.Set` calls exactly.

## Write actions & risks

None. Legacy `Write` always returns `connectors.ErrUnsupportedOperation`; `metadata.json` declares
`capabilities.write: false` and no `writes.json` file exists, matching legacy exactly (legacy's own
package comment notes writes are intentionally unexposed because Pennylane's write surface mutates
accounting data).

## Known limits

- `x-cursor-field: updated_at` is declared on every schema as catalog/candidate-cursor metadata
  only, mirroring legacy's own `CursorFields: []string{"updated_at"}` declaration — legacy's `Read`
  never applies any server-side or client-side incremental filter using this field (no
  `updated_at[gte]`-style query param, no client-side filtering by cursor value), so no
  `streams.json` `incremental` block is declared here either. This is exact parity, not a
  narrowing.
- `page_size`/`max_pages` config validation (legacy's numeric-range and `all`/`unlimited` keyword
  parsing) is not reproduced at the bundle-config level; the engine treats `page_size` as an opaque
  string substituted directly into the `limit` query param. This never changes emitted record DATA
  for any legacy-valid input; it only narrows client-side input validation, out of scope for wave2
  fan-out (Pass B).
