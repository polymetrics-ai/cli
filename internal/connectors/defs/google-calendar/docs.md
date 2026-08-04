# Google Calendar connector

## Overview

This connector reads Google Calendar API v3 data through 11 declarative GET streams and offers one bounded, typed direct read: `freebusy query`. It does not expose a generic HTTP command.

The official Google Discovery document reviewed on 2026-08-05 (revision `20260731`) lists 38 operations. Twelve are currently reachable: 11 GET operations through streams and `freeBusy.query` through the typed direct-read command. The other 26 documented mutation operations are recorded in `api_surface.json` as blocked; they require `rest_write`, whose schema exists but whose command runner has no execution dispatch.

## Auth setup

Use OAuth2 refresh-token credentials from environment variables, stdin, or the configured credential store. Do not put credential values in command arguments, prompts, logs, or documentation.

Required secret keys are `client_id`, `client_refresh_token_2`, and `client_secret`. The `calendarid` configuration value scopes calendar streams; use `primary` for the authenticated user's primary calendar.

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

No Google Calendar mutation is executable from this connector. All 26 documented non-read operations are explicit blocked ledger rows rather than advertised reverse-ETL commands:

- ACL, calendar-list, calendar, event, channel, and settings mutations all require the missing `rest_write` command-runner executor.
- This includes destructive deletes and clears, ownership transfer, notification-channel management, and event/calendar updates.

The connector therefore has no write plan, preview, approval, or execute surface to invoke. This is a runtime limitation, not a claim that the provider operations are absent.

## Known limits

- No live Google Calendar provider certification was performed; validation is fixture-backed.
- `cdc=false`: Calendar watch operations are ledger-blocked, and webhook delivery/changefeed state is outside this connector.
- The direct free/busy operation is intentionally typed and bounded; it is not a generic Google Calendar HTTP facility.
- Source audit: [Google Calendar API v3 Discovery](https://www.googleapis.com/discovery/v1/apis/calendar/v3/rest) and [Google Calendar API v3 reference](https://developers.google.com/workspace/calendar/api/v3/reference). The phase research ledger records a provider citation for every declared request-field use.
