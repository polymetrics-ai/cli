# Overview

Confluence reads Confluence Cloud spaces, pages, blog posts, labels, attachments, footer/inline
comments, tasks, and app-defined custom content, and writes pages, blog posts, and comments,
through the Confluence Cloud REST API v2. This bundle migrates `internal/connectors/confluence`
(the legacy hand-written connector, kept registered and unchanged until wave6's registry flip) to a
declarative defs bundle at capability parity, then expands past that legacy-parity floor with a
Pass B full-surface review of the live v2 resource-group documentation.

## Auth setup

Provide `base_url` (your site's full API base URL including the `/wiki/api/v2` path, e.g.
`https://mysite.atlassian.net/wiki/api/v2`), `email` (the Atlassian account email), and the
`api_token` secret (an Atlassian API token). Auth is HTTP Basic (`email:api_token`); the token is
only ever passed into the `basic` auth mode and never logged.

Legacy also accepted a bare `domain_name` config key and derived the base URL in code as
`https://<domain_name>/wiki/api/v2`. The engine's `spec.json` `"default"` mechanism only
materializes a fixed literal default, not one derived from another config value (see
`docs/migration/conventions.md` §3's "spec.json default values" section) — so this bundle requires
`base_url` directly and drops the `domain_name`-derivation convenience. See Known limits.

`custom_content_type` (optional, defaults to `com.atlassian.confluence.macro.core`) selects which
app-defined custom-content type key the `custom_content` stream reads — `GET /custom-content` has
no type-agnostic list mode; `type` is a required query parameter on the real API.

## Streams notes

All 9 streams are `GET` requests against the Confluence v2 list endpoints, with records at
`results`, primary key `["id"]`, and `limit=25` sent on every request (matches legacy's default
page size for the original 5 streams; the 4 new streams adopt the same page size for consistency).

Pagination follows Confluence v2's cursor-in-relative-URL convention: each page response includes
`_links.next`, a **relative** path (e.g. `/wiki/api/v2/spaces?cursor=<token>&limit=25`) carrying an
opaque cursor token — Confluence's own docs confirm `_links.next` (and the equivalent `Link` response
header) is always relative, never absolute. This bundle uses `pagination.type: next_url` with
`next_url_path: "_links.next"` and `allow_cross_host: true`. `allow_cross_host` is ordinarily an
opt-out of the engine's same-origin SSRF guard for a genuinely cross-host next-page URL; here it is
the only way to route around the guard's stricter "next URL has no host" fail-closed check, which
otherwise rejects every relative URL outright regardless of same-origin intent (the guard's
same-origin case implicitly assumes an absolute next URL). This is safe for Confluence specifically:
`_links.next` is a same-instance relative path by API design, incapable of ever pointing at a
different host — there is no cross-host redirection risk being waived here, only the guard's
overly strict host-required parsing for what is, in practice, always a same-host relative link.

`pages`, `blogposts`, `footer_comments`, `inline_comments`, and `custom_content` project a computed
`version` field from the raw nested `{"version": {"number": N}}` shape via `computed_fields`' bare
`{{ record.version.number }}` reference, which copies the raw typed integer value (typed extraction,
not stringified). `spaces`, `labels`, `attachments`, and `tasks` have no `version` field in the real
API response, matching legacy's own per-stream record mappers exactly for the 5 original streams;
`tasks`' own docs confirm it carries no version-history field at all (unlike every other content
type in v2).

Legacy declares `CursorFields: ["createdAt"]` on `pages`/`blogposts`/`attachments` in its published
catalog but never actually performs incremental filtering in `Read`/`harvest` (it is a full-refresh
connector in practice). This bundle mirrors that exactly: `x-cursor-field: "createdAt"` is declared
on those 3 schemas (publishing the same catalog metadata), but no stream declares an `incremental`
block, so no incremental lower-bound filtering happens here either — full parity with legacy's
actual (not just advertised) behavior. The 4 new Pass B streams declare no `x-cursor-field` either:
none of `footer_comments`/`inline_comments`/`tasks`/`custom_content`'s real API responses expose a
top-level `createdAt`/`updatedAt` field suitable as a cursor at the list level (comments/custom
content nest their timestamp under `version.createdAt`, and tasks' own `createdAt`/`updatedAt` are
not documented as filterable/sortable list parameters), so no incremental capability is advertised
for them, honestly reflecting the real surface rather than a narrowed one.

`footer_comments` and `inline_comments` are separate v2 resources (not two views of one `comments`
list) — `inline_comments` additionally carries a `resolutionStatus` field (`open`/`resolved`/
`reopened`) that footer comments do not have, matching the real API's distinct schemas for the two
comment kinds.

