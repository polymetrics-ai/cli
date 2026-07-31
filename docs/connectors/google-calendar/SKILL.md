---
name: pm-google-calendar
description: Google Calendar connector knowledge and safe action guide.
---

# pm-google-calendar

## Purpose

Reads Google Calendar calendars, calendar-list entries, events, recurring-event instances, colors, settings, and ACL rules; executes typed reverse-ETL mutations and bounded free/busy queries through the Calendar API v3 using an OAuth2 refresh token.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- calendarid
- event_id
- rule_id
- setting
- start_date
- client_id (secret)
- client_refresh_token_2 (secret)
- client_secret (secret)

## ETL Streams

- calendar_list:
  - primary key: id
  - fields: accessRole(), backgroundColor(), colorId(), conferenceProperties(), defaultReminders(), deleted(), description(), etag(), foregroundColor(), hidden(), id(), kind(), location(), notificationSettings(), primary(), selected(), summary(), summaryOverride(), timeZone()
- calendar_list_entry:
  - primary key: id
  - fields: accessRole(), backgroundColor(), colorId(), conferenceProperties(), defaultReminders(), deleted(), description(), etag(), foregroundColor(), hidden(), id(), kind(), location(), notificationSettings(), primary(), selected(), summary(), summaryOverride(), timeZone()
- calendar:
  - primary key: id
  - fields: conferenceProperties(), description(), etag(), id(), kind(), location(), summary(), timeZone()
- colors:
  - primary key: kind
  - fields: calendar(), event(), kind(), updated()
- events:
  - primary key: id
  - cursor: updated
  - fields: anyoneCanAddSelf(), attachments(), attendees(), attendeesOmitted(), birthdayProperties(), colorId(), conferenceData(), created(), creator(), description(), end(), endTimeUnspecified(), etag(), eventType(), extendedProperties(), focusTimeProperties(), gadget(), guestsCanInviteOthers(), guestsCanModify(), guestsCanSeeOtherGuests(), hangoutLink(), htmlLink(), iCalUID(), id(), kind(), location(), locked(), organizer(), originalStartTime(), outOfOfficeProperties(), privateCopy(), recurrence(), recurringEventId(), reminders(), sequence(), source(), start(), status(), summary(), transparency(), updated(), visibility(), workingLocationProperties()
- event:
  - primary key: id
  - fields: anyoneCanAddSelf(), attachments(), attendees(), attendeesOmitted(), birthdayProperties(), colorId(), conferenceData(), created(), creator(), description(), end(), endTimeUnspecified(), etag(), eventType(), extendedProperties(), focusTimeProperties(), gadget(), guestsCanInviteOthers(), guestsCanModify(), guestsCanSeeOtherGuests(), hangoutLink(), htmlLink(), iCalUID(), id(), kind(), location(), locked(), organizer(), originalStartTime(), outOfOfficeProperties(), privateCopy(), recurrence(), recurringEventId(), reminders(), sequence(), source(), start(), status(), summary(), transparency(), updated(), visibility(), workingLocationProperties()
- event_instances:
  - primary key: id
  - fields: anyoneCanAddSelf(), attachments(), attendees(), attendeesOmitted(), birthdayProperties(), colorId(), conferenceData(), created(), creator(), description(), end(), endTimeUnspecified(), etag(), eventType(), extendedProperties(), focusTimeProperties(), gadget(), guestsCanInviteOthers(), guestsCanModify(), guestsCanSeeOtherGuests(), hangoutLink(), htmlLink(), iCalUID(), id(), kind(), location(), locked(), organizer(), originalStartTime(), outOfOfficeProperties(), privateCopy(), recurrence(), recurringEventId(), reminders(), sequence(), source(), start(), status(), summary(), transparency(), updated(), visibility(), workingLocationProperties()
- settings:
  - primary key: id
  - fields: etag(), id(), kind(), value()
- setting:
  - primary key: id
  - fields: etag(), id(), kind(), value()
- acl:
  - primary key: id
  - fields: etag(), id(), kind(), role(), scope()
- acl_rule:
  - primary key: id
  - fields: etag(), id(), kind(), role(), scope()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- delete_acl_rule:
  - endpoint: DELETE /calendars/{{ record.calendar_id }}/acl/{{ record.rule_id }}
  - required fields: calendar_id, rule_id
  - risk: permanently removes a Google Calendar ACL rule from the target calendar
