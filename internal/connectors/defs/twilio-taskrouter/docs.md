# Overview

Reads Twilio TaskRouter workers, tasks, activities, task queues, and workflows for a workspace.

Readable streams: `workers`, `tasks`, `activities`, `task_queues`, `workflows`.

This connector is read-only; no write actions are declared.

Service API documentation: https://www.twilio.com/docs/taskrouter/api.

## Auth setup

Connection fields:

- `account_sid` (required, secret, string); Twilio account SID, used as the Basic auth username.
  Never logged.
- `auth_token` (required, secret, string); Twilio auth token, used as the Basic auth password. Never
  logged.
- `base_url` (optional, string); default `https://taskrouter.twilio.com`; format `uri`; Twilio
  TaskRouter API base URL override for tests or proxies.
- `workspace_sid` (required, string); TaskRouter workspace SID every stream is scoped to;
  substituted into each stream's workspace-scoped path.

Secret fields are redacted in logs and write previews: `account_sid`, `auth_token`.

Default configuration values: `base_url=https://taskrouter.twilio.com`.

Authentication behavior:

- HTTP Basic authentication using `secrets.account_sid`, `secrets.auth_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/v1/Workspaces/{{ config.workspace_sid }}/Workers` with query
`PageSize`=`1`.

## Streams notes

Default pagination: page-number pagination; size parameter `PageSize`; starts at 1; page size 50;
maximum 1 page(s).

- `workers`: GET `/v1/Workspaces/{{ config.workspace_sid }}/Workers` - records path `workers`;
  page-number pagination; size parameter `PageSize`; starts at 1; page size 50; maximum 1 page(s).
- `tasks`: GET `/v1/Workspaces/{{ config.workspace_sid }}/Tasks` - records path `tasks`; page-number
  pagination; size parameter `PageSize`; starts at 1; page size 50; maximum 1 page(s).
- `activities`: GET `/v1/Workspaces/{{ config.workspace_sid }}/Activities` - records path
  `activities`; page-number pagination; size parameter `PageSize`; starts at 1; page size 50;
  maximum 1 page(s).
- `task_queues`: GET `/v1/Workspaces/{{ config.workspace_sid }}/TaskQueues` - records path
  `task_queues`; page-number pagination; size parameter `PageSize`; starts at 1; page size 50;
  maximum 1 page(s).
- `workflows`: GET `/v1/Workspaces/{{ config.workspace_sid }}/Workflows` - records path `workflows`;
  page-number pagination; size parameter `PageSize`; starts at 1; page size 50; maximum 1 page(s).

## Write actions & risks

This connector is read-only. Read behavior: external Twilio TaskRouter API read of workspace
workers, tasks, activities, task queues, and workflows.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  non_data_endpoint=1, out_of_scope=2.