## Write actions & risks

Six write actions are exposed, all `external mutation` / `approval required` per `metadata.json`'s
`capabilities.write: true`:

- `create_page` (`POST /pages`) — creates a page in `spaceId` with `title` and a `body.value`/
  `body.representation` (Confluence storage-format XHTML is the conventional representation).
- `update_page` (`POST` via `PUT /pages/{{ record.id }}`) — mutates an existing page's `title`/
  `status`/`body`. Confluence's PUT requires the caller to supply the NEXT `version.number`
  (strictly greater than the page's current version) — a stale/repeated version number is rejected
  by the live API, not by this bundle; callers must track the current version (e.g. from a prior
  `pages` stream read) themselves.
- `create_blogpost` (`POST /blogposts`) — creates a blog post, same body shape as `create_page`
  minus `parentId` (blog posts have no page-hierarchy parent).
- `create_footer_comment` (`POST /footer-comments`) — creates a footer comment or, when
  `parentCommentId` is set, a reply, on a page or blog post.
- `create_inline_comment` (`POST /inline-comments`) — creates an inline comment anchored to a text
  selection (`inlineCommentProperties.textSelection`, required by the real API to locate the
  anchor) on a page or blog post.

No delete action is exposed for any resource (all deletes are `destructive_admin`, out of scope —
see `api_surface.json`), and no update action is exposed for blog posts/comments/custom content
(breadth-vs-cost triage for this probe run — see `api_surface.json`'s `out_of_scope` entries on
those PUT endpoints; `create_*` covers the primary lifecycle mutation for each resource).

## Known limits

- Legacy's configurable `page_size` (1-250, default 25) and `max_pages` (0/all/unlimited or a
  positive integer cap) config knobs are not modeled. The `next_url` pagination type does not read
  a page-size field at all (the next page's full URL, including its own `limit` query param, comes
  verbatim from `_links.next`), and there is no config-driven override mechanism for a
  `streams.json` pagination block's static fields (same class of limitation as searxng's
  `page_size`/`max_pages`, `docs/migration/conventions.md`'s read-only/no-auth golden). Every
  stream's static `query: {"limit": "25"}` sends the same page size legacy defaulted to; a caller
  needing a different page size or a bounded page count has no equivalent knob in this bundle. A
  declared-but-unwireable config key is worse than an absent one (F6, REVIEW.md), so `page_size`/
  `max_pages` are not declared in `spec.json`.
- `domain_name`-based base URL derivation (legacy's `https://<domain_name>/wiki/api/v2` fallback)
  is not modeled; `base_url` must be supplied in full. This is a documented config-surface
  narrowing per `docs/migration/conventions.md` §3 (derived defaults are not expressible via the
  engine's `spec.json` `"default"` mechanism), not a data or behavior deviation for any input this
  bundle does accept.
- Per-content-id-scoped sub-resources (content properties, version history, ancestors/children/
  descendants for every content type; space permissions/roles/properties; classification levels;
  redactions; likes) have no global/cross-item list endpoint in the real API — enumerating them
  would require a `fan_out` parent-id source over every already-read page/blogpost/attachment/
  comment/custom-content id, which this bundle does not declare config for. See
  `api_surface.json`'s per-endpoint `out_of_scope` reasons (each cites the specific fan_out gap,
  not a blanket bucket).
- Whiteboards, databases, folders, and smart links (embeds) have no global list endpoint at all in
  v2 (only single-item detail-by-id and creation) — they cannot be modeled as streams under this
  dialect's list-endpoint stream shape, and their creation payloads (freeform canvas/binary
  containers) have no meaningful declarative `record_schema`; see `api_surface.json`.
- `custom-content`'s create/update endpoints are excluded: the request body shape is defined per
  app-registered custom-content `type` with no single stable schema this probe run can express
  safely across arbitrary types; only the read stream (parameterized by `custom_content_type`) is
  covered.
- Per `docs/migration/conventions.md` §4's sanctioned exception, `fixtures/streams/**` ships a
  single page per stream (not a 2-page fixture): a `next_url` stream's next-page URL is the replay
  server's own address, unknown until the harness picks a port at runtime, so a static fixture file
  cannot embed a correct second-page URL. `pagination_terminates` is satisfied by the paginator's
  own defined behavior (no `_links.next` in the fixture response means pagination stops after page
  1), matching every other `next_url` bundle's accepted shape (e.g. bitly, calendly).
