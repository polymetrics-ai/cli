# Overview

Reads Apptivo CRM customers, contacts, leads, and opportunities through the Apptivo REST DAO API
(full refresh); deletes CRM customer records via the documented deleteCustomer DAO action.

Readable streams: `customers`, `contacts`, `leads`, `opportunities`.

Write actions: `remove_customer`.

Service API documentation: https://www.apptivo.com/api-reference/.

## Auth setup

Connection fields:

- `access_key` (required, secret, string); Apptivo accessKey, sent as the accessKey query parameter
  on every request. Never logged.
- `api_key` (required, secret, string); Apptivo apiKey, sent as the apiKey query parameter on every
  request. Never logged.
- `base_url` (optional, string); default `https://app.apptivo.com`; format `uri`; Apptivo API base
  URL override for tests or proxies.

Secret fields are redacted in logs and write previews: `access_key`, `api_key`.

Default configuration values: `base_url=https://app.apptivo.com`.

Authentication behavior:

- API key authentication in query parameter `apiKey` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/app/dao/v6/customers` with query `a`=`getAll`; `numRecords`=`1`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `startIndex`; limit parameter
`numRecords`; page size 100.

- `customers`: GET `/app/dao/v6/customers` - records path `data`; query `a`=`getAll`;
  `accessKey`=`{{ secrets.access_key }}`; offset/limit pagination; offset parameter `startIndex`;
  limit parameter `numRecords`; page size 100.
- `contacts`: GET `/app/dao/v6/contacts` - records path `data`; query `a`=`getAll`; `accessKey`=`{{
  secrets.access_key }}`; offset/limit pagination; offset parameter `startIndex`; limit parameter
  `numRecords`; page size 100.
- `leads`: GET `/app/dao/v6/leads` - records path `data`; query `a`=`getAll`; `accessKey`=`{{
  secrets.access_key }}`; offset/limit pagination; offset parameter `startIndex`; limit parameter
  `numRecords`; page size 100.
- `opportunities`: GET `/app/dao/v6/opportunities` - records path `data`; query `a`=`getAll`;
  `accessKey`=`{{ secrets.access_key }}`; offset/limit pagination; offset parameter `startIndex`;
  limit parameter `numRecords`; page size 100.

## Write actions & risks

Overall write risk: external mutation: irreversibly deletes a CRM customer record.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `remove_customer`: GET `/app/dao/v6/customers?a=delete&customerId={{ record.id }}&apiKey={{
  secrets.api_key }}&accessKey={{ secrets.access_key }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; confirmation `destructive`; risk:
  irreversible external deletion of a CRM customer record; approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s), 1 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=5, non_data_endpoint=4, out_of_scope=12.
