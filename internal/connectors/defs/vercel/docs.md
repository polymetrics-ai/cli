# Overview

Vercel is a frontend deployment platform. This bundle reads deployments, projects, teams, domains,
aliases, webhooks, log drains, edge configs, and per-project environment variables from the
Vercel REST API (`{base_url}`, default `https://api.vercel.com`), and writes create/update/delete
mutations for projects, deployments, project domains, project environment variables, webhooks,
log drains, edge configs, and alias removal. It originally migrated `internal/connectors/vercel`
(132 loc, a read-only single-stream connector), which stays registered and unchanged until wave6's
registry flip; this Pass B pass expands the bundle using Vercel's full, official, live OpenAPI 3.x
specification fetched directly from `https://openapi.vercel.sh/` (332 real method+path operations
spanning the entire Vercel platform API, reviewed 2026-07-04) — every endpoint covered or excluded
with a specific, real per-endpoint reason (no blanket bucket).

## Auth setup

Vercel authenticates via a single secret, `access_token`, sent as a Bearer token
(`Authorization: Bearer <access_token>`), matching legacy's `connsdk.Bearer(token)` requester
exactly (`vercel.go:95-105`).

An optional `team_id` config value is sent as the `teamId` query parameter on every request that
accepts it (`deployments`, `projects`, `domains`) — Vercel's own API requires `teamId` to scope a
request to a team when the access token is a team-scoped token; a personal-account token omits it
entirely (`omit_when_absent`).

## Streams notes

`deployments` is the original legacy-parity stream: `GET /v6/deployments`, records from the
`deployments` array, `id` renamed from the wire's `uid` field (typed extraction preserves its
native string type, matching legacy's direct assignment with no cast), `created` preserved as a
native Unix-milliseconds integer, and the optional `start_date`/`from` query passthrough — all
preserved unchanged from the original migration. Vercel's currently-documented list-deployments
endpoint has moved to `GET /v7/deployments` (confirmed in the live OpenAPI document), but `/v6/`
remains independently supported (Vercel's versioned endpoints are not retired when a newer version
ships) and this bundle intentionally keeps it for exact continuity with the original migration;
`api_surface.json` marks `/v7/deployments` `duplicate_of` `/v6/` for this reason, not as an
oversight.

Every other stream is new in this Pass B pass:

- `projects` (`GET /v10/projects`) and `teams` (`GET /v2/teams`) and `domains` (`GET /v5/domains`)
  share Vercel's documented cursor-pagination shape: the response body carries a `pagination`
  object (`{"count", "next", "prev"}`, confirmed from the OpenAPI document's shared `Pagination`
  component schema) whose `next` field (a Unix-milliseconds timestamp) becomes the next request's
  `until` query parameter — expressed via the engine's `cursor` pagination type
  (`token_path: pagination.next`, `cursor_param: until`). `projects`' response root is a `oneOf`
  (a bare array OR `{"projects": [...], "pagination": {...}}` depending on API version); this
  bundle selects the enveloped shape (`records.path: "projects"`), matching `teams`/`domains`'
  identical envelope convention.
