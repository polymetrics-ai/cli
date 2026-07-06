# Overview

Reads Apify dataset items and dataset metadata (item_collection, dataset_collection, dataset)
through the Apify API v2.

Readable streams: `item_collection`, `dataset_collection`, `dataset`.

This connector is read-only; no write actions are declared.

Service API documentation: https://docs.apify.com/api/v2.

## Auth setup

Connection fields:

- `base_url` (optional, string).
- `dataset_id` (required, string).
- `mode` (optional, string).
- `token` (required, secret, string); Personal API token of your Apify account. In Apify Console,
  you can find your API token in the <a
  href="https://console.apify.com/account/integrations">Settings section under the Integrations
  tab</a> after you login. See the <a
  href="https://docs.apify.com/platform/integrations/api#api-token">Apify Docs</a> for more
  information.

Secret fields are redacted in logs and write previews: `token`.

Provide the secret fields listed above. Authentication is applied by the connector-specific
implementation for this service.

Requests use the configured `base_url` value after applying defaults.

Connection checks use a connector-managed request.

## Streams notes

Default pagination: single request; no pagination.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `item_collection`: GET connector-managed request path - records path `data`.
- `dataset_collection`: GET connector-managed request path - records path `data`; incremental cursor
  `createdAt`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.
- `dataset`: GET connector-managed request path - records path `data`; incremental cursor
  `modifiedAt`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.

## Write actions & risks

This connector is read-only; no reverse-ETL write actions are declared.

## Known limits

- API coverage includes 3 stream-backed endpoint group(s).
- Client-side incremental filtering is used for: `dataset_collection`, `dataset`.
