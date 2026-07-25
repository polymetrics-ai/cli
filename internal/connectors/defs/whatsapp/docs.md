# Overview

Reads the WhatsApp Business Platform (Cloud API + Business Management API) on Meta Graph API
v25.0 and models the WhatsApp Web multidevice (whatsmeow) access mode from
`https://github.com/vicentereig/whatsapp-cli`. Framed for an HMS / healthcare deployment paired with the Bahmni EMR connector
(parent #516): WhatsApp is the patient engagement / outbound notification channel while Bahmni
holds the clinical system of record.

Executable ETL streams (cloud mode): `phone_numbers`, `message_templates`, `subscribed_apps`,
`waba`.

Bounded direct-read commands cover single phone-number/template/profile/media detail plus the four
typed analytics read-queries (messaging, conversation, pricing, template) under the `json_redacted`
output policy. WhatsApp sends (all 11 message types), template and phone-number
administration, business-profile updates, QR/subscription management, and bounded media upload are
modeled as 28 typed reverse-ETL write actions.

The WhatsApp Web (whatsmeow) mode is modeled as documented, config-scoped ops (`spec.mode=web`);
it requires a local QR session and is not executed by the declarative Graph HTTP engine.

Service documentation: https://developers.facebook.com/docs/whatsapp/cloud-api.

## Auth setup

Connection fields:

- `mode` (required, string, enum cloud|web); default `cloud`; selects the active access mode. The
  two modes' credentials are never conflated.
- `access_token` (optional, secret, string); Meta Graph access token for cloud-mode Bearer auth;
  never logged.
- `base_url` (optional, string); default `https://graph.facebook.com/v25.0`; Graph base URL (includes version) override.
- `waba_id` (optional, string); WhatsApp Business Account ID for Business Management reads.
- `phone_number_id` (optional, string); Phone Number ID for messaging/media/profile/registration.
- `page_size` (optional, string); default `100`; records per page (1-100).
- `max_pages` (optional, string); default `0`; 0 exhausts a paginated stream.
- `start_date` (optional, string); format `date-time`; RFC3339 lower bound for analytics windows.
- `store_dir` (optional, string); WhatsApp Web (whatsmeow) local store directory for the QR session.

Secret fields are redacted in logs and write previews: `access_token`. No secret value is read or
printed at `pm connectors inspect whatsapp --json`.

Authentication behavior: cloud mode uses Bearer auth on `secrets.access_token` (falling back to no
auth when unset, for credential-free conformance); web mode uses a local whatsmeow QR session and
`store_dir` with no printed session secret.

Connection checks call GET `/{WABA_ID}/phone_numbers` with query `limit`=`1`.

## Streams notes

Default pagination: Graph cursor pagination; cursor parameter `after`; next token read from
`paging.cursors.after`. The `waba` stream is a single-object read with pagination disabled.

Streams are full-refresh over Graph list endpoints; the `waba` stream reads WABA metadata as one
record. Analytics are exposed as bounded typed direct-read queries rather than streams because the
Graph analytics surface is a time-windowed field-expansion on the WABA node.

Webhook-fed inbound messages and statuses are push-only Graph callbacks, not pollable streams; they
are recorded as blocked ledger rows and should be ingested by the EMR/webhook receiver.

## Write actions & risks

Write actions are declared in `writes.json` for WhatsApp sends (text/image/audio/video/document/
sticker/location/contacts/interactive/template/reaction), mark-as-read and typing, template
create/edit/delete, phone-number register/deregister/request-code/verify/two-step-PIN, business-
profile update, app subscribe/unsubscribe, QR create/delete, and bounded multipart media upload +
media delete.

Safety gates:

- Use reverse ETL plan -> preview -> approval -> execute.
- Sends and destructive/admin actions declare `confirm: destructive` and require `--confirm
  destructive`.
- Recipient numbers and message bodies are patient PHI: redacted by default in command plans;
  sends require patient consent and Meta template pre-approval.
- Multipart media upload accepts only a declared local file path field and enforces a bounded byte
  limit; no generic upload or raw Graph body flag is exposed.
- No generic raw HTTP write, arbitrary Graph API method/path/body, raw whatsmeow method, generic
  shell write, or SQL write is exposed.

Read risk: external Meta Graph API read of WABA metadata, phone numbers, templates, subscribed apps, and analytics; direct reads are bounded and PHI-redacted.

Write risk: typed WhatsApp reverse ETL: patient message sends (PHI), template and phone-number administration, business-profile updates, QR/subscription management, and bounded media upload; message bodies and recipient numbers are redacted in plans.

Approval: reverse ETL requires plan, preview, approval, execute; sends and destructive/admin actions require --confirm destructive plus healthcare consent/template pre-approval.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage is inventoried from Meta WhatsApp Cloud API + Business Management API v25.0 (reviewed
  2026-07-25) and the vicentereig/whatsapp-cli command surface.
- Executable coverage (cloud mode): 4 stream endpoints, 4 GET direct reads, 4 typed analytics
  read-queries, and 28 typed reverse-ETL write actions.
- WhatsApp Web (whatsmeow) mode: modeled as documented, config-scoped ops
  (`unsupported_local`); live execution requires a whatsmeow QR session and is human-gated.
- Live sends, template submissions, phone-number administration, media payloads beyond fixtures,
  and any real patient messaging are human-gated (credentials, consent, template pre-approval).
