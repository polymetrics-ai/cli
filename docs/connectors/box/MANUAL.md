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
  box_subject_id
  box_subject_type
  folder_id
  mode
  token_url
  client_id (secret)
  client_secret (secret)

ETL STREAMS
  users:
    primary key: id
    cursor: modified_at
    fields: created_at(), id(), language(), login(), modified_at(), name(), status(), timezone(), type()
  groups:
    primary key: id
    cursor: modified_at
    fields: created_at(), group_type(), id(), modified_at(), name(), type()
  collections:
    primary key: id
    fields: collection_type(), id(), name(), type()
  folder_items:
    primary key: id
    cursor: modified_at
    fields: created_at(), etag(), id(), modified_at(), name(), sequence_id(), sha1(), size(), type()
  webhooks:
    primary key: id
    fields: address(), created_at(), created_by(), id(), target(), triggers(), type()
  retention_policies:
    primary key: id
    cursor: modified_at
    fields: are_owners_notified(), can_owner_extend_retention(), created_at(), created_by(), custom_notification_recipients(), description(), disposition_action(), id(), modified_at(), policy_name(), policy_type(), retention_length(), retention_type(), status(), type()
  legal_hold_policies:
    primary key: id
    cursor: modified_at
    fields: assignment_counts(), created_at(), created_by(), description(), filter_ended_at(), filter_started_at(), id(), modified_at(), policy_name(), status(), type()
  storage_policies:
    primary key: id
    fields: id(), name(), type()
  sign_requests:
    primary key: id
    cursor: created_at
    fields: auto_expire_at(), created_at(), finished_at(), id(), parent_folder(), prepare_url(), sender_email(), sign_files(), signers(), signing_log(), source_files(), status(), type()
  terms_of_services:
    primary key: id
    cursor: modified_at
    fields: created_at(), id(), modified_at(), status(), text(), tos_type(), type()
  metadata_templates:
    primary key: id
    fields: copy_instance_on_item_copy(), display_name(), fields(), hidden(), id(), scope(), template_key(), type()
  pending_collaborations:
    primary key: id
    cursor: modified_at
    fields: accessible_by(), acknowledged_at(), created_at(), created_by(), expires_at(), id(), invite_email(), is_access_only(), item(), modified_at(), role(), status(), type()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_group:
    endpoint: POST /groups
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
