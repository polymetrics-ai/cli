# Overview

Freshsales (Freshworks CRM) is a wave2 fan-out declarative-HTTP migration. It reads Freshsales
contacts, sales accounts, deals, and leads through the Freshsales REST API
(`GET https://<domain>/crm/sales/api/<resource>/view/<view_id>`). This bundle is engine-vs-legacy
capability-parity migrated from `internal/connectors/freshsales` (the hand-written connector it
migrates); the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Freshsales API key via the `api_key` secret; it is sent as
`Authorization: Token token=<api_key>` (an `api_key_header` auth candidate with a fixed `Token
token=` prefix and the secret as the templated value) and is never logged, matching legacy's
`connsdk.APIKeyHeader("Authorization", secret, "Token token=")` (`freshsales.go:264`).
`domain_name` (e.g. `mydomain.myfreshworks.com`) is required and combined with the fixed
`/crm/sales/api` path segment to build the base URL, matching legacy's domain-derived
`freshsalesBaseURL` fallback path.

## Streams notes

All 4 streams (`contacts`, `sales_accounts`, `deals`, `leads`) share the same shape: `GET` against
a view-scoped list endpoint (`<resource>/view/<view_id>`), records at the resource-named top-level
key (e.g. `{"contacts":[...]}`), primary key `["id"]`, cursor field `updated_at` (informational
only — see Known limits). `view_id` defaults to `"0"` (Freshsales's default/all-view alias,
matching legacy's `freshsalesDefaultView`) and is shared across every stream via the single
`config.view_id` template. Pagination is `page_number` (`page` query param, no size param sent —
Freshsales list endpoints do not accept a page-size override) with `page_size: 100` as the
short-page stop threshold. Legacy itself has no page-size concept at all (no config key, no
constant, no size param ever sent) — its real stop condition is `meta.total_pages` when present,
falling back to "stop only on a truly EMPTY page" when it is absent (`freshsales.go:171-183`;
the fallback branch never compares against a record count, only `len(records) == 0`). The engine's
`page_number` paginator has no `meta.total_pages`-aware stop mode, only a short-page (`recordCount
< page_size`) threshold, so `page_size` is set to a deliberately large, realistic value (100,
matching the sibling `page_number`+empty-`size_param` convention used by xero/judge-me-reviews/
buzzsprout for this exact "client-side stop threshold only, no server-side size param" shape) to
keep that threshold from ever firing on a genuine, non-final Freshsales page — every actual
termination in practice still comes from `meta.total_pages` (present on every real Freshsales list
response and on every fixture in this bundle) or a genuinely empty page, matching legacy's real
behavior far more closely than a small threshold would. A tiny `page_size` (e.g. 2) would invert
this: it would treat ordinary non-final pages as "short" and stop early whenever `meta.total_pages`
happened to be absent, which legacy itself never does.

## Write actions & risks

None. Freshsales is read-only (`capabilities.write: false`, no `writes.json`), matching legacy's
`Write` returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **Per-stream view-id overrides are not modeled.** Legacy accepts a per-stream config override
  (`<stream>_view_id`, e.g. `contacts_view_id`) in addition to the global `view_id`
  (`freshsales.go:325-335`, `viewID`). The engine's path templating resolves a single
  `{{ config.view_id }}` reference per stream with no per-stream-key indirection mechanism (a
  stream cannot template `config.<stream_name>_view_id` since the stream name is not itself
  interpolable into a config key lookup), so only the shared `view_id` config value is wired here;
  per-stream view overrides are out of scope for this migration. Every stream still defaults to
  view `"0"` exactly like legacy when no override is set.
- **`base_url` override is not modeled.** Legacy also accepts a direct `base_url` override
  (bypassing `domain_name`) for tests/proxies. The engine's `spec.json` `"default"` materialization
  mechanism only fills in a FIXED literal default, not a derived-from-another-config-value default
  (conventions.md §3): Freshsales's base URL is a function of `domain_name`, not a constant, so
  this bundle requires `domain_name` and does not expose a `base_url` escape hatch. This is a
  documented config-surface narrowing, not a behavior change for any `domain_name`-configured
  caller.
- **Freshsales exposes no incremental/server-side filter.** Legacy itself only supports full
  refresh (`freshsales.go:101-103`: "Freshsales only supports full_refresh upstream, so the cursor
  is informational"); this bundle declares no `incremental` block on any stream, matching legacy
  exactly. `updated_at` is still declared as `x-cursor-field` in each schema (as legacy declares
  `CursorFields`) so downstream `*_deduped` sync modes remain available, but no request-time
  filtering happens on either side.
- **`page_size` is not runtime-configurable** (there is no such legacy config key to restore in the
  first place — see Streams notes above). `spec.json` intentionally declares no `page_size`
  property (F6, conventions.md: a declared-but-unwireable key is worse than an absent one). The
  `contacts` stream's required 2-page conformance fixture (conventions.md §4) ships a full
  100-record page 1 (with `meta.total_pages: 2`, matching a real Freshsales response shape) followed
  by a genuinely short page 2, so the fixture exercises the engine's short-page-stop signal without
  relying on an unrealistically tiny `page_size` to do it; `sales_accounts`/`deals`/`leads` each
  ship a single terminal page (`meta.total_pages: 1`), unaffected by the `page_size` threshold at
  all.