- create_acl_rule:
  - endpoint: POST /calendars/{{ record.calendar_id }}/acl
  - required fields: calendar_id, role, scope
  - risk: creates a visible calendar sharing ACL rule
- patch_acl_rule:
  - endpoint: PATCH /calendars/{{ record.calendar_id }}/acl/{{ record.rule_id }}
  - required fields: calendar_id, rule_id
  - risk: patches an existing calendar sharing ACL rule
- update_acl_rule:
  - endpoint: PUT /calendars/{{ record.calendar_id }}/acl/{{ record.rule_id }}
  - required fields: calendar_id, rule_id, role, scope
  - risk: replaces an existing calendar sharing ACL rule
- delete_calendar_list_entry:
  - endpoint: DELETE /users/me/calendarList/{{ record.calendar_id }}
  - required fields: calendar_id
  - risk: removes a calendar from the authenticated user calendar list
- create_calendar_list_entry:
  - endpoint: POST /users/me/calendarList
  - required fields: id
  - risk: adds an existing calendar to the authenticated user calendar list
- patch_calendar_list_entry:
  - endpoint: PATCH /users/me/calendarList/{{ record.calendar_id }}
  - required fields: calendar_id
  - risk: patches an authenticated-user calendar-list entry
- update_calendar_list_entry:
  - endpoint: PUT /users/me/calendarList/{{ record.calendar_id }}
  - required fields: calendar_id
  - risk: replaces an authenticated-user calendar-list entry
- clear_calendar:
  - endpoint: POST /calendars/{{ record.calendar_id }}/clear
  - required fields: calendar_id
  - risk: removes all events from the target primary calendar while preserving the calendar
- delete_calendar:
  - endpoint: DELETE /calendars/{{ record.calendar_id }}
  - required fields: calendar_id
  - risk: permanently deletes the target secondary calendar
- create_calendar:
  - endpoint: POST /calendars
  - required fields: summary
  - risk: creates a secondary Google Calendar
- patch_calendar:
  - endpoint: PATCH /calendars/{{ record.calendar_id }}
  - required fields: calendar_id
  - risk: patches calendar metadata
- transfer_calendar_ownership:
  - endpoint: POST /calendars/{{ record.calendar_id }}/transferOwnership?newDataOwner={{ record.newDataOwner }}
  - required fields: calendar_id, newDataOwner
  - risk: transfers secondary-calendar ownership to another data owner via Calendar API transferOwnership
- update_calendar:
  - endpoint: PUT /calendars/{{ record.calendar_id }}
  - required fields: calendar_id, summary
  - risk: replaces calendar metadata
- stop_channel:
  - endpoint: POST /channels/stop
  - required fields: id, resourceId
  - risk: stops a Google Calendar notification channel for a known channel id/resource id pair
- delete_event:
  - endpoint: DELETE /calendars/{{ record.calendar_id }}/events/{{ record.event_id }}
  - required fields: calendar_id, event_id
  - risk: deletes or cancels a Google Calendar event
- import_event:
  - endpoint: POST /calendars/{{ record.calendar_id }}/events/import
  - required fields: calendar_id, summary, start, end
  - risk: imports a private copy of an event into the target calendar
- create_event:
  - endpoint: POST /calendars/{{ record.calendar_id }}/events
  - required fields: calendar_id, summary, start, end
  - risk: creates a Google Calendar event
- move_event:
  - endpoint: POST /calendars/{{ record.calendar_id }}/events/{{ record.event_id }}/move?destination={{ record.destination_calendar_id }}
  - required fields: calendar_id, event_id, destination_calendar_id
  - risk: moves an event to another calendar
- patch_event:
  - endpoint: PATCH /calendars/{{ record.calendar_id }}/events/{{ record.event_id }}
  - required fields: calendar_id, event_id
  - risk: patches a Google Calendar event
- quick_add_event:
  - endpoint: POST /calendars/{{ record.calendar_id }}/events/quickAdd?text={{ record.text }}
  - required fields: calendar_id, text
  - risk: creates an event from natural-language text using Google Calendar quickAdd
