# Overview

Amazon SQS (`amazon-sqs`) is a Tier-3 native connector for the AWS SQS Query API
(API version `2012-11-05`). It signs fixed SQS `Action` requests with AWS SigV4,
decodes XML responses, reads bounded message batches, and exposes typed
approval-gated write/admin actions. It does not expose arbitrary AWS actions,
raw HTTP, raw body/query passthrough, shell, or file escape hatches.

Official sources re-audited for this parity slice:

- AWS-owned Smithy model: https://raw.githubusercontent.com/aws/aws-sdk-go-v2/main/codegen/sdk-codegen/aws-models/sqs.json
- AWS SQS API Reference operation list: https://docs.aws.amazon.com/AWSSimpleQueueService/latest/APIReference/API_Operations.html

The complete operation ledger covers all 23 documented operations: 1 bounded
message stream (`ReceiveMessage`), 6 typed direct-read operations, and 16 typed
reverse-ETL write/admin actions. Fixture-only/local evidence is not a live AWS
certification claim.

## Auth setup

Required connection fields:

- `queue_url` — full SQS queue URL. Queue-specific reads/writes are fixed to
  this target. The URL is canonicalized and must be absolute, contain no
  userinfo/query/fragment, and use `https` except localhost fixture/test
  endpoints. Account-level operations derive the regional endpoint from this
  URL unless `endpoint_url` is set.
- `region` — AWS region used in the SigV4 credential scope.
- `access_key` (secret) — AWS access key id.
- `secret_key` (secret) — AWS secret access key.

Optional fields:

- `session_token` (secret) — STS/session token sent as `X-Amz-Security-Token`.
- `endpoint_url` — AWS-compatible regional endpoint override for local or
  AWS-compatible SQS implementations. Overrides must be root service endpoints
  with no userinfo/query/fragment and must use `https` except localhost
  fixture/test endpoints.
- `max_batch_size` — `ReceiveMessage.MaxNumberOfMessages`, clamped to 1-10.
- `max_wait_time` — `ReceiveMessage.WaitTimeSeconds`, clamped to 0-20.
- `visibility_timeout` — optional `ReceiveMessage.VisibilityTimeout`, clamped
  to 0-43200.
- `attributes_to_return` — comma-separated message attribute names; default
  `All`.
- `system_attributes_to_return` — comma-separated system attribute names;
  default `All`.
- `max_polls` — bounded ReceiveMessage polls per read, clamped to 1-100.
- `mode=fixture` — local fixture mode for credential-free tests; no network.

Secret fields are `access_key`, `secret_key`, and `session_token`. Do not place
secret values in prompt text, docs, fixtures, or issue comments.

## Streams notes

The `messages` stream executes bounded `ReceiveMessage` calls against the
configured queue and emits records shaped like:

- `message_id`
- `md5_of_body`
- `receipt_handle` (redacted by provider command surfaces)
- `body` (JSON-decoded when the message body is JSON; otherwise a string)
- message/system attributes projected to snake_case fields such as
  `sent_timestamp`, `approximate_receive_count`, `message_group_id`, and
  `sequence_number`

Reads do **not** delete messages. Standard SQS receive semantics still apply:
received messages may become temporarily invisible according to the queue or
request visibility timeout. `ReceiveMessage` has no timestamp or offset cursor,
so CDC/changefeed is not advertised; the parent audit counts it as a queue poll
operation, not CDC.

Typed direct-read commands cover the other read operations:

- `queue attributes` -> `GetQueueAttributes`
- `queue url` -> `GetQueueUrl`
- `dead-letter-source-queues list` -> `ListDeadLetterSourceQueues`
- `message-move-tasks list` -> `ListMessageMoveTasks`
- `queues list` -> `ListQueues`
- `queue tags` -> `ListQueueTags`

Every direct read is fixed-target, bounded by the command runner's operation
read byte cap, decoded from XML to JSON, and redacted through the
`json_redacted` output policy equivalent in the native implementation.

## Write actions & risks

All writes require the existing reverse ETL safety path: plan -> preview ->
explicit approval -> execute. Destructive actions also carry
`confirm: "destructive"`, so execution requires the typed destructive
confirmation challenge.

Typed actions:

| Action | Official operation | Confirmation | Risk |
| --- | --- | --- | --- |
| `add_permission` | `AddPermission` | approval | Adds a queue resource policy permission statement. |
| `cancel_message_move_task` | `CancelMessageMoveTask` | destructive | Cancels an active dead-letter redrive/move task. |
| `change_message_visibility` | `ChangeMessageVisibility` | approval | Changes visibility for one in-flight message. |
| `change_message_visibility_batch` | `ChangeMessageVisibilityBatch` | approval | Changes visibility for up to 10 in-flight messages per request. |
| `create_queue` | `CreateQueue` | approval | Creates a queue; SQS is idempotent for matching existing queue name/attributes. |
| `delete_message` | `DeleteMessage` | destructive | Deletes one message by receipt handle. |
| `delete_message_batch` | `DeleteMessageBatch` | destructive | Deletes up to 10 messages per request. |
| `delete_queue` | `DeleteQueue` | destructive | Deletes the configured queue. |
| `purge_queue` | `PurgeQueue` | destructive | Removes available messages from the configured queue. |
| `remove_permission` | `RemovePermission` | destructive | Removes a queue policy statement. |
| `send_message` | `SendMessage` | approval | Sends one message; FIFO queues may use `message_deduplication_id`. |
| `send_message_batch` | `SendMessageBatch` | approval | Sends up to 10 messages per request. |
| `set_queue_attributes` | `SetQueueAttributes` | approval | Mutates queue attributes such as policy, encryption, redrive, and retention settings. |
| `start_message_move_task` | `StartMessageMoveTask` | approval | Starts a dead-letter queue redrive/move task. |
| `tag_queue` | `TagQueue` | approval | Adds or updates queue tags. |
| `untag_queue` | `UntagQueue` | destructive | Removes queue tags. |

Closed schemas are declared in `writes.json`. Receipt handles, message bodies,
message attributes, task handles, and queue attribute maps are redacted in
previews/errors where they may carry sensitive payload or control data.

## Known limits

- This is a Tier-3 native SigV4/XML connector. Declarative dynamic conformance
  replay is explicitly skipped with native httptest fixtures as substitute
  evidence; `api_surface.json`, `fixtures/**`, and native tests are the
  operation-level proof for fixture-only/local validation.
- No live AWS call, credentialed check, certification claim, release claim, or
  provider write was made by this parity implementation.
- Batch write actions chunk records into SQS-supported batches of 10. SQS can
  return per-entry batch failures in a 200 response; the connector reports
  those as failed records.
- Account-level operations such as `CreateQueue`, `GetQueueUrl`, and
  `ListQueues` use the derived regional endpoint from `queue_url` (or
  `endpoint_url` when configured); queue-level operations remain fixed to the
  configured queue target.
- No generic provider search/query or binary file surface is applicable to the
  official SQS operation list.
