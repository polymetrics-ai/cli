# Overview

Reads and writes documented Rollbar API v1 resources through the connector engine.

Readable streams: `activated_counts_report`, `deploy`, `deploys`, `email_notification_rule`,
`email_notification_rules`, `environments`, `invitation`, `item_by_counter`, `item_by_id`,
`item_by_uuid`, `item_occurrences`, `items`, `occurrence`, `occurrence_counts_report`,
`occurrences`, `pagerduty_notification_rule`, `pagerduty_notification_rules`, `person_deletion_job`,
`project`, `project_teams`, `projects`, `rql_job`, `rql_job_result`, `rql_jobs`, `service_link`,
`service_links`, `session_replay`, `slack_notification_rule`, `slack_notification_rules`, `team`,
`team_invitations`, `team_project_assignment`, `team_projects`, `team_user_assignment`,
`team_users`, `teams`, `top_active_items_report`, `user`, `user_projects`, `user_teams`, `users`,
`version`, `version_items`, `webhook_notification_rule`, `webhook_notification_rules`.

Write actions: `assign_team_to_project`, `assign_user_to_team`, `cancel_invitation`,
`cancel_rql_job`, `configure_email_notifications`, `configure_slack_notifications`, `create_item`,
`create_project`, `create_service_link`, `create_team`, `delete_email_notification_rule`,
`delete_occurrence`, `delete_pagerduty_notification_rule`, `delete_project`, `delete_service_link`,
`delete_session_replay`, `delete_slack_notification_rule`, `delete_team`,
`delete_webhook_notification_rule`, `invite_team_user`, `remove_team_from_project`,
`remove_user_from_team`, `update_deploy`, `update_email_notification_rule`, `update_item`,
`update_pagerduty_notification_rule`, `update_service_link`, `update_slack_notification_rule`,
`update_webhook_notification_rule`.

Service API documentation: https://docs.rollbar.com/reference/getting-started-1.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); Rollbar API access token, sent as the
  X-Rollbar-Access-Token header. Never logged.
- `base_url` (optional, string); default `https://api.rollbar.com`; format `uri`; Rollbar API base
  URL override for tests or proxies.
- `counter` (optional, string); Project-local Rollbar item counter for item_by_counter.
- `deploy_id` (optional, string); Rollbar deploy ID.
- `environment` (optional, string); Rollbar environment for version and session replay streams.
- `event` (optional, string); default `new`; allowed values `new`, `repeated`, `reactivated`,
  `resolved`; Version item event filter required by Rollbar for version_items.
- `instance_id` (optional, string); Rollbar occurrence instance ID.
- `invite_id` (optional, string); Rollbar team invitation ID.
- `item_id` (optional, string); Rollbar item ID for item detail, item occurrence, and item mutation
  endpoints.
- `job_id` (optional, string); Rollbar RQL or person-deletion job ID.
- `project_id` (optional, string); Rollbar project ID for project-scoped streams and writes.
- `replay_id` (optional, string); Rollbar replay ID for session replay endpoints.
- `rule_id` (optional, string); Rollbar notification rule ID.
- `service_link_id` (optional, string); Rollbar service link ID.
- `session_id` (optional, string); Rollbar session ID for session replay endpoints.
- `team_id` (optional, string); Rollbar team ID.
- `user_id` (optional, string); Rollbar user ID.
- `uuid` (optional, string); Occurrence UUID for the item_by_uuid stream.
- `version` (optional, string); Application code version for Rollbar version streams.

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `base_url=https://api.rollbar.com`, `event=new`.

Authentication behavior:

