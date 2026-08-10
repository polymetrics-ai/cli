---
name: pm-rollbar
description: Rollbar connector knowledge and safe action guide.
---

# pm-rollbar

## Purpose

Reads and writes documented Rollbar API v1 resources through the declarative connector engine.

## Icon

- id: simple-icons-rollbar
- asset: icons/simple-icons/rollbar.svg
- title: Rollbar
- simple_icon_slug: rollbar
- simple_icon_hex: 3569F3
- source: simple-icons
- license: CC0-1.0
- review_status: cc0_with_trademark_caveat
- review_url: https://simpleicons.org/?q=Rollbar
- match: exact-name-or-slug
- matched_by: rollbar

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- counter
- deploy_id
- environment
- event
- instance_id
- invite_id
- item_id
- job_id
- project_id
- replay_id
- rule_id
- service_link_id
- session_id
- team_id
- user_id
- uuid
- version
- access_token (secret) (required)

## ETL Streams

- activated_counts_report:
  - fields: err(integer), id(integer), name(string), result(), status(string), title(string)
- deploy:
  - fields: err(integer), id(integer), name(string), result(), status(string), title(string)
- deploys:
  - fields: err(integer), id(integer), name(string), result(), status(string), title(string)
- email_notification_rule:
  - fields: err(integer), id(integer), name(string), result(), status(string), title(string)
- email_notification_rules:
  - fields: err(integer), id(integer), name(string), result(), status(string), title(string)
- environments:
  - fields: err(integer), id(integer), name(string), result(), status(string), title(string)
- invitation:
  - fields: err(integer), id(integer), name(string), result(), status(string), title(string)
- item_by_counter:
  - fields: err(integer), id(integer), name(string), result(), status(string), title(string)
- item_by_id:
  - fields: err(integer), id(integer), name(string), result(), status(string), title(string)
- item_by_uuid:
  - fields: err(integer), id(integer), name(string), result(), status(string), title(string)
- item_occurrences:
  - fields: err(integer), id(integer), name(string), result(), status(string), title(string)
- items:
  - fields: err(integer), id(integer), name(string), result(), status(string), title(string)
- occurrence:
  - fields: err(integer), id(integer), name(string), result(), status(string), title(string)
- occurrence_counts_report:
  - fields: err(integer), id(integer), name(string), result(), status(string), title(string)
- occurrences:
  - fields: err(integer), id(integer), name(string), result(), status(string), title(string)
- pagerduty_notification_rule:
  - fields: err(integer), id(integer), name(string), result(), status(string), title(string)
- pagerduty_notification_rules:
  - fields: err(integer), id(integer), name(string), result(), status(string), title(string)
- person_deletion_job:
  - fields: err(integer), id(integer), name(string), result(), status(string), title(string)
- project:
  - fields: err(integer), id(integer), name(string), result(), status(string), title(string)
- project_teams:
  - fields: err(integer), id(integer), name(string), result(), status(string), title(string)
- projects:
  - fields: err(integer), id(integer), name(string), result(), status(string), title(string)
- rql_job:
  - fields: err(integer), id(integer), name(string), result(), status(string), title(string)
- rql_job_result:
  - fields: err(integer), id(integer), name(string), result(), status(string), title(string)
- rql_jobs:
  - fields: err(integer), id(integer), name(string), result(), status(string), title(string)
- service_link:
  - fields: err(integer), id(integer), name(string), result(), status(string), title(string)
- service_links:
  - fields: err(integer), id(integer), name(string), result(), status(string), title(string)
- session_replay:
  - fields: err(integer), id(integer), name(string), result(), status(string), title(string)
- slack_notification_rule:
  - fields: err(integer), id(integer), name(string), result(), status(string), title(string)
- slack_notification_rules:
  - fields: err(integer), id(integer), name(string), result(), status(string), title(string)
- team:
  - fields: err(integer), id(integer), name(string), result(), status(string), title(string)
- team_invitations:
  - fields: err(integer), id(integer), name(string), result(), status(string), title(string)
- team_project_assignment:
  - fields: err(integer), id(integer), name(string), result(), status(string), title(string)
- team_projects:
  - fields: err(integer), id(integer), name(string), result(), status(string), title(string)
- team_user_assignment:
  - fields: err(integer), id(integer), name(string), result(), status(string), title(string)
- team_users:
  - fields: err(integer), id(integer), name(string), result(), status(string), title(string)
- teams:
  - fields: err(integer), id(integer), name(string), result(), status(string), title(string)
- top_active_items_report:
  - fields: err(integer), id(integer), name(string), result(), status(string), title(string)
- user:
  - fields: err(integer), id(integer), name(string), result(), status(string), title(string)
