---
name: pm-devin-ai
description: Devin AI connector knowledge and safe action guide.
---

# pm-devin-ai

## Purpose

Reads Devin AI sessions, session child resources, playbooks, knowledge notes, repositories, schedules, membership, metrics, consumption, and secret metadata through the Devin v3 REST API; writes documented organization-scoped JSON mutations.

## Icon

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
- org_id
- page_size
- repository_filter_name
- start_date
- user_email
- api_token (secret)

## ETL Streams

- sessions:
  - primary key: session_id
  - cursor: created_at
  - fields: acus_consumed(), category(), child_session_ids(), created_at(), is_archived(), org_id(), origin(), parent_session_id(), playbook_id(), pull_requests(), service_user_id(), session_id(), status(), status_detail(), structured_output(), subcategory(), tags(), title(), updated_at(), url(), user_id()
- sessions_insights:
  - primary key: session_id
  - cursor: created_at
  - fields: acus_consumed(), analysis(), category(), created_at(), is_archived(), message_count(), num_devin_messages(), num_user_messages(), org_id(), origin(), playbook_id(), pull_requests(), service_user_id(), session_id(), session_size(), status(), status_detail(), subcategory(), summary(), tags(), title(), updated_at(), url(), user_id()
- session_details:
  - primary key: session_id
  - cursor: created_at
  - fields: acus_consumed(), category(), child_session_ids(), created_at(), is_archived(), org_id(), origin(), parent_session_id(), playbook_id(), pull_requests(), service_user_id(), session_id(), status(), status_detail(), structured_output(), subcategory(), tags(), title(), updated_at(), url(), user_id()
- session_messages:
  - primary key: message_id
  - cursor: created_at
  - fields: content(), created_at(), event_id(), message(), message_id(), role(), session_id(), source(), type()
- session_attachments:
  - primary key: attachment_id
  - fields: attachment_id(), content_type(), name(), session_id(), source(), url()
- session_tags:
  - primary key: session_id
  - fields: session_id(), tags()
- playbooks:
  - primary key: playbook_id
  - fields: access_type(), body(), created_at(), created_by(), description(), macro(), name(), org_id(), playbook_id(), structured_output_schema(), title(), updated_at(), updated_by()
- secrets:
  - primary key: secret_id
  - fields: access_type(), created_at(), created_by(), is_sensitive(), key(), name(), note(), secret_id(), secret_type(), type(), updated_at(), updated_by()
- knowledge_notes:
  - primary key: note_id
  - fields: access_type(), body(), created_at(), folder_id(), folder_path(), is_enabled(), macro(), name(), note_id(), org_id(), pinned_repo(), trigger(), updated_at()
- knowledge_folders:
  - primary key: folder_id
  - fields: folder_id(), name(), note_count(), parent_folder_id(), path()
- repositories:
  - primary key: repo_path
  - fields: git_connection_host(), git_connection_id(), indexing_status(), last_updated_at(), provider_repository_id(), repo_description(), repo_language(), repo_name(), repo_path()
- indexed_repositories:
  - primary key: repository_path
  - fields: branches(), indexing_enabled(), indexing_status(), repository_path()
- schedules:
  - primary key: scheduled_session_id
  - fields: agent(), bypass_approval(), consecutive_failures(), created_at(), created_by(), enabled(), frequency(), interval_count(), last_edited_by(), last_error_at(), last_error_message(), last_executed_at(), name(), notify_on(), org_id(), platform(), playbook(), prompt(), schedule_type(), scheduled_at(), scheduled_session_id(), tags(), target_devin_id(), updated_at()
- organization_users:
  - primary key: user_id
  - fields: email(), name(), role_assignments(), user_id()
- organization_idp_group_users:
  - primary key: user_id
  - fields: email(), idp_role_assignments(), name(), user_id()
