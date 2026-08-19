# pm connectors inspect the-guardian-api

```text
NAME
  pm connectors inspect the-guardian-api - The Guardian API connector manual

SYNOPSIS
  pm connectors inspect the-guardian-api
  pm connectors inspect the-guardian-api --json
  pm credentials add <name> --connector the-guardian-api [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Guardian content search results through the Guardian Open Platform Content API.

ICON
  id: theguardian
  asset: icons/theguardian.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://open-platform.theguardian.com/documentation/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  content_id
  query
  api_key (secret) (required)

ETL STREAMS
  search:
    primary key: id
    cursor: published_at
    fields: id(string), published_at(string), title(string)
  tags:
    primary key: id
    fields: apiUrl(string), id(string), sectionId(string), sectionName(string), type(string), webTitle(string), webUrl(string)
  sections:
    primary key: id
    fields: apiUrl(string), editions(array), id(string), webTitle(string), webUrl(string)
  editions:
    primary key: id
    fields: apiUrl(string), edition(string), id(string), path(string), webTitle(string), webUrl(string)
  content:
    primary key: id
    fields: apiUrl(string), id(string), isHosted(boolean), pillarId(string), pillarName(string), published_at(string), sectionId(string), sectionName(string), title(string), type(string), webUrl(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Guardian Open Platform API read of published content search results
  approval: none; read-only, no obviously-safe reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect the-guardian-api

  # Inspect as structured JSON
  pm connectors inspect the-guardian-api --json

AGENT WORKFLOW
  - Run pm connectors inspect the-guardian-api before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
