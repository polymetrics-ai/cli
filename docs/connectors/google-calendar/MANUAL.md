# pm connectors inspect google-calendar

```text
NAME
  pm connectors inspect google-calendar - Google Calendar connector manual

SYNOPSIS
  pm connectors inspect google-calendar
  pm connectors inspect google-calendar --json
  pm credentials add <name> --connector google-calendar [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Google Calendar calendars, calendar-list entries, events, recurring-event instances, colors, settings, and ACL rules, plus a bounded typed free/busy query, through the Calendar API v3 using an OAuth2 refresh token.

ICON
  id: simple-icons-googlecalendar
  asset: icons/simple-icons/googlecalendar.svg
  title: Google Calendar
  simple_icon_slug: googlecalendar
  simple_icon_hex: 4285F4
  source: simple-icons
  license: CC0-1.0
  review_status: cc0_with_trademark_caveat
  review_url: https://simpleicons.org/?q=Google%20Calendar
  match: exact-name-or-slug
  matched_by: google-calendar

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  calendarid
  event_id
  rule_id
  setting
  start_date
  client_id (secret)
  client_refresh_token_2 (secret)
  client_secret (secret)

ETL STREAMS
  calendar_list:
    primary key: id
    fields: accessRole(), colorId(), deleted(), description(), etag(), hidden(), id(), kind(), primary(), selected(), summary(), timeZone()
  calendar_list_entry:
    primary key: id
    fields: accessRole(), backgroundColor(), colorId(), conferenceProperties(), defaultReminders(), deleted(), description(), etag(), foregroundColor(), hidden(), id(), kind(), location(), notificationSettings(), primary(), selected(), summary(), summaryOverride(), timeZone()
  calendar:
    primary key: id
    fields: conferenceProperties(), description(), etag(), id(), kind(), location(), summary(), timeZone()
  colors:
    primary key: kind
    fields: calendar(), event(), kind(), updated()
  events:
    primary key: id
    cursor: updated
    fields: attendees(), created(), creator(), description(), end(), etag(), htmlLink(), iCalUID(), id(), kind(), location(), organizer(), recurringEventId(), start(), status(), summary(), updated()
  event:
    primary key: id
    fields: anyoneCanAddSelf(), attachments(), attendees(), attendeesOmitted(), birthdayProperties(), colorId(), conferenceData(), created(), creator(), description(), end(), endTimeUnspecified(), etag(), eventType(), extendedProperties(), focusTimeProperties(), gadget(), guestsCanInviteOthers(), guestsCanModify(), guestsCanSeeOtherGuests(), hangoutLink(), htmlLink(), iCalUID(), id(), kind(), location(), locked(), organizer(), originalStartTime(), outOfOfficeProperties(), privateCopy(), recurrence(), recurringEventId(), reminders(), sequence(), source(), start(), status(), summary(), transparency(), updated(), visibility(), workingLocationProperties()
  event_instances:
    primary key: id
    fields: anyoneCanAddSelf(), attachments(), attendees(), attendeesOmitted(), birthdayProperties(), colorId(), conferenceData(), created(), creator(), description(), end(), endTimeUnspecified(), etag(), eventType(), extendedProperties(), focusTimeProperties(), gadget(), guestsCanInviteOthers(), guestsCanModify(), guestsCanSeeOtherGuests(), hangoutLink(), htmlLink(), iCalUID(), id(), kind(), location(), locked(), organizer(), originalStartTime(), outOfOfficeProperties(), privateCopy(), recurrence(), recurringEventId(), reminders(), sequence(), source(), start(), status(), summary(), transparency(), updated(), visibility(), workingLocationProperties()
  settings:
    primary key: id
    fields: etag(), id(), kind(), value()
  setting:
    primary key: id
    fields: etag(), id(), kind(), value()
  acl:
    primary key: id
    fields: etag(), id(), kind(), role(), scope()
  acl_rule:
    primary key: id
    fields: etag(), id(), kind(), role(), scope()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Google Calendar API reads for the authenticated account and configured calendar; bounded direct free/busy query responses are JSON-redacted
  write risk: the provider exposes documented mutations, but this connector has no executable write action: all 26 mutation operations are ledger-blocked because the shared rest_write executor has no command-runner dispatch
  approval: no reverse-ETL write action is exposed by this connector
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Read Google Calendar data and run a bounded typed free/busy query.
  Usage: pm google-calendar <command> [flags]
  Source CLI: Google Calendar API v3 (https://www.googleapis.com/discovery/v1/apis/calendar/v3/rest)
  Global flags:
    --json (boolean): Write machine-readable JSON output.
    --connection (string): Use a saved Google Calendar connector credential.: maps_to=connection
    --credential (string): Alias for selecting a saved connector credential.: maps_to=credential
    --config (string): Inline key=value connector configuration override; do not use for secrets.: maps_to=config
    --limit (integer): Maximum ETL records to read for stream-backed commands.: maps_to=limit
    --max-bytes (integer): Maximum direct-read response bytes.: maps_to=max_bytes
  Read streams
    calendar-list list - List calendars visible in the authenticated user calendar list. [intent=etl availability=implemented stream=calendar_list]
    calendar-list get - Read one calendar-list entry using config.calendarid. [intent=etl availability=implemented stream=calendar_list_entry]
    calendars get - Read configured calendar metadata. [intent=etl availability=implemented stream=calendar]
    colors get - Read color palettes for calendars and events. [intent=etl availability=implemented stream=colors]
    events list - List events with updatedMin incremental support. [intent=etl availability=implemented stream=events]
    events get - Read one event using config.calendarid and config.event_id. [intent=etl availability=implemented stream=event]
    events instances - List recurring-event instances for config.event_id. [intent=etl availability=implemented stream=event_instances]
    settings list - List user Calendar settings. [intent=etl availability=implemented stream=settings]
    settings get - Read one user Calendar setting. [intent=etl availability=implemented stream=setting]
    acl list - List ACL rules for config.calendarid. [intent=etl availability=implemented stream=acl]
    acl get - Read one ACL rule for config.rule_id. [intent=etl availability=implemented stream=acl_rule]
  Typed direct reads
    freebusy query - Run a bounded typed free/busy query for one calendar and time range. [intent=direct_read availability=implemented operation=google-calendar.freebusy.query]; approval: No write approval required; bounded direct read validates typed inputs.; risk: medium; flags: --calendar (required), --time-min (required), --time-max (required), --time-zone

EXAMPLES
  # Inspect as a manual
  pm connectors inspect google-calendar

  # Inspect as structured JSON
  pm connectors inspect google-calendar --json

AGENT WORKFLOW
  - Run pm connectors inspect google-calendar before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
