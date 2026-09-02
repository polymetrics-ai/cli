---
name: pm-google-calendar
description: Google Calendar connector knowledge and safe action guide.
---

# pm-google-calendar

## Purpose

Reads and safely reverse-ETLs Google Calendar calendars, calendar-list entries, events, ACL rules, and notification channels, plus a bounded typed free/busy query, through the Calendar API v3 using an OAuth2 refresh token.

## Icon

- id: simple-icons-googlecalendar
- asset: icons/simple-icons/googlecalendar.svg
- title: Google Calendar
- simple_icon_slug: googlecalendar
- simple_icon_hex: 4285F4
- source: simple-icons
- license: CC0-1.0
- review_status: cc0_with_trademark_caveat
- review_url: https://simpleicons.org/?q=Google%20Calendar
- match: exact-name-or-slug
- matched_by: google-calendar

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- calendarid
- event_id
- mode
- rule_id
- setting
- start_date
- client_id (secret) (required)
- client_refresh_token_2 (secret) (required)
- client_secret (secret) (required)

## ETL Streams

- calendar_list:
  - primary key: id
  - fields: accessRole(string), colorId(string), deleted(boolean), description(string), etag(string), hidden(boolean), id(string), kind(string), primary(boolean), selected(boolean), summary(string), timeZone(string)
- calendar_list_entry:
  - primary key: id
  - fields: accessRole(string), backgroundColor(string), colorId(string), conferenceProperties(object), defaultReminders(array), deleted(boolean), description(string), etag(string), foregroundColor(string), hidden(boolean), id(string), kind(string), location(string), notificationSettings(object), primary(boolean), selected(boolean), summary(string), summaryOverride(string), timeZone(string)
- calendar:
  - primary key: id
  - fields: conferenceProperties(object), description(string), etag(string), id(string), kind(string), location(string), summary(string), timeZone(string)
- colors:
  - primary key: kind
  - fields: calendar(object), event(object), kind(string), updated(string)
- events:
  - primary key: id
  - cursor: updated
  - fields: attendees(array), created(string), creator(object), description(string), end(object), etag(string), htmlLink(string), iCalUID(string), id(string), kind(string), location(string), organizer(object), recurringEventId(string), start(object), status(string), summary(string), updated(string)
- event:
  - primary key: id
  - fields: anyoneCanAddSelf(boolean), attachments(array), attendees(array), attendeesOmitted(boolean), birthdayProperties(object), colorId(string), conferenceData(object), created(string), creator(object), description(string), end(object), endTimeUnspecified(boolean), etag(string), eventType(string), extendedProperties(object), focusTimeProperties(object), gadget(object), guestsCanInviteOthers(boolean), guestsCanModify(boolean), guestsCanSeeOtherGuests(boolean), hangoutLink(string), htmlLink(string), iCalUID(string), id(string), kind(string), location(string), locked(boolean), organizer(object), originalStartTime(object), outOfOfficeProperties(object), privateCopy(boolean), recurrence(array), recurringEventId(string), reminders(object), sequence(integer), source(object), start(object), status(string), summary(string), transparency(string), updated(string), visibility(string), workingLocationProperties(object)
- event_instances:
  - primary key: id
  - fields: anyoneCanAddSelf(boolean), attachments(array), attendees(array), attendeesOmitted(boolean), birthdayProperties(object), colorId(string), conferenceData(object), created(string), creator(object), description(string), end(object), endTimeUnspecified(boolean), etag(string), eventType(string), extendedProperties(object), focusTimeProperties(object), gadget(object), guestsCanInviteOthers(boolean), guestsCanModify(boolean), guestsCanSeeOtherGuests(boolean), hangoutLink(string), htmlLink(string), iCalUID(string), id(string), kind(string), location(string), locked(boolean), organizer(object), originalStartTime(object), outOfOfficeProperties(object), privateCopy(boolean), recurrence(array), recurringEventId(string), reminders(object), sequence(integer), source(object), start(object), status(string), summary(string), transparency(string), updated(string), visibility(string), workingLocationProperties(object)
