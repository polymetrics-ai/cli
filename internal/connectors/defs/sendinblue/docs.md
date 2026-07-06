# Overview

Reads Sendinblue/Brevo contacts, campaigns, lists, and senders through the Brevo API.

Readable streams: `contacts`, `email_campaigns`, `contacts_lists`, `senders`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.brevo.com/reference.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Brevo (Sendinblue) API key, sent as the 'api-key' request
  header. Never logged.
- `base_url` (optional, string); default `https://api.brevo.com/v3`; format `uri`; Brevo API base
  URL override for tests or proxies.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.brevo.com/v3`.

Authentication behavior:

- API key authentication in `api-key` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/account`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `limit`;
page size 100.

- `contacts`: GET `/contacts` - records path `contacts`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100; emits passthrough records.
- `email_campaigns`: GET `/emailCampaigns` - records path `campaigns`; offset/limit pagination;
  offset parameter `offset`; limit parameter `limit`; page size 100; emits passthrough records.
- `contacts_lists`: GET `/contacts/lists` - records path `lists`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 100; emits passthrough records.
- `senders`: GET `/senders` - records path `senders`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100; emits passthrough records.

## Write actions & risks

This connector is read-only. Read behavior: external Brevo (Sendinblue) API read of contact,
campaign, list, and sender data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  non_data_endpoint=1, out_of_scope=2.
