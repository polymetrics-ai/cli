# Overview

Unleash reads Unleash projects, feature toggles, environments, and segments through the Unleash
admin API's list endpoints. This bundle migrates `internal/connectors/unleash` (the hand-written
connector) to a declarative defs bundle at capability parity; the legacy package stays registered
and unchanged until wave6's registry flip.

## Auth setup

Provide an Unleash admin API token via the `api_token` secret; it is used only for Bearer auth
(`Authorization: Bearer <api_token>`) and is never logged.

## Streams notes

Four streams, all `GET`, all list endpoints:

- `projects` (`GET /api/admin/projects`, records at `projects`, primary key `id`) — account-wide,
  not project-scoped.
- `features` (`GET /api/admin/projects/{{ config.project_id }}/features`, records at `features`,
  primary key `name`) — project-scoped; `project_id` defaults to `"default"` (Unleash's own default
  project), matching legacy's `defaultProject` constant.
- `environments` (`GET /api/admin/environments`, records at `environments`, primary key `id`).
- `segments` (`GET /api/admin/segments`, records at `segments`, primary key `id`).

Pagination is `offset_limit` (`limit`/`offset` query params, page size 100) with a short-page stop
— matches legacy's `harvest` loop exactly. `max_pages` is baked into `streams.json`'s
`base.pagination` block as a fixed value of `1` (legacy's own `defaultMaxPages = 1`, i.e. legacy
itself only reads a single page unless overridden) rather than exposed as a runtime `spec.json`
config property: the engine's `PaginationSpec.MaxPages` is a static bundle-authored integer with no
`{{ }}` templating/config-override mechanism (there is no way to wire a `config.max_pages` value
into it), so a genuinely runtime-adjustable page cap — like legacy's own — cannot be expressed;
`page_size` is similarly baked in as a fixed `100` for the same reason. This mirrors searxng's
identical `max_pages`/`page_size` "no runtime override mechanism" pattern (see
`docs/migration/conventions.md` §1's searxng worked example) — neither `page_size` nor `max_pages`
is declared in `spec.json` since no template anywhere in this bundle consumes them (a declared,
unwireable config key is worse than an absent one, per the F6 rule).

None of the 4 streams declare an `x-cursor-field`: legacy itself declares no `CursorFields` for any
Unleash stream (config toggles/projects/environments/segments have no natural "last modified"
timestamp legacy ever surfaced), so every stream is `full_refresh`-only here too, matching legacy
exactly.

## Write actions & risks

None. Unleash is read-only in both legacy and this bundle (`capabilities.write: false`) — feature
toggle mutation is a side-effecting action inappropriate for a generic reverse-ETL source.

## Known limits

- Legacy's `projects`/`environments`/`segments` record mappers derive `id` as `first(item, "id",
  "name")` — a defensive fallback to the `name` field when a raw record has no `id` at all. This
  bundle projects the raw `id` field directly with no fallback: the fallback path is untested in
  legacy (no test exercises it) and Unleash's real admin API always returns an `id` on every project/
  environment/segment object, so this narrowing never changes behavior for any record the real API
  actually returns. The dialect has no coalesce/fallback template combinator (`computed_fields`
  supports a single bare reference, a filter chain, or a static literal — not an either-or of two
  reference paths), so a hypothetical malformed record missing `id` entirely would be dropped from
  projection here (schema `required: ["id"]`) rather than falling back to `name` as legacy would.
- `page_size` (100) and `max_pages` (1) are fixed values baked into `streams.json`'s
  `base.pagination` block, not exposed as runtime `spec.json` config properties — the engine's
  pagination spec has no config-templating mechanism. This matches legacy's own DEFAULT behavior
  (`defaultPageSize = 100`, `defaultMaxPages = 1`) but drops legacy's ability to override either at
  request time via `config.page_size`/`config.max_pages`.
- Full Unleash admin API surface (feature toggle mutation, strategies, tags, addons, API tokens,
  events, playground) is out of scope for this wave; see `api_surface.json`'s `excluded` entries.
  Only the 4 legacy-parity streams are implemented.
