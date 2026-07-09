# Overview

Linear is a GraphQL-first connector. The implemented surface uses fixed, reviewed GraphQL
documents for approved reads and writes; it does not expose arbitrary GraphQL query or mutation
execution.

Readable streams: `issues`, `teams`, `projects`, `users`.

Direct-read view streams back `pm linear issue view`, `team view`, `project view`, and `user view`.

Approved write actions: `create_issue`, `update_issue`, `comment_issue`, and `create_project`.
They execute only through reverse ETL connector-command planning: plan, preview, approval token, and
execute. No write action is run during connector inspection or help rendering.

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
  local server origin. Streams append `/graphql`.
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

- `issues`: records path `data.issues.nodes`; cursor `updated_at`; includes state, team, assignee,
  creator, branch, and timestamp projections.
- `teams`: records path `data.teams.nodes`; cursor `updated_at`.
- `projects`: records path `data.projects.nodes`; cursor `updated_at`.
- `users`: records path `data.users.nodes`; cursor `updated_at`.
- `issue`, `team`, `project`, `user`: single-object GraphQL view streams used only by constrained
  direct-read CLI commands.

## Write actions & risks

Linear write actions are fixed GraphQL mutations and are never raw user-supplied GraphQL.

- `create_issue`: creates a visible Linear issue. Required fields: `team_id`, `title`.
- `update_issue`: mutates an existing Linear issue. Required field: `issue_id` plus at least one
  update field.
- `comment_issue`: creates a visible comment on a Linear issue. Required fields: `issue_id`, `body`.
- `create_project`: creates a visible Linear project. Required fields: `team_id`, `name`.

Every write is approval-gated through reverse ETL plan → preview → approval → execute. Sensitive,
admin, destructive, upload, auth, integration, webhook, invite, organization, and user-management
mutations are inventoried in `api_surface.json` and blocked by default.

## Known limits

- Raw arbitrary GraphQL is disallowed.
- The operation ledger inventories current `@linear/sdk` generated GraphQL root operations; only the
  approved stream/direct-read/write subset is executable.
- Binary/file uploads and admin/sensitive mutations require future operation-specific policy before
  they can become executable.
- Connector checks are metadata-safe locally; do not run credentialed Linear checks unless explicitly
  requested.
