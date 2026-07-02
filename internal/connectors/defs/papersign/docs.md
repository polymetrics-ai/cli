# Overview

PaperSign reads PaperSign documents, templates, and recipients through the REST API. This bundle
migrates `internal/connectors/papersign` (the hand-written legacy connector) to a declarative
Tier-1 defs bundle; the legacy package stays registered and unchanged until the wave6 registry
flip.

## Auth setup

Provide a PaperSign API key via the `api_key` secret; it is used only for Bearer auth
(`Authorization: Bearer <api_key>`) and is never logged.

## Streams notes

Three streams: `documents` (`GET /documents`), `templates` (`GET /templates`), `recipients`
(`GET /recipients`). All share the same shape: records at `data`, primary key `["id"]`, incremental
cursor field `created_at`. Pagination follows legacy's `harvestCursor`: a `cursor`-type pagination
block reads the next cursor from the body path `pagination.next_cursor` (`token_path`) and stops
when that value is empty — legacy's exact stop condition (`strings.TrimSpace(next) == ""`), with no
separate boolean stop signal (`stop_path` intentionally not declared, matching legacy exactly:
legacy relies solely on the empty-cursor stop, never a `has_more`-style flag). Every request sends
`limit=100` (matches legacy's default `page_size`) via each stream's static `query: {"limit":
"100"}`.

## Write actions & risks

None. This connector is read-only, matching legacy (`Capabilities.Write: false`,
`Write` returns `ErrUnsupportedOperation`).

## Known limits

- Only `documents`, `templates`, and `recipients` are implemented, matching legacy's exact stream
  set. Document creation/sending, template editing, and webhooks are out of scope for this wave;
  see `api_surface.json`'s `excluded` entries.
- Legacy's `limit`/`max_pages` runtime config overrides are not exposed as `spec.json` properties:
  the engine's `cursor` pagination type has no config-templated page-size or page-count field (both
  `page_size`/`max_pages` on a `pagination` block are static integers, not `{{ }}`-templatable), so
  a declared-but-unwireable `spec.json` property would be dead config (F6, `docs/migration/
  conventions.md` §3) — this bundle sends a fixed `limit=100` per request instead, matching
  legacy's own default exactly, and relies on the engine's hard `MaxPages` request-count cap (which
  this bundle leaves unbounded, matching legacy's `max_pages=0`/unlimited default) rather than a
  config-driven override.
