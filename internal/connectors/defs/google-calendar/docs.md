# Google Calendar connector

## Overview

Google Calendar API v3 connector for calendars, calendar-list entries, events, recurring-event instances, colors, settings, ACL rules, notification-channel setup/stop, and free/busy queries.

- **Read/check:** fixture-backed declarative HTTP reads for `calendar_list`, `calendar_list_entry`, `calendar`, `colors`, `events`, `event`, `event_instances`, `settings`, `setting`, `acl`, and `acl_rule`.
- **Direct read:** `freebusy query` is a bounded typed POST to `/calendar/v3/freeBusy` with JSON-redacted output. It is not a generic HTTP request facility.
- **Reverse ETL:** all Calendar API mutations are represented as named typed actions.
- **CDC:** `cdc=false`. Calendar API watch/channel setup operations are exposed as typed reverse-ETL management actions, but webhook delivery, channel renewal, and changefeed state consumption require shared runtime infrastructure outside this connector bundle.

## Auth setup

Use OAuth2 refresh-token credentials from environment variables, stdin, or the configured credential store. Do not paste credentials into prompts. The connector refreshes an access bearer value through Google OAuth. Fixture auth bypass is harness-internal only and is not a user configuration mode.

Required credential keys are `client_id`, `client_refresh_token_2`, and `client_secret`. The `calendarid` config value defaults to `primary` for read streams.

## Streams notes

The operation ledger covers 11 GET operations with 11 executable streams:

- `calendar_list` and `calendar_list_entry`
- `calendar`
- `colors`
- `events`, `event`, and `event_instances`
- `settings` and `setting`
- `acl` and `acl_rule`

The `events` stream is incremental on `updated`, sent as `updatedMin` using the configured `start_date` lower bound on first run.

## Write actions & risks

Named reverse-ETL actions: `delete_acl_rule`, `create_acl_rule`, `patch_acl_rule`, `update_acl_rule`, `watch_acl`, `delete_calendar_list_entry`, `create_calendar_list_entry`, `patch_calendar_list_entry`, `update_calendar_list_entry`, `watch_calendar_list`, `clear_calendar`, `delete_calendar`, `create_calendar`, `patch_calendar`, `transfer_calendar_ownership`, `update_calendar`, `stop_channel`, `delete_event`, `import_event`, `create_event`, `move_event`, `patch_event`, `quick_add_event`, `update_event`, `watch_events`, and `watch_settings`.

All write actions use closed record schemas. Destructive actions such as event deletion/cancellation, calendar deletion/clear, ACL removal, calendar-list removal, event move, ownership transfer, and channel stop require plan -> preview -> explicit approval -> execute plus destructive confirmation. Target identifiers and webhook/channel details are redacted in previews and errors where declared.

## Known limits

- No live Google Calendar provider certification was performed in this wave; validation is fixture-only.
- `cdc=false`: watch actions create Calendar API notification channels, but webhook delivery/changefeed state consumption is a separate runtime concern.
- The `freebusy query` direct read is bounded and typed; the connector intentionally does not expose a generic Google Calendar HTTP operation command.

Official-source audit: Google Calendar API v3 discovery (`https://www.googleapis.com/discovery/v1/apis/calendar/v3/rest`) and official reference (`https://developers.google.com/workspace/calendar/api/v3/reference`) were reviewed on 2026-07-31. The ledger records 38 operations exactly once: 11 GET, 15 POST, 4 PATCH, 4 PUT, and 4 DELETE.
