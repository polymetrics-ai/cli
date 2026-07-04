# Overview

Postmark App is a declarative HTTP bundle for the Postmark API. It keeps legacy-projected message activity streams and expands the server-token API surface with streams for server settings, bulk email status, bounces, templates, message streams, message details/dumps, opens/clicks, stats, inbound rules, webhooks, and suppressions. It also exposes server-token write actions for email sends, template and message-stream management, inbound message processing, inbound rules, webhooks, suppressions, bounce activation, and current-server settings.

## Auth setup

Provide a Postmark server token via the `X-Postmark-Server-Token` secret. It is sent as the `X-Postmark-Server-Token` header and is never logged. `base_url` defaults to `https://api.postmarkapp.com`. Account-token endpoints remain excluded because the current engine has no per-stream or per-action auth/header override.

## Streams notes

`outbound_messages` and `inbound_messages` retain legacy schema projection from the hand-written connector, including computed snake_case fields from Postmark's PascalCase message fields. Newly added streams use passthrough projection to preserve Postmark response shapes. Offset-list endpoints use `count`/`offset` pagination with a fixed page size of 100, matching legacy's default page size. Detail streams require config ids such as `message_id`, `bounce_id`, `template_id_or_alias`, `message_stream_id`, `webhook_id`, or `bulk_request_id`.

## Write actions & risks

Write actions can send live email and mutate Postmark resources. Destructive delete actions are marked with destructive confirmation semantics. All reverse ETL writes require plan preview and approval. Batch send endpoints that require root JSON arrays are not exposed because writes.json can only construct object-shaped JSON bodies from records.

## Known limits

- Account-token endpoints such as Servers, Domains, Sender Signatures, Data Removals, and template push are excluded and recorded in `docs/migration/quarantine.json`; the engine cannot select `X-Postmark-Account-Token` for only those streams/actions while using `X-Postmark-Server-Token` elsewhere.
- `POST /email/batch` and `POST /email/batchWithTemplates` require a root JSON array request body. The write dialect currently builds JSON objects from record maps and has no root-array body mode.
- Message dump and bounce dump endpoints return JSON objects containing raw source text in a `Body` field; they are included as single-object streams, not parsed into MIME sub-records.
