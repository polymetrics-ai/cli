# pm connectors inspect boldsign

```text
NAME
  pm connectors inspect boldsign - BoldSign connector manual

SYNOPSIS
  pm connectors inspect boldsign
  pm connectors inspect boldsign --json
  pm credentials add <name> --connector boldsign [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads BoldSign documents, templates, teams, contacts, brands, users, contact groups, and sender identities, and writes team/contact-group/document-lifecycle/user-lifecycle mutations, through the BoldSign REST API.

ICON
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  api_key (secret) (required)

ETL STREAMS
  documents:
    primary key: document_id
    cursor: created_date
    fields: created_date(integer), document_id(string), enable_signing_order(boolean), expiry_date(integer), is_deleted(boolean), labels(array), message_title(string), sender_detail(object), sender_email(string), signer_details(array), status(string)
  templates:
    primary key: document_id
    cursor: created_date
    fields: created_date(integer), document_id(string), is_shared_template(boolean), labels(array), sender_email(string), template_description(string), template_name(string)
  teams:
    primary key: team_id
    cursor: created_date
    fields: created_date(integer), team_id(string), team_name(string), users(array)
  contacts:
    primary key: id
    fields: company_name(string), email(string), id(string), name(string), phone_number(object)
  brands:
    primary key: brand_id
    fields: background_color(string), brand_id(string), brand_name(string), button_color(string), is_default(boolean)
  users:
    primary key: user_id
    cursor: created_date
    fields: created_date(integer), email(string), first_name(string), last_name(string), meta_data(object), modified_date(integer), role(string), team_id(string), team_name(string), user_id(string), user_status(string)
  contact_groups:
    primary key: group_id
    fields: contacts(array), directories(array), group_id(string), group_name(string)
  sender_identities:
    primary key: id
    fields: approved_date(string), brand_id(string), created_by(string), email(string), id(string), meta_data(object), name(string), notification_settings(object), redirect_url(string), status(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_team:
    endpoint: POST /v1/teams/create
    required fields: teamName
    risk: external mutation; creates a new BoldSign team; approval required
  update_team:
    endpoint: PUT /v1/teams/update
    required fields: teamId, teamName
    risk: external mutation; renames an existing BoldSign team; approval required
  update_contact:
    endpoint: PUT /v1/contacts/update?id={{ record.id }}
    required fields: id, email, name
    risk: external mutation; overwrites an existing BoldSign contact's details; approval required
  delete_contact:
    endpoint: DELETE /v1/contacts/delete?id={{ record.id }}
    required fields: id
    risk: destructive external mutation; permanently deletes a BoldSign contact; approval required
  create_contact_group:
    endpoint: POST /v1/contactGroups/create
    required fields: groupName
    risk: external mutation; creates a new BoldSign contact group; approval required
  update_contact_group:
    endpoint: PUT /v1/contactGroups/update?groupId={{ record.groupId }}
    required fields: groupId, groupName
    risk: external mutation; overwrites an existing BoldSign contact group's members/name; approval required
  delete_contact_group:
    endpoint: DELETE /v1/contactGroups/delete?groupId={{ record.groupId }}
    required fields: groupId
    risk: destructive external mutation; permanently deletes a BoldSign contact group; approval required
  revoke_document:
    endpoint: POST /v1/document/revoke?documentId={{ record.documentId }}
    required fields: documentId, message
    risk: destructive external mutation; revokes a BoldSign document, permanently ending its signature request; approval required
  remind_document:
    endpoint: POST /v1/document/remind?documentId={{ record.documentId }}
    required fields: documentId
    risk: external mutation; sends an email/SMS reminder to a document's pending signers; approval required
  delete_document:
    endpoint: DELETE /v1/document/delete?documentId={{ record.documentId }}&deletePermanently={{ record.deletePermanently }}
    required fields: documentId, deletePermanently
    risk: destructive external mutation; moves a BoldSign document to trash (or permanently deletes it when deletePermanently=true); approval required
  add_document_tags:
    endpoint: PATCH /v1/document/addTags
    required fields: documentId, tags
    risk: external mutation; adds label tags to a BoldSign document; approval required
  delete_document_tags:
    endpoint: DELETE /v1/document/deleteTags
    required fields: documentId, tags
    risk: external mutation; removes label tags from a BoldSign document; approval required
  update_user:
    endpoint: PUT /v1/users/update
    required fields: userId
    risk: external mutation; changes a BoldSign user's role or active/deactivated status; approval required
  change_user_team:
    endpoint: PUT /v1/users/changeTeam?userId={{ record.userId }}
    required fields: userId, toTeamId
    risk: external mutation; moves a BoldSign user to a different team; approval required

SECURITY
  read risk: external BoldSign API read of documents, templates, teams, contacts, brands, users, contact groups, and sender identities
  write risk: external mutation of BoldSign teams, contacts, contact groups, document lifecycle state (revoke/remind/delete/tags), and user role/team/status; includes 2 destructive (irreversible-effect) actions (delete_contact, delete_contact_group, delete_document, revoke_document)
  approval: required for every write action; read remains unapproved
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect boldsign

  # Inspect as structured JSON
  pm connectors inspect boldsign --json

AGENT WORKFLOW
  - Run pm connectors inspect boldsign before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