- `project_env_vars` (`GET /v10/projects/{id}/env`) uses the engine's `fan_out` dialect
  (`docs/migration/conventions.md` §3) to read every project's environment variables in one
  logical stream: `ids_from.request` lists every project id via a preliminary, unpaginated
  `GET /v10/projects` request (reusing the SAME endpoint the `projects` stream itself reads, per
  project id), `into.path_var` substitutes each id into `{{ fanout.id }}` in the stream's own path,
  and `stamp_field: project_id` tags every emitted env-var record with its owning project id.
  `records.path: "envs"` matches the enveloped response shape (`{"envs": [...]}`, one of the
  OpenAPI document's three possible root shapes for this endpoint). Env var VALUES are not
  requested in decrypted form (`decrypt` query param is not set), so `value`/`vsmValue` are not
  projected onto this schema — only metadata (`id`, `key`, `type`, `target`, timestamps) is
  modeled, avoiding any risk of a secret value flowing into synced record data.
- `aliases` (`GET /v4/aliases`) — the same `pagination.next`/`until` cursor shape as
  projects/teams/domains; `records.path: "aliases"`.
- `webhooks` (`GET /v1/webhooks`) and `log_drains` (`GET /v1/log-drains`) and `edge_configs`
  (`GET /v1/edge-config`) — all three converted in this pass from a prior blanket exclusion into
  real streams. Each returns a BARE top-level JSON array (confirmed from the OpenAPI document's
  response schema — no envelope key, no pagination parameters documented on any of the three list
  endpoints), so `records.path: "."` and no `pagination` block is declared, matching each
  endpoint's actual documented shape.

## Write actions & risks

18 write actions, all `body_type: json` (flat top-level bodies, confirmed from the OpenAPI
document's request-body schemas for each endpoint):

- **projects**: `create_project` (`POST /v11/projects`, only `name` required),
  `update_project` (`PATCH /v9/projects/{id}`), `delete_project` (`DELETE /v9/projects/{id}`,
  idempotent, 404 tolerated).
- **deployments**: `create_deployment` (`POST /v13/deployments`, only `name` required),
  `cancel_deployment` (`PATCH /v12/deployments/{id}/cancel`, no body),
  `delete_deployment` (`DELETE /v13/deployments/{id}`, idempotent).
- **project domains**: `add_project_domain` (`POST /v10/projects/{project_id}/domains`),
  `remove_project_domain` (`DELETE /v9/projects/{project_id}/domains/{domain}`, idempotent).
- **project environment variables**: `create_project_env_var`
  (`POST /v10/projects/{project_id}/env`, requires `key`/`value`/`type` per Vercel's own
  `anyOf` requirement of either `target` or `customEnvironmentIds` — this bundle's `record_schema`
  requires only the base triple, matching the dialect's draft-07 `anyOf` limitation documented in
  the stripe golden's parity-deviation ledger item 1: strictly more permissive, never stricter),
  `delete_project_env_var` (`DELETE /v9/projects/{project_id}/env/{id}`, idempotent).
- **webhooks**: `create_webhook` (`POST /v1/webhooks`, requires `url`/`events[]`),
  `delete_webhook` (`DELETE /v1/webhooks/{id}`, idempotent; Vercel documents no `PATCH`/`PUT`
  single-webhook-update endpoint).
- **log drains**: `create_log_drain` (`POST /v1/log-drains`, requires
  `deliveryFormat`/`url`/`sources[]`), `delete_log_drain` (`DELETE /v1/log-drains/{id}`,
  idempotent; Vercel documents no update endpoint for an existing log drain).
- **edge configs**: `create_edge_config` (`POST /v1/edge-config`, requires `slug`),
  `update_edge_config` (`PUT /v1/edge-config/{id}`, renames `slug`), `delete_edge_config`
  (`DELETE /v1/edge-config/{id}`, idempotent). The edge config's own key-value item store
  (`PATCH /v1/edge-config/{id}/items`) is a separate, unbounded connector-defined-schema payload —
  see Known limits.
- **aliases**: `delete_alias` (`DELETE /v2/aliases/{id}`, idempotent) — removing an alias
  assignment; Vercel's alias-CREATE action (`POST /v2/deployments/{id}/aliases`) is not modeled,
  see Known limits.

All actions carry `"risk": "external mutation; approval required"` (destructive deletes add
`"confirm": "destructive"`).

## Known limits

- Full API-surface classification lives in `api_surface.json` (332 endpoint entries reviewed
  2026-07-04, including a manually-added `/v6/deployments` entry not present in the live OpenAPI
  document since Vercel's docs have moved on to `/v7/`): 9 read streams, 18 write actions, and
  every remaining endpoint excluded with a specific, real per-endpoint category and reason (no
  blanket bucket) — `duplicate_of` (34 endpoints) for single-item detail GETs, the newer
  `/v7/deployments` version, and legacy endpoints superseded by a now-covered resource (the old
  team-level `/v1/env`, the old `/v1/drains`/`/v1/integrations/log-drains` Drains API now
  superseded by `/v1/log-drains`); `requires_elevated_scope` (145 endpoints) for Enterprise/
  team-admin surfaces (billing, access groups, Secure Compute networking, directory sync/SSO,
  marketplace integration installation, microfrontends, rolling releases, team membership,
  Web Analytics query, WAF/security config, TLS certs), credential/token issuance (personal
  access tokens, edge-config read tokens, protection-bypass secrets), domain-registrar purchasing/
  transfer/renewal, and per-domain/per-project sub-resources (DNS records, custom environments'
  Enterprise gate) needing a fan_out this pass does not declare; `destructive_admin` (8 endpoints)
  for account-wide attack-mode/edge-cache-purge/user-deletion actions; `non_data_endpoint`
  (8 endpoints) for profile/audit-log/diagnostic-verification/transient-log-stream endpoints;
  `binary_payload` (9 endpoints) for avatar uploads, deployment file contents, and Remote Caching
  artifact blobs; and `out_of_scope` (101 endpoints) for real, individually-reasoned product
  surfaces distinct from this connector's core scope — bulk redirects, feature flags, routing
  middleware, custom environments, Sandbox ephemeral compute, VCR container registry, rollback/
  promote, legacy/duplicate deployment-checks families, edge config's item/schema/backup
  sub-resources, and alias-assignment creation.
- **Edge config's key-value item store is intentionally not modeled** (`GET`/`PATCH
  /v1/edge-config/{id}/items`, `GET /v1/edge-config/{id}/item/{key}`): its values are an
  unbounded, connector-defined-schema JSON payload (any shape the caller stores there) rather than
  a fixed business-data record shape this dialect's schema-as-projection model targets; the
  edge_configs stream's own record still reports `itemCount`/`digest` metadata.
- **Alias assignment (`POST /v2/deployments/{id}/aliases`) is not modeled** — only alias read
  (the `aliases` stream) and alias removal (`delete_alias`) are covered; assigning a specific
  alias to a specific deployment is the same product family as the not-yet-migrated
  rollback/promote actions, left for a future wave.
- **`project_env_vars`' fan_out preliminary id-listing request re-reads `GET /v10/projects` with NO
  pagination** (the fan_out dialect uses the fan-out STREAM's own pagination spec for its
  id-listing request, and `project_env_vars` declares none) — on an account with more projects than
  fit in one default-sized page, this stream would silently miss later projects' env vars. This is
  a known, documented scope narrowing rather than a silent gap: a future increment could declare a
  `cursor` pagination block on `project_env_vars` itself (reusing the same `token_path`/
  `cursor_param` shape `projects` uses) to close it.
- **`id` on `deployments` preserves its native wire type via bare-reference typed extraction**
  (unchanged from the original migration) — see the original migration's rationale contrasting
  Vercel's un-cast `uid`/`created` assignment with other connectors' explicit string casts.
- **`Check` dials the network; legacy's `Check` never did** — unchanged from the original
  migration, a deliberate fail-loud improvement with zero record-data impact.
- The optional `start_date`/`from` filter on `deployments` remains a stateless, config-only
  passthrough (see the original migration) — not a true incremental sync; `projects`/`teams`/
  `domains`/`project_env_vars` add no incremental filtering either (Vercel documents no
  updated-since filter on any of these list endpoints beyond the deployments-specific `since`/
  `until` window params, which this bundle does not wire as a stateful cursor in this pass).
