# Overview

Reads Zoom users, meetings, webinars, and bounded module-specific data through the Zoom REST API.
The current direct-read surface includes Quality of Service (QoS), AI Companion, My Notes, and
Healthcare clinical-note routes; the Healthcare completion-status update remains approval-gated.

The provider-owned inventory contains 1,913 callable REST operations from Zoom's OpenAPI 3.1.1
reference corpus (881 reads and 1,032 writes), retrieved on 2026-08-05 from the docs static build
`2026-08-03T14-58-19-06-00`, and is being brought to full documented-operation parity one Zoom
provider module at a time (see issue #3915 for the module-by-module tracker). Wave 1 exposes the
three existing stream-backed reads: `pm zoom users list`, `pm zoom meetings list`, and
`pm zoom webinars list`. Later waves add bounded direct-read/write commands module by module; see
"Direct reads" below and "Executable today" in Known limits for the exact current set.

One high-risk Zoom write action is implemented for the Healthcare module. All remaining provider
operations stay explicitly disposed in `api_surface.json`; the ledger is not a claim that those
operations are executable.

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

### Direct reads (qss module)

Three bounded, single-request `direct_read` commands read the `qss` (Quality of Service
Subscription) provider module. Each takes one required typed path-parameter flag and returns the
decoded JSON response under the `json_redacted` output policy; unlike the ETL streams above these
are not paginated by the connector — call again with a different id, there is no `--limit` or
cursor auto-continuation. Only account/subscriptions with the QSS add-on and a subscribed QSS
summary event receive non-empty data; see
[Zoom's QSS product page](https://explore.zoom.us/en/qss/).

- `pm zoom qss meeting-participants list --meeting-id <id>` reads GET
  `/v2/metrics/meetings/{meetingId}/participants/qos_summary` (operation
  `zoom.list_meeting_participants_qos_summary`). Provider reference:
  https://developers.zoom.us/docs/api/qss.md.
- `pm zoom qss webinar-participants list --webinar-id <id>` reads GET
  `/v2/metrics/webinars/{webinarId}/participants/qos_summary` (operation
  `zoom.list_webinar_participants_qos_summary`). Provider reference:
  https://developers.zoom.us/docs/api/qss.md.
- `pm zoom qss session-users list --session-id <id>` reads GET
  `/v2/videosdk/sessions/{sessionId}/users/qos_summary` (operation
  `zoom.list_session_users_qos_summary`). Provider reference:
  https://developers.zoom.us/docs/api/qss.md.

All three responses carry a `next_page_token` field in Zoom's own schema (used to fetch a further
page of participants/users, not exposed as a command flag in this slice — see Known limits). The
shared `json_redacted` output policy redacts any field whose name contains `token`
(`engine/direct_read.go`'s `shouldRedactJSONField`), so `next_page_token` returns as
`next_page_token_redacted: true` rather than its wire value even though it is a pagination cursor,
not a credential.

### Direct reads (ai-companion module)

- `pm zoom ai-companion conversation-archive get --user-id <id>` reads GET
  `/v2/aic/users/{userId}/conversation_archive` (operation
  `zoom.get_ai_companion_conversation_archives`). Provider reference:
  https://developers.zoom.us/docs/api/ai-companion.md. The response's `aic_history_download_url`
  and `physical_files[].download_url` fields are redacted by `json_redacted` (both fields' names
  contain `download`+`url`, matching `shouldRedactJSONField`'s download-URL rule) since they are
  direct download links to a user's AI conversation history/attachments.

### Direct reads (my-notes module)

- `pm zoom my-notes list` reads GET `/v2/my_notes/notes` (operation `zoom.list_my_notes`). No
  request parameters; the live artifact documents none. Provider reference:
  https://developers.zoom.us/docs/api/my-notes.md.
- `pm zoom my-notes content get --note-id <id> [--include transcript]` reads GET
  `/v2/my_notes/notes/{noteId}/content` (operation `zoom.get_my_notes_content`). Unlike qss's
  response-only pagination fields, the `include=transcript` query parameter here is explicitly
  documented in the provider's operation description prose, so it is exposed as an optional
  `--include` flag. Provider reference: https://developers.zoom.us/docs/api/my-notes.md.

### Direct reads (healthcare module)

The Healthcare routes return clinical note content and clinical identifiers, so both commands use
the existing `clinical_json_redacted` policy. It redacts note content, EHR appointment/patient/
provider identifiers, note owner/modifier identifiers, and token-shaped response fields before
output. Sanitized fixtures exercise the returned structure; no provider clinical example is retained
in this repository.

- `pm zoom healthcare clinical-notes list [--note-owner-user-id <id>] [--meeting-id <id>]` reads
  GET `/v2/clinical_notes/notes` (operation `zoom.list_clinical_notes`). The two optional filters
  are explicitly described by Zoom's operation prose. `from`, `to`, `page_size`, and
  `next_page_token` appear only in the `200` response schema and are not request flags. Provider
  reference: https://developers.zoom.us/docs/api/healthcare.md.
- `pm zoom healthcare clinical-notes get --note-id <id>` reads GET
  `/v2/clinical_notes/notes/{noteId}` (operation `zoom.get_clinical_note`). Provider reference:
  https://developers.zoom.us/docs/api/healthcare.md.

## Write actions & risks

Read behavior includes external Zoom API reads of user, meeting, webinar, QoS, AI Companion, My
Notes, and healthcare clinical-note data.

- `pm zoom healthcare clinical-notes update --note-id <id> --is-note-completed <true|false>`
  plans the typed PATCH `/v2/clinical_notes/notes/{noteId}` action
  (`update_clinical_note`). It changes a clinical-note completion status and therefore must use the
  existing plan → preview → explicit approval → execute path. The note ID is redacted in write
  errors. Zoom's `204 No Content` success response is recorded as a successful action.

The provider inventory records 1,032 documented writes. Only the Healthcare completion-status
action is currently declared; all remaining write rows are either blocked on connector-local typed
contracts, safety/approval evidence, and fixtures, or on the corresponding Zoom account entitlement.
Future writes must use the existing plan → preview → explicit approval → execute path; destructive
operations additionally require the typed confirmation gate.

## Known limits

- Batch default: `read_page_size=100`.
- Provider inventory: 1,913 operations across 35 published modules (881 reads, 1,032 writes). See
  issue #3915 for the full module-by-module tracking table.
- Executable today: 12 operations — 3 stream-backed GET reads (`users`, `meetings`, `webinars`), 3
  bounded `qss` module direct reads, 1 bounded `ai-companion` module direct read, 2 bounded
  `my-notes` module direct reads, 2 Healthcare direct reads, and 1 approval-gated Healthcare PATCH
  action.
- Direct-read commands in this connector take only the request inputs the live provider artifact
  expressly documents. A response-body field of the same name (including Healthcare's `from`, `to`,
  `page_size`, and `next_page_token`) is not sufficient evidence of an accepted request parameter.
  This is a deliberate module-by-module scope-narrowing, not an oversight.
- Pending connector-local delivery: 1,830 operations have no shared foundation blocker, but still
  need bounded Zoom-specific contracts, schemas, safety evidence, and fixtures before they can
  become commands.
- Provider-side restrictions: 17 operations (five Information Barriers, seven Chat migration, one
  Meeting audit trail, and four Phone blocked-list routes) remain blocked until Zoom enables the
  corresponding product/account capability.
- Justified exclusions: 54 provider-deprecated operations remain recorded as `deprecated` ledger
  evidence and are not implemented.