- settings:
  - primary key: id
  - fields: etag(string), id(string), kind(string), value(string)
- setting:
  - primary key: id
  - fields: etag(string), id(string), kind(string), value(string)
- acl:
  - primary key: id
  - fields: etag(string), id(string), kind(string), role(string), scope(object)
- acl_rule:
  - primary key: id
  - fields: etag(string), id(string), kind(string), role(string), scope(object)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- delete_acl_rule:
  - endpoint: DELETE /calendars/{{ record.calendar_id }}/acl/{{ record.rule_id }}
  - required fields: calendar_id, rule_id
  - risk: destructive external mutation; removes a Google Calendar ACL rule; reverse-ETL approval and typed destructive confirmation required
- insert_acl_rule:
  - endpoint: POST /calendars/{{ record.calendar_id }}/acl
  - required fields: calendar_id, role, scope
  - risk: high-impact external mutation; grants or changes calendar sharing; requires reverse-ETL plan, preview, and explicit approval
- patch_acl_rule:
  - endpoint: PATCH /calendars/{{ record.calendar_id }}/acl/{{ record.rule_id }}
  - required fields: calendar_id, rule_id, role
  - optional fields: scope
  - risk: high-impact external mutation; changes calendar sharing; requires reverse-ETL plan, preview, and explicit approval
- update_acl_rule:
  - endpoint: PUT /calendars/{{ record.calendar_id }}/acl/{{ record.rule_id }}
  - required fields: calendar_id, rule_id, role, scope
  - risk: high-impact external mutation; replaces calendar sharing settings; requires reverse-ETL plan, preview, and explicit approval
- watch_acl:
  - endpoint: POST /calendars/{{ record.calendar_id }}/acl/watch
  - required fields: calendar_id, id, type, address
  - optional fields: token
  - risk: external mutation; creates an ACL notification channel; requires reverse-ETL plan, preview, and explicit approval
  - bulk reverse ETL: refused (non-batchable; run it as its own pm command, one record at a time)
- delete_calendar_list_entry:
  - endpoint: DELETE /users/me/calendarList/{{ record.calendar_id }}
  - required fields: calendar_id
  - risk: destructive external mutation; removes a calendar from the authenticated user's calendar list; reverse-ETL approval and typed destructive confirmation required
- insert_calendar_list_entry:
  - endpoint: POST /users/me/calendarList
  - required fields: id
  - optional fields: summaryOverride, colorId, hidden, selected
  - risk: external mutation; adds a calendar to the authenticated user's calendar list; requires reverse-ETL plan, preview, and explicit approval
- patch_calendar_list_entry:
  - endpoint: PATCH /users/me/calendarList/{{ record.calendar_id }}
  - required fields: calendar_id, summaryOverride
  - optional fields: colorId, hidden, selected
  - risk: external mutation; changes a calendar-list entry; requires reverse-ETL plan, preview, and explicit approval
- update_calendar_list_entry:
  - endpoint: PUT /users/me/calendarList/{{ record.calendar_id }}
  - required fields: calendar_id, summaryOverride
  - optional fields: colorId, hidden, selected
  - risk: external mutation; replaces a calendar-list entry's writable fields; requires reverse-ETL plan, preview, and explicit approval
- watch_calendar_list:
  - endpoint: POST /users/me/calendarList/watch
  - required fields: id, type, address
  - optional fields: token
  - risk: external mutation; creates a calendar-list notification channel; requires reverse-ETL plan, preview, and explicit approval
  - bulk reverse ETL: refused (non-batchable; run it as its own pm command, one record at a time)
