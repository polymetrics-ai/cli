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
  "name")` — the raw `id` field, falling back to `name` when a record has no `id`. `projects` and
  `segments` objects DO carry a real `id` from the Unleash admin API (distinct from `name`), so those
  streams project the raw `id` directly. The `environments` list endpoint (`GET
  /api/admin/environments`), by contrast, returns NO `id` on any record — per Unleash's own
  `environmentSchema` the object is keyed on `name` (properties `name`/`type`/`enabled`/`protected`/
  `sortOrder`, no `id`) — so legacy's `first(item, "id", "name")` ALWAYS resolves to `name` there.
  This bundle reproduces that exactly with a `computed_fields` rename (`"id": "{{ record.name }}"`) on
  the `environments` stream, emitting `{"id": name, "name": name}` for every record just as legacy
  does; the environments fixture carries the real wire shape (no fabricated `id`). Because the derived
  `id` always equals `name`, `x-primary-key: ["id"]` / `required: ["id"]` stay satisfied for every
  record the real API returns. (There is no coalesce/fallback combinator for a record that carries a
  real `id` on SOME rows and none on others within one stream, but no Unleash stream has that shape.)
- `page_size` (100) and `max_pages` (1) are fixed values baked into `streams.json`'s
  `base.pagination` block, not exposed as runtime `spec.json` config properties — the engine's
  pagination spec has no config-templating mechanism. This matches legacy's own DEFAULT behavior
  (`defaultPageSize = 100`, `defaultMaxPages = 1`) but drops legacy's ability to override either at
  request time via `config.page_size`/`config.max_pages`.
- Full Unleash admin API surface (feature toggle mutation, strategies, tags, addons, API tokens,
  events, playground) is out of scope for this wave; see `api_surface.json`'s `excluded` entries.
  Only the 4 legacy-parity streams are implemented.
