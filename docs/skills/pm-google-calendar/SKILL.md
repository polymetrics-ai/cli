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
- Source CLI: Google Calendar API v3 (https://www.googleapis.com/discovery/v1/apis/calendar/v3/rest)
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
  - acl delete - Plan deletion of an ACL rule with typed destructive confirmation. [intent=reverse_etl availability=implemented write=delete_acl_rule]; approval: Plan first, inspect preview output, then execute only with the generated approval token and --confirm destructive.; risk: Deletes calendar sharing permission; requires plan, preview, approval, and --confirm destructive.; flags: --calendar-id (required, non-empty) (string): Calendar identifier.: maps_to=record.calendar_id, --rule-id (required, non-empty) (string): ACL rule identifier.: maps_to=record.rule_id
  - acl insert - Plan creation of an ACL sharing rule. [intent=reverse_etl availability=implemented write=insert_acl_rule]; approval: Plan first, inspect preview output, then execute only with the generated approval token.; risk: Grants or changes calendar sharing; requires plan, preview, approval, and execute.; flags: --calendar-id (required, non-empty) (string): Calendar identifier.: maps_to=record.calendar_id, --role (required) (enum): ACL role.: values=none|freeBusyReader|reader|writer|owner: maps_to=record.role, --scope-type (required) (enum): ACL scope type.: values=default|user|group|domain: maps_to=record.scope.type, --scope-value (non-empty) (string): ACL scope value for user, group, or domain scopes.: maps_to=record.scope.value
  - acl patch - Plan a partial ACL sharing-rule update. [intent=reverse_etl availability=implemented write=patch_acl_rule]; approval: Plan first, inspect preview output, then execute only with the generated approval token.; risk: Changes calendar sharing; requires plan, preview, approval, and execute.; flags: --calendar-id (required, non-empty) (string): Calendar identifier.: maps_to=record.calendar_id, --rule-id (required, non-empty) (string): ACL rule identifier.: maps_to=record.rule_id, --role (required) (enum): Replacement ACL role.: values=none|freeBusyReader|reader|writer|owner: maps_to=record.role, --scope-type (enum): Optional ACL scope type.: values=default|user|group|domain: maps_to=record.scope.type, --scope-value (non-empty) (string): Optional ACL scope value.: maps_to=record.scope.value
  - acl update - Plan replacement of an ACL sharing rule. [intent=reverse_etl availability=implemented write=update_acl_rule]; approval: Plan first, inspect preview output, then execute only with the generated approval token.; risk: Replaces calendar sharing settings; requires plan, preview, approval, and execute.; flags: --calendar-id (required, non-empty) (string): Calendar identifier.: maps_to=record.calendar_id, --rule-id (required, non-empty) (string): ACL rule identifier.: maps_to=record.rule_id, --role (required) (enum): Replacement ACL role.: values=none|freeBusyReader|reader|writer|owner: maps_to=record.role, --scope-type (required) (enum): ACL scope type.: values=default|user|group|domain: maps_to=record.scope.type, --scope-value (non-empty) (string): ACL scope value for user, group, or domain scopes.: maps_to=record.scope.value
  - acl watch - Plan creation of an ACL notification channel. [intent=reverse_etl availability=implemented write=watch_acl]; approval: Plan first, inspect preview output, then execute only with the generated approval token.; risk: Creates an ACL notification channel; requires plan, preview, approval, and execute.; flags: --calendar-id (required, non-empty) (string): Calendar identifier.: maps_to=record.calendar_id, --channel-id (required, non-empty) (string): Unique channel identifier.: maps_to=record.id, --channel-type (required) (enum): Google channel transport type.: values=web_hook: maps_to=record.type, --channel-address (required, non-empty) (string): HTTPS webhook callback URL.: maps_to=record.address, --channel-token (non-empty) (string): Optional opaque channel token.: maps_to=record.token
- Calendar-list reads and reverse-ETL actions
  - calendar-list list - List calendars visible in the authenticated user calendar list. [intent=etl availability=implemented stream=calendar_list]
  - calendar-list get - Read one calendar-list entry using config.calendarid. [intent=etl availability=implemented stream=calendar_list_entry]
  - calendar-list delete - Plan removal of a calendar-list entry with typed destructive confirmation. [intent=reverse_etl availability=implemented write=delete_calendar_list_entry]; approval: Plan first, inspect preview output, then execute only with the generated approval token and --confirm destructive.; risk: Removes a calendar from the user's list; requires plan, preview, approval, and --confirm destructive.; flags: --calendar-id (required, non-empty) (string): Calendar identifier.: maps_to=record.calendar_id
  - calendar-list insert - Plan insertion of a calendar-list entry. [intent=reverse_etl availability=implemented write=insert_calendar_list_entry]; approval: Plan first, inspect preview output, then execute only with the generated approval token.; risk: Adds a calendar to the user's list; requires plan, preview, approval, and execute.; flags: --calendar-id (required, non-empty) (string): Existing calendar ID to add.: maps_to=record.id, --summary-override (non-empty) (string): Optional display-name override.: maps_to=record.summaryOverride, --color-id (non-empty) (string): Optional calendar-list color ID.: maps_to=record.colorId, --hidden (boolean): Optional hidden state.: maps_to=record.hidden, --selected (boolean): Optional selected state.: maps_to=record.selected
  - calendar-list patch - Plan a partial calendar-list entry update. [intent=reverse_etl availability=implemented write=patch_calendar_list_entry]; approval: Plan first, inspect preview output, then execute only with the generated approval token.; risk: Changes a calendar-list entry; requires plan, preview, approval, and execute.; flags: --calendar-id (required, non-empty) (string): Calendar identifier.: maps_to=record.calendar_id, --summary-override (required, non-empty) (string): Replacement display-name override.: maps_to=record.summaryOverride, --color-id (non-empty) (string): Optional calendar-list color ID.: maps_to=record.colorId, --hidden (boolean): Optional hidden state.: maps_to=record.hidden, --selected (boolean): Optional selected state.: maps_to=record.selected
  - calendar-list update - Plan replacement of calendar-list writable fields. [intent=reverse_etl availability=implemented write=update_calendar_list_entry]; approval: Plan first, inspect preview output, then execute only with the generated approval token.; risk: Replaces calendar-list writable fields; requires plan, preview, approval, and execute.; flags: --calendar-id (required, non-empty) (string): Calendar identifier.: maps_to=record.calendar_id, --summary-override (required, non-empty) (string): Replacement display-name override.: maps_to=record.summaryOverride, --color-id (non-empty) (string): Optional calendar-list color ID.: maps_to=record.colorId, --hidden (boolean): Optional hidden state.: maps_to=record.hidden, --selected (boolean): Optional selected state.: maps_to=record.selected
  - calendar-list watch - Plan creation of a calendar-list notification channel. [intent=reverse_etl availability=implemented write=watch_calendar_list]; approval: Plan first, inspect preview output, then execute only with the generated approval token.; risk: Creates a calendar-list notification channel; requires plan, preview, approval, and execute.; flags: --channel-id (required, non-empty) (string): Unique channel identifier.: maps_to=record.id, --channel-type (required) (enum): Google channel transport type.: values=web_hook: maps_to=record.type, --channel-address (required, non-empty) (string): HTTPS webhook callback URL.: maps_to=record.address, --channel-token (non-empty) (string): Optional opaque channel token.: maps_to=record.token
- Calendar reads and reverse-ETL actions
  - calendars get - Read configured calendar metadata. [intent=etl availability=implemented stream=calendar]
  - calendars clear - Plan clearing every event from a primary calendar with typed destructive confirmation. [intent=reverse_etl availability=implemented write=clear_calendar]; approval: Plan first, inspect preview output, then execute only with the generated approval token and --confirm destructive.; risk: Clears all events in a primary calendar; requires plan, preview, approval, and --confirm destructive.; flags: --calendar-id (required, non-empty) (string): Primary calendar identifier.: maps_to=record.calendar_id
  - calendars delete - Plan deletion of a secondary calendar with typed destructive confirmation. [intent=reverse_etl availability=implemented write=delete_calendar]; approval: Plan first, inspect preview output, then execute only with the generated approval token and --confirm destructive.; risk: Deletes a secondary calendar; requires plan, preview, approval, and --confirm destructive.; flags: --calendar-id (required, non-empty) (string): Secondary calendar identifier.: maps_to=record.calendar_id
  - calendars insert - Plan creation of a secondary calendar. [intent=reverse_etl availability=implemented write=insert_calendar]; approval: Plan first, inspect preview output, then execute only with the generated approval token.; risk: Creates a secondary calendar; requires plan, preview, approval, and execute.; flags: --summary (required, non-empty) (string): Calendar display name.: maps_to=record.summary, --description (non-empty) (string): Optional calendar description.: maps_to=record.description, --location (non-empty) (string): Optional calendar location.: maps_to=record.location, --time-zone (non-empty) (string): Optional IANA calendar timezone.: maps_to=record.timeZone
  - calendars patch - Plan a partial calendar metadata update. [intent=reverse_etl availability=implemented write=patch_calendar]; approval: Plan first, inspect preview output, then execute only with the generated approval token.; risk: Changes calendar metadata; requires plan, preview, approval, and execute.; flags: --calendar-id (required, non-empty) (string): Calendar identifier.: maps_to=record.calendar_id, --summary (required, non-empty) (string): Replacement calendar display name.: maps_to=record.summary, --description (non-empty) (string): Optional calendar description.: maps_to=record.description, --location (non-empty) (string): Optional calendar location.: maps_to=record.location, --time-zone (non-empty) (string): Optional IANA calendar timezone.: maps_to=record.timeZone
  - calendars transfer-ownership - Plan transfer of a secondary calendar to a Workspace user. [intent=reverse_etl availability=implemented write=transfer_calendar_ownership]; approval: Plan first, inspect preview output, then execute only with the generated approval token and --confirm destructive.; risk: Transfers calendar data ownership; requires plan, preview, approval, and --confirm destructive.; flags: --calendar-id (required, non-empty) (string): Secondary calendar identifier.: maps_to=record.calendar_id, --new-data-owner (required, non-empty) (string): Workspace user email that becomes the data owner.: maps_to=record.new_data_owner, --use-admin-access (required) (boolean): Required administrator-privilege acknowledgement.: maps_to=record.use_admin_access
  - calendars update - Plan replacement of calendar metadata. [intent=reverse_etl availability=implemented write=update_calendar]; approval: Plan first, inspect preview output, then execute only with the generated approval token.; risk: Replaces calendar metadata; requires plan, preview, approval, and execute.; flags: --calendar-id (required, non-empty) (string): Calendar identifier.: maps_to=record.calendar_id, --summary (required, non-empty) (string): Replacement calendar display name.: maps_to=record.summary, --description (non-empty) (string): Optional calendar description.: maps_to=record.description, --location (non-empty) (string): Optional calendar location.: maps_to=record.location, --time-zone (non-empty) (string): Optional IANA calendar timezone.: maps_to=record.timeZone
- Notification-channel reverse-ETL actions
  - channels stop - Plan stopping a Google Calendar notification channel. [intent=reverse_etl availability=implemented write=stop_channel]; approval: Plan first, inspect preview output, then execute only with the generated approval token.; risk: Stops a notification channel; requires plan, preview, approval, and execute.; flags: --channel-id (required, non-empty) (string): Channel identifier.: maps_to=record.id, --resource-id (required, non-empty) (string): Provider resource identifier returned with the channel.: maps_to=record.resourceId
- Color reads
  - colors get - Read color palettes for calendars and events. [intent=etl availability=implemented stream=colors]
- Event reads and reverse-ETL actions
  - events list - List events with updatedMin incremental support. [intent=etl availability=implemented stream=events]
  - events get - Read one event using config.calendarid and config.event_id. [intent=etl availability=implemented stream=event]
  - events instances - List recurring-event instances for config.event_id. [intent=etl availability=implemented stream=event_instances]
  - events delete - Plan deletion of an event with typed destructive confirmation. [intent=reverse_etl availability=implemented write=delete_event]; approval: Plan first, inspect preview output, then execute only with the generated approval token and --confirm destructive.; risk: Deletes an event; requires plan, preview, approval, and --confirm destructive.; flags: --calendar-id (required, non-empty) (string): Calendar identifier.: maps_to=record.calendar_id, --event-id (required, non-empty) (string): Event identifier.: maps_to=record.event_id
  - events import - Plan import of a private event copy. [intent=reverse_etl availability=implemented write=import_event]; approval: Plan first, inspect preview output, then execute only with the generated approval token.; risk: Imports a private event copy; requires plan, preview, approval, and execute.; flags: --calendar-id (required, non-empty) (string): Calendar identifier.: maps_to=record.calendar_id, --ical-uid (required, non-empty) (string): RFC 5545 iCalendar event UID.: maps_to=record.iCalUID, --summary (required, non-empty) (string): Event summary.: maps_to=record.summary, --start-date-time (required, non-empty, format=date-time) (string): RFC3339 event start.: maps_to=record.start.dateTime, --end-date-time (required, non-empty, format=date-time) (string): RFC3339 event end.: maps_to=record.end.dateTime, --description (non-empty) (string): Optional event description.: maps_to=record.description, --location (non-empty) (string): Optional event location.: maps_to=record.location, --color-id (non-empty) (string): Optional event color ID.: maps_to=record.colorId
  - events insert - Plan creation of a Google Calendar event. [intent=reverse_etl availability=implemented write=insert_event]; approval: Plan first, inspect preview output, then execute only with the generated approval token.; risk: Creates a calendar event; requires plan, preview, approval, and execute.; flags: --calendar-id (required, non-empty) (string): Calendar identifier.: maps_to=record.calendar_id, --summary (required, non-empty) (string): Event summary.: maps_to=record.summary, --start-date-time (required, non-empty, format=date-time) (string): RFC3339 event start.: maps_to=record.start.dateTime, --end-date-time (required, non-empty, format=date-time) (string): RFC3339 event end.: maps_to=record.end.dateTime, --description (non-empty) (string): Optional event description.: maps_to=record.description, --location (non-empty) (string): Optional event location.: maps_to=record.location, --color-id (non-empty) (string): Optional event color ID.: maps_to=record.colorId
  - events move - Plan moving an event to a destination calendar. [intent=reverse_etl availability=implemented write=move_event]; approval: Plan first, inspect preview output, then execute only with the generated approval token.; risk: Moves an event between calendars; requires plan, preview, approval, and execute.; flags: --calendar-id (required, non-empty) (string): Source calendar identifier.: maps_to=record.calendar_id, --event-id (required, non-empty) (string): Event identifier.: maps_to=record.event_id, --destination (required, non-empty) (string): Destination calendar identifier.: maps_to=record.destination
  - events patch - Plan a partial Google Calendar event update. [intent=reverse_etl availability=implemented write=patch_event]; approval: Plan first, inspect preview output, then execute only with the generated approval token.; risk: Patches a calendar event; requires plan, preview, approval, and execute.; flags: --calendar-id (required, non-empty) (string): Calendar identifier.: maps_to=record.calendar_id, --event-id (required, non-empty) (string): Event identifier.: maps_to=record.event_id, --summary (required, non-empty) (string): Replacement event summary.: maps_to=record.summary, --start-date-time (required, non-empty, format=date-time) (string): RFC3339 event start.: maps_to=record.start.dateTime, --end-date-time (required, non-empty, format=date-time) (string): RFC3339 event end.: maps_to=record.end.dateTime, --description (non-empty) (string): Optional event description.: maps_to=record.description, --location (non-empty) (string): Optional event location.: maps_to=record.location, --color-id (non-empty) (string): Optional event color ID.: maps_to=record.colorId
  - events quick-add - Plan creation of an event from natural-language text. [intent=reverse_etl availability=implemented write=quick_add_event]; approval: Plan first, inspect preview output, then execute only with the generated approval token.; risk: Creates an event from natural-language text; requires plan, preview, approval, and execute.; flags: --calendar-id (required, non-empty) (string): Calendar identifier.: maps_to=record.calendar_id, --text (required, non-empty) (string): Natural-language event description.: maps_to=record.text
  - events update - Plan replacement of a Google Calendar event. [intent=reverse_etl availability=implemented write=update_event]; approval: Plan first, inspect preview output, then execute only with the generated approval token.; risk: Replaces a calendar event; requires plan, preview, approval, and execute.; flags: --calendar-id (required, non-empty) (string): Calendar identifier.: maps_to=record.calendar_id, --event-id (required, non-empty) (string): Event identifier.: maps_to=record.event_id, --summary (required, non-empty) (string): Replacement event summary.: maps_to=record.summary, --start-date-time (required, non-empty, format=date-time) (string): RFC3339 event start.: maps_to=record.start.dateTime, --end-date-time (required, non-empty, format=date-time) (string): RFC3339 event end.: maps_to=record.end.dateTime, --description (non-empty) (string): Optional event description.: maps_to=record.description, --location (non-empty) (string): Optional event location.: maps_to=record.location, --color-id (non-empty) (string): Optional event color ID.: maps_to=record.colorId
  - events watch - Plan creation of an event notification channel. [intent=reverse_etl availability=implemented write=watch_events]; approval: Plan first, inspect preview output, then execute only with the generated approval token.; risk: Creates an event notification channel; requires plan, preview, approval, and execute.; flags: --calendar-id (required, non-empty) (string): Calendar identifier.: maps_to=record.calendar_id, --channel-id (required, non-empty) (string): Unique channel identifier.: maps_to=record.id, --channel-type (required) (enum): Google channel transport type.: values=web_hook: maps_to=record.type, --channel-address (required, non-empty) (string): HTTPS webhook callback URL.: maps_to=record.address, --channel-token (non-empty) (string): Optional opaque channel token.: maps_to=record.token
- Typed direct reads
  - freebusy query - Run a bounded typed free/busy query for one calendar and time range. [intent=direct_read availability=implemented operation=google-calendar.freebusy.query]; approval: No write approval required; bounded direct read validates typed inputs.; risk: medium; flags: --calendar (required, non-empty) (string): Calendar ID to query.: maps_to=body.items.0.id, --time-min (required, non-empty, format=date-time) (string): RFC3339 lower time bound.: maps_to=body.timeMin, --time-max (required, non-empty, format=date-time) (string): RFC3339 upper time bound.: maps_to=body.timeMax, --time-zone (non-empty) (string): Optional response timezone.: maps_to=body.timeZone, --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
- Settings reads and reverse-ETL actions
  - settings list - List user Calendar settings. [intent=etl availability=implemented stream=settings]
  - settings get - Read one user Calendar setting. [intent=etl availability=implemented stream=setting]
  - settings watch - Plan creation of a settings notification channel. [intent=reverse_etl availability=implemented write=watch_settings]; approval: Plan first, inspect preview output, then execute only with the generated approval token.; risk: Creates a settings notification channel; requires plan, preview, approval, and execute.; flags: --channel-id (required, non-empty) (string): Unique channel identifier.: maps_to=record.id, --channel-type (required) (enum): Google channel transport type.: values=web_hook: maps_to=record.type, --channel-address (required, non-empty) (string): HTTPS webhook callback URL.: maps_to=record.address, --channel-token (non-empty) (string): Optional opaque channel token.: maps_to=record.token

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