- API key authentication in `X-Rollbar-Access-Token` using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/api/1/items/` with query `page`=`1`; `per_page`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `per_page`; starts
at 1; page size 100.

Pagination by stream: none: `activated_counts_report`, `deploy`, `email_notification_rule`,
`invitation`, `item_by_counter`, `item_by_id`, `item_by_uuid`, `occurrence`,
`occurrence_counts_report`, `pagerduty_notification_rule`, `person_deletion_job`, `project`,
`rql_job`, `rql_job_result`, `service_link`, `session_replay`, `slack_notification_rule`, `team`,
`team_project_assignment`, `team_user_assignment`, `top_active_items_report`, `user`, `version`,
`webhook_notification_rule`; page_number: `deploys`, `email_notification_rules`, `environments`,
`item_occurrences`, `items`, `occurrences`, `pagerduty_notification_rules`, `project_teams`,
`projects`, `rql_jobs`, `service_links`, `slack_notification_rules`, `team_invitations`,
`team_projects`, `team_users`, `teams`, `user_projects`, `user_teams`, `users`, `version_items`,
`webhook_notification_rules`.

- `activated_counts_report`: GET `/api/1/reports/activated_counts` - single-object response; records
  path `.`; emits passthrough records.
- `deploy`: GET `/api/1/deploy/{{ config.deploy_id }}` - single-object response; records path
  `result`; emits passthrough records.
- `deploys`: GET `/api/1/deploys` - records path `result.deploys`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits passthrough
  records.
- `email_notification_rule`: GET `/api/1/notifications/email/rule/{{ config.rule_id }}` -
  single-object response; records path `result`; emits passthrough records.
- `email_notification_rules`: GET `/api/1/notifications/email/rules` - records path `result`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; emits passthrough records.
- `environments`: GET `/api/1/environments` - records path `result.environments`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits
  passthrough records.
- `invitation`: GET `/api/1/invite/{{ config.invite_id }}` - single-object response; records path
  `result`; emits passthrough records.
- `item_by_counter`: GET `/api/1/item_by_counter/{{ config.counter }}` - single-object response;
  records path `result`; emits passthrough records.
- `item_by_id`: GET `/api/1/item/{{ config.item_id }}` - single-object response; records path
  `result`; emits passthrough records.
- `item_by_uuid`: GET `/api/1/item/` - single-object response; records path `result`; query
  `uuid`=`{{ config.uuid }}`; emits passthrough records.
- `item_occurrences`: GET `/api/1/item/{{ config.item_id }}/instances` - records path
  `result.instances`; page-number pagination; page parameter `page`; size parameter `per_page`;
  starts at 1; page size 100; emits passthrough records.
- `items`: GET `/api/1/items/` - records path `result.items`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 100; emits passthrough records.
- `occurrence`: GET `/api/1/instance/{{ config.instance_id }}` - single-object response; records
  path `result`; emits passthrough records.
- `occurrence_counts_report`: GET `/api/1/reports/occurrence_counts` - single-object response;
  records path `.`; emits passthrough records.
- `occurrences`: GET `/api/1/instances` - records path `result.instances`; page-number pagination;
  page parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits passthrough
  records.
- `pagerduty_notification_rule`: GET `/api/1/notifications/pagerduty/rule/{{ config.rule_id }}` -
  single-object response; records path `result`; emits passthrough records.
- `pagerduty_notification_rules`: GET `/api/1/notifications/pagerduty/rules` - records path
  `result`; page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1;
  page size 100; emits passthrough records.
- `person_deletion_job`: GET `/api/1/people/delete_jobs/{{ config.job_id }}` - single-object
  response; records path `result`; emits passthrough records.
- `project`: GET `/api/1/project/{{ config.project_id }}` - single-object response; records path
  `result`; emits passthrough records.
- `project_teams`: GET `/api/1/project/{{ config.project_id }}/teams` - records path `result`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; emits passthrough records.
- `projects`: GET `/api/1/projects` - records path `result`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 100; emits passthrough records.
- `rql_job`: GET `/api/1/rql/job/{{ config.job_id }}` - single-object response; records path
  `result`; emits passthrough records.
- `rql_job_result`: GET `/api/1/rql/job/{{ config.job_id }}/result` - single-object response;
  records path `result`; emits passthrough records.
- `rql_jobs`: GET `/api/1/rql/jobs/` - records path `result`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 100; emits passthrough records.
- `service_link`: GET `/api/1/service_links/{{ config.service_link_id }}` - single-object response;
  records path `result`; emits passthrough records.
- `service_links`: GET `/api/1/service_links` - records path `result`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits passthrough
  records.
- `session_replay`: GET `/api/1/environment/{{ config.environment }}/session/{{ config.session_id
  }}/replay/{{ config.replay_id }}` - single-object response; records path `.`; emits passthrough
  records.
- `slack_notification_rule`: GET `/api/1/notifications/slack/rule/{{ config.rule_id }}` -
  single-object response; records path `result`; emits passthrough records.
- `slack_notification_rules`: GET `/api/1/notifications/slack/rules` - records path `result`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; emits passthrough records.
- `team`: GET `/api/1/team/{{ config.team_id }}` - single-object response; records path `result`;
  emits passthrough records.
- `team_invitations`: GET `/api/1/team/{{ config.team_id }}/invites` - records path `result`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; emits passthrough records.
- `team_project_assignment`: GET `/api/1/team/{{ config.team_id }}/project/{{ config.project_id }}`
  - single-object response; records path `result`; emits passthrough records.
- `team_projects`: GET `/api/1/team/{{ config.team_id }}/projects` - records path `result`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; emits passthrough records.
- `team_user_assignment`: GET `/api/1/team/{{ config.team_id }}/user/{{ config.user_id }}` -
  single-object response; records path `result`; emits passthrough records.
- `team_users`: GET `/api/1/team/{{ config.team_id }}/users` - records path `result`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits
  passthrough records.
- `teams`: GET `/api/1/teams` - records path `result`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 100; emits passthrough records.
- `top_active_items_report`: GET `/api/1/reports/top_active_items` - records path `result`; emits
  passthrough records.
- `user`: GET `/api/1/user/{{ config.user_id }}` - single-object response; records path `result`;
  emits passthrough records.
- `user_projects`: GET `/api/1/user/{{ config.user_id }}/projects` - records path `result.projects`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; emits passthrough records.
- `user_teams`: GET `/api/1/user/{{ config.user_id }}/teams` - records path `result.teams`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; emits passthrough records.
- `users`: GET `/api/1/users` - records path `result.users`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 100; emits passthrough records.
- `version`: GET `/api/1/versions/{{ config.version }}` - single-object response; records path `.`;
  query `environment`=`{{ config.environment }}`; emits passthrough records.
- `version_items`: GET `/api/1/versions/{{ config.version }}/items` - records path `result`; query
  `environment`=`{{ config.environment }}`; `event`=`{{ config.event }}`; page-number pagination;
  page parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits passthrough
  records.
- `webhook_notification_rule`: GET `/api/1/notifications/webhook/rule/{{ config.rule_id }}` -
  single-object response; records path `result`; emits passthrough records.
- `webhook_notification_rules`: GET `/api/1/notifications/webhook/rules` - records path `result`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; emits passthrough records.

## Write actions & risks

Overall write risk: external Rollbar API mutations for items, projects, teams, assignments,
invitations, notification settings/rules, service links, occurrences, and session replay deletes;
approval required before execution.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `assign_team_to_project`: PUT `/api/1/team/{{ record.team_id }}/project/{{ record.project_id }}` -
  kind `update`; body type `none`; path fields `team_id`, `project_id`; required record fields
  `team_id`, `project_id`; accepted fields `project_id`, `team_id`; risk: Changes Rollbar
  membership/assignment for /api/1/team/{team_id}/project/{project_id}; external access mapping
  mutation, approval required.
- `assign_user_to_team`: PUT `/api/1/team/{{ record.team_id }}/user/{{ record.user_id }}` - kind
  `update`; body type `none`; path fields `team_id`, `user_id`; required record fields `team_id`,
  `user_id`; accepted fields `team_id`, `user_id`; risk: Changes Rollbar membership/assignment for
  /api/1/team/{team_id}/user/{user_id}; external access mapping mutation, approval required.
- `cancel_invitation`: DELETE `/api/1/invite/{{ record.invite_id }}` - kind `delete`; body type
  `none`; path fields `invite_id`; required record fields `invite_id`; accepted fields `invite_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: Mutates
  Rollbar by deleting/removing/canceling /api/1/invite/{invite_id}; destructive external mutation,
  approval required.