- clear_calendar:
  - endpoint: POST /calendars/{{ record.calendar_id }}/clear
  - required fields: calendar_id
  - risk: critical destructive external mutation; clears every event from a primary calendar; reverse-ETL approval and typed destructive confirmation required
  - bulk reverse ETL: refused (non-batchable; run it as its own pm command, one record at a time)
- delete_calendar:
  - endpoint: DELETE /calendars/{{ record.calendar_id }}
  - required fields: calendar_id
  - risk: critical destructive external mutation; deletes a secondary calendar; reverse-ETL approval and typed destructive confirmation required
  - bulk reverse ETL: refused (non-batchable; run it as its own pm command, one record at a time)
- insert_calendar:
  - endpoint: POST /calendars
  - required fields: summary
  - optional fields: description, location, timeZone
  - risk: high-impact external mutation; creates a secondary calendar owned by the authenticated user; requires reverse-ETL plan, preview, and explicit approval
- patch_calendar:
  - endpoint: PATCH /calendars/{{ record.calendar_id }}
  - required fields: calendar_id, summary
  - optional fields: description, location, timeZone
  - risk: high-impact external mutation; changes calendar metadata; requires reverse-ETL plan, preview, and explicit approval
- transfer_calendar_ownership:
  - endpoint: POST /calendars/{{ record.calendar_id }}/transferOwnership
  - required fields: calendar_id, new_data_owner, use_admin_access
  - risk: critical administrative external mutation; transfers secondary-calendar data ownership; reverse-ETL approval and typed destructive confirmation required
  - bulk reverse ETL: refused (non-batchable; run it as its own pm command, one record at a time)
- update_calendar:
  - endpoint: PUT /calendars/{{ record.calendar_id }}
  - required fields: calendar_id, summary
  - optional fields: description, location, timeZone
  - risk: high-impact external mutation; replaces calendar metadata; requires reverse-ETL plan, preview, and explicit approval
- stop_channel:
  - endpoint: POST /channels/stop
  - required fields: id, resourceId
  - risk: external mutation; stops a Google Calendar notification channel; requires reverse-ETL plan, preview, and explicit approval
  - bulk reverse ETL: refused (non-batchable; run it as its own pm command, one record at a time)
- delete_event:
  - endpoint: DELETE /calendars/{{ record.calendar_id }}/events/{{ record.event_id }}
  - required fields: calendar_id, event_id
  - risk: destructive external mutation; deletes a Google Calendar event; reverse-ETL approval and typed destructive confirmation required
- import_event:
  - endpoint: POST /calendars/{{ record.calendar_id }}/events/import
  - required fields: calendar_id, iCalUID, summary, start, end
  - optional fields: description, location, attendees, recurrence, colorId, visibility, transparency
  - risk: high-impact external mutation; imports a private event copy; requires reverse-ETL plan, preview, and explicit approval
- insert_event:
  - endpoint: POST /calendars/{{ record.calendar_id }}/events
  - required fields: calendar_id, summary, start, end
  - optional fields: description, location, attendees, recurrence, colorId, visibility, transparency
  - risk: high-impact external mutation; creates a Google Calendar event; requires reverse-ETL plan, preview, and explicit approval
- move_event:
  - endpoint: POST /calendars/{{ record.calendar_id }}/events/{{ record.event_id }}/move
  - required fields: calendar_id, event_id, destination
  - risk: high-impact external mutation; moves an event between calendars; requires reverse-ETL plan, preview, and explicit approval
  - bulk reverse ETL: refused (non-batchable; run it as its own pm command, one record at a time)
- patch_event:
  - endpoint: PATCH /calendars/{{ record.calendar_id }}/events/{{ record.event_id }}
  - required fields: calendar_id, event_id, summary, start, end
  - optional fields: description, location, attendees, recurrence, colorId, visibility, transparency
  - risk: high-impact external mutation; patches a Google Calendar event; requires reverse-ETL plan, preview, and explicit approval
