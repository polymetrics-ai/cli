# pm connectors inspect posthog

```text
NAME
  pm connectors inspect posthog - PostHog connector manual

SYNOPSIS
  pm connectors inspect posthog
  pm connectors inspect posthog --json
  pm credentials add <name> --connector posthog [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads PostHog events and persons for a project via the PostHog REST API. Read-only.

ICON
  asset: icons/posthog.svg
  source: official
  review_status: official_verified
  review_url: https://posthog.com/docs/api

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  page_size
  project_id
  start_date
  api_key (secret)

ETL STREAMS
  events:
    primary key: id
    cursor: timestamp
    fields: distinct_id(), event(), id(), properties(), timestamp()
  persons:
    primary key: id
    fields: created_at(), distinct_id(), id(), properties()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external PostHog API read of project event and person data
  approval: none; read-only analytics API
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect posthog

  # Inspect as structured JSON
  pm connectors inspect posthog --json

AGENT WORKFLOW
  - Run pm connectors inspect posthog before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
