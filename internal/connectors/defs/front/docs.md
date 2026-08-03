# Overview

Front is modeled as a definition-owned connector bundle for the documented Front Platform API sources listed in `https://dev.frontapp.com/llms.txt` and the linked `reference/*.md` pages. This worker parsed the embedded OpenAPI snippets from those reference pages and reconciled 255 documented operation pages (254 unique method/path pairs; `PATCH /channels/{channel_id}` is documented twice and one row is retained as a duplicate ledger entry for count reconciliation).

Implemented read streams cover 115 fixed Front GET/changefeed surfaces. Implemented reverse ETL metadata covers 129 documented Core REST and Channels API write operations (118 Core REST writes, 1 application event trigger surface, and 10 custom-channel call/message sync writes: create/update call, call recording/summary/transcript, application message template sync/update, inbound/outbound message sync, and external message status update). Direct/provider search metadata covers 5 bounded JSON operations. Binary attachment downloads (5) remain represented as planned/blocked fixed operations without raw passthrough, pending a shared binary-download execution foundation this connector-local worker cannot add. The duplicate `PATCH /channels/{channel_id}` documentation row is recorded once for count reconciliation; its executable behavior is covered by the single `update_channel` write action.

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

Direct/provider-search commands for analytics exports/reports and conversation search are fixed, bounded, JSON-redacted direct reads. Binary attachment download operations are recorded in `operations.json` with max-byte bounds but remain planned because this connector-local worker cannot add the shared binary file-output execution foundation.

## Known limits

- No live Front credentials or provider calls were used; this bundle is fixture/conformance backed and remains uncertified for live behavior.
- The generated stream schemas are intentionally permissive (`projection: passthrough`, `additionalProperties: true`) pending a future field-tightening pass; they preserve records rather than dropping undocumented fields.
- Parameterized streams require the corresponding config values and are not safe as an all-stream default sync without operator selection.
- Binary downloads are typed and bounded in metadata but not claimed executable until a shared binary-download execution contract is available.
- `sync_inbound_message` and `sync_outbound_message` omit the provider's optional `attachments` field (a binary-payload array) from their record schemas; every other documented field is modeled. This is a strictly more permissive scope narrowing (never stricter than the provider's own requirements) pending the same shared binary-payload execution foundation the attachment-download operations need — sending an attachment through these two actions is not yet possible, everything else about the write is.
- `add_call_recording` is a multipart file-upload write with no dynamic conformance fixture (see Write actions & risks); its correctness rests on static schema/CLI-mapping validation and `go vet`/`go build`, the same bar the repo's only other multipart-write connector (`gong`) is held to.
- The duplicate `PATCH /channels/{channel_id}` docs row is recorded as a duplicate operation row; the executable Core API write action is declared once.
