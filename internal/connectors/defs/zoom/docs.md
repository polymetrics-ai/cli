# Overview

Reads Zoom users, meetings, webinars, and bounded module-specific data through the Zoom REST API.
The current direct-read surface includes Quality of Service (QoS), AI Companion, My Notes,
Healthcare clinical-note, Quality Management, Cobrowse SDK, and SCIM2 Group/User routes; every
declared write action, including Chatbot and SCIM2 actions, remains approval-gated.

The provider-owned inventory contains 1,913 callable REST operations from Zoom's OpenAPI 3.1.1
reference corpus (881 reads and 1,032 writes), retrieved on 2026-08-05 from the docs static build
`2026-08-03T14-58-19-06-00`, and is being brought to full documented-operation parity one Zoom
provider module at a time (see issue #3915 for the module-by-module tracker). Wave 1 exposes the
three existing stream-backed reads: `pm zoom users list`, `pm zoom meetings list`, and
`pm zoom webinars list`. Later waves add bounded direct-read/write commands module by module; see
"Direct reads" below and "Executable today" in Known limits for the exact current set.

High-risk Zoom write actions are implemented for Chatbot, SCIM2, Healthcare, Quality Management,
and Customer Managed Keys Hybrid. All remaining provider operations stay explicitly disposed in
`api_surface.json`; the ledger is not a claim that those operations are executable.

Service API documentation: https://developers.zoom.us/docs/api/.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); Zoom OAuth access token, sent as a Bearer token
  (Authorization: Bearer `<access_token>`). Never logged.
- `chatbot_client_id` (required for Chatbot commands, secret, string); client ID used only for the
  declared Chatbot client-credentials exchange. Never logged.
- `chatbot_client_secret` (required for Chatbot commands, secret, string); client secret used only
  for the declared Chatbot client-credentials exchange. Never logged.
- `base_url` (optional, string); default `https://api.zoom.us/v2`; format `uri`; Zoom API base URL
  override for tests or proxies.
- `scim2_base_url` (optional, string); default `https://api.zoom.us`; format `uri`; Zoom SCIM2 API
  root override for tests or proxies. It is used only by the declared `/scim2` operations and is
  paired with their declared Bearer authentication.
- `chatbot_token_url` (optional, string); default `https://api.zoom.us/oauth/token`; format `uri`;
  Chatbot-only client-credentials token endpoint.
- `max_pages` (optional, string); default `0`. The field remains in the connection specification,
  but the current Zoom cursor paginator does not consume it.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; records per page (1-300); sent as the `page_size`
  query parameter.
- `user_id` (optional, string); Zoom user ID or email that scopes the `meetings` and `webinars`
  streams. Set it in credential configuration or override it with `--user-id` for either command.

Secret fields are redacted in logs and write previews: `access_token`, `chatbot_client_id`,
`chatbot_client_secret`, and `key_connector_jwt`.

Default configuration values: `base_url=https://api.zoom.us/v2`,
`scim2_base_url=https://api.zoom.us`, `chatbot_token_url=https://api.zoom.us/oauth/token`,
`max_pages=0`, `page_size=100`. `max_pages` is materialized as configuration but does not affect
Zoom pagination.

Authentication behavior:

- Bearer token authentication using `secrets.access_token`.
- Chatbot actions use their own declared client-credentials exchange: the client ID and secret are
  sent in an HTTP Basic authorization header to `chatbot_token_url`, then the returned access token
  is used as the action's Bearer credential. They do not reuse `access_token`.
- SCIM2 Group and User operations use the declared `scim2_base_url` root and the ordinary
  `access_token` Bearer secret. They do not inherit an ordinary `/v2` request path or any unrelated
  operation header.

Ordinary API requests use the configured `base_url` value after applying defaults; SCIM2 requests
use their operation-scoped `scim2_base_url` declaration.

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

### Direct reads (quality-management module)

Quality Management responses can carry agent/consumer names, contact values, account/user
identifiers, and token-shaped pagination fields. The five bounded reads use `json_redacted` plus a
connector-local sensitive field policy, so those values are removed before CLI output. Fixtures are
synthetic and do not retain Zoom response examples.

- `pm zoom quality-management automated-evaluations list` reads GET
  `/v2/qm/automated_evaluations` (operation
  `zoom.list_quality_management_automated_evaluations`). The provider artifact declares no request
  parameters; `page_size` and `next_page_token` are response fields, not flags.
- `pm zoom quality-management evaluations list` reads GET `/v2/qm/evaluation` (operation
  `zoom.list_quality_management_evaluations`). The provider artifact declares no request parameters.
- `pm zoom quality-management evaluations get --evaluation-id <id>` reads GET
  `/v2/qm/evaluation/{evaluationId}` (operation `zoom.get_quality_management_evaluation`).
- `pm zoom quality-management interactions list` reads GET `/v2/qm/interactions` (operation
  `zoom.list_quality_management_interactions`). The provider artifact declares no request
  parameters; response-only date/pagination fields are not flags.
- `pm zoom quality-management interactions get --interaction-id <id>` reads GET
  `/v2/qm/interactions/{interactionId}` (operation `zoom.get_quality_management_interaction`).
  Provider reference for all five routes: https://developers.zoom.us/docs/api/quality-management.md.

### Direct reads (cobrowse-sdk module)

Cobrowse SDK responses can contain join-capable session pins, session/user identifiers, display
names, connection IDs, IP addresses, and token-shaped pagination fields. The four bounded reads use
`json_redacted` plus a connector-local sensitive field policy, so those values are removed before
CLI output. Fixtures are synthetic and do not retain Zoom response examples.

- `pm zoom cobrowse-sdk live-sessions list [--from <YYYY-MM-DD>] [--to <YYYY-MM-DD>]` reads GET
  `/v2/cobrowsesdk/live_sessions` (operation `zoom.list_cobrowse_live_sessions`). Zoom explicitly
  permits the optional monthly date range in the operation prose; the range must fall in the past
  six months. `page_size` and `next_page_token` are response fields, not flags.
- `pm zoom cobrowse-sdk past-sessions list [--from <YYYY-MM-DD>] [--to <YYYY-MM-DD>]` reads GET
  `/v2/cobrowsesdk/past_sessions` (operation `zoom.list_cobrowse_past_sessions`) with the same
  documented monthly range semantics. Response `from`, `to`, `page_size`, and `next_page_token`
  do not create additional inputs beyond the two explicitly documented query dates.
- `pm zoom cobrowse-sdk sessions get --session-id <id>` reads GET
  `/v2/cobrowsesdk/sessions/{sessionId}` (operation `zoom.get_cobrowse_session`).
- `pm zoom cobrowse-sdk sessions users list --session-id <id>` reads GET
  `/v2/cobrowsesdk/sessions/{sessionId}/users` (operation `zoom.list_cobrowse_session_users`).
  Provider reference for all four routes: https://developers.zoom.us/docs/api/cobrowse-sdk.md.

### Direct reads (SCIM2 module)

The SCIM2 artifact was re-fetched from
https://developers.zoom.us/docs/api/scim2.md on 2026-08-08T13:33:09Z (171,559 bytes; SHA-256
`ba86462a888677ea38a8bcc0557e9c4cf5809cd78fc6bc7655f85f79e5b27264`). Its four Group/User reads
use the provider's root `https://api.zoom.us` server rather than the ordinary `/v2` base. SCIM
user/group identifiers, names, contact and organization data, aliases, memberships, and extension
objects are redacted from `json_redacted` output. The artifact does not declare standalone paging
inputs for these commands, so none is exposed.

