# pm connectors inspect fastly

```text
NAME
  pm connectors inspect fastly - Fastly connector manual

SYNOPSIS
  pm connectors inspect fastly
  pm connectors inspect fastly --json
  pm credentials add <name> --connector fastly [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Fastly services, the current user, the current customer (account), and datacenters through the Fastly REST API. Read-only.

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
  fastly_api_token (secret)

ETL STREAMS
  services:
    primary key: id
    cursor: updated_at
    fields: comment(), created_at(), customer_id(), deleted_at(), id(), name(), paused(), type(), updated_at(), version()
  current_user:
    primary key: id
    fields: created_at(), customer_id(), email_hash(), id(), locked(), login(), name(), role(), two_factor_auth_enabled(), updated_at()
  current_customer:
    primary key: id
    fields: billing_contact_id(), can_stream_syslog(), created_at(), has_account_panel(), id(), name(), owner_id(), pricing_plan(), updated_at()
  datacenters:
    primary key: code
    fields: code(), coordinates(), group(), name(), shield()
  service_details:
    primary key: service_id
    fields: activated_version(), comment(), created_at(), customer_id(), deleted_at(), environments(), id(), name(), service_id(), type(), updated_at(), version(), versions()
  users:
    primary key: id
    fields: created_at(), customer_id(), email_hash(), id(), locked(), login(), name(), role(), two_factor_auth_enabled(), updated_at()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Fastly API read of service/account configuration metadata
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect fastly

  # Inspect as structured JSON
  pm connectors inspect fastly --json

AGENT WORKFLOW
  - Run pm connectors inspect fastly before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