- quick_add_event:
  - endpoint: POST /calendars/{{ record.calendar_id }}/events/quickAdd
  - required fields: calendar_id, text
  - risk: external mutation; creates an event from natural-language text; requires reverse-ETL plan, preview, and explicit approval
  - bulk reverse ETL: refused (non-batchable; run it as its own pm command, one record at a time)
- update_event:
  - endpoint: PUT /calendars/{{ record.calendar_id }}/events/{{ record.event_id }}
  - required fields: calendar_id, event_id, summary, start, end
  - optional fields: description, location, attendees, recurrence, colorId, visibility, transparency
  - risk: high-impact external mutation; replaces a Google Calendar event; requires reverse-ETL plan, preview, and explicit approval
- watch_events:
  - endpoint: POST /calendars/{{ record.calendar_id }}/events/watch
  - required fields: calendar_id, id, type, address
  - optional fields: token
  - risk: external mutation; creates an event notification channel; requires reverse-ETL plan, preview, and explicit approval
  - bulk reverse ETL: refused (non-batchable; run it as its own pm command, one record at a time)
- watch_settings:
  - endpoint: POST /users/me/settings/watch
  - required fields: id, type, address
  - optional fields: token
  - risk: external mutation; creates a settings notification channel; requires reverse-ETL plan, preview, and explicit approval
  - bulk reverse ETL: refused (non-batchable; run it as its own pm command, one record at a time)

## Security

- read risk: external Google Calendar API reads for the authenticated account and configured calendar; bounded direct free/busy query responses are JSON-redacted
- write risk: 26 typed Google Calendar reverse-ETL actions mutate sharing, calendar lists, calendars, events, or notification channels; destructive deletes, calendar clearing, and ownership transfer require typed destructive confirmation
- approval: every write uses plan, preview, explicit approval, and execute; destructive actions also require --confirm destructive
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Read Google Calendar data, run a bounded free/busy query, and safely plan typed Calendar reverse-ETL writes.
- Usage: pm google-calendar <command> [flags]
- Global flags:
  - --json (boolean): Write machine-readable JSON output.
  - --connection (string): Use a saved Google Calendar connector credential.: maps_to=connection
  - --credential (string): Alias for selecting a saved connector credential.: maps_to=credential
  - --config (string): Inline key=value connector configuration override; do not use for secrets.: maps_to=config
  - --limit (integer): Maximum ETL records to read for stream-backed commands.: maps_to=limit
  - --max-bytes (integer): Maximum direct-read response bytes.: maps_to=max_bytes
  - --plan (string): Execute an approved reverse-ETL plan by id.
  - --preview (boolean): Preview a reverse-ETL write command without making a network mutation.
  - --approval-token-stdin (boolean): Read the approval token as one bounded line from standard input.
  - --confirm (string): Typed confirmation challenge for destructive reverse-ETL writes.
- ACL reads and reverse-ETL actions
  - acl list - List ACL rules for config.calendarid. [intent=etl availability=implemented stream=acl]
  - acl get - Read one ACL rule for config.rule_id. [intent=etl availability=implemented stream=acl_rule]
  - acl delete - Plan deletion of an ACL rule with typed destructive confirmation. [intent=reverse_etl availability=implemented write=delete_acl_rule]; approval: Plan first, inspect preview output, then execute only with the generated approval token and --confirm destructive.; risk: Deletes calendar sharing permission; requires plan, preview, approval, and --confirm destructive.; flags: --calendar-id (required, non-empty), --rule-id (required, non-empty)
  - acl insert - Plan creation of an ACL sharing rule. [intent=reverse_etl availability=implemented write=insert_acl_rule]; approval: Plan first, inspect preview output, then execute only with the generated approval token.; risk: Grants or changes calendar sharing; requires plan, preview, approval, and execute.; flags: --calendar-id (required, non-empty), --role (required), --scope-type (required), --scope-value (non-empty)
  - acl patch - Plan a partial ACL sharing-rule update. [intent=reverse_etl availability=implemented write=patch_acl_rule]; approval: Plan first, inspect preview output, then execute only with the generated approval token.; risk: Changes calendar sharing; requires plan, preview, approval, and execute.; flags: --calendar-id (required, non-empty), --rule-id (required, non-empty), --role (required), --scope-type, --scope-value (non-empty)
  - acl update - Plan replacement of an ACL sharing rule. [intent=reverse_etl availability=implemented write=update_acl_rule]; approval: Plan first, inspect preview output, then execute only with the generated approval token.; risk: Replaces calendar sharing settings; requires plan, preview, approval, and execute.; flags: --calendar-id (required, non-empty), --rule-id (required, non-empty), --role (required), --scope-type (required), --scope-value (non-empty)
  - acl watch - Plan creation of an ACL notification channel. [intent=reverse_etl availability=implemented write=watch_acl]; approval: Plan first, inspect preview output, then execute only with the generated approval token.; risk: Creates an ACL notification channel; requires plan, preview, approval, and execute.; flags: --calendar-id (required, non-empty), --channel-id (required, non-empty), --channel-type (required), --channel-address (required, non-empty), --channel-token (non-empty)
