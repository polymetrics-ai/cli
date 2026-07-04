# Overview

Dremio reads and writes catalog entries, reflections, sources, users, and roles through the Dremio
REST API (defaulting to the Dremio Cloud US root, `https://api.dremio.cloud/v0`). This bundle
originally migrated `internal/connectors/dremio` (legacy) at capability parity; this Pass B pass
expands the surface with a `roles` read stream and 10 write actions covering user/role/reflection/
personal-access-token lifecycle mutation, researched directly against
`https://docs.dremio.com/current/reference/api/` (the current landing page for the
`docs_url`-declared `https://docs.dremio.com/software/rest-api/`). The legacy package stays
registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Dremio Personal Access Token (PAT) via the `api_key` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_key>`, `bearer` auth mode) and never logged, matching legacy's
`connsdk.Bearer(secret)` wiring exactly. The same PAT authorizes both reads and the new writes;
Dremio does not have separate read/write API keys.

## Streams notes

All 5 streams (`catalog`, `reflections`, `sources` (path `/source`), `users` (path `/user`), `roles`
(path `/role`, new)) share the same envelope: `GET` against the resource path, records at `data`,
primary key `["id"]`. Pagination is `cursor` with `token_path: nextPageToken` and `cursor_param:
pageToken` — Dremio returns `{"data":[...], "nextPageToken":"..."}`; the next page is requested with
`pageToken=<nextPageToken>`, and the engine's `tokenPathCursor` paginator stops on a
null/absent/non-advancing token exactly like legacy's `harvest` loop (`strings.TrimSpace(next) ==
""`), plus an empty-page stop. No `stop_path` is declared: legacy's stop condition is driven purely
by the token itself, so this bundle preserves that exact stop-on-empty-token-only behavior
(conventions.md §3).

Every request sends `maxResults` (default 100, matching legacy's `dremioDefaultPageSize`) via each
stream's static `query`. None of the five streams have a genuine incremental filter in the real API
(list endpoints accept no `updated_since`-style parameter), so no schema declares `x-cursor-field`
and no stream declares an `incremental` block — these streams are full-refresh only.

`roles` (new this pass) reads `GET /role`, an identically `data`-wrapped envelope to the other four
streams; each role object carries `id`/`name`/`type` (`INTERNAL`/`EXTERNAL`/`SYSTEM`)/`description`/
`memberCount`/a nested `roles` array of role references it belongs to.

## Write actions & risks

10 actions, all against Dremio's real documented v3 shapes (the new writes use the API's actual
singular `/reflection`/`/role` path segments; only the four pre-existing parity-locked READ stream
paths keep legacy's original — including its plural `/reflections` — path spelling, since those are
locked to legacy behavior and this pass does not touch them):

- `create_user` (`POST /user`), `update_user` (`PUT /user/{id}`), `delete_user` (`DELETE
  /user/{id}`, idempotent on 404, `confirm: destructive`) — Dremio user account lifecycle.
  `update_user`'s `active: false` can lock a user out immediately.
- `create_role` (`POST /role`), `update_role` (`PUT /role/{id}`), `delete_role` (`DELETE
  /role/{id}`, idempotent on 404, `confirm: destructive`) — Dremio role lifecycle. Role membership
  management (`PATCH /role/{id}/member`) is a diff-shaped add/remove list body, not a uniform
  single-record write this dialect targets — excluded, see `api_surface.json`.
  `create_role`/`update_role` do not manage grants; a role with no privileges granted separately
  has no practical access.
- `update_reflection` (`PUT /reflection/{id}`) — mutates name/enabled/tag; disabling a reflection
  removes its query-acceleration benefit until re-enabled and rebuilt.
- `refresh_reflection` (`POST /reflection/{id}/refresh`, `kind: custom`, no body) — forces an
  immediate rebuild; low risk, no data loss, no approval required (the only action in this bundle
  without an approval requirement).
- `delete_reflection` (`DELETE /reflection/{id}`, idempotent on 404, `confirm: destructive`) —
  permanently removes the reflection definition.
- `create_personal_access_token` (`POST /user/{user_id}/token`, `path_fields: ["user_id"]`,
  `body_fields: ["label", "millisToExpire"]`) — mints a new PAT for the named user. The real API's
  response carries the plaintext token exactly once; this bundle's write path never logs write
  response bodies, but any downstream consumer of the write result must treat it as a one-time
  secret reveal, same as the real Dremio UI does.
- `delete_personal_access_token` (`DELETE /user/{user_id}/token/{token_id}`, `path_fields:
  ["user_id", "token_id"]`, idempotent on 404, `confirm: destructive`) — revokes a single PAT.

Catalog object create/update/delete (`POST`/`PUT`/`DELETE /catalog[/{id}]`) and reflection create
(`POST /reflection`) are excluded as `requires_elevated_scope`/`destructive_admin`: catalog object
bodies are type-specific (a source's connection credentials, a view's SQL, an aggregation
reflection's dimension/measure fields), which this dialect's uniform per-action `record_schema` has
no type-dispatch primitive to express safely, and a generic catalog delete-by-id can remove an
entire source and everything nested beneath it — a materially larger blast radius than the
per-resource deletes this bundle does model. SQL job submission (`POST /sql`) is excluded as
`requires_elevated_scope`: it is an arbitrary-code-execution capability (DDL/DML against any
accessible source depending on grants), not a bounded record mutation; legacy never implemented it
either, and its dependent `/job/*` endpoints are consequently unreachable and also excluded. See
`api_surface.json` for the full endpoint-by-endpoint breakdown, including PAT/token system-wide
revocation, search, scripts, workload management, and Cloud-only engine-management, all excluded as
out-of-scope operational/admin surfaces rather than syncable data streams or bounded writes.

## Known limits

- The full Dremio API surface beyond the 5 read streams and 10 writes above (SQL execution and its
  dependent job endpoints, catalog object mutation, script/search/workload-management/engine
  administration, external-token-provider/LDAP configuration) is out of scope — see
  `api_surface.json`'s `excluded` entries for the endpoint-by-endpoint reasoning.
- Dremio Cloud vs. Dremio Software/self-hosted vs. EU-region deployments use different base URLs
  (and, per Dremio's current Cloud docs, Cloud now requires project-scoped paths like
  `/v0/projects/{project_id}/catalog` rather than the flat `/catalog` this bundle and legacy both
  use) — legacy resolves only a flat `base_url` config override with no derivation/project-scoping
  logic, and this bundle matches that exactly (`base_url` default is the Dremio Cloud US root; any
  other deployment, including any Cloud project requiring the newer project-scoped path shape, must
  set `base_url` explicitly to a compatible flat-routed endpoint, e.g. a self-hosted Dremio Software
  instance's `<host>/api/v3`).
- Legacy's `reflections` stream name and its `/reflections` (plural) wire path predate this pass and
  are kept exactly as-is for parity; the real Dremio v3 API's actual resource noun is singular
  (`/reflection`), which the new `update_reflection`/`refresh_reflection`/`delete_reflection` writes
  use directly since they are not constrained by legacy parity. This is an intentional
  read-vs-write path-spelling asymmetry within the same bundle, not an inconsistency to "fix" by
  breaking the parity-locked stream.
- `create_role`/`update_role` do not manage the role's own privilege grants (`PUT
  /catalog/{id}/grants`-style catalog-object grants, or the role-member `PATCH` sub-resource) —
  those are separate excluded surfaces (see Write actions & risks above), so a newly created role
  has no access until grants/membership are configured through some other path.
