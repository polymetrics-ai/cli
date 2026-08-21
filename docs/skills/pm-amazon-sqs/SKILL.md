---
name: pm-amazon-sqs
description: Amazon SQS connector knowledge and safe action guide.
---

# pm-amazon-sqs

## Purpose

Reads Amazon SQS queues and executes typed, approval-gated SQS message and queue actions through fixed AWS Query API operations.

## Icon

- id: amazon-sqs
- asset: icons/amazon-sqs.svg
- source: upstream_registry
- review_status: upstream_seeded

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: queue

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- queue_url
- region
- endpoint_url
- max_batch_size
- max_wait_time
- visibility_timeout
- attributes_to_return
- system_attributes_to_return
- max_polls
- mode
- access_key (secret)
- secret_key (secret)
- session_token (secret)

## ETL Streams

- messages: Messages received from the configured SQS queue. The connector does not delete messages.
  - primary key: message_id
  - fields: message_id(string), md5_of_body(string), receipt_handle(string), body(object), sent_timestamp(string), approximate_receive_count(string), sender_id(string), sequence_number(string), message_group_id(string), message_deduplication_id(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- add_permission:
  - endpoint: POST SQS.AddPermission
  - required fields: label, aws_account_ids, actions
  - risk: adds an SQS queue resource policy permission statement for listed AWS account ids
- cancel_message_move_task:
  - endpoint: POST SQS.CancelMessageMoveTask
  - required fields: task_handle
  - risk: cancels an in-flight dead-letter-queue message move task
- change_message_visibility:
  - endpoint: POST SQS.ChangeMessageVisibility
  - required fields: receipt_handle, visibility_timeout
  - risk: changes the visibility timeout for one in-flight message
- change_message_visibility_batch:
  - endpoint: POST SQS.ChangeMessageVisibilityBatch
  - required fields: receipt_handle, visibility_timeout
  - optional fields: id
  - risk: changes visibility timeout for up to 10 in-flight messages per SQS batch request
- create_queue:
  - endpoint: POST SQS.CreateQueue
  - required fields: queue_name
  - optional fields: attributes, tags
  - risk: creates an SQS queue; SQS returns an existing queue URL when name and attributes match
- delete_message:
  - endpoint: POST SQS.DeleteMessage
  - required fields: receipt_handle
  - risk: deletes one received message by receipt handle
- delete_message_batch:
  - endpoint: POST SQS.DeleteMessageBatch
  - required fields: receipt_handle
  - optional fields: id
  - risk: deletes up to 10 received messages per SQS batch request
- delete_queue:
  - endpoint: POST SQS.DeleteQueue
  - risk: deletes the configured SQS queue
- purge_queue:
  - endpoint: POST SQS.PurgeQueue
  - risk: purges all available messages from the configured queue
- remove_permission:
  - endpoint: POST SQS.RemovePermission
  - required fields: label
  - risk: removes an SQS queue resource policy permission statement
- send_message:
  - endpoint: POST SQS.SendMessage
  - required fields: message_body
  - optional fields: delay_seconds, message_attributes, message_system_attributes, message_deduplication_id, message_group_id
  - risk: sends one message to the configured queue; FIFO queues may use message_deduplication_id for provider-supported idempotency
- send_message_batch:
  - endpoint: POST SQS.SendMessageBatch
  - required fields: message_body
  - optional fields: id, delay_seconds, message_attributes, message_system_attributes, message_deduplication_id, message_group_id
  - risk: sends up to 10 messages per SQS batch request; FIFO queues may use message_deduplication_id for provider-supported idempotency
- set_queue_attributes:
  - endpoint: POST SQS.SetQueueAttributes
  - required fields: attributes or attribute_name + attribute_value
  - risk: sets typed SQS queue attributes such as policy, redrive, encryption, retention, and visibility settings
- start_message_move_task:
  - endpoint: POST SQS.StartMessageMoveTask
  - required fields: source_arn
  - optional fields: destination_arn, max_number_of_messages_per_second
  - risk: starts an SQS dead-letter queue redrive message move task
- tag_queue:
  - endpoint: POST SQS.TagQueue
  - required fields: tags or tag_key + tag_value
  - risk: adds or updates tags on the configured SQS queue
- untag_queue:
  - endpoint: POST SQS.UntagQueue
  - required fields: tag_keys
  - risk: removes tags from the configured SQS queue

## Security

- read risk: signed, bounded Amazon SQS Query API reads against the configured queue or derived regional endpoint; ReceiveMessage can change message visibility per SQS semantics
- write risk: typed SendMessage, queue administration, permission, tag, visibility, delete, purge, and dead-letter redrive operations against fixed SQS targets only
- approval: all reverse ETL/provider write commands require plan, preview, explicit approval, and destructive confirmation for delete/purge/cancel/remove operations
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Inspect and safely operate Amazon SQS queues through fixed typed commands.
- Usage: pm amazon-sqs <command> [options]
- Source CLI: AWS CLI sqs (https://docs.aws.amazon.com/AWSSimpleQueueService/latest/APIReference/API_Operations.html)
- Global flags:
  - --approval-token-stdin (boolean): Read the approval token as one bounded line from standard input.
- Read
- Typed write/admin actions
- Other Commands
  - messages receive - Read bounded messages from the configured queue without deleting them. [intent=etl availability=implemented stream=messages]
  - queue attributes - Read selected attributes for the configured queue. [intent=direct_read availability=implemented operation=get_queue_attributes]; flags: --attribute-names (string_array): Attribute names to request; defaults to All.: maps_to=body.attribute_names, --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - queue url - Resolve a queue URL by queue name and optional owner account id. [intent=direct_read availability=implemented operation=get_queue_url]; flags: --queue-name (required, non-empty) (string): Queue name to resolve.: maps_to=body.queue_name, --queue-owner-aws-account-id (non-empty) (string): Optional owner AWS account id.: maps_to=body.queue_owner_aws_account_id, --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - queues list - List queue URLs in the configured region endpoint. [intent=direct_read availability=implemented operation=list_queues]; flags: --queue-name-prefix (non-empty) (string): Optional queue name prefix.: maps_to=body.queue_name_prefix, --max-results (integer): Maximum queue URLs to return (bounded by SQS).: maps_to=body.max_results, --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - queue tags - Read tags for the configured queue. [intent=direct_read availability=implemented operation=list_queue_tags]; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - dead-letter-source-queues list - List queues configured with the current queue as a dead-letter queue. [intent=direct_read availability=implemented operation=list_dead_letter_source_queues]; flags: --max-results (integer): Maximum queue URLs to return (bounded by SQS).: maps_to=body.max_results, --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - message-move-tasks list - List recent dead-letter message move tasks for a source queue ARN. [intent=direct_read availability=implemented operation=list_message_move_tasks]; flags: --source-arn (required, non-empty) (string): Source queue ARN.: maps_to=body.source_arn, --max-results (integer): Maximum tasks to return (bounded by SQS).: maps_to=body.max_results, --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - permission add - Plan a typed Amazon SQS add permission operation. [intent=reverse_etl availability=implemented write=add_permission]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: adds an SQS queue resource policy permission statement for listed AWS account ids; flags: --label (required, non-empty) (string): Permission statement label.: maps_to=record.label, --aws-account-ids (required) (string_array): AWS account ids to grant.: maps_to=record.aws_account_ids, --actions (required) (string_array): SQS action names to allow.: maps_to=record.actions
  - message-move-task cancel - Plan a typed Amazon SQS cancel message move task operation. [intent=reverse_etl availability=implemented write=cancel_message_move_task]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: cancels an in-flight dead-letter-queue message move task; flags: --task-handle (required, non-empty) (string): Task handle returned by StartMessageMoveTask/ListMessageMoveTasks.: maps_to=record.task_handle
  - message change-visibility - Plan a typed Amazon SQS change message visibility operation. [intent=reverse_etl availability=implemented write=change_message_visibility]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: changes the visibility timeout for one in-flight message; flags: --receipt-handle (required, non-empty) (string): Receipt handle from ReceiveMessage.: maps_to=record.receipt_handle, --visibility-timeout (required) (integer): New visibility timeout in seconds.: maps_to=record.visibility_timeout
  - message change-visibility-batch - Plan a typed Amazon SQS change message visibility batch operation. [intent=reverse_etl availability=implemented write=change_message_visibility_batch]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: changes visibility timeout for up to 10 in-flight messages per SQS batch request; flags: --id (non-empty) (string): Optional batch entry id.: maps_to=record.id, --receipt-handle (required, non-empty) (string): Receipt handle from ReceiveMessage.: maps_to=record.receipt_handle, --visibility-timeout (required) (integer): New visibility timeout in seconds.: maps_to=record.visibility_timeout
  - queue create - Plan a typed Amazon SQS create queue operation. [intent=reverse_etl availability=implemented write=create_queue]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: creates an SQS queue; SQS returns an existing queue URL when name and attributes match; flags: --queue-name (required, non-empty) (string): Queue name to create.: maps_to=record.queue_name
  - message delete - Plan a typed Amazon SQS delete message operation. [intent=reverse_etl availability=implemented write=delete_message]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: deletes one received message by receipt handle; flags: --receipt-handle (required, non-empty) (string): Receipt handle from ReceiveMessage.: maps_to=record.receipt_handle
  - message delete-batch - Plan a typed Amazon SQS delete message batch operation. [intent=reverse_etl availability=implemented write=delete_message_batch]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: deletes up to 10 received messages per SQS batch request; flags: --id (non-empty) (string): Optional batch entry id.: maps_to=record.id, --receipt-handle (required, non-empty) (string): Receipt handle from ReceiveMessage.: maps_to=record.receipt_handle
  - queue delete - Plan a typed Amazon SQS delete queue operation. [intent=reverse_etl availability=implemented write=delete_queue]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: deletes the configured SQS queue
  - queue purge - Plan a typed Amazon SQS purge queue operation. [intent=reverse_etl availability=implemented write=purge_queue]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: purges all available messages from the configured queue
  - permission remove - Plan a typed Amazon SQS remove permission operation. [intent=reverse_etl availability=implemented write=remove_permission]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: removes an SQS queue resource policy permission statement; flags: --label (required, non-empty) (string): Permission statement label to remove.: maps_to=record.label
  - message send - Plan a typed Amazon SQS send message operation. [intent=reverse_etl availability=implemented write=send_message]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: sends one message to the configured queue; FIFO queues may use message_deduplication_id for provider-supported idempotency; flags: --message-body (required, non-empty) (string): Message body string.: maps_to=record.message_body, --delay-seconds (integer): Optional per-message delay seconds.: maps_to=record.delay_seconds, --message-deduplication-id (non-empty) (string): FIFO deduplication id.: maps_to=record.message_deduplication_id, --message-group-id (non-empty) (string): FIFO message group id.: maps_to=record.message_group_id
  - message send-batch - Plan a typed Amazon SQS send message batch operation. [intent=reverse_etl availability=implemented write=send_message_batch]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: sends up to 10 messages per SQS batch request; FIFO queues may use message_deduplication_id for provider-supported idempotency; flags: --id (non-empty) (string): Optional batch entry id.: maps_to=record.id, --message-body (required, non-empty) (string): Message body string.: maps_to=record.message_body, --delay-seconds (integer): Optional per-message delay seconds.: maps_to=record.delay_seconds, --message-deduplication-id (non-empty) (string): FIFO deduplication id.: maps_to=record.message_deduplication_id, --message-group-id (non-empty) (string): FIFO message group id.: maps_to=record.message_group_id
  - queue set-attributes - Plan a typed Amazon SQS set queue attributes operation. [intent=reverse_etl availability=implemented write=set_queue_attributes]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: sets typed SQS queue attributes such as policy, redrive, encryption, retention, and visibility settings; flags: --attribute-name (required, non-empty) (string): Queue attribute name to set, for example VisibilityTimeout or Policy.: maps_to=record.attribute_name, --attribute-value (required, non-empty) (string): Queue attribute value. Sensitive values are redacted in previews/errors.: maps_to=record.attribute_value
  - message-move-task start - Plan a typed Amazon SQS start message move task operation. [intent=reverse_etl availability=implemented write=start_message_move_task]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: starts an SQS dead-letter queue redrive message move task; flags: --source-arn (required, non-empty) (string): Source dead-letter queue ARN.: maps_to=record.source_arn, --destination-arn (non-empty) (string): Optional destination queue ARN.: maps_to=record.destination_arn, --max-number-of-messages-per-second (integer): Optional move rate limit.: maps_to=record.max_number_of_messages_per_second
  - queue tag - Plan a typed Amazon SQS tag queue operation. [intent=reverse_etl availability=implemented write=tag_queue]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: adds or updates tags on the configured SQS queue; flags: --tag-key (required, non-empty) (string): Tag key to add or update.: maps_to=record.tag_key, --tag-value (required, non-empty) (string): Tag value to add or update.: maps_to=record.tag_value
  - queue untag - Plan a typed Amazon SQS untag queue operation. [intent=reverse_etl availability=implemented write=untag_queue]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: removes tags from the configured SQS queue; flags: --tag-keys (required) (string_array): Tag keys to remove.: maps_to=record.tag_keys
- Help topics:
  - safety - Amazon SQS write commands always use reverse ETL plan, preview, approval, and execute; destructive commands require typed confirmation.

## Commands

### Inspect as a manual

```bash
pm connectors inspect amazon-sqs
```

### Inspect as structured JSON

```bash
pm connectors inspect amazon-sqs --json
```

## Agent Rules

- Run pm connectors inspect amazon-sqs before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
