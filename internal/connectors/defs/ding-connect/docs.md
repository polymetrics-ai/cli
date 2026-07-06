# Overview

Reads DingConnect reference catalogs (countries, currencies, regions, providers, products, product
descriptions, promotions, provider status, error code descriptions, account balance) through the
DingConnect REST API, and sends real-money mobile top-up transfers.

Readable streams: `countries`, `currencies`, `regions`, `providers`, `products`,
`product_descriptions`, `promotions`, `provider_status`, `error_code_descriptions`, `balance`.

Write actions: `send_transfer`.

Service API documentation: https://docs.ding.com/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); DingConnect API key, sent as the api_key request header.
  Used only for header auth; never logged.
- `base_url` (optional, string); default `https://api.dingconnect.com`; format `uri`; DingConnect
  API base URL override for tests or proxies.
- `mode` (optional, string).
- `x_correlation_id` (optional, string); Optional X-Correlation-Id header value for request tracing.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.dingconnect.com`.

Authentication behavior:

- API key authentication in `api_key` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/api/V1/GetCurrencies`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `Skip`; page size 100.

Pagination by stream: none: `product_descriptions`, `promotions`, `provider_status`,
`error_code_descriptions`, `balance`; offset_limit: `countries`, `currencies`, `regions`,
`providers`, `products`.

- `countries`: GET `/api/V1/GetCountries` - records path `Items`; offset/limit pagination; offset
  parameter `Skip`; page size 100; computed output fields `uuid`.
- `currencies`: GET `/api/V1/GetCurrencies` - records path `Items`; offset/limit pagination; offset
  parameter `Skip`; page size 100; computed output fields `uuid`.
- `regions`: GET `/api/V1/GetRegions` - records path `Items`; offset/limit pagination; offset
  parameter `Skip`; page size 100; computed output fields `uuid`.
- `providers`: GET `/api/V1/GetProviders` - records path `Items`; offset/limit pagination; offset
  parameter `Skip`; page size 100; computed output fields `uuid`.
- `products`: GET `/api/V1/GetProducts` - records path `Items`; offset/limit pagination; offset
  parameter `Skip`; page size 100; computed output fields `uuid`.
- `product_descriptions`: GET `/api/V1/GetProductDescriptions` - records path `Items`; computed
  output fields `uuid`.
- `promotions`: GET `/api/V1/GetPromotions` - records path `Items`; computed output fields `uuid`.
- `provider_status`: GET `/api/V1/GetProviderStatus` - records path `Items`; computed output fields
  `uuid`.
- `error_code_descriptions`: GET `/api/V1/GetErrorCodeDescriptions` - records path `Items`; computed
  output fields `uuid`.
- `balance`: GET `/api/V1/GetBalance` - single-object response; records path `.`; computed output
  fields `uuid`.

## Write actions & risks

Overall write risk: external mutation; sends a real-money mobile top-up/airtime transfer and deducts
the distributor's live DingConnect balance.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `send_transfer`: POST `/api/V1/SendTransfer` - kind `create`; body type `json`; required record
  fields `SkuCode`, `SendValue`, `AccountNumber`, `DistributorRef`; accepted fields `AccountNumber`,
  `DistributorRef`, `SendCurrencyIso`, `SendValue`, `SkuCode`, `ValidateOnly`; risk: external
  mutation; sends a real-money mobile top-up/airtime transfer to a live account and deducts the
  distributor's DingConnect balance unless ValidateOnly is set; approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 10 stream-backed endpoint group(s), 1 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=1, out_of_scope=4.
