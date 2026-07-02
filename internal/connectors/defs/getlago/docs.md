# Overview

Lago (getlago.com) is a read-only declarative-HTTP connector migrated from
`internal/connectors/getlago` (legacy wave2 fan-out). It reads Lago customers, invoices,
subscriptions, plans, and billable metrics through the Lago REST API. This bundle is
capability-parity with the legacy hand-written connector; the legacy package stays registered and
unchanged until wave6's registry flip.

## Auth setup

Provide a Lago API key via the `api_key` secret; it is used only for Bearer auth
(`Authorization: Bearer <api_key>`) and is never logged. `api_url` defaults to
`https://api.getlago.com/api/v1` and may be overridden for self-hosted Lago instances or test
proxies.

## Streams notes

All 5 streams (`customers`, `invoices`, `subscriptions`, `plans`, `billable_metrics`) share the
same shape: `GET` against the Lago list endpoint, records nested under a resource-named JSON key
(e.g. `{"customers": [...], "meta": {...}}`), primary key `["lago_id"]`. Pagination is
`page_number` (`page`/`per_page` query params, default page size 100) — Lago's real stop signal is
`meta.next_page` going `null`, but the engine's `page_number` paginator's short-page stop rule
(fewer than `page_size` records returned) fires at the exact same point in every legacy response
shape this bundle's fixtures model (a `total_pages`-bounded list never returns a full final page),
so no behavior is lost porting to the declarative paginator.

Every stream's schema declares `x-cursor-field: created_at` (matching legacy's own catalog
`CursorFields` entry for every stream), but legacy's `Read` path never actually sends any
incremental filter parameter, nor filters client-side, for ANY stream (the only place
`req.State["cursor"]` is consulted is fixture-mode's `previous_cursor` debug annotation). Declaring
an `incremental` block here would change accepted behavior (the engine would then compute and
possibly send a lower-bound parameter legacy never sent), so this bundle matches legacy's real
behavior: no stream declares an `incremental` block — every stream is a full-refresh read, exactly
like legacy.

Legacy accepts a config `base_url` as a secondary alias for `api_url` (`api_url` is checked first;
`base_url` is a fallback when `api_url` is unset). The engine's templating dialect resolves exactly
one config key per field with no fallback-chain primitive, so this bundle declares only `api_url`
(legacy's primary/first-checked key) in `spec.json`; the `base_url` alias name is dropped. This
never changes accepted behavior for any caller already using the primary `api_url` key — it only
removes a secondary, less-preferred alias name. See Known limits.

## Write actions & risks

None. Lago's source connector is read-only (full refresh only) in legacy (`capabilities.write:
false`, matching exactly); there is no `writes.json`.

## Known limits

- Only the 5 legacy-parity read streams are implemented; Lago's write endpoints (creating
  customers/subscriptions/wallets, applying add-ons, refreshing invoices) are out of scope for this
  migration wave — see `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass B
  capability expansion"}` entries.
- Every stream's `created_at` cursor field is catalog-informational only, matching legacy's own
  behavior of declaring a `CursorFields` entry without ever filtering by it (see Streams notes
  above) — this is a faithful port, not a scope narrowing.
- The legacy `base_url` config alias (secondary fallback for `api_url`) is dropped; only `api_url`
  is declared. ACCEPTABLE per the parity-deviation meta-rule: never changes behavior for any caller
  using the primary key, only removes an alternate alias name for the same base-URL override.