- `cancel_rql_job`: POST `/api/1/rql/job/{{ record.job_id }}/cancel` - kind `delete`; body type
  `none`; path fields `job_id`; required record fields `job_id`; accepted fields `job_id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: Mutates Rollbar by
  deleting/removing/canceling /api/1/rql/job/{job_id}/cancel; destructive external mutation,
  approval required.
- `configure_email_notifications`: PUT `/api/1/notifications/email` - kind `update`; body type
  `json`; body fields `enabled`; accepted fields `enabled`; risk: Changes Rollbar notification
  configuration for /api/1/notifications/email; external configuration mutation, approval required.
- `configure_slack_notifications`: PUT `/api/1/notifications/slack` - kind `update`; body type
  `json`; body fields `enabled`, `channel`, `show_message_buttons`; accepted fields `channel`,
  `enabled`, `show_message_buttons`; risk: Changes Rollbar notification configuration for
  /api/1/notifications/slack; external configuration mutation, approval required.
- `create_item`: POST `/api/1/item/` - kind `create`; body type `json`; body fields `data`; accepted
  fields `data`; risk: Mutates Rollbar through POST /api/1/item/; external API write, approval
  required.
- `create_project`: POST `/api/1/projects` - kind `create`; body type `json`; body fields `name`;
  accepted fields `name`; risk: Mutates Rollbar through POST /api/1/projects; external API write,
  approval required.
- `create_service_link`: POST `/api/1/service_links` - kind `create`; body type `json`; body fields
  `name`, `template`; accepted fields `name`, `template`; risk: Mutates Rollbar through POST
  /api/1/service_links; external API write, approval required.
- `create_team`: POST `/api/1/teams` - kind `create`; body type `json`; body fields `name`,
  `access_level`; accepted fields `access_level`, `name`; risk: Mutates Rollbar through POST
  /api/1/teams; external API write, approval required.
- `delete_email_notification_rule`: DELETE `/api/1/notifications/email/rule/{{ record.rule_id }}` -
  kind `delete`; body type `none`; path fields `rule_id`; required record fields `rule_id`; accepted
  fields `rule_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: Mutates Rollbar by deleting/removing/canceling /api/1/notifications/email/rule/{rule_id};
  destructive external mutation, approval required.
