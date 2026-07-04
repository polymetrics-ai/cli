# Overview

Auth0 is a declarative-HTTP bundle migrated from `internal/connectors/auth0` (the hand-written
legacy connector, which stays registered and unchanged until wave6's registry flip). It reads
Auth0 users, clients, connections, roles, organizations, per-role user assignments, and
per-organization memberships, and creates/updates users, clients, roles, and organizations,
through the Auth0 Management API v2. Both of legacy's previously-blocking `ENGINE_GAP`s are closed:
0-indexed pagination (`start_page: 0`, S4 engine mini-wave item 1) and the M2M `audience`
token-request parameter (`auth[].extra_params`, S4 engine mini-wave item 4) are both expressible in
this dialect now. Pass B full-surface expansion (this revision) adds the 2 fan-out membership
streams and 8 write actions on top of the wave2 legacy-parity read surface.

## Auth setup

Legacy accepts two credential shapes with `access_token` taking priority whenever both are
configured (dual-auth ordering is load-bearing, conventions.md §3) — this bundle's `base.auth`
candidate list reproduces that exact precedence:

1. **`access_token` secret** (bearer, first candidate): used directly as a Bearer token
   (`{"mode": "bearer", "token": "{{ secrets.access_token }}", "when": "{{ secrets.access_token }}"}`).
2. **M2M `client_id`/`client_secret` secrets** (`oauth2_client_credentials`, fallback candidate,
   `when: "{{ secrets.client_id }}"`): performs an OAuth2 client-credentials exchange against
   `{{ config.base_url }}/oauth/token`, always including the `audience` config value as the
   `extra_params.audience` token-request form parameter (Auth0 requires this to scope the issued
   token to the Management API) — matching legacy's `authenticator`'s M2M branch and its
   `<base_url>/api/v2/`-by-convention audience exactly, modulo the one narrowing noted below.

`base_url` (your tenant domain, e.g. `https://your-tenant.us.auth0.com`) is required with no
default, matching legacy (`auth0BaseURL` requires it explicitly — the base is tenant-specific,
never a fixed literal).

**Narrowed from legacy: `audience` has no derived default.** Legacy defaults `audience` to
`<base_url>/api/v2/` when the config value is unset (`authenticator`'s `if audience == ""`
branch). The engine's `extra_params` dialect resolves each value via ordinary `Interpolate` with NO
`omit_when_absent`/`default` tolerance (conventions.md §3: "a misconfigured audience/subject param
should fail loudly rather than silently omit a value a real OAuth2 provider may require"), and
`audience` is a function of ANOTHER config value (`base_url`), not a fixed literal. This bundle
takes the documented "require it explicitly" path: `audience` must be set when using the M2M auth
candidate (typically `<base_url>/api/v2/`, matching legacy's own default value). Never changes
emitted record data for any configuration legacy itself would accept. Documented parity deviation,
ACCEPTABLE.

The same `base.auth` candidates authorize both reads and the 8 write actions below — Auth0's
Management API uses one Bearer/M2M token for the entire surface, read or write, so no additional
write-specific scope wiring exists in this dialect (the operator's underlying token/M2M
application must itself carry the relevant `create:users`/`update:users`/`create:clients`/etc.
scopes; the bundle has no way to declare or check scopes ahead of a request).

## Streams notes

**Legacy-parity streams** (unchanged from wave2): all 5 of `users`, `clients`, `connections`,
`roles`, `organizations` share the identical `include_totals=true` envelope shape
(`{start,limit,length,total,<resource>:[...]}`) and 0-indexed `page`/`per_page` page-number
pagination — matching legacy's `harvest`, which advanced `page` until a short page (fewer than
`per_page`) or the reported `total` was reached. `users` is the only legacy-parity stream with a
cursor field (`x-cursor-field: updated_at`, matching legacy's `CursorFields:
[]string{"updated_at"}`); no `incremental` block is declared since legacy never actually filters
`users` requests by `updated_at` (catalog metadata only in `auth0Streams()`, not a wired
request-param filter). `page_size`/`max_pages` are not exposed as config for the same
static-`PaginationSpec`-field reason documented in aviationstack's/searxng's goldens
(`PaginationSpec.PageSize`/`MaxPages` are plain JSON ints with no template/config override);
`streams.json`'s `pagination.page_size: 50` matches legacy's own `auth0DefaultPageSize`.

**New Pass B fan-out streams** each list every parent id via the SAME endpoint its non-fan-out
sibling stream already reads (`fan_out.ids_from.request`), then repeat the identical
`include_totals=true`/page-number sub-sequence once per parent id (`fan_out.into.path_var`),
stamping the parent id onto every emitted record (`fan_out.stamp_field`):

- `role_users` — `GET /api/v2/roles/{id}/users`, ids from `roles` (`records_path: roles`,
  `id_field: id`), `stamp_field: role_id`, records at `users`. Primary key `[role_id, user_id]`
  (the same user can appear under multiple roles; the role_id/user_id pair is the real unique
  edge, not `user_id` alone).
- `organization_members` — `GET /api/v2/organizations/{id}/members`, ids from `organizations`
  (`records_path: organizations`, `id_field: id`), `stamp_field: organization_id`, records at
  `members`. Primary key `[organization_id, user_id]`, same reasoning as `role_users`.

Both new streams' sub-sequences reuse the exact same `include_totals=true`/`page`/`per_page`
pagination shape as every other stream in this bundle (no stream-level pagination override) — the
Auth0 Management API's `page`/`per_page`/`include_totals` convention is uniform across every
listable resource, parent or child.

## Write actions & risks

All 8 actions require a Bearer/M2M token carrying the corresponding Management API scope (the
bundle has no scope-checking of its own — an insufficiently-scoped token fails with Auth0's own
403, surfaced as an ordinary write error).

- **`create_user`** (`POST /api/v2/users`, required: `connection`) / **`update_user`**
  (`PATCH /api/v2/users/{user_id}`, `path_fields: ["user_id"]`): creates or updates a user account,
  including credential fields (`password`) and verification flags. **Risk: external mutation;
  creates/updates a live Auth0 user account and credential; approval required.**
- **`create_client`** (`POST /api/v2/clients`, required: `name`) / **`update_client`**
  (`PATCH /api/v2/clients/{client_id}`, `path_fields: ["client_id"]`): registers or updates an
  Auth0 application. **Risk: external mutation; a new client can obtain its own OAuth2
  credentials; approval required.**
- **`create_role`** (`POST /api/v2/roles`, required: `name`) / **`update_role`**
  (`PATCH /api/v2/roles/{id}`, `path_fields: ["id"]`): creates or updates an RBAC role's
  name/description (Auth0 attaches no permissions to a newly-created role by default — a separate,
  unmodeled `POST /api/v2/roles/{id}/permissions` call is required to grant any, out of scope
  here; see Known limits). **Risk: external mutation; approval required.**
- **`create_organization`** (`POST /api/v2/organizations`, required: `name`) /
  **`update_organization`** (`PATCH /api/v2/organizations/{id}`, `path_fields: ["id"]`): creates or
  updates an organization's name/display_name. **Risk: external mutation; approval required.**

Every action uses `body_type: json` (Auth0's Management API is a JSON-body REST API throughout, no
form-encoded endpoints) with the default body construction (every record field except
`path_fields`-named ones).

## Known limits

- **`audience` has no derived default** — see Auth setup above. Documented parity deviation,
  ACCEPTABLE.
- **Membership-edge writes are out of scope**: `POST /api/v2/roles/{id}/users` (assign existing
  users to a role) and `POST /api/v2/organizations/{id}/members` (add existing users to an
  organization) are many-to-many membership-edge mutations, not a create-or-update of a role/user/
  organization record itself — modeling them would need a write action whose "record" is really an
  edge (parent id + a list of child ids), a shape this dialect's flat `record_schema` +
  `path_fields` convention does not cleanly express. Documented Pass B breadth-vs-cost triage, not
  an `ENGINE_GAP`.
- **Deeper fan-out chains are out of scope**: `GET /api/v2/users/{id}/roles`,
  `GET /api/v2/users/{id}/organizations`, `GET /api/v2/organizations/{id}/members/{user_id}/roles`,
  and `GET /api/v2/users/{id}/permissions`/`GET /api/v2/roles/{id}/permissions` would each need
  either a redundant inverse fan-out over an edge this bundle already covers in one direction
  (`role_users`/`organization_members`), or a genuinely 2-level-deep fan-out chain (e.g. list orgs
  -> list members -> list each member's roles) that the engine's `fan_out` dialect does not support
  (it resolves exactly one id list per stream, once per `Read()` call). Pass B breadth-vs-cost
  triage.
- **Connection strategy-specific config (`PATCH /api/v2/connections/{id}`) is out of scope**: each
  connection strategy (database, social, enterprise SAML/OIDC) has its own polymorphic `options`
  schema; this dialect's flat JSON-schema `record_schema` cannot faithfully express "one of N
  different shapes depending on `strategy`" without silently dropping or misvalidating
  strategy-specific fields. Not modeled as a write action.
- Only the 7 read streams and 8 write actions above are implemented; the broader Management API
  surface (deprecated Rules/Hooks, Actions custom code, Logs, tenant branding/prompts/flows,
  resource servers, client grants, device credentials, custom domains, attack protection, MFA
  guardian config, signing keys, and every tenant-wide admin/security/observability singleton) is
  out of scope — every excluded endpoint in `api_surface.json` carries its own specific category
  and reason (deprecated / non_data_endpoint / requires_elevated_scope / destructive_admin /
  out_of_scope breadth-vs-cost triage), not one blanket bucket.
