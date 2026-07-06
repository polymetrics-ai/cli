# pm connectors inspect onesignal

```text
NAME
  pm connectors inspect onesignal - OneSignal connector manual

SYNOPSIS
  pm connectors inspect onesignal
  pm connectors inspect onesignal --json
  pm credentials add <name> --connector onesignal [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads OneSignal account-level applications through the OneSignal REST API. Device/notification/outcome streams remain quarantined (ENGINE_GAP: no per-stream auth override).

ICON
  asset: icons/onesignal.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://documentation.onesignal.com/reference

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  user_auth_key (secret)

ETL STREAMS
  apps:
    primary key: id
    fields: created_at(), id(), messageable_players(), name(), organization_id(), players(), updated_at()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external OneSignal API read of account-level application metadata
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect onesignal

  # Inspect as structured JSON
  pm connectors inspect onesignal --json

AGENT WORKFLOW
  - Run pm connectors inspect onesignal before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
