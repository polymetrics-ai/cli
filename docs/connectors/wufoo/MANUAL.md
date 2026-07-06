# pm connectors inspect wufoo

```text
NAME
  pm connectors inspect wufoo - Wufoo connector manual

SYNOPSIS
  pm connectors inspect wufoo
  pm connectors inspect wufoo --json
  pm credentials add <name> --connector wufoo [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Wufoo forms, fields, entries, comments, reports, and widgets, and writes entry submissions and webhook registrations through the Wufoo API.

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
  form_hash
  max_pages
  mode
  page_size
  report_hash
  api_key (secret)

ETL STREAMS
  forms:
    primary key: Hash
    cursor: DateUpdated
    fields: DateUpdated(), Hash(), Name()
  form_fields:
    primary key: ID
    fields: ClassNames(), ID(), Instructions(), IsRequired(), Title(), Type()
  entries:
    primary key: Hash
    cursor: DateUpdated
    fields: DateCreated(), DateUpdated(), EntryId(), Hash()
  form_comments:
    primary key: CommentId
    cursor: DateCreated
    fields: CommentId(), CommentedBy(), DateCreated(), EntryId(), Text()
  reports:
    primary key: Hash
    cursor: DateUpdated
    fields: DateUpdated(), Hash(), Name()
  report_fields:
    primary key: ID
    fields: ClassNames(), ID(), Instructions(), Title(), Type()
  report_entries:
    primary key: EntryId
    cursor: DateUpdated
    fields: DateCreated(), DateUpdated(), EntryId()
  report_widgets:
    primary key: Hash
    fields: Hash(), Name(), Size(), Type(), TypeDesc()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  submit_entry:
    endpoint: POST /forms/{{ config.form_hash }}/entries.json
    risk: external mutation; creates a live Wufoo form entry; approval required
  add_webhook:
    endpoint: PUT /forms/{{ config.form_hash }}/webhooks.json
    risk: external mutation; registers a webhook callback URL on the configured form; approval required
  delete_webhook:
    endpoint: DELETE /forms/{{ config.form_hash }}/webhooks/{{ record.hash }}.json
    required fields: hash
    risk: irreversible external deletion; removes a registered webhook from the configured form; approval required

SECURITY
  read risk: external Wufoo API read of form, field, entry, comment, report, and widget data
  write risk: external mutation: submits live form entries and registers/removes webhook callback URLs
  approval: required for all write actions (submit_entry, add_webhook, delete_webhook); reads require none
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect wufoo

  # Inspect as structured JSON
  pm connectors inspect wufoo --json

AGENT WORKFLOW
  - Run pm connectors inspect wufoo before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
