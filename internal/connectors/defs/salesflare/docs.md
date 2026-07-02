# Overview

Salesflare is a CRM for small and medium B2B businesses. This bundle reads Salesflare accounts,
contacts, and opportunities through the Salesflare REST API. It is read-only, matching legacy
`internal/connectors/salesflare` exactly (`Capabilities{Write: false}`).

## Auth setup

Provide a Salesflare API key via the `api_key` secret. It is sent as `Authorization: Bearer
<api_key>` via `base.auth`'s `mode: bearer` — identical to legacy's `connsdk.Bearer(token)`. Never
logged.

## Streams notes

All 3 streams (`accounts`, `contacts`, `opportunities`) share the same shape: `GET` against the
Salesflare list endpoint, records at `data` (`records.path: "data"`), primary key `["id"]`.
`projection: passthrough` is used on every stream because legacy's own `Read` re-emits each decoded
record verbatim (`emit(connectors.Record(rec))` at `salesflare.go:126`) with no field filtering at
all — unlike a `mapRecord`-shaped connector, legacy never drops any raw API field, so schema
projection alone (which would silently drop any undeclared field) would be a data-parity
regression. The declared `schemas/*.json` properties are still a realistic, honest field set (for
`records_match_schema`'s type-checking of the fields that ARE declared) but do not gate which
fields survive.

Pagination is `cursor` with `token_path: pagination.next_page` and `cursor_param: page`: legacy's
`readPages` reads `pagination.next_page` (falling back to top-level `next_page`/`next` only as
defensive dead code never exercised by the real API or legacy's own test fixture — see legacy's own
`salesflare_test.go`, which uses `"pagination":{"next_page":2}` / `"pagination":{"next_page":null}`)
and treats a non-empty value as the next page's literal `page` query param value; a `null`/absent
`next_page` stops pagination. This is exactly the `cursor`/`token_path` shape:
`connsdk.StringAt`'s numeric stringify handles the wire-real integer `next_page` value, and a
JSON `null` resolves to `""` (falsy), stopping pagination identically to legacy's
`next == ""` check. Legacy's other two branches (`next` starting with `http://`/`https://`/`/`,
treated as a full URL/path override) are not modeled: they are unreachable in practice (Salesflare's
real API — and legacy's own recorded test fixture — only ever populates `pagination.next_page` with
a bare page number), so no accepted legacy input's behavior changes; see Known limits.

`limit` (the page-size query param) is a fixed literal `100` on every stream, matching legacy's
`defaultPageSize = 100`; `max_pages` is a fixed `100` in `streams.json`'s
`base.pagination.max_pages` (legacy's `defaultMaxPages = 100`). Neither is a runtime config
override — see Known limits.

## Write actions & risks

None. Salesflare is exposed read-only, matching legacy's `Capabilities{Write: false}` and its
`Write` method, which always returns `connectors.ErrUnsupportedOperation`.

## Known limits

- Full Salesflare API surface (tasks, workflows, files, custom-field metadata, teams,
  notifications, etc.) is out of scope for this wave; see `api_surface.json`'s
  `excluded: {category: out_of_scope, reason: "Pass B capability expansion"}` entries. Only the 3
  legacy-parity read streams are implemented.
- **`page_size`/`max_pages` are not runtime config overrides, and neither is declared in
  `spec.json`.** Legacy accepts both as config-driven overrides (`page_size` unbounded,
  `max_pages` also accepting `all`/`unlimited` sentinels for "no cap") and threads them through to
  its manual page loop. The engine's declarative pagination has no per-read config-driven override
  mechanism for either `PaginationSpec.PageSize` or `PaginationSpec.MaxPages` — both are fixed
  values baked into `streams.json`'s `base.pagination` block (the same "no runtime override"
  limitation `docs/migration/conventions.md` documents for searxng's `page_size`/`max_pages`). This
  bundle bakes in `max_pages: 100` (legacy's own default) and a fixed `limit: 100` query literal per
  stream; a caller wanting a different page size or cap (or "unlimited") cannot express it here.
  This is a documented config-surface narrowing, never a data-parity change for any read actually
  exercised at the default.
- **Legacy's URL/path-shaped `next_page` fallback branches are not modeled.** Legacy's `readPages`
  treats a `next_page` value starting with `http://`, `https://`, or `/` as a full next-page
  URL/path override rather than a literal `page` param value. Salesflare's real API (and legacy's
  own `salesflare_test.go` fixture) only ever returns a bare integer page number at
  `pagination.next_page`, so this branch is unreachable dead code in practice; the `cursor`/
  `token_path` paginator here only implements the bare-page-number shape. If Salesflare's API were
  ever observed returning a URL/path-shaped `next_page`, this would need re-scoping (a `next_url`
  paginator variant), but no such shape has ever been recorded.
- Legacy declares no incremental/cursor-field behavior for any of these 3 streams (no
  `CursorFields` in its `Catalog()`); this bundle matches that — no `x-cursor-field` is declared on
  any schema, and no `incremental` block is declared on any stream.