- update_event:
  - endpoint: PUT /calendars/{{ record.calendar_id }}/events/{{ record.event_id }}
  - required fields: calendar_id, event_id, summary, start, end
  - risk: replaces a Google Calendar event
- watch_acl:
  - endpoint: POST /calendars/{{ record.calendar_id }}/acl/watch
  - required fields: calendar_id, id, type, address
  - risk: creates a notification channel for ACL changes; webhook delivery consumption is outside this connector slice
- watch_calendar_list:
  - endpoint: POST /users/me/calendarList/watch
  - required fields: id, type, address
  - risk: creates a notification channel for calendar-list changes; webhook delivery consumption is outside this connector slice
- watch_events:
  - endpoint: POST /calendars/{{ record.calendar_id }}/events/watch
  - required fields: calendar_id, id, type, address
  - risk: creates a notification channel for event changes; webhook delivery consumption is outside this connector slice
- watch_settings:
  - endpoint: POST /users/me/settings/watch
  - required fields: id, type, address
  - risk: creates a notification channel for Calendar settings changes; webhook delivery consumption is outside this connector slice

## Security

- read risk: external Google Calendar API reads for the authenticated account and configured calendar; bounded direct free/busy query responses are JSON-redacted
- write risk: external Google Calendar mutations including event creation/update/deletion, calendar metadata changes, ACL sharing changes, calendar-list edits, notification-channel watch setup, and channel stop requests
- approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive deletes, calendar clearing, ownership transfer, and channel stop actions require destructive confirmation
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Work with Google Calendar data and typed Calendar API operations.
- Usage: pm google-calendar <command> [flags]
- Source CLI: Google Calendar API v3 (https://www.googleapis.com/discovery/v1/apis/calendar/v3/rest)
- Global flags:
  - --json (boolean): Write machine-readable JSON output.
  - --connection (string): Use a saved Google Calendar connector credential.: maps_to=connection
  - --credential (string): Alias for selecting a saved connector credential.: maps_to=credential
  - --config (string): Inline key=value connector configuration override; do not use for secrets.: maps_to=config
  - --limit (integer): Maximum ETL records to read for stream-backed commands.: maps_to=limit
  - --max-bytes (integer): Maximum direct-read response bytes.: maps_to=max_bytes
  - --plan (boolean): Build a reverse-ETL write plan without executing.: maps_to=plan
  - --preview (boolean): Preview a reverse-ETL write request before approval/execution.: maps_to=preview
  - --approve (string): Explicit approval token for reverse-ETL execution.: maps_to=approval
  - --confirm (string): Typed destructive confirmation challenge.: maps_to=confirm
  - --plan-name (string): Stable name for a generated reverse-ETL plan.: maps_to=plan_name
- Read streams
  - calendar-list list - List calendars visible in the authenticated user calendar list. [intent=etl availability=implemented stream=calendar_list]
  - calendar-list get - Read one calendar-list entry using config.calendarid. [intent=etl availability=implemented stream=calendar_list_entry]
  - calendars get - Read configured calendar metadata. [intent=etl availability=implemented stream=calendar]
  - colors get - Read color palettes for calendars and events. [intent=etl availability=implemented stream=colors]
  - events list - List events with updatedMin incremental support. [intent=etl availability=implemented stream=events]
  - events get - Read one event using config.calendarid and config.event_id. [intent=etl availability=implemented stream=event]
  - events instances - List recurring-event instances for config.event_id. [intent=etl availability=implemented stream=event_instances]
  - settings list - List user Calendar settings. [intent=etl availability=implemented stream=settings]
  - settings get - Read one user Calendar setting. [intent=etl availability=implemented stream=setting]
  - acl list - List ACL rules for config.calendarid. [intent=etl availability=implemented stream=acl]
  - acl get - Read one ACL rule for config.rule_id. [intent=etl availability=implemented stream=acl_rule]
- Typed direct reads
  - freebusy query - Run a bounded typed free/busy query for one calendar and time range. [intent=direct_read availability=implemented]; approval: No write approval required; bounded direct read validates typed inputs.; risk: medium; flags: --calendar, --time-min, --time-max, --time-zone
- Reverse ETL writes
  - delete-acl-rule - permanently removes a Google Calendar ACL rule from the target calendar [intent=reverse_etl availability=implemented write=delete_acl_rule]; approval: Requires reverse ETL plan, preview, explicit approval, and execute. Destructive confirmation is required.; risk: permanently removes a Google Calendar ACL rule from the target calendar; flags: --calendar-id, --rule-id
  - create-acl-rule - creates a visible calendar sharing ACL rule [intent=reverse_etl availability=implemented write=create_acl_rule]; approval: Requires reverse ETL plan, preview, explicit approval, and execute.; risk: creates a visible calendar sharing ACL rule; flags: --calendar-id, --role, --scope-type, --scope-value
  - patch-acl-rule - patches an existing calendar sharing ACL rule [intent=reverse_etl availability=implemented write=patch_acl_rule]; approval: Requires reverse ETL plan, preview, explicit approval, and execute.; risk: patches an existing calendar sharing ACL rule; flags: --calendar-id, --rule-id, --role
  - update-acl-rule - replaces an existing calendar sharing ACL rule [intent=reverse_etl availability=implemented write=update_acl_rule]; approval: Requires reverse ETL plan, preview, explicit approval, and execute.; risk: replaces an existing calendar sharing ACL rule; flags: --calendar-id, --rule-id, --role, --scope-type, --scope-value
  - watch-acl - creates a notification channel for ACL changes; webhook delivery consumption is outside this connector slice [intent=reverse_etl availability=implemented write=watch_acl]; approval: Requires reverse ETL plan, preview, explicit approval, and execute.; risk: creates a notification channel for ACL changes; webhook delivery consumption is outside this connector slice; flags: --calendar-id, --id, --type, --address
  - delete-calendar-list-entry - removes a calendar from the authenticated user calendar list [intent=reverse_etl availability=implemented write=delete_calendar_list_entry]; approval: Requires reverse ETL plan, preview, explicit approval, and execute. Destructive confirmation is required.; risk: removes a calendar from the authenticated user calendar list; flags: --calendar-id
  - create-calendar-list-entry - adds an existing calendar to the authenticated user calendar list [intent=reverse_etl availability=implemented write=create_calendar_list_entry]; approval: Requires reverse ETL plan, preview, explicit approval, and execute.; risk: adds an existing calendar to the authenticated user calendar list; flags: --id
  - patch-calendar-list-entry - patches an authenticated-user calendar-list entry [intent=reverse_etl availability=implemented write=patch_calendar_list_entry]; approval: Requires reverse ETL plan, preview, explicit approval, and execute.; risk: patches an authenticated-user calendar-list entry; flags: --calendar-id, --summaryOverride
  - update-calendar-list-entry - replaces an authenticated-user calendar-list entry [intent=reverse_etl availability=implemented write=update_calendar_list_entry]; approval: Requires reverse ETL plan, preview, explicit approval, and execute.; risk: replaces an authenticated-user calendar-list entry; flags: --calendar-id, --summaryOverride
  - watch-calendar-list - creates a notification channel for calendar-list changes; webhook delivery consumption is outside this connector slice [intent=reverse_etl availability=implemented write=watch_calendar_list]; approval: Requires reverse ETL plan, preview, explicit approval, and execute.; risk: creates a notification channel for calendar-list changes; webhook delivery consumption is outside this connector slice; flags: --id, --type, --address
  - clear-calendar - removes all events from the target primary calendar while preserving the calendar [intent=reverse_etl availability=implemented write=clear_calendar]; approval: Requires reverse ETL plan, preview, explicit approval, and execute. Destructive confirmation is required.; risk: removes all events from the target primary calendar while preserving the calendar; flags: --calendar-id
  - delete-calendar - permanently deletes the target secondary calendar [intent=reverse_etl availability=implemented write=delete_calendar]; approval: Requires reverse ETL plan, preview, explicit approval, and execute. Destructive confirmation is required.; risk: permanently deletes the target secondary calendar; flags: --calendar-id
  - create-calendar - creates a secondary Google Calendar [intent=reverse_etl availability=implemented write=create_calendar]; approval: Requires reverse ETL plan, preview, explicit approval, and execute.; risk: creates a secondary Google Calendar; flags: --summary
  - patch-calendar - patches calendar metadata [intent=reverse_etl availability=implemented write=patch_calendar]; approval: Requires reverse ETL plan, preview, explicit approval, and execute.; risk: patches calendar metadata; flags: --calendar-id, --summary
  - transfer-calendar-ownership - transfers secondary-calendar ownership to another data owner via Calendar API transferOwnership [intent=reverse_etl availability=implemented write=transfer_calendar_ownership]; approval: Requires reverse ETL plan, preview, explicit approval, and execute. Destructive confirmation is required.; risk: transfers secondary-calendar ownership to another data owner via Calendar API transferOwnership; flags: --calendar-id, --newDataOwner
  - update-calendar - replaces calendar metadata [intent=reverse_etl availability=implemented write=update_calendar]; approval: Requires reverse ETL plan, preview, explicit approval, and execute.; risk: replaces calendar metadata; flags: --calendar-id, --summary
  - stop-channel - stops a Google Calendar notification channel for a known channel id/resource id pair [intent=reverse_etl availability=implemented write=stop_channel]; approval: Requires reverse ETL plan, preview, explicit approval, and execute. Destructive confirmation is required.; risk: stops a Google Calendar notification channel for a known channel id/resource id pair; flags: --id, --resourceId
  - delete-event - deletes or cancels a Google Calendar event [intent=reverse_etl availability=implemented write=delete_event]; approval: Requires reverse ETL plan, preview, explicit approval, and execute. Destructive confirmation is required.; risk: deletes or cancels a Google Calendar event; flags: --calendar-id, --event-id
  - import-event - imports a private copy of an event into the target calendar [intent=reverse_etl availability=implemented write=import_event]; approval: Requires reverse ETL plan, preview, explicit approval, and execute.; risk: imports a private copy of an event into the target calendar; flags: --calendar-id, --summary, --start-dateTime, --end-dateTime
  - create-event - creates a Google Calendar event [intent=reverse_etl availability=implemented write=create_event]; approval: Requires reverse ETL plan, preview, explicit approval, and execute.; risk: creates a Google Calendar event; flags: --calendar-id, --summary, --start-dateTime, --end-dateTime
  - move-event - moves an event to another calendar [intent=reverse_etl availability=implemented write=move_event]; approval: Requires reverse ETL plan, preview, explicit approval, and execute. Destructive confirmation is required.; risk: moves an event to another calendar; flags: --calendar-id, --event-id, --destination-calendar-id
  - patch-event - patches a Google Calendar event [intent=reverse_etl availability=implemented write=patch_event]; approval: Requires reverse ETL plan, preview, explicit approval, and execute.; risk: patches a Google Calendar event; flags: --calendar-id, --event-id, --summary
  - quick-add-event - creates an event from natural-language text using Google Calendar quickAdd [intent=reverse_etl availability=implemented write=quick_add_event]; approval: Requires reverse ETL plan, preview, explicit approval, and execute.; risk: creates an event from natural-language text using Google Calendar quickAdd; flags: --calendar-id, --text
  - update-event - replaces a Google Calendar event [intent=reverse_etl availability=implemented write=update_event]; approval: Requires reverse ETL plan, preview, explicit approval, and execute.; risk: replaces a Google Calendar event; flags: --calendar-id, --event-id, --summary, --start-dateTime, --end-dateTime
  - watch-events - creates a notification channel for event changes; webhook delivery consumption is outside this connector slice [intent=reverse_etl availability=implemented write=watch_events]; approval: Requires reverse ETL plan, preview, explicit approval, and execute.; risk: creates a notification channel for event changes; webhook delivery consumption is outside this connector slice; flags: --calendar-id, --id, --type, --address
  - watch-settings - creates a notification channel for Calendar settings changes; webhook delivery consumption is outside this connector slice [intent=reverse_etl availability=implemented write=watch_settings]; approval: Requires reverse ETL plan, preview, explicit approval, and execute.; risk: creates a notification channel for Calendar settings changes; webhook delivery consumption is outside this connector slice; flags: --id, --type, --address
- Help topics:
  - google-calendar-auth - Configure OAuth2 refresh-token credentials without printing secrets.
  - google-calendar-safety - Reverse ETL writes require plan, preview, approval, and execute.

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
