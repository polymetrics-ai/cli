# pm connectors inspect pingdom

```text
NAME
  pm connectors inspect pingdom - Pingdom connector manual

SYNOPSIS
  pm connectors inspect pingdom
  pm connectors inspect pingdom --json
  pm credentials add <name> --connector pingdom [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Pingdom checks, probes, actions, maintenance windows/occurrences, alerting contacts/teams, credits, transaction checks, and reference data, and writes check/contact/team/maintenance mutations through API 3.1.

ICON
  id: simple-icons-pingdom
  asset: icons/simple-icons/pingdom.svg
  title: Pingdom
  simple_icon_slug: pingdom
  simple_icon_hex: FFF000
  source: simple-icons
  license: CC0-1.0
  review_status: cc0_with_trademark_caveat
  review_url: https://simpleicons.org/?q=Pingdom
  match: exact-name-or-slug
  matched_by: pingdom

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  api_key (secret)

ETL STREAMS
  checks:
    primary key: id
    fields: hostname(string), id(integer), lasterrortime(integer), lastresponsetime(integer), lasttesttime(integer), name(string), resolution(integer), status(string), tags(array), type(string)
  probes:
    primary key: id
    fields: active(boolean), city(string), country(string), hostname(string), id(integer), ip(string), name(string)
  actions:
    primary key: id
    fields: checkid(integer), checkname(string), contactname(string), id(integer), status(string), time(integer), via(integer)
  maintenance:
    primary key: id
    fields: description(string), effectiveto(integer), from(integer), id(integer), recurrencetype(string), repeatevery(integer), to(integer)
  reference:
    fields: checktypes(object), probes(object), regions(object)
  alerting_contacts:
    primary key: id
    fields: id(integer), name(string), notification_targets(object), owner(boolean), paused(boolean), teams(array), type(string)
  alerting_teams:
    primary key: id
    fields: id(integer), members(array), name(string)
  maintenance_occurrences:
    primary key: id
    fields: from(integer), id(integer), maintenanceid(integer), to(integer)
  credits:
    fields: autofillsms(boolean), autofillsms_amount(integer), autofillsms_when_left(integer), availablechecks(integer), availablerumsites(integer), availablesms(integer), availablesmstests(integer), checklimit(integer), max_sms_overage(integer), useddefault(integer), usedtransaction(integer)
  tms_checks:
    primary key: id
    fields: active(boolean), created_at(integer), id(integer), interval(integer), last_downtime_end(integer), last_downtime_start(integer), modified_at(integer), name(string), region(string), status(string), tags(array), type(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_check:
    endpoint: POST /checks
    required fields: name, host, type
    risk: creates a new Pingdom uptime check (this action models the common HTTP-type check shape; Pingdom's other 8 check types share the same name/host/type/paused/resolution/notification fields plus type-specific attributes not modeled here, see docs.md Known limits); low-risk external mutation, no approval required
  update_check:
    endpoint: PUT /checks/{{ record.id }}
    required fields: id
    risk: updates an existing check's settings (name/host/paused/resolution/tags); external mutation, approval required
  delete_check:
    endpoint: DELETE /checks/{{ record.id }}
    required fields: id
    risk: permanently deletes an uptime check and its historical results; destructive external mutation, approval required
  create_contact:
    endpoint: POST /alerting/contacts
    required fields: name, notification_targets
    risk: creates a new alerting contact with email/SMS notification targets; low-risk external mutation, no approval required
  update_contact:
    endpoint: PUT /alerting/contacts/{{ record.id }}
    required fields: id, name, paused, notification_targets
    risk: updates an existing alerting contact's name/paused state/notification targets (Pingdom's PUT is a full replacement, requiring name/paused/notification_targets together); external mutation, approval required
  delete_contact:
    endpoint: DELETE /alerting/contacts/{{ record.id }}
    required fields: id
    risk: permanently deletes an alerting contact and its notification targets; destructive external mutation, approval required
  create_team:
    endpoint: POST /alerting/teams
    required fields: name, member_ids
    risk: creates a new alerting team from a list of contact ids; low-risk external mutation, no approval required
  update_team:
    endpoint: PUT /alerting/teams/{{ record.id }}
    required fields: id, name, member_ids
    risk: updates an existing alerting team's name/member list; external mutation, approval required
  delete_team:
    endpoint: DELETE /alerting/teams/{{ record.id }}
    required fields: id
    risk: permanently deletes an alerting team; destructive external mutation, approval required
  create_maintenance:
    endpoint: POST /maintenance
    required fields: description, from, to
    risk: creates a new maintenance window that suppresses alerting for the assigned checks during the scheduled period; low-risk external mutation, no approval required
  delete_maintenance:
    endpoint: DELETE /maintenance/{{ record.id }}
    required fields: id
    risk: permanently deletes a maintenance window, immediately resuming alerting for its assigned checks; destructive external mutation, approval required

SECURITY
  read risk: external Pingdom API read of uptime/transaction monitoring configuration, alerting configuration, account credits, and event data
  write risk: creates/updates/deletes uptime checks, alerting contacts and teams, and maintenance windows
  approval: required for update_check/update_contact/update_team/delete_check/delete_contact/delete_team/delete_maintenance; create_check/create_contact/create_team/create_maintenance require no approval (low-risk, non-destructive)
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run Pingdom's declared streams and reverse-ETL actions.
  Usage: pm pingdom <command> [flags]
  Read streams
  Reverse ETL writes
  Other Commands
    actions list - Run the actions ETL stream [intent=etl availability=implemented stream=actions]
    alerting contacts list - Run the alerting contacts ETL stream [intent=etl availability=implemented stream=alerting_contacts]
    alerting teams list - Run the alerting teams ETL stream [intent=etl availability=implemented stream=alerting_teams]
    api delete checks - Documented DELETE /checks (not implemented) [intent=direct_write availability=not_implemented operation=pingdom.delete.checks]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete maintenance - Documented DELETE /maintenance (not implemented) [intent=direct_write availability=not_implemented operation=pingdom.delete.maintenance]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete maintenance-occurrences - Documented DELETE /maintenance.occurrences (not implemented) [intent=direct_write availability=not_implemented operation=pingdom.delete.maintenance-occurrences]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete maintenance-occurrences id - Documented DELETE /maintenance.occurrences/{id} (not implemented) [intent=direct_write availability=not_implemented operation=pingdom.delete.maintenance-occurrences-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete tms check cid - Documented DELETE /tms/check/{cid} (not implemented) [intent=direct_write availability=not_implemented operation=pingdom.delete.tms-check-cid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api get alerting contacts contactid - Documented GET /alerting/contacts/{contactid} (not implemented) [intent=direct_read availability=not_implemented operation=pingdom.get.alerting-contacts-contactid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get alerting teams teamid - Documented GET /alerting/teams/{teamid} (not implemented) [intent=direct_read availability=not_implemented operation=pingdom.get.alerting-teams-teamid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get analysis checkid - Documented GET /analysis/{checkid} (not implemented) [intent=direct_read availability=not_implemented operation=pingdom.get.analysis-checkid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get analysis checkid analysisid - Documented GET /analysis/{checkid}/{analysisid} (not implemented) [intent=direct_read availability=not_implemented operation=pingdom.get.analysis-checkid-analysisid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get checks checkid - Documented GET /checks/{checkid} (not implemented) [intent=direct_read availability=not_implemented operation=pingdom.get.checks-checkid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get maintenance id - Documented GET /maintenance/{id} (not implemented) [intent=direct_read availability=not_implemented operation=pingdom.get.maintenance-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get maintenance-occurrences id - Documented GET /maintenance.occurrences/{id} (not implemented) [intent=direct_read availability=not_implemented operation=pingdom.get.maintenance-occurrences-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get results checkid - Documented GET /results/{checkid} (not implemented) [intent=direct_read availability=not_implemented operation=pingdom.get.results-checkid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get single - Documented GET /single (not implemented) [intent=direct_read availability=not_implemented operation=pingdom.get.single]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get summary-average checkid - Documented GET /summary.average/{checkid} (not implemented) [intent=direct_read availability=not_implemented operation=pingdom.get.summary-average-checkid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get summary-hoursofday checkid - Documented GET /summary.hoursofday/{checkid} (not implemented) [intent=direct_read availability=not_implemented operation=pingdom.get.summary-hoursofday-checkid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get summary-outage checkid - Documented GET /summary.outage/{checkid} (not implemented) [intent=direct_read availability=not_implemented operation=pingdom.get.summary-outage-checkid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get summary-performance checkid - Documented GET /summary.performance/{checkid} (not implemented) [intent=direct_read availability=not_implemented operation=pingdom.get.summary-performance-checkid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get summary-probes checkid - Documented GET /summary.probes/{checkid} (not implemented) [intent=direct_read availability=not_implemented operation=pingdom.get.summary-probes-checkid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get tms check check-id report performance - Documented GET /tms/check/{check_id}/report/performance (not implemented) [intent=direct_read availability=not_implemented operation=pingdom.get.tms-check-check-id-report-performance]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get tms check check-id report status - Documented GET /tms/check/{check_id}/report/status (not implemented) [intent=direct_read availability=not_implemented operation=pingdom.get.tms-check-check-id-report-status]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get tms check cid - Documented GET /tms/check/{cid} (not implemented) [intent=direct_read availability=not_implemented operation=pingdom.get.tms-check-cid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get tms check report status - Documented GET /tms/check/report/status (not implemented) [intent=direct_read availability=not_implemented operation=pingdom.get.tms-check-report-status]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get traceroute - Documented GET /traceroute (not implemented) [intent=direct_read availability=not_implemented operation=pingdom.get.traceroute]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api post tms check - Documented POST /tms/check (not implemented) [intent=direct_write availability=not_implemented operation=pingdom.post.tms-check]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put checks - Documented PUT /checks (not implemented) [intent=direct_write availability=not_implemented operation=pingdom.put.checks]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put maintenance id - Documented PUT /maintenance/{id} (not implemented) [intent=direct_write availability=not_implemented operation=pingdom.put.maintenance-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put maintenance-occurrences id - Documented PUT /maintenance.occurrences/{id} (not implemented) [intent=direct_write availability=not_implemented operation=pingdom.put.maintenance-occurrences-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put tms check cid - Documented PUT /tms/check/{cid} (not implemented) [intent=direct_write availability=not_implemented operation=pingdom.put.tms-check-cid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    checks list - Run the checks ETL stream [intent=etl availability=implemented stream=checks]
    create check apply - Plan and execute the create check reverse-ETL action [intent=reverse_etl availability=implemented write=create_check]; approval: requires plan, preview, approval, and execute; risk: creates a new Pingdom uptime check (this action models the common HTTP-type check shape; Pingdom's other 8 check types share the same name/host/type/paused/resolution/notification fields plus type-specific attributes not modeled here, see docs.md Known limits); low-risk external mutation, no approval required; flags: --host (required), --name (required), --type (required)
    create contact apply - Plan and execute the create contact reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_contact]; approval: requires plan, preview, approval, and execute; risk: creates a new alerting contact with email/SMS notification targets; low-risk external mutation, no approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create maintenance apply - Plan and execute the create maintenance reverse-ETL action [intent=reverse_etl availability=implemented write=create_maintenance]; approval: requires plan, preview, approval, and execute; risk: creates a new maintenance window that suppresses alerting for the assigned checks during the scheduled period; low-risk external mutation, no approval required; flags: --description (required), --from (required), --to (required)
    create team apply - Plan and execute the create team reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_team]; approval: requires plan, preview, approval, and execute; risk: creates a new alerting team from a list of contact ids; low-risk external mutation, no approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    credits list - Run the credits ETL stream [intent=etl availability=implemented stream=credits]
    delete check apply - Plan and execute the delete check reverse-ETL action [intent=reverse_etl availability=implemented write=delete_check]; approval: requires plan, preview, approval, and execute; risk: permanently deletes an uptime check and its historical results; destructive external mutation, approval required; flags: --id (required)
    delete contact apply - Plan and execute the delete contact reverse-ETL action [intent=reverse_etl availability=implemented write=delete_contact]; approval: requires plan, preview, approval, and execute; risk: permanently deletes an alerting contact and its notification targets; destructive external mutation, approval required; flags: --id (required)
    delete maintenance apply - Plan and execute the delete maintenance reverse-ETL action [intent=reverse_etl availability=implemented write=delete_maintenance]; approval: requires plan, preview, approval, and execute; risk: permanently deletes a maintenance window, immediately resuming alerting for its assigned checks; destructive external mutation, approval required; flags: --id (required)
    delete team apply - Plan and execute the delete team reverse-ETL action [intent=reverse_etl availability=implemented write=delete_team]; approval: requires plan, preview, approval, and execute; risk: permanently deletes an alerting team; destructive external mutation, approval required; flags: --id (required)
    maintenance list - Run the maintenance ETL stream [intent=etl availability=implemented stream=maintenance]
    maintenance occurrences list - Run the maintenance occurrences ETL stream [intent=etl availability=implemented stream=maintenance_occurrences]
    probes list - Run the probes ETL stream [intent=etl availability=implemented stream=probes]
    reference list - Run the reference ETL stream [intent=etl availability=implemented stream=reference]
    tms checks list - Run the tms checks ETL stream [intent=etl availability=implemented stream=tms_checks]
    update check apply - Plan and execute the update check reverse-ETL action [intent=reverse_etl availability=implemented write=update_check]; approval: requires plan, preview, approval, and execute; risk: updates an existing check's settings (name/host/paused/resolution/tags); external mutation, approval required; flags: --id (required)
    update contact apply - Plan and execute the update contact reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_contact]; approval: requires plan, preview, approval, and execute; risk: updates an existing alerting contact's name/paused state/notification targets (Pingdom's PUT is a full replacement, requiring name/paused/notification_targets together); external mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update team apply - Plan and execute the update team reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_team]; approval: requires plan, preview, approval, and execute; risk: updates an existing alerting team's name/member list; external mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags

EXAMPLES
  # Inspect as a manual
  pm connectors inspect pingdom

  # Inspect as structured JSON
  pm connectors inspect pingdom --json

AGENT WORKFLOW
  - Run pm connectors inspect pingdom before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
