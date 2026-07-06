# pm connectors inspect elasticsearch

```text
NAME
  pm connectors inspect elasticsearch - Elasticsearch connector manual

SYNOPSIS
  pm connectors inspect elasticsearch
  pm connectors inspect elasticsearch --json
  pm credentials add <name> --connector elasticsearch [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Elasticsearch index metadata and documents through the REST API. Read-only.

ICON
  asset: icons/elasticsearch.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://www.elastic.co/guide/en/elasticsearch/reference/current/rest-apis.html

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  endpoint
  index
  max_pages
  mode
  page_size
  username
  api_key_id (secret)
  api_key_secret (secret)
  password (secret)

ETL STREAMS
  indices:
    primary key: index
    fields: docs.count(), index()
  documents:
    primary key: id
    fields: id()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Elasticsearch cluster read of index metadata and documents
  approval: none; read-only cluster access
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect elasticsearch

  # Inspect as structured JSON
  pm connectors inspect elasticsearch --json

AGENT WORKFLOW
  - Run pm connectors inspect elasticsearch before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
