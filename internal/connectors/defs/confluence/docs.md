# Overview

Confluence reads Confluence Cloud spaces, pages, blog posts, labels, and attachments through the
Confluence Cloud REST API v2. This bundle migrates `internal/connectors/confluence` (the legacy
hand-written connector, kept registered and unchanged until wave6's registry flip) to a declarative
defs bundle at capability parity. Read-only.

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

## Streams notes

All 5 streams (`spaces`, `pages`, `blogposts`, `labels`, `attachments`) are `GET` requests against
the Confluence v2 list endpoints, with records at `results`, primary key `["id"]`, and
`limit=25` (matches legacy's default page size) sent on every request.

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
overly strict host-required parsing for what is, in practice, always a same-host relative link. See
Known limits for why this bundle does not use `pagination.type: cursor` with `token_path` instead.

`pages` and `blogposts` project a computed `version` field from the raw nested
`{"version": {"number": N}}` shape via `computed_fields`' bare `{{ record.version.number }}`
reference, which copies the raw typed integer value (typed extraction, not stringified). `spaces`,
`labels`, and `attachments` have no `version` field, matching legacy's own per-stream record
mappers exactly.

Legacy declares `CursorFields: ["createdAt"]` on `pages`/`blogposts`/`attachments` in its published
catalog but never actually performs incremental filtering in `Read`/`harvest` (it is a full-refresh
connector in practice). This bundle mirrors that exactly: `x-cursor-field: "createdAt"` is declared
on those 3 schemas (publishing the same catalog metadata), but no stream declares an `incremental`
block, so no incremental lower-bound filtering happens here either — full parity with legacy's
actual (not just advertised) behavior.

## Write actions & risks

None. Confluence is read-only in this port (`writes.json` is intentionally absent), matching
legacy's `Capabilities.Write: false`.

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
- Full Confluence v2 API surface (comments, content properties, custom content, folders,
  whiteboards, writes) is out of scope; see `api_surface.json`'s `excluded: {category:
  out_of_scope, reason: "Pass B capability expansion"}` entries. Only the 5 legacy-parity read
  streams are implemented.
- Per `docs/migration/conventions.md` §4's sanctioned exception, `fixtures/streams/**` ships a
  single page per stream (not a 2-page fixture): a `next_url` stream's next-page URL is the replay
  server's own address, unknown until the harness picks a port at runtime, so a static fixture file
  cannot embed a correct second-page URL. `pagination_terminates` is satisfied by the paginator's
  own defined behavior (no `_links.next` in the fixture response means pagination stops after page
  1), matching every other `next_url` bundle's accepted shape (e.g. bitly, calendly).