- `delete_occurrence`: DELETE `/api/1/instance/{{ record.instance_id }}` - kind `delete`; body type
  `none`; path fields `instance_id`; required record fields `instance_id`; accepted fields
  `instance_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: Mutates Rollbar by deleting/removing/canceling /api/1/instance/{instance_id}; destructive
  external mutation, approval required.
- `delete_pagerduty_notification_rule`: DELETE `/api/1/notifications/pagerduty/rule/{{
  record.rule_id }}` - kind `delete`; body type `none`; path fields `rule_id`; required record
  fields `rule_id`; accepted fields `rule_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: Mutates Rollbar by deleting/removing/canceling
  /api/1/notifications/pagerduty/rule/{rule_id}; destructive external mutation, approval required.
- `delete_project`: DELETE `/api/1/project/{{ record.project_id }}` - kind `delete`; body type
  `none`; path fields `project_id`; required record fields `project_id`; accepted fields
  `project_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: Mutates Rollbar by deleting/removing/canceling /api/1/project/{project_id}; destructive
  external mutation, approval required.
- `delete_service_link`: DELETE `/api/1/service_links/{{ record.service_link_id }}` - kind `delete`;
  body type `none`; path fields `service_link_id`; required record fields `service_link_id`;
  accepted fields `service_link_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: Mutates Rollbar by deleting/removing/canceling
  /api/1/service_links/{id}; destructive external mutation, approval required.
- `delete_session_replay`: DELETE `/api/1/environment/{{ record.environment }}/session/{{
  record.session_id }}/replay/{{ record.replay_id }}` - kind `delete`; body type `none`; path fields
  `environment`, `session_id`, `replay_id`; required record fields `environment`, `session_id`,
  `replay_id`; accepted fields `environment`, `replay_id`, `session_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: Mutates Rollbar by
  deleting/removing/canceling
  /api/1/environment/{environment}/session/{sessionId}/replay/{replayId}; destructive external
  mutation, approval required.
- `delete_slack_notification_rule`: DELETE `/api/1/notifications/slack/rule/{{ record.rule_id }}` -
  kind `delete`; body type `none`; path fields `rule_id`; required record fields `rule_id`; accepted
  fields `rule_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: Mutates Rollbar by deleting/removing/canceling /api/1/notifications/slack/rule/{rule_id};
  destructive external mutation, approval required.
- `delete_team`: DELETE `/api/1/team/{{ record.team_id }}` - kind `delete`; body type `none`; path
  fields `team_id`; required record fields `team_id`; accepted fields `team_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: Mutates Rollbar by
  deleting/removing/canceling /api/1/team/{team_id}; destructive external mutation, approval
  required.
- `delete_webhook_notification_rule`: DELETE `/api/1/notifications/webhook/rule/{{ record.rule_id
  }}` - kind `delete`; body type `none`; path fields `rule_id`; required record fields `rule_id`;
  accepted fields `rule_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: Mutates Rollbar by deleting/removing/canceling
  /api/1/notifications/webhook/rule/{rule_id}; destructive external mutation, approval required.
- `invite_team_user`: POST `/api/1/team/{{ record.team_id }}/invites` - kind `create`; body type
  `json`; path fields `team_id`; body fields `email`; required record fields `team_id`; accepted
  fields `email`, `team_id`; risk: Mutates Rollbar through POST /api/1/team/{team_id}/invites;
  external API write, approval required.
