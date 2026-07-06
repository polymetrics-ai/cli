# pm connectors inspect dixa

```text
NAME
  pm connectors inspect dixa - Dixa connector manual

SYNOPSIS
  pm connectors inspect dixa
  pm connectors inspect dixa --json
  pm credentials add <name> --connector dixa [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Dixa conversations (and their queue, rating, and assignment projections) from the Dixa conversation_export API. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

ICON
  asset: icons/dixa.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://docs.dixa.io/openapi/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  batch_size
  mode
  start_date
  api_token (secret)

ETL STREAMS
  conversations:
    primary key: id
    cursor: updated_at
    fields: closed_at(), created_at(), direction(), handling_duration(), id(), initial_channel(), last_message_created_at(), originating_country(), requester_email(), requester_id(), requester_name(), status(), subject(), total_duration(), updated_at()
  conversation_queue:
    primary key: id
    cursor: updated_at
    fields: direction(), id(), initial_channel(), queue_id(), queue_name(), queued_at(), updated_at()
  conversation_rating:
    primary key: id
    cursor: updated_at
    fields: id(), rating_message(), rating_score(), status(), updated_at()
  conversation_assignment:
    primary key: id
    cursor: updated_at
    fields: assigned_at(), assignee_email(), assignee_id(), assignee_name(), id(), transfer_time(), transferee_name(), updated_at()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Dixa API reads performed by the legacy connector via a Tier-2 hook
  write risk: unsupported
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect dixa

  # Inspect as structured JSON
  pm connectors inspect dixa --json

AGENT WORKFLOW
  - Run pm connectors inspect dixa before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
