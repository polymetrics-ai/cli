# Overview

Reads messages from Amazon SQS via signed ReceiveMessage calls. Read-only; messages are not deleted.

This connector discovers available streams and schemas from the configured service at runtime.

This connector is read-only; no write actions are declared.

Service API documentation:
https://docs.aws.amazon.com/AWSSimpleQueueService/latest/APIReference/Welcome.html.

## Auth setup

Connection fields:

- `access_key` (required, secret, string); AWS access key id used for SigV4 request signing. Never
  logged.
- `attributes_to_return` (optional, string); default `All`; Comma-separated list of
  MessageAttributeName values to request. Defaults to All.
- `max_batch_size` (optional, string); default `10`; Messages requested per ReceiveMessage call
  (MaxNumberOfMessages), clamped to 1-10.
- `max_polls` (optional, string); default `1`; Maximum number of ReceiveMessage calls per Read,
  clamped to 1-100. A poll returning zero messages stops early.
- `max_wait_time` (optional, string); default `0`; SQS long-poll WaitTimeSeconds, clamped to 0-20.
- `mode` (optional, string).
- `queue_url` (required, string); format `uri`; Full SQS queue URL (e.g.
  https://sqs.us-east-1.amazonaws.com/123456789012/my-queue).
- `region` (required, string); AWS region the queue lives in (e.g. us-east-1), used for SigV4
  request signing.
- `secret_key` (required, secret, string); AWS secret access key used for SigV4 request signing.
  Never logged.
- `session_token` (optional, secret, string); Optional AWS session token for temporary/STS
  credentials, sent as X-Amz-Security-Token when set.
- `visibility_timeout` (optional, string); Optional VisibilityTimeout override in seconds, clamped
  to 0-43200. Uses the queue's own default when unset.

Secret fields are redacted in logs and write previews: `access_key`, `secret_key`, `session_token`.

Default configuration values: `attributes_to_return=All`, `max_batch_size=10`, `max_polls=1`,
`max_wait_time=0`.

Provide the secret fields listed above. Authentication is applied by the connector-specific
implementation for this service.

## Streams notes

The connector discovers catalogs and records directly from the configured service instead of using
fixed stream declarations.

## Write actions & risks

This connector is read-only. Read behavior: external AWS SQS queue read via signed ReceiveMessage
calls (visibility timeout side effects per SQS semantics).

## Known limits

- Schemas and stream availability depend on the configured service at runtime.
