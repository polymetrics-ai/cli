---
name: pm-solarwinds-service-desk
description: SolarWinds Service Desk connector knowledge and safe action guide.
---

# pm-solarwinds-service-desk

## Purpose

Reads SolarWinds Service Desk incidents, problems, changes, change catalogs, releases, solutions, catalog items, configuration items, users, sites, departments, roles, groups, categories, hardware/mobile/other/software assets, printers, contracts, purchase orders, vendors, audits, and risks; writes delete actions for every resource with a documented delete-by-id endpoint.

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- page
- per_page
- start_date
- api_key (secret)
- api_key_2 (secret)

## ETL Streams

- incidents:
  - primary key: id
  - fields: created_at(string), id(integer), name(string), updated_at(string)
- users:
  - primary key: id
  - fields: created_at(string), id(integer), name(string), updated_at(string)
- departments:
  - primary key: id
  - fields: created_at(string), id(integer), name(string), updated_at(string)
- categories:
  - primary key: id
  - fields: created_at(string), id(integer), name(string), updated_at(string)
- problems:
  - primary key: id
  - fields: created_at(string), id(integer), name(string), updated_at(string)
- changes:
  - primary key: id
  - fields: created_at(string), id(integer), name(string), updated_at(string)
- change_catalogs:
  - primary key: id
  - fields: created_at(string), id(integer), name(string), updated_at(string)
- releases:
  - primary key: id
  - fields: created_at(string), id(integer), name(string), updated_at(string)
- solutions:
  - primary key: id
  - fields: created_at(string), id(integer), name(string), updated_at(string)
- catalog_items:
  - primary key: id
  - fields: created_at(string), id(integer), name(string), updated_at(string)
- configuration_items:
  - primary key: id
  - fields: created_at(string), id(integer), name(string), updated_at(string)
- sites:
  - primary key: id
  - fields: created_at(string), id(integer), name(string), updated_at(string)
- roles:
  - primary key: id
  - fields: created_at(string), id(integer), name(string), updated_at(string)
- groups:
  - primary key: id
  - fields: created_at(string), id(integer), name(string), updated_at(string)
- hardwares:
  - primary key: id
  - fields: created_at(string), id(integer), name(string), updated_at(string)
- mobiles:
  - primary key: id
  - fields: created_at(string), id(integer), name(string), updated_at(string)
- other_assets:
  - primary key: id
  - fields: created_at(string), id(integer), name(string), updated_at(string)
- softwares:
  - primary key: id
  - fields: created_at(string), id(integer), name(string), updated_at(string)
- printers:
  - primary key: id
  - fields: created_at(string), id(integer), name(string), updated_at(string)
- contracts:
  - primary key: id
  - fields: created_at(string), id(integer), name(string), updated_at(string)
- purchase_orders:
  - primary key: id
  - fields: created_at(string), id(integer), name(string), updated_at(string)
- vendors:
  - primary key: id
  - fields: created_at(string), id(integer), name(string), updated_at(string)
- audits:
  - primary key: id
  - fields: created_at(string), id(integer), name(string), updated_at(string)
