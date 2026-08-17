# pm connectors inspect safetyculture

```text
NAME
  pm connectors inspect safetyculture - SafetyCulture connector manual

SYNOPSIS
  pm connectors inspect safetyculture
  pm connectors inspect safetyculture --json
  pm credentials add <name> --connector safetyculture [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads SafetyCulture audits, templates, and users through the SafetyCulture API. Read-only. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

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
  api_key (secret)

ETL STREAMS
  audits:
    primary key: id
    fields: id(), modified_at(), name()
  templates:
    primary key: id
    fields: id(), modified_at(), name()
  users:
    primary key: id
    fields: id(), modified_at(), name()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external SafetyCulture API reads performed by the legacy connector via a Tier-2 hook
  write risk: unsupported
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect safetyculture

  # Inspect as structured JSON
  pm connectors inspect safetyculture --json

AGENT WORKFLOW
  - Run pm connectors inspect safetyculture before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
