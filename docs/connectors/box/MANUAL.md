# pm connectors inspect box

```text
NAME
  pm connectors inspect box - Box connector manual

SYNOPSIS
  pm connectors inspect box
  pm connectors inspect box --json
  pm credentials add <name> --connector box [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Box users, groups, collections, folder items, webhooks, retention policies, legal hold policies, storage policies, sign requests, terms of services, metadata templates, and pending collaborations, and writes group/webhook/collaboration lifecycle mutations, through the Box REST API using the OAuth2 client-credentials grant.

ICON
  id: simple-icons-box
  asset: icons/simple-icons/box.svg
  title: Box
  simple_icon_slug: box
  simple_icon_hex: 0061D5
  source: simple-icons
  license: CC0-1.0
  review_status: cc0_with_trademark_caveat
  review_url: https://simpleicons.org/?q=Box
  match: exact-name-or-slug
  matched_by: box

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  box_subject_id
  box_subject_type
  folder_id
  mode
  token_url
  client_id (secret) (required)
  client_secret (secret) (required)

ETL STREAMS
  users:
    primary key: id
    cursor: modified_at
    fields: created_at(string), id(string), language(string), login(string), modified_at(string), name(string), status(string), timezone(string), type(string)
  groups:
    primary key: id
    cursor: modified_at
    fields: created_at(string), group_type(string), id(string), modified_at(string), name(string), type(string)
  collections:
    primary key: id
    fields: collection_type(string), id(string), name(string), type(string)
  folder_items:
    primary key: id
    cursor: modified_at
    fields: created_at(string), etag(string), id(string), modified_at(string), name(string), sequence_id(string), sha1(string), size(integer), type(string)
  webhooks:
    primary key: id
    fields: address(string), created_at(string), created_by(object), id(string), target(object), triggers(array), type(string)
  retention_policies:
    primary key: id
    cursor: modified_at
    fields: are_owners_notified(boolean), can_owner_extend_retention(boolean), created_at(string), created_by(object), custom_notification_recipients(array), description(string), disposition_action(string), id(string), modified_at(string), policy_name(string), policy_type(string), retention_length(string), retention_type(string), status(string), type(string)
  legal_hold_policies:
    primary key: id
    cursor: modified_at
    fields: assignment_counts(object), created_at(string), created_by(object), description(string), filter_ended_at(string), filter_started_at(string), id(string), modified_at(string), policy_name(string), status(string), type(string)
  storage_policies:
    primary key: id
    fields: id(string), name(string), type(string)
  sign_requests:
    primary key: id
    cursor: created_at
    fields: auto_expire_at(string), created_at(string), finished_at(string), id(string), parent_folder(object), prepare_url(string), sender_email(string), sign_files(object), signers(array), signing_log(object), source_files(array), status(string), type(string)
  terms_of_services:
    primary key: id
    cursor: modified_at
    fields: created_at(string), id(string), modified_at(string), status(string), text(string), tos_type(string), type(string)
  metadata_templates:
    primary key: id
    fields: copy_instance_on_item_copy(boolean), display_name(string), fields(array), hidden(boolean), id(string), scope(string), template_key(string), type(string)
  pending_collaborations:
    primary key: id
    cursor: modified_at
    fields: accessible_by(object), acknowledged_at(string), created_at(string), created_by(object), expires_at(string), id(string), invite_email(string), is_access_only(boolean), item(object), modified_at(string), role(string), status(string), type(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_group:
    endpoint: POST /groups
    required fields: name
    risk: external mutation; creates a new Box enterprise group; approval required
  update_group:
    endpoint: PUT /groups/{{ record.id }}
    required fields: id
    risk: external mutation; updates an existing Box enterprise group's settings; approval required
  delete_group:
    endpoint: DELETE /groups/{{ record.id }}
    required fields: id
    risk: destructive external mutation; permanently deletes a Box enterprise group; approval required
  create_webhook:
    endpoint: POST /webhooks
    required fields: target, address, triggers
    risk: external mutation; creates a new Box webhook subscription that will POST event payloads to an external address; approval required
  update_webhook:
    endpoint: PUT /webhooks/{{ record.id }}
    required fields: id
    risk: external mutation; updates an existing Box webhook's target/address/triggers; approval required
  delete_webhook:
    endpoint: DELETE /webhooks/{{ record.id }}
    required fields: id
    risk: destructive external mutation; permanently deletes a Box webhook subscription; approval required
  create_collaboration:
    endpoint: POST /collaborations
    required fields: item, accessible_by, role
    risk: external mutation; grants a user or group access to a Box file/folder; approval required
  update_collaboration:
    endpoint: PUT /collaborations/{{ record.id }}
    required fields: id
    risk: external mutation; changes an existing Box collaboration's role, or accepts/rejects a pending invitation; approval required
  delete_collaboration:
    endpoint: DELETE /collaborations/{{ record.id }}
    required fields: id
    risk: destructive external mutation; permanently revokes a user or group's access to a Box file/folder; approval required

SECURITY
  read risk: external Box API read of enterprise users, groups, collections, folder items, webhooks, retention policies, legal hold policies, storage policies, sign requests, terms of services, metadata templates, and pending collaborations
  write risk: external mutation of Box enterprise groups, webhook subscriptions, and file/folder collaborations (access grants); includes 3 destructive (irreversible-effect) actions (delete_group, delete_webhook, delete_collaboration)
  approval: required for every write action; read remains unapproved
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect box

  # Inspect as structured JSON
  pm connectors inspect box --json

AGENT WORKFLOW
  - Run pm connectors inspect box before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
