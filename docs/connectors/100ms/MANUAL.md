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
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  management_token (secret) (required)

ETL STREAMS
  rooms:
    primary key: id
    cursor: created_at
    fields: created_at(string), customer_id(string), description(string), enabled(boolean), id(string), large_room(boolean), max_duration_seconds(integer), name(string), region(string), template_id(string), updated_at(string)
  sessions:
    primary key: id
    cursor: created_at
    fields: active(boolean), created_at(string), customer_id(string), id(string), room_id(string), updated_at(string)
  recordings:
    primary key: id
    cursor: created_at
    fields: created_at(string), duration(integer), id(string), room_id(string), session_id(string), size(integer), status(string), updated_at(string)
  templates:
    primary key: id
    cursor: created_at
    fields: created_at(string), customer_id(string), default(boolean), id(string), name(string), updated_at(string)
  live_streams:
    primary key: id
    cursor: created_at
    fields: created_at(string), destination(string), id(string), meeting_url(string), room_id(string), session_id(string), started_at(string), status(string), stopped_at(string)
  external_streams:
    primary key: id
    cursor: created_at
    fields: created_at(string), destination(string), id(string), meeting_url(string), recording(boolean), room_id(string), session_id(string), started_at(string), status(string), stopped_at(string)
  recording_assets:
    primary key: id
    fields: duration(integer), id(string), job_id(string), path(string), room_id(string), session_id(string), size(integer), status(string), type(string)
  webhook_events:
    primary key: event_id
    cursor: event_timestamp
    fields: event_id(string), event_name(string), event_timestamp(string), room_id(string)

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
    required fields: name
    risk: creates a new room-policy template (roles/settings); external mutation, approval required
  create_room_code:
    endpoint: POST /room-codes/room/{{ record.room_id }}
    required fields: room_id
    risk: generates join-authentication room codes for every role in the named room; codes act as join credentials, external mutation, approval required
  update_room_code:
    endpoint: POST /room-codes/code
    required fields: code, enabled
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
