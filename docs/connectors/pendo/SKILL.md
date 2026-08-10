---
name: pm-pendo
description: Pendo connector knowledge and safe action guide.
---

# pm-pendo

## Purpose

Reads Pendo Engage visitors, accounts, product objects, guides, reports, metadata, exclusion lists, servers, and feedback options; exposes safe segment, guide, and feedback mutations.

## Icon

- id: pendo
- asset: icons/pendo.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://engageapi.pendo.io/

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- account_id
- base_url
- blacklist_type
- bulkdelete_id
- feature_id
- feature_ids
- flag_name
- guide_id
- limit
- max_pages
- metadata_field_name
- metadata_group
- metadata_kind
- mode
- page_id
- page_ids
- report_id
- segment_id
- server_name
- tracktype_id
- visitor_history_starttime
- visitor_id
- integration_key (secret)

## ETL Streams

- visitors:
  - primary key: id
  - cursor: lastVisit
  - fields: accountId(string), email(string), id(string), lastVisit(string)
- accounts:
  - primary key: id
  - cursor: lastVisit
  - fields: id(string), lastVisit(string), name(string)
- pages:
  - primary key: id
  - cursor: lastUpdated
  - fields: id(string), lastUpdated(string), name(string)
- features:
  - primary key: id
  - cursor: lastUpdated
  - fields: id(string), lastUpdated(string), name(string)
- page_by_id:
  - primary key: id
  - fields: id(string)
- pages_by_ids:
  - primary key: id
  - fields: id(string)
- feature_by_id:
  - primary key: id
  - fields: id(string)
- features_by_ids:
  - primary key: id
  - fields: id(string)
- tracktypes:
  - primary key: id
  - fields: id(string)
- tracktype_by_id:
  - primary key: id
  - fields: id(string)
- visitor_by_id:
  - primary key: id
  - fields: id(string)
- visitor_history:
  - primary key: ts, type
  - fields: ts(string), type(string)
- account_by_id:
  - primary key: id
  - fields: id(string)
- bulkdelete_requests:
  - primary key: id
  - fields: id(string)
- bulkdelete_request:
  - primary key: id
  - fields: id(string)
- segments:
  - primary key: id
  - fields: id(string)
- segment_by_id:
  - primary key: id
  - fields: id(string)
- segment_status:
  - primary key: requestId
  - fields: requestId(string)
- reports:
  - primary key: id
  - fields: id(string)
- report_results_json:
- guides:
  - primary key: id
  - fields: id(string)
- guide_by_id:
  - primary key: id
  - fields: id(string)
- guide_history:
  - primary key: id
  - fields: id(string)
- guide_order:
- metadata_schema:
- metadata_dependencies:
- metadata_field_dependencies:
- blacklist:
  - primary key: id
  - fields: id(string)
- blacklist_by_type:
  - primary key: id
  - fields: id(string)
- servers:
  - primary key: id
  - fields: id(string)
- server_by_name:
  - primary key: id
  - fields: id(string)
- servers_by_flag:
  - primary key: id
  - fields: id(string)
- feedback_options:

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- start_segment_visitor_export:
  - endpoint: POST /segment/{{ record.segmentId }}/visitors
  - required fields: segmentId
  - risk: starts an asynchronous Pendo segment visitor export job; approval required
- create_segment:
  - endpoint: POST /segment/upload
  - required fields: name, visitors
  - risk: creates a shared Pendo segment from visitor ids; approval required
- update_segment:
  - endpoint: PUT /segment/{{ record.segmentId }}
  - required fields: segmentId, name, visitors
  - risk: replaces the visitor membership for a Pendo segment; approval required
- delete_segment:
  - endpoint: DELETE /segment/{{ record.segmentId }}
  - required fields: segmentId
  - risk: deletes a Pendo segment; destructive external mutation
- add_segment_visitor:
  - endpoint: PUT /segment/{{ record.segmentId }}/visitor/{{ record.visitorId }}
  - required fields: segmentId, visitorId
  - risk: adds a visitor to a Pendo segment; approval required
- remove_segment_visitor:
  - endpoint: DELETE /segment/{{ record.segmentId }}/visitor/{{ record.visitorId }}
  - required fields: segmentId, visitorId
  - risk: removes a visitor from a Pendo segment; destructive membership mutation
