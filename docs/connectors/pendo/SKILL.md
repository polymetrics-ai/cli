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
- integration_key (secret) (required)

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
