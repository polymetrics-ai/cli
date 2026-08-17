# pm connectors inspect nytimes

```text
NAME
  pm connectors inspect nytimes - New York Times connector manual

SYNOPSIS
  pm connectors inspect nytimes
  pm connectors inspect nytimes --json
  pm credentials add <name> --connector nytimes [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads New York Times Most Popular (viewed, emailed, shared) articles via the NYTimes Developer APIs.

ICON
  asset: icons/nytimes.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developer.nytimes.com/apis

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  period
  api_key (secret)

ETL STREAMS
  most_popular_viewed:
    primary key: id
    cursor: published_date
    fields: abstract(), byline(), id(), published_date(), section(), source(), title(), type(), updated(), uri(), url()
  most_popular_emailed:
    primary key: id
    cursor: published_date
    fields: abstract(), byline(), id(), published_date(), section(), source(), title(), type(), updated(), uri(), url()
  most_popular_shared:
    primary key: id
    cursor: published_date
    fields: abstract(), byline(), id(), published_date(), section(), source(), title(), type(), updated(), uri(), url()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external NYTimes API read of published article metadata (no PII)
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect nytimes

  # Inspect as structured JSON
  pm connectors inspect nytimes --json

AGENT WORKFLOW
  - Run pm connectors inspect nytimes before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
