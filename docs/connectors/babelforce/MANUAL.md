# pm connectors inspect babelforce

```text
NAME
  pm connectors inspect babelforce - Babelforce connector manual

SYNOPSIS
  pm connectors inspect babelforce
  pm connectors inspect babelforce --json
  pm credentials add <name> --connector babelforce [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Babelforce call reporting, recordings, numbers, and users through fixed Babelforce v2 REST routes.

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
  region
  access_key_id (secret)
  access_token (secret)

ETL STREAMS
  calls:
    primary key: id
    cursor: dateCreated
    fields: dateCreated(string), id(string)
  calls_extended:
    primary key: id
    cursor: dateCreated
    fields: dateCreated(string), id(string)
  recordings:
    primary key: id
    cursor: dateCreated
    fields: dateCreated(string), id(string)
  numbers:
    primary key: id
    cursor: dateCreated
    fields: dateCreated(string), id(string)
  users:
    primary key: id
    cursor: dateCreated
    fields: dateCreated(string), id(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: Bounded Babelforce v2 reads use a source-declared regional provider origin and dual-header authentication.
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
