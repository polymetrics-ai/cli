# pm connectors inspect retently

```text
NAME
  pm connectors inspect retently - Retently connector manual

SYNOPSIS
  pm connectors inspect retently
  pm connectors inspect retently --json
  pm credentials add <name> --connector retently [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Retently customers, survey responses, surveys, and campaigns through the REST API.

ICON
  asset: icons/retently.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://www.retently.com/api/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  campaign_id
  created_after
  email
  updated_after
  api_key (secret)

ETL STREAMS
  customers:
    primary key: id
    cursor: updated_at
    fields: company(), email(), full_name(), id(), stream(), updated_at()
  responses:
    primary key: id
    cursor: created_at
    fields: comment(), created_at(), customer_id(), id(), score(), stream()
  surveys:
    primary key: id
    cursor: updated_at
    fields: id(), name(), status(), stream(), type(), updated_at()
  campaigns:
    primary key: id
    cursor: updated_at
    fields: id(), name(), status(), stream(), updated_at()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Retently API read of customer and NPS/CSAT survey response data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect retently

  # Inspect as structured JSON
  pm connectors inspect retently --json

AGENT WORKFLOW
  - Run pm connectors inspect retently before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
