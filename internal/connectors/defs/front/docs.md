# Overview

Front is modeled as a definition-owned connector bundle for the documented Front Platform API sources listed in `https://dev.frontapp.com/llms.txt` and the linked `reference/*.md` pages. This worker parsed the embedded OpenAPI snippets from those reference pages and reconciled 255 documented operation pages (254 unique method/path pairs; `PATCH /channels/{channel_id}` is documented twice and one row is retained as a duplicate ledger entry for count reconciliation).

Implemented read streams cover 115 fixed Front GET/changefeed surfaces. Implemented reverse ETL metadata covers 129 documented Core REST and Channels API write operations (118 Core REST writes, 1 application event trigger surface, and 10 custom-channel call/message sync writes: create/update call, call recording/summary/transcript, application message template sync/update, inbound/outbound message sync, and external message status update). Direct/provider search metadata covers 5 bounded JSON operations. Five attachment downloads are implemented as fixed `binary_download` operations with project-relative output containment and per-operation byte limits. The duplicate `PATCH /channels/{channel_id}` documentation row is recorded once for count reconciliation; its executable behavior is covered by the single `update_channel` write action.

Service API documentation: https://dev.frontapp.com/llms.txt.

## Auth setup

Connection fields:

- `api_key` (required, secret, string): Front API token, sent as `Authorization: Bearer <api_key>`. Never log or store token values in fixtures.
- `base_url` (optional, string): default `https://api2.frontapp.com`.
- `page_limit` (optional, string): default `50`; sent as `limit` on paginated list streams.
- Optional path-parameter config fields such as `account_id`, `conversation_id`, `inbox_id`, and `team_id` are used only by parameterized streams. A parameterized stream fails closed if its required config value is absent.

Connection checks call `GET /inboxes` with bearer authentication.

## Streams notes

The bundle declares 115 Front read streams. List streams read records from `_results` and follow `_pagination.next`; single-resource streams use a single-object projection. Projection is `passthrough` for generated full-surface streams so the connector does not silently drop Front fields before a future schema-tightening pass.

The legacy fixture-backed streams remain available with their original names: `contacts`, `conversations`, `inboxes`, `tags`, `teammates`, and `channels`. Event/changefeed GET surfaces are represented as read streams; the mutating application event trigger is a fixed reverse-ETL action rather than a CDC reader claim. `metadata.capabilities.cdc` remains false because there is no streaming CDC implementation.

## Write actions & risks

`writes.json` declares fixed Front write actions only; there is no generic method/path/body escape hatch. Every action has a record schema, path field list where applicable, risk text, and `confirm: "destructive"` so execution stays behind the existing reverse ETL plan → preview → explicit approval → execute path plus typed destructive confirmation. DELETE actions declare idempotent 404 handling through `missing_ok_status: [404]`.

Custom-channel call and message sync actions (`create_call`, `update_call`, `add_call_summary`, `add_call_transcript`, `sync_application_message_template`, `update_application_message_template`, `sync_inbound_message`, `sync_outbound_message`, `update_external_message_status`) use the same fixed-endpoint reverse ETL path as every other write; they operate against channels this connector's own `create_a_channel`/`update_channel`/`validate_channel` actions can create and manage, not a Front Marketplace-partner-only surface. `add_call_recording` uses the engine's typed multipart write support (`body_type: "multipart"`) to upload a bounded, project-relative local audio file path as the call recording attachment, matching the pattern already used by the `gong` connector's `upload_call_media`/`upload_crm_entities` actions; like those, it has no `fixtures/writes/` dynamic-replay fixture (multipart write-shape conformance is not yet exercised by any bundle in this repo) and its coverage rests on `connectorgen validate`'s static schema/CLI-mapping checks plus `go build`/`go vet`.

Direct/provider-search commands for analytics exports/reports and conversation search are fixed, bounded, JSON-redacted direct reads. The five attachment download commands use the engine's implemented `binary_download` contract, fixed Front endpoints, project-relative destinations, and explicit maximum response sizes.

## Conditional request contracts

`schemas/request_contracts.json` preserves Front's draft-07 composed request rules for analytics export discriminators, analytics report filter alternatives, composed draft/template requiredness, exclusive link and conversation destinations, mutually exclusive conversation status selectors, channel/call variants, teammate-group contact selectors, the email-only imported-message body format, and merge cardinality. The CLI exposes fixed typed commands for each provider-valid branch: message, fixed-event, parameterized-event, and Smart QA exports; all six analytics report filter families; both conversation-link and link selectors; built-in and custom conversation statuses; inbox/teammate discussion/task destinations; custom/SMTP/Twilio channels; inbound/outbound calls; and both merge forms. Each command's help states its branch-specific requirements and never accepts a raw body.

