# Overview

RevenueCat is a wave2 fan-out declarative-HTTP migration. It reads RevenueCat projects, apps,
products, offerings, and customers through the REST API v2
(`GET https://api.revenuecat.com/v2/...`). This bundle targets capability parity with
`internal/connectors/revenuecat` (the hand-written connector it migrates); the legacy package
stays registered and unchanged until wave6's registry flip. Read-only (`capabilities.write` is
`false`, matching legacy's `Write` returning `connectors.ErrUnsupportedOperation`).

## Auth setup

Provide a RevenueCat secret API key via the `api_key` secret. It is sent as an
`Authorization: Bearer <api_key>` header, matching legacy's `connsdk.Bearer(key)`
(`revenuecat.go:179`). `base_url` defaults to `https://api.revenuecat.com/v2` and may be overridden
for tests/proxies.

## Streams notes

`projects` is the one project-independent stream (`GET /projects`); `apps`, `products`,
`offerings`, and `customers` are all scoped to a single project via a path-parameterized
`project_id` (`GET /projects/{{ config.project_id }}/apps` etc.), matching legacy's
`projectPath(resource)` helper (`revenuecat.go:186-194`) which hard-errors with
`"revenuecat stream %s requires config project_id"` when `project_id` is unset — the engine's
own path interpolation reproduces this exact requirement: an absent `config.project_id` is a hard
error at path-resolution time for these 4 streams, while `projects` never references it at all
(matching legacy's per-stream, not connection-wide, requirement). `project_id` is therefore
declared in `spec.json` but NOT in `required[]`.

All 5 streams share the same records envelope: RevenueCat v2's real wire shape is
`{"object": "list", "items": [...], "next_page": ...}` (confirmed against RevenueCat's published
v2 API reference), and `records.path: "items"` matches legacy's own `recordsPath: "items"`
declaration for every endpoint (`revenuecat.go:114-119`) exactly — the primary candidate legacy's
`recordsAt` fallback list tries first. Pagination is `page_number` (`page`/`limit`,
`page_size: 100`), stopping on a short page exactly as legacy's `connsdk.PageNumberPaginator`
does; this reproduces the effective behavior of RevenueCat's own `next_page`-based cursor even
though legacy itself never reads `next_page` (legacy always drives its own `page`/`limit` query
params, never following the returned `next_page` URL) — matching legacy exactly, not RevenueCat's
documented cursor convention.

Legacy applies three passthrough filters (`starting_after`, `created_after`, `updated_after`)
identically to every stream's request (`revenuecat.go:92-95`'s loop iterates a fixed key list
regardless of which stream is being read) — this bundle reproduces that exact blanket behavior via
`base.query`'s three `omit_when_absent` entries (shared across all streams), sent only when the
corresponding config value is set. `computed_fields` stamps a static `stream` marker on every
record, matching legacy's `mapRecord`'s `out["stream"] = stream`.

`updated_at` is declared as `x-cursor-field` only on `customers` (matching legacy's own
`CursorFields: []string{"updated_at"}`, declared only for that stream; the other 4 streams have no
`CursorFields` in legacy's `Catalog`). No `incremental` block is declared on any stream: legacy's
`Read` never reads a persisted sync cursor back into `updated_after`/`created_after`/
`starting_after` (`harvest` reads only `req.Config.Config[key]`, never `req.State["cursor"]`) — it
always resends the exact same raw config value on every sync, with no forward advancement.

## Write actions & risks

None. Legacy `revenuecat.go`'s `Write` returns `connectors.ErrUnsupportedOperation`
unconditionally; `capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes both as config-driven
  overrides (`page_size` bounded 1-100, default 100; `max_pages` 0/all/unlimited for unbounded).
  The engine's `page_number` paginator reads `PaginationSpec.PageSize`/`MaxPages` as static
  bundle-authored integers, not config templates — there is no mechanism to wire a `spec.json`
  property into either field. This bundle sends `page_size: 100` (legacy's own default) as a
  static value in `streams.json`'s `base.pagination` block; neither `page_size` nor `max_pages` is
  declared in `spec.json` (F6: dead config is worse than absent config). Pagination is otherwise
  unbounded (matches legacy's `max_pages: 0` = unlimited default) other than the short-page stop
  signal.
- **Legacy's `id` fallback (`uuid`/`app_user_id`/`identifier`/`name`) is not modeled.** Legacy's
  `mapRecord` falls back to a record's `uuid`, `app_user_id`, `identifier`, or `name` field when
  `id` is absent. Every RevenueCat v2 resource this bundle reads always carries an `id` in its real
  wire shape (confirmed against RevenueCat's published v2 API reference, and legacy's own
  `Catalog`/`PrimaryKey` declarations assume `id` unconditionally for all 5 streams), so this
  fallback is defensive dead code against the real API — not exercised by any input legacy itself
  would realistically receive. Documented here for completeness, not implemented via a hook.
- **RevenueCat's own `next_page` cursor-URL pagination convention is not modeled.** RevenueCat's v2
  API documents a `next_page` absolute-URL-based pagination scheme (in addition to
  `limit`/`starting_after`), but legacy itself never reads or follows `next_page` at all — it
  drives pagination purely with its own `page`/`limit` query params via
  `connsdk.PageNumberPaginator`. This bundle reproduces legacy's actual behavior (page-number
  pagination), not RevenueCat's documented alternative, per the meta-rule that legacy is ground
  truth over any doc.
- The full RevenueCat API surface (purchase receipt recording, subscriber attribute/attribution
  updates, promotional entitlement grants, purchase revoke/refund/cancel/defer/extend mutations)
  is out of scope for this wave; see `api_surface.json`'s `excluded: {category: out_of_scope}`
  entries.
