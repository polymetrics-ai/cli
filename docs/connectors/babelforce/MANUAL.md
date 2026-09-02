# pm connectors inspect babelforce

```text
NAME
  pm connectors inspect babelforce - Babelforce connector manual

SYNOPSIS
  pm connectors inspect babelforce
  pm connectors inspect babelforce --json
  pm credentials add <name> --connector babelforce [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Babelforce call reporting, recordings, numbers, and users through the Babelforce v2 REST API. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

ICON
  id: babelforce
  asset: icons/babelforce.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://api.babelforce.com/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  date_created_from
  date_created_to
  mode
  region (required)
  access_key_id (secret) (required)
  access_token (secret) (required)

ETL STREAMS
  calls:
    primary key: id
    cursor: dateCreated
    fields: anonymous(boolean), conversationId(string), dateCreated(string), dateEstablished(string), dateFinished(string), domain(string), duration(integer), finishReason(string), from(string), id(string), lastUpdated(string), parentId(string), sessionId(string), source(string), state(string), to(string), type(string)
  calls_extended:
    primary key: id
    cursor: dateCreated
    fields: anonymous(boolean), conversationId(string), dateCreated(string), dateEstablished(string), dateFinished(string), domain(string), duration(integer), finishReason(string), from(string), id(string), lastUpdated(string), parentId(string), sessionId(string), source(string), state(string), to(string), type(string)
  recordings:
    primary key: id
    cursor: dateCreated
    fields: dateCreated(string), duration(integer), id(string), lastUpdated(string), state(string), url(string)
  numbers:
    primary key: id
    cursor: dateCreated
    fields: dateCreated(string), id(string), lastUpdated(string), name(string), number(string), state(string)
  users:
    primary key: id
    cursor: dateCreated
    fields: dateCreated(string), id(string), lastUpdated(string), name(string), number(string), state(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Babelforce API reads performed by the legacy connector via a Tier-2 hook
  write risk: unsupported
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect babelforce

  # Inspect as structured JSON
  pm connectors inspect babelforce --json

AGENT WORKFLOW
  - Run pm connectors inspect babelforce before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
