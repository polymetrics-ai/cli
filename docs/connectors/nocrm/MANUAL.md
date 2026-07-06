# pm connectors inspect nocrm

```text
NAME
  pm connectors inspect nocrm - noCRM connector manual

SYNOPSIS
  pm connectors inspect nocrm
  pm connectors inspect nocrm --json
  pm credentials add <name> --connector nocrm [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads noCRM.io CRM objects and exposes declarative write actions for supported noCRM API v2 mutations.

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
  attachment_id
  base_url
  client_id
  lead_id
  prospecting_list_id
  step_id
  team_id
  user_id
  webhook_event_id
  api_key (secret)

ETL STREAMS
  leads:
    primary key: id
    fields: amount(), client_folder_id(), closed_at(), created_at(), currency(), id(), pipeline(), pipeline_id(), probability(), remind_date(), status(), step(), step_id(), title(), updated_at(), user_id()
  pipelines:
    primary key: id
    fields: created_at(), default(), id(), name(), position(), updated_at()
  users:
    primary key: id
    fields: active(), admin(), created_at(), email(), firstname(), id(), lastname(), team_id(), updated_at()
  teams:
    primary key: id
    fields: created_at(), id(), name(), updated_at()
  prospecting_lists:
    primary key: id
    fields: archived(), created_at(), id(), prospects_count(), title(), updated_at(), user_id()
  steps:
    primary key: id
    fields: created_at(), id(), name(), pipeline_id(), position(), updated_at()
  step:
    primary key: id
    fields: created_at(), id(), name(), pipeline_id(), position(), updated_at()
  client_folders:
    primary key: id
    fields: created_at(), description(), id(), name(), updated_at(), user_id()
  client_folder:
    primary key: id
    fields: created_at(), description(), id(), name(), updated_at(), user_id()
  categories:
    primary key: id
    fields: id(), name(), tags()
  predefined_tags:
    primary key: id
    fields: category_id(), id(), name()
  fields:
    primary key: id
    fields: id(), key(), name(), type()
  activities:
    primary key: id
    fields: created_at(), id(), name(), updated_at()
  lead:
    primary key: id
    fields: created_at(), id(), status(), title(), updated_at()
  unassigned_leads:
    primary key: id
    fields: created_at(), id(), status(), title(), updated_at()
  lead_comments:
    primary key: id
    fields: content(), created_at(), id(), updated_at(), user_id()
  lead_duplicates:
    primary key: id
    fields: id(), status(), title()
  lead_attachments:
    primary key: id
    fields: created_at(), id(), name(), url()
  lead_attachment:
    primary key: id
    fields: created_at(), id(), name(), url()
  lead_action_histories:
    primary key: id
    fields: action(), created_at(), id(), user_id()
  post_sales_tasks:
    primary key: id
    fields: created_at(), id(), status(), title(), updated_at()
  spreadsheets:
    primary key: id
    fields: created_at(), id(), title(), updated_at(), user_id()
  spreadsheet:
    primary key: id
    fields: created_at(), id(), title(), updated_at(), user_id()
  prospects:
    primary key: id
    fields: created_at(), id(), spreadsheet_id(), updated_at()
  prospects_called:
    primary key: id
    fields: created_at(), id(), spreadsheet_id(), updated_at()
  user:
    primary key: id
    fields: active(), email(), firstname(), id(), lastname()
  team:
    primary key: id
    fields: created_at(), id(), name(), updated_at()
  webhooks:
    primary key: id
    fields: active(), event(), id(), target()
  webhook_events:
    primary key: id
    fields: data(), event(), has_succeeded(), id()
  webhook_event:
    primary key: id
    fields: data(), event(), has_succeeded(), id()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_client_folder:
    endpoint: POST /clients
    risk: creates a noCRM client folder visible to account users; external mutation, approval required
  update_client_folder:
    endpoint: PUT /clients/{{ record.id }}
    required fields: id
    risk: updates a noCRM client folder; external mutation, approval required
  delete_client_folder:
    endpoint: DELETE /clients/{{ record.id }}
    required fields: id
    risk: deletes a noCRM client folder; destructive external mutation, approval required
  create_category:
    endpoint: POST /category
    risk: creates a noCRM category/tag grouping; account configuration mutation, approval required
  create_predefined_tag:
    endpoint: POST /predefined_tags
    risk: creates a predefined tag available in noCRM; account taxonomy mutation, approval required
  create_field:
    endpoint: POST /fields
    risk: creates a noCRM custom field; account schema mutation, approval required
  create_lead:
    endpoint: POST /leads
    risk: creates a live noCRM lead and may assign ownership/tags/step; external CRM mutation, approval required
  duplicate_lead:
    endpoint: POST /leads/{{ record.id }}/duplicate_lead
    required fields: id
    risk: duplicates a live noCRM lead; external CRM mutation, approval required
  update_lead:
    endpoint: PUT /leads/{{ record.id }}
    required fields: id
    risk: updates a live noCRM lead; external CRM mutation, approval required
  assign_lead:
    endpoint: POST /leads/{{ record.id }}/assign
    required fields: id
    risk: assigns a live noCRM lead to a user and can trigger notifications; external CRM mutation, approval required
  add_lead_to_client_folder:
    endpoint: POST /leads/{{ record.id }}/add_to_client
    required fields: id
    risk: links a live lead to a client folder; external CRM mutation, approval required
  delete_lead:
    endpoint: DELETE /leads/{{ record.id }}
    required fields: id
    risk: deletes a live noCRM lead; destructive external mutation, approval required
  delete_multiple_leads:
    endpoint: DELETE /leads/delete_multiple
    optional fields: ids
    risk: bulk-deletes live noCRM leads; destructive external mutation, approval required
  create_lead_comment:
    endpoint: POST /leads/{{ record.lead_id }}/comments
    required fields: lead_id
    risk: adds a comment to a live noCRM lead; external CRM mutation, approval required
  update_lead_comment:
    endpoint: PUT /leads/{{ record.lead_id }}/comments/{{ record.id }}
    required fields: lead_id, id
    risk: updates a lead comment in noCRM; external CRM mutation, approval required
  delete_lead_comment:
    endpoint: DELETE /leads/{{ record.lead_id }}/comments/{{ record.id }}
    required fields: lead_id, id
    risk: deletes a noCRM lead comment; destructive external mutation, approval required
  delete_lead_attachment:
    endpoint: DELETE /leads/{{ record.lead_id }}/attachments/{{ record.id }}
    required fields: lead_id, id
    risk: deletes an attachment from a live noCRM lead; destructive external mutation, approval required
  send_lead_email_from_template:
    endpoint: POST /leads/{{ record.lead_id }}/emails/send_email_from_template
    required fields: lead_id
    risk: sends an email from a noCRM template to a lead; external communication mutation, approval required
  create_lead_follow_up_from_template:
    endpoint: POST /leads/{{ record.lead_id }}/follow_ups/create_from_template
    required fields: lead_id
    risk: creates post-sales tasks for a live noCRM lead; workflow mutation, approval required
  create_prospecting_list:
    endpoint: POST /spreadsheets
    risk: creates a noCRM prospecting list and optional rows/tags/owner; external CRM mutation, approval required
  assign_prospecting_list:
    endpoint: POST /spreadsheets/{{ record.id }}/assign
    required fields: id
    risk: assigns a prospecting list to a user; external CRM mutation, approval required
  create_prospecting_list_comment:
    endpoint: POST /spreadsheets/{{ record.spreadsheet_id }}/comments
    required fields: spreadsheet_id
    risk: adds a comment to a noCRM prospecting list; external CRM mutation, approval required
  create_prospect_comment:
    endpoint: POST /spreadsheets/{{ record.spreadsheet_id }}/rows/{{ record.prospect_id }}/comments
    required fields: spreadsheet_id, prospect_id
    risk: adds a comment to a prospect row; external CRM mutation, approval required
  update_prospect_comment:
    endpoint: PUT /spreadsheets/{{ record.spreadsheet_id }}/rows/{{ record.prospect_id }}/comments/{{ record.id }}
    required fields: spreadsheet_id, prospect_id, id
    risk: updates a prospect-row comment; external CRM mutation, approval required
  create_prospects:
    endpoint: POST /spreadsheets/{{ record.spreadsheet_id }}/rows
    required fields: spreadsheet_id
    risk: adds prospect rows to a noCRM prospecting list; external CRM mutation, approval required
  update_prospect_fields:
    endpoint: PUT /spreadsheets/{{ record.spreadsheet_id }}/rows/{{ record.id }}/update_fields
    required fields: spreadsheet_id, id
    risk: updates named fields on a prospect row; external CRM mutation, approval required
  create_lead_from_prospect:
    endpoint: POST /spreadsheets/{{ record.spreadsheet_id }}/rows/{{ record.id }}/create_lead
    required fields: spreadsheet_id, id
    risk: converts a prospect row into a live noCRM lead; external CRM mutation, approval required
  delete_prospect:
    endpoint: DELETE /spreadsheets/{{ record.spreadsheet_id }}/rows/{{ record.id }}
    required fields: spreadsheet_id, id
    risk: deletes a prospect row from a prospecting list; destructive external mutation, approval required
  create_user:
    endpoint: POST /users
    risk: creates a noCRM user account and can send activation email depending on payload; administrative mutation, approval required
  disable_user:
    endpoint: PUT /users/{{ record.id }}/disable
    required fields: id
    risk: disables a noCRM user account; administrative mutation, approval required
  create_team:
    endpoint: POST /teams
    risk: creates a noCRM team; administrative mutation, approval required
  update_team:
    endpoint: PUT /teams/{{ record.id }}
    required fields: id
    risk: updates a noCRM team; administrative mutation, approval required
  delete_team:
    endpoint: DELETE /teams/{{ record.id }}
    required fields: id
    risk: deletes a noCRM team; destructive administrative mutation, approval required
  add_team_member:
    endpoint: POST /teams/{{ record.id }}/add_member
    required fields: id
    risk: adds a user to a noCRM team and can change manager status; administrative mutation, approval required
  remove_team_member:
    endpoint: DELETE /teams/{{ record.id }}/remove_member
    required fields: id
    optional fields: user_id
    risk: removes a user from a noCRM team; administrative mutation, approval required
  create_webhook:
    endpoint: POST /webhooks
    risk: creates a noCRM webhook/notification destination; outbound data delivery mutation, approval required
  activate_webhook:
    endpoint: PUT /webhooks/{{ record.id }}/activate
    required fields: id
    risk: activates a noCRM webhook/notification destination; outbound data delivery mutation, approval required
  delete_webhook:
    endpoint: DELETE /webhooks/{{ record.id }}
    required fields: id
    risk: disables or removes a noCRM webhook/notification destination; external mutation, approval required

SECURITY
  read risk: external noCRM API read of sales lead and pipeline data
  write risk: external noCRM API mutations can create, update, assign, email, disable, or delete live CRM/admin records
  approval: write actions require explicit approval; reads require none
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect nocrm

  # Inspect as structured JSON
  pm connectors inspect nocrm --json

AGENT WORKFLOW
  - Run pm connectors inspect nocrm before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
