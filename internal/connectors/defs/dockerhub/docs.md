# Overview

Docker Hub reads public repositories, image tags, and the namespace/user profile for a configured
`docker_username` through Docker Hub's public registry API (`https://hub.docker.com/v2`). This
bundle migrates `internal/connectors/dockerhub` (legacy) at capability parity: the legacy package
stays registered and unchanged until wave6's registry flip.

## Auth setup

Docker Hub's public registry API requires no authentication for public repositories — `base.auth`
declares `{"mode": "none"}` explicitly (matching legacy's credential-free `requester`, which wires no
authenticator at all). No secret is configured or required.

## Streams notes

- `repositories` (`GET /repositories/{docker_username}/`) and `tags` (`GET
  /repositories/{docker_username}/{repository}/tags`) both return Docker Hub's
  `{count,next,previous,results}` envelope; records are extracted at `results`. Pagination is
  `next_url` (`next_url_path: next`) — Docker Hub's `next` field is an absolute URL, exactly the
  shape `next_url` pagination is built for; the first request additionally sends `page=1` and
  `page_size` (default 100, matching legacy's `dockerhubDefaultPageSize`) via the stream's static
  `query`.
- `tags` requires the `repository` config field (the repository name without the namespace prefix),
  matching legacy's `requiresRepository` gate on the `tags` stream definition. There is no
  declarative way to make a query/path segment conditionally required per-stream in this dialect;
  an unset `repository` on a `tags` read hard-errors via the path's unresolved `config.repository`
  reference, which is the correct fail-loud behavior (legacy also hard-errors in this case, just
  with a friendlier message).
- `namespace` (`GET /users/{docker_username}/`) reads a single JSON object (no `results` wrapper);
  `records.path` is `""` (root) with `single_object: true`, and `pagination: {"type": "none"}`
  overrides the base `next_url` pagination for this one stream (legacy's `readSingle` never
  paginates the namespace endpoint).
- None of the three streams have a genuine server-side incremental filter in the legacy connector
  (Docker Hub's list endpoints do not accept a since/updated-after query parameter) — legacy exposes
  `CursorFields` purely as bookkeeping metadata for a full-refresh-only source. This bundle mirrors
  that: `repositories`/`tags` schemas declare `x-cursor-field: last_updated` (matching legacy's
  `CursorFields`) but no stream declares an `incremental` block, so only `full_refresh_append`/
  `full_refresh_append_deduped` sync modes apply — never `incremental_append*` — exactly matching
  legacy's real behavior (its `InitialState` cursor is bookkeeping only, never read back into a
  request).

## Write actions & risks

None. Docker Hub's public registry API exposes no safe generic reverse-ETL write actions; legacy
itself is read-only (`Capabilities.Write: false`), so no `writes.json` is shipped.

## Known limits

- Only the 3 legacy-parity read streams (`repositories`, `tags`, `namespace`) are implemented; the
  full Docker Hub API surface (repository create/update/delete, webhooks, per-tag image manifests)
  is out of scope for this wave — see `api_surface.json`'s `excluded` entries.
- `fixtures/streams/repositories/` and `fixtures/streams/tags/` ship a single-page fixture rather
  than a 2-page fixture: this bundle's pagination type is `next_url`, whose next-page URL is the
  replay server's own runtime-assigned address and cannot be embedded in a static fixture file
  (conventions.md §4's sanctioned `next_url` exception, formalized from bitly/calendly's identical
  shape). `pagination_terminates` exercises the `namespace` stream instead (a genuinely non-paginated
  single-object endpoint), and real 2-page `next_url` correctness is asserted by parity/unit tests
  authored in a future wave with Go-authoring scope (this migration wave is JSON/docs only).
