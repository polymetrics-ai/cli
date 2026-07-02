# Overview

Judge.me Reviews reads product reviews, products, and review widgets for a Shopify shop through
the Judge.me REST API (`https://judge.me/api/v1`). This bundle migrates the legacy
`internal/connectors/judge-me-reviews` package to the declarative engine at capability parity; the
legacy package stays registered and unchanged until wave6's registry flip. The connector is
read-only (Judge.me has no reverse-ETL write surface implemented here).

## Auth setup

Provide a Judge.me API token via the `api_key` secret; it is sent as the `api_token` query
parameter on every request (`auth: [{mode: api_key_query, param: api_token}]`) and is never
logged. `shop_domain` (the `*.myshopify.com` store domain) is a required config value, sent as the
`shop_domain` query parameter on every stream request via `stream.query` — Judge.me's API scopes
every list endpoint to a single shop this way, matching legacy's `queryAuth.Apply` (which sets both
`api_token` and `shop_domain` on every outgoing request).

## Streams notes

All 3 streams (`reviews`, `products`, `widgets`) share the same shape: `GET` against the Judge.me
list endpoint, records at the top-level key matching the resource name (`reviews`/`products`/
`widgets`), primary key `["id"]`, incremental cursor field `created_at`. Pagination is 1-based
`page_number` (`page_param: page`, `start_page: 1`); the request-size query param (`per_page`) is
sent via each stream's `query.per_page` (`config.page_size`, default `"100"`) rather than the
pagination block's `size_param` (left empty) — this avoids double-declaring the size param once via
`stream.query` (config-driven) and once via a fixed `pagination.size_param`+`page_size`, which would
conflict since only one can carry the caller's override.

`reviews` flattens Judge.me's nested `reviewer` object (`reviewer.id`/`reviewer.name`/
`reviewer.email`) onto the record via `computed_fields`, exactly matching legacy
`reviewRecord`'s manual flattening.

## Write actions & risks

None. Judge.me Reviews is read-only in both legacy and this bundle (`capabilities.write: false`,
no `writes.json`).

## Known limits

- **Pagination stop-threshold parity narrowing (ACCEPTABLE, documented)**: legacy's `page_size`
  config (1-100, default 100) drives BOTH the `per_page` request query param AND the short-page
  stop check (`len(records) < pageSize`). The engine's `page_number` paginator's own stop-threshold
  (`pagination.page_size`) is a fixed literal in `streams.json` (set to `100`, the legacy default)
  and cannot be wired to the same runtime `config.page_size` value that drives the request query
  param (`PaginationSpec.PageSize` is a plain JSON int, not a template) — the same class of gap
  documented for `page_size`/`max_pages` in `docs/migration/conventions.md`'s searxng entry. This
  bundle still wires `config.page_size` into the actual `per_page` request parameter (so callers CAN
  request a different page size), but a caller who overrides `page_size` away from `100` would see
  the engine's short-page-detection threshold diverge from the actual requested size. Never wrong
  for the common (default `page_size=100`) case; only imprecise for a non-default override, which
  legacy itself never exercised in its own tests either.
- **`max_pages` config (legacy: default 1000, `0`/`all`/`unlimited` = unbounded) is not
  runtime-wireable**: `PaginationSpec.MaxPages` is a fixed literal, so this bundle hardcodes
  `max_pages: 1000` (legacy's own default) in `streams.json`'s `base.pagination` and does not
  declare a `max_pages` spec property at all (a declared-but-unwireable key is worse than an absent
  one — F6, `conventions.md`). A sync that legitimately needs more than 1000 pages of a single
  stream is out of scope for this bundle; this matches legacy's own default bound for the vast
  majority of real shops.
- Full Judge.me API surface (creating/updating reviews, other endpoints) is out of scope; see
  `api_surface.json`'s `excluded: {category: out_of_scope}` entries. Only the 3 legacy-parity read
  streams are implemented.
- Fixtures for each stream ship a 100-record page 1 (matching the fixed `pagination.page_size: 100`
  short-page threshold) plus a 1-record page 2, proving the engine issues exactly 2 requests and
  terminates on the short second page.