- Calendar-list reads and reverse-ETL actions
  - calendar-list list - List calendars visible in the authenticated user calendar list. [intent=etl availability=implemented stream=calendar_list]
  - calendar-list get - Read one calendar-list entry using config.calendarid. [intent=etl availability=implemented stream=calendar_list_entry]
  - calendar-list delete - Plan removal of a calendar-list entry with typed destructive confirmation. [intent=reverse_etl availability=implemented write=delete_calendar_list_entry]; approval: Plan first, inspect preview output, then execute only with the generated approval token and --confirm destructive.; risk: Removes a calendar from the user's list; requires plan, preview, approval, and --confirm destructive.; flags: --calendar-id (required, non-empty)
  - calendar-list insert - Plan insertion of a calendar-list entry. [intent=reverse_etl availability=implemented write=insert_calendar_list_entry]; approval: Plan first, inspect preview output, then execute only with the generated approval token.; risk: Adds a calendar to the user's list; requires plan, preview, approval, and execute.; flags: --calendar-id (required, non-empty), --summary-override (non-empty), --color-id (non-empty), --hidden, --selected
  - calendar-list patch - Plan a partial calendar-list entry update. [intent=reverse_etl availability=implemented write=patch_calendar_list_entry]; approval: Plan first, inspect preview output, then execute only with the generated approval token.; risk: Changes a calendar-list entry; requires plan, preview, approval, and execute.; flags: --calendar-id (required, non-empty), --summary-override (required, non-empty), --color-id (non-empty), --hidden, --selected
  - calendar-list update - Plan replacement of calendar-list writable fields. [intent=reverse_etl availability=implemented write=update_calendar_list_entry]; approval: Plan first, inspect preview output, then execute only with the generated approval token.; risk: Replaces calendar-list writable fields; requires plan, preview, approval, and execute.; flags: --calendar-id (required, non-empty), --summary-override (required, non-empty), --color-id (non-empty), --hidden, --selected
  - calendar-list watch - Plan creation of a calendar-list notification channel. [intent=reverse_etl availability=implemented write=watch_calendar_list]; approval: Plan first, inspect preview output, then execute only with the generated approval token.; risk: Creates a calendar-list notification channel; requires plan, preview, approval, and execute.; flags: --channel-id (required, non-empty), --channel-type (required), --channel-address (required, non-empty), --channel-token (non-empty)
