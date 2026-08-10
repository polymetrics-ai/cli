# pm connectors inspect tyntec-sms

```text
NAME
  pm connectors inspect tyntec-sms - tyntec SMS connector manual

SYNOPSIS
  pm connectors inspect tyntec-sms
  pm connectors inspect tyntec-sms --json
  pm credentials add <name> --connector tyntec-sms [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads tyntec SMS messages, templates, sender IDs, and delivery reports through API list endpoints, and sends approved SMS messages through the Messaging API.

ICON
  id: tyntec
  asset: icons/tyntec.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://api.tyntec.com/reference/messaging

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  api_key (secret)

ETL STREAMS
  messages:
    primary key: id
    cursor: created_at
    fields: created_at(string), from(string), id(string), status(string), to(string)
  templates:
    primary key: id
    fields: id(string), name(string)
  sender_ids:
    primary key: id
    fields: id(string), name(string)
  delivery_reports:
    primary key: id
    cursor: created_at
    fields: created_at(string), from(string), id(string), status(string), to(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  send_message:
    endpoint: POST sms/v1/messages
    required fields: to, from, text
    risk: sends a billable SMS message to the recipient phone number and may notify an external user

SECURITY
  read risk: external tyntec SMS API read of messages, templates, sender IDs, and delivery reports
  write risk: sends billable SMS messages to recipient phone numbers; approval required before delivery
  approval: reverse ETL plan approval required before writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run tyntec SMS's declared streams and reverse-ETL actions.
  Usage: pm tyntec-sms <command> [flags]
  Read streams
  Reverse ETL writes
  Other Commands
    api delete byon contacts v1 contactid - Documented DELETE /byon/contacts/v1/{contactId} (not implemented) [intent=direct_write availability=not_implemented operation=tyntec-sms.delete.byon-contacts-v1-contactid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api get byon contacts v1 - Documented GET /byon/contacts/v1 (not implemented) [intent=direct_read availability=not_implemented operation=tyntec-sms.get.byon-contacts-v1]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get byon contacts v1 contactid - Documented GET /byon/contacts/v1/{contactId} (not implemented) [intent=direct_read availability=not_implemented operation=tyntec-sms.get.byon-contacts-v1-contactid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get byon phonebook v1 numbers - Documented GET /byon/phonebook/v1/numbers (not implemented) [intent=direct_read availability=not_implemented operation=tyntec-sms.get.byon-phonebook-v1-numbers]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get byon phonebook v1 numbers phonenumber - Documented GET /byon/phonebook/v1/numbers/{phoneNumber} (not implemented) [intent=direct_read availability=not_implemented operation=tyntec-sms.get.byon-phonebook-v1-numbers-phonenumber]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get byon provisioning v1 - Documented GET /byon/provisioning/v1 (not implemented) [intent=direct_read availability=not_implemented operation=tyntec-sms.get.byon-provisioning-v1]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get byon provisioning v1 requestid - Documented GET /byon/provisioning/v1/{requestId} (not implemented) [intent=direct_read availability=not_implemented operation=tyntec-sms.get.byon-provisioning-v1-requestid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get messaging v1 messages requestid - Documented GET /messaging/v1/messages/{requestId} (not implemented) [intent=direct_read availability=not_implemented operation=tyntec-sms.get.messaging-v1-messages-requestid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get messaging v1 sms - Documented GET /messaging/v1/sms (not implemented) [intent=direct_read availability=not_implemented operation=tyntec-sms.get.messaging-v1-sms]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api post byon contacts v1 - Documented POST /byon/contacts/v1 (not implemented) [intent=direct_write availability=not_implemented operation=tyntec-sms.post.byon-contacts-v1]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post byon contacts v1 contactid - Documented POST /byon/contacts/v1/{contactId} (not implemented) [intent=direct_write availability=not_implemented operation=tyntec-sms.post.byon-contacts-v1-contactid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post byon provisioning v1 - Documented POST /byon/provisioning/v1 (not implemented) [intent=direct_write availability=not_implemented operation=tyntec-sms.post.byon-provisioning-v1]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post incoming - Documented POST /incoming (not implemented) [intent=direct_write availability=not_implemented operation=tyntec-sms.post.incoming]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post messaging v1 conversion - Documented POST /messaging/v1/conversion (not implemented) [intent=direct_write availability=not_implemented operation=tyntec-sms.post.messaging-v1-conversion]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post messaging v1 sms - Documented POST /messaging/v1/sms (not implemented) [intent=direct_write availability=not_implemented operation=tyntec-sms.post.messaging-v1-sms]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    delivery reports list - Run the delivery reports ETL stream [intent=etl availability=implemented stream=delivery_reports]; notes: discrepancy=present-in-surface-absent-from-artifact
    messages list - Run the messages ETL stream [intent=etl availability=implemented stream=messages]; notes: discrepancy=present-in-surface-absent-from-artifact
    send message apply - Plan and execute the send message reverse-ETL action [intent=reverse_etl availability=implemented write=send_message]; approval: requires plan, preview, approval, and execute; risk: sends a billable SMS message to the recipient phone number and may notify an external user; flags: --from (required), --text (required), --to (required)
    sender ids list - Run the sender ids ETL stream [intent=etl availability=implemented stream=sender_ids]; notes: discrepancy=present-in-surface-absent-from-artifact
    templates list - Run the templates ETL stream [intent=etl availability=implemented stream=templates]; notes: discrepancy=present-in-surface-absent-from-artifact

EXAMPLES
  # Inspect as a manual
  pm connectors inspect tyntec-sms

  # Inspect as structured JSON
  pm connectors inspect tyntec-sms --json

AGENT WORKFLOW
  - Run pm connectors inspect tyntec-sms before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
