# pm connectors inspect customerly

```text
NAME
  pm connectors inspect customerly - Customerly connector manual

SYNOPSIS
  pm connectors inspect customerly
  pm connectors inspect customerly --json
  pm credentials add <name> --connector customerly [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Customerly users, leads, and accounts, and writes user/lead/tag/message/attribute/company mutations through the Customerly v1 REST API.

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
  page_size
  api_key (secret)

ETL STREAMS
  users:
    primary key: user_id, email
    cursor: last_update
    fields: city(), country(), create_date(), crmhero_user_id(), email(), first_seen_at(), last_activity(), last_update(), name(), role(), sub_active(), sub_status(), timezone(), user_id(), username()
  leads:
    primary key: crmhero_user_id
    cursor: last_update
    fields: city(), country(), create_date(), crmhero_user_id(), email(), last_update(), name(), role(), sub_active(), sub_status(), timezone(), username()
  accounts:
    primary key: account_id
    fields: account_id(), email()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  delete_user:
    endpoint: DELETE /users?user_id={{ record.user_id }}
    required fields: user_id
    risk: external mutation; irreversibly deletes a live Customerly user and every conversation/survey/campaign record tied to them; approval required
  delete_lead:
    endpoint: DELETE /leads?email={{ record.email }}
    required fields: email
    risk: external mutation; irreversibly deletes a live Customerly lead and every associated record; approval required
  unsubscribe_user:
    endpoint: POST /users/unsubscribe/{{ record.user_id }}
    required fields: user_id
    risk: external mutation; unsubscribes a live user from Customerly messaging; approval required
  add_tag:
    endpoint: POST /tags
    risk: external mutation; adds or removes a tag on one or more live users/leads
  delete_tag:
    endpoint: DELETE /tags
    optional fields: tag
    risk: external mutation; permanently removes a tag definition from the app; it is un-applied from every contact that carried it; approval required
  send_message:
    endpoint: POST /messages
    risk: sends a user-visible message from Customerly on the sender's behalf; may notify the recipient
  add_user_attributes:
    endpoint: POST /users/add-attributes/{{ record.user_id }}
    required fields: user_id
    risk: external mutation; adds/overwrites custom attribute values on a live user
  add_company_attributes:
    endpoint: POST /company/add-attributes/{{ record.company_id }}
    required fields: company_id
    risk: external mutation; adds/overwrites custom attribute values (and optionally renames) a live company
  add_user_to_company:
    endpoint: POST /users/add-to-company
    risk: external mutation; links a live user to a company, creating the company if it does not already exist

SECURITY
  read risk: external Customerly API read of user, lead, and account contact data
  write risk: external mutation of live Customerly users/leads/tags/messages/attributes/companies, including irreversible user and lead deletion; approval required
  approval: read: none; write: required
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect customerly

  # Inspect as structured JSON
  pm connectors inspect customerly --json

AGENT WORKFLOW
  - Run pm connectors inspect customerly before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
