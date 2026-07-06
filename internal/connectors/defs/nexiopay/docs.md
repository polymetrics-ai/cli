# Overview

Reads Nexio Pay card tokens, payout recipients, spendbacks, payment types, terminals, and the API
user via the Nexio REST API.

Readable streams: `card_tokens`, `recipients`, `spendbacks`, `payment_types`, `terminal_list`,
`user`.

This connector is read-only; no write actions are declared.

Service API documentation: https://docs.nexiopay.com/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Nexio Pay API key, used as the Basic auth password. Never
  logged.
- `base_url` (optional, string); default `https://api.nexiopay.com`; format `uri`; Nexio Pay API
  base URL override for tests, proxies, or a non-default subdomain.
- `mode` (optional, string).
- `username` (required, secret, string); Nexio Pay API username, used as the Basic auth username.
  Part of the credential pair; never logged.

Secret fields are redacted in logs and write previews: `api_key`, `username`.

Default configuration values: `base_url=https://api.nexiopay.com`.

Authentication behavior:

- HTTP Basic authentication using `secrets.username`, `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/user/v3/account/whoAmI`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `limit`;
page size 10.

Pagination by stream: none: `payment_types`, `terminal_list`, `user`; offset_limit: `card_tokens`,
`recipients`, `spendbacks`.

- `card_tokens`: GET `/card/v3` - records path `rows`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 10.
- `recipients`: GET `/payout/v3/recipient` - records path `rows`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 10.
- `spendbacks`: GET `/payout/v3/spendback` - records path `rows`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 10.
- `payment_types`: GET `/transaction/v3/paymentTypes` - records at response root.
- `terminal_list`: GET `/pay/v3/getTerminalList` - records at response root.
- `user`: GET `/user/v3/account/whoAmI` - records at response root.

## Write actions & risks

This connector is read-only. Read behavior: external Nexio Pay API read of card tokens, payout, and
account data.

## Known limits

- Batch defaults: read_page_size=10.
- API coverage includes 6 stream-backed endpoint group(s).
