# pm connectors inspect amazon-sqs

```text
NAME
  pm connectors inspect amazon-sqs - Amazon SQS connector manual

SYNOPSIS
  pm connectors inspect amazon-sqs
  pm connectors inspect amazon-sqs --json
  pm credentials add <name> --connector amazon-sqs [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Retains evidence-backed visibility for DeleteQueue and PurgeQueue while their configured-queue actions lack a schema-4 typed record binding. No Amazon SQS provider action is executable from this bundle.

ICON
  id: amazon-sqs
  asset: icons/amazon-sqs.svg
  source: upstream_registry
  review_status: upstream_seeded

CAPABILITIES
  check=false catalog=true read=false write=false query=false
  Integration type: queue

AUTHENTICATION
  No secret authentication is required for this connector.

CONFIGURATION
  queue_url

ETL STREAMS
  messages:
    primary key: message_id
    fields: approximate_receive_count(string), body(), fixture(boolean), md5_of_body(string), message_deduplication_id(string), message_group_id(string), message_id(string), receipt_handle(string), sender_id(string), sent_timestamp(string), sequence_number(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

REVERSE ETL ACTIONS
  add_permission:
    endpoint: POST SQS.AddPermission
    required fields: label, aws_account_ids, actions
    risk: adds an SQS queue resource policy permission statement for listed AWS account ids
  cancel_message_move_task:
    endpoint: POST SQS.CancelMessageMoveTask
    required fields: task_handle
    risk: cancels an in-flight dead-letter-queue message move task
  change_message_visibility:
    endpoint: POST SQS.ChangeMessageVisibility
    required fields: receipt_handle, visibility_timeout
    risk: changes the visibility timeout for one in-flight message
  change_message_visibility_batch:
    endpoint: POST SQS.ChangeMessageVisibilityBatch
    required fields: receipt_handle, visibility_timeout
    risk: changes visibility timeout for up to 10 in-flight messages per SQS batch request
  create_queue:
    endpoint: POST SQS.CreateQueue
    required fields: queue_name
    risk: creates an SQS queue; SQS returns an existing queue URL when name and attributes match
  delete_message:
    endpoint: POST SQS.DeleteMessage
    required fields: receipt_handle
    risk: deletes one received message by receipt handle
  delete_message_batch:
    endpoint: POST SQS.DeleteMessageBatch
    required fields: receipt_handle
    risk: deletes up to 10 received messages per SQS batch request
  delete_queue:
    endpoint: POST SQS.DeleteQueue
    risk: deletes the configured SQS queue
  purge_queue:
    endpoint: POST SQS.PurgeQueue
    risk: purges all available messages from the configured queue
  remove_permission:
    endpoint: POST SQS.RemovePermission
    required fields: label
    risk: removes an SQS queue resource policy permission statement
  send_message:
    endpoint: POST SQS.SendMessage
    required fields: message_body
    risk: sends one message to the configured queue; FIFO queues may use message_deduplication_id for provider-supported idempotency
  send_message_batch:
    endpoint: POST SQS.SendMessageBatch
    required fields: message_body
    risk: sends up to 10 messages per SQS batch request; FIFO queues may use message_deduplication_id for provider-supported idempotency
  set_queue_attributes:
    endpoint: POST SQS.SetQueueAttributes
    risk: sets typed SQS queue attributes such as policy, redrive, encryption, retention, and visibility settings
  start_message_move_task:
    endpoint: POST SQS.StartMessageMoveTask
    required fields: source_arn
    risk: starts an SQS dead-letter queue redrive message move task
  tag_queue:
    endpoint: POST SQS.TagQueue
    risk: adds or updates tags on the configured SQS queue
  untag_queue:
    endpoint: POST SQS.UntagQueue
    required fields: tag_keys
    risk: removes tags from the configured SQS queue

SECURITY
  read risk: No provider read is executable from this bounded disposition bundle.
  write risk: DeleteQueue and PurgeQueue remain visible with evidence-backed unsupported typed-binding gaps; the runtime never sends a request for either command.
  approval: No write can reach planning, approval, or provider I/O from this bundle.
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Amazon SQS queue actions with explicit runtime dispositions.
  Usage: pm amazon-sqs <command>
  Queue actions
  Other Commands
    queue delete - DeleteQueue is visible but unsupported without a schema-4 typed configured-queue binding. [intent=docs_only availability=unsupported_with_provider_evidence]; approval: not executable; risk: deletes the configured SQS queue
    queue purge - PurgeQueue is visible but unsupported without a schema-4 typed configured-queue binding. [intent=docs_only availability=unsupported_with_provider_evidence]; approval: not executable; risk: purges all available messages from the configured queue

EXAMPLES
  # Inspect as a manual
  pm connectors inspect amazon-sqs

  # Inspect as structured JSON
  pm connectors inspect amazon-sqs --json

AGENT WORKFLOW
  - Run pm connectors inspect amazon-sqs before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
