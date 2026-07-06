# Overview

Reads AWS CloudTrail management events (last 90 days) via the LookupEvents API using AWS Signature
V4 authentication. Read-only.

Readable streams: `management_events`, `read_only_events`, `write_only_events`, `console_logins`.

This connector is read-only; no write actions are declared.

Service API documentation:
https://docs.aws.amazon.com/awscloudtrail/latest/userguide/cloudtrail-user-guide.html.

## Auth setup

Connection fields:

- `aws_key_id` (required, secret, string).
- `aws_region_name` (required, string); The default AWS Region to use, for example, us-west-1 or
  us-west-2. When specifying a Region inline during client initialization, this property is named
  region_name.
- `aws_secret_key` (required, secret, string).
- `base_url` (optional, string).
- `lookup_attributes_filter` (optional, string).
- `mode` (optional, string).
- `start_date` (optional, string); The date you would like to replicate data. Data in AWS CloudTrail
  is available for last 90 days only. Format: YYYY-MM-DD.

Secret fields are redacted in logs and write previews: `aws_key_id`, `aws_secret_key`.

Provide the secret fields listed above. Authentication is applied by the connector-specific
implementation for this service.

Requests use the configured `base_url` value after applying defaults.

Connection checks use a connector-managed request.

## Streams notes

Default pagination: single request; no pagination.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `management_events`: GET connector-managed request path - records path `data`; incremental cursor
  `EventTime`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.
- `read_only_events`: GET connector-managed request path - records path `data`; incremental cursor
  `EventTime`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.
- `write_only_events`: GET connector-managed request path - records path `data`; incremental cursor
  `EventTime`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.
- `console_logins`: GET connector-managed request path - records path `data`; incremental cursor
  `EventTime`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.

## Write actions & risks

This connector is read-only; no reverse-ETL write actions are declared.

## Known limits

- API coverage includes 4 stream-backed endpoint group(s).
- Client-side incremental filtering is used for: `management_events`, `read_only_events`,
  `write_only_events`, `console_logins`.
