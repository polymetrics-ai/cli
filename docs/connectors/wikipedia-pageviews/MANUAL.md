# pm connectors inspect wikipedia-pageviews

```text
NAME
  pm connectors inspect wikipedia-pageviews - Wikipedia Pageviews connector manual

SYNOPSIS
  pm connectors inspect wikipedia-pageviews
  pm connectors inspect wikipedia-pageviews --json
  pm credentials add <name> --connector wikipedia-pageviews [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Wikimedia pageview metrics for articles and top-article reports through the public Wikimedia REST API.

ICON
  id: wikipedia-pageviews
  asset: icons/wikipedia-pageviews.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://wikitech.wikimedia.org/wiki/Analytics/AQS/Pageviews

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  No secret authentication is required for this connector.

CONFIGURATION
  access (required)
  agent (required)
  article (required)
  base_url
  country
  day
  end (required)
  month
  project (required)
  start (required)
  year

ETL STREAMS
  pageviews:
    primary key: id
    cursor: timestamp
    fields: access(string), agent(string), article(string), granularity(string), id(string), project(string), timestamp(string), views(integer)
  top_articles:
    primary key: id
    fields: access(string), articles(array), country(string), day(string), id(string), month(string), project(string), year(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Wikimedia public API read of aggregate pageview metrics; no authentication, no PII
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect wikipedia-pageviews

  # Inspect as structured JSON
  pm connectors inspect wikipedia-pageviews --json

AGENT WORKFLOW
  - Run pm connectors inspect wikipedia-pageviews before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
