# Overview

Rollbar is a wave2 fan-out declarative-HTTP migration. It reads Rollbar projects and error items
through the Rollbar API (`GET https://api.rollbar.com/api/1/...`). This bundle is migrated from
`internal/connectors/rollbar` (the hand-written connector it replaces); the legacy package stays
registered and unchanged until wave6's registry flip. Read-only (`capabilities.write` is `false`,
matching legacy's `Write` returning `connectors.ErrUnsupportedOperation`).

## Auth setup

Provide a Rollbar project access token via the `access_token` secret; it is sent as the
`X-Rollbar-Access-Token` header (`mode: api_key_header`, no prefix), matching legacy's
`connsdk.APIKeyHeader("X-Rollbar-Access-Token", token, "")`. `base_url` defaults to
`https://api.rollbar.com` and may be overridden for tests/proxies.

## Streams notes

`items` (`GET /api/1/items/`, records at `result.items`) and `projects` (`GET /api/1/projects`,
records at `result`) share the same `page_number` pagination (`page`/`per_page`, `page_size:
100`), matching legacy's `page`/`per_page` query params and default page size. Both streams'
primary key is `["id"]`; neither declares an incremental cursor (legacy exposes none for either
endpoint).

## Write actions & risks

None. Legacy `rollbar.go`'s `Write` returns `connectors.ErrUnsupportedOperation`
unconditionally; `capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- **Pagination stop-condition parity is DATA-equivalent but not request-count-identical.**
  Legacy's `readPages` stops when `len(records)==0` OR the response body's own
  `result.total_pages`/`result.page` indicate the current page is the last one — it can stop
  exactly at the last full page without ever requesting an empty page after it. The engine's
  `page_number` paginator only recognizes a single stop signal: a page returning fewer records
  than `page_size`. When Rollbar's true last page happens to be exactly `page_size` (100) records
  long, the engine issues one additional request for the next page (which Rollbar returns empty)
  before stopping, where legacy would have stopped immediately using `total_pages`. This never
  changes the set of emitted records (the extra page is empty) and never loops — it is a strictly
  data-neutral extra-request difference, acceptable per `docs/migration/conventions.md` §5's
  meta-rule (never changes emitted record data for any input legacy would accept).
  `total_pages`/`page` fields are not modeled in `streams.json` because the engine's `page_number`
  paginator has no body-introspection stop condition beyond record-count.
- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes both as config-driven
  overrides (both non-negative integers, `0` meaning unbounded pages for `max_pages`). The
  engine's `page_number` paginator reads `PaginationSpec.PageSize`/`MaxPages` as static
  bundle-authored integers, not config templates. This bundle sends `page_size: 100` (legacy's
  own default) as a static value; neither key is declared in `spec.json` (F6: dead config is
  worse than absent config).
- The full Rollbar API surface (deploys, teams, users, RQL queries, item status mutation) is out
  of scope for this wave; see `api_surface.json`'s `excluded` entries.