- Calendar reads and reverse-ETL actions
  - calendars get - Read configured calendar metadata. [intent=etl availability=implemented stream=calendar]
  - calendars clear - Plan clearing every event from a primary calendar with typed destructive confirmation. [intent=reverse_etl availability=implemented write=clear_calendar]; approval: Plan first, inspect preview output, then execute only with the generated approval token and --confirm destructive.; risk: Clears all events in a primary calendar; requires plan, preview, approval, and --confirm destructive.; flags: --calendar-id (required, non-empty)
  - calendars delete - Plan deletion of a secondary calendar with typed destructive confirmation. [intent=reverse_etl availability=implemented write=delete_calendar]; approval: Plan first, inspect preview output, then execute only with the generated approval token and --confirm destructive.; risk: Deletes a secondary calendar; requires plan, preview, approval, and --confirm destructive.; flags: --calendar-id (required, non-empty)
  - calendars insert - Plan creation of a secondary calendar. [intent=reverse_etl availability=implemented write=insert_calendar]; approval: Plan first, inspect preview output, then execute only with the generated approval token.; risk: Creates a secondary calendar; requires plan, preview, approval, and execute.; flags: --summary (required, non-empty), --description (non-empty), --location (non-empty), --time-zone (non-empty)
  - calendars patch - Plan a partial calendar metadata update. [intent=reverse_etl availability=implemented write=patch_calendar]; approval: Plan first, inspect preview output, then execute only with the generated approval token.; risk: Changes calendar metadata; requires plan, preview, approval, and execute.; flags: --calendar-id (required, non-empty), --summary (required, non-empty), --description (non-empty), --location (non-empty), --time-zone (non-empty)
  - calendars transfer-ownership - Plan transfer of a secondary calendar to a Workspace user. [intent=reverse_etl availability=implemented write=transfer_calendar_ownership]; approval: Plan first, inspect preview output, then execute only with the generated approval token and --confirm destructive.; risk: Transfers calendar data ownership; requires plan, preview, approval, and --confirm destructive.; flags: --calendar-id (required, non-empty), --new-data-owner (required, non-empty), --use-admin-access (required)
  - calendars update - Plan replacement of calendar metadata. [intent=reverse_etl availability=implemented write=update_calendar]; approval: Plan first, inspect preview output, then execute only with the generated approval token.; risk: Replaces calendar metadata; requires plan, preview, approval, and execute.; flags: --calendar-id (required, non-empty), --summary (required, non-empty), --description (non-empty), --location (non-empty), --time-zone (non-empty)
- Notification-channel reverse-ETL actions
  - channels stop - Plan stopping a Google Calendar notification channel. [intent=reverse_etl availability=implemented write=stop_channel]; approval: Plan first, inspect preview output, then execute only with the generated approval token.; risk: Stops a notification channel; requires plan, preview, approval, and execute.; flags: --channel-id (required, non-empty), --resource-id (required, non-empty)
- Color reads
  - colors get - Read color palettes for calendars and events. [intent=etl availability=implemented stream=colors]