- self:
  - primary key: principal_type
  - fields: api_key_id(), api_key_name(), creator_service_user_id(), devin_id(), org_id(), principal_type(), service_user_id(), service_user_name(), user_id(), user_name()
- org_daily_consumption:
  - primary key: metric
  - fields: metric(), session_id(), user_id()
- session_daily_consumption:
  - primary key: metric
  - fields: metric(), session_id(), user_id()
- user_daily_consumption:
  - primary key: metric
  - fields: metric(), session_id(), user_id()
- org_usage_metrics:
  - primary key: metric
  - fields: metric(), session_id(), user_id()
- org_session_metrics:
  - primary key: metric
  - fields: metric(), session_id(), user_id()
- org_active_users_metrics:
  - primary key: metric
  - fields: metric(), session_id(), user_id()
- org_daily_active_users:
  - primary key: metric
  - fields: metric(), session_id(), user_id()
- org_monthly_active_users:
  - primary key: metric
  - fields: metric(), session_id(), user_id()
- org_pr_metrics:
  - primary key: metric
  - fields: metric(), session_id(), user_id()
- org_search_metrics:
  - primary key: metric
  - fields: metric(), session_id(), user_id()
- org_weekly_active_users:
  - primary key: metric
  - fields: metric(), session_id(), user_id()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- create_session:
  - endpoint: POST /v3/organizations/{{ config.org_id }}/sessions
  - risk: creates a new Devin session in the organization and can consume ACUs
- send_session_message:
  - endpoint: POST /v3/organizations/{{ config.org_id }}/sessions/{{ record.devin_id }}/messages
  - required fields: devin_id
  - risk: sends a message to an active or suspended Devin session and may resume work
- append_session_tags:
  - endpoint: POST /v3/organizations/{{ config.org_id }}/sessions/{{ record.devin_id }}/tags
  - required fields: devin_id
  - risk: adds tags to a Devin session
- replace_session_tags:
  - endpoint: PUT /v3/organizations/{{ config.org_id }}/sessions/{{ record.devin_id }}/tags
  - required fields: devin_id
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
  - risk: creates an organization-level Devin playbook
- update_playbook:
  - endpoint: PUT /v3/organizations/{{ config.org_id }}/playbooks/{{ record.playbook_id }}
  - required fields: playbook_id
  - risk: replaces an organization-level Devin playbook
- delete_playbook:
  - endpoint: DELETE /v3/organizations/{{ config.org_id }}/playbooks/{{ record.playbook_id }}
  - required fields: playbook_id
  - risk: deletes an organization-level Devin playbook
- create_knowledge_note:
  - endpoint: POST /v3/organizations/{{ config.org_id }}/knowledge/notes
  - risk: creates an organization-level Devin knowledge note
- update_knowledge_note:
  - endpoint: PUT /v3/organizations/{{ config.org_id }}/knowledge/notes/{{ record.note_id }}
  - required fields: note_id
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
  - risk: enables indexing for multiple repositories and can trigger indexing jobs
- remove_repository_indexing:
  - endpoint: DELETE /v3beta1/organizations/{{ config.org_id }}/repositories/{{ record.encoded_repository_path }}/indexing
  - required fields: encoded_repository_path
  - risk: disables indexing and clears configured branches for a repository
- bulk_remove_repository_indexing:
  - endpoint: DELETE /v3beta1/organizations/{{ config.org_id }}/repositories/indexing
  - optional fields: repository_paths
  - risk: disables indexing and clears configured branches for multiple repositories
- remove_repository_branch_indexing:
  - endpoint: DELETE /v3beta1/organizations/{{ config.org_id }}/repositories/{{ record.encoded_repository_path }}/indexing/branches/{{ record.encoded_branch_name }}
  - required fields: encoded_repository_path, encoded_branch_name
  - risk: removes one branch from repository indexing and can disable indexing if no branches remain
- trigger_pr_review:
  - endpoint: POST /v3/organizations/{{ config.org_id }}/pr-reviews
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