- `remove_team_from_project`: DELETE `/api/1/team/{{ record.team_id }}/project/{{ record.project_id
  }}` - kind `delete`; body type `none`; path fields `team_id`, `project_id`; required record fields
  `team_id`, `project_id`; accepted fields `project_id`, `team_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: Mutates Rollbar by
  deleting/removing/canceling /api/1/team/{team_id}/project/{project_id}; destructive external
  mutation, approval required.
- `remove_user_from_team`: DELETE `/api/1/team/{{ record.team_id }}/user/{{ record.user_id }}` -
  kind `delete`; body type `none`; path fields `team_id`, `user_id`; required record fields
  `team_id`, `user_id`; accepted fields `team_id`, `user_id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: Mutates Rollbar by deleting/removing/canceling
  /api/1/team/{team_id}/user/{user_id}; destructive external mutation, approval required.
- `update_deploy`: PATCH `/api/1/deploy/{{ record.deploy_id }}` - kind `update`; body type `json`;
  path fields `deploy_id`; body fields `environment`, `revision`, `rollbar_username`,
  `local_username`, `comment`, `status`; required record fields `deploy_id`; accepted fields
  `comment`, `deploy_id`, `environment`, `local_username`, `revision`, `rollbar_username`, `status`;
  risk: Mutates Rollbar through PATCH /api/1/deploy/{deploy_id}; external API write, approval
  required.
- `update_email_notification_rule`: PUT `/api/1/notifications/email/rule/{{ record.rule_id }}` -
  kind `update`; body type `json`; path fields `rule_id`; body fields `trigger`, `status`,
  `filters`, `config`; required record fields `rule_id`; accepted fields `config`, `filters`,
  `rule_id`, `status`, `trigger`; risk: Mutates Rollbar through PUT
  /api/1/notifications/email/rule/{rule_id}; external API write, approval required.
- `update_item`: PATCH `/api/1/item/{{ record.item_id }}` - kind `update`; body type `json`; path
  fields `item_id`; body fields `status`, `resolved_in_version`, `title`, `level`,
  `assigned_user_id`, `assigned_team_id`, `snooze_enabled`, `snooze_expiration_in_seconds`; required
  record fields `item_id`; accepted fields `assigned_team_id`, `assigned_user_id`, `item_id`,
  `level`, `resolved_in_version`, `snooze_enabled`, `snooze_expiration_in_seconds`, `status`,
  `title`; risk: Mutates Rollbar through PATCH /api/1/item/{itemid}; external API write, approval
  required.
- `update_pagerduty_notification_rule`: PUT `/api/1/notifications/pagerduty/rule/{{ record.rule_id
  }}` - kind `update`; body type `json`; path fields `rule_id`; body fields `trigger`, `status`,
  `filters`, `config`; required record fields `rule_id`; accepted fields `config`, `filters`,
  `rule_id`, `status`, `trigger`; risk: Mutates Rollbar through PUT
  /api/1/notifications/pagerduty/rule/{rule_id}; external API write, approval required.
- `update_service_link`: PUT `/api/1/service_links/{{ record.service_link_id }}` - kind `update`;
  body type `json`; path fields `service_link_id`; body fields `name`, `template`; required record
  fields `service_link_id`; accepted fields `name`, `service_link_id`, `template`; risk: Mutates
  Rollbar through PUT /api/1/service_links/{id}; external API write, approval required.
- `update_slack_notification_rule`: PUT `/api/1/notifications/slack/rule/{{ record.rule_id }}` -
  kind `update`; body type `json`; path fields `rule_id`; body fields `trigger`, `status`,
  `filters`, `config`; required record fields `rule_id`; accepted fields `config`, `filters`,
  `rule_id`, `status`, `trigger`; risk: Mutates Rollbar through PUT
  /api/1/notifications/slack/rule/{rule_id}; external API write, approval required.
- `update_webhook_notification_rule`: PUT `/api/1/notifications/webhook/rule/{{ record.rule_id }}` -
  kind `update`; body type `json`; path fields `rule_id`; body fields `trigger`, `status`,
  `filters`, `config`; required record fields `rule_id`; accepted fields `config`, `filters`,
  `rule_id`, `status`, `trigger`; risk: Mutates Rollbar through PUT
  /api/1/notifications/webhook/rule/{rule_id}; external API write, approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 45 stream-backed endpoint group(s), 29 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=4, non_data_endpoint=1, out_of_scope=13, requires_elevated_scope=9.