- Event reads and reverse-ETL actions
  - events list - List events with updatedMin incremental support. [intent=etl availability=implemented stream=events]
  - events get - Read one event using config.calendarid and config.event_id. [intent=etl availability=implemented stream=event]
  - events instances - List recurring-event instances for config.event_id. [intent=etl availability=implemented stream=event_instances]
  - events delete - Plan deletion of an event with typed destructive confirmation. [intent=reverse_etl availability=implemented write=delete_event]; approval: Plan first, inspect preview output, then execute only with the generated approval token and --confirm destructive.; risk: Deletes an event; requires plan, preview, approval, and --confirm destructive.; flags: --calendar-id (required, non-empty), --event-id (required, non-empty)
  - events import - Plan import of a private event copy. [intent=reverse_etl availability=implemented write=import_event]; approval: Plan first, inspect preview output, then execute only with the generated approval token.; risk: Imports a private event copy; requires plan, preview, approval, and execute.; flags: --calendar-id (required, non-empty), --ical-uid (required, non-empty), --summary (required, non-empty), --start-date-time (required, non-empty, format=date-time), --end-date-time (required, non-empty, format=date-time), --description (non-empty), --location (non-empty), --color-id (non-empty)
  - events insert - Plan creation of a Google Calendar event. [intent=reverse_etl availability=implemented write=insert_event]; approval: Plan first, inspect preview output, then execute only with the generated approval token.; risk: Creates a calendar event; requires plan, preview, approval, and execute.; flags: --calendar-id (required, non-empty), --summary (required, non-empty), --start-date-time (required, non-empty, format=date-time), --end-date-time (required, non-empty, format=date-time), --description (non-empty), --location (non-empty), --color-id (non-empty)
  - events move - Plan moving an event to a destination calendar. [intent=reverse_etl availability=implemented write=move_event]; approval: Plan first, inspect preview output, then execute only with the generated approval token.; risk: Moves an event between calendars; requires plan, preview, approval, and execute.; flags: --calendar-id (required, non-empty), --event-id (required, non-empty), --destination (required, non-empty)
  - events patch - Plan a partial Google Calendar event update. [intent=reverse_etl availability=implemented write=patch_event]; approval: Plan first, inspect preview output, then execute only with the generated approval token.; risk: Patches a calendar event; requires plan, preview, approval, and execute.; flags: --calendar-id (required, non-empty), --event-id (required, non-empty), --summary (required, non-empty), --start-date-time (required, non-empty, format=date-time), --end-date-time (required, non-empty, format=date-time), --description (non-empty), --location (non-empty), --color-id (non-empty)
  - events quick-add - Plan creation of an event from natural-language text. [intent=reverse_etl availability=implemented write=quick_add_event]; approval: Plan first, inspect preview output, then execute only with the generated approval token.; risk: Creates an event from natural-language text; requires plan, preview, approval, and execute.; flags: --calendar-id (required, non-empty), --text (required, non-empty)
  - events update - Plan replacement of a Google Calendar event. [intent=reverse_etl availability=implemented write=update_event]; approval: Plan first, inspect preview output, then execute only with the generated approval token.; risk: Replaces a calendar event; requires plan, preview, approval, and execute.; flags: --calendar-id (required, non-empty), --event-id (required, non-empty), --summary (required, non-empty), --start-date-time (required, non-empty, format=date-time), --end-date-time (required, non-empty, format=date-time), --description (non-empty), --location (non-empty), --color-id (non-empty)
  - events watch - Plan creation of an event notification channel. [intent=reverse_etl availability=implemented write=watch_events]; approval: Plan first, inspect preview output, then execute only with the generated approval token.; risk: Creates an event notification channel; requires plan, preview, approval, and execute.; flags: --calendar-id (required, non-empty), --channel-id (required, non-empty), --channel-type (required), --channel-address (required, non-empty), --channel-token (non-empty)
- Typed direct reads
  - freebusy query - Run a bounded typed free/busy query for one calendar and time range. [intent=direct_read availability=implemented operation=google-calendar.freebusy.query]; approval: No write approval required; bounded direct read validates typed inputs.; risk: medium; flags: --calendar (required, non-empty), --time-min (required, non-empty, format=date-time), --time-max (required, non-empty, format=date-time), --time-zone (non-empty), --page, --page-cursor
- Settings reads and reverse-ETL actions
  - settings list - List user Calendar settings. [intent=etl availability=implemented stream=settings]
  - settings get - Read one user Calendar setting. [intent=etl availability=implemented stream=setting]
  - settings watch - Plan creation of a settings notification channel. [intent=reverse_etl availability=implemented write=watch_settings]; approval: Plan first, inspect preview output, then execute only with the generated approval token.; risk: Creates a settings notification channel; requires plan, preview, approval, and execute.; flags: --channel-id (required, non-empty), --channel-type (required), --channel-address (required, non-empty), --channel-token (non-empty)

## Commands

### Inspect as a manual

```bash
pm connectors inspect google-calendar
```

### Inspect as structured JSON

```bash
pm connectors inspect google-calendar --json
```

## Agent Rules

- Run pm connectors inspect google-calendar before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
