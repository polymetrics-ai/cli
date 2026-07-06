# Overview

Reads tyntec SMS messages, templates, sender IDs, and delivery reports through API list endpoints,
and sends approved SMS messages through the Messaging API.

Readable streams: `messages`, `templates`, `sender_ids`, `delivery_reports`.

Write actions: `send_message`.

Service API documentation: https://api.tyntec.com/reference/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); tyntec API key, sent as the apikey header on every request.
- `base_url` (optional, string); default `https://api.tyntec.com`; format `uri`; tyntec API base URL
  override for tests or proxies.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.tyntec.com`.

Authentication behavior:

- API key authentication in `apikey` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `sms/v1/messages` with query `limit`=`1`; `page`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `limit`; starts at
1; page size 100; maximum 1 page(s).

- `messages`: GET `sms/v1/messages` - records path `messages`; page-number pagination; page
  parameter `page`; size parameter `limit`; starts at 1; page size 100; maximum 1 page(s); computed
  output fields `created_at`.
- `templates`: GET `sms/v1/templates` - records path `templates`; page-number pagination; page
  parameter `page`; size parameter `limit`; starts at 1; page size 100; maximum 1 page(s).
- `sender_ids`: GET `sms/v1/sender-ids` - records path `sender_ids`; page-number pagination; page
  parameter `page`; size parameter `limit`; starts at 1; page size 100; maximum 1 page(s).
- `delivery_reports`: GET `sms/v1/reports` - records path `reports`; page-number pagination; page
  parameter `page`; size parameter `limit`; starts at 1; page size 100; maximum 1 page(s); computed
  output fields `created_at`.

## Write actions & risks

Overall write risk: sends billable SMS messages to recipient phone numbers; approval required before
delivery.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `send_message`: POST `sms/v1/messages` - kind `create`; body type `json`; required record fields
  `to`, `from`, `text`; accepted fields `callbackUrl`, `from`, `reference`, `text`, `to`; risk:
  sends a billable SMS message to the recipient phone number and may notify an external user.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s), 1 write-backed endpoint group(s).
