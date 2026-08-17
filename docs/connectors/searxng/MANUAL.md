# pm connectors inspect searxng

```text
NAME
  pm connectors inspect searxng - SearXNG connector manual

SYNOPSIS
  pm connectors inspect searxng
  pm connectors inspect searxng --json
  pm credentials add <name> --connector searxng [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads web and Reddit search results from a SearXNG metasearch instance's JSON API (format=json). Read-only. Requires base_url; no credentials by default.

ICON
  asset: icons/searxng.svg
  source: official_site
  review_status: manual_override
  review_url: https://docs.searxng.org/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  query
  api_key (secret)

ETL STREAMS
  search:
    primary key: url
    cursor: published_date
    fields: category(), content(), engine(), engines(), published_date(), score(), stream(), thumbnail(), title(), url()
  reddit:
    primary key: url
    cursor: published_date
    fields: category(), content(), engine(), engines(), published_date(), score(), stream(), thumbnail(), title(), url()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external SearXNG instance read of web/Reddit search results
  approval: none; read-only public search API
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect searxng

  # Inspect as structured JSON
  pm connectors inspect searxng --json

AGENT WORKFLOW
  - Run pm connectors inspect searxng before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
