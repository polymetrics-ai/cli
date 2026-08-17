# pm connectors inspect mendeley

```text
NAME
  pm connectors inspect mendeley - Mendeley connector manual

SYNOPSIS
  pm connectors inspect mendeley
  pm connectors inspect mendeley --json
  pm credentials add <name> --connector mendeley [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads documents, folders, groups, and annotations from the Mendeley reference manager REST API. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  name_for_institution
  query_for_catalog
  start_date
  client_id (secret)
  client_refresh_token (secret)
  client_secret (secret)

ETL STREAMS
  documents:
    primary key: id
    cursor: last_modified
    fields: abstract(), created(), group_id(), id(), last_modified(), profile_id(), source(), title(), type(), year()
  folders:
    primary key: id
    cursor: modified
    fields: created(), group_id(), id(), modified(), name(), parent_id()
  groups:
    primary key: id
    fields: access_level(), created(), description(), id(), name(), owning_profile_id(), role(), webpage()
  annotations:
    primary key: id
    cursor: last_modified
    fields: created(), document_id(), filehash(), id(), last_modified(), privacy_level(), profile_id(), text(), type()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Mendeley API reads performed by the legacy connector via a Tier-2 hook
  write risk: unsupported
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect mendeley

  # Inspect as structured JSON
  pm connectors inspect mendeley --json

AGENT WORKFLOW
  - Run pm connectors inspect mendeley before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
