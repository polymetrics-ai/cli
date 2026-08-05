# Google Calendar request-field research

Audited: 2026-08-05. Provider operation inventory: [Google Calendar API v3 Discovery, revision 20260731](https://www.googleapis.com/discovery/v1/apis/calendar/v3/rest). This is a convention-neutral research ledger until the shared citation-schema lane lands; it cites every request-field use currently declared by this bundle and records the source section, evidence type, confidence, and requiredness rationale.

| Declared use | Field mapping | Provider source and section | Evidence | Confidence | Requiredness rationale |
| --- | --- | --- | --- | --- |
| `base.check` | `query.maxResults=1` | [calendarList.list](https://developers.google.com/workspace/calendar/api/v3/reference/calendarList/list), Parameters | provider reference | high | Provider-optional; static bounded check policy. |
| `calendar_list` | `query.maxResults=250` | [calendarList.list](https://developers.google.com/workspace/calendar/api/v3/reference/calendarList/list), Parameters | provider reference | high | Provider-optional; static bounded-page policy. |
| `calendar_list` | runtime `query.pageToken` | [calendarList.list](https://developers.google.com/workspace/calendar/api/v3/reference/calendarList/list), Parameters | provider reference | high | Provider-optional continuation token, emitted only after a response token. |
| `calendar_list_entry` | `path.calendarId ← config.calendarid` | [calendarList.get](https://developers.google.com/workspace/calendar/api/v3/reference/calendarList/get), Parameters | provider reference | high | Provider-required path parameter. |
| `calendar` | `path.calendarId ← config.calendarid` | [calendars.get](https://developers.google.com/workspace/calendar/api/v3/reference/calendars/get), Parameters | provider reference | high | Provider-required path parameter. |
| `colors` | no declared request field | [colors.get](https://developers.google.com/workspace/calendar/api/v3/reference/colors/get), method reference | provider reference | high | Operation-level source retained although this fixed endpoint accepts no bundle request mapping. |
| `events` | `path.calendarId ← config.calendarid` | [events.list](https://developers.google.com/workspace/calendar/api/v3/reference/events/list), Parameters | provider reference | high | Provider-required path parameter. |
| `events` | `query.maxResults=250` | [events.list](https://developers.google.com/workspace/calendar/api/v3/reference/events/list), Parameters | provider reference | high | Provider-optional; static bounded-page policy. |
| `events` | runtime `query.pageToken` | [events.list](https://developers.google.com/workspace/calendar/api/v3/reference/events/list), Parameters | provider reference | high | Provider-optional continuation token, emitted only after a response token. |
| `events` | `query.updatedMin ← config.start_date` | [events.list](https://developers.google.com/workspace/calendar/api/v3/reference/events/list), Parameters | provider reference | high | Provider-optional lower bound, emitted only when explicitly configured or restored from state; an unconfigured fresh read is unfiltered. |
| `event` | `path.calendarId ← config.calendarid` | [events.get](https://developers.google.com/workspace/calendar/api/v3/reference/events/get), Parameters | provider reference | high | Provider-required path parameter. |
| `event` | `path.eventId ← config.event_id` | [events.get](https://developers.google.com/workspace/calendar/api/v3/reference/events/get), Parameters | provider reference | high | Provider-required path parameter. |
| `event_instances` | `path.calendarId ← config.calendarid` | [events.instances](https://developers.google.com/workspace/calendar/api/v3/reference/events/instances), Parameters | provider reference | high | Provider-required path parameter. |
| `event_instances` | `path.eventId ← config.event_id` | [events.instances](https://developers.google.com/workspace/calendar/api/v3/reference/events/instances), Parameters | provider reference | high | Provider-required path parameter. |
| `event_instances` | `query.maxResults=250` | [events.instances](https://developers.google.com/workspace/calendar/api/v3/reference/events/instances), Parameters | provider reference | high | Provider-optional; static bounded-page policy. |
| `event_instances` | runtime `query.pageToken` | [events.instances](https://developers.google.com/workspace/calendar/api/v3/reference/events/instances), Parameters | provider reference | high | Provider-optional continuation token, emitted only after a response token. |
| `setting` | `path.setting ← config.setting` | [settings.get](https://developers.google.com/workspace/calendar/api/v3/reference/settings/get), Parameters | provider reference | high | Provider-required path parameter. |
| `settings` | `query.maxResults=250` | [settings.list](https://developers.google.com/workspace/calendar/api/v3/reference/settings/list), Parameters | provider reference | high | Provider-optional; static bounded-page policy at the documented maximum. |
| `settings` | runtime `query.pageToken` | [settings.list](https://developers.google.com/workspace/calendar/api/v3/reference/settings/list), Parameters | provider reference | high | Provider-optional continuation token, emitted only after a response token. |
| `acl` | `path.calendarId ← config.calendarid` | [acl.list](https://developers.google.com/workspace/calendar/api/v3/reference/acl/list), Parameters | provider reference | high | Provider-required path parameter. |
| `acl` | `query.maxResults=250` | [acl.list](https://developers.google.com/workspace/calendar/api/v3/reference/acl/list), Parameters | provider reference | high | Provider-optional; static bounded-page policy. |
| `acl` | runtime `query.pageToken` | [acl.list](https://developers.google.com/workspace/calendar/api/v3/reference/acl/list), Parameters | provider reference | high | Provider-optional continuation token, emitted only after a response token. |
| `acl_rule` | `path.calendarId ← config.calendarid` | [acl.get](https://developers.google.com/workspace/calendar/api/v3/reference/acl/get), Parameters | provider reference | high | Provider-required path parameter. |
| `acl_rule` | `path.ruleId ← config.rule_id` | [acl.get](https://developers.google.com/workspace/calendar/api/v3/reference/acl/get), Parameters | provider reference | high | Provider-required path parameter. |
| `freebusy query` | `body.items[0].id ← --calendar` | [freebusy.query](https://developers.google.com/workspace/calendar/api/v3/reference/freebusy/query), Request body `items[].id` | provider reference + bundle schema | high | Provider documents a calendar-ID item. The CLI requires one item as a bounded typed-command policy. |
| `freebusy query` | `body.timeMin ← --time-min` | [freebusy.query](https://developers.google.com/workspace/calendar/api/v3/reference/freebusy/query), Request body `timeMin` | provider reference + bundle schema | high | Provider documents the lower bound. The CLI requires it for a bounded query even though Discovery's request schema has no global `required` list. |
| `freebusy query` | `body.timeMax ← --time-max` | [freebusy.query](https://developers.google.com/workspace/calendar/api/v3/reference/freebusy/query), Request body `timeMax` | provider reference + bundle schema | high | Provider documents the upper bound. The CLI requires it and validates `timeMin < timeMax` as bounded-command policy. |
| `freebusy query` | `body.timeZone ← --time-zone` | [freebusy.query](https://developers.google.com/workspace/calendar/api/v3/reference/freebusy/query), Request body `timeZone` | provider reference + bundle schema | high | Provider-optional response timezone. |

## Executable write-action request-field matrix

The 2026-08-05 contract correction establishes that record-driven `writes.json` actions execute today. Every documented mutation below was therefore assessed as a record action, not a `rest_write` candidate: **all 26 fit and are authored**. Each table row is one declared request-field use with its primary-provider citation, section, evidence type, confidence, and requiredness rationale. The referenced operation pages use the Discovery request parameter/resource names; the connector's record names are shown after `←`.

| Operation | Field mapping | Provider source and section | Evidence | Confidence | Requiredness rationale |
| --- | --- | --- | --- | --- | --- |
| `acl.delete` | `path.calendarId ← record.calendar_id` | [acl.delete], Parameters > `calendarId` | provider reference | high | Provider-required path parameter. |
| `acl.delete` | `path.ruleId ← record.rule_id` | [acl.delete], Parameters > `ruleId` | provider reference | high | Provider-required path parameter. |
| `acl.insert` | `path.calendarId ← record.calendar_id` | [acl.insert], Parameters > `calendarId` | provider reference | high | Provider-required path parameter. |
| `acl.insert` | `body.role ← record.role` | [acl.insert], Request body > `AclRule.role` | provider reference | high | Required by the typed action to create a sharing rule. |
| `acl.insert` | `body.scope ← record.scope` | [acl.insert], Request body > `AclRule.scope` | provider reference | high | Required by the typed action to name the sharing subject. |
| `acl.patch` | `path.calendarId ← record.calendar_id` | [acl.patch], Parameters > `calendarId` | provider reference | high | Provider-required path parameter. |
| `acl.patch` | `path.ruleId ← record.rule_id` | [acl.patch], Parameters > `ruleId` | provider reference | high | Provider-required path parameter. |
| `acl.patch` | `body.role ← record.role` | [acl.patch], Request body > `AclRule.role` | provider reference | high | Required typed patch field. |
| `acl.patch` | `body.scope ← record.scope` | [acl.patch], Request body > `AclRule.scope` | provider reference | high | Optional provider body field. |
| `acl.update` | `path.calendarId ← record.calendar_id` | [acl.update], Parameters > `calendarId` | provider reference | high | Provider-required path parameter. |
| `acl.update` | `path.ruleId ← record.rule_id` | [acl.update], Parameters > `ruleId` | provider reference | high | Provider-required path parameter. |
| `acl.update` | `body.role ← record.role` | [acl.update], Request body > `AclRule.role` | provider reference | high | Required by the typed replacement action. |
| `acl.update` | `body.scope ← record.scope` | [acl.update], Request body > `AclRule.scope` | provider reference | high | Required by the typed replacement action. |
| `acl.watch` | `path.calendarId ← record.calendar_id` | [acl.watch], Parameters > `calendarId` | provider reference | high | Provider-required path parameter. |
| `acl.watch` | `body.id ← record.id` | [acl.watch], Request body > `Channel.id` | provider reference | high | Required typed channel identifier. |
| `acl.watch` | `body.type ← record.type` | [acl.watch], Request body > `Channel.type` | provider reference | high | Required typed channel transport (`web_hook`). |
| `acl.watch` | `body.address ← record.address` | [acl.watch], Request body > `Channel.address` | provider reference | high | Required typed webhook address. |
| `acl.watch` | `body.token ← record.token` | [acl.watch], Request body > `Channel.token` | provider reference | high | Optional provider channel token. |
| `calendarList.delete` | `path.calendarId ← record.calendar_id` | [calendar-list.delete], Parameters > `calendarId` | provider reference | high | Provider-required path parameter. |
| `calendarList.insert` | `body.id ← record.id` | [calendar-list.insert], Request body > `CalendarListEntry.id` | provider reference | high | Required by the typed action to add an existing calendar. |
| `calendarList.insert` | `body.summaryOverride ← record.summaryOverride` | [calendar-list.insert], Request body > `CalendarListEntry.summaryOverride` | provider reference | high | Optional provider body field. |
| `calendarList.insert` | `body.colorId ← record.colorId` | [calendar-list.insert], Request body > `CalendarListEntry.colorId` | provider reference | high | Optional provider body field. |
| `calendarList.insert` | `body.hidden ← record.hidden` | [calendar-list.insert], Request body > `CalendarListEntry.hidden` | provider reference | high | Optional provider body field. |
| `calendarList.insert` | `body.selected ← record.selected` | [calendar-list.insert], Request body > `CalendarListEntry.selected` | provider reference | high | Optional provider body field. |
| `calendarList.patch` | `path.calendarId ← record.calendar_id` | [calendar-list.patch], Parameters > `calendarId` | provider reference | high | Provider-required path parameter. |
| `calendarList.patch` | `body.summaryOverride ← record.summaryOverride` | [calendar-list.patch], Request body > `CalendarListEntry.summaryOverride` | provider reference | high | Required typed patch field. |
| `calendarList.patch` | `body.colorId ← record.colorId` | [calendar-list.patch], Request body > `CalendarListEntry.colorId` | provider reference | high | Optional provider body field. |
| `calendarList.patch` | `body.hidden ← record.hidden` | [calendar-list.patch], Request body > `CalendarListEntry.hidden` | provider reference | high | Optional provider body field. |
| `calendarList.patch` | `body.selected ← record.selected` | [calendar-list.patch], Request body > `CalendarListEntry.selected` | provider reference | high | Optional provider body field. |
| `calendarList.update` | `path.calendarId ← record.calendar_id` | [calendar-list.update], Parameters > `calendarId` | provider reference | high | Provider-required path parameter. |
| `calendarList.update` | `body.summaryOverride ← record.summaryOverride` | [calendar-list.update], Request body > `CalendarListEntry.summaryOverride` | provider reference | high | Required by the typed replacement action. |
| `calendarList.update` | `body.colorId ← record.colorId` | [calendar-list.update], Request body > `CalendarListEntry.colorId` | provider reference | high | Optional provider body field. |
| `calendarList.update` | `body.hidden ← record.hidden` | [calendar-list.update], Request body > `CalendarListEntry.hidden` | provider reference | high | Optional provider body field. |
| `calendarList.update` | `body.selected ← record.selected` | [calendar-list.update], Request body > `CalendarListEntry.selected` | provider reference | high | Optional provider body field. |
| `calendarList.watch` | `body.id ← record.id` | [calendar-list.watch], Request body > `Channel.id` | provider reference | high | Required typed channel identifier. |
| `calendarList.watch` | `body.type ← record.type` | [calendar-list.watch], Request body > `Channel.type` | provider reference | high | Required typed channel transport (`web_hook`). |
| `calendarList.watch` | `body.address ← record.address` | [calendar-list.watch], Request body > `Channel.address` | provider reference | high | Required typed webhook address. |
| `calendarList.watch` | `body.token ← record.token` | [calendar-list.watch], Request body > `Channel.token` | provider reference | high | Optional provider channel token. |
| `calendars.clear` | `path.calendarId ← record.calendar_id` | [calendars.clear], Parameters > `calendarId` | provider reference | high | Provider-required path parameter. |
| `calendars.delete` | `path.calendarId ← record.calendar_id` | [calendars.delete], Parameters > `calendarId` | provider reference | high | Provider-required path parameter. |
| `calendars.insert` | `body.summary ← record.summary` | [calendars.insert], Request body > `Calendar.summary` | provider reference | high | Required by the typed create action. |
| `calendars.insert` | `body.description ← record.description` | [calendars.insert], Request body > `Calendar.description` | provider reference | high | Optional provider body field. |
| `calendars.insert` | `body.location ← record.location` | [calendars.insert], Request body > `Calendar.location` | provider reference | high | Optional provider body field. |
| `calendars.insert` | `body.timeZone ← record.timeZone` | [calendars.insert], Request body > `Calendar.timeZone` | provider reference | high | Optional provider body field. |
| `calendars.patch` | `path.calendarId ← record.calendar_id` | [calendars.patch], Parameters > `calendarId` | provider reference | high | Provider-required path parameter. |
| `calendars.patch` | `body.summary ← record.summary` | [calendars.patch], Request body > `Calendar.summary` | provider reference | high | Required typed patch field. |
| `calendars.patch` | `body.description ← record.description` | [calendars.patch], Request body > `Calendar.description` | provider reference | high | Optional provider body field. |
| `calendars.patch` | `body.location ← record.location` | [calendars.patch], Request body > `Calendar.location` | provider reference | high | Optional provider body field. |
| `calendars.patch` | `body.timeZone ← record.timeZone` | [calendars.patch], Request body > `Calendar.timeZone` | provider reference | high | Optional provider body field. |
| `calendars.transferOwnership` | `path.calendarId ← record.calendar_id` | [calendars.transfer-ownership], Parameters > `calendarId` | provider reference | high | Provider-required path parameter. |
| `calendars.transferOwnership` | `query.newDataOwner ← record.new_data_owner` | [calendars.transfer-ownership], Parameters > `newDataOwner` | provider reference | high | Provider-required query parameter. |
| `calendars.transferOwnership` | `query.useAdminAccess ← record.use_admin_access` | [calendars.transfer-ownership], Parameters > `useAdminAccess` | provider reference | high | Provider-required query parameter; action requires `true`. |
| `calendars.update` | `path.calendarId ← record.calendar_id` | [calendars.update], Parameters > `calendarId` | provider reference | high | Provider-required path parameter. |
| `calendars.update` | `body.summary ← record.summary` | [calendars.update], Request body > `Calendar.summary` | provider reference | high | Required by the typed replacement action. |
| `calendars.update` | `body.description ← record.description` | [calendars.update], Request body > `Calendar.description` | provider reference | high | Optional provider body field. |
| `calendars.update` | `body.location ← record.location` | [calendars.update], Request body > `Calendar.location` | provider reference | high | Optional provider body field. |
| `calendars.update` | `body.timeZone ← record.timeZone` | [calendars.update], Request body > `Calendar.timeZone` | provider reference | high | Optional provider body field. |
| `channels.stop` | `body.id ← record.id` | [channels.stop], Request body > `Channel.id` | provider reference | high | Required by the typed channel-stop action. |
| `channels.stop` | `body.resourceId ← record.resourceId` | [channels.stop], Request body > `Channel.resourceId` | provider reference | high | Required by the typed channel-stop action. |
| `events.delete` | `path.calendarId ← record.calendar_id` | [events.delete], Parameters > `calendarId` | provider reference | high | Provider-required path parameter. |
| `events.delete` | `path.eventId ← record.event_id` | [events.delete], Parameters > `eventId` | provider reference | high | Provider-required path parameter. |
| `events.import` | `path.calendarId ← record.calendar_id` | [events.import], Parameters > `calendarId` | provider reference | high | Provider-required path parameter. |
| `events.import` | `body.iCalUID ← record.iCalUID` | [events.import], Request body > `Event.iCalUID` | provider reference | high | Provider-required import identifier; uniquely identifies the event across calendaring systems. |
| `events.import` | `body.summary ← record.summary` | [events.import], Request body > `Event.summary` | provider reference | high | Required by the typed import action. |
| `events.import` | `body.description ← record.description` | [events.import], Request body > `Event.description` | provider reference | high | Optional provider body field. |
| `events.import` | `body.location ← record.location` | [events.import], Request body > `Event.location` | provider reference | high | Optional provider body field. |
| `events.import` | `body.start ← record.start` | [events.import], Request body > `Event.start` | provider reference | high | Required by the typed import action; CLI maps RFC3339 `start.dateTime`. |
| `events.import` | `body.end ← record.end` | [events.import], Request body > `Event.end` | provider reference | high | Required by the typed import action; CLI maps RFC3339 `end.dateTime`. |
| `events.import` | `body.attendees ← record.attendees` | [events.import], Request body > `Event.attendees` | provider reference | high | Optional provider body field. |
| `events.import` | `body.recurrence ← record.recurrence` | [events.import], Request body > `Event.recurrence` | provider reference | high | Optional provider body field. |
| `events.import` | `body.colorId ← record.colorId` | [events.import], Request body > `Event.colorId` | provider reference | high | Optional provider body field. |
| `events.import` | `body.visibility ← record.visibility` | [events.import], Request body > `Event.visibility` | provider reference | high | Optional provider body field. |
| `events.import` | `body.transparency ← record.transparency` | [events.import], Request body > `Event.transparency` | provider reference | high | Optional provider body field. |
| `events.insert` | `path.calendarId ← record.calendar_id` | [events.insert], Parameters > `calendarId` | provider reference | high | Provider-required path parameter. |
| `events.insert` | `body.summary ← record.summary` | [events.insert], Request body > `Event.summary` | provider reference | high | Required by the typed create action. |
| `events.insert` | `body.description ← record.description` | [events.insert], Request body > `Event.description` | provider reference | high | Optional provider body field. |
| `events.insert` | `body.location ← record.location` | [events.insert], Request body > `Event.location` | provider reference | high | Optional provider body field. |
| `events.insert` | `body.start ← record.start` | [events.insert], Request body > `Event.start` | provider reference | high | Required by the typed create action; CLI maps RFC3339 `start.dateTime`. |
| `events.insert` | `body.end ← record.end` | [events.insert], Request body > `Event.end` | provider reference | high | Required by the typed create action; CLI maps RFC3339 `end.dateTime`. |
| `events.insert` | `body.attendees ← record.attendees` | [events.insert], Request body > `Event.attendees` | provider reference | high | Optional provider body field. |
| `events.insert` | `body.recurrence ← record.recurrence` | [events.insert], Request body > `Event.recurrence` | provider reference | high | Optional provider body field. |
| `events.insert` | `body.colorId ← record.colorId` | [events.insert], Request body > `Event.colorId` | provider reference | high | Optional provider body field. |
| `events.insert` | `body.visibility ← record.visibility` | [events.insert], Request body > `Event.visibility` | provider reference | high | Optional provider body field. |
| `events.insert` | `body.transparency ← record.transparency` | [events.insert], Request body > `Event.transparency` | provider reference | high | Optional provider body field. |
| `events.move` | `path.calendarId ← record.calendar_id` | [events.move], Parameters > `calendarId` | provider reference | high | Provider-required path parameter. |
| `events.move` | `path.eventId ← record.event_id` | [events.move], Parameters > `eventId` | provider reference | high | Provider-required path parameter. |
| `events.move` | `query.destination ← record.destination` | [events.move], Parameters > `destination` | provider reference | high | Provider-required query parameter. |
| `events.patch` | `path.calendarId ← record.calendar_id` | [events.patch], Parameters > `calendarId` | provider reference | high | Provider-required path parameter. |
| `events.patch` | `path.eventId ← record.event_id` | [events.patch], Parameters > `eventId` | provider reference | high | Provider-required path parameter. |
| `events.patch` | `body.summary ← record.summary` | [events.patch], Request body > `Event.summary` | provider reference | high | Required typed patch field. |
| `events.patch` | `body.description ← record.description` | [events.patch], Request body > `Event.description` | provider reference | high | Optional provider body field. |
| `events.patch` | `body.location ← record.location` | [events.patch], Request body > `Event.location` | provider reference | high | Optional provider body field. |
| `events.patch` | `body.start ← record.start` | [events.patch], Request body > `Event.start` | provider reference | high | Required typed patch field; CLI maps RFC3339 `start.dateTime`. |
| `events.patch` | `body.end ← record.end` | [events.patch], Request body > `Event.end` | provider reference | high | Required typed patch field; CLI maps RFC3339 `end.dateTime`. |
| `events.patch` | `body.attendees ← record.attendees` | [events.patch], Request body > `Event.attendees` | provider reference | high | Optional provider body field. |
| `events.patch` | `body.recurrence ← record.recurrence` | [events.patch], Request body > `Event.recurrence` | provider reference | high | Optional provider body field. |
| `events.patch` | `body.colorId ← record.colorId` | [events.patch], Request body > `Event.colorId` | provider reference | high | Optional provider body field. |
| `events.patch` | `body.visibility ← record.visibility` | [events.patch], Request body > `Event.visibility` | provider reference | high | Optional provider body field. |
| `events.patch` | `body.transparency ← record.transparency` | [events.patch], Request body > `Event.transparency` | provider reference | high | Optional provider body field. |
| `events.quickAdd` | `path.calendarId ← record.calendar_id` | [events.quick-add], Parameters > `calendarId` | provider reference | high | Provider-required path parameter. |
| `events.quickAdd` | `query.text ← record.text` | [events.quick-add], Parameters > `text` | provider reference | high | Provider-required query parameter. |
| `events.update` | `path.calendarId ← record.calendar_id` | [events.update], Parameters > `calendarId` | provider reference | high | Provider-required path parameter. |
| `events.update` | `path.eventId ← record.event_id` | [events.update], Parameters > `eventId` | provider reference | high | Provider-required path parameter. |
| `events.update` | `body.summary ← record.summary` | [events.update], Request body > `Event.summary` | provider reference | high | Required by the typed replacement action. |
| `events.update` | `body.description ← record.description` | [events.update], Request body > `Event.description` | provider reference | high | Optional provider body field. |
| `events.update` | `body.location ← record.location` | [events.update], Request body > `Event.location` | provider reference | high | Optional provider body field. |
| `events.update` | `body.start ← record.start` | [events.update], Request body > `Event.start` | provider reference | high | Required by the typed replacement action; CLI maps RFC3339 `start.dateTime`. |
| `events.update` | `body.end ← record.end` | [events.update], Request body > `Event.end` | provider reference | high | Required by the typed replacement action; CLI maps RFC3339 `end.dateTime`. |
| `events.update` | `body.attendees ← record.attendees` | [events.update], Request body > `Event.attendees` | provider reference | high | Optional provider body field. |
| `events.update` | `body.recurrence ← record.recurrence` | [events.update], Request body > `Event.recurrence` | provider reference | high | Optional provider body field. |
| `events.update` | `body.colorId ← record.colorId` | [events.update], Request body > `Event.colorId` | provider reference | high | Optional provider body field. |
| `events.update` | `body.visibility ← record.visibility` | [events.update], Request body > `Event.visibility` | provider reference | high | Optional provider body field. |
| `events.update` | `body.transparency ← record.transparency` | [events.update], Request body > `Event.transparency` | provider reference | high | Optional provider body field. |
| `events.watch` | `path.calendarId ← record.calendar_id` | [events.watch], Parameters > `calendarId` | provider reference | high | Provider-required path parameter. |
| `events.watch` | `body.id ← record.id` | [events.watch], Request body > `Channel.id` | provider reference | high | Required typed channel identifier. |
| `events.watch` | `body.type ← record.type` | [events.watch], Request body > `Channel.type` | provider reference | high | Required typed channel transport (`web_hook`). |
| `events.watch` | `body.address ← record.address` | [events.watch], Request body > `Channel.address` | provider reference | high | Required typed webhook address. |
| `events.watch` | `body.token ← record.token` | [events.watch], Request body > `Channel.token` | provider reference | high | Optional provider channel token. |
| `settings.watch` | `body.id ← record.id` | [settings.watch], Request body > `Channel.id` | provider reference | high | Required typed channel identifier. |
| `settings.watch` | `body.type ← record.type` | [settings.watch], Request body > `Channel.type` | provider reference | high | Required typed channel transport (`web_hook`). |
| `settings.watch` | `body.address ← record.address` | [settings.watch], Request body > `Channel.address` | provider reference | high | Required typed webhook address. |
| `settings.watch` | `body.token ← record.token` | [settings.watch], Request body > `Channel.token` | provider reference | high | Optional provider channel token. |

### Per-mutation authoring determination

Each documented mutation is a record-shaped reverse-ETL request and is therefore executable through the current write executor. No operation needs the unavailable `rest_write` executor.

| Documented operation | Executable determination |
| --- | --- |
| `acl.delete` | `writes.json:delete_acl_rule` — typed delete record. |
| `acl.insert` | `writes.json:insert_acl_rule` — typed create record. |
| `acl.patch` | `writes.json:patch_acl_rule` — typed update record. |
| `acl.update` | `writes.json:update_acl_rule` — typed update record. |
| `acl.watch` | `writes.json:watch_acl` — typed channel-create record. |
| `calendarList.delete` | `writes.json:delete_calendar_list_entry` — typed delete record. |
| `calendarList.insert` | `writes.json:insert_calendar_list_entry` — typed create record. |
| `calendarList.patch` | `writes.json:patch_calendar_list_entry` — typed update record. |
| `calendarList.update` | `writes.json:update_calendar_list_entry` — typed update record. |
| `calendarList.watch` | `writes.json:watch_calendar_list` — typed channel-create record. |
| `calendars.clear` | `writes.json:clear_calendar` — typed custom record, destructive confirmation. |
| `calendars.delete` | `writes.json:delete_calendar` — typed delete record, destructive confirmation. |
| `calendars.insert` | `writes.json:insert_calendar` — typed create record. |
| `calendars.patch` | `writes.json:patch_calendar` — typed update record. |
| `calendars.transferOwnership` | `writes.json:transfer_calendar_ownership` — typed custom record, destructive confirmation. |
| `calendars.update` | `writes.json:update_calendar` — typed update record. |
| `channels.stop` | `writes.json:stop_channel` — typed custom record. |
| `events.delete` | `writes.json:delete_event` — typed delete record, destructive confirmation. |
| `events.import` | `writes.json:import_event` — typed create record. |
| `events.insert` | `writes.json:insert_event` — typed create record. |
| `events.move` | `writes.json:move_event` — typed custom record. |
| `events.patch` | `writes.json:patch_event` — typed update record. |
| `events.quickAdd` | `writes.json:quick_add_event` — typed custom record. |
| `events.update` | `writes.json:update_event` — typed update record. |
| `events.watch` | `writes.json:watch_events` — typed channel-create record. |
| `settings.watch` | `writes.json:watch_settings` — typed channel-create record. |

Coverage: **149/149 declared request-field uses have a primary-provider citation**: 27 pre-existing read/direct-read uses plus the 122 write-action uses above. No field is deferred under tier 5. The 11 stream endpoints, `freeBusy.query`, and all 26 write actions have explicit provider references, for **38/38 operation-level source coverage**.

[acl.delete]: https://developers.google.com/workspace/calendar/api/v3/reference/acl/delete
[acl.insert]: https://developers.google.com/workspace/calendar/api/v3/reference/acl/insert
[acl.patch]: https://developers.google.com/workspace/calendar/api/v3/reference/acl/patch
[acl.update]: https://developers.google.com/workspace/calendar/api/v3/reference/acl/update
[acl.watch]: https://developers.google.com/workspace/calendar/api/v3/reference/acl/watch
[calendar-list.delete]: https://developers.google.com/workspace/calendar/api/v3/reference/calendarList/delete
[calendar-list.insert]: https://developers.google.com/workspace/calendar/api/v3/reference/calendarList/insert
[calendar-list.patch]: https://developers.google.com/workspace/calendar/api/v3/reference/calendarList/patch
[calendar-list.update]: https://developers.google.com/workspace/calendar/api/v3/reference/calendarList/update
[calendar-list.watch]: https://developers.google.com/workspace/calendar/api/v3/reference/calendarList/watch
[calendars.clear]: https://developers.google.com/workspace/calendar/api/v3/reference/calendars/clear
[calendars.delete]: https://developers.google.com/workspace/calendar/api/v3/reference/calendars/delete
[calendars.insert]: https://developers.google.com/workspace/calendar/api/v3/reference/calendars/insert
[calendars.patch]: https://developers.google.com/workspace/calendar/api/v3/reference/calendars/patch
[calendars.transfer-ownership]: https://developers.google.com/workspace/calendar/api/v3/reference/calendars/transferOwnership
[calendars.update]: https://developers.google.com/workspace/calendar/api/v3/reference/calendars/update
[channels.stop]: https://developers.google.com/workspace/calendar/api/v3/reference/channels/stop
[events.delete]: https://developers.google.com/workspace/calendar/api/v3/reference/events/delete
[events.import]: https://developers.google.com/workspace/calendar/api/v3/reference/events/import
[events.insert]: https://developers.google.com/workspace/calendar/api/v3/reference/events/insert
[events.move]: https://developers.google.com/workspace/calendar/api/v3/reference/events/move
[events.patch]: https://developers.google.com/workspace/calendar/api/v3/reference/events/patch
[events.quick-add]: https://developers.google.com/workspace/calendar/api/v3/reference/events/quickAdd
[events.update]: https://developers.google.com/workspace/calendar/api/v3/reference/events/update
[events.watch]: https://developers.google.com/workspace/calendar/api/v3/reference/events/watch
[settings.watch]: https://developers.google.com/workspace/calendar/api/v3/reference/settings/watch
