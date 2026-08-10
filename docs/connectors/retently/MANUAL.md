# pm connectors inspect retently

```text
NAME
  pm connectors inspect retently - Retently connector manual

SYNOPSIS
  pm connectors inspect retently
  pm connectors inspect retently --json
  pm credentials add <name> --connector retently [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Retently customers, survey responses, surveys, and campaigns through the REST API.

ICON
  id: retently
  asset: icons/retently.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://www.retently.com/api/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  campaign_id
  created_after
  email
  updated_after
  api_key (secret)

ETL STREAMS
  customers:
    primary key: id
    cursor: updated_at
    fields: company(string), email(string), full_name(string), id(string), stream(string), updated_at(string)
  responses:
    primary key: id
    cursor: created_at
    fields: comment(string), created_at(string), customer_id(string), id(string), score(string), stream(string)
  surveys:
    primary key: id
    cursor: updated_at
    fields: id(string), name(string), status(string), stream(string), type(string), updated_at(string)
  campaigns:
    primary key: id
    cursor: updated_at
    fields: id(string), name(string), status(string), stream(string), updated_at(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Retently API read of customer and NPS/CSAT survey response data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run Retently's declared streams and reverse-ETL actions.
  Usage: pm retently <command> [flags]
  Read streams
  Reverse ETL writes
  Other Commands
    api delete api v2 customers - Documented DELETE /api/v2/customers (not implemented) [intent=direct_write availability=not_implemented operation=retently.delete.api-v2-customers]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete api v2 feedback - Documented DELETE /api/v2/feedback (not implemented) [intent=direct_write availability=not_implemented operation=retently.delete.api-v2-feedback]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete api v2 outbox - Documented DELETE /api/v2/outbox (not implemented) [intent=direct_write availability=not_implemented operation=retently.delete.api-v2-outbox]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete api v2 suppressions domains id - Documented DELETE /api/v2/suppressions/domains/{id} (not implemented) [intent=direct_write availability=not_implemented operation=retently.delete.api-v2-suppressions-domains-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete api v2 suppressions emails by-email - Documented DELETE /api/v2/suppressions/emails/by-email (not implemented) [intent=direct_write availability=not_implemented operation=retently.delete.api-v2-suppressions-emails-by-email]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete api v2 suppressions emails id - Documented DELETE /api/v2/suppressions/emails/{id} (not implemented) [intent=direct_write availability=not_implemented operation=retently.delete.api-v2-suppressions-emails-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api get api v2 companies - Documented GET /api/v2/companies{ (not implemented) [intent=direct_read availability=not_implemented operation=retently.get.api-v2-companies]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v2 companies companyid - Documented GET /api/v2/companies/{companyId (not implemented) [intent=direct_read availability=not_implemented operation=retently.get.api-v2-companies-companyid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v2 customers - Documented GET /api/v2/customers{ (not implemented) [intent=direct_read availability=not_implemented operation=retently.get.api-v2-customers]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v2 customers customerid - Documented GET /api/v2/customers/{customerId} (not implemented) [intent=direct_read availability=not_implemented operation=retently.get.api-v2-customers-customerid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v2 feedback - Documented GET /api/v2/feedback (not implemented) [intent=direct_read availability=not_implemented operation=retently.get.api-v2-feedback]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v2 feedback feedbackid - Documented GET /api/v2/feedback/{feedbackId} (not implemented) [intent=direct_read availability=not_implemented operation=retently.get.api-v2-feedback-feedbackid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v2 feedback operation - Documented GET /api/v2/feedback{ (not implemented) [intent=direct_read availability=not_implemented operation=retently.get.api-v2-feedback-2]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v2 metric score - Documented GET /api/v2/{metric}/score (not implemented) [intent=direct_read availability=not_implemented operation=retently.get.api-v2-metric-score]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v2 nps campaigns - Documented GET /api/v2/nps/campaigns (not implemented) [intent=direct_read availability=not_implemented operation=retently.get.api-v2-nps-campaigns]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v2 nps customers response - Documented GET /api/v2/nps/customers/response{ (not implemented) [intent=direct_read availability=not_implemented operation=retently.get.api-v2-nps-customers-response]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v2 nps templates - Documented GET /api/v2/nps/templates (not implemented) [intent=direct_read availability=not_implemented operation=retently.get.api-v2-nps-templates]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v2 outbox - Documented GET /api/v2/outbox (not implemented) [intent=direct_read availability=not_implemented operation=retently.get.api-v2-outbox]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v2 outbox operation - Documented GET /api/v2/outbox{ (not implemented) [intent=direct_read availability=not_implemented operation=retently.get.api-v2-outbox-2]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v2 ping - Documented GET /api/v2/ping (not implemented) [intent=direct_read availability=not_implemented operation=retently.get.api-v2-ping]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v2 reports - Documented GET /api/v2/reports/ (not implemented) [intent=direct_read availability=not_implemented operation=retently.get.api-v2-reports]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v2 suppressions domains - Documented GET /api/v2/suppressions/domains (not implemented) [intent=direct_read availability=not_implemented operation=retently.get.api-v2-suppressions-domains]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v2 suppressions emails - Documented GET /api/v2/suppressions/emails (not implemented) [intent=direct_read availability=not_implemented operation=retently.get.api-v2-suppressions-emails]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v2 templates - Documented GET /api/v2/templates (not implemented) [intent=direct_read availability=not_implemented operation=retently.get.api-v2-templates]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v2 templates templateid - Documented GET /api/v2/templates/{templateId} (not implemented) [intent=direct_read availability=not_implemented operation=retently.get.api-v2-templates-templateid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v2 templates templateid operation - Documented GET /api/v2/templates/{templateId}. (not implemented) [intent=direct_read availability=not_implemented operation=retently.get.api-v2-templates-templateid-2]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v2 trends - Documented GET /api/v2/trends (not implemented) [intent=direct_read availability=not_implemented operation=retently.get.api-v2-trends]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v2 trends groupid - Documented GET /api/v2/trends/:groupId (not implemented) [intent=direct_read availability=not_implemented operation=retently.get.api-v2-trends-groupid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api post api v2 customers - Documented POST /api/v2/customers (not implemented) [intent=direct_write availability=not_implemented operation=retently.post.api-v2-customers]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api v2 customers resubscribe - Documented POST /api/v2/customers/resubscribe (not implemented) [intent=direct_write availability=not_implemented operation=retently.post.api-v2-customers-resubscribe]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api v2 customers unsubscribe - Documented POST /api/v2/customers/unsubscribe (not implemented) [intent=direct_write availability=not_implemented operation=retently.post.api-v2-customers-unsubscribe]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api v2 import - Documented POST /api/v2/import (not implemented) [intent=direct_write availability=not_implemented operation=retently.post.api-v2-import]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api v2 nps customers response tags - Documented POST /api/v2/nps/customers/response/tags (not implemented) [intent=direct_write availability=not_implemented operation=retently.post.api-v2-nps-customers-response-tags]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api v2 nps customers survey - Documented POST /api/v2/nps/customers/survey (not implemented) [intent=direct_write availability=not_implemented operation=retently.post.api-v2-nps-customers-survey]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api v2 response tags - Documented POST /api/v2/response/tags (not implemented) [intent=direct_write availability=not_implemented operation=retently.post.api-v2-response-tags]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api v2 response topics - Documented POST /api/v2/response/topics (not implemented) [intent=direct_write availability=not_implemented operation=retently.post.api-v2-response-topics]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api v2 suppressions domains - Documented POST /api/v2/suppressions/domains (not implemented) [intent=direct_write availability=not_implemented operation=retently.post.api-v2-suppressions-domains]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api v2 suppressions emails - Documented POST /api/v2/suppressions/emails (not implemented) [intent=direct_write availability=not_implemented operation=retently.post.api-v2-suppressions-emails]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api v2 survey - Documented POST /api/v2/survey (not implemented) [intent=direct_write availability=not_implemented operation=retently.post.api-v2-survey]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api v2 sync-attributes - Documented POST /api/v2/sync-attributes (not implemented) [intent=direct_write availability=not_implemented operation=retently.post.api-v2-sync-attributes]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    campaigns list - Run the campaigns ETL stream [intent=etl availability=implemented stream=campaigns]
    customers list - Run the customers ETL stream [intent=etl availability=implemented stream=customers]; notes: discrepancy=present-in-surface-absent-from-artifact
    responses list - Run the responses ETL stream [intent=etl availability=implemented stream=responses]; notes: discrepancy=present-in-surface-absent-from-artifact
    surveys list - Run the surveys ETL stream [intent=etl availability=implemented stream=surveys]; notes: discrepancy=present-in-surface-absent-from-artifact

EXAMPLES
  # Inspect as a manual
  pm connectors inspect retently

  # Inspect as structured JSON
  pm connectors inspect retently --json

AGENT WORKFLOW
  - Run pm connectors inspect retently before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