- user_projects:
  - fields: err(integer), id(integer), name(string), result(), status(string), title(string)
- user_teams:
  - fields: err(integer), id(integer), name(string), result(), status(string), title(string)
- users:
  - fields: err(integer), id(integer), name(string), result(), status(string), title(string)
- version:
  - fields: err(integer), id(integer), name(string), result(), status(string), title(string)
- version_items:
  - fields: err(integer), id(integer), name(string), result(), status(string), title(string)
- webhook_notification_rule:
  - fields: err(integer), id(integer), name(string), result(), status(string), title(string)
- webhook_notification_rules:
  - fields: err(integer), id(integer), name(string), result(), status(string), title(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Reverse ETL Actions

- assign_team_to_project:
  - endpoint: PUT /api/1/team/{{ record.team_id }}/project/{{ record.project_id }}
  - required fields: team_id, project_id
  - risk: Changes Rollbar membership/assignment for /api/1/team/{team_id}/project/{project_id}; external access mapping mutation, approval required.
- assign_user_to_team:
  - endpoint: PUT /api/1/team/{{ record.team_id }}/user/{{ record.user_id }}
  - required fields: team_id, user_id
  - risk: Changes Rollbar membership/assignment for /api/1/team/{team_id}/user/{user_id}; external access mapping mutation, approval required.
- cancel_invitation:
  - endpoint: DELETE /api/1/invite/{{ record.invite_id }}
  - required fields: invite_id
  - risk: Mutates Rollbar by deleting/removing/canceling /api/1/invite/{invite_id}; destructive external mutation, approval required.
- cancel_rql_job:
  - endpoint: POST /api/1/rql/job/{{ record.job_id }}/cancel
  - required fields: job_id
  - risk: Mutates Rollbar by deleting/removing/canceling /api/1/rql/job/{job_id}/cancel; destructive external mutation, approval required.
- configure_email_notifications:
  - endpoint: PUT /api/1/notifications/email
  - optional fields: enabled
  - risk: Changes Rollbar notification configuration for /api/1/notifications/email; external configuration mutation, approval required.
- configure_slack_notifications:
  - endpoint: PUT /api/1/notifications/slack
  - optional fields: enabled, channel, show_message_buttons
  - risk: Changes Rollbar notification configuration for /api/1/notifications/slack; external configuration mutation, approval required.
- create_item:
  - endpoint: POST /api/1/item/
  - optional fields: data
  - risk: Mutates Rollbar through POST /api/1/item/; external API write, approval required.
- create_project:
  - endpoint: POST /api/1/projects
  - optional fields: name
  - risk: Mutates Rollbar through POST /api/1/projects; external API write, approval required.
- create_service_link:
  - endpoint: POST /api/1/service_links
  - optional fields: name, template
  - risk: Mutates Rollbar through POST /api/1/service_links; external API write, approval required.
- create_team:
  - endpoint: POST /api/1/teams
  - optional fields: name, access_level
  - risk: Mutates Rollbar through POST /api/1/teams; external API write, approval required.
- delete_email_notification_rule:
  - endpoint: DELETE /api/1/notifications/email/rule/{{ record.rule_id }}
  - required fields: rule_id
  - risk: Mutates Rollbar by deleting/removing/canceling /api/1/notifications/email/rule/{rule_id}; destructive external mutation, approval required.
- delete_occurrence:
  - endpoint: DELETE /api/1/instance/{{ record.instance_id }}
  - required fields: instance_id
  - risk: Mutates Rollbar by deleting/removing/canceling /api/1/instance/{instance_id}; destructive external mutation, approval required.
- delete_pagerduty_notification_rule:
  - endpoint: DELETE /api/1/notifications/pagerduty/rule/{{ record.rule_id }}
  - required fields: rule_id
  - risk: Mutates Rollbar by deleting/removing/canceling /api/1/notifications/pagerduty/rule/{rule_id}; destructive external mutation, approval required.
- delete_project:
  - endpoint: DELETE /api/1/project/{{ record.project_id }}
  - required fields: project_id
  - risk: Mutates Rollbar by deleting/removing/canceling /api/1/project/{project_id}; destructive external mutation, approval required.
- delete_service_link:
  - endpoint: DELETE /api/1/service_links/{{ record.service_link_id }}
  - required fields: service_link_id
  - risk: Mutates Rollbar by deleting/removing/canceling /api/1/service_links/{id}; destructive external mutation, approval required.
- delete_session_replay:
  - endpoint: DELETE /api/1/environment/{{ record.environment }}/session/{{ record.session_id }}/replay/{{ record.replay_id }}
  - required fields: environment, session_id, replay_id
  - risk: Mutates Rollbar by deleting/removing/canceling /api/1/environment/{environment}/session/{sessionId}/replay/{replayId}; destructive external mutation, approval required.
- delete_slack_notification_rule:
  - endpoint: DELETE /api/1/notifications/slack/rule/{{ record.rule_id }}
  - required fields: rule_id
  - risk: Mutates Rollbar by deleting/removing/canceling /api/1/notifications/slack/rule/{rule_id}; destructive external mutation, approval required.
- delete_team:
  - endpoint: DELETE /api/1/team/{{ record.team_id }}
  - required fields: team_id
  - risk: Mutates Rollbar by deleting/removing/canceling /api/1/team/{team_id}; destructive external mutation, approval required.
- delete_webhook_notification_rule:
  - endpoint: DELETE /api/1/notifications/webhook/rule/{{ record.rule_id }}
  - required fields: rule_id
  - risk: Mutates Rollbar by deleting/removing/canceling /api/1/notifications/webhook/rule/{rule_id}; destructive external mutation, approval required.
- invite_team_user:
  - endpoint: POST /api/1/team/{{ record.team_id }}/invites
  - required fields: team_id
  - optional fields: email
  - risk: Mutates Rollbar through POST /api/1/team/{team_id}/invites; external API write, approval required.
- remove_team_from_project:
  - endpoint: DELETE /api/1/team/{{ record.team_id }}/project/{{ record.project_id }}
  - required fields: team_id, project_id
  - risk: Mutates Rollbar by deleting/removing/canceling /api/1/team/{team_id}/project/{project_id}; destructive external mutation, approval required.
- remove_user_from_team:
  - endpoint: DELETE /api/1/team/{{ record.team_id }}/user/{{ record.user_id }}
  - required fields: team_id, user_id
  - risk: Mutates Rollbar by deleting/removing/canceling /api/1/team/{team_id}/user/{user_id}; destructive external mutation, approval required.
- update_deploy:
  - endpoint: PATCH /api/1/deploy/{{ record.deploy_id }}
  - required fields: deploy_id
  - optional fields: environment, revision, rollbar_username, local_username, comment, status
  - risk: Mutates Rollbar through PATCH /api/1/deploy/{deploy_id}; external API write, approval required.
- update_email_notification_rule:
  - endpoint: PUT /api/1/notifications/email/rule/{{ record.rule_id }}
  - required fields: rule_id
  - optional fields: trigger, status, filters, config
  - risk: Mutates Rollbar through PUT /api/1/notifications/email/rule/{rule_id}; external API write, approval required.
- update_item:
  - endpoint: PATCH /api/1/item/{{ record.item_id }}
  - required fields: item_id
  - optional fields: status, resolved_in_version, title, level, assigned_user_id, assigned_team_id, snooze_enabled, snooze_expiration_in_seconds
  - risk: Mutates Rollbar through PATCH /api/1/item/{itemid}; external API write, approval required.
- update_pagerduty_notification_rule:
  - endpoint: PUT /api/1/notifications/pagerduty/rule/{{ record.rule_id }}
  - required fields: rule_id
  - optional fields: trigger, status, filters, config
  - risk: Mutates Rollbar through PUT /api/1/notifications/pagerduty/rule/{rule_id}; external API write, approval required.
- update_service_link:
  - endpoint: PUT /api/1/service_links/{{ record.service_link_id }}
  - required fields: service_link_id
  - optional fields: name, template
  - risk: Mutates Rollbar through PUT /api/1/service_links/{id}; external API write, approval required.
- update_slack_notification_rule:
  - endpoint: PUT /api/1/notifications/slack/rule/{{ record.rule_id }}
  - required fields: rule_id
  - optional fields: trigger, status, filters, config
  - risk: Mutates Rollbar through PUT /api/1/notifications/slack/rule/{rule_id}; external API write, approval required.
- update_webhook_notification_rule:
  - endpoint: PUT /api/1/notifications/webhook/rule/{{ record.rule_id }}
  - required fields: rule_id
  - optional fields: trigger, status, filters, config
  - risk: Mutates Rollbar through PUT /api/1/notifications/webhook/rule/{rule_id}; external API write, approval required.

## Security

- read risk: external Rollbar API read of projects, items, occurrences, users, teams, reports, notification rules, service links, versions, and related operational data
- write risk: external Rollbar API mutations for items, projects, teams, assignments, invitations, notification settings/rules, service links, occurrences, and session replay deletes; approval required before execution
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect rollbar
```

### Inspect as structured JSON

```bash
pm connectors inspect rollbar --json
```

## Agent Rules

- Run pm connectors inspect rollbar before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
