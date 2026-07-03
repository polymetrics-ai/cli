# Overview

Bugsnag is a fully declarative (Tier 1) migration. It reads a user's Bugsnag organizations,
projects, collaborators, errors, events, and releases through the Bugsnag Data Access API v2
(`GET {{ config.base_url }}/...`), a hierarchical resource tree (organizations -> projects ->
errors/events/releases). This bundle targets full capability parity with
`internal/connectors/bugsnag` (the hand-written connector it replaces); the legacy package stays
registered and unchanged until wave6's registry flip. Read-only (`capabilities.write` is `false`,
matching legacy's `Write` returning `connectors.ErrUnsupportedOperation`).

## Auth setup

Provide a Bugsnag personal auth token via the `auth_token` secret; it is sent as
`Authorization: token <auth_token>` (`mode: api_key_header`, `prefix: "token "`), matching
legacy's `connsdk.APIKeyHeader("Authorization", secret, "token ")`. Every request also carries the
mandatory `X-Version: 2` header (a static-literal `base.headers` entry, matching legacy's
`bugsnagAPIVersion` constant). `base_url` defaults to `https://api.bugsnag.com` and may be
overridden for tests/proxies.

## Streams notes

`organizations` is the root stream: `GET /user/organizations`, records at the response root
(`records.path: ""`, matching `connsdk.RecordsAt(resp.Body, "")`'s bare-array behavior), no
`fan_out`. The other 5 streams are sub-resource fan-out reads, expressible via the engine's
`stream.fan_out` primitive (`docs/migration/conventions.md` §3):

- `projects` (`GET /organizations/{organization_id}/projects`) and `collaborators`
  (`GET /organizations/{organization_id}/collaborators`) fan out over the `organization_id` config
  value (comma-separated list, `fan_out.ids_from.config_key`), forwarded into the path as
  `{{ fanout.id }}`.
- `errors` (`GET /projects/{project_id}/errors`), `events`
  (`GET /projects/{project_id}/events`), and `releases`
  (`GET /projects/{project_id}/releases`) fan out over the `project_id` config value the same way.

All 6 streams share `link_header` (RFC 5988 `Link: <url>; rel="next"`) pagination, matching
legacy's `connsdk.LinkHeaderPaginator`. `per_page` is templated from `config.page_size` (default
`100`, matching legacy's `bugsnagDefaultPageSize`/`bugsnagMaxPageSize`).

`projects`/`errors`/`events`/`releases` stamp the current fan-out id onto their
`organization_id`/`project_id` field respectively (`fan_out.stamp_field`) — Bugsnag's real API
already returns these as ordinary fields on child objects, so this guarantees the value even on a
response shape that omits it (a common real-world "implicit from the URL" API quirk), matching
the value the URL parameter itself carries either way. `errors`'s `events_count` field requires a
`computed_fields` rename from the raw API's `events` key (schema projection only copies by exact
key match; without the rename the field would silently drop). `errors`/`events`/`releases`
declare a bare `incremental.cursor_field` (`last_seen`/`received_at`/`release_time`) with no
`request_param` — legacy publishes `CursorFields` for these 3 streams (a downstream incremental
sync watermark) but Bugsnag's Data Access API v2 has no server-side date-range filter parameter
legacy ever sends, matching `docs/migration/conventions.md` §8 rule 2's truth table exactly
(bare cursor_field iff legacy publishes CursorFields with no server-side filter).

Primary key `id` on every stream (every Bugsnag resource exposes a string id, matching legacy's
uniform `PrimaryKey: []string{"id"}}`).

## Write actions & risks

None. Legacy `bugsnag.go`'s `Write` returns `connectors.ErrUnsupportedOperation`
unconditionally; `capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- **Auto-discovery of `organization_id`/`project_id` is NOT modeled (documented config-surface
  narrowing).** Legacy's `resolveParents` prefers the configured id(s) and, only when unset,
  auto-discovers organizations via `/user/organizations` (for `projects`/`collaborators`) or
  projects across all the user's organizations via a nested org-then-project lookup (for
  `errors`/`events`/`releases`) — a genuine TWO-LEVEL discovery chain in the worst case (neither
  `organization_id` nor `project_id` configured). The engine's `fan_out` primitive resolves its id
  list from EXACTLY ONE preliminary request (or a config value) — it has no support for a
  multi-level "discover A, then discover B from each A" chain, and no way to express "prefer
  config, else auto-discover" as a single `ids_from` declaration (`config_key` and `request` are
  mutually exclusive). `organization_id`/`project_id` are both real, legacy-documented config keys
  a caller could already set explicitly; this migration requires them to be set for
  `projects`/`collaborators`/`errors`/`events`/`releases` (matching appfollow's identical
  documented `app_collection_ids`/`ext_ids`-required precedent) rather than silently
  auto-discovering with zero configuration. This never changes emitted record DATA for any
  configured-id input legacy itself would accept — only the zero-config auto-discovery convenience
  is out of scope.
- **`max_pages`** is not modeled (F6, dead config): legacy's `bugsnagMaxPages` accepts an optional
  cap (integer, `all`/`unlimited`/`0` synonyms for unbounded), but the engine's `link_header`
  paginator has no config-driven request-count override mechanism, so it is not declared in
  `spec.json`.
- The full Bugsnag Data Access API surface (single-resource lookups, error/event mutation
  endpoints, release group management) is out of scope for this wave; see `api_surface.json`'s
  `excluded` entries.
