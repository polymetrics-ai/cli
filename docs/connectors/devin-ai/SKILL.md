---
name: pm-devin-ai
description: Devin AI connector knowledge and safe action guide.
---

# pm-devin-ai

## Purpose

Reads Devin AI sessions, session child resources, playbooks, knowledge notes, repositories, schedules, membership, metrics, consumption, and secret metadata through the Devin v3 REST API; writes documented organization-scoped JSON mutations.

## Icon

- id: devin-ai
- asset: icons/devin-ai.svg
- source: upstream_registry
- review_status: upstream_seeded

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- metrics_time_after
- metrics_time_before
- mode
- org_id (required)
- page_size
- repository_filter_name
- start_date
- user_email
- api_token (secret) (required)

## ETL Streams

- sessions:
  - primary key: session_id
  - cursor: created_at
  - fields: acus_consumed(number), category(string), child_session_ids(array), created_at(integer), is_archived(boolean), org_id(string), origin(string), parent_session_id(string), playbook_id(string), pull_requests(array), service_user_id(string), session_id(string), status(string), status_detail(string), structured_output(object), subcategory(string), tags(array), title(string), updated_at(integer), url(string), user_id(string)
- sessions_insights:
  - primary key: session_id
  - cursor: created_at
  - fields: acus_consumed(number), analysis(object), category(string), created_at(integer), is_archived(boolean), message_count(integer), num_devin_messages(integer), num_user_messages(integer), org_id(string), origin(string), playbook_id(string), pull_requests(array), service_user_id(string), session_id(string), session_size(string), status(string), status_detail(string), subcategory(string), summary(string), tags(array), title(string), updated_at(integer), url(string), user_id(string)
- session_details:
  - primary key: session_id
  - cursor: created_at
  - fields: acus_consumed(number), category(string), child_session_ids(array), created_at(integer), is_archived(boolean), org_id(string), origin(string), parent_session_id(string), playbook_id(string), pull_requests(array), service_user_id(string), session_id(string), status(string), status_detail(string), structured_output(object), subcategory(string), tags(array), title(string), updated_at(integer), url(string), user_id(string)
- session_messages:
  - primary key: message_id
  - cursor: created_at
  - fields: content(string), created_at(integer), event_id(string), message(string), message_id(string), role(string), session_id(string), source(string), type(string)
- session_attachments:
  - primary key: attachment_id
  - fields: attachment_id(string), content_type(string), name(string), session_id(string), source(string), url(string)
- session_tags:
  - primary key: session_id
  - fields: session_id(string), tags(array)
- playbooks:
  - primary key: playbook_id
  - fields: access_type(string), body(string), created_at(integer), created_by(string), description(string), macro(string), name(string), org_id(string), playbook_id(string), structured_output_schema(object), title(string), updated_at(integer), updated_by(string)
- secrets:
  - primary key: secret_id
  - fields: access_type(string), created_at(integer), created_by(string), is_sensitive(boolean), key(string), name(string), note(string), secret_id(string), secret_type(string), type(string), updated_at(integer), updated_by(string)
- knowledge_notes:
  - primary key: note_id
  - fields: access_type(string), body(string), created_at(integer), folder_id(string), folder_path(string), is_enabled(boolean), macro(string), name(string), note_id(string), org_id(string), pinned_repo(string), trigger(string), updated_at(integer)
- knowledge_folders:
  - primary key: folder_id
  - fields: folder_id(string), name(string), note_count(integer), parent_folder_id(string), path(string)
- repositories:
  - primary key: repo_path
  - fields: git_connection_host(string), git_connection_id(string), indexing_status(object), last_updated_at(integer), provider_repository_id(string), repo_description(string), repo_language(string), repo_name(string), repo_path(string)
- indexed_repositories:
  - primary key: repository_path
  - fields: branches(array), indexing_enabled(boolean), indexing_status(object), repository_path(string)