- patch_segment_visitors:
  - endpoint: PATCH /segment/{{ record.segmentId }}/visitor
  - required fields: segmentId, patch
  - risk: adds/removes a small batch of visitors for a Pendo segment; approval required
- reset_guide_for_visitor:
  - endpoint: POST /guide/{{ record.guideId }}/visitor/{{ record.visitorId }}/reset
  - required fields: guideId, visitorId
  - risk: resets guide-seen state for one visitor; approval required
- reset_all_guides_for_visitor:
  - endpoint: POST /guide/all/visitor/{{ record.visitorId }}/reset
  - required fields: visitorId
  - risk: resets guide-seen state for one visitor across all guides; approval required
- reset_staged_guide:
  - endpoint: POST /guide/{{ record.guideId }}/reset
  - required fields: guideId
  - risk: resets one staged guide; approval required
- reset_all_staged_guides:
  - endpoint: POST /guide/staged/reset
  - risk: resets all staged guides in the subscription; approval required
- change_guide_segment:
  - endpoint: PUT /guide/{{ record.guideId }}/segment
  - required fields: guideId, segmentId
  - risk: changes the segment assigned to a Pendo guide; approval required
- change_guide_state:
  - endpoint: PUT /guide/{{ record.guideId }}/state
  - required fields: guideId, state
  - risk: changes a Pendo guide state such as public, staged, disabled, or draft; approval required
- create_feedback:
  - endpoint: POST /feedback
  - required fields: accountId, visitorId, title
  - risk: creates a Pendo Listen feedback item; approval required
- update_feedback:
  - endpoint: PATCH /feedback/{{ record.id }}
  - required fields: id
  - risk: updates a Pendo Listen feedback item; approval required
- delete_feedback:
  - endpoint: DELETE /feedback
  - required fields: ids
  - risk: deletes Pendo Listen feedback items by id; destructive external mutation

## Security

