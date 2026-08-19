# pm connectors inspect airtable

```text
NAME
  pm connectors inspect airtable - Airtable connector manual

SYNOPSIS
  pm connectors inspect airtable
  pm connectors inspect airtable --json
  pm credentials add <name> --connector airtable [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Airtable bases, tables, records, webhooks, and record comments, and writes record/table/field/comment/webhook mutations, through the Airtable Web API.

ICON
  id: airtable
  asset: icons/airtable.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://airtable.com/developers/web/api/changelog

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_id
  base_url
  page_size
  table_id
  access_token (secret)
  api_key (secret)

ETL STREAMS
  bases:
    primary key: id
    fields: id(string), name(string), permissionLevel(string)
  tables:
    primary key: id
    fields: description(string), fields(array), id(string), name(string), primaryFieldId(string), views(array)
  records:
    primary key: id
    fields: createdTime(string), fields(object), id(string)
  webhooks:
    primary key: id
    fields: areNotificationsEnabled(boolean), cursorForNextPayload(integer), expirationTime(string), id(string), isHookEnabled(boolean), lastSuccessfulNotificationTime(string), notificationUrl(string), specification(object)
  comments:
    primary key: id
    fields: author(object), createdTime(string), id(string), lastUpdatedTime(string), parentCommentId(string), record_id(string), text(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

REVERSE ETL ACTIONS
  create_record:
    endpoint: POST /{{ config.base_id }}/{{ config.table_id }}
    required fields: fields
    risk: creates a new record in the configured base/table; low-risk external mutation, no approval required
  update_record:
    endpoint: PATCH /{{ config.base_id }}/{{ config.table_id }}/{{ record.id }}
    required fields: id, fields
    risk: mutates only the field values included in the request (non-destructive PATCH); unincluded cell values are left unchanged
  delete_record:
    endpoint: DELETE /{{ config.base_id }}/{{ config.table_id }}/{{ record.id }}
    required fields: id
    risk: permanently removes a record from the base/table; irreversible
  create_table:
    endpoint: POST /meta/bases/{{ config.base_id }}/tables
    required fields: name, fields
    risk: creates a new table (schema mutation) in the configured base; low-risk but changes the base's structure, visible to every collaborator
  update_table:
    endpoint: PATCH /meta/bases/{{ config.base_id }}/tables/{{ record.id }}
    required fields: id
    risk: renames or redescribes an existing table; a visible schema change for every collaborator on the base
  create_field:
    endpoint: POST /meta/bases/{{ config.base_id }}/tables/{{ record.table_id }}/fields
    required fields: table_id, name, type
    risk: creates a new column (schema mutation) in the target table; low-risk but changes the table's structure, visible to every collaborator
  update_field:
    endpoint: PATCH /meta/bases/{{ config.base_id }}/tables/{{ record.table_id }}/fields/{{ record.id }}
    required fields: table_id, id
    risk: renames or redescribes an existing column; a visible schema change for every collaborator on the base
  create_comment:
    endpoint: POST /{{ config.base_id }}/{{ config.table_id }}/{{ record.record_id }}/comments
    required fields: record_id, text
    risk: adds a visible comment to a record; every base collaborator with record access can see it, no external side effect
  update_comment:
    endpoint: PATCH /{{ config.base_id }}/{{ config.table_id }}/{{ record.record_id }}/comments/{{ record.id }}
    required fields: record_id, id, text
    risk: edits the text of an existing comment; visible to every base collaborator with record access
  delete_comment:
    endpoint: DELETE /{{ config.base_id }}/{{ config.table_id }}/{{ record.record_id }}/comments/{{ record.id }}
    required fields: record_id, id
    risk: permanently removes a comment from a record; irreversible
  create_webhook:
    endpoint: POST /bases/{{ config.base_id }}/webhooks
    required fields: notificationUrl, specification
    risk: registers a new outbound webhook that will POST live base-change notifications to an external URL of the caller's choosing; verify the target endpoint before enabling
  delete_webhook:
    endpoint: DELETE /bases/{{ config.base_id }}/webhooks/{{ record.id }}
    required fields: id
    risk: permanently removes a webhook subscription; notification delivery to its target URL stops immediately

SECURITY
  read risk: external Airtable API read of base/table/record metadata
  write risk: external Airtable API mutation of records, table/field schema, comments, and webhooks; schema mutations (create_table/update_table/create_field/update_field) are visible to every base collaborator, approval required
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect airtable

  # Inspect as structured JSON
  pm connectors inspect airtable --json

AGENT WORKFLOW
  - Run pm connectors inspect airtable before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
