# pm connectors inspect freshcaller

```text
NAME
  pm connectors inspect freshcaller - Freshcaller connector manual

SYNOPSIS
  pm connectors inspect freshcaller
  pm connectors inspect freshcaller --json
  pm credentials add <name> --connector freshcaller [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Freshcaller calls, agents, teams, and phone numbers through the Freshcaller REST API.

ICON
  id: freshcaller
  asset: icons/freshcaller.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.freshcaller.com/api/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  max_pages
  mode
  page_size
  api_key (secret)

ETL STREAMS
  calls:
    primary key: id
    cursor: call_time
    fields: agent_id(integer), call_time(string), direction(string), duration(integer), id(integer), phone_number(string), status(string)
  agents:
    primary key: id
    fields: email(string), id(integer), name(string), status(string)
  teams:
    primary key: id
    fields: id(integer), name(string)
  numbers:
    primary key: id
    fields: id(integer), name(string), phone_number(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Freshcaller API read of call, agent, team, and phone number data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run Freshcaller's declared streams and reverse-ETL actions.
  Usage: pm freshcaller <command> [flags]
  Read streams
  Reverse ETL writes
  Other Commands
    agents list - Run the agents ETL stream [intent=etl availability=implemented stream=agents]; notes: discrepancy=present-in-surface-absent-from-artifact
    api delete api v1 calls call-id recording recording-id - Documented DELETE /api/v1/calls/{call_id}/recording/{recording_id} (not implemented) [intent=direct_write availability=not_implemented operation=freshcaller.delete.api-v1-calls-call-id-recording-recording-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api get api v1 business-calendars - Documented GET /api/v1/business_calendars (not implemented) [intent=direct_read availability=not_implemented operation=freshcaller.get.api-v1-business-calendars]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v1 call-metrics - Documented GET /api/v1/call_metrics (not implemented) [intent=direct_read availability=not_implemented operation=freshcaller.get.api-v1-call-metrics]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v1 calls call-id - Documented GET /api/v1/calls/{call_id} (not implemented) [intent=direct_read availability=not_implemented operation=freshcaller.get.api-v1-calls-call-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v1 calls call-id call-metrics - Documented GET /api/v1/calls/{call_id}/call_metrics (not implemented) [intent=direct_read availability=not_implemented operation=freshcaller.get.api-v1-calls-call-id-call-metrics]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v1 calls call-id recording recording-id - Documented GET /api/v1/calls/{call_id}/recording/{recording_id} (not implemented) [intent=direct_read availability=not_implemented operation=freshcaller.get.api-v1-calls-call-id-recording-recording-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v1 ivrs - Documented GET /api/v1/ivrs (not implemented) [intent=direct_read availability=not_implemented operation=freshcaller.get.api-v1-ivrs]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v1 teams team-id - Documented GET /api/v1/teams/{team_id} (not implemented) [intent=direct_read availability=not_implemented operation=freshcaller.get.api-v1-teams-team-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v1 user-statuses - Documented GET /api/v1/user-statuses (not implemented) [intent=direct_read availability=not_implemented operation=freshcaller.get.api-v1-user-statuses]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v1 user-statuses operation - Documented GET /api/v1/user_statuses (not implemented) [intent=direct_read availability=not_implemented operation=freshcaller.get.api-v1-user-statuses-2]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v1 users - Documented GET /api/v1/users (not implemented) [intent=direct_read availability=not_implemented operation=freshcaller.get.api-v1-users]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v1 users user-id - Documented GET /api/v1/users/{user_id} (not implemented) [intent=direct_read availability=not_implemented operation=freshcaller.get.api-v1-users-user-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get jobs job-id - Documented GET /jobs/{job_id} (not implemented) [intent=direct_read availability=not_implemented operation=freshcaller.get.jobs-job-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api post account export - Documented POST /account/export (not implemented) [intent=direct_write availability=not_implemented operation=freshcaller.post.account-export]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api v1 calls id note - Documented POST /api/v1/calls/{id}/note (not implemented) [intent=direct_write availability=not_implemented operation=freshcaller.post.api-v1-calls-id-note]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api v1 teams - Documented POST /api/v1/teams (not implemented) [intent=direct_write availability=not_implemented operation=freshcaller.post.api-v1-teams]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api v1 users - Documented POST /api/v1/users (not implemented) [intent=direct_write availability=not_implemented operation=freshcaller.post.api-v1-users]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put api v1 teams team-id - Documented PUT /api/v1/teams/{team_id} (not implemented) [intent=direct_write availability=not_implemented operation=freshcaller.put.api-v1-teams-team-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put api v1 users user-id - Documented PUT /api/v1/users/{user_id} (not implemented) [intent=direct_write availability=not_implemented operation=freshcaller.put.api-v1-users-user-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    calls list - Run the calls ETL stream [intent=etl availability=implemented stream=calls]
    numbers list - Run the numbers ETL stream [intent=etl availability=implemented stream=numbers]; notes: discrepancy=present-in-surface-absent-from-artifact
    teams list - Run the teams ETL stream [intent=etl availability=implemented stream=teams]

EXAMPLES
  # Inspect as a manual
  pm connectors inspect freshcaller

  # Inspect as structured JSON
  pm connectors inspect freshcaller --json

AGENT WORKFLOW
  - Run pm connectors inspect freshcaller before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
