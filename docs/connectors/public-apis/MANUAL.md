# pm connectors inspect public-apis

```text
NAME
  pm connectors inspect public-apis - Public APIs connector manual

SYNOPSIS
  pm connectors inspect public-apis
  pm connectors inspect public-apis --json
  pm credentials add <name> --connector public-apis [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads public API directory entries and categories from the api.publicapis.org directory API. Read-only and credential-free.

ICON
  asset: icons/public-apis.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://github.com/public-apis/public-apis

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  No secret authentication is required for this connector.

CONFIGURATION
  base_url
  mode

ETL STREAMS
  entries:
    primary key: id
    fields: api(), auth(), category(), cors(), description(), https(), id(), link()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external public-apis.org directory read of API listing metadata
  approval: none; read-only, credential-free public directory API
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect public-apis

  # Inspect as structured JSON
  pm connectors inspect public-apis --json

AGENT WORKFLOW
  - Run pm connectors inspect public-apis before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
