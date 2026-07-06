# Overview

Reads subscribers, lists, and campaigns, and writes subscriber create/upsert actions, through the
tinyEmail API.

Readable streams: `subscribers`, `lists`, `campaigns`.

Write actions: `create_subscriber`.

Service API documentation: https://tinyemail.com/api.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); tinyEmail API key, sent as the X-API-Key request header.
  Never logged.
- `base_url` (optional, string); default `https://api.tinyemail.com/v1`; format `uri`; tinyEmail API
  base URL override for tests or proxies.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.tinyemail.com/v1`.

Authentication behavior:

- API key authentication in `X-API-Key` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/subscribers` with query `limit`=`1`; `page`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `limit`; starts at
1; page size 100.

- `subscribers`: GET `/subscribers` - records path `data`; page-number pagination; page parameter
  `page`; size parameter `limit`; starts at 1; page size 100; emits passthrough records.
- `lists`: GET `/lists` - records path `data`; page-number pagination; page parameter `page`; size
  parameter `limit`; starts at 1; page size 100; emits passthrough records.
- `campaigns`: GET `/campaigns` - records path `data`; page-number pagination; page parameter
  `page`; size parameter `limit`; starts at 1; page size 100; emits passthrough records.

## Write actions & risks

Overall write risk: external tinyEmail API mutation: creates or upserts a subscriber (customer)
record, optionally assigning it to a named audience segment.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_subscriber`: POST `/segment/customer` - kind `create`; body type `json`; required record
  fields `email`; accepted fields `address1`, `address2`, `birthday`, `city`, `company`, `country`,
  `currency`, `email`, `firstName`, `lastName`, `lastOrderDate`, `lastOrderName`, `lastOrderPrice`,
  `ordersCount`, `phone`, `postalCode`, `province`, `segmentName`, and 5 more; risk: creates or
  upserts a subscriber (customer) into the caller's tinyEmail account, optionally into a named
  audience segment; low-risk external mutation, no approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 3 stream-backed endpoint group(s), 1 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=2.