- `pm zoom scim2 groups list` reads GET `/scim2/Groups` (operation
  `zoom.list_scim2_groups`).
- `pm zoom scim2 groups get --group-id <id>` reads GET `/scim2/Groups/{groupId}` (operation
  `zoom.get_scim2_group`).
- `pm zoom scim2 users list` reads GET `/scim2/Users` (operation
  `zoom.list_scim2_users`).
- `pm zoom scim2 users get --user-id <id>` reads GET `/scim2/Users/{userId}` (operation
  `zoom.get_scim2_user`).

Provider reference for all four routes: https://developers.zoom.us/docs/api/scim2.md.

## Write actions & risks

Read behavior includes external Zoom API reads of user, meeting, webinar, QoS, AI Companion, My
Notes, healthcare clinical-note, Quality Management, Cobrowse SDK session, and SCIM2 user/group
data.

- `pm zoom healthcare clinical-notes update --note-id <id> --is-note-completed <true|false>`
  plans the typed PATCH `/v2/clinical_notes/notes/{noteId}` action
  (`update_clinical_note`). It changes a clinical-note completion status and therefore must use the
  existing plan → preview → explicit approval → execute path. The note ID is redacted in write
  errors. Zoom's `204 No Content` success response is recorded as a successful action.

- `pm zoom quality-management interactions create --download-url <url>` plans the typed POST
  `/v2/qm/interactions` action (`create_quality_management_interaction`). It imports a third-party
  interaction into Quality Management and must use plan → preview → explicit approval → execute.
  The command exposes each documented optional scalar request field, including nested
  `interaction_info` fields; if an interaction-info field is supplied, `--interaction-channel-type`
  is required by Zoom. Download URLs and interaction-info fields are redacted in generic write
  errors. Fixture execution proves the documented `201 Created` response succeeds.

- `pm zoom chatbot messages send --account-id <id> --content <json-object> --robot-jid <jid>
  --to-jid <jid> --user-jid <jid>` plans the typed POST `/v2/im/chat/messages` action
  (`send_chatbot_message`). `content` accepts exactly one provider-defined JSON object, rather than
  a raw arbitrary body. Account, message, and JID values are redacted.

