# pm connectors inspect box-data-extract

```text
NAME
  pm connectors inspect box-data-extract - Box Data Extract connector manual

SYNOPSIS
  pm connectors inspect box-data-extract
  pm connectors inspect box-data-extract --json
  pm credentials add <name> --connector box-data-extract [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Box folder files and per-file detail metadata, and writes file rename/description updates, through the Box REST API using the OAuth2 client-credentials grant.

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
  box_folder_id
  box_subject_id
  box_subject_type
  mode
  page_size
  token_url
  client_id (secret)
  client_secret (secret)

ETL STREAMS
  files:
    primary key: id
    fields: id(), name(), type()
  file_details:
    primary key: id
    cursor: modified_at
    fields: content_created_at(), content_modified_at(), created_at(), created_by(), description(), etag(), file_id(), id(), item_status(), modified_at(), modified_by(), name(), owned_by(), parent(), path_collection(), purged_at(), sha1(), shared_link(), size(), trashed_at(), type()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  update_file:
    endpoint: PUT /files/{{ record.id }}
    required fields: id
    risk: external mutation; renames or updates the description of a Box file; approval required

SECURITY
  read risk: external Box API read of folder files and per-file detail metadata
  write risk: external mutation renaming or updating the description of a Box file
  approval: required for the update_file write action; read remains unapproved
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect box-data-extract

  # Inspect as structured JSON
  pm connectors inspect box-data-extract --json

AGENT WORKFLOW
  - Run pm connectors inspect box-data-extract before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