Front's [application-template schema](https://dev.frontapp.com/reference/sync-application-message-template.md) requires the `variable_mappings` property but declares no minimum array length, so record-driven sync accepts an empty array for a static template; when a mapping is present, its `uid`, `name`, and `type` fields are required. The CLI command deliberately constructs one complete mapping and identifies itself as the one-mapping variant. Front's [call schema](https://dev.frontapp.com/reference/create-call.md) requires a non-null string `parent_external_call_id` in both direction branches. Front's [channel schema](https://dev.frontapp.com/reference/create-a-channel.md) restricts `settings.sid` and `settings.auth_token` to Twilio channels, and the composed custom branch rejects both fields.

The executable schemas preserve Front's documented nullable transitions for unassignment, reminder cancellation, signature removal, root-folder moves, unrestricted template inboxes, unrestricted signature channels, and root-tag moves. Provider-declared `maxItems` bounds are enforced for contact, follower, and link arrays. Front's tag-name, conversation-description, call-summary, and Channel API external-ID string bounds are preserved in `schemas/request_contracts.json`; executable enforcement is blocked because `internal/connectors/engine/schema.go` currently rejects draft-07 `minLength` and `maxLength` as unknown keywords, and this connector-owned lane does not change shared validation. Analytics columns, analytics filter arrays, contact handles, and outbound recipients remain required where Front requires the property, but no local `minItems` is invented when Front declares none.

Front's [channel creation](https://dev.frontapp.com/reference/create-a-channel.md) and [Core channel update](https://dev.frontapp.com/reference/update-channel.md) schemas restrict `settings.undo_send_time` to `0`, `5`, `10`, `15`, `30`, or `60`; the duplicate [Channel API update](https://dev.frontapp.com/reference/update-channel-1.md) restricts `status` to `offline` or `ok`. The three template-creation references require a supplied `inbox_ids` array to be non-empty while preserving omission and `null` as unrestricted forms. Front's teammate-group request schemas restrict contact access to `all`, `contact_groups`, `contact_lists`, or `none`, with each ID array valid only for its matching selector. The [imported-message schema](https://dev.frontapp.com/reference/import-inbox-message.md) permits `body_format` only for email imports, and the [teammate update schema](https://dev.frontapp.com/reference/update-teammate.md) restricts usernames to lowercase letters, numbers, and underscores. Executable schemas enforce the non-composed enum, cardinality, and username-pattern rules; the two cross-field rules remain in the composed sidecar pending shared evaluation.

Current main's minimal schema compiler does not evaluate `oneOf`, `allOf`, `if`/`then`, `not`, or `anyOf`, so the executable `record_schema`/`body_schema` copies retain only the compatible typed field and cardinality subset. Full record-driven enforcement depends on the in-flight shared normalized-schema composition support; this connector does not implement a private composition evaluator or weaken shared validation. The fixed CLI variants and fixtures are provider-valid while that shared dependency is pending.

Outbound message sends, message replies, and `mark_message_seen` are explicitly non-batchable. Their single-record commands remain available, and seen-state changes must represent a real end-user action.

## Deferred multipart fields

The following optional provider file fields are deliberately absent from the JSON write schemas because Front requires multipart form data when they are supplied:

- `attachments` on [`create_draft`](https://dev.frontapp.com/reference/create-draft.md), [`receive_custom_messages`](https://dev.frontapp.com/reference/receive-custom-messages.md), [`create_message`](https://dev.frontapp.com/reference/create-message.md), [`add_comment_reply`](https://dev.frontapp.com/reference/add-comment-reply.md), [`add_comment`](https://dev.frontapp.com/reference/add-comment.md), [`create_draft_reply`](https://dev.frontapp.com/reference/create-draft-reply.md), [`create_message_reply`](https://dev.frontapp.com/reference/create-message-reply.md), [`create_conversation.comment`](https://dev.frontapp.com/reference/create-conversation.md), [`edit_draft`](https://dev.frontapp.com/reference/edit-draft.md), [`import_inbox_message`](https://dev.frontapp.com/reference/import-inbox-message.md), [`create_message_template`](https://dev.frontapp.com/reference/create-message-template.md), [`create_team_message_template`](https://dev.frontapp.com/reference/create-team-message-template.md), [`sync_inbound_message`](https://dev.frontapp.com/reference/sync-inbound-message.md), and [`sync_outbound_message`](https://dev.frontapp.com/reference/sync-outbound-message.md).
- `avatar` on [`update_a_contact`](https://dev.frontapp.com/reference/update-a-contact.md), [`create_contact`](https://dev.frontapp.com/reference/create-contact.md), [`create_teammate_contact`](https://dev.frontapp.com/reference/create-teammate-contact.md), and [`create_team_contact`](https://dev.frontapp.com/reference/create-team-contact.md).

Supporting these fields requires an array-aware extension to the typed multipart contract: repeated bounded file parts (including `attachments[]`), repeated scalar parts for array-valued form fields, deterministic encoding for nested object fields, project-root containment, approved payload digests, media-type checks, and per-file plus aggregate byte limits. The existing flat declared-part contract already expresses `add_call_recording`, so that action remains implemented without exposing a generic multipart or raw-body path.

## Known limits

- No live Front credentials or provider calls were used; this bundle is fixture/conformance backed and remains uncertified for live behavior.
- The generated stream schemas are intentionally permissive (`projection: passthrough`, `additionalProperties: true`) pending a future field-tightening pass; they preserve records rather than dropping undocumented fields.
- Parameterized streams require the corresponding config values and are not safe as an all-stream default sync without operator selection.
- File-valued `attachments` and `avatar` request fields remain deferred exactly as listed above; supplying them requires the bounded array-aware multipart extension described there.
- `add_call_recording` is a multipart file-upload write with no dynamic conformance fixture (see Write actions & risks); its correctness rests on static schema/CLI-mapping validation and `go vet`/`go build`, the same bar the repo's only other multipart-write connector (`gong`) is held to.
- Record-driven enforcement of the composed contracts in `schemas/request_contracts.json` remains paused on the shared normalized-schema composition dependency; fixed CLI variants already construct provider-valid branches without a generic body escape.
- The duplicate `PATCH /channels/{channel_id}` docs row is recorded as a duplicate operation row; the executable Core API write action is declared once.
