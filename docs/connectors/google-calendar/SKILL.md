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
- client_id (secret)
- client_refresh_token_2 (secret)
- client_secret (secret)

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

- Run Google Calendar's declared streams and reverse-ETL actions.
- Usage: pm google-calendar <command> [flags]
- Read streams
- Reverse ETL writes
- Other Commands
  - acl list - Run the acl ETL stream [intent=etl availability=implemented stream=acl]; notes: discrepancy=present-in-surface-absent-from-artifact
  - acl rule list - Run the acl rule ETL stream [intent=etl availability=implemented stream=acl_rule]; notes: discrepancy=present-in-surface-absent-from-artifact
  - api delete calendars calendarid - Documented DELETE /calendars/{calendarId} (not implemented) [intent=direct_write availability=not_implemented operation=google-calendar.delete.calendars-calendarid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete calendars calendarid acl ruleid - Documented DELETE /calendars/{calendarId}/acl/{ruleId} (not implemented) [intent=direct_write availability=not_implemented operation=google-calendar.delete.calendars-calendarid-acl-ruleid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete calendars calendarid events eventid - Documented DELETE /calendars/{calendarId}/events/{eventId} (not implemented) [intent=direct_write availability=not_implemented operation=google-calendar.delete.calendars-calendarid-events-eventid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete users me calendarlist calendarid - Documented DELETE /users/me/calendarList/{calendarId} (not implemented) [intent=direct_write availability=not_implemented operation=google-calendar.delete.users-me-calendarlist-calendarid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api get calendars calendarid - Documented GET /calendars/{calendarId} (not implemented) [intent=direct_read availability=not_implemented operation=google-calendar.get.calendars-calendarid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get calendars calendarid acl - Documented GET /calendars/{calendarId}/acl (not implemented) [intent=direct_read availability=not_implemented operation=google-calendar.get.calendars-calendarid-acl]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get calendars calendarid acl ruleid - Documented GET /calendars/{calendarId}/acl/{ruleId} (not implemented) [intent=direct_read availability=not_implemented operation=google-calendar.get.calendars-calendarid-acl-ruleid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get calendars calendarid events - Documented GET /calendars/{calendarId}/events (not implemented) [intent=direct_read availability=not_implemented operation=google-calendar.get.calendars-calendarid-events]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get calendars calendarid events eventid - Documented GET /calendars/{calendarId}/events/{eventId} (not implemented) [intent=direct_read availability=not_implemented operation=google-calendar.get.calendars-calendarid-events-eventid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get calendars calendarid events eventid instances - Documented GET /calendars/{calendarId}/events/{eventId}/instances (not implemented) [intent=direct_read availability=not_implemented operation=google-calendar.get.calendars-calendarid-events-eventid-instances]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get colors - Documented GET /colors (not implemented) [intent=direct_read availability=not_implemented operation=google-calendar.get.colors]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get users me calendarlist - Documented GET /users/me/calendarList (not implemented) [intent=direct_read availability=not_implemented operation=google-calendar.get.users-me-calendarlist]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get users me calendarlist calendarid - Documented GET /users/me/calendarList/{calendarId} (not implemented) [intent=direct_read availability=not_implemented operation=google-calendar.get.users-me-calendarlist-calendarid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get users me settings - Documented GET /users/me/settings (not implemented) [intent=direct_read availability=not_implemented operation=google-calendar.get.users-me-settings]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get users me settings setting - Documented GET /users/me/settings/{setting} (not implemented) [intent=direct_read availability=not_implemented operation=google-calendar.get.users-me-settings-setting]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api patch calendars calendarid - Documented PATCH /calendars/{calendarId} (not implemented) [intent=direct_write availability=not_implemented operation=google-calendar.patch.calendars-calendarid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api patch calendars calendarid acl ruleid - Documented PATCH /calendars/{calendarId}/acl/{ruleId} (not implemented) [intent=direct_write availability=not_implemented operation=google-calendar.patch.calendars-calendarid-acl-ruleid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api patch calendars calendarid events eventid - Documented PATCH /calendars/{calendarId}/events/{eventId} (not implemented) [intent=direct_write availability=not_implemented operation=google-calendar.patch.calendars-calendarid-events-eventid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api patch users me calendarlist calendarid - Documented PATCH /users/me/calendarList/{calendarId} (not implemented) [intent=direct_write availability=not_implemented operation=google-calendar.patch.users-me-calendarlist-calendarid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post calendars - Documented POST /calendars (not implemented) [intent=direct_write availability=not_implemented operation=google-calendar.post.calendars]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post calendars calendarid acl - Documented POST /calendars/{calendarId}/acl (not implemented) [intent=direct_write availability=not_implemented operation=google-calendar.post.calendars-calendarid-acl]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post calendars calendarid acl watch - Documented POST /calendars/{calendarId}/acl/watch (not implemented) [intent=direct_write availability=not_implemented operation=google-calendar.post.calendars-calendarid-acl-watch]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post calendars calendarid clear - Documented POST /calendars/{calendarId}/clear (not implemented) [intent=direct_write availability=not_implemented operation=google-calendar.post.calendars-calendarid-clear]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post calendars calendarid events - Documented POST /calendars/{calendarId}/events (not implemented) [intent=direct_write availability=not_implemented operation=google-calendar.post.calendars-calendarid-events]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post calendars calendarid events eventid move - Documented POST /calendars/{calendarId}/events/{eventId}/move (not implemented) [intent=direct_write availability=not_implemented operation=google-calendar.post.calendars-calendarid-events-eventid-move]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post calendars calendarid events import - Documented POST /calendars/{calendarId}/events/import (not implemented) [intent=direct_write availability=not_implemented operation=google-calendar.post.calendars-calendarid-events-import]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post calendars calendarid events quickadd - Documented POST /calendars/{calendarId}/events/quickAdd (not implemented) [intent=direct_write availability=not_implemented operation=google-calendar.post.calendars-calendarid-events-quickadd]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post calendars calendarid events watch - Documented POST /calendars/{calendarId}/events/watch (not implemented) [intent=direct_write availability=not_implemented operation=google-calendar.post.calendars-calendarid-events-watch]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post calendars calendarid transferownership - Documented POST /calendars/{calendarId}/transferOwnership (not implemented) [intent=direct_write availability=not_implemented operation=google-calendar.post.calendars-calendarid-transferownership]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post channels stop - Documented POST /channels/stop (not implemented) [intent=direct_write availability=not_implemented operation=google-calendar.post.channels-stop]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post freebusy - Documented POST /freeBusy (not implemented) [intent=direct_write availability=not_implemented operation=google-calendar.post.freebusy]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post users me calendarlist - Documented POST /users/me/calendarList (not implemented) [intent=direct_write availability=not_implemented operation=google-calendar.post.users-me-calendarlist]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post users me calendarlist watch - Documented POST /users/me/calendarList/watch (not implemented) [intent=direct_write availability=not_implemented operation=google-calendar.post.users-me-calendarlist-watch]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post users me settings watch - Documented POST /users/me/settings/watch (not implemented) [intent=direct_write availability=not_implemented operation=google-calendar.post.users-me-settings-watch]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put calendars calendarid - Documented PUT /calendars/{calendarId} (not implemented) [intent=direct_write availability=not_implemented operation=google-calendar.put.calendars-calendarid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put calendars calendarid acl ruleid - Documented PUT /calendars/{calendarId}/acl/{ruleId} (not implemented) [intent=direct_write availability=not_implemented operation=google-calendar.put.calendars-calendarid-acl-ruleid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put calendars calendarid events eventid - Documented PUT /calendars/{calendarId}/events/{eventId} (not implemented) [intent=direct_write availability=not_implemented operation=google-calendar.put.calendars-calendarid-events-eventid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put users me calendarlist calendarid - Documented PUT /users/me/calendarList/{calendarId} (not implemented) [intent=direct_write availability=not_implemented operation=google-calendar.put.users-me-calendarlist-calendarid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - calendar list - Run the calendar ETL stream [intent=etl availability=implemented stream=calendar]; notes: discrepancy=present-in-surface-absent-from-artifact
  - calendar list entry list - Run the calendar list entry ETL stream [intent=etl availability=implemented stream=calendar_list_entry]; notes: discrepancy=present-in-surface-absent-from-artifact
  - calendar list list - Run the calendar list ETL stream [intent=etl availability=implemented stream=calendar_list]; notes: discrepancy=present-in-surface-absent-from-artifact
  - clear calendar apply - Plan and execute the clear calendar reverse-ETL action [intent=reverse_etl availability=implemented write=clear_calendar]; approval: requires plan, preview, approval, and execute; risk: critical destructive external mutation; clears every event from a primary calendar; reverse-ETL approval and typed destructive confirmation required; flags: --calendar_id (required)
  - colors list - Run the colors ETL stream [intent=etl availability=implemented stream=colors]; notes: discrepancy=present-in-surface-absent-from-artifact
  - delete acl rule apply - Plan and execute the delete acl rule reverse-ETL action [intent=reverse_etl availability=implemented write=delete_acl_rule]; approval: requires plan, preview, approval, and execute; risk: destructive external mutation; removes a Google Calendar ACL rule; reverse-ETL approval and typed destructive confirmation required; flags: --calendar_id (required), --rule_id (required)
  - delete calendar apply - Plan and execute the delete calendar reverse-ETL action [intent=reverse_etl availability=implemented write=delete_calendar]; approval: requires plan, preview, approval, and execute; risk: critical destructive external mutation; deletes a secondary calendar; reverse-ETL approval and typed destructive confirmation required; flags: --calendar_id (required)
  - delete calendar list entry apply - Plan and execute the delete calendar list entry reverse-ETL action [intent=reverse_etl availability=implemented write=delete_calendar_list_entry]; approval: requires plan, preview, approval, and execute; risk: destructive external mutation; removes a calendar from the authenticated user's calendar list; reverse-ETL approval and typed destructive confirmation required; flags: --calendar_id (required)
  - delete event apply - Plan and execute the delete event reverse-ETL action [intent=reverse_etl availability=implemented write=delete_event]; approval: requires plan, preview, approval, and execute; risk: destructive external mutation; deletes a Google Calendar event; reverse-ETL approval and typed destructive confirmation required; flags: --calendar_id (required), --event_id (required)
  - event instances list - Run the event instances ETL stream [intent=etl availability=implemented stream=event_instances]; notes: discrepancy=present-in-surface-absent-from-artifact
  - event list - Run the event ETL stream [intent=etl availability=implemented stream=event]; notes: discrepancy=present-in-surface-absent-from-artifact
  - events list - Run the events ETL stream [intent=etl availability=implemented stream=events]; notes: discrepancy=present-in-surface-absent-from-artifact
  - freebusy query - Run a bounded typed free/busy query for one calendar and time range. [intent=direct_read availability=implemented operation=google-calendar.freebusy.query]; approval: No write approval required; bounded direct read validates typed inputs.; risk: medium; flags: --calendar (required), --time-min (required), --time-max (required), --time-zone, --page, --page-cursor
  - import event apply - Plan and execute the import event reverse-ETL action [intent=reverse_etl availability=not_implemented write=import_event]; approval: requires plan, preview, approval, and execute; risk: high-impact external mutation; imports a private event copy; requires reverse-ETL plan, preview, and explicit approval; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - insert acl rule apply - Plan and execute the insert acl rule reverse-ETL action [intent=reverse_etl availability=not_implemented write=insert_acl_rule]; approval: requires plan, preview, approval, and execute; risk: high-impact external mutation; grants or changes calendar sharing; requires reverse-ETL plan, preview, and explicit approval; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - insert calendar apply - Plan and execute the insert calendar reverse-ETL action [intent=reverse_etl availability=implemented write=insert_calendar]; approval: requires plan, preview, approval, and execute; risk: high-impact external mutation; creates a secondary calendar owned by the authenticated user; requires reverse-ETL plan, preview, and explicit approval; flags: --summary (required)
  - insert calendar list entry apply - Plan and execute the insert calendar list entry reverse-ETL action [intent=reverse_etl availability=implemented write=insert_calendar_list_entry]; approval: requires plan, preview, approval, and execute; risk: external mutation; adds a calendar to the authenticated user's calendar list; requires reverse-ETL plan, preview, and explicit approval; flags: --id (required)
  - insert event apply - Plan and execute the insert event reverse-ETL action [intent=reverse_etl availability=not_implemented write=insert_event]; approval: requires plan, preview, approval, and execute; risk: high-impact external mutation; creates a Google Calendar event; requires reverse-ETL plan, preview, and explicit approval; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - move event apply - Plan and execute the move event reverse-ETL action [intent=reverse_etl availability=implemented write=move_event]; approval: requires plan, preview, approval, and execute; risk: high-impact external mutation; moves an event between calendars; requires reverse-ETL plan, preview, and explicit approval; flags: --calendar_id (required), --destination (required), --event_id (required)
  - patch acl rule apply - Plan and execute the patch acl rule reverse-ETL action [intent=reverse_etl availability=implemented write=patch_acl_rule]; approval: requires plan, preview, approval, and execute; risk: high-impact external mutation; changes calendar sharing; requires reverse-ETL plan, preview, and explicit approval; flags: --calendar_id (required), --role (required), --rule_id (required)
  - patch calendar apply - Plan and execute the patch calendar reverse-ETL action [intent=reverse_etl availability=implemented write=patch_calendar]; approval: requires plan, preview, approval, and execute; risk: high-impact external mutation; changes calendar metadata; requires reverse-ETL plan, preview, and explicit approval; flags: --calendar_id (required), --summary (required)
  - patch calendar list entry apply - Plan and execute the patch calendar list entry reverse-ETL action [intent=reverse_etl availability=implemented write=patch_calendar_list_entry]; approval: requires plan, preview, approval, and execute; risk: external mutation; changes a calendar-list entry; requires reverse-ETL plan, preview, and explicit approval; flags: --calendar_id (required), --summaryOverride (required)
  - patch event apply - Plan and execute the patch event reverse-ETL action [intent=reverse_etl availability=not_implemented write=patch_event]; approval: requires plan, preview, approval, and execute; risk: high-impact external mutation; patches a Google Calendar event; requires reverse-ETL plan, preview, and explicit approval; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - quick add event apply - Plan and execute the quick add event reverse-ETL action [intent=reverse_etl availability=implemented write=quick_add_event]; approval: requires plan, preview, approval, and execute; risk: external mutation; creates an event from natural-language text; requires reverse-ETL plan, preview, and explicit approval; flags: --calendar_id (required), --text (required)
  - setting list - Run the setting ETL stream [intent=etl availability=implemented stream=setting]; notes: discrepancy=present-in-surface-absent-from-artifact
  - settings list - Run the settings ETL stream [intent=etl availability=implemented stream=settings]; notes: discrepancy=present-in-surface-absent-from-artifact
  - stop channel apply - Plan and execute the stop channel reverse-ETL action [intent=reverse_etl availability=implemented write=stop_channel]; approval: requires plan, preview, approval, and execute; risk: external mutation; stops a Google Calendar notification channel; requires reverse-ETL plan, preview, and explicit approval; flags: --id (required), --resourceId (required)
  - transfer calendar ownership apply - Plan and execute the transfer calendar ownership reverse-ETL action [intent=reverse_etl availability=implemented write=transfer_calendar_ownership]; approval: requires plan, preview, approval, and execute; risk: critical administrative external mutation; transfers secondary-calendar data ownership; reverse-ETL approval and typed destructive confirmation required; flags: --calendar_id (required), --new_data_owner (required), --use_admin_access (required)
  - update acl rule apply - Plan and execute the update acl rule reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_acl_rule]; approval: requires plan, preview, approval, and execute; risk: high-impact external mutation; replaces calendar sharing settings; requires reverse-ETL plan, preview, and explicit approval; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - update calendar apply - Plan and execute the update calendar reverse-ETL action [intent=reverse_etl availability=implemented write=update_calendar]; approval: requires plan, preview, approval, and execute; risk: high-impact external mutation; replaces calendar metadata; requires reverse-ETL plan, preview, and explicit approval; flags: --calendar_id (required), --summary (required)
  - update calendar list entry apply - Plan and execute the update calendar list entry reverse-ETL action [intent=reverse_etl availability=implemented write=update_calendar_list_entry]; approval: requires plan, preview, approval, and execute; risk: external mutation; replaces a calendar-list entry's writable fields; requires reverse-ETL plan, preview, and explicit approval; flags: --calendar_id (required), --summaryOverride (required)
  - update event apply - Plan and execute the update event reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_event]; approval: requires plan, preview, approval, and execute; risk: high-impact external mutation; replaces a Google Calendar event; requires reverse-ETL plan, preview, and explicit approval; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - watch acl apply - Plan and execute the watch acl reverse-ETL action [intent=reverse_etl availability=implemented write=watch_acl]; approval: requires plan, preview, approval, and execute; risk: external mutation; creates an ACL notification channel; requires reverse-ETL plan, preview, and explicit approval; flags: --address (required), --calendar_id (required), --id (required), --type (required)
  - watch calendar list apply - Plan and execute the watch calendar list reverse-ETL action [intent=reverse_etl availability=implemented write=watch_calendar_list]; approval: requires plan, preview, approval, and execute; risk: external mutation; creates a calendar-list notification channel; requires reverse-ETL plan, preview, and explicit approval; flags: --address (required), --id (required), --type (required)
  - watch events apply - Plan and execute the watch events reverse-ETL action [intent=reverse_etl availability=implemented write=watch_events]; approval: requires plan, preview, approval, and execute; risk: external mutation; creates an event notification channel; requires reverse-ETL plan, preview, and explicit approval; flags: --address (required), --calendar_id (required), --id (required), --type (required)
  - watch settings apply - Plan and execute the watch settings reverse-ETL action [intent=reverse_etl availability=implemented write=watch_settings]; approval: requires plan, preview, approval, and execute; risk: external mutation; creates a settings notification channel; requires reverse-ETL plan, preview, and explicit approval; flags: --address (required), --id (required), --type (required)

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
