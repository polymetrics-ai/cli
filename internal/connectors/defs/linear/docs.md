# Overview

Linear is a GraphQL-first connector. The implemented surface uses fixed, reviewed GraphQL
documents for reads and writes; it does not expose arbitrary GraphQL query or mutation execution.

Runtime reads include the original issue/team/project/user streams plus generated fixed-document
streams for every live Linear query root inventoried in `api_surface.json`. The provider-style CLI
surface intentionally highlights common commands (`issue/team/project/user list|view`) while the
connector manifest exposes the broader stream catalog for ETL.

Write coverage is modeled as fixed GraphQL reverse-ETL actions for every non-deprecated live Linear
mutation row in the operation ledger. These writes execute only through reverse ETL planning: plan,
preview, approval token, and execute. No write action is run during connector inspection or help
rendering.

Service API documentation: https://developers.linear.app/docs.

## Auth setup

Connection fields:

- `access_token` (optional, secret, string); Linear OAuth access token, sent as a Bearer
  Authorization header. Provide either `api_key` or `access_token`; `access_token` takes priority
  when both are set.
- `api_key` (optional, secret, string); Linear personal API key, sent as a bare Authorization header
  by default. Set `auth_type=oauth` or `auth_type=oauth2.0` to send it with a Bearer prefix.
- `auth_type` (optional, string); allowed values `oauth`, `oauth2.0`; default empty.
- `base_url` (optional, string); default `https://api.linear.app`; tests may override it with a
  local server origin. Streams and writes append `/graphql`.
- `max_pages` (optional, string); reserved for future configured page caps. Use command `--limit`
  or ETL `--batch-size` for bounded local reads today.

Secret fields are redacted in logs and write previews: `access_token`, `api_key`.

Default configuration values: `auth_type=`, `base_url=https://api.linear.app`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token` when present.
- Bearer token authentication using `secrets.api_key` when `auth_type` is `oauth` or `oauth2.0`.
- API key authentication in `Authorization` using `secrets.api_key` when present.
- No authentication only for local fixtures or public schema checks.

## Streams notes

All Linear runtime reads use `POST /graphql` with fixed documents stored in `streams.json`.
Connection streams use cursor pagination through `pageInfo.hasNextPage` and `pageInfo.endCursor`.

Curated CLI-backed streams:

- `issues`: records path `data.issues.nodes`; cursor `updated_at`; includes state, team, assignee,
  creator, branch, and timestamp projections.
- `teams`: records path `data.teams.nodes`; cursor `updated_at`.
- `projects`: records path `data.projects.nodes`; cursor `updated_at`.
- `users`: records path `data.users.nodes`; cursor `updated_at`.
- `issue`, `team`, `project`, `user`: single-object GraphQL view streams used by constrained
  direct-read CLI commands.

Generated query streams cover the remaining live query roots with fixed documents and explicit
query-variable mappings. Streams that require IDs or filters expect those values in the request query
map; connector inspection and help rendering never call Linear.

## Write actions & risks

Linear write actions are fixed GraphQL mutations and are never raw user-supplied GraphQL. The common
CLI-backed actions are:

- `create_issue`: creates a visible Linear issue.
- `update_issue`: mutates an existing Linear issue.
- `comment_issue`: creates a visible comment on a Linear issue.
- `create_project`: creates a visible Linear project.

The generated write catalog covers every non-deprecated live mutation root in the operation ledger
using explicit `record_schema` fields and fixed GraphQL variables. Every write is approval-gated
through reverse ETL plan → preview → approval → execute. Destructive/admin/sensitive-shaped mutations
carry risk text and `confirm: destructive` so the reverse-ETL execution path requires typed
confirmation.

## Known limits

- Raw arbitrary GraphQL is disallowed and remains blocked.
- The refreshed prompt's official non-deprecated target is fully inventoried and covered: 514 Linear
  GraphQL fields (156 query fields and 358 mutation fields).
- The live schema inventory contains 531 root fields including deprecated fields. This bundle covers
  530 fixed-document query/mutation rows, blocks the deprecated `integrationSettingsUpdate` row with
  exact schema evidence, and blocks the extra raw-arbitrary-GraphQL escape hatch.
- Generated query streams use a generic object schema unless a curated stream schema already exists;
  add operation-specific schemas/fixtures as follow-up hardening when a query becomes a primary CLI
  workflow.
- Connector checks are metadata-safe locally; do not run credentialed Linear checks unless explicitly
  requested.
