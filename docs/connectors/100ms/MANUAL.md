# pm connectors inspect 100ms

```text
NAME
  pm connectors inspect 100ms - 100ms connector manual

SYNOPSIS
  pm connectors inspect 100ms
  pm connectors inspect 100ms --json
  pm credentials add <name> --connector 100ms [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads 100ms rooms, sessions, recordings, templates, live streams, external streams, recording assets, and webhook events, and writes room/template/room-code/recording lifecycle mutations, through the 100ms server-side REST API.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  management_token (secret)

ETL STREAMS
  rooms:
    primary key: id
    cursor: created_at
    fields: created_at(), customer_id(), description(), enabled(), id(), large_room(), max_duration_seconds(), name(), region(), template_id(), updated_at()
  sessions:
    primary key: id
    cursor: created_at
    fields: active(), created_at(), customer_id(), id(), room_id(), updated_at()
  recordings:
    primary key: id
    cursor: created_at
    fields: created_at(), duration(), id(), room_id(), session_id(), size(), status(), updated_at()
  templates:
    primary key: id
    cursor: created_at
    fields: created_at(), customer_id(), default(), id(), name(), updated_at()
  live_streams:
    primary key: id
    cursor: created_at
    fields: created_at(), destination(), id(), meeting_url(), room_id(), session_id(), started_at(), status(), stopped_at()
  external_streams:
    primary key: id
    cursor: created_at
    fields: created_at(), destination(), id(), meeting_url(), recording(), room_id(), session_id(), started_at(), status(), stopped_at()
  recording_assets:
    primary key: id
    fields: duration(), id(), job_id(), path(), room_id(), session_id(), size(), status(), type()
  webhook_events:
    primary key: event_id
    cursor: event_timestamp
    fields: event_id(), event_name(), event_timestamp(), room_id()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_room:
    endpoint: POST /rooms
    risk: creates a new 100ms room, or upserts an existing room's template if the same name is reused (100ms's own documented create-with-existing-name behavior); external mutation, approval required
  update_room:
    endpoint: POST /rooms/{{ record.id }}
    required fields: id
    risk: mutates an existing room's metadata, or disables/re-enables it via the enabled field (100ms's disable/enable API is the same POST /rooms/{id} endpoint); disabling blocks all future joins to that room. External mutation, approval required
  create_template:
    endpoint: POST /templates
    risk: creates a new room-policy template (roles/settings); external mutation, approval required
  create_room_code:
    endpoint: POST /room-codes/room/{{ record.room_id }}
    required fields: room_id
    risk: generates join-authentication room codes for every role in the named room; codes act as join credentials, external mutation, approval required
  update_room_code:
    endpoint: POST /room-codes/code
    risk: enables or disables a specific join-credential room code; disabling revokes that code's ability to join. External mutation, approval required
  start_recording:
    endpoint: POST /recordings/room/{{ record.room_id }}/start
    required fields: room_id
    optional fields: meeting_url, resolution
    risk: starts a composite recording job for the named room; consumes recording/storage quota. External mutation, approval required
  stop_recording:
    endpoint: POST /recordings/room/{{ record.room_id }}/stop
    required fields: room_id
    risk: stops all recording jobs currently running in the named room; external mutation, approval required

SECURITY
  read risk: external 100ms API read of rooms, sessions, recordings, templates, live streams, external streams, recording assets, and webhook events
  write risk: external 100ms mutation: creates/updates rooms, creates templates, creates/updates room join-codes, and starts/stops room recordings; approval required
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect 100ms

  # Inspect as structured JSON
  pm connectors inspect 100ms --json

AGENT WORKFLOW
  - Run pm connectors inspect 100ms before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
