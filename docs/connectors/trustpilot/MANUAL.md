# pm connectors inspect trustpilot

```text
NAME
  pm connectors inspect trustpilot - Trustpilot connector manual

SYNOPSIS
  pm connectors inspect trustpilot
  pm connectors inspect trustpilot --json
  pm credentials add <name> --connector trustpilot [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Trustpilot business-unit reviews, invitations, and business-unit profile metadata.

ICON
  asset: icons/trustpilot.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.trustpilot.com/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  business_unit_id
  api_key (secret)

ETL STREAMS
  reviews:
    primary key: id
    cursor: created_at
    fields: created_at(), id(), stars(), title()
  invitations:
    primary key: id
    fields: created_at(), id(), status()
  business_units:
    primary key: id
    fields: display_name(), id()
  categories:
    primary key: category_id
    fields: category_id(), display_name(), is_primary(), name(), relevance(), source()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Trustpilot API read of business-unit reviews, invitations, and profile metadata
  approval: none; read-only, no reverse-ETL writes implemented by legacy
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect trustpilot

  # Inspect as structured JSON
  pm connectors inspect trustpilot --json

AGENT WORKFLOW
  - Run pm connectors inspect trustpilot before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
