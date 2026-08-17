# pm connectors inspect cal-com

```text
NAME
  pm connectors inspect cal-com - Cal.com connector manual

SYNOPSIS
  pm connectors inspect cal-com
  pm connectors inspect cal-com --json
  pm credentials add <name> --connector cal-com [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Cal.com bookings, event types, schedules, webhooks, and profile, and manages bookings/event types/schedules/webhooks through the Cal.com v2 REST API.

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
  api_version
  base_url
  api_key (secret)

ETL STREAMS
  bookings:
    primary key: id
    fields: createdAt(), description(), end(), eventTypeId(), id(), start(), status(), title(), uid(), updatedAt()
  schedules:
    primary key: id
    fields: id(), isDefault(), name(), ownerId(), timeZone()
  event_types:
    primary key: id
    fields: description(), hidden(), id(), length(), position(), slug(), title()
  webhooks:
    primary key: id
    fields: active(), id(), payloadTemplate(), secret(), subscriberUrl(), triggers(), userId()
  my_profile:
    primary key: id
    fields: email(), id(), name(), timeFormat(), timeZone(), username(), weekStart()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_booking:
    endpoint: POST /v2/bookings
    risk: external mutation; books a real meeting slot on the target event type and notifies attendees; approval required
  cancel_booking:
    endpoint: POST /v2/bookings/{{ record.uid }}/cancel
    required fields: uid
    risk: external mutation; cancels a real booking and notifies attendees; approval required
  confirm_booking:
    endpoint: POST /v2/bookings/{{ record.uid }}/confirm
    required fields: uid
    risk: external mutation; confirms a booking pending host approval, notifying the attendee; approval required
  decline_booking:
    endpoint: POST /v2/bookings/{{ record.uid }}/decline
    required fields: uid
    risk: external mutation; declines a booking pending host approval, notifying the attendee; approval required
  reschedule_booking:
    endpoint: POST /v2/bookings/{{ record.uid }}/reschedule
    required fields: uid
    risk: external mutation; moves a real booking to a new time and notifies attendees; approval required
  create_event_type:
    endpoint: POST /v2/event-types
    risk: external mutation; creates a new publicly-bookable event type; approval required
  update_event_type:
    endpoint: PATCH /v2/event-types/{{ record.id }}
    required fields: id
    risk: external mutation; changes the public scheduling configuration of an existing event type; approval required
  delete_event_type:
    endpoint: DELETE /v2/event-types/{{ record.id }}
    required fields: id
    risk: destructive; permanently deletes an event type, breaking any existing public booking links; approval required
  create_schedule:
    endpoint: POST /v2/schedules
    risk: external mutation; creates a new availability schedule, which can be attached to event types and change public availability; approval required
  update_schedule:
    endpoint: PATCH /v2/schedules/{{ record.id }}
    required fields: id
    risk: external mutation; changes a real availability schedule's hours/timezone, directly affecting public bookable slots; approval required
  delete_schedule:
    endpoint: DELETE /v2/schedules/{{ record.id }}
    required fields: id
    risk: destructive; permanently deletes an availability schedule; approval required
  create_webhook:
    endpoint: POST /v2/webhooks
    risk: external mutation; registers a new webhook endpoint that will receive live booking event payloads; approval required
  delete_webhook:
    endpoint: DELETE /v2/webhooks/{{ record.id }}
    required fields: id
    risk: destructive; permanently deletes a webhook subscription; approval required

SECURITY
  read risk: external Cal.com API read of bookings, event types, schedules, webhooks, and profile data
  write risk: external mutation of live scheduling data: creates/cancels/confirms/declines/reschedules real bookings (notifying attendees), creates/updates/deletes event types and availability schedules (changes public booking availability), and creates/deletes webhook subscriptions
  approval: required for every write action
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect cal-com

  # Inspect as structured JSON
  pm connectors inspect cal-com --json

AGENT WORKFLOW
  - Run pm connectors inspect cal-com before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