- schedules:
  - primary key: scheduled_session_id
  - fields: agent(string), bypass_approval(boolean), consecutive_failures(integer), created_at(string), created_by(string), enabled(boolean), frequency(string), interval_count(integer), last_edited_by(string), last_error_at(string), last_error_message(string), last_executed_at(string), name(string), notify_on(string), org_id(string), platform(string), playbook(object), prompt(string), schedule_type(string), scheduled_at(string), scheduled_session_id(string), tags(array), target_devin_id(string), updated_at(string)
- organization_users:
  - primary key: user_id
  - fields: email(string), name(string), role_assignments(array), user_id(string)
- organization_idp_group_users:
  - primary key: user_id
  - fields: email(string), idp_role_assignments(array), name(string), user_id(string)
- self:
  - primary key: principal_type
  - fields: api_key_id(string), api_key_name(string), creator_service_user_id(string), devin_id(string), org_id(string), principal_type(string), service_user_id(string), service_user_name(string), user_id(string), user_name(string)
- org_daily_consumption:
  - primary key: metric
  - fields: metric(string), session_id(string), user_id(string)
- session_daily_consumption:
  - primary key: metric
  - fields: metric(string), session_id(string), user_id(string)
- user_daily_consumption:
  - primary key: metric
  - fields: metric(string), session_id(string), user_id(string)
- org_usage_metrics:
  - primary key: metric
  - fields: metric(string), session_id(string), user_id(string)
- org_session_metrics:
  - primary key: metric
  - fields: metric(string), session_id(string), user_id(string)
- org_active_users_metrics:
  - primary key: metric
  - fields: metric(string), session_id(string), user_id(string)
- org_daily_active_users:
  - primary key: metric
  - fields: metric(string), session_id(string), user_id(string)
- org_monthly_active_users:
  - primary key: metric
  - fields: metric(string), session_id(string), user_id(string)
- org_pr_metrics:
  - primary key: metric
  - fields: metric(string), session_id(string), user_id(string)
- org_search_metrics:
  - primary key: metric
  - fields: metric(string), session_id(string), user_id(string)