- risks:
  - primary key: id
  - fields: created_at(string), id(integer), name(string), updated_at(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- delete_incident:
  - endpoint: DELETE /incidents/{{ record.id }}.json
  - required fields: id
  - risk: external mutation; permanently deletes a SolarWinds Service Desk incident record; approval required
- delete_problem:
  - endpoint: DELETE /problems/{{ record.id }}.json
  - required fields: id
  - risk: external mutation; permanently deletes a SolarWinds Service Desk problem record; approval required
- delete_change:
  - endpoint: DELETE /changes/{{ record.id }}.json
  - required fields: id
  - risk: external mutation; permanently deletes a SolarWinds Service Desk change record; approval required
- delete_change_catalog:
  - endpoint: DELETE /change_catalogs/{{ record.id }}.json
  - required fields: id
  - risk: external mutation; permanently deletes a SolarWinds Service Desk change catalog record; approval required
- delete_release:
  - endpoint: DELETE /releases/{{ record.id }}.json
  - required fields: id
  - risk: external mutation; permanently deletes a SolarWinds Service Desk release record; approval required
- delete_solution:
  - endpoint: DELETE /solutions/{{ record.id }}.json
  - required fields: id
  - risk: external mutation; permanently deletes a SolarWinds Service Desk solution record; approval required
- delete_catalog_item:
  - endpoint: DELETE /catalog_items/{{ record.id }}.json
  - required fields: id
  - risk: external mutation; permanently deletes a SolarWinds Service Desk catalog item record; approval required
- delete_configuration_item:
  - endpoint: DELETE /configuration_items/{{ record.id }}.json
  - required fields: id
  - risk: external mutation; permanently deletes a SolarWinds Service Desk configuration item record; approval required
- delete_user:
  - endpoint: DELETE /users/{{ record.id }}.json
  - required fields: id
  - risk: external mutation; permanently deletes a SolarWinds Service Desk user record; approval required
- delete_site:
  - endpoint: DELETE /sites/{{ record.id }}.json
  - required fields: id
  - risk: external mutation; permanently deletes a SolarWinds Service Desk site record; approval required
- delete_department:
  - endpoint: DELETE /departments/{{ record.id }}.json
  - required fields: id
  - risk: external mutation; permanently deletes a SolarWinds Service Desk department record; approval required
- delete_role:
  - endpoint: DELETE /roles/{{ record.id }}.json
  - required fields: id
  - risk: external mutation; permanently deletes a SolarWinds Service Desk role record; approval required
- delete_group:
  - endpoint: DELETE /groups/{{ record.id }}.json
  - required fields: id
  - risk: external mutation; permanently deletes a SolarWinds Service Desk group record; approval required
- delete_category:
  - endpoint: DELETE /categories/{{ record.id }}.json
  - required fields: id
  - risk: external mutation; permanently deletes a SolarWinds Service Desk category record; approval required
- delete_hardware:
  - endpoint: DELETE /hardwares/{{ record.id }}.json
  - required fields: id
  - risk: external mutation; permanently deletes a SolarWinds Service Desk hardware record; approval required
- delete_mobile:
  - endpoint: DELETE /mobiles/{{ record.id }}.json
  - required fields: id
  - risk: external mutation; permanently deletes a SolarWinds Service Desk mobile record; approval required
- delete_other_asset:
  - endpoint: DELETE /other_assets/{{ record.id }}.json
  - required fields: id
  - risk: external mutation; permanently deletes a SolarWinds Service Desk other asset record; approval required
- delete_contract:
  - endpoint: DELETE /contracts/{{ record.id }}.json
  - required fields: id
  - risk: external mutation; permanently deletes a SolarWinds Service Desk contract record; approval required
- delete_purchase_order:
  - endpoint: DELETE /purchase_orders/{{ record.id }}.json
  - required fields: id
  - risk: external mutation; permanently deletes a SolarWinds Service Desk purchase order record; approval required
- delete_vendor:
  - endpoint: DELETE /vendors/{{ record.id }}.json
  - required fields: id
  - risk: external mutation; permanently deletes a SolarWinds Service Desk vendor record; approval required

## Security

- read risk: external SolarWinds Service Desk API read of incident, problem, change, asset, and organizational (user/site/department/role/group) data
- write risk: external SolarWinds Service Desk API delete mutations against incidents, problems, changes, assets, and organizational records
- approval: required for all write actions; every write action is an irreversible external delete
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Run SolarWinds Service Desk's declared streams and reverse-ETL actions.
- Usage: pm solarwinds-service-desk <command> [flags]
- Read streams
- Reverse ETL writes
- Other Commands
  - api delete contracts contract-id items item-id - Documented DELETE /contracts/{contract_id}/items/{item_id} (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.delete.contracts-contract-id-items-item-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete hardwares hardware-id warranties warranty-id - Documented DELETE /hardwares/{hardware_id}/warranties/{warranty_id} (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.delete.hardwares-hardware-id-warranties-warranty-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete memberships id - Documented DELETE /memberships/{id} (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.delete.memberships-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete object-type id comments comment-id - Documented DELETE /{object_type}/{id}/comments/{comment_id} (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.delete.object-type-id-comments-comment-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete object-type id purchases purchase-id - Documented DELETE /{object_type}/{id}/purchases/{purchase_id} (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.delete.object-type-id-purchases-purchase-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete object-type id tasks task-id - Documented DELETE /{object_type}/{id}/tasks/{task_id} (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.delete.object-type-id-tasks-task-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete object-type id time-tracks time-track-id - Documented DELETE /{object_type}/{id}/time_tracks/{time_track_id} (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.delete.object-type-id-time-tracks-time-track-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api get catalog-items id - Documented GET /catalog_items/{id} (not implemented) [intent=direct_read availability=not_implemented operation=solarwinds-service-desk.get.catalog-items-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get categories id - Documented GET /categories/{id} (not implemented) [intent=direct_read availability=not_implemented operation=solarwinds-service-desk.get.categories-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get change-catalogs id - Documented GET /change_catalogs/{id} (not implemented) [intent=direct_read availability=not_implemented operation=solarwinds-service-desk.get.change-catalogs-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get changes id - Documented GET /changes/{id} (not implemented) [intent=direct_read availability=not_implemented operation=solarwinds-service-desk.get.changes-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get configuration-items id - Documented GET /configuration_items/{id} (not implemented) [intent=direct_read availability=not_implemented operation=solarwinds-service-desk.get.configuration-items-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get contracts id - Documented GET /contracts/{id} (not implemented) [intent=direct_read availability=not_implemented operation=solarwinds-service-desk.get.contracts-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get departments id - Documented GET /departments/{id} (not implemented) [intent=direct_read availability=not_implemented operation=solarwinds-service-desk.get.departments-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get groups id - Documented GET /groups/{id} (not implemented) [intent=direct_read availability=not_implemented operation=solarwinds-service-desk.get.groups-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get hardwares id - Documented GET /hardwares/{id} (not implemented) [intent=direct_read availability=not_implemented operation=solarwinds-service-desk.get.hardwares-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get hardwares id warranties - Documented GET /hardwares/{id}/warranties (not implemented) [intent=direct_read availability=not_implemented operation=solarwinds-service-desk.get.hardwares-id-warranties]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get incidents id - Documented GET /incidents/{id} (not implemented) [intent=direct_read availability=not_implemented operation=solarwinds-service-desk.get.incidents-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get mobiles id - Documented GET /mobiles/{id} (not implemented) [intent=direct_read availability=not_implemented operation=solarwinds-service-desk.get.mobiles-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get object-type id audits - Documented GET /{object_type}/{id}/audits (not implemented) [intent=direct_read availability=not_implemented operation=solarwinds-service-desk.get.object-type-id-audits]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get object-type id time-tracks - Documented GET /{object_type}/{id}/time_tracks (not implemented) [intent=direct_read availability=not_implemented operation=solarwinds-service-desk.get.object-type-id-time-tracks]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get other-assets id - Documented GET /other_assets/{id} (not implemented) [intent=direct_read availability=not_implemented operation=solarwinds-service-desk.get.other-assets-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get printers id - Documented GET /printers/{id} (not implemented) [intent=direct_read availability=not_implemented operation=solarwinds-service-desk.get.printers-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get problems id - Documented GET /problems/{id} (not implemented) [intent=direct_read availability=not_implemented operation=solarwinds-service-desk.get.problems-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get purchase-orders id - Documented GET /purchase_orders/{id} (not implemented) [intent=direct_read availability=not_implemented operation=solarwinds-service-desk.get.purchase-orders-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get releases id - Documented GET /releases/{id} (not implemented) [intent=direct_read availability=not_implemented operation=solarwinds-service-desk.get.releases-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get roles id - Documented GET /roles/{id} (not implemented) [intent=direct_read availability=not_implemented operation=solarwinds-service-desk.get.roles-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get sites id - Documented GET /sites/{id} (not implemented) [intent=direct_read availability=not_implemented operation=solarwinds-service-desk.get.sites-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get softwares id - Documented GET /softwares/{id} (not implemented) [intent=direct_read availability=not_implemented operation=solarwinds-service-desk.get.softwares-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get solutions id - Documented GET /solutions/{id} (not implemented) [intent=direct_read availability=not_implemented operation=solarwinds-service-desk.get.solutions-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get users id - Documented GET /users/{id} (not implemented) [intent=direct_read availability=not_implemented operation=solarwinds-service-desk.get.users-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get vendors id - Documented GET /vendors/{id} (not implemented) [intent=direct_read availability=not_implemented operation=solarwinds-service-desk.get.vendors-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api post attachments - Documented POST /attachments (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.post.attachments]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post catalog-items - Documented POST /catalog_items (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.post.catalog-items]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post catalog-items id service-requests - Documented POST /catalog_items/{id}/service_requests (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.post.catalog-items-id-service-requests]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post categories - Documented POST /categories (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.post.categories]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post change-catalogs - Documented POST /change_catalogs (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.post.change-catalogs]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post change-catalogs id change-requests - Documented POST /change_catalogs/{id}/change_requests (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.post.change-catalogs-id-change-requests]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post changes - Documented POST /changes (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.post.changes]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post configuration-items - Documented POST /configuration_items (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.post.configuration-items]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post contracts - Documented POST /contracts (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.post.contracts]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post contracts id items - Documented POST /contracts/{id}/items (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.post.contracts-id-items]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post departments - Documented POST /departments (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.post.departments]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post groups - Documented POST /groups (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.post.groups]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post hardwares - Documented POST /hardwares (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.post.hardwares]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post hardwares id warranties - Documented POST /hardwares/{id}/warranties (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.post.hardwares-id-warranties]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post incidents - Documented POST /incidents (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.post.incidents]; approval: not implemented: the REST write executor lacks the provider-specific top-level body envelope required by this operation; risk: high; notes: named_dependency=engine.rest_write_body_envelope: the REST write executor lacks the provider-specific top-level body envelope required by this operation
  - api post memberships - Documented POST /memberships (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.post.memberships]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post mobiles - Documented POST /mobiles (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.post.mobiles]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post object-type id comments - Documented POST /{object_type}/{id}/comments (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.post.object-type-id-comments]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post object-type id purchases - Documented POST /{object_type}/{id}/purchases (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.post.object-type-id-purchases]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post object-type id tasks - Documented POST /{object_type}/{id}/tasks (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.post.object-type-id-tasks]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post object-type id time-tracks - Documented POST /{object_type}/{id}/time_tracks (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.post.object-type-id-time-tracks]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post other-assets - Documented POST /other_assets (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.post.other-assets]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post problems - Documented POST /problems (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.post.problems]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post purchase-orders - Documented POST /purchase_orders (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.post.purchase-orders]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post releases - Documented POST /releases (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.post.releases]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post roles - Documented POST /roles (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.post.roles]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post sites - Documented POST /sites (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.post.sites]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post solutions - Documented POST /solutions (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.post.solutions]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post users - Documented POST /users (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.post.users]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post vendors - Documented POST /vendors (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.post.vendors]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put asset-links delete-asset-link-by-id - Documented PUT /asset_links/delete_asset_link_by_id (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.put.asset-links-delete-asset-link-by-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put catalog-items id - Documented PUT /catalog_items/{id} (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.put.catalog-items-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put categories id - Documented PUT /categories/{id} (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.put.categories-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put change-catalogs id - Documented PUT /change_catalogs/{id} (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.put.change-catalogs-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put changes id - Documented PUT /changes/{id} (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.put.changes-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put configuration-items id - Documented PUT /configuration_items/{id} (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.put.configuration-items-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put configuration-items id append-multiple-dependent-assets - Documented PUT /configuration_items/{id}/append_multiple_dependent_assets (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.put.configuration-items-id-append-multiple-dependent-assets]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put contracts contract-id items item-id - Documented PUT /contracts/{contract_id}/items/{item_id} (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.put.contracts-contract-id-items-item-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put contracts id - Documented PUT /contracts/{id} (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.put.contracts-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put departments id - Documented PUT /departments/{id} (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.put.departments-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put groups id - Documented PUT /groups/{id} (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.put.groups-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put hardwares hardware-id warranties warranty-id - Documented PUT /hardwares/{hardware_id}/warranties/{warranty_id} (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.put.hardwares-hardware-id-warranties-warranty-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put hardwares id - Documented PUT /hardwares/{id} (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.put.hardwares-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put incidents id - Documented PUT /incidents/{id} (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.put.incidents-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put mobiles id - Documented PUT /mobiles/{id} (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.put.mobiles-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put object-type id comments comment-id - Documented PUT /{object_type}/{id}/comments/{comment_id} (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.put.object-type-id-comments-comment-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put object-type id purchases purchase-id - Documented PUT /{object_type}/{id}/purchases/{purchase_id} (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.put.object-type-id-purchases-purchase-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put object-type id tasks task-id - Documented PUT /{object_type}/{id}/tasks/{task_id} (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.put.object-type-id-tasks-task-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put object-type id time-tracks time-track-id - Documented PUT /{object_type}/{id}/time_tracks/{time_track_id} (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.put.object-type-id-time-tracks-time-track-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put other-assets id - Documented PUT /other_assets/{id} (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.put.other-assets-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put printers id - Documented PUT /printers/{id} (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.put.printers-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put problems id - Documented PUT /problems/{id} (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.put.problems-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put purchase-orders id - Documented PUT /purchase_orders/{id} (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.put.purchase-orders-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put releases id - Documented PUT /releases/{id} (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.put.releases-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put roles id - Documented PUT /roles/{id} (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.put.roles-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put sites id - Documented PUT /sites/{id} (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.put.sites-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put solutions id - Documented PUT /solutions/{id} (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.put.solutions-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put users id - Documented PUT /users/{id} (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.put.users-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put vendors id - Documented PUT /vendors/{id} (not implemented) [intent=direct_write availability=not_implemented operation=solarwinds-service-desk.put.vendors-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - audits list - Run the audits ETL stream [intent=etl availability=implemented stream=audits]
  - catalog items list - Run the catalog items ETL stream [intent=etl availability=implemented stream=catalog_items]
  - categories list - Run the categories ETL stream [intent=etl availability=implemented stream=categories]
  - change catalogs list - Run the change catalogs ETL stream [intent=etl availability=implemented stream=change_catalogs]
  - changes list - Run the changes ETL stream [intent=etl availability=implemented stream=changes]
  - configuration items list - Run the configuration items ETL stream [intent=etl availability=implemented stream=configuration_items]
  - contracts list - Run the contracts ETL stream [intent=etl availability=implemented stream=contracts]
  - delete catalog item apply - Plan and execute the delete catalog item reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_catalog_item]; approval: requires plan, preview, approval, and execute; risk: external mutation; permanently deletes a SolarWinds Service Desk catalog item record; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - delete category apply - Plan and execute the delete category reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_category]; approval: requires plan, preview, approval, and execute; risk: external mutation; permanently deletes a SolarWinds Service Desk category record; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - delete change apply - Plan and execute the delete change reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_change]; approval: requires plan, preview, approval, and execute; risk: external mutation; permanently deletes a SolarWinds Service Desk change record; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - delete change catalog apply - Plan and execute the delete change catalog reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_change_catalog]; approval: requires plan, preview, approval, and execute; risk: external mutation; permanently deletes a SolarWinds Service Desk change catalog record; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - delete configuration item apply - Plan and execute the delete configuration item reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_configuration_item]; approval: requires plan, preview, approval, and execute; risk: external mutation; permanently deletes a SolarWinds Service Desk configuration item record; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - delete contract apply - Plan and execute the delete contract reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_contract]; approval: requires plan, preview, approval, and execute; risk: external mutation; permanently deletes a SolarWinds Service Desk contract record; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - delete department apply - Plan and execute the delete department reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_department]; approval: requires plan, preview, approval, and execute; risk: external mutation; permanently deletes a SolarWinds Service Desk department record; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - delete group apply - Plan and execute the delete group reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_group]; approval: requires plan, preview, approval, and execute; risk: external mutation; permanently deletes a SolarWinds Service Desk group record; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - delete hardware apply - Plan and execute the delete hardware reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_hardware]; approval: requires plan, preview, approval, and execute; risk: external mutation; permanently deletes a SolarWinds Service Desk hardware record; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - delete incident apply - Plan and execute the delete incident reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_incident]; approval: requires plan, preview, approval, and execute; risk: external mutation; permanently deletes a SolarWinds Service Desk incident record; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - delete mobile apply - Plan and execute the delete mobile reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_mobile]; approval: requires plan, preview, approval, and execute; risk: external mutation; permanently deletes a SolarWinds Service Desk mobile record; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - delete other asset apply - Plan and execute the delete other asset reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_other_asset]; approval: requires plan, preview, approval, and execute; risk: external mutation; permanently deletes a SolarWinds Service Desk other asset record; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - delete problem apply - Plan and execute the delete problem reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_problem]; approval: requires plan, preview, approval, and execute; risk: external mutation; permanently deletes a SolarWinds Service Desk problem record; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - delete purchase order apply - Plan and execute the delete purchase order reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_purchase_order]; approval: requires plan, preview, approval, and execute; risk: external mutation; permanently deletes a SolarWinds Service Desk purchase order record; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - delete release apply - Plan and execute the delete release reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_release]; approval: requires plan, preview, approval, and execute; risk: external mutation; permanently deletes a SolarWinds Service Desk release record; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - delete role apply - Plan and execute the delete role reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_role]; approval: requires plan, preview, approval, and execute; risk: external mutation; permanently deletes a SolarWinds Service Desk role record; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - delete site apply - Plan and execute the delete site reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_site]; approval: requires plan, preview, approval, and execute; risk: external mutation; permanently deletes a SolarWinds Service Desk site record; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - delete solution apply - Plan and execute the delete solution reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_solution]; approval: requires plan, preview, approval, and execute; risk: external mutation; permanently deletes a SolarWinds Service Desk solution record; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - delete user apply - Plan and execute the delete user reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_user]; approval: requires plan, preview, approval, and execute; risk: external mutation; permanently deletes a SolarWinds Service Desk user record; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - delete vendor apply - Plan and execute the delete vendor reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_vendor]; approval: requires plan, preview, approval, and execute; risk: external mutation; permanently deletes a SolarWinds Service Desk vendor record; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - departments list - Run the departments ETL stream [intent=etl availability=implemented stream=departments]
  - groups list - Run the groups ETL stream [intent=etl availability=implemented stream=groups]
  - hardwares list - Run the hardwares ETL stream [intent=etl availability=implemented stream=hardwares]
  - incidents list - Run the incidents ETL stream [intent=etl availability=implemented stream=incidents]
  - mobiles list - Run the mobiles ETL stream [intent=etl availability=implemented stream=mobiles]
  - other assets list - Run the other assets ETL stream [intent=etl availability=implemented stream=other_assets]
  - printers list - Run the printers ETL stream [intent=etl availability=implemented stream=printers]
  - problems list - Run the problems ETL stream [intent=etl availability=implemented stream=problems]
  - purchase orders list - Run the purchase orders ETL stream [intent=etl availability=implemented stream=purchase_orders]
  - releases list - Run the releases ETL stream [intent=etl availability=implemented stream=releases]
  - risks list - Run the risks ETL stream [intent=etl availability=implemented stream=risks]
  - roles list - Run the roles ETL stream [intent=etl availability=implemented stream=roles]
  - sites list - Run the sites ETL stream [intent=etl availability=implemented stream=sites]
  - softwares list - Run the softwares ETL stream [intent=etl availability=implemented stream=softwares]
  - solutions list - Run the solutions ETL stream [intent=etl availability=implemented stream=solutions]
  - users list - Run the users ETL stream [intent=etl availability=implemented stream=users]
  - vendors list - Run the vendors ETL stream [intent=etl availability=implemented stream=vendors]

## Commands

### Inspect as a manual

```bash
pm connectors inspect solarwinds-service-desk
```

### Inspect as structured JSON

```bash
pm connectors inspect solarwinds-service-desk --json
```

## Agent Rules

- Run pm connectors inspect solarwinds-service-desk before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
