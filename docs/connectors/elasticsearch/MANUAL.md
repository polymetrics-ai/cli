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
  id: elasticsearch
  asset: icons/elasticsearch.svg
  source: official
  review_status: official_verified
  review_url: https://www.elastic.co/docs/reference/elasticsearch

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  endpoint (required)
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
    fields: docs.count(string), index(string)
  documents:
    primary key: id
    fields: id(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

SECURITY
  read risk: external Elasticsearch cluster read of index metadata and documents
  approval: none; read-only cluster access
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Read Elasticsearch declared streams and bounded REST metadata.
  Usage: pm elasticsearch <command> [flags]
  Source CLI: Elasticsearch REST API (https://raw.githubusercontent.com/elastic/elasticsearch-specification/main/output/openapi/elasticsearch-openapi.json)
  Global flags:
    --credential (string): Named Elasticsearch credential; secrets are loaded from the credential store.
    --json (boolean): Emit machine-readable JSON output.
    --max-bytes (integer): Clamp direct-read response size; these operations are capped at 1 MiB.
  Elasticsearch cluster direct reads
  Elasticsearch lifecycle and license direct reads
  Other Commands
    cluster info get - Get Elasticsearch cluster information. [intent=direct_read availability=implemented operation=elasticsearch.cluster.info.get]; approval: none; risk: bounded read; requires the Elasticsearch monitor cluster privilege; the response is capped at 1 MiB and redacted before JSON output; flags: --page, --page-cursor
    cluster remote info get - Get Elasticsearch remote-cluster information. [intent=direct_read availability=implemented operation=elasticsearch.cluster.remote.info.get]; approval: none; risk: bounded read; requires the Elasticsearch monitor cluster privilege; the response is capped at 1 MiB and redacted before JSON output; flags: --page, --page-cursor
    ilm status get - Get Elasticsearch index lifecycle management status. [intent=direct_read availability=implemented operation=elasticsearch.ilm.status.get]; approval: none; risk: bounded read; requires the Elasticsearch read_ilm cluster privilege; the response is capped at 1 MiB and redacted before JSON output; flags: --page, --page-cursor
    license basic status get - Get Elasticsearch basic-license status. [intent=direct_read availability=implemented operation=elasticsearch.license.basic_status.get]; approval: none; risk: bounded read; requires the Elasticsearch monitor cluster privilege; the response is capped at 1 MiB and redacted before JSON output; flags: --page, --page-cursor
    license trial status get - Get Elasticsearch trial-license status. [intent=direct_read availability=implemented operation=elasticsearch.license.trial_status.get]; approval: none; risk: bounded read; requires the Elasticsearch monitor cluster privilege; the response is capped at 1 MiB and redacted before JSON output; flags: --page, --page-cursor

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