- org_weekly_active_users:
  - primary key: metric
  - fields: metric(string), session_id(string), user_id(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- create_session:
  - endpoint: POST /v3/organizations/{{ config.org_id }}/sessions
  - required fields: prompt
  - risk: creates a new Devin session in the organization and can consume ACUs
- send_session_message:
  - endpoint: POST /v3/organizations/{{ config.org_id }}/sessions/{{ record.devin_id }}/messages
  - required fields: devin_id, message
  - risk: sends a message to an active or suspended Devin session and may resume work
- append_session_tags:
  - endpoint: POST /v3/organizations/{{ config.org_id }}/sessions/{{ record.devin_id }}/tags
  - required fields: devin_id, tags
  - risk: adds tags to a Devin session
- replace_session_tags:
  - endpoint: PUT /v3/organizations/{{ config.org_id }}/sessions/{{ record.devin_id }}/tags
  - required fields: devin_id, tags
  - risk: replaces all tags on a Devin session
- archive_session:
  - endpoint: POST /v3/organizations/{{ config.org_id }}/sessions/{{ record.devin_id }}/archive
  - required fields: devin_id
  - risk: archives a Devin session and puts it to sleep if currently running
- terminate_session:
  - endpoint: DELETE /v3/organizations/{{ config.org_id }}/sessions/{{ record.devin_id }}
  - required fields: devin_id
  - risk: terminates a Devin session
- generate_session_insights:
  - endpoint: POST /v3/organizations/{{ config.org_id }}/sessions/{{ record.devin_id }}/insights/generate
  - required fields: devin_id
  - risk: triggers on-demand generation of session insights
- create_schedule:
  - endpoint: POST /v3/organizations/{{ config.org_id }}/schedules
  - required fields: name, prompt
  - risk: creates a scheduled Devin session that can run automatically
- update_schedule:
  - endpoint: PATCH /v3/organizations/{{ config.org_id }}/schedules/{{ record.schedule_id }}
  - required fields: schedule_id
  - risk: updates an existing scheduled Devin session
- delete_schedule:
  - endpoint: DELETE /v3/organizations/{{ config.org_id }}/schedules/{{ record.schedule_id }}
  - required fields: schedule_id
  - risk: soft-deletes a schedule
- create_playbook:
  - endpoint: POST /v3/organizations/{{ config.org_id }}/playbooks
  - required fields: title, body
  - risk: creates an organization-level Devin playbook
- update_playbook:
  - endpoint: PUT /v3/organizations/{{ config.org_id }}/playbooks/{{ record.playbook_id }}
  - required fields: playbook_id, title, body
  - risk: replaces an organization-level Devin playbook
- delete_playbook:
  - endpoint: DELETE /v3/organizations/{{ config.org_id }}/playbooks/{{ record.playbook_id }}
  - required fields: playbook_id
  - risk: deletes an organization-level Devin playbook
- create_knowledge_note:
  - endpoint: POST /v3/organizations/{{ config.org_id }}/knowledge/notes
  - required fields: name, body
  - risk: creates an organization-level Devin knowledge note
- update_knowledge_note:
  - endpoint: PUT /v3/organizations/{{ config.org_id }}/knowledge/notes/{{ record.note_id }}
  - required fields: note_id, name, body
  - risk: replaces an organization-level Devin knowledge note
- delete_knowledge_note:
  - endpoint: DELETE /v3/organizations/{{ config.org_id }}/knowledge/notes/{{ record.note_id }}
  - required fields: note_id
  - risk: deletes an organization-level Devin knowledge note
- index_repository:
  - endpoint: PUT /v3beta1/organizations/{{ config.org_id }}/repositories/{{ record.encoded_repository_path }}/indexing
  - required fields: encoded_repository_path
  - risk: enables indexing for a repository and can trigger indexing jobs
- bulk_index_repositories:
  - endpoint: PUT /v3beta1/organizations/{{ config.org_id }}/repositories/indexing
  - required fields: repositories
  - risk: enables indexing for multiple repositories and can trigger indexing jobs
- remove_repository_indexing:
  - endpoint: DELETE /v3beta1/organizations/{{ config.org_id }}/repositories/{{ record.encoded_repository_path }}/indexing
  - required fields: encoded_repository_path
  - risk: disables indexing and clears configured branches for a repository
- bulk_remove_repository_indexing:
  - endpoint: DELETE /v3beta1/organizations/{{ config.org_id }}/repositories/indexing
  - required fields: repository_paths
  - risk: disables indexing and clears configured branches for multiple repositories
- remove_repository_branch_indexing:
  - endpoint: DELETE /v3beta1/organizations/{{ config.org_id }}/repositories/{{ record.encoded_repository_path }}/indexing/branches/{{ record.encoded_branch_name }}
  - required fields: encoded_repository_path, encoded_branch_name
  - risk: removes one branch from repository indexing and can disable indexing if no branches remain
- trigger_pr_review:
  - endpoint: POST /v3/organizations/{{ config.org_id }}/pr-reviews
  - required fields: pr_url
  - risk: triggers a Devin Review for a pull or merge request
- delete_secret:
  - endpoint: DELETE /v3/organizations/{{ config.org_id }}/secrets/{{ record.secret_id }}
  - required fields: secret_id
  - risk: deletes Devin secret metadata and its stored value from the organization

## Security

- read risk: external Devin AI API reads of organization-scoped sessions, context, repositories, schedules, membership, usage, and metadata
- write risk: creates or mutates Devin sessions, session tags/messages, schedules, playbooks, knowledge notes, repository indexing state, and PR reviews; destructive actions can terminate sessions or delete objects
- approval: reverse ETL writes require plan preview and approval token
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect devin-ai
```

### Inspect as structured JSON

```bash
pm connectors inspect devin-ai --json
```

## Agent Rules

- Run pm connectors inspect devin-ai before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
