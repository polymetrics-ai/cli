# pm connectors inspect iterable

```text
NAME
  pm connectors inspect iterable - Iterable connector manual

SYNOPSIS
  pm connectors inspect iterable
  pm connectors inspect iterable --json
  pm credentials add <name> --connector iterable [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Iterable lists, campaigns, and templates through the Iterable REST API. Read-only.

ICON
  asset: icons/iterable.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://api.iterable.com/api/docs

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  page_size
  api_key (secret)

ETL STREAMS
  lists:
    primary key: id
    fields: createdAt(), id(), listType(), name(), updatedAt()
  campaigns:
    primary key: id
    fields: createdAt(), id(), name(), updatedAt()
  templates:
    primary key: id
    fields: createdAt(), id(), name(), updatedAt()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Iterable API read of lists, campaigns, and templates
  approval: none; read-only marketing-data API
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect iterable

  # Inspect as structured JSON
  pm connectors inspect iterable --json

AGENT WORKFLOW
  - Run pm connectors inspect iterable before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
