# Overview

Devin AI reads and writes the organization-scoped JSON surface of the current
Devin v3 REST API. The connector covers sessions and session child resources,
playbooks, secret metadata, knowledge notes and folders, repositories,
schedules, members, consumption, metrics, and self metadata.

The bundle also defines documented JSON mutations for session lifecycle and
messages, schedules, playbooks, knowledge notes, repository indexing, PR review
triggers, and secret deletion. Mutations require the normal reverse-ETL plan,
preview, approval, and execute flow.

## Auth setup

Provide `api_token` as a secret Devin service-user API key. The engine sends it
as `Authorization: Bearer <api_token>`. Provide `org_id` as non-secret config;
organization-scoped streams and writes use `/v3/organizations/{org_id}/...` or
the documented `/v3beta1/organizations/{org_id}/...` beta resources.

`base_url` defaults to `https://api.devin.ai`. Override it only for tests or a
trusted proxy; the engine applies the same-origin and path traversal guards
described in `docs/migration/conventions.md`.

Optional config:

- `start_date`: RFC3339 lower bound for session-derived streams, converted to
  Devin's Unix-seconds `created_after` query param.
- `page_size`: page size sent as `first` on cursor-paginated list streams.
- `metrics_time_after` and `metrics_time_before`: optional Unix-seconds bounds
  for metrics and consumption streams.
- `repository_filter_name`: optional repository-name filter.
- `user_email`: optional membership email filter.

## Streams notes

Most Devin list endpoints return an `items` array plus cursor pagination with
`end_cursor` and `has_next_page`. Those streams use the engine cursor paginator
with `after` as the request cursor. Schedule listing uses Devin's
`limit`/`offset` pagination. Singleton metrics, consumption, session detail,
session tags, session attachments, and self endpoints use no pagination.

Streams implemented:

- `sessions`, `sessions_insights`, `session_details`, `session_messages`,
  `session_attachments`, and `session_tags`
- `playbooks` and `secrets`
- `knowledge_notes` and `knowledge_folders`
- `repositories` and `indexed_repositories`
- `schedules`
- `organization_users`, `organization_idp_group_users`, and `self`
- `org_daily_consumption`, `session_daily_consumption`, and
  `user_daily_consumption`
- `org_usage_metrics`, `org_session_metrics`, `org_active_users_metrics`,
  `org_daily_active_users`, `org_monthly_active_users`, `org_pr_metrics`,
  `org_search_metrics`, and `org_weekly_active_users`

Session-child streams fan out from `sessions`. User consumption fans out from
`organization_users`. The `repositories` stream requests Devin's default
repository indexing status payload, so the per-repository indexing status point
endpoint is not modeled as a separate fan-out stream.

## Write actions & risks

Implemented actions:

- Session writes: `create_session`, `send_session_message`,
  `append_session_tags`, `replace_session_tags`, `archive_session`,
  `terminate_session`, and `generate_session_insights`
- Schedule writes: `create_schedule`, `update_schedule`, and `delete_schedule`
- Playbook writes: `create_playbook`, `update_playbook`, and `delete_playbook`
- Knowledge writes: `create_knowledge_note`, `update_knowledge_note`, and
  `delete_knowledge_note`
- Repository writes: `index_repository`, `bulk_index_repositories`,
  `remove_repository_indexing`, `bulk_remove_repository_indexing`, and
  `remove_repository_branch_indexing`
- Other writes: `trigger_pr_review` and `delete_secret`

Write actions can launch Devin work, send session messages, mutate reusable
automation content, change repository indexing state, trigger PR reviews, or
delete objects. `delete_secret` is intentionally limited to deleting metadata by
id; secret creation is excluded because it would require accepting secret
values as write-record data.

Single-repository indexing writes use `encoded_repository_path` because Devin's
path parameter is a full repository path such as `org/repo-name`, while the
engine does not URL-encode interpolated path variables. Branch-removal writes
also use `encoded_branch_name` for the same reason.

## Known limits

- Enterprise administration endpoints are excluded with explicit
  `requires_elevated_scope` or `destructive_admin` reasons in
  `api_surface.json`; this connector is organization-scoped.
- Binary and presigned-file endpoints, including attachments and snapshot
  blueprint file payloads, are excluded because the engine dialect is JSON-only.
- Point lookups that require non-enumerable caller-supplied identifiers and do
  not form practical ETL streams are excluded as `non_data_endpoint` or covered
  by an equivalent list stream.
- The single-repository indexing status point endpoint is covered by
  `repositories.indexing_status`; fan-out to it would require URL-encoding a
  full repository path.
- Legacy v1 and v2 API families are excluded as deprecated. This bundle targets
  the current Devin service-user v3 API documented from
  `metadata.json.docs_url`.
- Legacy `mode: fixture` remains a test/conformance affordance only; live reads
  use the documented Devin API.
