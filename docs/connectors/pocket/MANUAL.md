# pm connectors inspect pocket

```text
NAME
  pm connectors inspect pocket - Pocket connector manual

SYNOPSIS
  pm connectors inspect pocket
  pm connectors inspect pocket --json
  pm credentials add <name> --connector pocket [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads saved Pocket items through the v3 retrieve API. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

ICON
  asset: icons/pocket.svg
  source: upstream_registry
  review_status: upstream_seeded

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  content_type
  detail_type
  domain
  favorite
  mode
  search
  since
  sort
  state
  tag
  access_token (secret)
  consumer_key (secret)

ETL STREAMS
  items:
    primary key: item_id
    cursor: updated_at
    fields: excerpt(), item_id(), title(), updated_at(), url()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Pocket API reads performed by the legacy connector via a Tier-2 hook
  write risk: unsupported
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect pocket

  # Inspect as structured JSON
  pm connectors inspect pocket --json

AGENT WORKFLOW
  - Run pm connectors inspect pocket before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
