# pm connectors inspect mailosaur

```text
NAME
  pm connectors inspect mailosaur - Mailosaur connector manual

SYNOPSIS
  pm connectors inspect mailosaur
  pm connectors inspect mailosaur --json
  pm credentials add <name> --connector mailosaur [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Mailosaur virtual servers, message summaries, and account usage transactions through the Mailosaur REST API.

ICON
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  items_per_page
  mode
  received_after
  server
  username
  password (secret)

ETL STREAMS
  servers:
    primary key: id
    fields: id(string), messages(integer), name(string), users(array)
  messages:
    primary key: id
    cursor: received
    fields: bcc(array), cc(array), from(array), id(string), received(string), server(string), subject(string), to(array), type(string)
  transactions:
    primary key: timestamp
    cursor: timestamp
    fields: email(integer), previews(integer), sms(integer), timestamp(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Mailosaur API read of virtual-server, message-summary, and usage-transaction data
  approval: none; read-only source connector
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run Mailosaur's declared streams and reverse-ETL actions.
  Usage: pm mailosaur <command> [flags]
  Read streams
  Reverse ETL writes
  Other Commands
    api delete api devices id - Documented DELETE /api/devices/:id (not implemented) [intent=direct_write availability=not_implemented operation=mailosaur.delete.api-devices-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete api messages - Documented DELETE /api/messages (not implemented) [intent=direct_write availability=not_implemented operation=mailosaur.delete.api-messages]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete api messages id - Documented DELETE /api/messages/:id (not implemented) [intent=direct_write availability=not_implemented operation=mailosaur.delete.api-messages-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete api servers id - Documented DELETE /api/servers/:id (not implemented) [intent=direct_write availability=not_implemented operation=mailosaur.delete.api-servers-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete messages id - Documented DELETE /messages/{id} (not implemented) [intent=direct_write availability=not_implemented operation=mailosaur.delete.messages-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete servers id - Documented DELETE /servers/{id} (not implemented) [intent=direct_write availability=not_implemented operation=mailosaur.delete.servers-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api get analysis spam id - Documented GET /analysis/spam/{id} (not implemented) [intent=direct_read availability=not_implemented operation=mailosaur.get.analysis-spam-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api analysis deliverability id - Documented GET /api/analysis/deliverability/:id (not implemented) [intent=direct_read availability=not_implemented operation=mailosaur.get.api-analysis-deliverability-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api devices - Documented GET /api/devices (not implemented) [intent=direct_read availability=not_implemented operation=mailosaur.get.api-devices]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api devices id otp - Documented GET /api/devices/:id/otp (not implemented) [intent=direct_read availability=not_implemented operation=mailosaur.get.api-devices-id-otp]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api files attachments id - Documented GET /api/files/attachments/:id (not implemented) [intent=direct_read availability=not_implemented operation=mailosaur.get.api-files-attachments-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api files email id - Documented GET /api/files/email/:id (not implemented) [intent=direct_read availability=not_implemented operation=mailosaur.get.api-files-email-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api messages - Documented GET /api/messages (not implemented) [intent=direct_read availability=not_implemented operation=mailosaur.get.api-messages]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api messages id - Documented GET /api/messages/:id (not implemented) [intent=direct_read availability=not_implemented operation=mailosaur.get.api-messages-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api servers - Documented GET /api/servers (not implemented) [intent=direct_read availability=not_implemented operation=mailosaur.get.api-servers]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api servers id - Documented GET /api/servers/:id (not implemented) [intent=direct_read availability=not_implemented operation=mailosaur.get.api-servers-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api servers id password - Documented GET /api/servers/:id/password (not implemented) [intent=direct_read availability=not_implemented operation=mailosaur.get.api-servers-id-password]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api usage limits - Documented GET /api/usage/limits (not implemented) [intent=direct_read availability=not_implemented operation=mailosaur.get.api-usage-limits]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api usage transactions - Documented GET /api/usage/transactions (not implemented) [intent=direct_read availability=not_implemented operation=mailosaur.get.api-usage-transactions]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get messages id - Documented GET /messages/{id} (not implemented) [intent=direct_read availability=not_implemented operation=mailosaur.get.messages-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get usage limits - Documented GET /usage/limits (not implemented) [intent=direct_read availability=not_implemented operation=mailosaur.get.usage-limits]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api post api devices - Documented POST /api/devices (not implemented) [intent=direct_write availability=not_implemented operation=mailosaur.post.api-devices]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api devices otp - Documented POST /api/devices/otp (not implemented) [intent=direct_write availability=not_implemented operation=mailosaur.post.api-devices-otp]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api messages - Documented POST /api/messages (not implemented) [intent=direct_write availability=not_implemented operation=mailosaur.post.api-messages]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api messages id forward - Documented POST /api/messages/:id/forward (not implemented) [intent=direct_write availability=not_implemented operation=mailosaur.post.api-messages-id-forward]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api messages id reply - Documented POST /api/messages/:id/reply (not implemented) [intent=direct_write availability=not_implemented operation=mailosaur.post.api-messages-id-reply]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api messages search - Documented POST /api/messages/search (not implemented) [intent=direct_write availability=not_implemented operation=mailosaur.post.api-messages-search]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api servers - Documented POST /api/servers (not implemented) [intent=direct_write availability=not_implemented operation=mailosaur.post.api-servers]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post servers - Documented POST /servers (not implemented) [intent=direct_write availability=not_implemented operation=mailosaur.post.servers]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put api servers id - Documented PUT /api/servers/:id (not implemented) [intent=direct_write availability=not_implemented operation=mailosaur.put.api-servers-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    messages list - Run the messages ETL stream [intent=etl availability=implemented stream=messages]; notes: discrepancy=present-in-surface-absent-from-artifact
    servers list - Run the servers ETL stream [intent=etl availability=implemented stream=servers]; notes: discrepancy=present-in-surface-absent-from-artifact
    transactions list - Run the transactions ETL stream [intent=etl availability=implemented stream=transactions]; notes: discrepancy=present-in-surface-absent-from-artifact

EXAMPLES
  # Inspect as a manual
  pm connectors inspect mailosaur

  # Inspect as structured JSON
  pm connectors inspect mailosaur --json

AGENT WORKFLOW
  - Run pm connectors inspect mailosaur before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
