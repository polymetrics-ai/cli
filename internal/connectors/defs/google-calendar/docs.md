# Google Calendar connector

## Overview

This connector exposes every documented Google Calendar API v3 operation through the declarative runtime: 11 GET streams, one bounded typed `freebusy query` direct read, and 26 typed reverse-ETL write actions. It does not expose generic HTTP access.

The official Google Discovery document reviewed on 2026-08-05 (revision `20260731`) lists 38 operations: 11 GET, 15 POST, 4 PATCH, 4 PUT, and 4 DELETE. All 38 are reachable. The 26 mutation operations use the working `writes.json` record executor; none relies on the unavailable `rest_write` executor and no operation remains `planned` or blocked.

## Auth setup

Use OAuth2 refresh-token credentials from environment variables, stdin, or the configured credential store. Do not put credential values in command arguments, prompts, logs, or documentation.

Required secret keys are `client_id`, `client_refresh_token_2`, and `client_secret`. The `calendarid` configuration value scopes calendar streams; use `primary` for the authenticated user's primary calendar.

Set `mode=fixture` for credential-free deterministic replay of all 11 bundled read streams. Fixture mode performs no external requests and rejects reads without a bundled fixture; write validation, previews, execution, direct reads, and binary downloads fail closed.

## Streams notes

The stream-backed GET operations are:

- `calendar_list` and `calendar_list_entry`
- `calendar`
- `colors`
- `events`, `event`, and `event_instances`
- `settings` and `setting`
- `acl` and `acl_rule`

The `events` stream is incremental on `updated`. An explicitly configured `start_date` is sent as the `updatedMin` first-run lower bound; when it is omitted, a fresh read is unfiltered. Cursor-paginated streams, including `settings`, use the provider's `nextPageToken` response field.

`freebusy query` accepts one calendar ID and RFC3339 `timeMin`/`timeMax` bounds, validates that the lower bound precedes the upper bound, caps the response, and returns JSON-redacted output.

## Write actions & risks

All 26 documented mutations are executable typed reverse-ETL actions. Every action has a record schema and replay fixture under `fixtures/writes/`, and every write command follows the shared plan → preview → explicit approval → execute flow.

- ACL: delete, insert, patch, update, and watch.
- Calendar list: delete, insert, patch, update, and watch.
- Calendars: clear, delete, insert, patch, transfer ownership, and update.
- Channels: stop.
- Events: delete, import, insert, move, patch, quick-add, update, and watch.
- Settings: watch.

Deletes, calendar clearing, and ownership transfer require `--confirm destructive` in addition to the normal reverse-ETL approval. Notification-channel lifecycle actions, calendar clear/delete, ownership transfer, event move, and quick-add are non-batchable. Provider mutation request fields are typed and source-cited in the phase research ledger.

Event imports require the provider's RFC 5545 `iCalUID` in addition to start and end times. Watch actions accept HTTPS callback addresses only.

## Known limits

- No live Google Calendar provider certification was performed; reads and every write request shape are fixture-backed.
- `cdc=false`: watch creation and stopping are supported as explicit write actions, but webhook delivery and changefeed state are outside this connector.
- The direct free/busy operation is intentionally typed and bounded; it is not a generic Google Calendar HTTP facility.
- Source audit: [Google Calendar API v3 Discovery](https://www.googleapis.com/discovery/v1/apis/calendar/v3/rest) and [Google Calendar API v3 reference](https://developers.google.com/workspace/calendar/api/v3/reference). The phase research ledger records a primary-provider citation for all 149 declared request-field uses and all 38 operations.
