# Overview

Dolibarr was quarantined in wave1 for an `ENGINE_GAP`: Dolibarr's REST API is genuinely 0-indexed
(legacy sends `page=0` for the first page, `page=1` for the second, per `dolibarr.go`'s `for page :=
0; ...` harvest loop), and the engine's `page_number` pagination could not express a 0-indexed
start (a plain Go `int` `StartPage` field could not distinguish an explicit `0` from an omitted
key). This gap was closed by the S4 engine mini-wave's `PaginationSpec.StartPage *int`
(`"start_page": 0` is now distinguishable and honored verbatim) — this bundle is the unblock build
using that dialect addition. It reads Dolibarr third parties, contacts, products, customer
invoices, and orders through the Dolibarr REST API. This bundle migrates
`internal/connectors/dolibarr` (the hand-written connector it replaces at capability parity); the
legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Dolibarr API key via the `api_key` secret; it is sent as the `DOLAPIKEY` request header
(`api_key_header` auth mode) and is never logged.

## Streams notes

All 5 streams (`thirdparties`, `contacts`, `products`, `invoices`, `orders`) share the identical
shape: `GET` against the Dolibarr list endpoint, records at the response body root (`records.path:
""`, matching legacy's top-level-JSON-array `connsdk.RecordsAt(resp.Body, "")`), and every request
sends `sortfield=t.rowid&sortorder=ASC` as static query params (matching legacy's `harvest` query
construction exactly). Pagination is genuinely 0-indexed (`pagination.type: page_number`,
`page_param: page`, `size_param: limit`, `start_page: 0`, `page_size: 100` matching legacy's
`dolibarrDefaultPageSize`) — the first request sends `page=0`, matching legacy's loop exactly; a
page returning fewer than `limit` records stops the read (legacy's `len(records) < pageSize`
short-page stop).

None of the 5 streams declares an `incremental` block: legacy's `harvest` sends no server-side
filter parameter derived from a cursor or `start_date`-shaped config value at all (only the static
`sortfield`/`sortorder`/`limit`/`page` params above) — per conventions.md §8 rule 2, an
`incremental` block is only declared when legacy actually sends a server-side filter, which it does
not here. `x-cursor-field: date_modification` is still declared on every schema (matching legacy's
published `CursorFields`) for catalog/sync-mode-derivation parity even though no request-time
filtering happens.

## Write actions & risks

None. Dolibarr is exposed read-only here (legacy's `Write` returns
`connectors.ErrUnsupportedOperation`); `capabilities.write` is `false` and this bundle ships no
`writes.json`.

## Known limits

- **Dolibarr's 404-as-empty-page end-of-data signal is approximated, not reproduced 1:1.** Legacy
  treats an HTTP 404 ("No record found") past the end of the data set as a clean end-of-data signal
  identical to a short/empty 200 page (`isNotFound(err)` inside `harvest`, and again in `Check`).
  The engine's `page_number` paginator's only stop signal is a short/empty successful page; it does
  not special-case a 404 response as "clean end of data" — a 404 propagates as a request error
  instead. In practice this only diverges from legacy when the LAST page happens to be exactly a
  multiple of `page_size` (so the very next page is empty AND some deployments return 404 rather
  than `200 []` for it); every deployment/page shape observed in fixtures and the parity-relevant
  legacy tests returns a `200 []`-style empty/short page, which both sides handle identically. Not
  modeled as an engine change since it recurs only for this one connector (below the §6 recurrence
  threshold) and true production Dolibarr instances vary in whether they emit 404 vs. empty-200 for
  an out-of-range page.
- **Base URL is `base_url`-only; the legacy `my_dolibarr_domain_url` bare-domain convenience is
  dropped.** Legacy accepts either an explicit `base_url` override OR a bare
  `my_dolibarr_domain_url` (e.g. `"mydomain.com/dolibarr"`), deriving
  `https://<domain>/api/index.php` from the latter in Go (`domainToBaseURL`). The engine's
  `spec.json` `"default"` materialization only fills in a FIXED literal default, not a
  config-value-derived one (conventions.md §3's `spec.json "default"` paragraph — "for a DERIVED
  default ... this mechanism alone is not enough; either require `base_url` and drop the derivation
  ... or express the derivation as a `computed_fields`-style template if/when the dialect grows
  one for base-URL construction"); no such mechanism exists yet. This bundle requires the full
  `base_url` (e.g. `https://your-dolibarr-host/api/index.php`) and does not declare
  `my_dolibarr_domain_url` at all (a declared-but-unwireable key is worse than an absent one, per
  conventions.md F6). Documented config-surface narrowing, not a silent behavior change for any
  input this bundle itself accepts.
- Full Dolibarr API surface (users, projects, warehouses/stock, bank accounts, expense reports, and
  any write/mutation endpoints) is out of scope; see `api_surface.json`'s `excluded` entries. Only
  the 5 legacy-parity read streams are implemented.
