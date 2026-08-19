# Overview

Reads Zoom users, meetings, webinars, and 70 bounded module-specific direct-read routes through the Zoom REST API.

The provider-owned inventory contains 1,913 callable REST operations from Zoom's OpenAPI 3.1.1
reference corpus (881 reads and 1,032 writes), retrieved on 2026-08-05 from the docs static build
`2026-08-03T14-58-19-06-00`. The existing stream-backed reads remain `pm zoom users list`,
`pm zoom meetings list`, and `pm zoom webinars list`; Wave 2 adds the reviewed direct-read cohort.

No Zoom write action is implemented in this slice. The remaining provider operations stay explicitly
disposed in `api_surface.json`; the ledger is not a claim that those operations are executable.

Service API documentation: https://developers.zoom.us/docs/api/.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); Zoom OAuth access token, sent as a Bearer token
  (Authorization: Bearer `<access_token>`). Never logged.
- `base_url` (optional, string); default `https://api.zoom.us/v2`; format `uri`; Zoom API base URL
  override for tests or proxies.
- `max_pages` (optional, string); default `0`. The field remains in the connection specification,
  but the current Zoom cursor paginator does not consume it.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; records per page (1-300); sent as the `page_size`
  query parameter.
- `user_id` (optional, string); Zoom user ID or email that scopes the `meetings` and `webinars`
  streams. Set it in credential configuration or override it with `--user-id` for either command.

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `base_url=https://api.zoom.us/v2`, `max_pages=0`, `page_size=100`.
`max_pages` is materialized as configuration but does not affect Zoom pagination.

Authentication behavior:

- Bearer token authentication using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/users` with query `page_size`=`1`.

## Streams notes

Default pagination is cursor pagination: send `next_page_token` and take the next token from
`next_page_token`. The stream command `--limit` bounds emitted records (default 100).

- `users`: `pm zoom users list` reads GET `/v2/users` through stream `users`. Records are at
  `users`; query `page_size`=`{{ config.page_size }}`; computed output fields are `name` and
  `updated_at`; emits passthrough records. Provider reference:
  https://developers.zoom.us/docs/api/users.md.
- `meetings`: `pm zoom meetings list` reads GET `/v2/users/{userId}/meetings` through stream
  `meetings`. Supply `--user-id` or configure `user_id` in the credential. Records are at
  `meetings`; query `page_size`=`{{ config.page_size }}`; computed output fields are `id`, `name`,
  and `updated_at`; emits passthrough records. Provider reference:
  https://developers.zoom.us/docs/api/meetings.md.
- `webinars`: `pm zoom webinars list` reads GET `/v2/users/{userId}/webinars` through stream
  `webinars`. Supply `--user-id` or configure `user_id` in the credential. Records are at
  `webinars`; query `page_size`=`{{ config.page_size }}`; computed output fields are `id`, `name`,
  and `updated_at`; emits passthrough records. Provider reference:
  https://developers.zoom.us/docs/api/meetings.md.

## Direct reads

The 70 `direct_read` commands are bounded to one declared provider request and use the
`json_redacted` output policy. Their command groups are `qss`, `ai-companion`, `my-notes`,
`healthcare`, `quality-management`, `cobrowse-sdk`, `scim2`, `virtual-agent`, `auto-dialer`,
`tasks`, `workforce-management`, `clips`, and `crc`; run `pm zoom --help` or a group help command
to inspect the exact required path flags and provider citations.

All 70 commands pass connector preflight and 52 sanitized response fixtures execute through the
real operation runner. This is fixture proof, not a live-certification claim: Zoom remains outside
the centrally owned certification scope and no credential is resolved by these checks.

Zoom's SCIM2 provider routes use the bare `https://api.zoom.us` origin while the ordinary API base
defaults to `https://api.zoom.us/v2`. Until separately approved operation-specific base-origin
support exists, invoke a SCIM2 direct-read command with `--config base_url=https://api.zoom.us`.
That setting is command-local; do not reuse it for the existing stream commands.

## Write actions & risks

This Wave 2 connector surface remains read-only. Read behavior includes external Zoom API reads of
user, meeting, webinar, Quality of Service, AI Companion, My Notes, Healthcare, Quality
Management, Cobrowse SDK, SCIM2, Virtual Agent, Auto Dialer, Tasks, Workforce Management, Clips,
and Conference Room Connector data.

The provider inventory records 1,032 documented writes, but none is a declared Zoom write action.
The 997 write operations classified as implementable now remain blocked on connector-local typed
request contracts, safety/approval evidence, and fixtures; that is not a shared runtime blocker.
The 12 provider-restricted writes remain unavailable pending the corresponding Zoom account
entitlement. Future writes must use the existing plan → preview → explicit approval → execute path;
destructive operations additionally require the typed confirmation gate.

## Known limits

- Batch default: `read_page_size=100`.
- Provider inventory: 1,913 operations across 35 published modules (881 reads, 1,032 writes).
- Executable today: 3 stream-backed GET operations plus 70 bounded direct-read GET operations.
- Pending connector-local delivery: 1,769 operations remain ledger-blocked and need bounded
  Zoom-specific contracts, schemas, safety evidence, and fixtures before they can become commands.
- Provider-side restrictions: 17 operations (five Information Barriers, seven Chat migration, one
  Meeting audit trail, and four Phone blocked-list routes) remain blocked until Zoom enables the
  corresponding product/account capability.
- Justified exclusions: 54 provider-deprecated operations remain recorded as `deprecated` ledger
  evidence and are not implemented.
