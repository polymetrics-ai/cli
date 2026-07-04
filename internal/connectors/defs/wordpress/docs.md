# Overview

WordPress is a Pass B full-surface declarative-HTTP connector for the WordPress core REST API v2
(`GET/POST/DELETE {base_url}/wp-json/wp/v2/...`), reachable on any WordPress site exposing the
standard REST API. It originated as a wave2 fan-out migration of `internal/connectors/wordpress`
(the hand-written connector it migrates, which reads only 7 streams and writes nothing); the legacy
package stays registered and unchanged until wave6's registry flip. This bundle now reads 10 streams
and writes 20 create/update/delete actions across the core content-management surface (verified
against the reference docs at `https://developer.wordpress.org/rest-api/reference/` and live
core-only WordPress installs' own `/wp-json/` route index and direct endpoint responses).

## Auth setup

Credentials are optional. When both `username` and `password` secrets are set, HTTP Basic auth is
applied (`Authorization: Basic ...`), matching legacy's `connsdk.Basic(username, password)`
(`wordpress.go:131-136`) — legacy applies Basic auth only when BOTH secrets are non-empty; if only
one is set, legacy sends no auth at all. This bundle's `auth` candidate gates on `secrets.password`
truthiness alone (the `when` grammar has no two-key AND), so a username-only-set, password-unset
edge case would, unlike legacy, still omit auth (password absent, so the bearer candidate's `when`
is false) — behaviorally identical to legacy in that specific case. The only divergence is a
password-set-but-username-unset configuration, which this bundle would send as Basic auth with an
empty username where legacy would send no auth at all; this is a narrow, rarely-hit edge case
(documented in Known limits) rather than a common-path deviation.

`base_url` must be provided explicitly (required). Legacy also accepts a bare `domain` config value
and derives `https://` + domain when `base_url` is unset (`wordpress.go:166-178`) — the engine's
`spec.json` `"default"` mechanism only supports a fixed literal, not a value derived from another
config key, so this bundle narrows the config surface to `base_url` only (see Known limits).

**Write actions require authenticated credentials with sufficient WordPress capabilities**
(`edit_posts`/`moderate_comments`/`manage_categories`/`edit_users`/`upload_files`, etc., per
WordPress's own core role-capability model) — an unauthenticated or under-privileged caller
attempting any write receives a real `401`/`403` from WordPress itself; this bundle declares no
capability-checking of its own (WordPress is the enforcement point, exactly as for every other
authenticated action).

## Streams notes

10 streams:

`posts`, `pages`, `comments`, `media`, `users`, `categories`, `tags` are unchanged from the prior
migration — all 7 are `GET` list endpoints under `/wp-json/wp/v2/`, each response body a bare
top-level JSON array (`records.path: "."`), 1-based `page_number` pagination (`page_param: page`,
`size_param: per_page`, stopping on a short page), and `"projection": "passthrough"` (legacy's
`Read` forwards every raw response object unfiltered — see below).

`taxonomies` (`GET /wp-json/wp/v2/taxonomies`), `types` (`GET /wp-json/wp/v2/types`), and `statuses`
(`GET /wp-json/wp/v2/statuses`) are new. All 3 are metadata-registry endpoints whose real response
body is a JSON OBJECT keyed by slug (`{"category": {...}, "post_tag": {...}}`), not an array — the
opposite envelope shape from the 7 content streams — modeled with `records.keyed_object: true` /
`key_field: slug` (each exploded record already carries its own `slug` field natively; `key_field`
stamps the identical value, a no-op in practice but keeps the declaration self-documenting). None of
the 3 accepts pagination or an `after`/date filter at all (they are small, fixed-size registries, not
growing content collections), so each declares `"pagination": {"type": "none"}` to override the
base's `page_number` default rather than sending meaningless `page`/`per_page` params to an endpoint
that ignores them.

Every stream declares `"projection": "passthrough"`. Legacy's `Read` does `emit(connectors.Record(item))`
— it forwards each raw WordPress API object completely, unfiltered (`wordpress.go:117-119`); the
`Fields` list attached to each `streamEndpoint` (e.g. `id`/`date`/`slug`/`status` for `posts`) is only
ever read by `streams()` to populate `Catalog()`'s advertised field list — it is display metadata, not
a filter applied anywhere in `Read`. The default `"schema"` projection mode would silently drop every
field the schema doesn't itemize (`title`, `content`, `excerpt`, `link`, `guid`, `author`,
`featured_media`, `categories`, `tags`, `_links`, etc. on `posts` alone), which would be a silent,
undocumented emitted-record-data change from legacy — `passthrough` avoids that. Each stream's
`schemas/<stream>.json` still declares the well-known WordPress REST API v2 fields (for
`x-primary-key`/`x-cursor-field` and typed documentation purposes), but under `passthrough` this is
non-exhaustive by design: any additional raw field a live site returns still survives to the emitted
record.

Legacy sends `after={{start_date}}` unconditionally on **every** stream's request whenever
`start_date` is configured (`wordpress.go:112-115`, computed once before the per-stream endpoint
lookup) — including `users`, `categories`, and `tags`, none of which have a `date` field or an
`after` filter semantics on the real WordPress REST API (WordPress silently ignores the
unrecognized param on those endpoints). This bundle reproduces that exact behavior for the 7
legacy-parity streams: every one of their `query` blocks declares `"after": {"template": "{{
config.start_date }}", "omit_when_absent": true}` (omitted entirely when `start_date` is unset,
present verbatim when set), matching legacy's `if after != "" { base.Set("after", after) }` gate.
The 3 new metadata streams (`taxonomies`/`types`/`statuses`) do NOT send `after` at all — legacy
never had these streams to begin with, so there is no legacy behavior to reproduce, and sending an
inert filter param on a fixed metadata registry has no precedent to preserve.

Crucially, legacy has **no state-cursor persistence anywhere** — `after` is recomputed from the
static `cfg.Config["start_date"]` value on every single call, never from an advancing,
app-persisted cursor. This bundle therefore does **not** declare `request_param`/`start_config_key`
on any stream's `incremental` block: doing so would make the engine treat `after` as a genuine
incremental filter, persisting and advancing a state cursor across syncs and sending a DIFFERENT,
advancing `after` value on every subsequent run — a silent, genuine change to what gets requested
and what records come back that legacy never does. The plain `query.after` entry above (not tied to
`incremental`) is what reproduces legacy's actual static, non-advancing behavior.

`posts`, `pages`, `comments`, and `media` declare a bare `incremental: {"cursor_field": "date"}` (no
`request_param`) — these 4 streams have a genuine `date` field and legacy's own `streams()` publishes
`CursorFields: []string{"date"}` for them via `Catalog()`, so the bare cursor-field declaration
reproduces that published catalog metadata without introducing a server-side filter legacy never had.
`users`/`categories`/`tags` still send the (inert) `after` param when `start_date` is configured,
matching legacy's request shape exactly, but declare no `x-cursor-field`/`incremental` block: legacy's
own `streams()` sets `CursorFields: []string{"date"}` unconditionally for every stream even though
`users`/`categories`/`tags` never emit a `date` field at all — this bundle does not reproduce that
inconsistency in the schema layer, since `x-cursor-field` must name a property that actually exists in
that same schema (a hard `connectorgen validate` rule); it is a legacy catalog-metadata quirk with no
bearing on emitted record data. `taxonomies`/`types`/`statuses` declare no `incremental` block either
— they are small fixed registries with no date field of any kind.

## Write actions & risks

20 write actions across 7 resources — every action is `external mutation; approval required`:

| Resource | Create | Update | Delete |
|---|---|---|---|
| posts | `create_post` | `update_post` | `delete_post` |
| pages | `create_page` | `update_page` | `delete_page` |
| comments | `create_comment` | `update_comment` | `delete_comment` |
| media | — (see Known limits) | `update_media` | `delete_media` |
| users | `create_user` | `update_user` | `delete_user` |
| categories | `create_category` | `update_category` | `delete_category` |
| tags | `create_tag` | `update_tag` | `delete_tag` |

Every action uses `body_type: "json"` and `POST` for update (matching WordPress core's own REST
API, which accepts `POST` as a `PUT` alias on every one of these endpoints and is what the
reference docs list first). `delete_post`/`delete_page`/`delete_comment` use WordPress's default
trash-then-permanently-delete-on-second-call semantics (no `force` query param embedded — a first
call moves the item to trash, matching the least-destructive default); `delete_media`/
`delete_category`/`delete_tag`/`delete_user` embed a literal `?force=true` (and, for `delete_user`,
also `&reassign={{ record.reassign }}`) directly in the write action's `path` template — WordPress
core's real API requires `force=true` for these 4 resource types since they do not support trashing
at all (attachments, taxonomy terms, and user accounts), and `delete_user` additionally requires a
`reassign` target user ID by WordPress's own validation, so `record_schema` marks `reassign` as
`required` alongside `id`. This is the dialect's standard way to express a write action's own fixed
query parameters (`write.go` has no dedicated query-building support, only path/body interpolation)
— see `docs/migration/conventions.md` precedent in `commercetools`'s `delete_customer` (`?version=
{{ record.version }}`) and `smartreach`'s several `?team_id={{ config.team_id }}` actions.
`create_user` requires `username`/`email`/`password` (WordPress's own required-field set for user
creation); `create_category`/`create_tag` require `name` (WordPress's only required field for term
creation); `create_comment` requires `post`/`content`.

## Known limits

- **`create_media` (media/attachment creation) is not covered.** WordPress core's real
  `POST /wp-json/wp/v2/media` requires a raw binary file upload in the request body with a
  `Content-Disposition: attachment; filename="..."` header — this dialect's write actions support
  only JSON/form bodies (`write.go`'s `body_type`), with no multipart/binary body construction of
  any kind. `update_media`/`delete_media` (pure JSON-metadata operations against an EXISTING
  attachment) are covered; only the binary upload itself is out of scope. See
  `api_surface.json`'s `binary_payload` exclusion category.
- **`domain`-derived `base_url` is not modeled; `base_url` is required instead.** Unchanged from
  the prior migration — see Auth setup above and `docs/migration/conventions.md`'s
  hostname-derivation gap class.
- **Basic-auth "both secrets required" is approximated by a single-key `when` gate on
  `password`.** Unchanged from the prior migration — see Auth setup above.
- **`page_size`/`max_pages` are not runtime-configurable.** Unchanged from the prior migration:
  both are fixed in `streams.json`'s `base.pagination` block (`page_size: 100`, `max_pages: 1`).
  The 3 new metadata streams (`taxonomies`/`types`/`statuses`) declare `"pagination": {"type":
  "none"}` instead, since their real endpoints accept no page/size parameters at all.
- **`plugins`/`themes`/`settings`/`sidebars`/`widgets` are excluded as `requires_elevated_scope`.**
  Verified live against a real WordPress install: an unauthenticated (or non-admin) caller receives
  `401 rest_cannot_view`/`rest_cannot_view_themes` from these endpoints even though some appear
  under the reference docs' general endpoint listing. These are site-administration/appearance
  surfaces, not general content data, and this bundle's optional Basic auth is intended for
  content-author-level access, not full site-admin capabilities.
- **The Gutenberg/full-site-editing surface (blocks, templates, template-parts, navigation, menus,
  menu-items, global-styles, font-families/font-collections, block-types, block-patterns,
  block-directory, widget-types, application-passwords) is entirely out of scope.** These are
  block-editor construction/theming metadata endpoints, not the general content data (posts/pages/
  comments/media/users/taxonomy) this connector targets. See `api_surface.json` for the full
  per-endpoint disposition (over 30 excluded paths).
- **`search` and `menu-locations` are excluded.** `search` is a keyword-query endpoint with no
  natural incremental-sync shape and no connector-level default search term; `menu-locations` was
  verified live to return `401 rest_cannot_view` for an unauthenticated/non-admin caller despite
  being documented as a general reference endpoint.
- **Legacy's fixture-mode-only fields are not modeled.** Unchanged from the prior migration.
- **Single-page fixtures only for the 7 paginated streams, matching `max_pages: 1`'s real,
  always-enforced behavior.** Unchanged from the prior migration; the 3 new metadata streams are
  inherently single-response (no pagination block at all), so this does not apply to them.
- **Schemas are non-exhaustive by design under `passthrough` projection.** Unchanged from the prior
  migration — every schema documents the well-known stable fields; a live site may emit additional
  fields verbatim.
