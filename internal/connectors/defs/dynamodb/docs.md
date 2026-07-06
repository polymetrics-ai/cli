# Overview

Reads DynamoDB table items through the AWS JSON HTTP API (DynamoDB_20120810.Scan), authenticated
with hand-rolled AWS Signature Version 4 request signing. Read-only source; no write support.

Readable streams: `items`.

This connector is read-only; no write actions are declared.

Service API documentation: https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/.

## Auth setup

Connection fields:

- `access_key_id` (required, secret, string); AWS access key id used to derive the AWS Signature
  Version 4 (SigV4) Authorization header. Never logged.
- `base_url` (optional, string); format `uri`; Ignored when endpoint is set.
- `endpoint` (optional, string); format `uri`; DynamoDB JSON HTTP API endpoint, e.g.
  https://dynamodb.us-east-1.amazonaws.com. One of endpoint/base_url is required for live mode --
  enforced by native/dynamodb's own Go config resolution (connection.go's resolveEndpoint), not this
  schema's required[] (JSON Schema's required[] cannot express an either-or pair).
- `max_pages` (optional, string); default `100`; Maximum Scan pages to request (a positive integer).
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Scan Limit per page (a positive integer).
- `region` (required, string); AWS region used in the SigV4 credential scope (e.g. us-east-1) and,
  conventionally, the endpoint's own region.
- `secret_access_key` (required, secret, string); AWS secret access key used to derive the SigV4
  signing key HMAC chain. Never placed in a header or logged.
- `table` (optional, string); Ignored when table_name is set.
- `table_name` (optional, string); DynamoDB table name to Scan. One of table_name/table is required
  for live mode -- enforced by native/dynamodb's own Go config resolution (reader.go's tableName),
  not this schema's required[].

Secret fields are redacted in logs and write previews: `access_key_id`, `secret_access_key`.

Default configuration values: `max_pages=100`, `page_size=100`.

Provide the secret fields listed above. Authentication is applied by the connector-specific
implementation for this service.

Requests use the configured `endpoint` value after applying defaults.

Connection checks call POST `/`.

## Streams notes

Default pagination: single request; no pagination.

- `items`: POST `/` - records path `Items`.

## Write actions & risks

This connector is read-only. Read behavior: external AWS DynamoDB Scan of table items.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 1 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=1, out_of_scope=7.
