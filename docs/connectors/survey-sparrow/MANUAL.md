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
  id: surveysparrow
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
  access_token (secret) (required)

ETL STREAMS
  surveys:
    primary key: id
    fields: id(integer), name(string), survey_type(string)
  contacts:
    primary key: id
    fields: email(string), id(integer), name(string)
  responses:
    primary key: id
    cursor: completed_time
    fields: completed_time(string), id(integer), survey_id(integer)
  questions:
    primary key: id
    fields: id(integer), question(string), survey_id(integer)
  channels:
    primary key: id
    fields: id(integer), name(string), properties(object), status(string), type(string)
  contact_lists:
    primary key: id
    fields: description(string), id(integer), name(string)
  contact_properties:
    primary key: id
    fields: contact_property_group_id(integer), description(string), group(string), id(integer), label(string), name(string), type(string)
  reminders:
    primary key: id
    fields: account_id(integer), after_days(integer), created_at(string), frequency(string), id(integer), message(string), sent_count(integer), subject(string), survey_id(integer), type(string), updated_at(string)
  reputation_platforms:
    primary key: id
    fields: id(integer), label(string), logo_url(string), type(string)
  reputation_app_platforms:
    primary key: id
    fields: created_at(string), data_fetch_address(string), id(integer), is_active(boolean), location(string), platform_id(integer), updated_at(string)
  reputation_reviews:
    primary key: id
    fields: app_platform_id(integer), id(integer), rating(number), review_content(string), review_date(string), review_title(string), reviewer_name(string), reviewer_photo_url(string)
  survey_folders:
    primary key: id
    fields: auto_created(boolean), description(string), id(integer), name(string), parent_survey_folder_id(integer), teams(array), users(array), visibility(string)
  ticket_fields:
    primary key: id
    fields: created_at(string), description(string), id(integer), internal_name(string), is_default(boolean), mandatory(boolean), name(string), options(array), type(string), updated_at(string)
  tickets:
    primary key: id
    fields: agent(object), created_at(string), custom_fields(object), deleted_at(string), description(string), description_html(string), id(integer), priority(object), requester(object), source(object), status(object), subject(string), team(object), template_id(integer), updated_at(string)
  teams:
    primary key: id
    fields: account_id(integer), business_hour_id(integer), created_at(string), deleted_at(string), description(string), id(integer), name(string), round_robin_enabled(boolean), type(string), updated_at(string)
  roles:
    primary key: id
    fields: account_id(integer), created_at(string), deleted_at(string), description(string), id(integer), label(string), name(string), updated_at(string)
  variables:
    primary key: id
    fields: description(string), id(integer), label(string), name(string), type(string)
  webhooks:
    primary key: id
    fields: description(string), eventType(string), httpMethod(string), id(integer), name(string), objectType(string), url(string)
  users:
    primary key: id
    fields: admin(boolean), agency_owner(boolean), email(string), id(integer), name(string), owner(boolean), phone(string), role_id(integer), verified(boolean)
  templates:
    primary key: id
    fields: created_at(string), deleted_at(string), description(string), id(integer), name(string), updated_at(string)
  email_themes:
    primary key: id
    fields: created_at(string), id(integer), is_public(boolean), name(string), properties(object), updated_at(string)
  expressions:
    primary key: id
    fields: id(integer), name(string), representation(array)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  create_survey:
    endpoint: POST /surveys
    required fields: name, survey_type
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
    required fields: survey_id, text, type
    risk: external mutation; approval required
  update_question:
    endpoint: PUT /questions/{{ record.question_id }}
    required fields: question_id, survey_id
    risk: external mutation; approval required
  delete_question:
    endpoint: DELETE /questions/{{ record.question_id }}
    required fields: question_id
    risk: irreversible external deletion; approval required
  create_contact_list:
    endpoint: POST /contact_lists
    required fields: name
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
    required fields: type, label
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
    required fields: name
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
    required fields: name
    risk: external mutation; approval required
  create_ticket:
    endpoint: POST /tickets
    required fields: subject, priority, status
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
    required fields: url, survey_id, http_method
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
    required fields: name, email, role_id
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
    required fields: channel_id, survey_id, frequency, type, interval, embed_first_question, custom_footer
    risk: external mutation; approval required
  delete_reminder:
    endpoint: DELETE /reminders/{{ record.id }}
    required fields: id
    risk: irreversible external deletion; approval required
  create_variable:
    endpoint: POST /variables
    required fields: survey_id, label, name, type
    risk: external mutation; approval required
  delete_variable:
    endpoint: DELETE /variables/{{ record.variable_id }}
    required fields: variable_id
    risk: irreversible external deletion; approval required
  create_channel:
    endpoint: POST /channels
    required fields: type
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
