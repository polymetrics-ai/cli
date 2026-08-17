# pm connectors inspect klaviyo

```text
NAME
  pm connectors inspect klaviyo - Klaviyo connector manual

SYNOPSIS
  pm connectors inspect klaviyo
  pm connectors inspect klaviyo --json
  pm credentials add <name> --connector klaviyo [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Klaviyo profiles, events, campaigns, lists, metrics, and segments through the Klaviyo REST (JSON:API) API.

ICON
  asset: icons/klaviyo.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.klaviyo.com/en/docs/api_versioning_and_deprecation_policy

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  revision
  api_key (secret)

ETL STREAMS
  profiles:
    primary key: id
    cursor: updated
    fields: created(), email(), external_id(), first_name(), id(), last_name(), organization(), phone_number(), type(), updated()
  events:
    primary key: id
    cursor: datetime
    fields: datetime(), id(), timestamp(), type(), uuid()
  campaigns:
    primary key: id
    cursor: updated_at
    fields: archived(), channel(), created_at(), id(), name(), scheduled_at(), send_time(), status(), type(), updated_at()
  lists:
    primary key: id
    cursor: updated
    fields: created(), id(), name(), type(), updated()
  metrics:
    primary key: id
    cursor: updated
    fields: created(), id(), integration_name(), name(), type(), updated()
  segments:
    primary key: id
    cursor: updated
    fields: created(), id(), is_active(), is_processing(), name(), type(), updated()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Klaviyo API read of customer profile, event, and campaign data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect klaviyo

  # Inspect as structured JSON
  pm connectors inspect klaviyo --json

AGENT WORKFLOW
  - Run pm connectors inspect klaviyo before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
