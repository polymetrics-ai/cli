# Overview

Reads SendPulse address books, campaigns, senders, per-book emails, and the account blacklist, and
writes address-book/sender/blacklist lifecycle mutations and campaign create/cancel actions through
the SendPulse API.

Readable streams: `addressbooks`, `campaigns`, `senders`, `blacklist`, `emails_in_book`.

Write actions: `create_addressbook`, `update_addressbook`, `delete_addressbook`,
`add_emails_to_book`, `remove_emails_from_book`, `create_campaign`, `cancel_campaign`, `add_sender`,
`remove_sender`, `add_to_blacklist`, `remove_from_blacklist`.

Service API documentation: https://sendpulse.com/integrations/api.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api.sendpulse.com`; format `uri`; SendPulse API
  base URL override for tests or proxies.
- `client_id` (required, secret, string); SendPulse API client ID (OAuth2 client-credentials). Never
  logged.
- `client_secret` (required, secret, string); SendPulse API client secret (OAuth2
  client-credentials). Never logged.
- `token_url` (optional, string); default `https://api.sendpulse.com/oauth/access_token`; format
  `uri`.

Secret fields are redacted in logs and write previews: `client_id`, `client_secret`.

Default configuration values: `base_url=https://api.sendpulse.com`,
`token_url=https://api.sendpulse.com/oauth/access_token`.

Authentication behavior:

- OAuth 2.0 client credentials authentication using `config.token_url`, `secrets.client_id`,
  `secrets.client_secret`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/addressbooks`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `limit`; starts at
1; page size 100; maximum 1 page(s).

Pagination by stream: none: `blacklist`; page_number: `addressbooks`, `campaigns`, `senders`,
`emails_in_book`.

- `addressbooks`: GET `/addressbooks` - records path `.`; page-number pagination; page parameter
  `page`; size parameter `limit`; starts at 1; page size 100; maximum 1 page(s); emits passthrough
  records.
- `campaigns`: GET `/campaigns` - records path `.`; page-number pagination; page parameter `page`;
  size parameter `limit`; starts at 1; page size 100; maximum 1 page(s); emits passthrough records.
- `senders`: GET `/senders` - records path `.`; page-number pagination; page parameter `page`; size
  parameter `limit`; starts at 1; page size 100; maximum 1 page(s); emits passthrough records.
- `blacklist`: GET `/blacklist` - records path `.`; emits passthrough records.
- `emails_in_book`: GET `/addressbooks/{{ fanout.id }}/emails` - records path `.`; page-number
  pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100; maximum 1
  page(s); fan-out; ids from request `/addressbooks`; id-list records path `.`; id field `id`; id
  inserted into the request path; stamps `book_id`; emits passthrough records.

## Write actions & risks

Overall write risk: external SendPulse API mutation, including creating email campaigns that may
send to real subscribers.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_addressbook`: POST `/addressbooks` - kind `create`; body type `json`; required record
  fields `bookName`; accepted fields `bookName`; risk: creates a new address book (mailing list);
  external mutation, approval required.
- `update_addressbook`: PUT `/addressbooks/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`, `name`; accepted fields `id`, `name`; risk: renames an
  existing address book.
- `delete_addressbook`: DELETE `/addressbooks/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; risk: permanently removes an address book and all its subscriber
  associations; irreversible.
- `add_emails_to_book`: POST `/addressbooks/{{ record.id }}/emails` - kind `update`; body type
  `json`; path fields `id`; body fields `emails`; required record fields `id`, `emails`; accepted
  fields `emails`, `id`; risk: subscribes new email addresses to an address book; each add may
  trigger a double opt-in confirmation email depending on account settings.
- `remove_emails_from_book`: DELETE `/addressbooks/{{ record.id }}/emails` - kind `delete`; body
  type `json`; path fields `id`; body fields `emails`; required record fields `id`, `emails`;
  accepted fields `emails`, `id`; missing records treated as success for status `404`; risk:
  unsubscribes the given email addresses from an address book; irreversible without re-adding them.
- `create_campaign`: POST `/campaigns` - kind `create`; body type `json`; required record fields
  `sender_name`, `sender_email`, `subject`, `body`, `list_id`; accepted fields `body`, `list_id`,
  `name`, `sender_email`, `sender_name`, `subject`.
- `cancel_campaign`: DELETE `/campaigns/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; risk: cancels a scheduled/in-progress campaign; stops further sends but does not
  un-send already-delivered emails.
- `add_sender`: POST `/senders` - kind `create`; body type `json`; required record fields `email`,
  `name`; accepted fields `email`, `name`; risk: registers a new sender email address, which
  SendPulse will send an activation email to; low-risk external mutation.
- `remove_sender`: DELETE `/senders` - kind `delete`; body type `json`; body fields `email`;
  required record fields `email`; accepted fields `email`; missing records treated as success for
  status `404`; risk: removes a sender email address; any campaign still referencing it as its
  sender will fail to send.
- `add_to_blacklist`: POST `/blacklist` - kind `create`; body type `json`; required record fields
  `emails`; accepted fields `comment`, `emails`; risk: permanently suppresses future sends to the
  given address(es) account-wide.
- `remove_from_blacklist`: DELETE `/blacklist` - kind `delete`; body type `json`; body fields
  `emails`; required record fields `emails`; accepted fields `emails`; missing records treated as
  success for status `404`; risk: removes an address from the account-wide suppression list; future
  campaigns can reach it again.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s), 11 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=4, non_data_endpoint=9, out_of_scope=38.
