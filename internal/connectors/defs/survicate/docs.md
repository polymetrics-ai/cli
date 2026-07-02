# Overview

Survicate is a wave2 fan-out declarative-HTTP migration. It reads Survicate surveys through the
Survicate Data API (`GET https://data-api.survicate.com/v1/surveys`). This bundle targets
capability parity with `internal/connectors/survicate` (the hand-written connector it migrates);
the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Survicate API key via the `api_key` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_key>`) and is never logged, matching legacy's `connsdk.Bearer(secret)`
(`survicate.go:113`). `base_url` defaults to `https://data-api.survicate.com/v1` and may be
overridden for tests/proxies.

## Streams notes

The sole stream, `surveys` (`GET /surveys`), is the only stream legacy implements (`Read` rejects
any other stream name with `"survicate stream %q not found"`); records live at the `data` key.
`surveys` declares `incremental.cursor_field: updated_at`, matching legacy's own
`CursorFields: []string{"updated_at"}` declaration; neither this bundle nor legacy sends a
server-side lower-bound filter or performs client-side filtering (legacy's `Read` performs no
incremental filtering at all) — this bundle matches that exactly (no `request_param`/
`client_filtered` declared), not introducing new filtering under the guise of a migration.

Pagination is page-number (`pagination.type: page_number`, `page_param: page`, `size_param:
per_page`, `start_page: 1`, `page_size: 100`), identical to legacy's
`connsdk.PageNumberPaginator{PageParam: "page", SizeParam: "per_page", StartPage: 1, PageSize:
pageSize}` with legacy's default `pageSize` of 100. The `check` request sends both `page=1` and
`per_page=1`, matching legacy's `Check` exactly (`survicate.go:56`).

## Write actions & risks

None. Legacy's `Write` always returns `connectors.ErrUnsupportedOperation`; `capabilities.write` is
`false` and this bundle ships no `writes.json`.

## Known limits

- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes both as config-driven
  overrides (`survicate.go`'s `pageSize`/`maxPages`, bounded 1-100 / a non-negative integer or
  `all`/`unlimited`). The engine's `page_number` paginator has no config-driven page-size or
  max-pages knob (`PaginationSpec.PageSize`/`MaxPages` are static bundle JSON, never templated), so
  this bundle uses legacy's own default (`page_size: 100`) as a fixed bundle value and does not
  declare `page_size`/`max_pages` in `spec.json` at all (a declared-but-unwireable config key is
  worse than an absent one, per `docs/migration/conventions.md` F6). Pagination is unbounded by
  default (reads every page until a short page), matching legacy's own default of `maxPages=0`
  (unbounded) when `max_pages` is unset.
- Full Survicate API surface (responses, respondents, targeting) is out of scope for this wave; see
  `api_surface.json`'s `excluded: {category: out_of_scope}` entries.
