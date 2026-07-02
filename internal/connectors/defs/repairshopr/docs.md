# Overview

RepairShopr is a wave2 fan-out declarative-HTTP migration. It reads RepairShopr customers,
tickets, invoices, estimates, and customer assets through the RepairShopr REST API v1
(`GET https://<subdomain>.repairshopr.com/api/v1/...`). This bundle is migrated from
`internal/connectors/repairshopr` (the hand-written connector it replaces); the legacy package
stays registered and unchanged until wave6's registry flip. Read-only (`capabilities.write` is
`false`, matching legacy's `Write` returning `connectors.ErrUnsupportedOperation`).

## Auth setup

Provide a RepairShopr API token via the `api_token` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_token>`) and is never logged, matching legacy's
`connsdk.Bearer(token)`. `base_url` is required (see Known limits below for the
subdomain-derivation narrowing).

## Streams notes

All 5 streams (`customers`, `tickets`, `invoices`, `estimates`, `assets`) share the same shape:
`GET` against the RepairShopr list endpoint (`/customers`, `/tickets`, `/invoices`, `/estimates`,
`/customer_assets`), records at the stream's own top-level key (`customers`/`tickets`/`invoices`/
`estimates`/`assets`), primary key `["id"]`. Pagination is `page_number` (`page`/`per_page`,
`page_size: 100`), stopping on a short page exactly as legacy's `connsdk.PageNumberPaginator`
does. Every stream optionally forwards three passthrough filters — `created_after`,
`updated_after`, `query` — as query params only when the corresponding config value is set
(`omit_when_absent`), matching legacy's own `strings.TrimSpace(...) != ""` gate before adding
each to the base `url.Values{}`. `computed_fields` stamps a static `stream` marker on every
record (`"customers"`/`"tickets"`/etc.), matching legacy's `mapRecord`'s `out["stream"] = stream`.

`updated_at` is declared as `x-cursor-field` on every schema, matching legacy's own
`CursorFields: []string{"updated_at"}` Catalog declaration. No `incremental` block is declared:
legacy's `Read` never actually uses a persisted sync cursor to filter requests server-side (the
three passthrough filters above are static per-run config values, not a computed incremental
lower bound) — declaring an `incremental` block here would introduce new, behavior-changing
state-driven filtering legacy never had. Full refresh (and `_deduped` sync modes, since
`x-primary-key` is present) are what this bundle actually supports, matching legacy exactly.

## Write actions & risks

None. Legacy `repairshopr.go`'s `Write` returns `connectors.ErrUnsupportedOperation`
unconditionally; `capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- **Subdomain-derived `base_url` is not modeled.** Legacy accepts either an explicit `base_url`
  config value or a bare `subdomain`, deriving `https://<subdomain>.repairshopr.com/api/v1` in
  code when only `subdomain` is set. The engine's `spec.json` `"default"` materialization
  mechanism only supports a FIXED literal default, not one derived from another config value's
  runtime content (conventions.md §3's derived-default note) — expressing this would need either
  a new engine templating primitive for `base_url` construction or a Tier-2 hook, neither of
  which is warranted for a single derived-URL convenience. This bundle therefore requires
  `base_url` directly; `subdomain` is not declared in `spec.json` at all (a declared-but-unwireable
  key would be worse than an absent one, per F6/REVIEW.md).
- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes both as config-driven
  overrides (`page_size` bounded 1-100, default 100; `max_pages` 0/all/unlimited for unbounded).
  The engine's `page_number` paginator reads `PaginationSpec.PageSize`/`MaxPages` as static
  bundle-authored integers, not config templates — there is no mechanism to wire a `spec.json`
  property into either field. This bundle sends `page_size: 100` (legacy's own default) as a
  static value in `streams.json`'s `base.pagination` block; neither `page_size` nor `max_pages`
  is declared in `spec.json` (F6: dead config is worse than absent config). Pagination is
  otherwise unbounded (matches legacy's `max_pages: 0` = unlimited default) other than the
  short-page stop signal.
- **Legacy's `id` fallback (`uuid`/`number`) is not modeled.** Legacy's `mapRecord` falls back to
  a record's `uuid` or `number` field when `id` is absent. Every RepairShopr resource this bundle
  reads always carries a numeric `id` in its real wire shape (legacy's own `Catalog`/`PrimaryKey`
  declarations assume `id` unconditionally for all 5 streams), so this fallback is defensive dead
  code against the real API — not exercised by any input legacy itself would realistically
  receive. Documented here for completeness, not implemented via a hook.
- The full RepairShopr API surface (line items, payments, appointments, users, etc.) is out of
  scope for this wave; see `api_surface.json`'s `excluded: {category: out_of_scope}` entries.