- `pm zoom chatbot messages edit --message-id <id> --account-id <id> --content <json-object>
  --robot-jid <jid>` plans the typed PUT `/v2/im/chat/messages/{message_id}` action
  (`edit_chatbot_message`). It has the same Chatbot-only client-credentials transport and
  redaction policy.

- `pm zoom chatbot messages delete --message-id <id>` plans the typed DELETE
  `/v2/im/chat/messages/{message_id}` action (`delete_chatbot_message`). It requires the normal
  plan/preview/approval flow plus destructive typed confirmation before execute.

- `pm zoom chatbot link-unfurls create --user-id <id> --trigger-id <id> --content <string>` plans
  the typed POST `/v2/im/chat/users/{userId}/unfurls/{triggerId}` action
  (`create_chatbot_link_unfurl`). Zoom's `204 No Content` is recorded as a successful status-only
  action; no response body is invented.

- `pm zoom scim2 groups create --resource <json-object>` plans the typed POST `/scim2/Groups`
  action (`create_scim2_group`). `--resource` accepts exactly one documented SCIM Group resource at
  that fixed endpoint; aliases, display names, memberships, and extension fields are redacted.
- `pm zoom scim2 groups update --group-id <id> --patch <json-object>` plans the typed PATCH
  `/scim2/Groups/{groupId}` action (`update_scim2_group`). `--patch` accepts exactly one documented
  SCIM PatchOp object; the documented `204 No Content` response is status-only.
- `pm zoom scim2 groups delete --group-id <id>` plans the typed DELETE
  `/scim2/Groups/{groupId}` action (`delete_scim2_group`). It requires destructive typed
  confirmation and records the documented `204 No Content` status without inventing a body.
- `pm zoom scim2 users create --resource <json-object>` and `pm zoom scim2 users update --user-id
  <id> --resource <json-object>` plan the typed POST `/scim2/Users` and PUT
  `/scim2/Users/{userId}` actions (`create_scim2_user`, `update_scim2_user`). Each named resource
  flag accepts one documented extensible SCIM User object only for its fixed endpoint; profile,
  contact, organization, and extension fields are redacted.
- `pm zoom scim2 users deactivate --user-id <id> --patch <json-object>` plans the typed PATCH
  `/scim2/Users/{userId}` action (`deactivate_scim2_user`) with one documented activation-state
  SCIM PatchOp object.
- `pm zoom scim2 users delete --user-id <id>` plans the typed DELETE
  `/scim2/Users/{userId}` action (`delete_scim2_user`). It requires destructive typed confirmation
  and records the documented `204 No Content` status without inventing a body.

The provider inventory records 1,032 documented writes. The four Chatbot actions above, seven
SCIM2 actions, the Healthcare completion-status action, the Quality Management interaction-creation
action, and the Customer Managed Keys Hybrid archival-key action are currently declared; all
remaining write rows are either blocked on connector-local typed contracts, safety/approval
evidence, and fixtures, or on the corresponding Zoom account entitlement. Future writes must use
the existing plan → preview → explicit approval → execute path; destructive operations additionally
require the typed confirmation gate.

## Known limits

- Batch default: `read_page_size=100`.
- Provider inventory: 1,913 operations across 35 published modules (881 reads, 1,032 writes). See
  issue #3915 for the full module-by-module tracking table.
- Executable today: 38 operations — 3 stream-backed GET reads (`users`, `meetings`, `webinars`), 3
  bounded `qss` module direct reads, 1 bounded `ai-companion` module direct read, 2 bounded
  `my-notes` module direct reads, 2 Healthcare direct reads, 1 Healthcare PATCH action, 5 Quality
  Management direct reads, 1 approval-gated Quality Management POST action, 4 sensitive Cobrowse
  SDK direct reads, 4 sensitive SCIM2 direct reads, 7 approval-gated SCIM2 actions, 4 approval-gated
  Chatbot actions, and 1 redacted Customer Managed Keys Hybrid archival-key action.
- Direct-read commands in this connector take only the request inputs the live provider artifact
  expressly documents. Cobrowse's `from`/`to` range is exposed because the operation prose declares
  it as a query input; a response-body field alone (including `page_size` and `next_page_token`) is
  not sufficient evidence of an accepted request parameter. This is a deliberate module-by-module
  scope-narrowing, not an oversight.
- Pending connector-local delivery: 1,804 operations have no shared foundation blocker, but still
  need bounded Zoom-specific contracts, schemas, safety evidence, and fixtures before they can
  become commands.
- Provider-side restrictions: 17 operations (five Information Barriers, seven Chat migration, one
  Meeting audit trail, and four Phone blocked-list routes) remain blocked until Zoom enables the
  corresponding product/account capability.
- Justified exclusions: 54 provider-deprecated operations remain recorded as `deprecated` ledger
  evidence and are not implemented.