- read risk: external Pendo API read of product analytics, visitor/account, guide, report, metadata, segment, and feedback data
- write risk: mutates Pendo segments, guide state/seen status, and Listen feedback records
- approval: reverse ETL writes require plan preview and approval token
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Run Pendo's declared streams and reverse-ETL actions.
- Usage: pm pendo <command> [flags]
- Read streams
- Reverse ETL writes
- Other Commands
  - account by id list - Run the account by id ETL stream [intent=etl availability=implemented stream=account_by_id]; notes: discrepancy=present-in-surface-absent-from-artifact
  - accounts list - Run the accounts ETL stream [intent=etl availability=implemented stream=accounts]; notes: discrepancy=present-in-surface-absent-from-artifact
  - add segment visitor apply - Plan and execute the add segment visitor reverse-ETL action [intent=reverse_etl availability=not_implemented write=add_segment_visitor]; approval: requires plan, preview, approval, and execute; risk: adds a visitor to a Pendo segment; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - api delete api v1 feedback - Documented DELETE /api/v1/feedback (not implemented) [intent=direct_write availability=not_implemented operation=pendo.delete.api-v1-feedback]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete api v1 metadata schema kind group field - Documented DELETE /api/v1/metadata/schema/{kind}/{group}/{field} (not implemented) [intent=direct_write availability=not_implemented operation=pendo.delete.api-v1-metadata-schema-kind-group-field]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete api v1 segment segmentid - Documented DELETE /api/v1/segment/{segmentId} (not implemented) [intent=direct_write availability=not_implemented operation=pendo.delete.api-v1-segment-segmentid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete api v1 segment segmentid visitor visitorid - Documented DELETE /api/v1/segment/{segmentId}/visitor/{visitorId} (not implemented) [intent=direct_write availability=not_implemented operation=pendo.delete.api-v1-segment-segmentid-visitor-visitorid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete metadata schema kind group field - Documented DELETE /metadata/schema/{kind}/{group}/{field} (not implemented) [intent=direct_write availability=not_implemented operation=pendo.delete.metadata-schema-kind-group-field]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api get account b64-id - Documented GET /account/{b64_id} (not implemented) [intent=direct_read availability=not_implemented operation=pendo.get.account-b64-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 account accountid - Documented GET /api/v1/account/{accountId} (not implemented) [intent=direct_read availability=not_implemented operation=pendo.get.api-v1-account-accountid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 account b64-id - Documented GET /api/v1/account/{b64_id} (not implemented) [intent=direct_read availability=not_implemented operation=pendo.get.api-v1-account-b64-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 blacklist - Documented GET /api/v1/blacklist (not implemented) [intent=direct_read availability=not_implemented operation=pendo.get.api-v1-blacklist]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 blacklist type type - Documented GET /api/v1/blacklist/type/{type} (not implemented) [intent=direct_read availability=not_implemented operation=pendo.get.api-v1-blacklist-type-type]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 bulkdelete - Documented GET /api/v1/bulkdelete (not implemented) [intent=direct_read availability=not_implemented operation=pendo.get.api-v1-bulkdelete]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 bulkdelete id - Documented GET /api/v1/bulkdelete/{id} (not implemented) [intent=direct_read availability=not_implemented operation=pendo.get.api-v1-bulkdelete-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 feature - Documented GET /api/v1/feature (not implemented) [intent=direct_read availability=not_implemented operation=pendo.get.api-v1-feature]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 feedback options - Documented GET /api/v1/feedback/options (not implemented) [intent=direct_read availability=not_implemented operation=pendo.get.api-v1-feedback-options]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 guide - Documented GET /api/v1/guide (not implemented) [intent=direct_read availability=not_implemented operation=pendo.get.api-v1-guide]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 guide guideid history - Documented GET /api/v1/guide/{guideId}/history (not implemented) [intent=direct_read availability=not_implemented operation=pendo.get.api-v1-guide-guideid-history]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 guide localization export - Documented GET /api/v1/guide/localization/export (not implemented) [intent=direct_read availability=not_implemented operation=pendo.get.api-v1-guide-localization-export]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 guide order - Documented GET /api/v1/guide/order (not implemented) [intent=direct_read availability=not_implemented operation=pendo.get.api-v1-guide-order]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 metadata dependencies - Documented GET /api/v1/metadata/dependencies (not implemented) [intent=direct_read availability=not_implemented operation=pendo.get.api-v1-metadata-dependencies]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 metadata dependencies kind group fieldname - Documented GET /api/v1/metadata/dependencies/{kind}/{group}/{fieldName} (not implemented) [intent=direct_read availability=not_implemented operation=pendo.get.api-v1-metadata-dependencies-kind-group-fieldname]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 metadata kind group value id fieldname - Documented GET /api/v1/metadata/{kind}/{group}/value/{id}/{fieldName} (not implemented) [intent=direct_read availability=not_implemented operation=pendo.get.api-v1-metadata-kind-group-value-id-fieldname]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 metadata schema kind - Documented GET /api/v1/metadata/schema/{kind} (not implemented) [intent=direct_read availability=not_implemented operation=pendo.get.api-v1-metadata-schema-kind]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 page - Documented GET /api/v1/page (not implemented) [intent=direct_read availability=not_implemented operation=pendo.get.api-v1-page]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 report - Documented GET /api/v1/report (not implemented) [intent=direct_read availability=not_implemented operation=pendo.get.api-v1-report]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 report reportid results-csv - Documented GET /api/v1/report/{reportId}/results.csv (not implemented) [intent=direct_read availability=not_implemented operation=pendo.get.api-v1-report-reportid-results-csv]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 report reportid results-json - Documented GET /api/v1/report/{reportId}/results.json (not implemented) [intent=direct_read availability=not_implemented operation=pendo.get.api-v1-report-reportid-results-json]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 segment - Documented GET /api/v1/segment (not implemented) [intent=direct_read availability=not_implemented operation=pendo.get.api-v1-segment]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 segment id - Documented GET /api/v1/segment/{id} (not implemented) [intent=direct_read availability=not_implemented operation=pendo.get.api-v1-segment-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 segment segmentid status - Documented GET /api/v1/segment/{segmentId}/status (not implemented) [intent=direct_read availability=not_implemented operation=pendo.get.api-v1-segment-segmentid-status]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 segment segmentid visitors jobid results contenttype - Documented GET /api/v1/segment/{segmentId}/visitors/{jobId}/results/{contentType} (not implemented) [intent=direct_read availability=not_implemented operation=pendo.get.api-v1-segment-segmentid-visitors-jobid-results-contenttype]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 segment segmentid visitors jobid status - Documented GET /api/v1/segment/{segmentId}/visitors/{jobId}/status (not implemented) [intent=direct_read availability=not_implemented operation=pendo.get.api-v1-segment-segmentid-visitors-jobid-status]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 servername - Documented GET /api/v1/servername (not implemented) [intent=direct_read availability=not_implemented operation=pendo.get.api-v1-servername]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 servername flag flagname - Documented GET /api/v1/servername/flag/{flagname} (not implemented) [intent=direct_read availability=not_implemented operation=pendo.get.api-v1-servername-flag-flagname]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 servername name - Documented GET /api/v1/servername/{name} (not implemented) [intent=direct_read availability=not_implemented operation=pendo.get.api-v1-servername-name]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 token verify - Documented GET /api/v1/token/verify (not implemented) [intent=direct_read availability=not_implemented operation=pendo.get.api-v1-token-verify]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 tracktype - Documented GET /api/v1/tracktype (not implemented) [intent=direct_read availability=not_implemented operation=pendo.get.api-v1-tracktype]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 visitor b64-id - Documented GET /api/v1/visitor/{b64_id} (not implemented) [intent=direct_read availability=not_implemented operation=pendo.get.api-v1-visitor-b64-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 visitor visitorid - Documented GET /api/v1/visitor/{visitorId} (not implemented) [intent=direct_read availability=not_implemented operation=pendo.get.api-v1-visitor-visitorid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 visitor visitorid history - Documented GET /api/v1/visitor/{visitorId}/history (not implemented) [intent=direct_read availability=not_implemented operation=pendo.get.api-v1-visitor-visitorid-history]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get guide localization export-guideids guide-ids - Documented GET /guide/localization/export?guideids={guide_ids} (not implemented) [intent=direct_read availability=not_implemented operation=pendo.get.guide-localization-export-guideids-guide-ids]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get metadata kind group value id fieldname - Documented GET /metadata/{kind}/{group}/value/{id}/{fieldName} (not implemented) [intent=direct_read availability=not_implemented operation=pendo.get.metadata-kind-group-value-id-fieldname]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get report reportid results-csv - Documented GET /report/{reportId}/results.csv (not implemented) [intent=direct_read availability=not_implemented operation=pendo.get.report-reportid-results-csv]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get segment segmentid visitors jobid results contenttype - Documented GET /segment/{segmentId}/visitors/{jobId}/results/{contentType} (not implemented) [intent=direct_read availability=not_implemented operation=pendo.get.segment-segmentid-visitors-jobid-results-contenttype]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get segment segmentid visitors jobid status - Documented GET /segment/{segmentId}/visitors/{jobId}/status (not implemented) [intent=direct_read availability=not_implemented operation=pendo.get.segment-segmentid-visitors-jobid-status]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get token verify - Documented GET /token/verify (not implemented) [intent=direct_read availability=not_implemented operation=pendo.get.token-verify]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get visitor b64-id - Documented GET /visitor/{b64_id} (not implemented) [intent=direct_read availability=not_implemented operation=pendo.get.visitor-b64-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api patch api v1 feedback id - Documented PATCH /api/v1/feedback/{id} (not implemented) [intent=direct_write availability=not_implemented operation=pendo.patch.api-v1-feedback-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api patch api v1 segment segmentid visitor - Documented PATCH /api/v1/segment/{segmentId}/visitor (not implemented) [intent=direct_write availability=not_implemented operation=pendo.patch.api-v1-segment-segmentid-visitor]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post aggregation - Documented POST /aggregation (not implemented) [intent=direct_write availability=not_implemented operation=pendo.post.aggregation]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post api v1 aggregation - Documented POST /api/v1/aggregation (not implemented) [intent=direct_write availability=not_implemented operation=pendo.post.api-v1-aggregation]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post api v1 bulkdelete account - Documented POST /api/v1/bulkdelete/account (not implemented) [intent=direct_write availability=not_implemented operation=pendo.post.api-v1-bulkdelete-account]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post api v1 bulkdelete visitor - Documented POST /api/v1/bulkdelete/visitor (not implemented) [intent=direct_write availability=not_implemented operation=pendo.post.api-v1-bulkdelete-visitor]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post api v1 feedback - Documented POST /api/v1/feedback (not implemented) [intent=direct_write availability=not_implemented operation=pendo.post.api-v1-feedback]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post api v1 guide all visitor visitorid reset - Documented POST /api/v1/guide/all/visitor/{visitorId}/reset (not implemented) [intent=direct_write availability=not_implemented operation=pendo.post.api-v1-guide-all-visitor-visitorid-reset]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post api v1 guide guideid experiment export - Documented POST /api/v1/guide/{guideId}/experiment/export (not implemented) [intent=direct_write availability=not_implemented operation=pendo.post.api-v1-guide-guideid-experiment-export]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post api v1 guide guideid reset - Documented POST /api/v1/guide/{guideId}/reset (not implemented) [intent=direct_write availability=not_implemented operation=pendo.post.api-v1-guide-guideid-reset]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post api v1 guide guideid visitor visitorid reset - Documented POST /api/v1/guide/{guideId}/visitor/{visitorId}/reset (not implemented) [intent=direct_write availability=not_implemented operation=pendo.post.api-v1-guide-guideid-visitor-visitorid-reset]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post api v1 guide localization import - Documented POST /api/v1/guide/localization/import (not implemented) [intent=direct_write availability=not_implemented operation=pendo.post.api-v1-guide-localization-import]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post api v1 guide staged reset - Documented POST /api/v1/guide/staged/reset (not implemented) [intent=direct_write availability=not_implemented operation=pendo.post.api-v1-guide-staged-reset]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post api v1 metadata kind group value - Documented POST /api/v1/metadata/{kind}/{group}/value (not implemented) [intent=direct_write availability=not_implemented operation=pendo.post.api-v1-metadata-kind-group-value]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post api v1 metadata schema kind group - Documented POST /api/v1/metadata/schema/{kind}/{group} (not implemented) [intent=direct_write availability=not_implemented operation=pendo.post.api-v1-metadata-schema-kind-group]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post api v1 segment segmentid visitors - Documented POST /api/v1/segment/{segmentId}/visitors (not implemented) [intent=direct_write availability=not_implemented operation=pendo.post.api-v1-segment-segmentid-visitors]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post api v1 segment upload - Documented POST /api/v1/segment/upload (not implemented) [intent=direct_write availability=not_implemented operation=pendo.post.api-v1-segment-upload]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post bulkdelete account - Documented POST /bulkdelete/account (not implemented) [intent=direct_write availability=not_implemented operation=pendo.post.bulkdelete-account]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post bulkdelete visitor - Documented POST /bulkdelete/visitor (not implemented) [intent=direct_write availability=not_implemented operation=pendo.post.bulkdelete-visitor]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post data agentic - Documented POST /data/agentic (not implemented) [intent=direct_write availability=not_implemented operation=pendo.post.data-agentic]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post data track - Documented POST /data/track (not implemented) [intent=direct_write availability=not_implemented operation=pendo.post.data-track]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post guide guideid experiment export - Documented POST /guide/{guideId}/experiment/export (not implemented) [intent=direct_write availability=not_implemented operation=pendo.post.guide-guideid-experiment-export]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post guide localization import - Documented POST /guide/localization/import (not implemented) [intent=direct_write availability=not_implemented operation=pendo.post.guide-localization-import]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post https data-pendo-io data agentic - Documented POST https://data.pendo.io/data/agentic (not implemented) [intent=direct_write availability=not_implemented operation=pendo.post.https-data-pendo-io-data-agentic]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post https data-pendo-io data track - Documented POST https://data.pendo.io/data/track (not implemented) [intent=direct_write availability=not_implemented operation=pendo.post.https-data-pendo-io-data-track]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post metadata kind group value - Documented POST /metadata/{kind}/{group}/value (not implemented) [intent=direct_write availability=not_implemented operation=pendo.post.metadata-kind-group-value]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post metadata schema kind group - Documented POST /metadata/schema/{kind}/{group} (not implemented) [intent=direct_write availability=not_implemented operation=pendo.post.metadata-schema-kind-group]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put api v1 datasync storagetypes azure destinations default credentials - Documented PUT /api/v1/datasync/storageTypes/azure/destinations/Default/credentials (not implemented) [intent=direct_write availability=not_implemented operation=pendo.put.api-v1-datasync-storagetypes-azure-destinations-default-credentials]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put api v1 datasync storagetypes gcs destinations default credentials - Documented PUT /api/v1/datasync/storageTypes/gcs/destinations/Default/credentials (not implemented) [intent=direct_write availability=not_implemented operation=pendo.put.api-v1-datasync-storagetypes-gcs-destinations-default-credentials]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put api v1 datasync storagetypes s3 destinations default credentials - Documented PUT /api/v1/datasync/storageTypes/s3/destinations/Default/credentials (not implemented) [intent=direct_write availability=not_implemented operation=pendo.put.api-v1-datasync-storagetypes-s3-destinations-default-credentials]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put api v1 guide guideid segment - Documented PUT /api/v1/guide/{guideId}/segment (not implemented) [intent=direct_write availability=not_implemented operation=pendo.put.api-v1-guide-guideid-segment]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put api v1 guide guideid state - Documented PUT /api/v1/guide/{guideId}/state (not implemented) [intent=direct_write availability=not_implemented operation=pendo.put.api-v1-guide-guideid-state]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put api v1 metadata kind group value id fieldname - Documented PUT /api/v1/metadata/{kind}/{group}/value/{id}/{fieldName} (not implemented) [intent=direct_write availability=not_implemented operation=pendo.put.api-v1-metadata-kind-group-value-id-fieldname]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put api v1 metadata kind pendo value id blacklistguides - Documented PUT /api/v1/metadata/{kind}/pendo/value/{id}/blacklistguides (not implemented) [intent=direct_write availability=not_implemented operation=pendo.put.api-v1-metadata-kind-pendo-value-id-blacklistguides]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put api v1 metadata kind pendo value id donotprocess - Documented PUT /api/v1/metadata/{kind}/pendo/value/{id}/donotprocess (not implemented) [intent=direct_write availability=not_implemented operation=pendo.put.api-v1-metadata-kind-pendo-value-id-donotprocess]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put api v1 segment segmentid - Documented PUT /api/v1/segment/{segmentId} (not implemented) [intent=direct_write availability=not_implemented operation=pendo.put.api-v1-segment-segmentid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put api v1 segment segmentid visitor visitorid - Documented PUT /api/v1/segment/{segmentId}/visitor/{visitorId} (not implemented) [intent=direct_write availability=not_implemented operation=pendo.put.api-v1-segment-segmentid-visitor-visitorid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put datasync storagetypes azure destinations default credentials - Documented PUT /datasync/storageTypes/azure/destinations/Default/credentials (not implemented) [intent=direct_write availability=not_implemented operation=pendo.put.datasync-storagetypes-azure-destinations-default-credentials]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put datasync storagetypes gcs destinations default credentials - Documented PUT /datasync/storageTypes/gcs/destinations/Default/credentials (not implemented) [intent=direct_write availability=not_implemented operation=pendo.put.datasync-storagetypes-gcs-destinations-default-credentials]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put datasync storagetypes s3 destinations default credentials - Documented PUT /datasync/storageTypes/s3/destinations/Default/credentials (not implemented) [intent=direct_write availability=not_implemented operation=pendo.put.datasync-storagetypes-s3-destinations-default-credentials]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put metadata kind group value id fieldname - Documented PUT /metadata/{kind}/{group}/value/{id}/{fieldName} (not implemented) [intent=direct_write availability=not_implemented operation=pendo.put.metadata-kind-group-value-id-fieldname]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put metadata kind pendo value id blacklistguides - Documented PUT /metadata/{kind}/pendo/value/{id}/blacklistguides (not implemented) [intent=direct_write availability=not_implemented operation=pendo.put.metadata-kind-pendo-value-id-blacklistguides]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put metadata kind pendo value id donotprocess - Documented PUT /metadata/{kind}/pendo/value/{id}/donotprocess (not implemented) [intent=direct_write availability=not_implemented operation=pendo.put.metadata-kind-pendo-value-id-donotprocess]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - blacklist by type list - Run the blacklist by type ETL stream [intent=etl availability=implemented stream=blacklist_by_type]; notes: discrepancy=present-in-surface-absent-from-artifact
  - blacklist list - Run the blacklist ETL stream [intent=etl availability=implemented stream=blacklist]; notes: discrepancy=present-in-surface-absent-from-artifact
  - bulkdelete request list - Run the bulkdelete request ETL stream [intent=etl availability=implemented stream=bulkdelete_request]; notes: discrepancy=present-in-surface-absent-from-artifact
  - bulkdelete requests list - Run the bulkdelete requests ETL stream [intent=etl availability=implemented stream=bulkdelete_requests]; notes: discrepancy=present-in-surface-absent-from-artifact
  - change guide segment apply - Plan and execute the change guide segment reverse-ETL action [intent=reverse_etl availability=not_implemented write=change_guide_segment]; approval: requires plan, preview, approval, and execute; risk: changes the segment assigned to a Pendo guide; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - change guide state apply - Plan and execute the change guide state reverse-ETL action [intent=reverse_etl availability=not_implemented write=change_guide_state]; approval: requires plan, preview, approval, and execute; risk: changes a Pendo guide state such as public, staged, disabled, or draft; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - create feedback apply - Plan and execute the create feedback reverse-ETL action [intent=reverse_etl availability=implemented write=create_feedback]; approval: requires plan, preview, approval, and execute; risk: creates a Pendo Listen feedback item; approval required; flags: --accountId (required), --title (required), --visitorId (required)
  - create segment apply - Plan and execute the create segment reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_segment]; approval: requires plan, preview, approval, and execute; risk: creates a shared Pendo segment from visitor ids; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - delete feedback apply - Plan and execute the delete feedback reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_feedback]; approval: requires plan, preview, approval, and execute; risk: deletes Pendo Listen feedback items by id; destructive external mutation; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - delete segment apply - Plan and execute the delete segment reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_segment]; approval: requires plan, preview, approval, and execute; risk: deletes a Pendo segment; destructive external mutation; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - feature by id list - Run the feature by id ETL stream [intent=etl availability=implemented stream=feature_by_id]; notes: discrepancy=present-in-surface-absent-from-artifact
  - features by ids list - Run the features by ids ETL stream [intent=etl availability=implemented stream=features_by_ids]; notes: discrepancy=present-in-surface-absent-from-artifact
  - features list - Run the features ETL stream [intent=etl availability=implemented stream=features]; notes: discrepancy=present-in-surface-absent-from-artifact
  - feedback options list - Run the feedback options ETL stream [intent=etl availability=implemented stream=feedback_options]; notes: discrepancy=present-in-surface-absent-from-artifact
  - guide by id list - Run the guide by id ETL stream [intent=etl availability=implemented stream=guide_by_id]; notes: discrepancy=present-in-surface-absent-from-artifact
  - guide history list - Run the guide history ETL stream [intent=etl availability=implemented stream=guide_history]; notes: discrepancy=present-in-surface-absent-from-artifact
  - guide order list - Run the guide order ETL stream [intent=etl availability=implemented stream=guide_order]; notes: discrepancy=present-in-surface-absent-from-artifact
  - guides list - Run the guides ETL stream [intent=etl availability=implemented stream=guides]; notes: discrepancy=present-in-surface-absent-from-artifact
  - metadata dependencies list - Run the metadata dependencies ETL stream [intent=etl availability=implemented stream=metadata_dependencies]; notes: discrepancy=present-in-surface-absent-from-artifact
  - metadata field dependencies list - Run the metadata field dependencies ETL stream [intent=etl availability=implemented stream=metadata_field_dependencies]; notes: discrepancy=present-in-surface-absent-from-artifact
  - metadata schema list - Run the metadata schema ETL stream [intent=etl availability=implemented stream=metadata_schema]; notes: discrepancy=present-in-surface-absent-from-artifact
  - page by id list - Run the page by id ETL stream [intent=etl availability=implemented stream=page_by_id]; notes: discrepancy=present-in-surface-absent-from-artifact
  - pages by ids list - Run the pages by ids ETL stream [intent=etl availability=implemented stream=pages_by_ids]; notes: discrepancy=present-in-surface-absent-from-artifact
  - pages list - Run the pages ETL stream [intent=etl availability=implemented stream=pages]; notes: discrepancy=present-in-surface-absent-from-artifact
  - patch segment visitors apply - Plan and execute the patch segment visitors reverse-ETL action [intent=reverse_etl availability=not_implemented write=patch_segment_visitors]; approval: requires plan, preview, approval, and execute; risk: adds/removes a small batch of visitors for a Pendo segment; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - remove segment visitor apply - Plan and execute the remove segment visitor reverse-ETL action [intent=reverse_etl availability=not_implemented write=remove_segment_visitor]; approval: requires plan, preview, approval, and execute; risk: removes a visitor from a Pendo segment; destructive membership mutation; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - report results json list - Run the report results json ETL stream [intent=etl availability=implemented stream=report_results_json]; notes: discrepancy=present-in-surface-absent-from-artifact
  - reports list - Run the reports ETL stream [intent=etl availability=implemented stream=reports]; notes: discrepancy=present-in-surface-absent-from-artifact
  - reset all guides for visitor apply - Plan and execute the reset all guides for visitor reverse-ETL action [intent=reverse_etl availability=not_implemented write=reset_all_guides_for_visitor]; approval: requires plan, preview, approval, and execute; risk: resets guide-seen state for one visitor across all guides; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - reset all staged guides apply - Plan and execute the reset all staged guides reverse-ETL action [intent=reverse_etl availability=implemented write=reset_all_staged_guides]; approval: requires plan, preview, approval, and execute; risk: resets all staged guides in the subscription; approval required
  - reset guide for visitor apply - Plan and execute the reset guide for visitor reverse-ETL action [intent=reverse_etl availability=not_implemented write=reset_guide_for_visitor]; approval: requires plan, preview, approval, and execute; risk: resets guide-seen state for one visitor; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - reset staged guide apply - Plan and execute the reset staged guide reverse-ETL action [intent=reverse_etl availability=not_implemented write=reset_staged_guide]; approval: requires plan, preview, approval, and execute; risk: resets one staged guide; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - segment by id list - Run the segment by id ETL stream [intent=etl availability=implemented stream=segment_by_id]; notes: discrepancy=present-in-surface-absent-from-artifact
  - segment status list - Run the segment status ETL stream [intent=etl availability=implemented stream=segment_status]; notes: discrepancy=present-in-surface-absent-from-artifact
  - segments list - Run the segments ETL stream [intent=etl availability=implemented stream=segments]; notes: discrepancy=present-in-surface-absent-from-artifact
  - server by name list - Run the server by name ETL stream [intent=etl availability=implemented stream=server_by_name]; notes: discrepancy=present-in-surface-absent-from-artifact
  - servers by flag list - Run the servers by flag ETL stream [intent=etl availability=implemented stream=servers_by_flag]; notes: discrepancy=present-in-surface-absent-from-artifact
  - servers list - Run the servers ETL stream [intent=etl availability=implemented stream=servers]; notes: discrepancy=present-in-surface-absent-from-artifact
  - start segment visitor export apply - Plan and execute the start segment visitor export reverse-ETL action [intent=reverse_etl availability=not_implemented write=start_segment_visitor_export]; approval: requires plan, preview, approval, and execute; risk: starts an asynchronous Pendo segment visitor export job; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - tracktype by id list - Run the tracktype by id ETL stream [intent=etl availability=implemented stream=tracktype_by_id]; notes: discrepancy=present-in-surface-absent-from-artifact
  - tracktypes list - Run the tracktypes ETL stream [intent=etl availability=implemented stream=tracktypes]; notes: discrepancy=present-in-surface-absent-from-artifact
  - update feedback apply - Plan and execute the update feedback reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_feedback]; approval: requires plan, preview, approval, and execute; risk: updates a Pendo Listen feedback item; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - update segment apply - Plan and execute the update segment reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_segment]; approval: requires plan, preview, approval, and execute; risk: replaces the visitor membership for a Pendo segment; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - visitor by id list - Run the visitor by id ETL stream [intent=etl availability=implemented stream=visitor_by_id]; notes: discrepancy=present-in-surface-absent-from-artifact
  - visitor history list - Run the visitor history ETL stream [intent=etl availability=implemented stream=visitor_history]; notes: discrepancy=present-in-surface-absent-from-artifact
  - visitors list - Run the visitors ETL stream [intent=etl availability=implemented stream=visitors]; notes: discrepancy=present-in-surface-absent-from-artifact

## Commands

### Inspect as a manual

```bash
pm connectors inspect pendo
```

### Inspect as structured JSON

```bash
pm connectors inspect pendo --json
```

## Agent Rules

- Run pm connectors inspect pendo before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
