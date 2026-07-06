# pm connectors inspect survey-sparrow

```text
NAME
  pm connectors inspect survey-sparrow - SurveySparrow connector manual

SYNOPSIS
  pm connectors inspect survey-sparrow
  pm connectors inspect survey-sparrow --json
  pm credentials add <name> --connector survey-sparrow [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and manages SurveySparrow surveys, contacts, responses, questions, channels, contact lists/properties, reminders, reputation platforms/reviews, survey folders, ticket fields, tickets, teams, roles, variables, webhooks, users, templates, email themes, and expressions through the SurveySparrow API.

ICON
  asset: icons/surveysparrow.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.surveysparrow.com/

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  survey_id
  access_token (secret)

ETL STREAMS
  surveys:
    primary key: id
    fields: id(), name(), survey_type()
  contacts:
    primary key: id
    fields: email(), id(), name()
  responses:
    primary key: id
    cursor: completed_time
    fields: completed_time(), id(), survey_id()
  questions:
    primary key: id
    fields: id(), question(), survey_id()
  channels:
    primary key: id
    fields: id(), name(), properties(), status(), type()
  contact_lists:
    primary key: id
    fields: description(), id(), name()
  contact_properties:
    primary key: id
    fields: contact_property_group_id(), description(), group(), id(), label(), name(), type()
  reminders:
    primary key: id
    fields: account_id(), after_days(), created_at(), frequency(), id(), message(), sent_count(), subject(), survey_id(), type(), updated_at()
  reputation_platforms:
    primary key: id
    fields: id(), label(), logo_url(), type()
  reputation_app_platforms:
    primary key: id
    fields: created_at(), data_fetch_address(), id(), is_active(), location(), platform_id(), updated_at()
  reputation_reviews:
    primary key: id
    fields: app_platform_id(), id(), rating(), review_content(), review_date(), review_title(), reviewer_name(), reviewer_photo_url()
  survey_folders:
    primary key: id
    fields: auto_created(), description(), id(), name(), parent_survey_folder_id(), teams(), users(), visibility()
  ticket_fields:
    primary key: id
    fields: created_at(), description(), id(), internal_name(), is_default(), mandatory(), name(), options(), type(), updated_at()
  tickets:
    primary key: id
    fields: agent(), created_at(), custom_fields(), deleted_at(), description(), description_html(), id(), priority(), requester(), source(), status(), subject(), team(), template_id(), updated_at()
  teams:
    primary key: id
    fields: account_id(), business_hour_id(), created_at(), deleted_at(), description(), id(), name(), round_robin_enabled(), type(), updated_at()
  roles:
    primary key: id
    fields: account_id(), created_at(), deleted_at(), description(), id(), label(), name(), updated_at()
  variables:
    primary key: id
    fields: description(), id(), label(), name(), type()
  webhooks:
    primary key: id
    fields: description(), eventType(), httpMethod(), id(), name(), objectType(), url()
  users:
    primary key: id
    fields: admin(), agency_owner(), email(), id(), name(), owner(), phone(), role_id(), verified()
  templates:
    primary key: id
    fields: created_at(), deleted_at(), description(), id(), name(), updated_at()
  email_themes:
    primary key: id
    fields: created_at(), id(), is_public(), name(), properties(), updated_at()
  expressions:
    primary key: id
    fields: id(), name(), representation()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  create_survey:
    endpoint: POST /surveys
    risk: external mutation; approval required
  update_survey:
    endpoint: PATCH /surveys/{{ record.id }}
    required fields: id
    risk: external mutation; approval required
  create_contact:
    endpoint: POST /contacts
    risk: external mutation; approval required
  update_contact:
    endpoint: PUT /contacts/{{ record.id }}
    required fields: id
    risk: external mutation; approval required
  delete_contact:
    endpoint: DELETE /contacts/{{ record.id }}
    required fields: id
    risk: irreversible external deletion; approval required
  create_question:
    endpoint: POST /questions
    risk: external mutation; approval required
  update_question:
    endpoint: PUT /questions/{{ record.question_id }}
    required fields: question_id
    risk: external mutation; approval required
  delete_question:
    endpoint: DELETE /questions/{{ record.question_id }}
    required fields: question_id
    risk: irreversible external deletion; approval required
  create_contact_list:
    endpoint: POST /contact_lists
    risk: external mutation; approval required
  update_contact_list:
    endpoint: PATCH /contact_lists/{{ record.id }}
    required fields: id
    risk: external mutation; approval required
  delete_contact_list:
    endpoint: DELETE /contact_lists/{{ record.id }}
    required fields: id
    risk: irreversible external deletion; approval required
  create_contact_property:
    endpoint: POST /contact_properties
    risk: external mutation; approval required
  update_contact_property:
    endpoint: PATCH /contact_properties/{{ record.id }}
    required fields: id
    risk: external mutation; approval required
  delete_contact_property:
    endpoint: DELETE /contact_properties/{{ record.id }}
    required fields: id
    risk: irreversible external deletion; approval required
  create_survey_folder:
    endpoint: POST /survey_folders
    risk: external mutation; approval required
  update_survey_folder:
    endpoint: PATCH /survey_folders/{{ record.id }}
    required fields: id
    risk: external mutation; approval required
  delete_survey_folder:
    endpoint: DELETE /survey_folders/{{ record.id }}
    required fields: id
    risk: irreversible external deletion; approval required
  create_team:
    endpoint: POST /teams
    risk: external mutation; approval required
  create_ticket:
    endpoint: POST /tickets
    risk: external mutation; approval required
  update_ticket:
    endpoint: PUT /tickets/{{ record.id }}
    required fields: id
    risk: external mutation; approval required
  delete_ticket:
    endpoint: DELETE /tickets/{{ record.id }}
    required fields: id
    risk: irreversible external deletion; approval required
  create_webhook:
    endpoint: POST /webhooks
    risk: external mutation; approval required
  update_webhook:
    endpoint: PUT /webhooks/{{ record.id }}
    required fields: id
    risk: external mutation; approval required
  delete_webhook:
    endpoint: DELETE /webhooks/{{ record.id }}
    required fields: id
    risk: irreversible external deletion; approval required
  create_user:
    endpoint: POST /users
    risk: external mutation creating a live user account with console access; approval required
  update_user:
    endpoint: PATCH /users/{{ record.id }}
    required fields: id
    risk: external mutation; approval required
  delete_user:
    endpoint: DELETE /users/{{ record.id }}
    required fields: id
    risk: irreversible external deletion of a user account; approval required
  create_reminder:
    endpoint: POST /reminders
    risk: external mutation; approval required
  delete_reminder:
    endpoint: DELETE /reminders/{{ record.id }}
    required fields: id
    risk: irreversible external deletion; approval required
  create_variable:
    endpoint: POST /variables
    risk: external mutation; approval required
  delete_variable:
    endpoint: DELETE /variables/{{ record.variable_id }}
    required fields: variable_id
    risk: irreversible external deletion; approval required
  create_channel:
    endpoint: POST /channels
    risk: external mutation; approval required
  delete_channel:
    endpoint: DELETE /channels/{{ record.id }}
    required fields: id
    risk: irreversible external deletion; approval required

SECURITY
  read risk: external SurveySparrow API read of survey, contact, response, question, and 18 additional catalog resource types
  write risk: external mutation of SurveySparrow surveys, contacts, questions, contact lists/properties, survey folders, teams, tickets, webhooks, users, reminders, variables, and channels, including irreversible deletes and live-user-account creation/deletion
  approval: read: none; write: required for all mutation actions
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect survey-sparrow

  # Inspect as structured JSON
  pm connectors inspect survey-sparrow --json

AGENT WORKFLOW
  - Run pm connectors inspect survey-sparrow before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
