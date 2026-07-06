# pm connectors inspect salesflare

```text
NAME
  pm connectors inspect salesflare - Salesflare connector manual

SYNOPSIS
  pm connectors inspect salesflare
  pm connectors inspect salesflare --json
  pm credentials add <name> --connector salesflare [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Salesflare accounts, contacts, opportunities, users, tags, tasks, workflows, groups, stages, pipelines, persons, currencies, custom-field types, and email data sources, and writes CRM lifecycle mutations, through the Salesflare REST API.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  api_key (secret)

ETL STREAMS
  accounts:
    primary key: id
    fields: city(), country(), created_at(), domain(), email(), id(), name(), phone_number(), updated_at()
  contacts:
    primary key: id
    fields: account_id(), created_at(), email(), first_name(), id(), last_name(), name(), phone_number(), updated_at()
  opportunities:
    primary key: id
    fields: account_id(), created_at(), currency(), email(), id(), name(), stage_id(), updated_at(), value()
  users:
    primary key: id
    fields: email(), enabled(), id(), name()
  tags:
    primary key: id
    fields: id(), name()
  tasks:
    primary key: id
    fields: account_id(), assignee_id(), completed(), description(), due_date(), id(), name()
  workflows:
    primary key: id
    fields: id(), name(), state()
  groups:
    primary key: id
    fields: id(), name()
  stages:
    primary key: id
    fields: id(), name(), pipeline_id()
  pipelines:
    primary key: id
    fields: id(), name()
  persons:
    primary key: id
    fields: email(), id(), name()
  currencies:
    primary key: code
    fields: code(), name()
  custom_field_types:
    primary key: type
    fields: name(), type()
  email_data_sources:
    primary key: id
    fields: email(), id()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_account:
    endpoint: POST /accounts
    risk: creates a new CRM account; low-risk external mutation, no approval required
  update_account:
    endpoint: PUT /accounts/{{ record.id }}
    required fields: id
    risk: external mutation updating a CRM account; approval required
  delete_account:
    endpoint: DELETE /accounts/{{ record.id }}
    required fields: id
    risk: destructive/irreversible: permanently deletes a CRM account; approval required
  create_contact:
    endpoint: POST /contacts
    risk: creates a new CRM contact; low-risk external mutation, no approval required
  update_contact:
    endpoint: PUT /contacts/{{ record.id }}
    required fields: id
    risk: external mutation updating a CRM contact; approval required
  delete_contact:
    endpoint: DELETE /contacts/{{ record.id }}
    required fields: id
    risk: destructive/irreversible: permanently deletes a CRM contact; approval required
  create_opportunity:
    endpoint: POST /opportunities
    risk: creates a new CRM opportunity/deal; low-risk external mutation, no approval required
  update_opportunity:
    endpoint: PUT /opportunities/{{ record.id }}
    required fields: id
    risk: external mutation updating a CRM opportunity/deal (may change stage/close state); approval required
  delete_opportunity:
    endpoint: DELETE /opportunities/{{ record.id }}
    required fields: id
    risk: destructive/irreversible: permanently deletes a CRM opportunity/deal; approval required
  create_tag:
    endpoint: POST /tags
    risk: creates a new CRM tag; low-risk external mutation, no approval required
  update_tag:
    endpoint: PUT /tags/{{ record.id }}
    required fields: id
    risk: external mutation renaming a CRM tag; approval required
  delete_tag:
    endpoint: DELETE /tags/{{ record.id }}
    required fields: id
    risk: destructive/irreversible: permanently deletes a CRM tag from every record it's applied to; approval required
  create_task:
    endpoint: POST /tasks
    risk: creates a new CRM task; low-risk external mutation, no approval required
  update_task:
    endpoint: PUT /tasks/{{ record.id }}
    required fields: id
    risk: external mutation updating a CRM task (may mark complete); approval required
  delete_task:
    endpoint: DELETE /tasks/{{ record.id }}
    required fields: id
    risk: destructive/irreversible: permanently deletes a CRM task; approval required
  create_meeting:
    endpoint: POST /meetings
    risk: creates a new CRM meeting/calendar entry; low-risk external mutation, no approval required
  update_meeting:
    endpoint: PUT /meetings/{{ record.meeting_id }}
    required fields: meeting_id
    risk: external mutation updating a CRM meeting/calendar entry; approval required
  delete_meeting:
    endpoint: DELETE /meetings/{{ record.meeting_id }}
    required fields: meeting_id
    risk: destructive/irreversible: permanently deletes a CRM meeting/calendar entry; approval required
  create_call:
    endpoint: POST /calls
    risk: logs a new call activity against a CRM account; low-risk external mutation, no approval required
  create_internal_note:
    endpoint: POST /messages
    risk: creates a new internal note on a CRM record; low-risk external mutation, no approval required
  update_internal_note:
    endpoint: PUT /messages/{{ record.message_id }}
    required fields: message_id
    risk: external mutation editing a CRM internal note; approval required
  delete_internal_note:
    endpoint: DELETE /messages/{{ record.message_id }}
    required fields: message_id
    risk: destructive/irreversible: permanently deletes a CRM internal note; approval required

SECURITY
  read risk: external Salesflare API read of CRM account, contact, opportunity, task, workflow, and reference data
  write risk: external Salesflare mutations: account/contact/opportunity/tag/task/meeting/call/internal-note create-update-delete
  approval: required for update/delete actions; create_* actions on accounts/contacts/opportunities/tags/tasks/meetings/calls/internal notes are low-risk and do not require approval; delete_* actions are destructive and irreversible
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect salesflare

  # Inspect as structured JSON
  pm connectors inspect salesflare --json

AGENT WORKFLOW
  - Run pm connectors inspect salesflare before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
