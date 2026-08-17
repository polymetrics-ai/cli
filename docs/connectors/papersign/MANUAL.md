# pm connectors inspect papersign

```text
NAME
  pm connectors inspect papersign - PaperSign connector manual

SYNOPSIS
  pm connectors inspect papersign
  pm connectors inspect papersign --json
  pm credentials add <name> --connector papersign [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads PaperSign documents, templates, and recipients through the REST API.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  api_key (secret)

ETL STREAMS
  documents:
    primary key: id
    cursor: created_at
    fields: created_at(), id(), name(), status(), updated_at()
  templates:
    primary key: id
    cursor: created_at
    fields: created_at(), id(), name(), updated_at()
  recipients:
    primary key: id
    cursor: created_at
    fields: created_at(), document_id(), email(), id(), status()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external PaperSign API read of document, template, and recipient data
  approval: none; read-only, no obviously-safe reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect papersign

  # Inspect as structured JSON
  pm connectors inspect papersign --json

AGENT WORKFLOW
  - Run pm connectors inspect papersign before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
