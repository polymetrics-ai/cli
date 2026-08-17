# pm connectors inspect serpstat

```text
NAME
  pm connectors inspect serpstat - Serpstat connector manual

SYNOPSIS
  pm connectors inspect serpstat
  pm connectors inspect serpstat --json
  pm credentials add <name> --connector serpstat [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Serpstat SEO domain keyword, competitor, and top-URL data through the Serpstat JSON-RPC-over-HTTP API. Read-only.

ICON
  asset: icons/serpstat.svg
  source: upstream_registry
  review_status: upstream_seeded

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  domain
  page_size
  pages_to_fetch
  region_id
  api_key (secret)

ETL STREAMS
  domain_keywords:
    primary key: keyword, url
    fields: keyword(), position(), updated_at(), url()
  domain_competitors:
    primary key: domain
    fields: domain(), visibility()
  domain_urls:
    primary key: url
    fields: keywords(), url()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Serpstat API read of domain keyword/competitor/top-URL SEO metrics
  approval: none; read-only source connector
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect serpstat

  # Inspect as structured JSON
  pm connectors inspect serpstat --json

AGENT WORKFLOW
  - Run pm connectors inspect serpstat before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
