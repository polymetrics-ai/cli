# pm connectors inspect ticketmaster

```text
NAME
  pm connectors inspect ticketmaster - Ticketmaster connector manual

SYNOPSIS
  pm connectors inspect ticketmaster
  pm connectors inspect ticketmaster --json
  pm credentials add <name> --connector ticketmaster [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads events, venues, attractions, and classifications from the Ticketmaster Discovery API.

ICON
  id: ticketmaster
  asset: icons/ticketmaster.svg
  source: official
  review_status: official_verified
  review_url: https://developer.ticketmaster.com/products-and-docs/apis/discovery-api/v2/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  country_code
  keyword
  locale
  api_key (secret) (required)

ETL STREAMS
  events:
    primary key: id
    fields: id(string), locale(string), name(string), type(string), url(string)
  venues:
    primary key: id
    fields: city(object), country(object), id(string), name(string), url(string)
  attractions:
    primary key: id
    fields: id(string), locale(string), name(string), type(string), url(string)
  classifications:
    primary key: id
    fields: genre(object), id(string), name(string), segment(object), subGenre(object)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

SECURITY
  read risk: external Ticketmaster Discovery API read of public event/venue data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect ticketmaster

  # Inspect as structured JSON
  pm connectors inspect ticketmaster --json

AGENT WORKFLOW
  - Run pm connectors inspect ticketmaster before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
